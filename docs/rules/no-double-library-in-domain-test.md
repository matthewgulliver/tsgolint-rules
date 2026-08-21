# no-double-library-in-domain-test

Type-aware. Written in Go, run by `npx lintcn lint`, **not** by
`oxlint`.

### What it does

Reports `fn`, `mock`, and `spyOn` calls in a domain test when their receiver
resolves to a configured test-double library such as Vitest or Jest. A local
object with the same method name is not reported.

### Why is this bad?

[`domain-test.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/tests/domain-test.md) defines its subject as
a pure aggregate command or domain service: it has no collaborators to fake.
A mock-library double in that test is evidence that collaboration leaked into
the domain boundary.

### Why it is type-aware

The rule resolves the receiver symbol through aliases to its declaration file.
That catches `import { vi as doubles } from "vitest"` without banning an
unrelated local helper named `vi`.

### Examples

Incorrect:

```ts
import { vi } from "vitest"
const persistence = vi.fn()
```

Correct:

```ts
const result = pledgeOccasion(occasion, command)
expect(result).toEqual({ ok: true, value: expected })
```

### Options

These are the values the rule uses. They are **not** overridable through
lintcn today: `runner.go` passes `nil` options to every rule on every file,
so the defaults below are the shipped behaviour and the option names are a
test-only surface.

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | domain `*.test.ts(x)` files | Tests the rule judges. |
| `doubleMethodNames` | `string[]` | `[`"fn"`, `"mock"`, `"spyOn"`]` | Double-library methods to reject. |
| `doubleLibraryPathFragments` | `string[]` | Vitest/Jest declaration paths | Declaration-file fragments identifying a double library. |

### Limitations

It does not prove that every test subject is called directly or that a hand-made
object is not a double. It also does not judge assertions. Those remain runtime
test-design evidence.
