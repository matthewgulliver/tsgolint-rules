# use-case-result-is-discriminated

Type-aware. Written in Go, run by `npx lintcn lint`, **not** by
`oxlint`.

### What it does

Two failures on one subject: the resolved result of an exported function
**inside the hexagon** — the application tree and the domain tree.

`undiscriminatedResult` reports a result that is a union of two or more object
shapes with **no property that is a literal type in every member**.

`genericFailureReason` reports a result any of whose shapes carries a member
named in `failureReasonMemberPatterns` whose resolved type **admits any
string**: a bare `string`, an alias of it, a literal widened by a `string`
sibling, or an array of any of those. Inferred results count, so `return { ok:
false, error: message }` with no annotation reports.

It resolves the declaration's type, follows a curried use case through each
call to the outcome the caller finally receives, unwraps `Promise`, drops
`null` and `undefined` constituents, and then reads the members.

### Why is this bad?

The use-case docs return expected outcomes as a discriminated union and never
throw them. That only pays off if a caller can *narrow*: with a discriminant,
`switch (result.outcome)` narrows each branch, and a new outcome added to the
union breaks every incomplete `switch` at compile time. Without one, callers
probe for fields — `if ("occasionId" in result)` — a new outcome compiles
everywhere, and the surface that forgot it fails at runtime. A failure member
that admits any string defeats the same mechanism from inside one branch:
`error-handling.md`'s "never a generic `error: string`" — the caller cannot
switch on it and the compiler cannot prove every reason is handled.

An array of them is the same failure once per element. `readonly string[]` is a
reference type rather than a string-flagged one, so the elements have to be
read: a `{ outcome: "refused"; violations: readonly string[] }` is discriminated,
deeply readonly, and still tells the caller nothing it can branch on.

### Why it is type-aware

Union membership is the checker's own question, and the failure is invisible in
the source. These two are the same text and opposite contracts:

```ts
type Saved = { readonly outcome: "saved"; readonly occasionId: OccasionId }
type Saved = { readonly outcome: string; readonly occasionId: OccasionId }
```

The second discriminates nothing. Only the resolved type says which one is on
the page.

### The other half of the claim is already shipped

"…and handles it **exhaustively**" is not this rule. tsgolint's upstream
`switch-exhaustiveness-check` already runs in the same binary — it is
compiled into the same binary and needs no configuration here. Writing a second
one would be duplicating a shipped rule under a local name.

### The domain tree

Both [`aggregate-root.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/domain/aggregate-root.md) and
[`domain-service.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/domain/domain-service.md) make the same
claim one tree over — the aggregate's `PledgeCommandResult` ("every state
transition either returns a new valid occasion or an explicit business
refusal") and the domain service's `PledgeDecision` — so the domain tree is a
`files` entry, not a second rule. Both ledgers proposed it independently.

This is the **union half** of "returns the next state or an explicit refusal".
The return half is [`no-void-return-in-domain`](no-void-return-in-domain.md),
which is syntactic: it reads a written `void` annotation. A transition whose
`void` is inferred rather than written escapes that rule, and is not this one's
subject either — a `void` return is not a union.

Both docs' Perfect Examples discriminate on a boolean literal
(`success: true` / `success: false`) and pass unchanged.

**A factory is not this rule's subject.** Both precedents above are state
transitions. A factory returns its value or throws — an invariant violation is a
bug, not an outcome a caller handles — so there is no union to discriminate.
`createMoney` in [`shared-kernel.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/domain/shared-kernel.md) is
the tree's one factory and it throws, per
`domain-driven-design/resources/error-modeling.md:21-48`. A factory written as
`T | Refusal` instead is a modelling mistake this rule will not catch for you:
give it a discriminant and it passes, because the discriminant is all this rule
reads.

### Examples

Examples of **incorrect** code for this rule:

```ts
export const pledgeToOccasion = async (): Promise<
  { readonly occasionId: OccasionId } | { readonly reason: string }
> => ({ occasionId })
```

```ts
// A shared key widened to `string` narrows nothing.
type Saved = { readonly outcome: string; readonly occasionId: OccasionId }
type Conflict = { readonly outcome: string }
export function pledgeToOccasion(): Saved | Conflict { … }
```

```ts
// A generic failure text, hidden behind an alias the annotation never shows —
// and the same result with no annotation, inferred from `message`.
type FailureText = string
type PledgeResult =
  | { readonly ok: true }
  | { readonly ok: false; readonly error: FailureText }
export const pledgeToOccasion = (message: string) =>
  message ? { ok: false as const, error: message } : { ok: true as const }
```

Examples of **correct** code for this rule:

```ts
type PledgeResult =
  | { readonly outcome: "saved"; readonly occasionId: OccasionId }
  | { readonly outcome: "conflict" }

export const createPledgingToOccasions =
  (persistence: PledgePersistence) =>
  async (command: PledgeCommand): Promise<PledgeResult> => { … }
```

```ts
// An absent answer is not an undiscriminated union.
export const findOccasion = async (): Promise<Occasion | null> => null
```

### Options

These are the values the rule uses. They are **not** overridable through
lintcn today: `runner.go` passes `nil` options to every rule on every file,
so the defaults below are the shipped behaviour and the option names are a
test-only surface.

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/hexagon/application/**", "**/hexagon/domain/**"]` | The trees this rule judges. |
| `failureReasonMemberPatterns` | `string[]` | `["^error$", "^violations$", "^faults$", "^problems$"]` | Result members whose resolved type may not admit any `string`. The three plurals are the names a batch-shaped refusal takes, and no Perfect Example or skill declares one. `reason` is left to configuration: [`bounded-context.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/domain/bounded-context.md) and [`domain-probe.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/application/domain-probe.md) declare `reason: string` in their own Perfect Examples. `errors` is deliberately absent — it is RFC 9457's validation-error array, endorsed by example in `api-design/resources/problem-details.md:51-54`. [`no-generic-failure-reason`](no-generic-failure-reason.md) draws the same lines syntactically. |

### Limitations

**Only exported declarations**, and only function declarations and `const`
function initialisers. A module-private helper returning a union is a local
choice.

**A union containing a non-object constituent is skipped entirely** by the
discriminant arm. `Occasion | "not-found"` never reports — mixing a shape with a
literal is a different convention, and reporting it here would say the wrong
thing. The failure-reason arm still reads the object shapes of such a union.

**`boolean` counts as a discriminant** — `{ success: true } | { success: false }`
passes, because `true` and `false` are literal types and that union narrows
correctly.

**It cannot see whether callers narrow.** A perfectly discriminated union whose
every caller probes with `in` reports nothing. Exhaustiveness at the call site is
`switch-exhaustiveness-check`'s job.

**Curried chains are followed eight calls deep**, which is seven more than any
use case in the docs. **Nested collections are followed four deep**, and only
arrays and tuples count as collections — a `ReadonlySet<string>` reports
nothing.

**Scope is the rule's own**, from the `files` above. The rule judges the file it
is handed and carries no scope check of its own.
