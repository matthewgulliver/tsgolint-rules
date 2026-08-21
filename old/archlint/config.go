package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/typescript-eslint/tsgolint/internal/archscope"
)

// The configuration file `archlint` reads, in `.oxlintrc.json`'s grammar: a
// `rules` map of name to severity, and `overrides` blocks that re-decide a rule
// for the files they name.
//
// Scope is decided here and nowhere else. A rule judges the file it is handed;
// which files those are is this file's answer, from the block that named the
// rule, the `files` option the rule was given, or the tree the rule's own
// package declares.
const configFileBasename = ".archtypesrc.json"

type ruleSetting struct {
	enabled bool
	// files is the tree this setting names, and nil when it names none — the
	// rule's own declared tree then stands.
	files   []string
	options map[string]any
}

// registration is what a rule name resolves to: whether it runs unless the
// configuration says otherwise, and the tree it judges by default. Upstream's
// rules declare no tree, and a configuration that asks for one gets it over
// every file the tsconfig includes unless a block narrows it.
type registration struct {
	onByDefault bool
	files       []string
}

type registry map[string]registration

type override struct {
	files []string
	rules map[string]ruleSetting
}

type config struct {
	rules     map[string]ruleSetting
	overrides []override
}

func empty() config { return config{} }

type rawOverride struct {
	Files []string                   `json:"files"`
	Rules map[string]json.RawMessage `json:"rules"`
}

type rawConfig struct {
	Schema    string                     `json:"$schema"`
	Rules     map[string]json.RawMessage `json:"rules"`
	Overrides []rawOverride              `json:"overrides"`
}

func parseConfig(source []byte) (config, error) {
	decoder := json.NewDecoder(strings.NewReader(string(source)))
	decoder.DisallowUnknownFields()

	var raw rawConfig
	if err := decoder.Decode(&raw); err != nil {
		return config{}, fmt.Errorf("%s is not valid: %w", configFileBasename, err)
	}

	parsed := config{}

	var err error
	if parsed.rules, err = parseRules(raw.Rules); err != nil {
		return config{}, err
	}

	for i, block := range raw.Overrides {
		if len(block.Files) == 0 {
			return config{}, fmt.Errorf("overrides[%d] names no files, so it decides nothing", i)
		}
		rules, err := parseRules(block.Rules)
		if err != nil {
			return config{}, fmt.Errorf("overrides[%d]: %w", i, err)
		}
		parsed.overrides = append(parsed.overrides, override{files: block.Files, rules: rules})
	}

	return parsed, nil
}

func parseRules(raw map[string]json.RawMessage) (map[string]ruleSetting, error) {
	if raw == nil {
		return nil, nil
	}
	rules := make(map[string]ruleSetting, len(raw))
	for name, value := range raw {
		setting, err := parseRuleSetting(value)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		rules[name] = setting
	}
	return rules, nil
}

// A rule is `"error"`, `"off"`, or `["error", { … }]` — the shapes
// `.oxlintrc.json` already uses, so one grammar covers both halves.
func parseRuleSetting(value json.RawMessage) (ruleSetting, error) {
	var severity string
	if err := json.Unmarshal(value, &severity); err == nil {
		enabled, err := parseSeverity(severity)
		return ruleSetting{enabled: enabled}, err
	}

	var pair []json.RawMessage
	if err := json.Unmarshal(value, &pair); err != nil || len(pair) == 0 || len(pair) > 2 {
		return ruleSetting{}, errors.New(`must be "error", "off", or ["error", { … }]`)
	}

	if err := json.Unmarshal(pair[0], &severity); err != nil {
		return ruleSetting{}, errors.New(`the first entry must be "error" or "off"`)
	}
	enabled, err := parseSeverity(severity)
	if err != nil || len(pair) == 1 {
		return ruleSetting{enabled: enabled}, err
	}

	var options map[string]any
	if err := json.Unmarshal(pair[1], &options); err != nil {
		return ruleSetting{}, errors.New("the second entry must be an options object")
	}
	files, err := takeFiles(options)
	if err != nil {
		return ruleSetting{}, err
	}
	return ruleSetting{enabled: enabled, files: files, options: options}, nil
}

// takeFiles reads `files` out of a rule's options. It is the one option
// `archlint` interprets itself; the rest are the rule's own vocabulary and
// pass through untouched. An empty list names no tree, and leaves the rule's
// own declared one standing.
func takeFiles(options map[string]any) ([]string, error) {
	given, named := options["files"]
	if !named {
		return nil, nil
	}
	delete(options, "files")

	globs, ok := given.([]any)
	if !ok {
		return nil, errors.New("`files` must be an array of globs")
	}
	files := make([]string, 0, len(globs))
	for _, glob := range globs {
		pattern, ok := glob.(string)
		if !ok {
			return nil, errors.New("`files` must be an array of globs")
		}
		files = append(files, pattern)
	}
	if len(files) == 0 {
		return nil, nil
	}
	return files, nil
}

func parseSeverity(severity string) (bool, error) {
	switch severity {
	case "error":
		return true, nil
	case "off":
		return false, nil
	}
	// No `warn`: this binary either fails a build or does not, and a severity
	// that reports without failing is a gate nobody notices going out.
	return false, fmt.Errorf(`severity must be "error" or "off", not %q`, severity)
}

// validate refuses a name no rule answers to. Left alone, a typo is a rule
// silently not running, which is the failure this repository keeps writing
// down: a rule that judges nothing is a green run.
func (c config) validate(reg registry) error {
	unknown := []string{}
	check := func(rules map[string]ruleSetting) {
		for name := range rules {
			if _, known := reg[name]; !known {
				unknown = append(unknown, name)
			}
		}
	}
	check(c.rules)
	for _, block := range c.overrides {
		check(block.rules)
	}
	if len(unknown) == 0 {
		return nil
	}
	slices.Sort(unknown)
	return fmt.Errorf("no rule is called %s", strings.Join(slices.Compact(unknown), ", "))
}

// resolve is the rules that judge one file, and the options each runs with. A
// rule is absent when the configuration switches it off, and when the file is
// outside the tree it judges — the `files` option it was given, else the block
// that named it, else the tree its own package declares.
func (c config) resolve(file string, reg registry) map[string]map[string]any {
	settings := make(map[string]ruleSetting, len(reg))
	for name, declared := range reg {
		settings[name] = ruleSetting{enabled: declared.onByDefault}
	}
	for name, setting := range c.rules {
		settings[name] = setting
	}
	for _, block := range c.overrides {
		if !archscope.Includes(block.files, file) {
			continue
		}
		for name, setting := range block.rules {
			if setting.files == nil {
				setting.files = block.files
			}
			settings[name] = setting
		}
	}

	running := make(map[string]map[string]any, len(settings))
	for name, setting := range settings {
		if !setting.enabled {
			continue
		}
		judged := setting.files
		if judged == nil {
			judged = reg[name].files
		}
		// A rule that names no tree judges every file: that is what asking for
		// one of upstream's rules means, and no rule of ours may do it.
		if len(judged) > 0 && !archscope.Includes(judged, file) {
			continue
		}
		running[name] = withFiles(setting.options, judged)
	}
	return running
}

// enabledRules is every rule this configuration lets run somewhere, for the
// summary line. Which rules run on a given file is a per-file answer; how many
// rules are switched on at all is not.
func (c config) enabledRules(reg registry) []string {
	enabled := map[string]bool{}
	for name, declared := range reg {
		enabled[name] = declared.onByDefault
	}
	for name, setting := range c.rules {
		enabled[name] = setting.enabled
	}
	for _, block := range c.overrides {
		for name, setting := range block.rules {
			if setting.enabled {
				enabled[name] = true
			}
		}
	}

	names := []string{}
	for name, on := range enabled {
		if on {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

// withFiles tells a rule the tree it was judged against. Only
// `domain-type-is-declared-once` reads it — it indexes that tree instead of
// judging one file at a time — but every rule is told, because a rule that
// needs to know says so in its own options and not here.
func withFiles(options map[string]any, files []string) map[string]any {
	if len(files) == 0 {
		return options
	}
	told := make(map[string]any, len(options)+1)
	for key, value := range options {
		told[key] = value
	}
	told["files"] = files
	return told
}
