package domain_function_returns_an_answer

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inDomain = "packages/gifting/hexagon/domain/src/occasions/occasion.ts"

const occasion = `
  type Occasion = { readonly id: string; readonly total: number; readonly settled: boolean }
  declare const audit: (o: Occasion) => void
`

func TestDomainFunctionReturnsAnAnswer(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &DomainFunctionReturnsAnAnswerRule, []rule_tester.ValidTestCase{
		{
			Code:     occasion + `export const settle = (o: Occasion) => ({ ...o, settled: true })`,
			FileName: inDomain,
		},
		{
			Code: occasion + `
        export function settle(o: Occasion): { readonly success: false; readonly reason: "closed" } | Occasion {
          return o.settled ? { success: false, reason: "closed" } : { ...o, settled: true }
        }`,
			FileName: inDomain,
		},
		{
			Code:     occasion + `export const find = (all: ReadonlyArray<Occasion>, id: string) => all.find((o) => o.id === id)`,
			FileName: inDomain,
		},
		{
			Code:     occasion + `export const settle = (o: Occasion): void => { audit(o) }`,
			FileName: inDomain,
		},
		{
			Code:     occasion + `export const settleWith = (reason: string) => (o: Occasion) => ({ ...o, settled: true })`,
			FileName: inDomain,
		},
		{
			Code:     occasion + `const settle = (o: Occasion) => { audit(o) }`,
			FileName: inDomain,
		},
		{
			Code: occasion + `export const settle = (o: Occasion) => { audit(o) }`,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:     occasion + `export const settle = (o: Occasion) => { audit(o) }`,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "answerlessDomainFunction"}},
		},
		{
			Code: occasion + `
        export function record(o: { total: number }) {
          o.total += 1
        }`,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "answerlessDomainFunction"}},
		},
		{
			Code:     occasion + `export const settle = (o: Occasion) => { if (o.settled) return undefined; return }`,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "answerlessDomainFunction"}},
		},
		{
			Code:     occasion + `export const settleWith = (reason: string) => (o: Occasion) => { audit(o) }`,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "answerlessDomainFunction"}},
		},
	})
}
