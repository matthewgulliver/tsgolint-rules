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
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoDoubleLibraryInDomainTestRule.Rule, []rule_tester.ValidTestCase{
		// A local `vi` is not the library, which only the declaration says.
		{
			Code: `
        const vi = { fn: () => "a domain value" }
        vi.fn()
      `,
			FileName: inDomainTest,
		},
	}, []rule_tester.InvalidTestCase{
		// The doc's boundary: a domain test calls the model directly.
		{
			Code: `
        import { vi } from "vitest"
        vi.fn()
      `,
			FileName: inDomainTest,
			Files:    vitest,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "doubleLibrary"}},
		},
		// An alias hides the name, never the declaration.
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
