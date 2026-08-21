package no_double_library_in_use_case

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inUseCaseTest = "packages/gifting/hexagon/application/src/pledge-to-occasion.test.ts"

var vitest = map[string]string{
	"node_modules/vitest/index.d.ts": `
    export declare const vi: {
      fn(): unknown
      mock(path: string): void
      spyOn(target: object, method: string): unknown
    }
  `,
}

func TestNoDoubleLibraryInUseCaseTest(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoDoubleLibraryInUseCaseTestRule, []rule_tester.ValidTestCase{
		// The doc's own answer: a stateful in-memory fake, hand-written.
		{
			Code: `
        type PledgePersistence = { save(occasion: { id: string }): Promise<void> }
        const createFakePledgePersistence = (): PledgePersistence & { readonly saved: { id: string }[] } => {
          const saved: { id: string }[] = []
          return { saved, save: async (occasion) => { saved.push(occasion) } }
        }
        const persistence = createFakePledgePersistence()
      `,
			FileName: inUseCaseTest,
		},
		// A local `vi` is not the library, which only the declaration says.
		{
			Code: `
        const vi = { fn: () => "a hand-written double" }
        vi.fn()
      `,
			FileName: inUseCaseTest,
		},
		// The gate: a file outside the application test tree (default file name
		// `file.ts`) reports nothing, however much double-library scaffolding it
		// carries.
		{
			Code: `
        import { vi } from "vitest"
        const persistence = { save: vi.fn() }
      `,
			Files: vitest,
		},
	}, []rule_tester.InvalidTestCase{
		// A mock standing in for the port the use-case test exists to exercise.
		{
			Code: `
        import { vi } from "vitest"
        const persistence = { save: vi.fn() }
      `,
			FileName: inUseCaseTest,
			Files:    vitest,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "doubleLibrary"}},
		},
		// An alias hides the name, never the declaration.
		{
			Code: `
        import { vi as doubles } from "vitest"
        doubles.spyOn({}, "save")
      `,
			FileName: inUseCaseTest,
			Files:    vitest,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "doubleLibrary"}},
		},
		// Module mocking a port, in a `.tsx` use-case test.
		{
			Code: `
        import { vi } from "vitest"
        vi.mock("./ports/pledge-persistence")
      `,
			FileName: "packages/gifting/hexagon/application/src/occasions/list-dashboard.test.tsx",
			Files:    vitest,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "doubleLibrary"}},
		},
	})
}
