package no_constructed_collaborators

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inApplication = "packages/gifting/hexagon/application/src/pledge-to-occasion.ts"

var pg = map[string]string{
	"node_modules/pg/index.d.ts": `
    export declare class Pool {
      query(text: string): Promise<unknown>;
    }
  `,
	// A package whose classes are reached through a namespace, so `new` is
	// written against a property access rather than a bare name.
	"node_modules/zod/index.d.ts": `
    export declare namespace z { class ZodError { constructor(issues: unknown[]) } }
  `,
}

func TestNoConstructedCollaborators(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoConstructedCollaboratorsRule, []rule_tester.ValidTestCase{
		// The values that made the syntactic version unshippable.
		{Code: `const now = new Date()`, FileName: inApplication},
		// Standard-library values, which is the line `new` alone could not draw and
		// the reason the syntactic version of this rule was refused.
		{Code: `const seen = new Map<string, number>()`, FileName: inApplication},
		{Code: `const at = new URL("https://example.test")`, FileName: inApplication},
		// A type this repository declares is not a provider.
		{
			Code: `
        class Occasion { constructor(readonly id: string) {} }
        const occasion = new Occasion("o-1")
      `,
			FileName: inApplication,
		},
		// The gate: outside the inside trees (default file name `file.ts`) the
		// rule reports nothing, however provider-shaped the construction is.
		{
			Code:     `import { Pool } from "pg"; const pool = new Pool()`,
			Files:    pg,
		},
	}, []rule_tester.InvalidTestCase{
		// A dependency constructed where it should have been injected.
		{
			Code:     `import { Pool } from "pg"; const pool = new Pool()`,
			FileName: inApplication,
			Files:    pg,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerConstruction"}},
		},
		// An alias hides the name, never the declaration.
		{
			Code: `
        import { Pool } from "pg"
        const Connection = Pool
        const pool = new Connection()
      `,
			FileName: inApplication,
			Files:    pg,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerConstruction"}},
		},
		// A class reached through its namespace. Asking a property access for
		// its text as if it were an identifier took the whole run down, so
		// every file after it went unjudged.
		{
			Code: `
        import { z } from "zod"
        const thrown = new z.ZodError([])
      `,
			FileName: inApplication,
			Files:    pg,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerConstruction"}},
		},
	})
}
