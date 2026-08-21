package domain_type_is_declared_once

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const canonical = "packages/gifting/hexagon/domain/src/occasions/occasion.ts"

// The second declaration is in another file, which is the whole point: no
// per-file rule can see it.
var elsewhere = map[string]string{
	"packages/gifting/hexagon/domain/src/pledges/occasion.ts": `
    export type Occasion = { readonly id: string; readonly closed: boolean }
  `,
	// The same word in another context is the language test working, not a
	// duplicate.
	"packages/notifications/hexagon/domain/src/reminders/reminder.ts": `
    export type Reminder = { readonly at: string }
  `,
	// Outside the judged tree.
	"packages/gifting/hexagon/application/src/occasion.ts": `
    export type Pledge = { readonly id: string }
  `,
}

// Two files in one context declaring one name, both outside the tree this
// rule declares.
var applicationReminders = map[string]string{
	"packages/notifications/hexagon/application/src/reminder.ts": `
    export type Reminder = { readonly at: string }
  `,
}

func TestDomainTypeIsDeclaredOnce(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &DomainTypeIsDeclaredOnceRule.Rule, []rule_tester.ValidTestCase{
		// Declared here and nowhere else in this context.
		{
			Code:     `export type OccasionId = string & { readonly brand: unique symbol }`,
			FileName: canonical,
			Files:    elsewhere,
		},
		// The language test: Gifting declares `Occasion` too, and the same
		// word meaning something else in another context is the boundary
		// working rather than a duplicate.
		{
			Code:     `export type Occasion = { readonly id: string; readonly remindAt: string }`,
			FileName: "packages/notifications/hexagon/domain/src/occasions/occasion.ts",
			Files:    elsewhere,
		},
		// The same name outside the domain tree is not this rule's subject.
		{
			Code:     `export type Pledge = { readonly id: string; readonly amount: number }`,
			FileName: canonical,
			Files:    elsewhere,
		},
		// Unexported, so not the model's published vocabulary.
		{
			Code:     `type Occasion = { readonly id: string }`,
			FileName: canonical,
			Files:    elsewhere,
		},
		// Two application files, one name, and the default tree indexes neither.
		{
			Code:     `export type Reminder = { readonly at: string; readonly channel: string }`,
			FileName: "packages/notifications/hexagon/application/src/read-models/reminder.ts",
			Files:    applicationReminders,
		},
	}, []rule_tester.InvalidTestCase{
		// The canonical-type claim: two files in one context, two `Occasion`s,
		// both compiling.
		{
			Code:     `export type Occasion = { readonly id: string; readonly budget: number }`,
			FileName: canonical,
			Files:    elsewhere,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "redeclaredDomainType"}},
		},
		// The same two files, once the tree `archlint` resolved says to index
		// them. This is the one rule that is told its own tree, because it
		// reads the tree rather than judging one file at a time.
		{
			Code:     `export type Reminder = { readonly at: string; readonly channel: string }`,
			FileName: "packages/notifications/hexagon/application/src/read-models/reminder.ts",
			Files:    applicationReminders,
			Options:  Options{Files: []string{"**/hexagon/application/**"}},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "redeclaredDomainType"}},
		},
		// Spelled as an interface, which is the same claim.
		{
			Code: `
        export interface Occasion {
          readonly id: string
        }
      `,
			FileName: canonical,
			Files:    elsewhere,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "redeclaredDomainType"}},
		},
	})
}
