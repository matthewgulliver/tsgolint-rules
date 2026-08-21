package context_model_does_not_cross_the_boundary

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inNotifications = "packages/notifications/src/reminders/reminder.ts"

var contexts = map[string]string{
	// Gifting's model, declared in its internals.
	"packages/gifting/src/occasions/occasion.ts": `
    export type Occasion = { readonly id: string }
  `,
	// Its published contract re-exports the model rather than declaring one.
	// That is the leak §2.3 names, and it is what the path rule cannot see.
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
	// A barrel at the context root. dependency-cruiser's no-barrel-files
	// forbids importing one at all, so it cannot also be the surface this
	// rule lets a second context depend on.
	"packages/gifting/index.ts": `
    export type Voucher = { readonly code: string }
  `,
}

// A checkout whose own path holds a `packages/` segment, which is what this
// repository's fixtures became when they moved to `packages/tsgolint/fixtures/`.
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
		// This context's own model.
		{
			Code: `
        import type { Schedule } from "./schedule"
        export const nextReminder = (schedule: Schedule): Schedule => schedule
      `,
			FileName: inNotifications,
			Files:    contexts,
		},
		// A type the other context genuinely declares in its published
		// contract is the surface it agreed to keep stable.
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
		// The shared kernel belongs to no context.
		{
			Code: `
        import type { Money } from "../../../shared-kernel/src/money"
        export const owed = (amount: Money): Money => amount
      `,
			FileName: inNotifications,
			Files:    contexts,
		},
		// Not exported, so not part of this context's surface.
		{
			Code: `
        import type { Occasion } from "../../../gifting/src/occasions/occasion"
        const describe = (occasion: Occasion): string => occasion.id
      `,
			FileName: inNotifications,
			Files:    contexts,
		},
		// Its own context, reached by the other spelling of the same identity.
		{
			Code: `
        import type { Schedule } from "./schedule"
        export type Reminder = { readonly schedule: Schedule }
      `,
			FileName: inNotifications,
			Files:    contexts,
		},
		// No pattern identifies this path as a context, so there is nothing to
		// compare and the rule keeps its hands off.
		{
			Code: `
        import type { Occasion } from "../packages/gifting/src/occasions/occasion"
        export const describe = (occasion: Occasion): string => occasion.id
      `,
			FileName: "apps/web/src/page.ts",
			Files:    contexts,
		},
		// The innermost `packages/<name>` is the context, so a file's own
		// context is still its own when an outer directory is named that way.
		{
			Code: `
        import type { Schedule } from "./schedule"
        export const nextReminder = (schedule: Schedule): Schedule => schedule
      `,
			FileName: inNestedNotifications,
			Files:    nestedContexts,
		},
	}, []rule_tester.InvalidTestCase{
		// One context's model named directly in another's signature.
		{
			Code: `
        import type { Occasion } from "../../../gifting/src/occasions/occasion"
        export const remindAbout = (occasion: Occasion): string => occasion.id
      `,
			FileName: inNotifications,
			Files:    contexts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "crossContextModelType"}},
		},
		// The blind spot the path rule cannot close: the import is legal —
		// `packages/gifting/public` is the front door — and the type it hands
		// over is still declared in Gifting's internals.
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
		// A published type carries the model just as a signature does.
		{
			Code: `
        import type { Occasion } from "../../../gifting/public"
        export type Reminder = { readonly about: Occasion }
      `,
			FileName: inNotifications,
			Files:    contexts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "crossContextModelType"}},
		},
		// The return half.
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
		// The same crossing under an outer `packages/` segment. Read from the
		// left, every context below it answers `monorepo`, they compare equal,
		// and a real crossing goes unreported.
		{
			Code: `
        import type { Occasion } from "../../../gifting/src/occasions/occasion"
        export const remindAbout = (occasion: Occasion): string => occasion.id
      `,
			FileName: inNestedNotifications,
			Files:    nestedContexts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "crossContextModelType"}},
		},
		// A root barrel is not a published contract.
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
