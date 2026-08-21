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
		// Outside a port tree the same declaration is nobody's business: this
		// rule reported on the repository's own `packages/oxlint/rule.ts` until it
		// learnt to read its own path.
		{
			Code:     `export type Rule = { readonly create: (context: string) => void }`,
			FileName: "packages/oxlint/rule.ts",
		},
		// The gate: outside the ports tree (default file name `file.ts`) the
		// rule reports nothing, however behaviour-shaped the alias is.
		{
			Code: `export type ForPledgingToOccasions = (command: string) => Promise<string>`,
		},
		// A data shape, which is what `type` is reserved for here.
		{Code: `export type StoredOccasion = { readonly value: string; readonly version: number }`, FileName: inPorts},
		// Behaviour already declared with the keyword this rule asks for.
		{Code: `
      export interface PledgePersistence {
        readonly findOccasionById: (id: string) => Promise<string | null>
      }
    `, FileName: inPorts},
		// Data reached through an alias is still data.
		{Code: `
      type Occasion = { readonly id: string };
      export type StoredOccasion = { readonly value: Occasion; readonly version: number }
    `, FileName: inPorts},
		// A union of literals has no members to be callable.
		{Code: `export type PledgeResult = "saved" | "conflict"`, FileName: inPorts},
		// Not exported, so not a published contract.
		{Code: `type ToCard = (row: string) => string`, FileName: inPorts},
		// Written as a function type on an exported alias, which is exactly what
		// `port-contract-is-an-interface` reads. Both rules reporting made one
		// mistake produce two diagnostics saying nearly the same sentence.
		{Code: `export type ForPledgingToOccasions = (command: string) => Promise<string>`, FileName: inPorts},
		// A type literal whose member is written as a function type: the syntactic
		// rule reports this one, so this rule does not.
		{Code: `
      export type PledgePersistence = {
        readonly findOccasionById: (id: string) => Promise<string | null>
      }
    `, FileName: inPorts},
	}, []rule_tester.InvalidTestCase{
		{
			// The recorded gap: the member's callability is behind an alias, so the
			// syntactic rule reads it as data.
			Code: `
        type PledgeToOccasion = (command: string) => Promise<string>;
        export type ForPledgingToOccasions = { readonly pledgeToOccasion: PledgeToOccasion }
      `,
			FileName: inPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "behaviourContractAsTypeAlias"}},
		},
		{
			// The alias itself is behind an alias.
			Code: `
        type Pledging = (command: string) => Promise<string>;
        export type ForPledgingToOccasions = Pledging
      `,
			FileName: inPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "bareFunctionTypeAlias"}},
		},
		{
			// An intersection carrying behaviour is still a behaviour contract.
			Code: `
        type Reads = { readonly findOccasionById: (id: string) => Promise<string> };
        type Version = { readonly version: number };
        export type PledgePersistence = Reads & Version
      `,
			FileName: inPorts,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "behaviourContractAsTypeAlias"}},
		},
		// Configured to judge unexported aliases too, which the syntactic rule
		// never opens — so there is no second diagnostic to collide with.
		{
			Code:     `type ToCard = (row: string) => string`,
			FileName: inPorts,
			Options:  map[string]any{"exportedOnly": false},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "bareFunctionTypeAlias"}},
		},
	})
}
