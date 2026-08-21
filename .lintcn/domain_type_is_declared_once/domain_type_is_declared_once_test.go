package domain_type_is_declared_once

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const canonical = "packages/gifting/hexagon/domain/src/occasions/occasion.ts"

var elsewhere = map[string]string{
	"packages/gifting/hexagon/domain/src/pledges/occasion.ts": `
    export type Occasion = { readonly id: string; readonly closed: boolean }
  `,
	"packages/notifications/hexagon/domain/src/reminders/reminder.ts": `
    export type Reminder = { readonly at: string }
  `,
	"packages/gifting/hexagon/application/src/occasion.ts": `
    export type Pledge = { readonly id: string }
  `,
}

var applicationReminders = map[string]string{
	"packages/notifications/hexagon/application/src/reminder.ts": `
    export type Reminder = { readonly at: string }
  `,
}

func TestDomainTypeIsDeclaredOnce(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &DomainTypeIsDeclaredOnceRule, []rule_tester.ValidTestCase{
		{
			Code:     `export type OccasionId = string & { readonly brand: unique symbol }`,
			FileName: canonical,
			Files:    elsewhere,
		},
		{
			Code:     `export type Occasion = { readonly id: string; readonly remindAt: string }`,
			FileName: "packages/notifications/hexagon/domain/src/occasions/occasion.ts",
			Files:    elsewhere,
		},
		{
			Code:     `export type Pledge = { readonly id: string; readonly amount: number }`,
			FileName: canonical,
			Files:    elsewhere,
		},
		{
			Code:     `type Occasion = { readonly id: string }`,
			FileName: canonical,
			Files:    elsewhere,
		},
		{
			Code:     `export type Reminder = { readonly at: string; readonly channel: string }`,
			FileName: "packages/notifications/hexagon/application/src/read-models/reminder.ts",
			Files:    applicationReminders,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:     `export type Occasion = { readonly id: string; readonly budget: number }`,
			FileName: canonical,
			Files:    elsewhere,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "redeclaredDomainType"}},
		},
		{
			Code:     `export type Reminder = { readonly at: string; readonly channel: string }`,
			FileName: "packages/notifications/hexagon/application/src/read-models/reminder.ts",
			Files:    applicationReminders,
			Options:  Options{Files: []string{"**/hexagon/application/**"}},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "redeclaredDomainType"}},
		},
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
