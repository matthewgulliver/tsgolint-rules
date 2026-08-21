# domain-type-is-declared-once

### What it does

Reports an exported `interface` or `type` in the domain tree whose name is also
exported by **another file in the same bounded context's domain tree**.

It builds one index per program — every top-level exported type name in the
judged tree, and the files declaring it — and compares within a context.

`messageId`: `redeclaredDomainType`.

### Why is this bad?

`aggregate-root.md` states it on the
type itself: the canonical `Occasion` is imported everywhere and redefined
nowhere.

Two declarations of one concept both compile, and nothing ever forces them to
agree. One grows a field; the other does not. Code converts between them with a
spread that silently drops what the other side needed, and the model's own
definition becomes a question of which file you happened to import from — which
is the ubiquitous-language failure the DDD docs are about, arriving as a
type-shaped bug.

### Why it is type-aware

**The second declaration is in another file, and a per-file linter sees one
file.** That is the whole reason this could not be an `arch/*` rule: nothing in
`occasion.ts` looks wrong, and nothing in `pledges/occasion.ts` looks wrong
either. Only the program has both.

### It compares within one context, deliberately

`bounded-context.md` §2.1 is the
language test: the same word meaning different things in two contexts is the
boundary working. Gifting's `Occasion` and Notifications' `Occasion` are two
concepts, and reporting them would be arguing against the doc that motivates the
whole `packages/*` layout.

### Examples

Examples of **incorrect** code for this rule:

```ts
// packages/gifting/hexagon/domain/src/occasions/occasion.ts
export type Occasion = { readonly id: OccasionId; readonly budget: Money }
```
```ts
// packages/gifting/hexagon/domain/src/pledges/occasion.ts
export type Occasion = { readonly id: string; readonly closed: boolean }
```

Examples of **correct** code for this rule:

```ts
// packages/notifications/hexagon/domain/src/occasions/occasion.ts
// Another context's `Occasion` is another concept — the language test working.
export type Occasion = { readonly id: string; readonly remindAt: string }
```

```ts
// One declaration, imported everywhere else.
export type Occasion = { readonly id: OccasionId; readonly budget: Money }
```

### Options

These are the values the rule uses. They are **not** overridable through
lintcn today: `runner.go` passes `nil` options to every rule on every file,
so the defaults below are the shipped behaviour and the option names are a
test-only surface.

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/hexagon/domain/**"]` | The trees this rule judges, and the trees it indexes. |
| `contextRootPatterns` | `string[]` | `["(^\|/)packages/(?<name>[^/]+)(/\|$)"]` | Each captures a context's identity in the group named `name`. |

Spelled without lookaround, for the reason
[`context-model-does-not-cross-the-boundary`](context-model-does-not-cross-the-boundary.md)
records.

### Limitations

**Top-level exported declarations only.** A nested or unexported type is not the
model's published vocabulary and cannot be the canonical one. A type reached
only through `export { X } from …` is indexed where it is declared, not where it
is re-exported — which is what makes a barrel harmless here.

**Two unidentified files compare equal.** Both answer `""` for their context, so
a repository with no `packages/` layout has one implicit context rather than
none, and the rule still works.

**It reports both declarations, once each.** Neither is the canonical one as far
as this rule knows; deciding which file owns the concept is the human part.

**Name equality, not shape equality.** Two identical types under different names
are a duplicate this rule cannot see, and two unrelated concepts that happen to
share a name inside one context report — which is the rule working, since one
context is meant to have one meaning per word.

**The index is built once per program and cached.** A rule that rebuilt it per
file would be quadratic over the repository.

**Scope is the rule's own**, from the `files` above — and this is the one rule
the rule hands its resolved tree to, because the tree is also what the index
reads. Every other rule judges the file it is handed and never sees a glob.
