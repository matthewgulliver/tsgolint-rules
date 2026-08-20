# no-double-library-in-use-case-test

Type-aware. Written in Go, run by `archlint`, **not** by
`oxlint`.

### What it does

Reports `fn`, `mock`, and `spyOn` calls in a use-case test — a `*.test.ts(x)`
under `**/hexagon/application/**` — when their receiver resolves to a configured
test-double library such as Vitest or Jest. A local object with the same method
name is not reported. Sibling of
[`no-double-library-in-domain-test`](no-double-library-in-domain-test.md): same
matcher, one tree over, a different repair.

### Why is this bad?

[`driven-port.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/application/driven-port.md) says every driven
port gets a stateful in-memory fake, *never a mock*, and lists "doubled by
mock-verifying call sequences" as an anti-pattern;
[`driving-port.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/application/driving-port.md) refuses a
mock-verifying test driver; [`in-memory-fake.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/tests/in-memory-fake.md)
names the double for its role — `createFake…`, `createStub…`, `createSpy…` —
never for the library that made it, because a hand-written double implements
the port's real interface and a contract change breaks at compile time. A
`vi.fn()` standing in for a port is scaffolding that a renamed member never
reaches.

This reverses one row of
[`reports/decompose-use-case-test/ledger.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/reports/decompose-use-case-test/ledger.md)
and [`reports/decompose-in-memory-fake/ledger.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/reports/decompose-in-memory-fake/ledger.md)
— the library-double slice only. Whether a fake or a spy was the right double,
and whether an assertion is adequate, stay refused.

### Why it is type-aware

The rule resolves the receiver symbol through aliases to its declaration file, so
`import { vi as doubles } from "vitest"` reports and an unrelated local helper
named `vi` does not.

### Examples

Incorrect:

```ts
import { vi } from "vitest"
const persistence = { save: vi.fn() }
```

Correct:

```ts
const persistence = createFakePledgePersistence()
const result = await pledgeToOccasion(command)
expect(persistence.saved).toHaveLength(0)
```

### Options

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/hexagon/application/**/*.test.ts", "**/hexagon/application/**/*.test.tsx"]` | Tests the rule judges. |
| `doubleMethodNames` | `string[]` | `["fn", "mock", "spyOn"]` | Double-library methods to reject. |
| `doubleLibraryPathFragments` | `string[]` | `["/node_modules/vitest/", "/node_modules/jest/", "/node_modules/@types/jest/"]` | Declaration-file fragments identifying a double library. |

### Limitations

It judges the library, not the double: a hand-written object that verifies
call sequences passes, and so does a double library not in the fragment list.
It does not judge assertions, does not prove the doubled thing was a port, and
does not see a double built in a helper file outside the test tree.
