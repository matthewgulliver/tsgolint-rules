package stored_state_switch_has_a_throwing_default

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inRepository = "packages/gifting/hexagon/adapters/driven/postgres/src/occasion-repository.ts"

var trees = map[string]string{
	// The row type, declared where `no-row-type-in-domain` keeps it.
	"packages/gifting/hexagon/adapters/driven/postgres/src/rows.ts": `
    export type OccasionRow = {
      readonly id: string
      readonly state: "Open" | "Settled"
    }
  `,
	// A command the model itself declares. Same syntax at the switch, opposite
	// provenance: nothing read it out of a database.
	"packages/gifting/hexagon/domain/src/occasions/occasion.ts": `
    export type SettleCommand = {
      readonly id: string
      readonly state: "Open" | "Settled"
    }
  `,
}

func TestStoredStateSwitchHasAThrowingDefault(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &StoredStateSwitchHasAThrowingDefaultRule, []rule_tester.ValidTestCase{
		// aggregate-reconstitution.md's own mapper, `never` binding and all.
		{
			Code: `
        import type { OccasionRow } from "./rows"
        export const toOccasion = (row: OccasionRow): string => {
          switch (row.state) {
            case "Open":
              return "open"
            case "Settled":
              return "settled"
            default: {
              const unknownState: never = row.state
              throw new Error("unknown state " + String(unknownState))
            }
          }
        }
      `,
			FileName: inRepository,
			Files:    trees,
		},
		// A domain command's discriminant is not a stored one, and upstream
		// `switch-exhaustiveness-check` is the right rule for it. A throwing
		// default here would be the noise this rule is careful not to demand.
		{
			Code: `
        import type { SettleCommand } from "../../../../domain/src/occasions/occasion"
        export const describe = (command: SettleCommand): string => {
          switch (command.state) {
            case "Open":
              return "open"
            case "Settled":
              return "settled"
          }
        }
      `,
			FileName: inRepository,
			Files:    trees,
		},
		// Not a property access, so nothing says where the value came from.
		{
			Code: `
        export const describe = (state: "Open" | "Settled"): string => {
          switch (state) {
            case "Open":
              return "open"
            default:
              return "settled"
          }
        }
      `,
			FileName: inRepository,
			Files:    trees,
		},
		// Outside the judged tree.
		{
			Code: `
        import type { OccasionRow } from "../adapters/driven/postgres/src/rows"
        export const describe = (row: OccasionRow): string => {
          switch (row.state) {
            case "Open":
              return "open"
            case "Settled":
              return "settled"
          }
        }
      `,
			FileName: "packages/gifting/hexagon/application/src/pledge-to-occasion.ts",
			Files:    trees,
		},
		// The gate: outside the driven-adapter tree (default file name
		// `file.ts`) the rule reports nothing, however stored the discriminant
		// is — here the very switch the first invalid case reports.
		{
			Code: `
        import type { OccasionRow } from "./packages/gifting/hexagon/adapters/driven/postgres/src/rows"
        export const toOccasion = (row: OccasionRow): string => {
          switch (row.state) {
            case "Open":
              return "open"
            case "Settled":
              return "settled"
          }
          return "unreachable"
        }
      `,
			Files: trees,
		},
	}, []rule_tester.InvalidTestCase{
		// Exhaustive over the row type, and silent about the database. This is
		// exactly what upstream `switch-exhaustiveness-check` is satisfied by.
		{
			Code: `
        import type { OccasionRow } from "./rows"
        export const toOccasion = (row: OccasionRow): string => {
          switch (row.state) {
            case "Open":
              return "open"
            case "Settled":
              return "settled"
          }
          return "unreachable"
        }
      `,
			FileName: inRepository,
			Files:    trees,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "storedStateSwitchWithoutDefault"}},
		},
		// A `default` that returns instead of throwing turns a corrupt row into
		// a value the model accepts.
		{
			Code: `
        import type { OccasionRow } from "./rows"
        export const toOccasion = (row: OccasionRow): string => {
          switch (row.state) {
            case "Open":
              return "open"
            default:
              return "settled"
          }
        }
      `,
			FileName: inRepository,
			Files:    trees,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "storedStateDefaultDoesNotThrow"}},
		},
		// A `throw` inside a nested function is not this clause's.
		{
			Code: `
        import type { OccasionRow } from "./rows"
        export const toOccasion = (row: OccasionRow): string => {
          switch (row.state) {
            case "Open":
              return "open"
            default: {
              const explode = () => {
                throw new Error("nope")
              }
              return String(explode)
            }
          }
        }
      `,
			FileName: inRepository,
			Files:    trees,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "storedStateDefaultDoesNotThrow"}},
		},
	})
}
