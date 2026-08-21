# no-provider-type-in-signature

Type-aware. Written in Go, run by `npx lintcn lint`, **not** by
`oxlint`.

### What it does

Four failures on one subject: the types written in the hexagon's signatures.

`providerParameterType` reports a parameter whose annotation names a **class or
interface declared by a package dependency** — `Context` from Hono, `Pool` from
`pg`. `providerReturnType` asks the same of a return type annotation and of a
property signature's type, so a port answering `Promise<QueryResult<Row>>` or a
deps object holding `readonly pool: Pool` reports too.

`transportTypeInSignature` reports a written name in `transportTypeNamePatterns`
— `Request`, `Response`, `Headers`, `FormData` — whose declaration belongs to the
platform or a package rather than this repository. A domain `PledgeRequest`, or
a local `interface Request`, is not the transport and passes.

`unbrandedPrincipal` reports a parameter the vocabulary calls identity —
`principal`, `actor`, `currentUser` — whose type carries **no brand**: no
property keyed by a `unique symbol` this repository declares.

It resolves each type name written in the annotation, including the arguments of
a generic and the members of a type literal, follows import aliases, and judges
the file that declares the symbol.

### Why is this bad?

The driven doc: a port never speaks SQL, HTTP or vendor language, in what it
takes or in what it answers. The driving doc and `route-handler.md`: a driving
port takes typed application commands; no `Request`, headers or cookies cross
the boundary. A parameter typed `Context` means the inside can only be called by
code holding Hono's object — the CLI, the test and the next transport all have
to fabricate one; a port answering in the driver's `QueryResult` couples every
consumer the same way from the other side. And
`cross-cutting/authentication.md`: authentication ends at a **provider-free
branded principal**, because a `principal: string` can be assembled by anybody
holding the ids.

### Why it is type-aware

The syntactic [`port-speaks-domain-language`](port-speaks-domain-language.md)
catches *spellings*; rename the import (`import type { Context as Ambient }`)
and it goes quiet. Ownership of a declaration cannot be renamed, and it is what
tells the platform's `Request` from a domain `PledgeRequest`.

### Examples

Examples of **incorrect** code for this rule:

```ts
import type { Context } from "hono"
export const createPledgingToOccasions = () => (context: Context) =>
  context.req.url
```

```ts
// The driver's answer, and the vendor's object held as a member.
import type { Pool, QueryResult } from "pg"
export interface OccasionReads {
  rowsFor(id: OccasionId): Promise<QueryResult<Row>>
  readonly pool: Pool
}
```

```ts
// The platform's Request, resolved to lib.dom or a package, not ours.
export const pledge = (request: Request) => request.url
```

```ts
// Unbranded: possession of the string is enough to be anybody.
export const pledge = (principal: string) => principal
```

Examples of **correct** code for this rule:

```ts
export const createPledgingToOccasions =
  (persistence: PledgePersistence) => (command: PledgeCommand) =>
    persistence.saveWithOutbox(command)
```

```ts
// A schema-derived command. `z.infer` is declared in `node_modules` and is a
// type alias, so the shape it describes stays this repository's own.
export const pledge = (command: z.infer<typeof PledgeCommand>) => command
```

```ts
// Standard-library types are not a vendor's object, and a name that only
// resembles the transport is this repository's own.
export const at = (when: Date, ids: ReadonlyArray<OccasionId>) => [when, ids]
type PledgeRequest = { readonly occasionId: OccasionId }
export const pledge = (request: PledgeRequest) => request.occasionId
```

```ts
// Branded by a symbol nothing exports.
declare const authenticated: unique symbol
type AuthenticatedPrincipal = { readonly [authenticated]: true; readonly userId: UserId }
export const pledge = (principal: AuthenticatedPrincipal) => principal.userId
```

### Options

These are the values the rule uses. They are **not** overridable through
lintcn today: `runner.go` passes `nil` options to every rule on every file,
so the defaults below are the shipped behaviour and the option names are a
test-only surface.

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/hexagon/application/**", "**/hexagon/domain/**"]` | The trees this rule judges — both halves of the inside, because the claim is about the inside as a whole. Not `**/hexagon/**`: that also matches `hexagon/adapters/**`, where holding the transport is the adapter's job. |
| `principalParameterPatterns` | `string[]` | `["^principal$", "^actor$", "^currentUser$"]` | A parameter whose name matches must be branded. |
| `transportTypeNamePatterns` | `string[]` | `["^(?:Request\|Response\|Headers\|FormData)$"]` | A written type name that matches, and resolves to a platform or package declaration, is the transport crossing the boundary. |

### How to use

```bash
npx lintcn lint            # build the binary if needed, then run it
cd .lintcn && TSGOLINT_SNAPSHOT_CWD=true go test ./...   # the rule tests
```

### Limitations

**Only classes and interfaces.** A generic type alias from a package — `z.infer`,
a `Simplify<T>` helper — describes a shape this repository owns and is
deliberately not reported. Deciding otherwise would flag every schema-derived
command, which the docs recommend.

**Signatures only.** A provider type reached through a generic alias, an
inline structural type (`{ rows: ReadonlyArray<{ id: string }> }`), a variable
annotation or an inferred type is not reported. Standard-library types that are
not in `transportTypeNamePatterns` — `URL`, `ReadableStream` — pass, on the same
line `no-constructed-collaborators` draws.

**The brand arm is keyed on the parameter's name.** A principal passed as
`caller` is not judged until the vocabulary says so, and a `principal` that is
not one reports. Only a symbol this repository declares brands anything —
`string` carries `[Symbol.iterator]` from the standard library and does not
count.

**Scope is the rule's own**, from the `files` above. The rule judges the file it
is handed and carries no scope check of its own.
