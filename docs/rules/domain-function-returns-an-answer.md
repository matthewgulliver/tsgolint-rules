# domain-function-returns-an-answer

Type-aware. Written in Go, run by `npx lintcn lint`, **not** by
`oxlint`.

### What it does

Reports an **exported** function in the domain tree that carries no written
return type and whose *inferred* return type is `void` or `undefined` on every
path — a `function` declaration, or a `const` bound to an arrow or function
expression. A curried chain is followed to the type its last call answers with.

A signature with a written return type is left alone: the written `: void` is
[`no-void-return-in-domain`](no-void-return-in-domain.md)'s subject, and this
rule is its complement, so the two never report the same function.

### Why is this bad?

[`aggregate-root.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/domain/aggregate-root.md): every command
"either preserves the invariant or returns an explicit refusal". A function that
returns nothing decided something and put the answer somewhere the model cannot
show — a mutated argument, a provider it called — so the caller learns neither
the next state nor why the transition was refused, and the test has to reach
for whatever the function touched instead of what it returned.

### Why it is type-aware

`export const settle = (o: Occasion) => { audit(o) }` has no annotation to read.
Only the checker knows the block body answers with nothing.

### Examples

Examples of **incorrect** code for this rule:

```ts
// packages/gifting/hexagon/domain/src/occasions/occasion.ts
export const settle = (o: Occasion) => { audit(o) }

export function record(o: { total: number }) {
  o.total += 1
}
```

Examples of **correct** code for this rule:

```ts
export const settle = (o: Occasion) => ({ ...o, settled: true })

// A lookup that found nothing still answers.
export const find = (all: ReadonlyArray<Occasion>, id: string) => all.find((o) => o.id === id)

// Written `: void` is `no-void-return-in-domain`'s to report.
export const settleWith = (o: Occasion): void => { audit(o) }
```

### Options

These are the values the rule uses. They are **not** overridable through
lintcn today: `runner.go` passes `nil` options to every rule on every file,
so the defaults below are the shipped behaviour and the option names are a
test-only surface.

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/hexagon/domain/**"]` | The trees this rule judges. |

### Limitations

**`Promise<void>` is not unwrapped.** The domain is synchronous, and a promise
here is [`no-async-in-domain`](no-async-in-domain.md)'s report; a function that
answers with a promise of nothing passes this rule.

**A union with an answer passes.** `Occasion | undefined` is a lookup that
found nothing, and only a type whose every constituent is void-like reports.

**Only exported declarations**, and only function declarations and `const`
function initialisers. A module-private helper is a local choice.

**It does not say the returned value is the right one.** That the refusal is
explicit and the result discriminates is
[`use-case-result-is-discriminated`](use-case-result-is-discriminated.md)'s.

**Scope is the rule's own**, from the `files` above. The rule judges the file it
is handed and carries no scope check of its own.
