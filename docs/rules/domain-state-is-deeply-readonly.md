# domain-state-is-deeply-readonly

### What it does

Follows the members of every **exported** `interface` and `type` in the domain
and shared-kernel trees through the types they name, and reports state that is
writable once resolved. Two `messageId`s:

- `mutableMemberResolved` — a member reached through a named type
  (`readonly funding: FundingState`, where `FundingState` declares `pledges`
  without `readonly`) is writable.
- `mutableCollectionResolved` — a member's resolved type is a mutable `Array`,
  `Map` or `Set`, however the annotation spelled it: an alias
  (`readonly pledges: PledgeList` with `type PledgeList = string[]`), a union
  with `undefined`, or an alias reached one level down.

Element types of a `ReadonlyArray`, `ReadonlyMap` and `ReadonlySet` are followed
too: `ReadonlyArray<Pledge>` where `Pledge` has a writable `amount` reports.

What it deliberately leaves to the syntactic rules: a member written mutable in
the judged declaration itself — no `readonly`, or a written `T[]` / `Array<T>` —
is `domain-state-is-readonly`'s and
`shared-kernel-export-is-an-immutable-value`'s
report, and a member of *another* exported declaration is reported when that
declaration is visited, not through every type that names it. Callable members
are behaviour, not state, and are skipped; members the standard library or a
package declares are not the model's.

### Why is this bad?

`aggregate-root.md`: the root is the
only write surface, and `ReadonlyArray` and immutable values are what stop a
caller altering the collection directly. `shared-kernel.md`:
`Money` is immutable. A `readonly` keyword on the outer member keeps that
promise only as far as the type it names — one unexported `FundingState` with a
bare `pledges: Pledge[]`, or one `type PledgeList = Pledge[]`, and the aggregate
can be changed through a member the root never sees changing.

### Why it is type-aware

`readonly funding: FundingState` and `readonly pledges: PledgeList` are opaque
to a rule reading one declaration: what those names resolve to — and whether the
member behind them carries `readonly` — is the checker's answer.

### Examples

Examples of **incorrect** code for this rule:

```ts
// packages/gifting/hexagon/domain/src/occasions/occasion.ts
type FundingState = { pledges: ReadonlyArray<string>; readonly closed: boolean }
export type Occasion = { readonly id: string; readonly funding: FundingState }
```

```ts
type PledgeList = string[]
export type Occasion = { readonly pledges: PledgeList }
```

```ts
type Pledge = { amount: number }
export type Occasion = { readonly pledges: ReadonlyArray<Pledge> }
```

Examples of **correct** code for this rule:

```ts
type PledgeList = ReadonlyArray<{ readonly id: string }>
type FundingState = { readonly pledges: PledgeList; readonly closed: boolean }
export type Occasion = { readonly id: string; readonly funding: FundingState }
```

```ts
// Written mutable in the declaration itself: `domain-state-is-readonly` reports it.
export type Occasion = { id: string; pledges: string[] }
```

### Options

These are the values the rule uses. They are **not** overridable through
lintcn today: `runner.go` passes `nil` options to every rule on every file,
so the defaults below are the shipped behaviour and the option names are a
test-only surface.

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/hexagon/domain/**", "**/shared-kernel/**"]` | The trees this rule judges. |
| `mutableCollectionTypeNames` | `string[]` | `["Array", "Map", "Set"]` | Resolved type names read as a mutable collection. |

### Limitations

**Depth is capped at six members** and each type is followed once per
declaration, so a self-referential type terminates and a mutable member deeper
than that passes.

**`readonly` is a compile-time claim.** A cast or an `any` still writes; this
rule reads declarations, not code.

**A mutable member of an exported type reached from another tree passes here.**
`**/hexagon/domain/**` and `**/shared-kernel/**` are the only trees visited, so
an application-tree type named by a domain member is
[`domain-signature-stays-in-the-domain`](domain-signature-stays-in-the-domain.md)'s
subject rather than this rule's.

**Reports land on the exported declaration's name**, with the member path in
the message; the fix is where the path ends.

**Scope is the rule's own**, from the `files` above. The rule judges the file it
is handed and carries no scope check of its own.
