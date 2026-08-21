package domain_state_is_deeply_readonly

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inDomain = "packages/gifting/hexagon/domain/src/occasions/occasion.ts"
const inKernel = "packages/shared-kernel/src/money.ts"

var vendor = map[string]string{
	"node_modules/decimal.js/index.d.ts": `
    declare module "decimal.js" {
      class Decimal { d: number[]; toNumber(): number }
    }
  `,
}

func TestDomainStateIsDeeplyReadonly(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &DomainStateIsDeeplyReadonlyRule, []rule_tester.ValidTestCase{
		{
			Code: `
        type PledgeList = ReadonlyArray<{ readonly id: string }>
        type FundingState = { readonly pledges: PledgeList; readonly closed: boolean }
        export type Occasion = { readonly id: string; readonly funding: FundingState }
      `,
			FileName: inDomain,
		},
		{
			Code:     `export type Occasion = { id: string; pledges: string[]; tags: Array<string> }`,
			FileName: inDomain,
		},
		{
			Code:     `export type Occasion = { readonly tags: readonly string[] }`,
			FileName: inDomain,
		},
		{
			Code: `
        import { Decimal } from "decimal.js"
        export type Money = { readonly amount: Decimal }
      `,
			FileName: inKernel,
			Files:    vendor,
		},
		{
			Code:     `export type Occasion = { readonly pattern: RegExp; readonly at: Date }`,
			FileName: inDomain,
		},
		{
			Code: `
        type Rules = { settle: (id: string) => string; readonly limit: number }
        export type Occasion = { readonly rules: Rules }
      `,
			FileName: inDomain,
		},
		{
			Code:     `export type Tree = { readonly value: number; readonly children: ReadonlyArray<Tree> }`,
			FileName: inDomain,
		},
		{
			Code: `
        export type FundingState = { readonly pledges: ReadonlyArray<string> }
        export type Occasion = { readonly funding: FundingState }
      `,
			FileName: inDomain,
		},
		{
			Code: `
        type PledgeList = string[]
        type Occasion = { readonly pledges: PledgeList }
      `,
			FileName: inDomain,
		},
		{
			Code: `
        type FundingState = { pledges: ReadonlyArray<string> }
        export type Occasion = { readonly funding: FundingState }
      `,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        type FundingState = { pledges: ReadonlyArray<string>; readonly closed: boolean }
        export type Occasion = { readonly id: string; readonly funding: FundingState }
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableMemberResolved"}},
		},
		{
			Code: `
        type PledgeList = string[]
        export type Occasion = { readonly pledges: PledgeList }
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableCollectionResolved"}},
		},
		{
			Code: `
        type PledgeList = Array<string>
        type FundingState = { readonly pledges: PledgeList }
        export interface Occasion { readonly funding: FundingState }
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableCollectionResolved"}},
		},
		{
			Code: `
        type Index = Map<string, number>
        export type Occasion = { readonly index: Index | undefined }
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableCollectionResolved"}},
		},
		{
			Code: `
        type Pledge = { amount: number }
        export type Occasion = { readonly pledges: ReadonlyArray<Pledge> }
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableMemberResolved"}},
		},
		{
			Code: `
        type Minor = { units: number }
        export type Money = { readonly minor: Minor; readonly currency: "AUD" }
      `,
			FileName: inKernel,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableMemberResolved"}},
		},
		{
			Code: `
        type PledgeList = string[]
        export type FundingState = { readonly pledges: PledgeList }
        export type Occasion = { readonly funding: FundingState }
      `,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableCollectionResolved"}},
		},
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
