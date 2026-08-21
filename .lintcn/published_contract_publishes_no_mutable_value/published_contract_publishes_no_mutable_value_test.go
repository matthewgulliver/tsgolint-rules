package published_contract_publishes_no_mutable_value

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const contract = "packages/gifting/public.ts"

func TestPublishedContractPublishesNoMutableValue(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &PublishedContractPublishesNoMutableValueRule, []rule_tester.ValidTestCase{
		{
			Code:     `export const CURRENCIES = Object.freeze({ gbp: "GBP" })`,
			FileName: contract,
		},
		{
			Code: `
        type Limits = { readonly maxOpenPledges: number }
        const build = (): Limits => ({ maxOpenPledges: 5 })
        export const LIMITS = build()
      `,
			FileName: contract,
		},
		{
			Code:     `export const CODES: ReadonlyArray<string> = ["gbp"]`,
			FileName: contract,
		},
		{
			Code:     `export const BY_CODE: ReadonlyMap<string, string> = new Map()`,
			FileName: contract,
		},
		{
			Code:     `export const describeOccasion = (id: string): string => id`,
			FileName: contract,
		},
		{
			Code: `
        type Preferences = { readonly set: (key: string) => void }
        declare const preferences: Preferences
        export const PREFERENCES = preferences
      `,
			FileName: contract,
		},
		{
			Code: `
        const registry = new Map<string, string>()
        export const size = registry.size
      `,
			FileName: contract,
		},
		{
			Code: `export const CODES: Array<string> = ["gbp"]`,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        const buildRegistry = (): Map<string, string> => new Map()
        export const REGISTRY = buildRegistry()
      `,
			FileName: contract,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableContainerInContract"}},
		},
		{
			Code: `
        type Limits = { maxOpenPledges: number }
        const build = (): Limits => ({ maxOpenPledges: 5 })
        export const LIMITS = build()
      `,
			FileName: contract,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableMemberInContract"}},
		},
		{
			Code:     `export const CODES: Array<string> = ["gbp"]`,
			FileName: contract,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableContainerInContract"}},
		},
		{
			Code:     `export const SEEN: Set<string> = new Set()`,
			FileName: contract,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableContainerInContract"}},
		},
	})
}
