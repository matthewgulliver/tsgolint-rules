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
	"node_modules/zod/index.d.ts": `
    export declare namespace z { class ZodError { constructor(issues: unknown[]) } }
  `,
}

func TestNoConstructedCollaborators(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoConstructedCollaboratorsRule, []rule_tester.ValidTestCase{
		{Code: `const now = new Date()`, FileName: inApplication},
		{Code: `const seen = new Map<string, number>()`, FileName: inApplication},
		{Code: `const at = new URL("https://example.test")`, FileName: inApplication},
		{
			Code: `
        class Occasion { constructor(readonly id: string) {} }
        const occasion = new Occasion("o-1")
      `,
			FileName: inApplication,
		},
		{
			Code:  `import { Pool } from "pg"; const pool = new Pool()`,
			Files: pg,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:     `import { Pool } from "pg"; const pool = new Pool()`,
			FileName: inApplication,
			Files:    pg,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerConstruction"}},
		},
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
