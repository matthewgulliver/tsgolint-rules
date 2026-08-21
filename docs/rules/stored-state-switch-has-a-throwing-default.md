# stored-state-switch-has-a-throwing-default

### What it does

Reports a `switch` in a driven adapter whose discriminant is a property of a
**persistence row type** and which has no `default`, or whose `default` does not
`throw`.

Two `messageId`s: `storedStateSwitchWithoutDefault` and
`storedStateDefaultDoesNotThrow`.

### Why is this bad?

`aggregate-reconstitution.md`
writes the mapper's `default` as a `never` binding and a `throw`, and the reason
is not exhaustiveness. **The row type is a claim about the database, not a fact
about it.** A column typed `"Open" | "Settled"` can hold `"Archived"` — written
by an older deploy, a migration, or a hand-edited row — and the type system was
never in a position to prevent it.

Without a throwing `default`, that row falls out of the `switch` as `undefined`
and enters the model as an aggregate that was never valid. The corruption is
then somewhere else, in something that read a field that is not there.

### Why it is type-aware, and why upstream does not cover it

`switch-exhaustiveness-check` already runs in the same binary and is
**satisfied by exhaustive cases and no `default` at all** — which is exactly the
code this rule reports. The two rules want different things and both are right:
upstream wants every declared case handled, this one wants the undeclared case
to fail loudly.

The type question is provenance. These are the same syntax:

```ts
switch (row.state) { … }      // read out of a database
switch (command.state) { … }  // built by the model, and genuinely exhaustive
```

Demanding a throwing `default` of the second would be noise, and the only thing
separating them is where the type of the thing being read was declared.

### Examples

Examples of **incorrect** code for this rule:

```ts
// packages/gifting/hexagon/adapters/driven/postgres/src/occasion-repository.ts
const toOccasion = (row: OccasionRow): Occasion => {
  switch (row.state) {
    case "Open": return openOccasion(recorded)
    case "Settled": return settledOccasion(recorded)
  }
}
```

Examples of **correct** code for this rule:

```ts
const toOccasion = (row: OccasionRow): Occasion => {
  switch (row.state) {
    case "Open": return openOccasion(recorded)
    case "Settled": return settledOccasion(recorded)
    default: {
      const unknownState: never = row.state
      throw new Error(`Stored occasion ${row.id} has unknown state ${String(unknownState)}`)
    }
  }
}
```

```ts
// A command the model built. Upstream's exhaustiveness check is the right rule.
switch (command.state) {
  case "Open": return "open"
  case "Settled": return "settled"
}
```

### Options

These are the values the rule uses. They are **not** overridable through
lintcn today: `runner.go` passes `nil` options to every rule on every file,
so the defaults below are the shipped behaviour and the option names are a
test-only surface.

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/adapters/driven/**"]` | The trees this rule judges. |
| `rowTypeFiles` | `string[]` | `["**/adapters/driven/**"]` | Where a persistence row type is declared. A discriminant whose object resolves to a type declared here is a stored one. |

The two default to the same tree because that is where
`no-row-type-in-domain` keeps a row: the adapter
declares it, and the domain never names it.

### Limitations

**A property access only.** `switch (state)`, where the value was destructured
or assigned to a local first, carries no object to resolve and is not judged.
That is the common shape this rule misses, and the reason its report is a floor
rather than a guarantee.

**A `throw` anywhere in the `default` counts**, including behind an `if`. The
rule asks whether the clause can throw, not whether it always does — the second
question is control flow, not types. A `throw` inside a nested function
declared in the clause does *not* count.

**Provenance is approximated by the declaring file, not traced.** A row type
declared outside `rowTypeFiles` is invisible here, and a non-row type that
happens to be declared in the adapter tree is treated as stored. Real dataflow
— did this value come out of a query? — is not a type question, which is
why a sibling proposal about brand factories was refused rather than
approximated.

**It says nothing about what is thrown.** `use-case-throws-a-domain-error` is
the rule with an opinion there, and it does not judge this tree.

**Scope is the rule's own**, from the `files` above. The rule judges the file it
is handed and carries no scope check of its own.
