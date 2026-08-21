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
		// A modelled command, which is what the doc asks a driving port to take.
		{
			Code: `
        type PledgeCommand = { readonly occasionId: string; readonly amount: number }
        export interface ForPledgingToOccasions {
          readonly pledgeToOccasion: (command: PledgeCommand) => Promise<void>
        }
      `,
			FileName: inPorts,
		},
		// A branded value is an intersection, not a bare primitive — the whole
		// reason this is a type-aware rule.
		{
			Code: `
        type OccasionId = string & { readonly __brand: "OccasionId" }
        export interface ForCancellingAnOccasion {
          readonly cancel: (id: OccasionId) => Promise<void>
        }
      `,
			FileName: inPorts,
		},
		// A literal union is a modelled vocabulary.
		{
			Code: `
        export interface ForRatingAPledge {
          readonly rate: (rating: "up" | "down") => Promise<void>
        }
      `,
			FileName: inPorts,
		},
		// A driven port answering by id is not this rule's business.
		{
			Code: `
        export interface PledgePersistence {
          readonly findOccasionById: (id: string) => Promise<string | null>
        }
      `,
			FileName: inPorts,
		},
		// The gate: outside the ports tree (default file name `file.ts`) the
		// rule reports nothing, however bare the parameter is.
		{
			Code: `
        export interface ForPledgingToOccasions {
          readonly pledgeToOccasion: (occasionId: string) => Promise<void>
        }
      `,
		},
	}, []rule_tester.InvalidTestCase{
		// Loose primitives where a command belongs: the caller has to know the
		// argument order, and nothing names what the operation needs.
		{
			Code: `
        export interface ForPledgingToOccasions {
          readonly pledgeToOccasion: (occasionId: string) => Promise<void>
        }
      `,
			FileName: inPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unmodelledCommand"}},
		},
		// An alias to a primitive is still a primitive.
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
		// A configured naming vocabulary reaches a port not called `For…`.
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
