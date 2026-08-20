package domain_state_is_deeply_readonly

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inDomain = "packages/gifting/hexagon/domain/src/occasions/occasion.ts"
const inKernel = "packages/shared-kernel/src/money.ts"

// An ambient module, the shape many `@types` packages take: nothing inside
// carries an `export` keyword, so only the declaring file says it is a vendor's.
var vendor = map[string]string{
	"node_modules/decimal.js/index.d.ts": `
    declare module "decimal.js" {
      class Decimal { d: number[]; toNumber(): number }
    }
  `,
}

func TestDomainStateIsDeeplyReadonly(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &DomainStateIsDeeplyReadonlyRule.Rule, []rule_tester.ValidTestCase{
		// Readonly all the way down through the alias.
		{
			Code: `
        type PledgeList = ReadonlyArray<{ readonly id: string }>
        type FundingState = { readonly pledges: PledgeList; readonly closed: boolean }
        export type Occasion = { readonly id: string; readonly funding: FundingState }
      `,
			FileName: inDomain,
		},
		// A member written mutable in the declaration itself is
		// `domain-state-is-readonly`'s to report, not this rule's.
		{
			Code:     `export type Occasion = { id: string; pledges: string[]; tags: Array<string> }`,
			FileName: inDomain,
		},
		// A dependency's members are not this repository's state.
		{
			Code: `
        import { Decimal } from "decimal.js"
        export type Money = { readonly amount: Decimal }
      `,
			FileName: inKernel,
			Files:    vendor,
		},
		// The standard library's own writable members are not the model's.
		{
			Code:     `export type Occasion = { readonly pattern: RegExp; readonly at: Date }`,
			FileName: inDomain,
		},
		// Behaviour is not state.
		{
			Code: `
        type Rules = { settle: (id: string) => string; readonly limit: number }
        export type Occasion = { readonly rules: Rules }
      `,
			FileName: inDomain,
		},
		// Recursion terminates on a self-referential type.
		{
			Code:     `export type Tree = { readonly value: number; readonly children: ReadonlyArray<Tree> }`,
			FileName: inDomain,
		},
		// Another exported declaration is judged when it is visited, not
		// through every type that names it.
		{
			Code: `
        export type FundingState = { readonly pledges: ReadonlyArray<string> }
        export type Occasion = { readonly funding: FundingState }
      `,
			FileName: inDomain,
		},
		// Not exported, so not the model's published state.
		{
			Code: `
        type PledgeList = string[]
        type Occasion = { readonly pledges: PledgeList }
      `,
			FileName: inDomain,
		},
	}, []rule_tester.InvalidTestCase{
		// A member reached through a named type is writable.
		{
			Code: `
        type FundingState = { pledges: ReadonlyArray<string>; readonly closed: boolean }
        export type Occasion = { readonly id: string; readonly funding: FundingState }
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableMemberResolved"}},
		},
		// An alias resolving to a mutable array.
		{
			Code: `
        type PledgeList = string[]
        export type Occasion = { readonly pledges: PledgeList }
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableCollectionResolved"}},
		},
		// The same alias, one level down.
		{
			Code: `
        type PledgeList = Array<string>
        type FundingState = { readonly pledges: PledgeList }
        export interface Occasion { readonly funding: FundingState }
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableCollectionResolved"}},
		},
		// A `Map` behind an alias, in a union with undefined.
		{
			Code: `
        type Index = Map<string, number>
        export type Occasion = { readonly index: Index | undefined }
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableCollectionResolved"}},
		},
		// The element type of a readonly collection is still state.
		{
			Code: `
        type Pledge = { amount: number }
        export type Occasion = { readonly pledges: ReadonlyArray<Pledge> }
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableMemberResolved"}},
		},
		// The shared kernel is judged too.
		{
			Code: `
        type Minor = { units: number }
        export type Money = { readonly minor: Minor; readonly currency: "AUD" }
      `,
			FileName: inKernel,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableMemberResolved"}},
		},
		// Reported once, at the declaration that owns the member: `Occasion`
		// does not repeat what `FundingState`'s own visit already said.
		{
			Code: `
        type PledgeList = string[]
        export type FundingState = { readonly pledges: PledgeList }
        export type Occasion = { readonly funding: FundingState }
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableCollectionResolved"}},
		},
		// Configured collection names.
		{
			Code: `
        type Bag = Set<string>
        export type Occasion = { readonly tags: Bag }
      `,
			FileName: inDomain,
			Options:  rule_tester.OptionsFromJSON[Options](`{"mutableCollectionTypeNames": ["Set"]}`),
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableCollectionResolved"}},
		},
	})
}
