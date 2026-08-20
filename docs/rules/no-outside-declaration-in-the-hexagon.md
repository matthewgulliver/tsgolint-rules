# no-outside-declaration-in-the-hexagon

Type-aware. Written in Go, run by `archlint`, **not** by
`oxlint`.

### What it does

Reports an import inside the hexagon whose binding — after aliases are followed —
is *declared* in an outside tree: an adapter, a host app or the composition root.
It reads every named, default and namespace import specifier, and judges the
target module of a side-effect `import "…"` and of an `export … from "…"`. Type
imports count.

It is the type-aware complement of
[`no-driven-import-in-query-use-case`](no-driven-import-in-query-use-case.md) and
[`no-driving-adapter-import-in-use-case`](no-driving-adapter-import-in-use-case.md),
which read the specifier text: a tsconfig `paths` alias, a workspace package
name resolving to a repository path, and a barrel that re-exports an adapter all
land on text those rules cannot test and on a declaring file this rule can. A
plain relative import reports under both, which is fine — the messages agree.

### Why is this bad?

[`driven-port.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/application/driven-port.md) puts ports inside
beside their consumer and adapters outside, and
[`route-handler.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/adapters/route-handler.md) says the adapter
calls the application and is never called by it. An inside file that depends on
an outside declaration — however the import was spelled — can no longer be
driven by a test, a CLI or a second transport without that adapter, and a change
to the adapter now breaks policy.

### Why it is type-aware

The specifier is not the fact; the declaring file is. `import { Repo } from
"./index"` resolves through the barrel to `adapters/driven/postgres/src/repo.ts`
only in the checker, and `@gifting/adapters-postgres` resolves through `paths`
only in the program.

### Examples

Incorrect:

```ts
import { PostgresPledgeRepository } from "@gifting/adapters-postgres"
import { PostgresPledgeRepository } from "./index" // a barrel re-exporting the adapter
import type { StoredRow } from "../../../adapters/driven/postgres/src/repo"
export { PostgresPledgeRepository } from "../../adapters/driven/postgres/src/repo"
```

Correct:

```ts
import type { PledgePersistence } from "./ports/pledge-persistence"
import { z } from "zod"
```

### Options

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/hexagon/application/**", "**/hexagon/domain/**"]` | The inside trees the rule judges. |
| `outsideFiles` | `string[]` | `["**/adapters/**", "**/apps/*/src/**", "**/composition/**"]` | Trees whose declarations may not be reached from inside. |

### Limitations

A dynamic `import()`, `require`, and a value reached through a global or a
callback handed in at composition are not imports and are not read. A
declaration under `node_modules` never reports, whatever its path contains, so a
workspace package that resolves into `node_modules` rather than to a repository
path is invisible — configure `paths` or let the JS specifier rules name it.
Application test files are inside the default `files` and are judged; a
use-case test that imports an adapter is reported, which is the convention.
