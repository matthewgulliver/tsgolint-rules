# driving-port-command-is-modelled

Type-aware. Written in Go, run by `archlint`, **not** by
`oxlint`.

### What it does

Reports a **driving port** — a declaration in the ports tree whose name matches
`drivingPortPatterns`, `^For` by default — with a member whose parameter
*resolves* to a bare `string`, `number`, `boolean`, `any` or `unknown`.

### Why is this bad?

[`driving-port.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/application/driving-port.md) says the driving
port takes typed application commands. A method taking `(occasionId: string)`
accepts every string in the program: a contributor id, a slug, an empty string,
a value read straight off a query parameter. The port carries no vocabulary of
its own, so the checking that should have happened once at the boundary happens
in each caller, or nowhere.

### Why it is type-aware

This is the case a syntactic rule inverts. `(id: OccasionId)` and `(id: string)`
are the same written shape when `OccasionId` is an alias, and opposite shapes
when it is a brand:

```ts
type OccasionId = string & { readonly __brand: "OccasionId" } // modelled
type OccasionId = string                                      // bare
```

Only the checker separates them, and the whole point of a branded id is that it
is not a string at the boundary.

### Examples

Examples of **incorrect** code for this rule:

```ts
export interface ForPledgingToOccasions {
  readonly pledgeToOccasion: (occasionId: string) => Promise<PledgeResult>
}
```

```ts
// An alias to a primitive is still a primitive.
type OccasionId = string
export interface ForPledgingToOccasions {
  readonly pledgeToOccasion: (occasionId: OccasionId) => Promise<PledgeResult>
}
```

Examples of **correct** code for this rule:

```ts
export interface ForPledgingToOccasions {
  readonly pledgeToOccasion: (command: PledgeCommand) => Promise<PledgeResult>
}
```

```ts
type OccasionId = string & { readonly __brand: "OccasionId" }
export interface ForCancellingAnOccasion {
  readonly cancel: (id: OccasionId) => Promise<void>
}
```

```ts
// A literal union is a modelled vocabulary.
export interface ForRatingAPledge {
  readonly rate: (rating: "up" | "down") => Promise<void>
}
```

### Options

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/ports/**"]` | The trees this rule judges. |
| `drivingPortPatterns` | `string[]` | `["^For"]` | A declaration whose name matches one of these is treated as a driving port. |

### How to use

```bash
./archlint/build.sh
./archlint/build.sh --test
archlint
```

### Limitations

**Driving ports are found by name, and the docs only *prefer* the `For…` name.**
A driving port called `PledgeIntake` is not judged at all. The rule uses the
naming preference as a filter and does not enforce it — no rule here requires a
`For` prefix, because the driven doc's own gold-standard names have none. Where
a repository names driving ports another way, set `drivingPortPatterns`, or put
them under a `ports/driving/**` tree and set `files`.

**A driven port answering by id is untouched**, which is why the filter exists:
`findOccasionById(id: string)` is a legitimate driven-side lookup, and this
convention is not about it.

**One report per port**, naming the first bare parameter found.

**Scope is `archlint`'s**, from the `files` above. The rule judges the file it
is handed and carries no scope check of its own.
