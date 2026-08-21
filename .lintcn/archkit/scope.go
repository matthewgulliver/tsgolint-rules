// Package archscope matches a file against the `.gitignore`-style globs that
// `.oxlintrc.json` uses to scope a JS rule.
//
// `archlint` matches every rule's tree here once per file, before the rule
// runs. A rule reaches for this package only for a second path vocabulary of
// its own — the trees a domain signature may name, where a row type is
// declared — never to decide whether it should be looking at this file.
package archkit

import (
	"path"
	"regexp"
	"strings"
)

// Includes reports whether fileName is judged by any of the patterns.
//
// Patterns are `.gitignore`-style globs over path segments: `**` matches any
// run of segments including none, and `*`/`?`/`[…]` match within one segment.
func Includes(patterns []string, fileName string) bool {
	segments := split(fileName)
	for _, pattern := range patterns {
		if matches(split(pattern), segments) {
			return true
		}
	}
	return false
}

func split(p string) []string {
	normalised := strings.ReplaceAll(p, "\\", "/")
	parts := strings.Split(normalised, "/")
	kept := parts[:0]
	for _, part := range parts {
		if part != "" && part != "." {
			kept = append(kept, part)
		}
	}
	return kept
}

func matches(pattern []string, name []string) bool {
	if len(pattern) == 0 {
		return len(name) == 0
	}

	if pattern[0] == "**" {
		// `**` matches nothing at all, or one more segment and then itself.
		for consumed := 0; consumed <= len(name); consumed++ {
			if matches(pattern[1:], name[consumed:]) {
				return true
			}
		}
		return false
	}

	if len(name) == 0 {
		return false
	}
	ok, err := path.Match(pattern[0], name[0])
	if err != nil || !ok {
		return false
	}
	return matches(pattern[1:], name[1:])
}

// IsOutsideDependency reports whether a declaration file belongs to a package
// dependency or to the TypeScript standard library rather than to this
// repository's own source.
//
// This is the one cross-file fact a per-file syntactic rule cannot reach, and
// the whole reason several of these rules can exist at all.
func IsOutsideDependency(fileName string) bool {
	return IsPackageDependency(fileName) || IsStandardLibrary(fileName)
}

// IsPackageDependency reports whether a declaration file was installed rather
// than written here.
func IsPackageDependency(fileName string) bool {
	for _, segment := range split(fileName) {
		if segment == "node_modules" {
			return true
		}
	}
	return false
}

// IsStandardLibrary reports whether a declaration file is one of TypeScript's
// bundled `lib.*.d.ts` files — where `Date`, `Map` and `Response` alike are
// declared, so a rule must decide deliberately which side of the line it wants.
func IsStandardLibrary(fileName string) bool {
	base := path.Base(strings.ReplaceAll(fileName, "\\", "/"))
	return strings.HasPrefix(base, "lib.") && strings.HasSuffix(base, ".d.ts")
}

// Compile turns configured patterns into expressions, dropping any that do not
// compile: a misconfigured pattern identifies nothing rather than failing a run.
func Compile(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if expression, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, expression)
		}
	}
	return compiled
}

// ContextOf returns the identity the first matching pattern captures in a
// `(?<name>…)` group, taken at the innermost place that pattern matches.
//
// Innermost, because these file names are absolute and a pattern is free to
// match above the repository as well as inside it. A checkout under
// `~/packages/…`, or this repository's own fixtures under
// `packages/tsgolint/fixtures/`, made every context below answer that outer
// segment; they then compared equal and a real crossing went unreported.
//
// Matches are found by restarting one byte past each match's start rather than
// past its end, so a pattern that consumes the separator between two contexts
// — as `(^|/)packages/(?<name>[^/]+)(/|$)` does — still sees the second one.
func ContextOf(patterns []*regexp.Regexp, file string) (string, bool) {
	for _, expression := range patterns {
		group := captureIndex(expression)
		if group < 0 {
			continue
		}
		identity, identified := "", false
		for at := 0; at <= len(file); {
			match := expression.FindStringSubmatchIndex(file[at:])
			if match == nil {
				break
			}
			start, end := match[2*group], match[2*group+1]
			if start >= 0 && end > start {
				identity, identified = file[at+start:at+end], true
			}
			at += match[0] + 1
		}
		if identified {
			return identity, true
		}
	}
	return "", false
}

// captureIndex is the position of the `name` group, or -1 when the pattern has
// none and so identifies nothing.
func captureIndex(expression *regexp.Regexp) int {
	for index, group := range expression.SubexpNames() {
		if group == "name" {
			return index
		}
	}
	return -1
}
