# domain-probe-returns-void

Type-aware. Written in Go, run by `npx lintcn lint`, **not** by
`oxlint`.

### What it does

Reports an exported port declaration whose name matches `probeNamePatterns`
(`…Instrumentation` or `…Probe`) carrying a callable member whose resolved
return type is not `void` or `undefined`. Reported once per declaration: the
diagnostic names the first offending member, because the report lands on the
declaration and a report per member would stack diagnostics on one span. `Promise<void>` reports: a Domain Probe is
fire-and-forget, not eventually answerless.

### Examples

Examples of **incorrect** code for this rule:

```ts
export interface PledgeInstrumentation {
  readonly pledgeAccepted: (facts: PledgeFacts) => Promise<void>
}
```

Examples of **correct** code for this rule:

```ts
export interface PledgeInstrumentation {
  readonly pledgeAccepted: (facts: PledgeFacts) => void
  readonly pledgeRefused: (facts: PledgeFacts) => void
}
```

### Options

These are the values the rule uses. They are **not** overridable through
lintcn today: `runner.go` passes `nil` options to every rule on every file,
so the defaults below are the shipped behaviour and the option names are a
test-only surface.

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `["**/hexagon/application/**/ports/**"]` | The files this rule judges. |
| `probeNamePatterns` | `string[]` | `["(?:Instrumentation\|Probe)$"]` | An exported declaration whose name matches is treated as a Domain Probe. |

### Limitations

Probe selection is naming-based; a differently named probe is not judged until
its pattern is configured. The rule checks a declared return type only. It
cannot prove that callers do not await, branch on, or otherwise infer from a
probe call.
