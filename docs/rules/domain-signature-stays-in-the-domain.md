# domain-signature-stays-in-the-domain

Type-aware. Written in Go, run by `npx lintcn lint`, **not** by
`oxlint`.

### What it does

Reports an exported function in the domain tree whose parameter or return
annotation names a type declared **outside the domain and the shared kernel**.

For every type name written in the signature — including the arguments of a
generic and the members of a type literal — it resolves the name to a symbol,
follows the import alias to the real declaration, and judges the file that
declaration lives in.

Two `messageId`s, because the two halves are different repairs:
`foreignParameterType` and `foreignReturnType`.

### Why is this bad?

[`domain-service.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/domain/domain-service.md) §2 is that the
service "takes domain values and returns domain results only", and §5 that
"everything it decides with arrives as an argument". A signature is where that
promise is kept or broken.

Once a domain function takes an application port, the model has a collaborator:
only a caller holding that port can ask for the decision, the test needs a
double, and the rule the function exists to enforce is no longer observable on
its own. Once it *returns* someone else's type, the model's answer is expressed
in a vocabulary it does not control, and a change over there is a change to the
model.

### Why it is type-aware

The declaration is in another file. `PledgePersistence` and `Occasion` are the
same syntax at the point of use, and only the resolved symbol says which tree
declared each. This is the one fact a per-file rule cannot have.

It also reads through an alias: `import type { X } from "…"` binds a local name
whose own declaration is the import specifier, so a rule that stops at the first
declaration learns nothing.

### Examples

Examples of **incorrect** code for this rule:

```ts
// packages/gifting/hexagon/domain/src/pledges/pledge-contribution.ts
import type { PledgePersistence } from "../../../application/src/ports/pledge-persistence"

export const pledgeContribution = (
  occasion: Occasion,
  persistence: PledgePersistence,
): Occasion => occasion
```

```ts
import type { StripeCharge } from "stripe"

export function settle(occasion: Occasion, charges: ReadonlyArray<StripeCharge>): Occasion {
  return occasion
}
```

Examples of **correct** code for this rule:

```ts
import type { Occasion, PledgeDecision } from "../occasions/occasion"
import type { Money } from "@/shared-kernel/money"

export const pledgeContribution = (
  occasion: Occasion,
  amount: Money,
): PledgeDecision => ({ success: false, reason: "funding-closed" })
```

```ts
// The standard library is not foreign — `aggregate-root.md` holds `pledgedAt: Date`.
export const settleAt = (occasion: Occasion, at: Date): Occasion => occasion
```

### Options

These are the values the rule uses. They are **not** overridable through
lintcn today: `runner.go` passes `nil` options to every rule on every file,
so the defaults below are the shipped behaviour and the option names are a
test-only surface.

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/hexagon/domain/**"]` | The trees this rule judges. |
| `ownFiles` | `string[]` | `["**/hexagon/domain/**", "**/shared-kernel/**"]` | The trees a domain signature may name. |

The shared kernel is in `ownFiles` because `Money` is the domain's own
vocabulary, published one level up —
[`shared-kernel.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/domain/shared-kernel.md) is the doc that
argues it.

### Limitations

**The standard library is always allowed, and is not configurable.** Without
that carve-out the rule flags `Date`, `ReadonlyArray` and `Promise`, which is to
say every domain type in the repository. It is the trap the
[domain-service ledger](https://github.com/matthewgulliver/typescript-examples/blob/main/reports/decompose-domain-service/ledger.md) named
when it proposed the rule, and it is designed around rather than left to
configuration.

**A package dependency's type *alias* is excused**, on the precedent
[`no-provider-type-in-signature`](no-provider-type-in-signature.md) already
records: `z.infer<typeof Schema>` is declared in `node_modules` and describes a
shape this repository owns. A class or interface from a dependency is nominal
and is not excused.

**It overlaps `no-provider-type-in-signature` on purpose, and only partly.**
That rule denies a package dependency's class or interface anywhere in the
hexagon, parameters only. This one requires a *domain-declared* type, covers the
return position too, and catches the case that rule cannot see at all — an
application port, which no package declares. A dependency's interface in a
domain parameter reports under both.

**Only exported declarations**, and only function declarations and `const`
function initialisers. A module-private helper is a local choice.

**The outer signature only.** A curried domain function's inner parameters are
not read; the domain in these docs does not curry.

**Another context's domain tree satisfies `ownFiles`.** `**/hexagon/domain/**`
does not distinguish contexts, so a type from a *different* context's domain
passes here. That crossing is
[`context-model-does-not-cross-the-boundary`](context-model-does-not-cross-the-boundary.md)'s
subject, not this rule's.

**Scope is the rule's own**, from the `files` above. The rule judges the file it
is handed and carries no scope check of its own.
