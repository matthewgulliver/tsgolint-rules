package no_double_library_in_domain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inDomainTest = "packages/gifting/hexagon/domain/src/occasions/occasion.test.ts"

var vitest = map[string]string{
	"node_modules/vitest/index.d.ts": `
    export declare const vi: {
      fn(): unknown
      mock(path: string): void
      spyOn(target: object, method: string): unknown
    }
  `,
}

func TestNoDoubleLibraryInDomainTest(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoDoubleLibraryInDomainTestRule, []rule_tester.ValidTestCase{
		{
			Code: `
        const vi = { fn: () => "a domain value" }
        vi.fn()
      `,
			FileName: inDomainTest,
		},
		{
			Code: `
        const makeVi = () => ({ fn: () => "a domain value" })
        makeVi().fn()
      `,
			FileName: inDomainTest,
		},
		{
			Code: `
        import { vi } from "vitest"
        vi.fn()
      `,
			Files: vitest,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        import { vi } from "vitest"
        vi.fn()
      `,
			FileName: inDomainTest,
			Files:    vitest,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "doubleLibrary"}},
		},
		{
			Code: `
        import { vi as doubles } from "vitest"
        doubles.spyOn({}, "save")
      `,
			FileName: inDomainTest,
			Files:    vitest,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "doubleLibrary"}},
		},
	})
}
