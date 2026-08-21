# port-behaviour-is-an-interface

### What it does

Reports an exported `type` alias in the ports tree that carries behaviour, by
*resolving* the alias rather than reading its written shape:

- `bareFunctionTypeAlias` — the alias resolves to a type with a call signature.
- `behaviourContractAsTypeAlias` — the alias resolves to an object or
  intersection with at least one member whose type resolves to a callable.

An alias whose members all resolve to data is untouched — that is the shape the
driven-port doc reserves `type` for.

### Why is this bad?

`driven-port.md` draws the line as
`interface` for behaviour contracts, `type` and schemas for pure data shapes, so
a reader can tell at a glance which declarations an adapter has to satisfy.
`driving-port.md` sharpens the
bare-function case: the interface is the **role**, the method is the operation.
Export the function type and the role has no name — an adapter holds a callable
rather than a collaborator, and the actor's second operation arrives as a second
injected parameter threaded through every construction site instead of one more
method nobody had to re-plumb.

This is why the rule is type-aware. The syntactic
`port-contract-is-an-interface` records the
gap in its own Limitations: a member typed through an alias
(`readonly pledgeToOccasion: PledgeToOccasion`) "reads as data and escapes
`behaviourContractAsTypeAlias` — the more indirection, the less this rule sees".
That indirection is not a matcher weakness to be tightened; it is a type
question, and asking it syntactically would leave the convention unenforced for
exactly the ports that named their operations well.

### Examples

Examples of **incorrect** code for this rule:

```ts
export type ForPledgingToOccasions = (
  command: PledgeCommand
) => Promise<PledgeResult>
```

```ts
export type PledgePersistence = {
  readonly findOccasionById: (id: string) => Promise<StoredOccasion | null>
}
```

```ts
// The alias the syntactic rule reads as data.
type PledgeToOccasion = (command: PledgeCommand) => Promise<PledgeResult>
export type ForPledgingToOccasions = {
  readonly pledgeToOccasion: PledgeToOccasion
}
```

```ts
// An intersection that mixes data in still publishes behaviour.
type Reads = { readonly findOccasionById: (id: string) => Promise<StoredOccasion> }
type Version = { readonly version: number }
export type PledgePersistence = Reads & Version
```

Examples of **correct** code for this rule:

```ts
export type StoredOccasion = {
  readonly value: Occasion
  readonly version: number
}
```

```ts
export type PledgeResult = "saved" | "conflict"
```

```ts
export interface PledgePersistence {
  readonly findOccasionById: (id: string) => Promise<StoredOccasion | null>
  readonly saveWithOutbox: (
    occasion: Occasion,
    events: ReadonlyArray<DomainEvent>,
    expectedVersion: number
  ) => Promise<"saved" | "conflict">
}
```

### Options

These are the values the rule uses. They are **not** overridable through
lintcn today: `runner.go` passes `nil` options to every rule on every file,
so the defaults below are the shipped behaviour and the option names are a
test-only surface.

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/ports/**"]` | The trees this rule judges. `**` spans any number of segments. |
| `exportedOnly` | `boolean` | `true` | Judge only aliases the module publishes. A module-private `type ToCard = (row: Row) => Card` beside a port is a local helper, not a boundary. Set `false` to judge every alias in scope. |

### How to use

```bash
npx lintcn lint            # build the binary if needed, then run it
cd .lintcn && TSGOLINT_SNAPSHOT_CWD=true go test ./...   # the rule tests
```

First build clones tsgolint and typescript-go and takes a few minutes; after
that it is incremental.

### Limitations

**A callable written in place belongs to the syntactic rule.** An exported alias
whose type is written `(x: T) => U`, or whose type literal has a member written
that way, is reported by
`port-contract-is-an-interface` and is
deliberately silent here — both firing turned one mistake into two diagnostics
saying nearly the same sentence. What is left is what that rule cannot see: a
member typed through an alias, an alias of an alias, an intersection, and any
declaration it does not open, which is every non-exported one.

**Scope is the rule's own**, from the `files` above. Before anything scoped this
rule it reported `Context` and `Visitor` in `packages/oxlint/rule.ts` —
right about those aliases, and wrong to be looking at them.

Only type alias declarations. An `interface` holding nothing but data fields is
a data shape wearing the contract keyword, and this rule no more reports it than
the syntactic one does — the keyword it judges is `type`.

A union is not inspected for members. `type X = A | { run: () => void }` passes,
because the properties of a union are the ones common to every constituent and a
union of shapes is not the declaration this convention is about.

The member scan stops at the first callable, so a contract with three behaviour
members reports once, naming one. And callability is not intent: a data shape
whose field genuinely holds a function — a comparator, a formatter — reports,
and whether that field is a boundary an adapter implements is the modelling
question no matcher reaches.
