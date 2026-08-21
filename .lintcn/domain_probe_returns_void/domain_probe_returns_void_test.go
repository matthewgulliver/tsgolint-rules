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
		// Fire-and-forget, spelled both ways the language offers.
		{
			Code: `
        export interface PledgeInstrumentation {
          readonly pledgeAccepted: (facts: PledgeFacts) => void
          readonly pledgeRefused: (facts: PledgeFacts) => undefined
        }
      `,
			FileName: inApplicationPorts,
		},
		// A persistence port is not a probe, and its `Promise<void>` is an
		// acknowledgement the caller is meant to wait for.
		{
			Code: `
        export interface PledgePersistence {
          readonly save: (facts: PledgeFacts) => Promise<void>
        }
      `,
			FileName: inApplicationPorts,
		},
		// The gate: outside the ports tree (default file name `file.ts`) the
		// rule reports nothing, however answering the probe is.
		{
			Code: `
        export interface PledgeInstrumentation {
          readonly pledgeAccepted: (facts: PledgeFacts) => PledgeReceipt
        }
      `,
		},
	}, []rule_tester.InvalidTestCase{
		// A probe that answers is a probe the caller can now depend on.
		{
			Code: `
        export interface PledgeInstrumentation {
          readonly pledgeAccepted: (facts: PledgeFacts) => PledgeReceipt
        }
      `,
			FileName: inApplicationPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "probeMemberReturnsValue"}},
		},
		// The recorded gap: the answer is behind an alias, which a syntactic rule
		// reads as `void`.
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
		// Two answering members, reported once: the keyword is settled by the first.
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
