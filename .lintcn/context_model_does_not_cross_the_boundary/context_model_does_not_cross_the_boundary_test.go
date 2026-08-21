package context_model_does_not_cross_the_boundary

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inNotifications = "packages/notifications/src/reminders/reminder.ts"

var contexts = map[string]string{
	"packages/gifting/src/occasions/occasion.ts": `
    export type Occasion = { readonly id: string }
  `,
	"packages/gifting/public.ts": `
    export type { Occasion } from "./src/occasions/occasion"
    export type OccasionSummary = { readonly id: string; readonly closed: boolean }
  `,
	"packages/shared-kernel/src/money.ts": `
    export type Money = { readonly minorUnits: number }
  `,
	"packages/notifications/src/reminders/schedule.ts": `
    export type Schedule = { readonly at: string }
  `,
	"packages/gifting/index.ts": `
    export type Voucher = { readonly code: string }
  `,
}

const inNestedNotifications = "packages/monorepo/packages/notifications/src/reminders/reminder.ts"

var nestedContexts = map[string]string{
	"packages/monorepo/packages/gifting/src/occasions/occasion.ts": `
    export type Occasion = { readonly id: string }
  `,
	"packages/monorepo/packages/notifications/src/reminders/schedule.ts": `
    export type Schedule = { readonly at: string }
  `,
}

func TestContextModelDoesNotCrossTheBoundary(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &ContextModelDoesNotCrossTheBoundaryRule, []rule_tester.ValidTestCase{
		{
			Code: `
        import type { Schedule } from "./schedule"
        export const nextReminder = (schedule: Schedule): Schedule => schedule
      `,
			FileName: inNotifications,
			Files:    contexts,
		},
		{
			Code: `
        import type { OccasionSummary } from "../../../gifting/public"
        import type { Schedule } from "./schedule"
        export const remindAbout = (summary: OccasionSummary): Schedule =>
          ({ at: summary.id })
      `,
			FileName: inNotifications,
			Files:    contexts,
		},
		{
			Code: `
        import type { Money } from "../../../shared-kernel/src/money"
        export const owed = (amount: Money): Money => amount
      `,
			FileName: inNotifications,
			Files:    contexts,
		},
		{
			Code: `
        import type { Occasion } from "../../../gifting/src/occasions/occasion"
        const describe = (occasion: Occasion): string => occasion.id
      `,
			FileName: inNotifications,
			Files:    contexts,
		},
		{
			Code: `
        import type { Schedule } from "./schedule"
        export type Reminder = { readonly schedule: Schedule }
      `,
			FileName: inNotifications,
			Files:    contexts,
		},
		{
			Code: `
        import type { Occasion } from "../packages/gifting/src/occasions/occasion"
        export const describe = (occasion: Occasion): string => occasion.id
      `,
			FileName: "apps/web/src/page.ts",
			Files:    contexts,
		},
		{
			Code: `
        import type { Schedule } from "./schedule"
        export const nextReminder = (schedule: Schedule): Schedule => schedule
      `,
			FileName: inNestedNotifications,
			Files:    nestedContexts,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        import type { Occasion } from "../../../gifting/src/occasions/occasion"
        export const remindAbout = (occasion: Occasion): string => occasion.id
      `,
			FileName: inNotifications,
			Files:    contexts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "crossContextModelType"}},
		},
		{
			Code: `
        import type { Occasion } from "../../../gifting/public"
        import type { Schedule } from "./schedule"
        export const remindAbout = (occasion: Occasion): Schedule =>
          ({ at: occasion.id })
      `,
			FileName: inNotifications,
			Files:    contexts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "crossContextModelType"}},
		},
		{
			Code: `
        import type { Occasion } from "../../../gifting/public"
        export type Reminder = { readonly about: Occasion }
      `,
			FileName: inNotifications,
			Files:    contexts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "crossContextModelType"}},
		},
		{
			Code: `
        import type { Occasion } from "../../../gifting/src/occasions/occasion"
        import type { Schedule } from "./schedule"
        export function subjectOf(schedule: Schedule): Occasion {
          return { id: schedule.at }
        }
      `,
			FileName: inNotifications,
			Files:    contexts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "crossContextModelType"}},
		},
		{
			Code: `
        import type { Occasion } from "../../../gifting/src/occasions/occasion"
        export const remindAbout = (occasion: Occasion): string => occasion.id
      `,
			FileName: inNestedNotifications,
			Files:    nestedContexts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "crossContextModelType"}},
		},
		{
			Code: `
        import type { Voucher } from "../../../gifting/index"
        export const codeOf = (voucher: Voucher): string => voucher.code
      `,
			FileName: inNotifications,
			Files:    contexts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "crossContextModelType"}},
		},
	})
}
