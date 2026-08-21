package use_case_throws_a_domain_error

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inApplication = "packages/gifting/hexagon/application/src/pledge-to-occasion.ts"
const inDomain = "packages/gifting/hexagon/domain/src/occasion.ts"

var vendor = map[string]string{
	"node_modules/pg/index.d.ts": `
    export declare class DatabaseError extends Error { constructor(message: string); }
  `,
}

func TestUseCaseThrowsADomainError(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &UseCaseThrowsADomainErrorRule, []rule_tester.ValidTestCase{
		{
			Code: `
        class OccasionAlreadyClosed extends Error {}
        export const pledge = () => { throw new OccasionAlreadyClosed() }
      `,
			FileName: inApplication,
		},
		{
			Code: `
        export const pledge = (error: unknown) => { throw error }
      `,
			FileName: inApplication,
		},
		{
			Code:     `export const query = () => { throw new Error("no connection") }`,
			FileName: "packages/gifting/adapters/driven/src/pledge-postgres.ts",
		},
		{
			Code: `
        export const createMoney = (minorUnits: number) => {
          if (!Number.isSafeInteger(minorUnits)) throw new Error("Invalid money")
          return { minorUnits }
        }
      `,
			FileName: inDomain,
		},
		{
			Code: `export const pledge = () => { throw new Error("closed") }`,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:     `export const pledge = () => { throw new Error("occasion closed") }`,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "foreignThrow"}},
		},
		{
			Code: `
        import { DatabaseError } from "pg"
        export const pledge = () => { throw new DatabaseError("conflict") }
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "foreignThrow"}},
		},
		{
			Code: `
        const Failure = TypeError
        export const pledge = () => { throw new Failure("closed") }
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "foreignThrow"}},
		},
		{
			Code:     `export const pledge = () => { throw new Error("closed") }`,
			FileName: inDomain,
			Options:  Options{StandardLibraryFiles: []string{"**/hexagon/domain/**"}},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "foreignThrow"}},
		},
		{
			Code: `
        import { DatabaseError } from "pg"
        export const pledge = () => { throw new DatabaseError("conflict") }
      `,
			FileName: inDomain,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "foreignThrow"}},
		},
	})
}
