package domain_probe_returns_void

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inApplicationPorts = "packages/gifting/hexagon/application/src/ports/pledge-instrumentation.ts"

func TestDomainProbeReturnsVoid(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &DomainProbeReturnsVoidRule, []rule_tester.ValidTestCase{
		{
			Code: `
        export interface PledgeInstrumentation {
          readonly pledgeAccepted: (facts: PledgeFacts) => void
          readonly pledgeRefused: (facts: PledgeFacts) => undefined
        }
      `,
			FileName: inApplicationPorts,
		},
		{
			Code: `
        export interface PledgePersistence {
          readonly save: (facts: PledgeFacts) => Promise<void>
        }
      `,
			FileName: inApplicationPorts,
		},
		{
			Code: `
        export interface PledgeInstrumentation {
          readonly pledgeAccepted: (facts: PledgeFacts) => PledgeReceipt
        }
      `,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        export interface PledgeInstrumentation {
          readonly pledgeAccepted: (facts: PledgeFacts) => PledgeReceipt
        }
      `,
			FileName: inApplicationPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "probeMemberReturnsValue"}},
		},
		{
			Code: `
        type Acknowledged = Promise<void>
        export interface PledgeProbe {
          readonly pledgeAccepted: (facts: PledgeFacts) => Acknowledged
        }
      `,
			FileName: inApplicationPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "probeMemberReturnsValue"}},
		},
		{
			Code: `
        export interface PledgeInstrumentation {
          readonly pledgeAccepted: (facts: PledgeFacts) => PledgeReceipt
          readonly pledgeRefused: (facts: PledgeFacts) => PledgeReceipt
        }
      `,
			FileName: inApplicationPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "probeMemberReturnsValue"}},
		},
	})
}
