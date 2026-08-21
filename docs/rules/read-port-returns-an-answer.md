# read-port-returns-an-answer

### What it does

Reports a **read port** — an exported declaration in the ports tree whose name
matches `readPortPatterns` (`…Rows`, `…Reader`, `…View`, `…Dashboard`) — with a
member whose return type, once `Promise` is unwrapped, is `void` or `undefined`.

### Why is this bad?

A member that answers nothing is called only for its effect. Declared on a read
port, that effect sits outside the consistency boundary an aggregate guards,
reachable from query policy no invariant protects — and it is the one shape a
reviewer skims past, because the port's name says the whole file is a question.

### Why it is type-aware, and how it differs from the JS rule

`arch/read-port-writes-nothing` judges the same
declarations by **member name**, against a list of write verbs. The two catch
different mistakes:

- `recordDashboardView(): Promise<void>` — both.
- `touch(id): Acknowledged`, where `type Acknowledged = Promise<void>` — only
  this rule. The annotation reads as a value.
- `refresh(id): Promise<void>` — only this rule. No write verb in the name.
- `saveWithOutbox(...): Promise<"saved" | "conflict">` on a `…Rows` port — only
  the JS rule. It answers, so the effect is invisible to the checker.

Neither subsumes the other, which is why they are two rules rather than one with
two message ids.

### Examples

Examples of **incorrect** code for this rule:

```ts
export interface OccasionDashboardRows {
  readonly recordDashboardView: (id: ContributorId) => Promise<void>
}
```

```ts
type Acknowledged = Promise<void>
export interface OccasionDashboardReader {
  readonly touch: (id: ContributorId) => Acknowledged
}
```

Examples of **correct** code for this rule:

```ts
export interface OccasionDashboardRows {
  readonly forContributor: (
    id: ContributorId
  ) => Promise<ReadonlyArray<OccasionDashboardRow>>
}
```

```ts
// An answer that may be absent is still an answer.
export interface OccasionDashboardView {
  readonly forContributor: (id: ContributorId) => Promise<OccasionCard | undefined>
}
```

```ts
// A write port is allowed to answer with nothing.
export interface PledgePersistence {
  readonly saveWithOutbox: (occasion: Occasion) => Promise<void>
}
```

### Options

These are the values the rule uses. They are **not** overridable through
lintcn today: `runner.go` passes `nil` options to every rule on every file,
so the defaults below are the shipped behaviour and the option names are a
test-only surface.

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/ports/**"]` | The trees this rule judges. |
| `readPortPatterns` | `string[]` | `["Rows$", "Reader$", "View$", "Dashboard$"]` | A declaration whose name matches one of these is treated as a read port. The same vocabulary `arch/read-port-writes-nothing` uses. |

### Limitations

**Read ports are found by name.** A read port called `OccasionQueries` is not
judged; extend `readPortPatterns`. Nothing here requires the naming.

**Returning `void` is the shape of a write, not proof of one**, and returning a
value is no proof of purity: a member that answers *and* writes passes. Effects
are not visible to this or any rule here.

**A union containing `undefined` is an answer**, deliberately — `Promise<Card |
undefined>` is a lookup that found nothing, not a write.

**One report per port**, naming the first answerless member.

**Scope is the rule's own**, from the `files` above. The rule judges the file it
is handed and carries no scope check of its own.
