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
		// The next state, inferred rather than annotated — which is the case the
		// syntactic rule cannot judge.
		{
			Code:     occasion + `export const settle = (o: Occasion) => ({ ...o, settled: true })`,
			FileName: inDomain,
		},
		// An explicit refusal or the next state: a union, and an answer either way.
		{
			Code: occasion + `
        export function settle(o: Occasion): { readonly success: false; readonly reason: "closed" } | Occasion {
          return o.settled ? { success: false, reason: "closed" } : { ...o, settled: true }
        }`,
			FileName: inDomain,
		},
		// A lookup that found nothing still answers.
		{
			Code:     occasion + `export const find = (all: ReadonlyArray<Occasion>, id: string) => all.find((o) => o.id === id)`,
			FileName: inDomain,
		},
		// The written `: void` is `no-void-return-in-domain`'s to report.
		{
			Code:     occasion + `export const settle = (o: Occasion): void => { audit(o) }`,
			FileName: inDomain,
		},
		// A curried transition answers at the end of the chain.
		{
			Code:     occasion + `export const settleWith = (reason: string) => (o: Occasion) => ({ ...o, settled: true })`,
			FileName: inDomain,
		},
		// Not exported, so a local choice.
		{
			Code:     occasion + `const settle = (o: Occasion) => { audit(o) }`,
			FileName: inDomain,
		},
		// The gate: outside the domain tree (default file name `file.ts`) the
		// rule reports nothing, however answerless the function is.
		{
			Code: occasion + `export const settle = (o: Occasion) => { audit(o) }`,
		},
	}, []rule_tester.InvalidTestCase{
		// Inferred `void`: whatever it decided happened to something else, and
		// the recorded hole in `no-void-return-in-domain`, which reads only what
		// was written.
		{
			Code:     occasion + `export const settle = (o: Occasion) => { audit(o) }`,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "answerlessDomainFunction"}},
		},
		// The decision landed in the caller's object, and nothing came back.
		{
			Code: occasion + `
        export function record(o: { total: number }) {
          o.total += 1
        }`,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "answerlessDomainFunction"}},
		},
		// Inferred `undefined` on every path is no answer either.
		{
			Code:     occasion + `export const settle = (o: Occasion) => { if (o.settled) return undefined; return }`,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "answerlessDomainFunction"}},
		},
		// Curried, and answerless at the end of the chain.
		{
			Code:     occasion + `export const settleWith = (reason: string) => (o: Occasion) => { audit(o) }`,
			FileName: inDomain,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "answerlessDomainFunction"}},
		},
	})
}
