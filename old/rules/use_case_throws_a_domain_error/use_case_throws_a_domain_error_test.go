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
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &UseCaseThrowsADomainErrorRule.Rule, []rule_tester.ValidTestCase{
		// An error type the domain declares, which carries the business meaning.
		{
			Code: `
        class OccasionAlreadyClosed extends Error {}
        export const pledge = () => { throw new OccasionAlreadyClosed() }
      `,
			FileName: inApplication,
		},
		// A rethrow has nothing declared to judge.
		{
			Code: `
        export const pledge = (error: unknown) => { throw error }
      `,
			FileName: inApplication,
		},
		// An adapter translating a vendor failure is doing its job.
		{
			Code:     `export const query = () => { throw new Error("no connection") }`,
			FileName: "packages/gifting/adapters/driven/src/pledge-postgres.ts",
		},
		// A factory rejecting an impossible state. The hexagonal skill's own
		// worked example throws a bare `Error` from `createMoney` while
		// returning `exceeds-budget` as a result value in the same function:
		// an invariant violation is not a business outcome, and the domain has
		// no result union to put it in.
		{
			Code: `
        export const createMoney = (minorUnits: number) => {
          if (!Number.isSafeInteger(minorUnits)) throw new Error("Invalid money")
          return { minorUnits }
        }
      `,
			FileName: inDomain,
		},
	}, []rule_tester.InvalidTestCase{
		// A bare `Error` for a business outcome: every surface must invent copy.
		{
			Code:     `export const pledge = () => { throw new Error("occasion closed") }`,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "foreignThrow"}},
		},
		// A vendor's error escaping the inside, which the package-dependency arm
		// judges in the domain tree as well.
		{
			Code: `
        import { DatabaseError } from "pg"
        export const pledge = () => { throw new DatabaseError("conflict") }
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "foreignThrow"}},
		},
		// An alias hides the name, never the declaration.
		{
			Code: `
        const Failure = TypeError
        export const pledge = () => { throw new Failure("closed") }
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "foreignThrow"}},
		},
		// A repository that wants the stricter reading opts the domain tree in.
		{
			Code:     `export const pledge = () => { throw new Error("closed") }`,
			FileName: inDomain,
			Options: Options{StandardLibraryFiles: []string{"**/hexagon/domain/**"}},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "foreignThrow"}},
		},
		// A vendor error thrown in the domain still reports: no skill sanctions
		// a provider's failure type reaching the model.
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
