package no_provider_type_in_signature

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inApplication = "packages/gifting/hexagon/application/src/pledge-to-occasion.ts"
const inPorts = "packages/gifting/hexagon/application/src/ports/for-pledging-to-occasions.ts"

var vendor = map[string]string{
	"node_modules/pg/index.d.ts": `
    export declare class Pool { query(text: string): Promise<unknown>; }
  `,
	"node_modules/hono/index.d.ts": `
    export interface Context { readonly req: { url: string }; }
  `,
	"node_modules/pg-types/index.d.ts": `
    export interface QueryResult<R> { rows: R[] }
  `,
	// The platform's transport type, declared by a package rather than lib.dom.
	"node_modules/undici-types/index.d.ts": `
    export interface Request { readonly url: string }
    export interface Headers { get(name: string): string | null }
  `,
	// A generic helper is a type alias, the shape it describes is ours.
	"node_modules/zod/index.d.ts": `
    export type infer<T> = T extends { _output: infer O } ? O : never;
    export declare const z: { object: (shape: unknown) => { _output: unknown } };
  `,
	// A vendor that publishes its types inside a namespace, so every reference
	// to one is written `namespace.Type` rather than as a bare name.
	"node_modules/aws-sdk/index.d.ts": `
    export declare namespace storage {
      export interface Client { send(command: string): Promise<void> }
      export namespace inner { export interface Handle { readonly key: string } }
    }
  `,
}

func TestNoProviderTypeInSignature(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoProviderTypeInSignatureRule, []rule_tester.ValidTestCase{
		// A port this repository declares, which is what a signature here should say.
		{
			Code: `
        interface PledgePersistence { readonly save: (id: string) => Promise<void> }
        export const createPledgingToOccasions =
          (persistence: PledgePersistence) => async (id: string) => persistence.save(id)
      `,
			FileName: inApplication,
		},
		// Standard-library types are not a vendor's object.
		{
			Code:     `export const at = (when: Date, ids: ReadonlyArray<string>) => [when, ids]`,
			FileName: inApplication,
		},
		// A schema-derived command: declared through an alias in node_modules,
		// describing a shape this repository owns.
		{
			Code: `
        import type { infer as Infer } from "zod"
        declare const Command: { _output: { readonly occasionId: string } }
        export const pledge = (command: Infer<typeof Command>) => command
      `,
			FileName: inApplication,
			Files:    vendor,
		},
		// The doc's own principal: branded with a non-exported unique symbol,
		// and narrowed by intersection rather than replaced.
		{
			Code: `
        declare const authenticated: unique symbol
        type AuthenticatedPrincipal = {
          readonly [authenticated]: true
          readonly userId: string
          readonly tenantId: string
        }
        type AuthenticatedPledger = AuthenticatedPrincipal & { readonly contributorId: string }
        export const pledge = (principal: AuthenticatedPledger) => principal.contributorId
      `,
			FileName: inApplication,
		},
		// A parameter the vocabulary does not call identity.
		{
			Code: `
        export const pledge = (occasionId: string) => occasionId
      `,
			FileName: inApplication,
		},
		// A domain command that merely shares the transport's name.
		{
			Code: `
        type PledgeRequest = { readonly occasionId: string }
        export const pledge = (request: PledgeRequest) => request.occasionId
      `,
			FileName: inApplication,
		},
		// A local `Request` is this repository's own declaration, not the platform's.
		{
			Code: `
        interface Request { readonly body: string }
        export const pledge = (request: Request): Request => request
      `,
			FileName: inApplication,
		},
		// A return type this repository owns, wrapped in the language's own Promise.
		{
			Code: `
        type Occasion = { readonly id: string }
        export interface OccasionPersistence {
          findById(id: string): Promise<Occasion | undefined>
          readonly all: () => Promise<ReadonlyArray<Occasion>>
        }
      `,
			FileName: inPorts,
		},
		// A schema-derived return: a package alias describing our shape.
		{
			Code: `
        import type { infer as Infer } from "zod"
        declare const Command: { _output: { readonly occasionId: string } }
        export const parse = (): Infer<typeof Command> => Command._output
      `,
			FileName: inApplication,
			Files:    vendor,
		},
		// A driving adapter holding the transport, which is its whole job.
		{
			Code:     `export const setCookie = (res: Response, id: string) => res.headers.set("Set-Cookie", id)`,
			FileName: "packages/gifting/hexagon/adapters/driving/bff/src/session-cookie.ts",
		},
		// Destructuring a type this repository owns is ordinary, and naming no
		// parameter must not stop the rule reaching the next file.
		{
			Code: `
        export type PledgeCommand = { readonly occasionId: string }
        export const pledge = ({ occasionId }: PledgeCommand) => occasionId
      `,
			FileName: inApplication,
			Files:    vendor,
		},
		// A namespaced type this repository declares. Asking a qualified name
		// for its text as if it were an identifier took the whole run down, so
		// every file after it went unjudged.
		{
			Code: `
        namespace shapes { export interface Of { readonly value: string } }
        export const value = (held: shapes.Of): shapes.Of => held
      `,
			FileName: inApplication,
		},
		// The gate: outside the inside trees (default file name `file.ts`) the
		// rule reports nothing, however vendor-heavy the signature is.
		{
			Code: `
        import type { Context } from "hono"
        export const pledge = (context: Context) => context.req.url
      `,
			Files: vendor,
		},
	}, []rule_tester.InvalidTestCase{
		// A framework's request object reaching application policy.
		{
			Code: `
        import type { Context } from "hono"
        export const pledge = (context: Context) => context.req.url
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerParameterType"}},
		},
		// A port method's parameter, and the vendor type nested in a generic.
		{
			Code: `
        import type { Pool } from "pg"
        export interface ForPledgingToOccasions {
          readonly pledgeToOccasion: (connection: ReadonlyArray<Pool>) => Promise<void>
        }
      `,
			FileName: inPorts,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerParameterType"}},
		},
		// Identity as a bare id: possession of the string is enough.
		{
			Code: `
        export const pledge = (principal: string) => principal
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unbrandedPrincipal"}},
		},
		// A principal-shaped object anyone can write out.
		{
			Code: `
        type Caller = { readonly userId: string; readonly tenantId: string }
        export const pledge = (actor: Caller) => actor.userId
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unbrandedPrincipal"}},
		},
		// A parameter named identity only by configuration.
		{
			Code: `
        export const pledge = (caller: string) => caller
      `,
			FileName: inApplication,
			Options: Options{
				PrincipalParameterPatterns: []string{"^caller$"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unbrandedPrincipal"}},
		},
		// A port method answering in the driver's vocabulary, nested in a Promise.
		{
			Code: `
        import type { QueryResult } from "pg-types"
        type Row = { readonly id: string }
        export interface OccasionReads {
          rowsFor(id: string): Promise<QueryResult<Row>>
        }
      `,
			FileName: inPorts,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerReturnType"}},
		},
		// A vendor object held as a member of an inside contract.
		{
			Code: `
        import type { Pool } from "pg"
        export interface PledgingDeps { readonly pool: Pool }
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerReturnType"}},
		},
		// Renamed on import; the checker follows the alias.
		{
			Code: `
        import type { Pool as Connection } from "pg"
        export const connect = (): Connection => ({} as Connection)
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerReturnType"}},
		},
		// A function-typed member returning the vendor's object reports once.
		{
			Code: `
        import type { QueryResult } from "pg-types"
        export interface OccasionReads {
          readonly rowsFor: (id: string) => Promise<QueryResult<{ id: string }>>
        }
      `,
			FileName: inPorts,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerReturnType"}},
		},
		// The platform's Request crossing into the inside.
		{
			Code: `
        import type { Request } from "undici-types"
        export const pledge = (request: Request) => request.url
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "transportTypeInSignature"}},
		},
		// Headers as a port method's answer.
		{
			Code: `
        import type { Headers } from "undici-types"
        export interface ForPledgingToOccasions {
          pledge(id: string): Promise<Headers>
        }
      `,
			FileName: inPorts,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "transportTypeInSignature"}},
		},
		// The standard library's own transport type, pulled in by reference.
		{
			Code: `
        /// <reference lib="dom" />
        export const pledge = (request: Request) => request.url
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "transportTypeInSignature"}},
		},
		// A transport type named only by configuration.
		{
			Code: `
        import type { Context } from "hono"
        export interface Deps { readonly context: Context }
      `,
			FileName: inApplication,
			Files:    vendor,
			Options:  Options{TransportTypeNamePatterns: []string{"^Context$"}},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "transportTypeInSignature"}},
		},
		// The transport renamed on import: a new name, the same declaration.
		{
			Code: `
        import type { Request as Incoming } from "undici-types"
        export const pledge = (): Incoming => ({} as Incoming)
      `,
			FileName: inApplication,
			Files:    vendor,
			Options:  Options{TransportTypeNamePatterns: []string{"^Incoming$"}},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "transportTypeInSignature"}},
		},
		// A destructured parameter has no name to blame, and the rule used to
		// ask for one anyway and take the whole run down with it. The type is
		// still the signature's, so the vendor question is still asked.
		{
			Code: `
        import type { Pool } from "pg"
        export const pledge = ({ pool }: Pool) => pool
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerParameterType"}},
		},
		// A vendor's namespaced type, reported by the name it is declared under
		// rather than the namespace it is reached through.
		{
			Code: `
        import type { storage } from "aws-sdk"
        export const send = (client: storage.Client) => client.send("put")
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerParameterType"}},
		},
		// The same, nested one namespace deeper.
		{
			Code: `
        import type { storage } from "aws-sdk"
        export const keyOf = (handle: storage.inner.Handle) => handle.key
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerParameterType"}},
		},
	})
}
