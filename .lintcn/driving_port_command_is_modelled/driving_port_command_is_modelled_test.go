package driving_port_command_is_modelled

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inPorts = "packages/gifting/hexagon/application/src/ports/for-pledging-to-occasions.ts"

func TestDrivingPortCommandIsModelled(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &DrivingPortCommandIsModelledRule, []rule_tester.ValidTestCase{
		{
			Code: `
        type PledgeCommand = { readonly occasionId: string; readonly amount: number }
        export interface ForPledgingToOccasions {
          readonly pledgeToOccasion: (command: PledgeCommand) => Promise<void>
        }
      `,
			FileName: inPorts,
		},
		{
			Code: `
        type OccasionId = string & { readonly __brand: "OccasionId" }
        export interface ForCancellingAnOccasion {
          readonly cancel: (id: OccasionId) => Promise<void>
        }
      `,
			FileName: inPorts,
		},
		{
			Code: `
        export interface ForRatingAPledge {
          readonly rate: (rating: "up" | "down") => Promise<void>
        }
      `,
			FileName: inPorts,
		},
		{
			Code: `
        export interface PledgePersistence {
          readonly findOccasionById: (id: string) => Promise<string | null>
        }
      `,
			FileName: inPorts,
		},
		{
			Code: `
        export interface ForPledgingToOccasions {
          readonly pledgeToOccasion: (occasionId: string) => Promise<void>
        }
      `,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        export interface ForPledgingToOccasions {
          readonly pledgeToOccasion: (occasionId: string) => Promise<void>
        }
      `,
			FileName: inPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unmodelledCommand"}},
		},
		{
			Code: `
        type OccasionId = string
        export interface ForPledgingToOccasions {
          readonly pledgeToOccasion: (occasionId: OccasionId) => Promise<void>
        }
      `,
			FileName: inPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unmodelledCommand"}},
		},
		{
			Code: `
        export interface DrivesPledging {
          readonly pledge: (amount: number) => Promise<void>
        }
      `,
			FileName: inPorts,
			Options:  Options{DrivingPortPatterns: []string{"^Drives"}},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unmodelledCommand"}},
		},
	})
}
