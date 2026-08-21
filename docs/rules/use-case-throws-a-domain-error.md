# use-case-throws-a-domain-error

Type-aware. Written in Go, run by `npx lintcn lint`, **not** by
`oxlint`.

### What it does

Reports a `throw` inside the hexagon whose thrown value resolves to a type
declared **outside this repository** — by a package dependency anywhere in the
hexagon, or by the TypeScript standard library (`Error`, `TypeError`,
`RangeError`) in the application tree only.

### Why is this bad?

The use-case docs give a use case two failure channels: expected outcomes come
back as members of the result union, and everything else propagates unhandled to
the edge that owns the provider. `throw new Error("occasion closed")` is neither.
It is a business outcome dressed as a crash: no surface can distinguish it from
a genuine fault, every one of them has to match on a message string to write
copy for it, and the type system says nothing at all.

A `pg` `DatabaseError` thrown from application policy is the same mistake
inverted — infrastructure vocabulary escaping into the inside.

### Why it is type-aware

"Requires knowing what was thrown" is the recorded reason the claim was refused
for a syntactic rule, and it is exactly right: `throw new Failure(…)` says
nothing about whether `Failure` is a domain error or an alias for `TypeError`.
The checker resolves the thrown expression to a type and the type to the file
that declares it.

The full claim — *is this throw an expected business outcome?* — stays refused.
That is a modelling judgement. This rule enforces the narrower fact underneath
it: whatever a use case throws, this repository declared it.

### Examples

Examples of **incorrect** code for this rule:

```ts
export const pledge = () => {
  throw new Error("occasion closed")
}
```

```ts
import { DatabaseError } from "pg"
export const pledge = () => {
  throw new DatabaseError("conflict")
}
```

```ts
// The alias hides the name, never the declaration.
const Failure = TypeError
throw new Failure("closed")
```

Examples of **correct** code for this rule:

```ts
export class OccasionAlreadyClosed extends Error {}
export const pledge = () => {
  throw new OccasionAlreadyClosed()
}
```

```ts
// The outcome the docs actually recommend: returned, not thrown.
return { outcome: "conflict" }
```

### Options

These are the values the rule uses. They are **not** overridable through
lintcn today: `runner.go` passes `nil` options to every rule on every file,
so the defaults below are the shipped behaviour and the option names are a
test-only surface.

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/hexagon/application/**", "**/hexagon/domain/**"]` | The trees this rule judges. The domain scope carries the driven doc's "the domain never catches HTTP errors, manages transactions or leaks a logger" family; the catch clause of that sentence is [`arch/no-catch-in-domain`](no-catch-in-domain.md). |
| `standardLibraryFiles` | `string[]` | `["**/hexagon/application/**"]` | Of those trees, the ones where a **standard-library** throw also reports. A repository that wants the stricter reading adds its domain tree. |

**Why the standard-library arm skips the domain by default.** A bare `Error` is
how the hexagonal skill itself rejects an impossible state: its worked example
throws one from `createMoney` while returning `exceeds-budget` as a result value
in the same function. An invariant violation is not a business outcome and has
no result union to live in, so reporting it would condemn the model answer the
domain docs are written from. A provider's error type reaching the domain still
reports — no skill sanctions that.

### Limitations

**A domain error class extending `Error` passes**, and should: the subclass is
declared here. Only the base class thrown directly reports.

**A rethrow is invisible.** `throw error` where `error` is `unknown` has no
declaration to judge, and guessing is how a rule starts reporting on nothing.

**It does not decide whether the throw was expected.** A domain-declared error
thrown where the docs would have returned an outcome passes. That is the
modelling half, and it stays with human review.

**Scope is the rule's own**, from the `files` above. The rule judges the file it
is handed and carries no scope check of its own.

**A business rule thrown as a bare `Error` in the domain now passes.** That is
the cost of letting invariant violations through: the two are the same syntax
and the same type, and only the modelling tells them apart. The application
tree still reports, which is where an expected outcome thrown instead of
returned actually crosses a boundary.
