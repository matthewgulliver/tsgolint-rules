package stored_state_switch_has_a_throwing_default

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inRepository = "packages/gifting/hexagon/adapters/driven/postgres/src/occasion-repository.ts"

var trees = map[string]string{
	"packages/gifting/hexagon/adapters/driven/postgres/src/rows.ts": `
    export type OccasionRow = {
      readonly id: string
      readonly state: "Open" | "Settled"
    }
  `,
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
