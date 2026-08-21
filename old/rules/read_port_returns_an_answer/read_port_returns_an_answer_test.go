package read_port_returns_an_answer

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inPorts = "packages/gifting/hexagon/application/src/ports/occasion-dashboard-rows.ts"

func TestReadPortReturnsAnAnswer(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &ReadPortReturnsAnAnswerRule.Rule, []rule_tester.ValidTestCase{
		// A read that answers with rows.
		{
			Code: `
        type OccasionDashboardRow = { readonly occasionId: string }
        export interface OccasionDashboardRows {
          readonly forContributor: (id: string) => Promise<ReadonlyArray<OccasionDashboardRow>>
        }
      `,
			FileName: inPorts,
		},
		// A write port is allowed to answer with nothing.
		{
			Code: `
        export interface PledgePersistence {
          readonly saveWithOutbox: (id: string) => Promise<void>
        }
      `,
			FileName: inPorts,
		},
		// An answer that may be absent is still an answer.
		{
			Code: `
        export interface OccasionDashboardView {
          readonly forContributor: (id: string) => Promise<string | undefined>
        }
      `,
			FileName: inPorts,
		},
		// An unexported local type is no contract another module depends on.
		{
			Code: `
        interface OccasionDashboardRows {
          readonly recordDashboardView: (id: string) => Promise<void>
        }
      `,
			FileName: inPorts,
		},
	}, []rule_tester.InvalidTestCase{
		// A configured `files` scope replaces the default tree.
		{
			Code: `
        export interface OccasionDashboardRows {
          readonly recordDashboardView: (id: string) => Promise<void>
        }
      `,
			FileName: inPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "answerlessRead"}},
		},
		// The alias the syntactic rule cannot follow: the annotation says
		// `Acknowledged`, the type says nothing at all.
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
		// A member whose name carries no write verb at all, so the JS rule is
		// silent on it.
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
