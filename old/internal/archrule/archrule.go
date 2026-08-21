// Package archrule is what `packages/oxlint/rule.ts` is to the JavaScript
// half: the shared contract every rule in this repository is written against.
package archrule

import "github.com/typescript-eslint/tsgolint/internal/rule"

// Where the reasoning behind a rule is written down. Kept identical to
// `packages/oxlint/rule.ts`'s own constant, so both halves send a reader to the
// same place.
const source = "https://github.com/matthewgulliver/tsgolint-rules/blob/main/docs/rules"

// DocumentedAt is the page a rule's diagnostic sends its reader to.
//
// `packages/oxlint/rule.ts` requires a `docs.url` of every JavaScript rule in
// the type itself, so a rule without one does not compile. `rule.Rule` is
// upstream's and carries no field for it, so the Go half derives the page from
// the rule's own name when the diagnostic is printed. Nothing to forget, and
// nothing to keep in step by hand.
func DocumentedAt(name string) string {
	return source + "/" + name + ".md"
}

// Rule is one of this repository's rules together with the tree it judges.
//
// `archlint` runs a rule over the files its tree names and no others, so a
// rule contains no scope check of its own: scope is decided once, in the
// entrypoint, from `.archtypesrc.json` and this declaration.
type Rule struct {
	rule.Rule
	// Files is the tree the rule judges when the configuration does not name
	// one. Never empty for a rule of ours: a rule that names no tree judges
	// every file the tsconfig includes, which is how `lint:arch-types` came to
	// print 580 diagnostics that belonged to nobody.
	Files []string
}
