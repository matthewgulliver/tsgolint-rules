package port_behaviour_is_an_interface

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inPorts = "packages/gifting/hexagon/application/src/ports/pledge-persistence.ts"

func TestPortBehaviourIsAnInterface(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &PortBehaviourIsAnInterfaceRule, []rule_tester.ValidTestCase{
		{
			Code:     `export type Rule = { readonly create: (context: string) => void }`,
			FileName: "packages/oxlint/rule.ts",
		},
		{
			Code: `export type ForPledgingToOccasions = (command: string) => Promise<string>`,
		},
		{Code: `export type StoredOccasion = { readonly value: string; readonly version: number }`, FileName: inPorts},
		{Code: `
      export interface PledgePersistence {
        readonly findOccasionById: (id: string) => Promise<string | null>
      }
    `, FileName: inPorts},
		{Code: `
      type Occasion = { readonly id: string };
      export type StoredOccasion = { readonly value: Occasion; readonly version: number }
    `, FileName: inPorts},
		{Code: `export type PledgeResult = "saved" | "conflict"`, FileName: inPorts},
		{Code: `type ToCard = (row: string) => string`, FileName: inPorts},
		{Code: `export type ForPledgingToOccasions = (command: string) => Promise<string>`, FileName: inPorts},
		{Code: `
      export type PledgePersistence = {
        readonly findOccasionById: (id: string) => Promise<string | null>
      }
    `, FileName: inPorts},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        type PledgeToOccasion = (command: string) => Promise<string>;
        export type ForPledgingToOccasions = { readonly pledgeToOccasion: PledgeToOccasion }
      `,
			FileName: inPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "behaviourContractAsTypeAlias"}},
		},
		{
			Code: `
        type Pledging = (command: string) => Promise<string>;
        export type ForPledgingToOccasions = Pledging
      `,
			FileName: inPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "bareFunctionTypeAlias"}},
		},
		{
			Code: `
        type Reads = { readonly findOccasionById: (id: string) => Promise<string> };
        type Version = { readonly version: number };
        export type PledgePersistence = Reads & Version
      `,
			FileName: inPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "behaviourContractAsTypeAlias"}},
		},
		{
			Code:     `type ToCard = (row: string) => string`,
			FileName: inPorts,
			Options:  map[string]any{"exportedOnly": false},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "bareFunctionTypeAlias"}},
		},
	})
}
