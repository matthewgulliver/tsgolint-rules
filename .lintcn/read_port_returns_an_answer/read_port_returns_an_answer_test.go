package read_port_returns_an_answer

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inPorts = "packages/gifting/hexagon/application/src/ports/occasion-dashboard-rows.ts"

func TestReadPortReturnsAnAnswer(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &ReadPortReturnsAnAnswerRule, []rule_tester.ValidTestCase{
		{
			Code: `
        type OccasionDashboardRow = { readonly occasionId: string }
        export interface OccasionDashboardRows {
          readonly forContributor: (id: string) => Promise<ReadonlyArray<OccasionDashboardRow>>
        }
      `,
			FileName: inPorts,
		},
		{
			Code: `
        export interface PledgePersistence {
          readonly saveWithOutbox: (id: string) => Promise<void>
        }
      `,
			FileName: inPorts,
		},
		{
			Code: `
        export interface OccasionDashboardView {
          readonly forContributor: (id: string) => Promise<string | undefined>
        }
      `,
			FileName: inPorts,
		},
		{
			Code: `
        interface OccasionDashboardRows {
          readonly recordDashboardView: (id: string) => Promise<void>
        }
      `,
			FileName: inPorts,
		},
		{
			Code: `
        export interface OccasionDashboardRows {
          readonly recordDashboardView: (id: string) => Promise<void>
        }
      `,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        export interface OccasionDashboardRows {
          readonly recordDashboardView: (id: string) => Promise<void>
        }
      `,
			FileName: inPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "answerlessRead"}},
		},
		{
			Code: `
        type Acknowledged = Promise<void>
        export interface OccasionDashboardReader {
          readonly touch: (id: string) => Acknowledged
        }
      `,
			FileName: inPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "answerlessRead"}},
		},
		{
			Code: `
        export interface ContributorDashboard {
          readonly refresh: (id: string) => Promise<void>
        }
      `,
			FileName: inPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "answerlessRead"}},
		},
	})
}
