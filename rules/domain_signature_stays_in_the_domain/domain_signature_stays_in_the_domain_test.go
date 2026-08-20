package domain_signature_stays_in_the_domain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inDomain = "packages/gifting/hexagon/domain/src/pledges/pledge-contribution.ts"

// The three trees a domain signature may reach into, and the one it may not.
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
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &DomainSignatureStaysInTheDomainRule.Rule, []rule_tester.ValidTestCase{
		// domain-service.md's Perfect Example signature: domain values in,
		// a domain decision out.
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
		// The trap the ledger named: the standard library is not foreign.
		// aggregate-root.md holds `readonly pledgedAt: Date`.
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
		// A package dependency's *type alias* is excused, on the precedent
		// `no-provider-type-in-signature` records for `z.infer<typeof Schema>`.
		{
			Code: `
        import type { Infer } from "stripe"
        type Parsed = Infer<{ _out: { readonly id: string } }>
        export const identify = (parsed: Parsed): Parsed => parsed
      `,
			FileName: inDomain,
			Files:    trees,
		},
		// Not exported, so not part of the model's published vocabulary.
		{
			Code: `
        import type { PledgePersistence } from "../../../application/src/ports/pledge-persistence"
        const save = (persistence: PledgePersistence): void => {}
      `,
			FileName: inDomain,
			Files:    trees,
		},
	}, []rule_tester.InvalidTestCase{
		// The claim: an application port is not a domain value, and no shipped
		// rule sees it — `no-application-port-in-domain` judges declarations in
		// the domain tree, not references to one declared elsewhere.
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
		// The return half, which `no-provider-type-in-signature` never reads:
		// it visits parameters only.
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
		// A provider's nominal interface, reached through a generic argument.
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
