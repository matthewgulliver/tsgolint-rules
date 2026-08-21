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
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoOutsideDeclarationInTheHexagonRule.Rule, []rule_tester.ValidTestCase{
		// A port beside it and a schema library: both allowed, and the reason the
		// rule resolves the symbol rather than banning package imports.
		{
			Code: `
        import type { PledgePersistence } from "./ports/pledge-persistence"
        import { z } from "zod"
        export const createPledging = (persistence: PledgePersistence) => ({ persistence, schema: z.object({}) })
      `,
			FileName: useCase,
			Files:    tree,
		},
		// `adapters` in a package's own subpath is not this repository's adapter
		// tree, which a path rule reading the specifier could not tell.
		{
			Code: `
        import { adapt } from "some-lib/adapters"
        export const adapted = adapt({})
      `,
			FileName: useCase,
			Files:    tree,
		},
		// The barrel re-exports from inside as well as outside; this import
		// resolves to the inside half.
		{
			Code: `
        import { PledgePersistence } from "./index"
        export const createPledging = (persistence: PledgePersistence) => persistence
      `,
			FileName: useCase,
			Files:    tree,
		},
		// The domain importing its own neighbour.
		{
			Code: `
        import type { Money } from "./money"
        export const zero: Money = { minorUnits: 0 }
      `,
			FileName: domainFile,
			Files:    tree,
		},
		// Configured scope replaces the default, so the adapter tree stops being
		// outside and this import stops being anyone's business.
		{
			Code: `
        import { PostgresPledgeRepository } from "../../adapters/driven/postgres/src/repo"
        export const repository = new PostgresPledgeRepository()
      `,
			FileName: useCase,
			Files:    tree,
			Options:  map[string]any{"outsideFiles": []string{"**/apps/*/src/**"}},
		},
	}, []rule_tester.InvalidTestCase{
		// A use case reaching straight into the adapter tree.
		{
			Code: `
        import { PostgresPledgeRepository } from "../../adapters/driven/postgres/src/repo"
        export const repository = new PostgresPledgeRepository()
      `,
			FileName: useCase,
			Files:    tree,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "outsideDeclarationImported"}},
		},
		// The blind spot the path rules cannot close: the specifier is a legal
		// neighbour, and the symbol behind it was declared in an adapter.
		{
			Code: `
        import { PostgresPledgeRepository } from "./index"
        export const repository = new PostgresPledgeRepository()
      `,
			FileName: useCase,
			Files:    tree,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "outsideDeclarationImported"}},
		},
		// The same, through a workspace alias that only the resolver can follow.
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
		// A type-only import is still the domain knowing a stored row's shape.
		{
			Code: `
        import type { StoredRow } from "../../../adapters/driven/postgres/src/repo"
        export const toOccasion = (row: StoredRow) => row.id
      `,
			FileName: domainFile,
			Files:    tree,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "outsideDeclarationImported"}},
		},
		// A namespace import hides the name, never the declaration.
		{
			Code: `
        import * as postgres from "../../adapters/driven/postgres/src/repo"
        export const repository = new postgres.PostgresPledgeRepository()
      `,
			FileName: useCase,
			Files:    tree,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "outsideDeclarationImported"}},
		},
		// A side-effect import binds nothing, so there is no symbol to resolve —
		// the module itself is what crossed the line.
		{
			Code: `
        import "../../adapters/driven/postgres/src/repo"
      `,
			FileName: useCase,
			Files:    tree,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "outsideModuleImported"}},
		},
		// A re-export hands the outside on without ever binding it either.
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
