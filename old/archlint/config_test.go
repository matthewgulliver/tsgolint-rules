package main

import (
	"testing"
)

const domainFile = "packages/gifting/hexagon/domain/src/occasions/occasion.ts"
const portFile = "packages/gifting/hexagon/application/src/ports/occasion-repository.ts"
const adapterFile = "packages/gifting/hexagon/adapters/driven/postgres/src/repo.ts"

// Two of this repository's rules and one of upstream's, with the trees they
// really declare. Scope is decided here and nowhere else, so these are the
// cases that say which rule looks at which file.
var registered = registry{
	"domain-state-is-deeply-readonly": {onByDefault: true, files: []string{"**/hexagon/domain/**", "**/shared-kernel/**"}},
	"port-behaviour-is-an-interface":  {onByDefault: true, files: []string{"**/ports/**"}},
	"switch-exhaustiveness-check":     {},
}

func resolveOrFail(t *testing.T, source string, file string) map[string]map[string]any {
	t.Helper()
	config, err := parseConfig([]byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := config.validate(registered); err != nil {
		t.Fatalf("validate: %v", err)
	}
	return config.resolve(file, registered)
}

func TestARuleJudgesTheTreeItDeclares(t *testing.T) {
	t.Parallel()

	if _, on := empty().resolve(domainFile, registered)["domain-state-is-deeply-readonly"]; !on {
		t.Fatal("a rule did not run inside the tree it declares")
	}
	if _, on := empty().resolve(portFile, registered)["domain-state-is-deeply-readonly"]; on {
		t.Fatal("a rule ran outside the tree it declares")
	}
	if _, on := empty().resolve(domainFile, registered)["port-behaviour-is-an-interface"]; on {
		t.Fatal("a rule ran outside the tree it declares")
	}
	if _, on := empty().resolve(domainFile, registered)["switch-exhaustiveness-check"]; on {
		t.Fatal("an upstream rule ran without being asked for")
	}
}

// The tree a rule is judged against is handed back to it. Only
// `domain-type-is-declared-once` reads it, and it indexes that tree.
func TestARuleIsToldTheTreeItWasJudgedAgainst(t *testing.T) {
	t.Parallel()
	options := empty().resolve(domainFile, registered)["domain-state-is-deeply-readonly"]

	files, told := options["files"].([]string)
	if !told || len(files) != 2 || files[0] != "**/hexagon/domain/**" {
		t.Fatalf("the rule was not told the tree it judges: %v", options)
	}
}

func TestOffSilencesARuleEverywhere(t *testing.T) {
	t.Parallel()
	running := resolveOrFail(t, `{"rules": {"domain-state-is-deeply-readonly": "off"}}`, domainFile)

	if _, on := running["domain-state-is-deeply-readonly"]; on {
		t.Fatal(`a rule set to "off" still ran`)
	}
}

// Upstream's rules declare no tree, so asking for one is asking for it over
// every file the tsconfig includes.
func TestAnUpstreamRuleRunsOnlyWhenNamedAndThenEverywhere(t *testing.T) {
	t.Parallel()
	source := `{"rules": {"switch-exhaustiveness-check": "error"}}`

	for _, file := range []string{domainFile, portFile, adapterFile} {
		if _, on := resolveOrFail(t, source, file)["switch-exhaustiveness-check"]; !on {
			t.Fatalf("an upstream rule named in the config did not run on %s", file)
		}
	}
}

func TestAnOverrideScopesARuleToItsOwnFiles(t *testing.T) {
	t.Parallel()
	source := `{
	  "rules": {"domain-state-is-deeply-readonly": "off"},
	  "overrides": [
	    {"files": ["**/adapters/**"], "rules": {"domain-state-is-deeply-readonly": "error"}}
	  ]
	}`

	if _, on := resolveOrFail(t, source, domainFile)["domain-state-is-deeply-readonly"]; on {
		t.Fatal("a rule ran on a file its override does not name")
	}

	options, on := resolveOrFail(t, source, adapterFile)["domain-state-is-deeply-readonly"]
	if !on {
		t.Fatal("a rule did not run on the file its override names — an override could not widen a rule")
	}
	files, told := options["files"].([]string)
	if !told || len(files) != 1 || files[0] != "**/adapters/**" {
		t.Fatalf("the override's files did not reach the rule: %v", options)
	}
}

func TestARulesOwnOptionsSurvive(t *testing.T) {
	t.Parallel()
	source := `{
	  "overrides": [
	    {
	      "files": ["**/domain/**"],
	      "rules": {
	        "domain-state-is-deeply-readonly": ["error", {"mutableCollectionTypeNames": ["Array"]}]
	      }
	    }
	  ]
	}`
	options := resolveOrFail(t, source, domainFile)["domain-state-is-deeply-readonly"]

	names, ok := options["mutableCollectionTypeNames"].([]any)
	if !ok || len(names) != 1 || names[0] != "Array" {
		t.Fatalf("a rule's own option did not reach it: %v", options)
	}
}

func TestExplicitFilesBeatTheBlock(t *testing.T) {
	t.Parallel()
	source := `{
	  "overrides": [
	    {
	      "files": ["**/domain/**"],
	      "rules": {"domain-state-is-deeply-readonly": ["error", {"files": ["**/elsewhere/**"]}]}
	    }
	  ]
	}`

	if _, on := resolveOrFail(t, source, domainFile)["domain-state-is-deeply-readonly"]; on {
		t.Fatal("a rule ran on a file the block names and its own `files` do not")
	}

	elsewhere := "packages/gifting/elsewhere/src/occasion.ts"
	if _, on := resolveOrFail(t, source, elsewhere)["domain-state-is-deeply-readonly"]; on {
		t.Fatal("a rule ran outside the block that named it")
	}
}

func TestALaterOverrideWins(t *testing.T) {
	t.Parallel()
	source := `{
	  "overrides": [
	    {"files": ["**/domain/**"], "rules": {"port-behaviour-is-an-interface": "error"}},
	    {"files": ["**/occasions/**"], "rules": {"port-behaviour-is-an-interface": "off"}}
	  ]
	}`

	if _, on := resolveOrFail(t, source, domainFile)["port-behaviour-is-an-interface"]; on {
		t.Fatal("an earlier override beat a later one")
	}
}

func TestAnUnknownRuleNameIsAnError(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`{"rules": {"no-such-rule": "error"}}`,
		`{"overrides": [{"files": ["**"], "rules": {"no-such-rule": "error"}}]}`,
	} {
		config, err := parseConfig([]byte(source))
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if err := config.validate(registered); err == nil {
			t.Fatalf("an unknown rule name passed validation: %s", source)
		}
	}
}

func TestAMalformedConfigIsAnError(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`{`,
		`{"rules": {"port-behaviour-is-an-interface": "warn"}}`,
		`{"rules": {"port-behaviour-is-an-interface": 3}}`,
		`{"overrides": [{"rules": {"port-behaviour-is-an-interface": "error"}}]}`,
		`{"rules": {"port-behaviour-is-an-interface": ["error", {"files": "**/ports/**"}]}}`,
		`{"rules": {"port-behaviour-is-an-interface": ["error", {"files": [3]}]}}`,
	} {
		if _, err := parseConfig([]byte(source)); err == nil {
			t.Fatalf("a malformed config parsed cleanly: %s", source)
		}
	}
}

// A rule of ours that named no tree would judge every file the tsconfig
// includes, and nothing at run time would say so — it would simply start
// reporting on files that are none of its business.
func TestEveryRuleOfOursNamesTheTreeItJudges(t *testing.T) {
	t.Parallel()
	for _, r := range archRules {
		if len(r.Files) == 0 {
			t.Errorf("%s names no tree to judge", r.Name)
		}
	}
}
