package published_contract_publishes_no_mutable_value

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const contract = "packages/gifting/public.ts"

func TestPublishedContractPublishesNoMutableValue(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &PublishedContractPublishesNoMutableValueRule.Rule, []rule_tester.ValidTestCase{
		// The blind spot the syntactic rule records: `Object.freeze({…})` and
		// `buildRegistry()` are the same syntax, and opposite contracts.
		{
			Code:     `export const CURRENCIES = Object.freeze({ gbp: "GBP" })`,
			FileName: contract,
		},
		// The blind spot the syntactic rule cannot close: a built value, whose
		// immutability is only visible once the call's return type resolves.
		{
			Code: `
        type Limits = { readonly maxOpenPledges: number }
        const build = (): Limits => ({ maxOpenPledges: 5 })
        export const LIMITS = build()
      `,
			FileName: contract,
		},
		// Readonly containers, which publish nothing a consumer can change.
		{
			Code:     `export const CODES: ReadonlyArray<string> = ["gbp"]`,
			FileName: contract,
		},
		// `new Map()` is the same syntax as a mutable one; only the annotation
		// says which this is.
		{
			Code:     `export const BY_CODE: ReadonlyMap<string, string> = new Map()`,
			FileName: contract,
		},
		// A published operation is behaviour, not shared state.
		{
			Code:     `export const describeOccasion = (id: string): string => id`,
			FileName: contract,
		},
		// This contract's own `set` is a command, not a container's writer.
		{
			Code: `
        type Preferences = { readonly set: (key: string) => void }
        declare const preferences: Preferences
        export const PREFERENCES = preferences
      `,
			FileName: contract,
		},
		// Not exported.
		{
			Code: `
        const registry = new Map<string, string>()
        export const size = registry.size
      `,
			FileName: contract,
		},
	}, []rule_tester.InvalidTestCase{
		// The blind spot, the other way round: same syntax as the frozen one.
		{
			Code: `
        const buildRegistry = (): Map<string, string> => new Map()
        export const REGISTRY = buildRegistry()
      `,
			FileName: contract,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableContainerInContract"}},
		},
		// The same built value, one `readonly` short.
		{
			Code: `
        type Limits = { maxOpenPledges: number }
        const build = (): Limits => ({ maxOpenPledges: 5 })
        export const LIMITS = build()
      `,
			FileName: contract,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableMemberInContract"}},
		},
		// A published array a consumer can push to.
		{
			Code:     `export const CODES: Array<string> = ["gbp"]`,
			FileName: contract,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableContainerInContract"}},
		},
		// Shared mutable state behind a published name.
		{
			Code:     `export const SEEN: Set<string> = new Set()`,
			FileName: contract,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "mutableContainerInContract"}},
		},
	})
}
