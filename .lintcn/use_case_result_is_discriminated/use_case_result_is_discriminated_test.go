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
		// A discriminated union: one literal field says which member this is.
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
		// The curried use case the docs actually ship.
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
		// An absent answer is not an undiscriminated union.
		{
			Code: `
        type Occasion = { readonly id: string }
        export const findOccasion = async (): Promise<Occasion | null> => null
      `,
			FileName: inApplication,
		},
		// Not exported, so not the use case's published outcome.
		{
			Code: `
        const decide = (): { readonly a: string } | { readonly b: string } => ({ a: "" })
      `,
			FileName: inApplication,
		},
		// aggregate-root.md's own transition: the next state or an explicit
		// refusal, discriminated on a boolean literal.
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
		// Specific reason literals: the compiler can prove every failure handled.
		{
			Code: `
        type PledgeResult =
          | { readonly ok: true }
          | { readonly ok: false; readonly error: "exceeds-budget" | "funding-closed" }
        export const pledgeToOccasion = (): PledgeResult => ({ ok: true })
      `,
			FileName: inApplication,
		},
		// `reason: string` is left to configuration: two docs declare it themselves.
		{
			Code: `
        type PaymentResult =
          | { readonly ok: true }
          | { readonly ok: false; readonly reason: string }
        export const charge = (): PaymentResult => ({ ok: true })
      `,
			FileName: inApplication,
		},
		// RFC 9457's validation-error array, which problem-details.md endorses
		// by example, so `errors` is deliberately not seeded.
		{
			Code: `
        export const pledgeToOccasion = (): { readonly errors: readonly string[] } =>
          ({ errors: [] })
      `,
			FileName: inApplication,
		},
		// An array whose elements are modelled reasons: the caller can still
		// switch, once per element.
		{
			Code: `
        type Violation = "exceeds-budget" | "funding-closed"
        export const pledgeContribution = (): { readonly violations: readonly Violation[] } =>
          ({ violations: [] })
      `,
			FileName: inDomain,
		},
		// An array of identifiers, which are strings because they identify things.
		{
			Code: `
        export const pledgeToOccasion = (): { readonly pledgeIds: readonly string[] } =>
          ({ pledgeIds: [] })
      `,
			FileName: inApplication,
		},
		// A generic that carries a string without being a collection of them:
		// only arrays and tuples have elements this rule looks inside.
		{
			Code: `
        type Boxed<T> = { readonly inner: T }
        export const pledgeContribution = (): { readonly violations: Boxed<string> } =>
          ({ violations: { inner: "" } })
      `,
			FileName: inDomain,
		},
		// The gate: outside both hexagon trees (default file name `file.ts`)
		// the rule reports nothing, however undiscriminated the result is.
		{
			Code: `
        export const pledgeToOccasion = async (): Promise<
          { readonly occasionId: string } | { readonly reason: string }
        > => ({ occasionId: "o-1" })
      `,
		},
	}, []rule_tester.InvalidTestCase{
		// A union with nothing to switch on — the caller has to guess by probing
		// for a field.
		{
			Code: `
        export const pledgeToOccasion = async (): Promise<
          { readonly occasionId: string } | { readonly reason: string }
        > => ({ occasionId: "o-1" })
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undiscriminatedResult"}},
		},
		// A shared key whose type is widened to `string` discriminates nothing,
		// and reads identically to the good version in the source.
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
		// Named members still need a shared discriminant; these have none in common.
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
		// The same claim one tree over: a transition returning the next state
		// or a refusal, with nothing to narrow on.
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
		// A widening hidden behind an alias, which the written annotation never shows.
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
		// No annotation at all: the inferred `error` is whatever `message` was.
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
		// A single shape, not a union, still answers with a bare string.
		{
			Code: `
        export const pledgeToOccasion = (): { readonly ok: boolean; readonly error: string } =>
          ({ ok: true, error: "" })
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "genericFailureReason"}},
		},
		// A literal widened by a bare `string` sibling collapses to `string`.
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
		// `reason` judged only once configured.
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
		// The same claim one tree over.
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
		// Both arms on one function: undiscriminated, and a generic error.
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
		// An array of unconstrained strings is the same failure once per element,
		// and a `readonly string[]` is a reference type rather than a flagged one.
		{
			Code: `
        export const pledgeToOccasion = (): { readonly ok: boolean; readonly error: readonly string[] } =>
          ({ ok: true, error: [] })
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "genericFailureReason"}},
		},
		// The refusal shape a repository following this tree actually shipped.
		{
			Code: `
        export const pledgeContribution = (): { readonly outcome: "refused"; readonly violations: readonly string[] } =>
          ({ outcome: "refused", violations: [] })
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "genericFailureReason"}},
		},
		// The other two spellings the same shape takes.
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
