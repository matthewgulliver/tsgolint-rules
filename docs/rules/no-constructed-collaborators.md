# no-constructed-collaborators

Type-aware. Written in Go, run by `npx lintcn lint`, **not** by
`oxlint`.

### What it does

Reports a `new` expression inside the hexagon whose constructed type is
**declared by a package dependency** — a class or interface whose declaration
file sits under `node_modules`.

It resolves the constructed expression to its type and asks which file declares
it. It does not read the name being constructed.

### Why is this bad?

[`command-use-case.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/application/command-use-case.md) says a
use case constructs no dependency: it receives its collaborators as ports so a
test can substitute a fake and the composition root stays the one place that
knows which provider is wired. A `new Pool()` in application policy pins the
file to that vendor, and the only way to test it is to run the vendor.

### Why it is type-aware, and why the syntactic version was dropped

This rule was first proposed as `no-constructed-collaborators` over `new` alone,
and [dropped](README.md#rejected-candidates): the application tree is full of
legitimate `new Date()`, `new Map()`, `new URL()`, the matcher could not tell a
dependency from a value, and a rule that reports mostly false positives gets
switched off and stops judging the case it was written for.

The checker draws the line the syntax could not. `Date` and `Map` are declared
in `lib.*.d.ts`; `Pool` is declared in `node_modules/pg`. That is the whole
difference, it is a fact rather than a vocabulary, and no denylist of package
names has to be maintained for it.

### Examples

Examples of **incorrect** code for this rule:

```ts
import { Pool } from "pg"
export const createPledgingToOccasions = () => {
  const pool = new Pool()
}
```

```ts
// The alias hides the name, never the declaration.
import { Pool } from "pg"
const Connection = Pool
const pool = new Connection()
```

Examples of **correct** code for this rule:

```ts
const now = new Date()
const seen = new Map<string, number>()
const at = new URL("https://example.test")
```

```ts
export const createPledgingToOccasions =
  (persistence: PledgePersistence) => async (command: PledgeCommand) =>
    persistence.saveWithOutbox(command)
```

### Options

These are the values the rule uses. They are **not** overridable through
lintcn today: `runner.go` passes `nil` options to every rule on every file,
so the defaults below are the shipped behaviour and the option names are a
test-only surface.

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/hexagon/application/**", "**/hexagon/domain/**"]` | The trees this rule judges. `**` spans any number of segments. |

### How to use

```bash
npx lintcn lint            # build the binary if needed, then run it
cd .lintcn && TSGOLINT_SNAPSHOT_CWD=true go test ./...   # the rule tests
```

### Limitations

**A first-party provider wrapper is invisible.** A class this repository writes
that wraps `pg` is declared here, not in `node_modules`, and passes. The rule
judges who declared the type, which is a proxy for who owns it.

**It does not judge what the value is for.** A vendor value class — a decimal, a
duration, an id type from a package — is constructed the same way a connection
pool is, and reports the same. Whether the constructed thing is a collaborator
or a value is the modelling question, and this rule is honest about answering a
narrower one.

**Only `new`.** A factory call (`createPool()`) constructs a dependency without
the keyword and is not reported.

**Scope is the rule's own**, from the `files` above. The rule judges the file it
is handed and carries no scope check of its own.
