package use_case_result_is_discriminated

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inApplication = "packages/gifting/hexagon/application/src/pledge-to-occasion.ts"

const inDomain = "packages/gifting/hexagon/domain/src/occasions/occasion.ts"

func TestUseCaseResultIsDiscriminated(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &UseCaseResultIsDiscriminatedRule, []rule_tester.ValidTestCase{
		{
			Code: `
        type PledgeResult =
          | { readonly success: true; readonly occasionId: string }
          | { readonly success: false; readonly reason: "occasion-closed" }
        export const pledgeToOccasion = async (): Promise<PledgeResult> =>
          ({ success: true, occasionId: "o-1" })
      `,
			FileName: inApplication,
		},
		{
			Code: `
        type PledgeResult =
          | { readonly outcome: "saved"; readonly occasionId: string }
          | { readonly outcome: "conflict" }
        export const createPledgingToOccasions =
          (save: (id: string) => Promise<void>) =>
          async (id: string): Promise<PledgeResult> => {
            await save(id)
            return { outcome: "saved", occasionId: id }
          }
      `,
			FileName: inApplication,
		},
		{
			Code: `
        type Occasion = { readonly id: string }
        export const findOccasion = async (): Promise<Occasion | null> => null
      `,
			FileName: inApplication,
		},
		{
			Code: `
        const decide = (): { readonly a: string } | { readonly b: string } => ({ a: "" })
      `,
			FileName: inApplication,
		},
		{
			Code: `
        type Occasion = { readonly id: string }
        type PledgeCommandResult =
          | { readonly success: true; readonly occasion: Occasion }
          | { readonly success: false; readonly reason: "unknown-pledge" }
        export const settlePledge = (occasion: Occasion): PledgeCommandResult =>
          ({ success: true, occasion })
      `,
			FileName: inDomain,
		},
		{
			Code: `
        type PledgeResult =
          | { readonly ok: true }
          | { readonly ok: false; readonly error: "exceeds-budget" | "funding-closed" }
        export const pledgeToOccasion = (): PledgeResult => ({ ok: true })
      `,
			FileName: inApplication,
		},
		{
			Code: `
        type PaymentResult =
          | { readonly ok: true }
          | { readonly ok: false; readonly reason: string }
        export const charge = (): PaymentResult => ({ ok: true })
      `,
			FileName: inApplication,
		},
		{
			Code: `
        export const pledgeToOccasion = (): { readonly errors: readonly string[] } =>
          ({ errors: [] })
      `,
			FileName: inApplication,
		},
		{
			Code: `
        type Violation = "exceeds-budget" | "funding-closed"
        export const pledgeContribution = (): { readonly violations: readonly Violation[] } =>
          ({ violations: [] })
      `,
			FileName: inDomain,
		},
		{
			Code: `
        export const pledgeToOccasion = (): { readonly pledgeIds: readonly string[] } =>
          ({ pledgeIds: [] })
      `,
			FileName: inApplication,
		},
		{
			Code: `
        type Boxed<T> = { readonly inner: T }
        export const pledgeContribution = (): { readonly violations: Boxed<string> } =>
          ({ violations: { inner: "" } })
      `,
			FileName: inDomain,
		},
		{
			Code: `
        export const pledgeToOccasion = async (): Promise<
          { readonly occasionId: string } | { readonly reason: string }
        > => ({ occasionId: "o-1" })
      `,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        export const pledgeToOccasion = async (): Promise<
          { readonly occasionId: string } | { readonly reason: string }
        > => ({ occasionId: "o-1" })
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undiscriminatedResult"}},
		},
		{
			Code: `
        type Saved = { readonly outcome: string; readonly occasionId: string }
        type Conflict = { readonly outcome: string }
        export function pledgeToOccasion(): Saved | Conflict {
          return { outcome: "saved", occasionId: "o-1" }
        }
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undiscriminatedResult"}},
		},
		{
			Code: `
        type Saved = { readonly outcome: "saved" }
        type Conflict = { readonly reason: "conflict" }
        export const createPledgingToOccasions =
          () => async (): Promise<Saved | Conflict> => ({ outcome: "saved" })
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undiscriminatedResult"}},
		},
		{
			Code: `
        type Occasion = { readonly id: string }
        export const settlePledge = (
          occasion: Occasion,
        ): { readonly occasion: Occasion } | { readonly reason: string } =>
          ({ occasion })
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undiscriminatedResult"}},
		},
		{
			Code: `
        type FailureText = string
        type PledgeResult =
          | { readonly ok: true }
          | { readonly ok: false; readonly error: FailureText }
        export const pledgeToOccasion = (): PledgeResult => ({ ok: true })
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "genericFailureReason"}},
		},
		{
			Code: `
        export const pledgeToOccasion = (message: string) => {
          if (message.length > 3) return { ok: false as const, error: message }
          return { ok: true as const }
        }
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "genericFailureReason"}},
		},
		{
			Code: `
        export const pledgeToOccasion = (): { readonly ok: boolean; readonly error: string } =>
          ({ ok: true, error: "" })
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "genericFailureReason"}},
		},
		{
			Code: `
        type PledgeResult =
          | { readonly ok: true }
          | { readonly ok: false; readonly error: "exceeds-budget" | string }
        export const pledgeToOccasion = (): PledgeResult => ({ ok: true })
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "genericFailureReason"}},
		},
		{
			Code: `
        type PledgeResult =
          | { readonly ok: true }
          | { readonly ok: false; readonly reason: string }
        export const pledgeToOccasion = (): PledgeResult => ({ ok: true })
      `,
			FileName: inApplication,
			Options:  Options{FailureReasonMemberPatterns: []string{"^reason$"}},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "genericFailureReason"}},
		},
		{
			Code: `
        type Occasion = { readonly id: string }
        type PledgeDecision =
          | { readonly success: true; readonly occasion: Occasion }
          | { readonly success: false; readonly error: string }
        export const pledgeContribution = (occasion: Occasion): PledgeDecision =>
          ({ success: true, occasion })
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "genericFailureReason"}},
		},
		{
			Code: `
        export const pledgeToOccasion = (): { readonly occasionId: string } | { readonly error: string } =>
          ({ occasionId: "o-1" })
      `,
			FileName: inApplication,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "genericFailureReason"},
				{MessageId: "undiscriminatedResult"},
			},
		},
		{
			Code: `
        export const pledgeToOccasion = (): { readonly ok: boolean; readonly error: readonly string[] } =>
          ({ ok: true, error: [] })
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "genericFailureReason"}},
		},
		{
			Code: `
        export const pledgeContribution = (): { readonly outcome: "refused"; readonly violations: readonly string[] } =>
          ({ outcome: "refused", violations: [] })
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "genericFailureReason"}},
		},
		{
			Code: `
        export const pledgeContribution = (): { readonly faults: string[]; readonly problems: ReadonlyArray<string> } =>
          ({ faults: [], problems: [] })
      `,
			FileName: inDomain,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "genericFailureReason"},
				{MessageId: "genericFailureReason"},
			},
		},
	})
}
