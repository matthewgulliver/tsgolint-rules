package domain_signature_stays_in_the_domain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inDomain = "packages/gifting/hexagon/domain/src/pledges/pledge-contribution.ts"

var trees = map[string]string{
	"packages/gifting/hexagon/domain/src/occasions/occasion.ts": `
    export type Occasion = { readonly id: string }
    export type PledgeDecision =
      | { readonly success: true; readonly occasion: Occasion }
      | { readonly success: false; readonly reason: "funding-closed" }
  `,
	"packages/shared-kernel/src/money.ts": `
    export type Money = { readonly minorUnits: number; readonly currency: "GBP" }
  `,
	"packages/gifting/hexagon/application/src/ports/pledge-persistence.ts": `
    export interface PledgePersistence {
      readonly save: (id: string) => Promise<void>
    }
    export type PledgeResult = { readonly saved: boolean }
  `,
	"node_modules/stripe/index.d.ts": `
    export interface StripeCharge { readonly id: string }
    export type Infer<T> = T extends { _out: infer O } ? O : never
  `,
}

func TestDomainSignatureStaysInTheDomain(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &DomainSignatureStaysInTheDomainRule, []rule_tester.ValidTestCase{
		{
			Code: `
        import type { Occasion, PledgeDecision } from "../occasions/occasion"
        import type { Money } from "../../../../../shared-kernel/src/money"
        export const pledgeContribution = (
          occasion: Occasion,
          amount: Money,
        ): PledgeDecision => ({ success: false, reason: "funding-closed" })
      `,
			FileName: inDomain,
			Files:    trees,
		},
		{
			Code: `
        import type { Occasion } from "../occasions/occasion"
        export const settleAt = (
          occasion: Occasion,
          at: Date,
          ids: ReadonlyArray<string>,
        ): Occasion => occasion
      `,
			FileName: inDomain,
			Files:    trees,
		},
		{
			Code: `
        import type { Infer } from "stripe"
        type Parsed = Infer<{ _out: { readonly id: string } }>
        export const identify = (parsed: Parsed): Parsed => parsed
      `,
			FileName: inDomain,
			Files:    trees,
		},
		{
			Code: `
        import type { PledgePersistence } from "../../../application/src/ports/pledge-persistence"
        const save = (persistence: PledgePersistence): void => {}
      `,
			FileName: inDomain,
			Files:    trees,
		},
		{
			Code: `
        import type { PledgePersistence } from "./ports/pledge-persistence"
        export const pledgeContribution = (
          persistence: PledgePersistence,
        ): PledgeResult => ({ saved: true })
      `,
			Files: trees,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        import type { PledgePersistence } from "../../../application/src/ports/pledge-persistence"
        import type { Occasion } from "../occasions/occasion"
        export const pledgeContribution = (
          occasion: Occasion,
          persistence: PledgePersistence,
        ): Occasion => occasion
      `,
			FileName: inDomain,
			Files:    trees,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "foreignParameterType"}},
		},
		{
			Code: `
        import type { PledgeResult } from "../../../application/src/ports/pledge-persistence"
        import type { Occasion } from "../occasions/occasion"
        export const pledgeContribution = (occasion: Occasion): PledgeResult =>
          ({ saved: true })
      `,
			FileName: inDomain,
			Files:    trees,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "foreignReturnType"}},
		},
		{
			Code: `
        import type { StripeCharge } from "stripe"
        import type { Occasion } from "../occasions/occasion"
        export function settle(
          occasion: Occasion,
          charges: ReadonlyArray<StripeCharge>,
        ): Occasion {
          return occasion
        }
      `,
			FileName: inDomain,
			Files:    trees,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "foreignParameterType"}},
		},
	})
}
