package no_outside_declaration_in_the_hexagon

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const useCase = "packages/gifting/hexagon/application/src/pledge-to-occasion.ts"
const domainFile = "packages/gifting/hexagon/domain/src/occasions/occasion.ts"

const repository = `
  export class PostgresPledgeRepository { save(): Promise<void> { return Promise.resolve() } }
  export type StoredRow = { readonly id: string }
`

var tree = map[string]string{
	"packages/gifting/hexagon/adapters/driven/postgres/src/repo.ts":  repository,
	"packages/gifting/hexagon/adapters/driven/postgres/src/index.ts": `export * from "./repo"`,
	"packages/gifting/hexagon/application/src/index.ts": `
    export { PostgresPledgeRepository } from "../../adapters/driven/postgres/src/repo"
    export type { PledgePersistence } from "./ports/pledge-persistence"
  `,
	"packages/gifting/hexagon/application/src/ports/pledge-persistence.ts": `
    export interface PledgePersistence { save(id: string): Promise<void> }
  `,
	"packages/gifting/hexagon/domain/src/occasions/money.ts": `export type Money = { readonly minorUnits: number }`,
	"node_modules/zod/index.d.ts":                            `export declare const z: { object(shape: object): unknown }`,
	"node_modules/some-lib/adapters/index.d.ts":              `export declare const adapt: (value: unknown) => unknown`,
	"tsconfig.paths.json": `{
    "extends": "./tsconfig.minimal.json",
    "compilerOptions": {
      "paths": { "@gifting/adapters-postgres": ["./packages/gifting/hexagon/adapters/driven/postgres/src/index.ts"] }
    }
  }`,
}

func TestNoOutsideDeclarationInTheHexagon(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoOutsideDeclarationInTheHexagonRule, []rule_tester.ValidTestCase{
		{
			Code: `
        import type { PledgePersistence } from "./ports/pledge-persistence"
        import { z } from "zod"
        export const createPledging = (persistence: PledgePersistence) => ({ persistence, schema: z.object({}) })
      `,
			FileName: useCase,
			Files:    tree,
		},
		{
			Code: `
        import { adapt } from "some-lib/adapters"
        export const adapted = adapt({})
      `,
			FileName: useCase,
			Files:    tree,
		},
		{
			Code: `
        import { PledgePersistence } from "./index"
        export const createPledging = (persistence: PledgePersistence) => persistence
      `,
			FileName: useCase,
			Files:    tree,
		},
		{
			Code: `
        import type { Money } from "./money"
        export const zero: Money = { minorUnits: 0 }
      `,
			FileName: domainFile,
			Files:    tree,
		},
		{
			Code: `
        import { PostgresPledgeRepository } from "../../adapters/driven/postgres/src/repo"
        export const repository = new PostgresPledgeRepository()
      `,
			FileName: useCase,
			Files:    tree,
			Options:  map[string]any{"outsideFiles": []string{"**/apps/*/src/**"}},
		},
		{
			Code: `
        import { PostgresPledgeRepository } from "./repo"
        export const repository = new PostgresPledgeRepository()
      `,
			Files: tree,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        import { PostgresPledgeRepository } from "../../adapters/driven/postgres/src/repo"
        export const repository = new PostgresPledgeRepository()
      `,
			FileName: useCase,
			Files:    tree,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "outsideDeclarationImported"}},
		},
		{
			Code: `
        import { PostgresPledgeRepository } from "./index"
        export const repository = new PostgresPledgeRepository()
      `,
			FileName: useCase,
			Files:    tree,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "outsideDeclarationImported"}},
		},
		{
			Code: `
        import { PostgresPledgeRepository } from "@gifting/adapters-postgres"
        export const repository = new PostgresPledgeRepository()
      `,
			FileName: useCase,
			Files:    tree,
			TSConfig: "tsconfig.paths.json",
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "outsideDeclarationImported"}},
		},
		{
			Code: `
        import type { StoredRow } from "../../../adapters/driven/postgres/src/repo"
        export const toOccasion = (row: StoredRow) => row.id
      `,
			FileName: domainFile,
			Files:    tree,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "outsideDeclarationImported"}},
		},
		{
			Code: `
        import * as postgres from "../../adapters/driven/postgres/src/repo"
        export const repository = new postgres.PostgresPledgeRepository()
      `,
			FileName: useCase,
			Files:    tree,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "outsideDeclarationImported"}},
		},
		{
			Code: `
        import "../../adapters/driven/postgres/src/repo"
      `,
			FileName: useCase,
			Files:    tree,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "outsideModuleImported"}},
		},
		{
			Code: `
        export { PostgresPledgeRepository } from "../../adapters/driven/postgres/src/repo"
      `,
			FileName: useCase,
			Files:    tree,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "outsideModuleImported"}},
		},
	})
}
