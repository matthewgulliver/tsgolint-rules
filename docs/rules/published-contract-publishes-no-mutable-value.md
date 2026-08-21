# published-contract-publishes-no-mutable-value

Type-aware. Written in Go, run by `npx lintcn lint`, **not** by
`oxlint`.

### What it does

Reports an exported `const` in a published contract file whose **resolved type**
is mutable — either a writable member, or a standard-library container that
exposes a writer.

Two `messageId`s, because freezing an object is not the repair for a published
`Map`: `mutableMemberInContract` and `mutableContainerInContract`.

### Why is this bad?

[`bounded-context.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/domain/bounded-context.md) §2.3: "never
share mutable state between contexts — communicate via events or explicit
contracts". A published mutable value is a channel nobody declared. One context
writes, another reads a different value than it did a moment ago, and the change
appears in no event, no contract and no diff.

### Why it is type-aware

This is the recorded blind spot of the syntactic
[`no-mutable-export-in-public-contract`](no-mutable-export-in-public-contract.md).
These two are the same syntax — an exported `const` bound to a call — and
opposite contracts:

```ts
export const CURRENCIES = Object.freeze({ gbp: "GBP" })
export const REGISTRY = buildRegistry()
```

The first resolves to `Readonly<{ gbp: string }>`, the second to
`Map<string, string>`. Only the type says which.

The two rules ship together on purpose: the syntactic one catches `let`, a bare
object literal and a literal `new Map()` without a checker, and this one catches
what a call hides.

### Examples

Examples of **incorrect** code for this rule:

```ts
export const REGISTRY = buildRegistry()          // Map<string, string>
export const CODES: Array<string> = ["gbp"]
export const LIMITS = build()                    // { maxOpenPledges: number }
```

Examples of **correct** code for this rule:

```ts
export const CURRENCIES = Object.freeze({ gbp: "GBP" })
export const CODES: ReadonlyArray<string> = ["gbp"]
export const BY_CODE: ReadonlyMap<string, string> = new Map()
export const describeOccasion = (id: OccasionId): string => …
```

### Options

These are the values the rule uses. They are **not** overridable through
lintcn today: `runner.go` passes `nil` options to every rule on every file,
so the defaults below are the shipped behaviour and the option names are a
test-only surface.

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/public.ts", "**/public.tsx"]` | The files this rule judges. |
| `mutatingMemberNames` | `string[]` | `["set", "add", "delete", "clear", "push", "pop", "shift", "unshift", "splice", "sort", "reverse", "fill", "copyWithin"]` | A member with one of these names, declared by the standard library, reports `mutableContainerInContract`. |

### Limitations

**The container arm only judges members the standard library declares.** A
published contract of this repository's own with an operation named `set` is a
command, not a mutation of shared state, and is deliberately not reported.
Remove that gate and it is; there is a test holding it down.

**A method is never read as state.** Judging callable members as writable
properties reads every interface with a method as mutable — `ReadonlyMap`
included — which is how this check would stop meaning anything.

**A published factory is not judged.** What a function returns is the caller's
own value, not one every context reads.

**`mutatingMemberNames` is a starting vocabulary**, like every list in this
repository. A container the language adds later, or a `WeakMap`-shaped type
whose writer is spelled differently, passes until it is listed.

**One diagnostic per export.** The first mutable member found is reported and
the rest of the type is not walked.

**Deep immutability is not checked.** `Object.freeze({ pledges: [] })` has a
`readonly pledges` whose array is mutable, and passes.

**Scope is the rule's own**, from the `files` above. The rule judges the file it
is handed and carries no scope check of its own.
