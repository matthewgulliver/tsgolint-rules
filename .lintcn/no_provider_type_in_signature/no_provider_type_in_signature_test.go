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
	"node_modules/undici-types/index.d.ts": `
    export interface Request { readonly url: string }
    export interface Headers { get(name: string): string | null }
  `,
	"node_modules/zod/index.d.ts": `
    export type infer<T> = T extends { _output: infer O } ? O : never;
    export declare const z: { object: (shape: unknown) => { _output: unknown } };
  `,
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
		{
			Code: `
        interface PledgePersistence { readonly save: (id: string) => Promise<void> }
        export const createPledgingToOccasions =
          (persistence: PledgePersistence) => async (id: string) => persistence.save(id)
      `,
			FileName: inApplication,
		},
		{
			Code:     `export const at = (when: Date, ids: ReadonlyArray<string>) => [when, ids]`,
			FileName: inApplication,
		},
		{
			Code: `
        import type { infer as Infer } from "zod"
        declare const Command: { _output: { readonly occasionId: string } }
        export const pledge = (command: Infer<typeof Command>) => command
      `,
			FileName: inApplication,
			Files:    vendor,
		},
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
		{
			Code: `
        export const pledge = (occasionId: string) => occasionId
      `,
			FileName: inApplication,
		},
		{
			Code: `
        type PledgeRequest = { readonly occasionId: string }
        export const pledge = (request: PledgeRequest) => request.occasionId
      `,
			FileName: inApplication,
		},
		{
			Code: `
        interface Request { readonly body: string }
        export const pledge = (request: Request): Request => request
      `,
			FileName: inApplication,
		},
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
		{
			Code: `
        import type { infer as Infer } from "zod"
        declare const Command: { _output: { readonly occasionId: string } }
        export const parse = (): Infer<typeof Command> => Command._output
      `,
			FileName: inApplication,
			Files:    vendor,
		},
		{
			Code:     `export const setCookie = (res: Response, id: string) => res.headers.set("Set-Cookie", id)`,
			FileName: "packages/gifting/hexagon/adapters/driving/bff/src/session-cookie.ts",
		},
		{
			Code: `
        export type PledgeCommand = { readonly occasionId: string }
        export const pledge = ({ occasionId }: PledgeCommand) => occasionId
      `,
			FileName: inApplication,
			Files:    vendor,
		},
		{
			Code: `
        namespace shapes { export interface Of { readonly value: string } }
        export const value = (held: shapes.Of): shapes.Of => held
      `,
			FileName: inApplication,
		},
		{
			Code: `
        import type { Context } from "hono"
        export const pledge = (context: Context) => context.req.url
      `,
			Files: vendor,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        import type { Context } from "hono"
        export const pledge = (context: Context) => context.req.url
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerParameterType"}},
		},
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
		{
			Code: `
        export const pledge = (principal: string) => principal
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unbrandedPrincipal"}},
		},
		{
			Code: `
        type Caller = { readonly userId: string; readonly tenantId: string }
        export const pledge = (actor: Caller) => actor.userId
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unbrandedPrincipal"}},
		},
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
		{
			Code: `
        import type { Pool } from "pg"
        export interface PledgingDeps { readonly pool: Pool }
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerReturnType"}},
		},
		{
			Code: `
        import type { Pool as Connection } from "pg"
        export const connect = (): Connection => ({} as Connection)
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerReturnType"}},
		},
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
		{
			Code: `
        import type { Request } from "undici-types"
        export const pledge = (request: Request) => request.url
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "transportTypeInSignature"}},
		},
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
		{
			Code: `
        /// <reference lib="dom" />
        export const pledge = (request: Request) => request.url
      `,
			FileName: inApplication,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "transportTypeInSignature"}},
		},
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
		{
			Code: `
        import type { Pool } from "pg"
        export const pledge = ({ pool }: Pool) => pool
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerParameterType"}},
		},
		{
			Code: `
        import type { storage } from "aws-sdk"
        export const send = (client: storage.Client) => client.send("put")
      `,
			FileName: inApplication,
			Files:    vendor,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "providerParameterType"}},
		},
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
