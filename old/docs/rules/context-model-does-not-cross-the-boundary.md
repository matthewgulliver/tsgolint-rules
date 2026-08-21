# context-model-does-not-cross-the-boundary

Type-aware. Written in Go, run by `archlint`, **not** by
`oxlint`.

### What it does

Reports an exported function, `interface` or `type` in one bounded context whose
signature names a type **declared inside another context's internals** — however
that type arrived.

A context's identity comes from a `(?<name>…)` capture over the file path, the
same decision
[`no-cross-context-internal-import`](no-cross-context-internal-import.md)
records: one context spelled two ways must capture the same text, or an
intra-context reference reads as a crossing.

`messageId`: `crossContextModelType`.

### Why is this bad?

[`bounded-context.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/domain/bounded-context.md) §2.3: an
Anti-Corruption Layer exists so that an external model's types never leak into
yours. §2.4 adds that what a context publishes is versioned and kept stable.

A type that crosses makes the two models one. The other context can no longer
rename a field without breaking you, which is the freedom the boundary was drawn
to give it, and the "breakthroughs" the doc quotes Evans on stop being
affordable.

### Why it is type-aware

**Because the import can be legal and the leak still happen.**
`no-cross-context-internal-import` judges the module specifier, and
`@acme/gifting/public` is the front door. If that front door merely
*re-exports* a type declared in `@acme/gifting/src/occasions/occasion.ts`, the
specifier rule is satisfied and the model has crossed anyway. Only the resolved
symbol says where the declaration actually lives.

That is also the line this rule draws: a type **declared in** a public contract
file is the surface that context offered, and passes. A type merely re-exported
through one does not — the declaration is still in the internals, which is the
leak §2.3 names.

### Examples

Examples of **incorrect** code for this rule:

```ts
// packages/notifications/src/reminders/reminder.ts
import type { Occasion } from "@acme/gifting/public" // a legal import…

// …of a type `public.ts` only re-exports from `src/occasions/occasion.ts`.
export const remindAbout = (occasion: Occasion): Schedule => ({ at: occasion.id })
```

```ts
export type Reminder = { readonly about: Occasion }
```

Examples of **correct** code for this rule:

```ts
// `OccasionSummary` is declared in `packages/gifting/public.ts` itself.
import type { OccasionSummary } from "@acme/gifting/public"

export const remindAbout = (summary: OccasionSummary): Schedule =>
  ({ at: summary.id })
```

```ts
// The shared kernel belongs to no context.
import type { Money } from "@acme/shared-kernel/public"
export const owed = (amount: Money): Money => amount
```

### Options

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/packages/*/**"]` | The trees this rule judges. |
| `contextRootPatterns` | `string[]` | `["(^\|/)packages/(?<name>[^/]+)(/\|$)"]` | Each captures a context's identity in the group named `name`. |
| `publicContractPathPatterns` | `string[]` | `["/public\\.tsx?$", "/index\\.tsx?$", "/public/"]` | A file another context may declare its published types in. |
| `sharedFiles` | `string[]` | `["**/shared-kernel/**"]` | Trees belonging to no context. |

**The patterns are not spelled identically to the JavaScript rule's, and cannot
be.** Go's regexp is RE2 and has no lookaround, so
`(^|/)packages/(?<name>[^/]+)(?=/|$)` becomes
`(^|/)packages/(?<name>[^/]+)(/|$)`. Only the capture is read, so the trailing
group costs nothing — but the two rules' `contextRootPatterns` have to be
configured separately and kept in step.

### Limitations

**A file no pattern identifies is not judged.** The rule returns before
comparing anything, so a `contextRootPatterns` entry whose `name` group does not
exist switches this rule off and shows a green run. That is the shipped
precedent and the failure
[`reports/decompose-shared-kernel/ledger.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/reports/decompose-shared-kernel/ledger.md)
records having actually hit. Whether a pattern that captures nothing should be a
config error rather than a silent pass is still open.

**Two unidentified files compare equal.** Both answer `""`, so a repository with
no `packages/` layout has one implicit context rather than none.

**A package dependency and the standard library are never crossings.** Judging
them is [`no-provider-type-in-signature`](no-provider-type-in-signature.md)'s
job and this rule stays off it.

**Signatures and published type declarations only.** A type named inside a
function body, or inferred rather than written, is not read. The subject is what
the context *publishes*, and inference is the case this rule does not yet reach.

**Only exported declarations.** A module-private signature is not the context's
surface.

**A context that re-exports its whole model through `public.ts` will light this
rule up in every consumer.** That is the finding, not a false positive — but it
is a big finding, and worth landing deliberately rather than in the same change
as anything else.

**Scope is `archlint`'s**, from the `files` above. The rule judges the file it
is handed and carries no scope check of its own.
