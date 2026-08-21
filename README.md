# tsgolint-rules

Type-aware architecture lint rules for TypeScript, written in Go against
[tsgolint](https://github.com/oxc-project/tsgolint)'s checker and distributed
as [lintcn](https://github.com/remorses/lintcn) rules.

They enforce hexagonal boundaries: what the domain may name, what a port may
declare, what crosses a context edge, and what a test may double.

## Install

```bash
npx lintcn add https://github.com/matthewgulliver/tsgolint-rules
npx lintcn lint
```

The repository is private; `lintcn add` needs credentials that can read it.

## Rules

Severity is `error` unless marked.

| Rule | Disallows |
| --- | --- |
| `context-model-does-not-cross-the-boundary` | an exported signature naming a type declared inside another context's internals |
| `domain-function-returns-an-answer` | an exported domain function with no written return type whose inferred return is `void` or `undefined` on every path |
| `domain-probe-returns-void` | a Domain Probe method returning anything — a probe announces a fact, it does not answer |
| `domain-signature-stays-in-the-domain` | a domain signature naming a type the domain does not own |
| `domain-state-is-deeply-readonly` | domain state writable deeper than the root guards |
| `domain-type-is-declared-once` | the same domain type declared twice in one context |
| `driving-port-command-is-modelled` | a driving port method taking bare primitives where a command belongs |
| `no-constructed-collaborators` | the inside of the hexagon constructing a provider a package dependency declares |
| `no-double-library-in-domain-test` *(warn)* | test-double library scaffolding in a domain test |
| `no-double-library-in-use-case-test` *(warn)* | test-double library scaffolding in a use-case test |
| `no-outside-declaration-in-the-hexagon` | the inside of the hexagon importing symbols declared in adapters, apps or composition |
| `no-page-request-in-journey` *(warn)* | `page.request` in an e2e journey — it bypasses the browser the journey exists to prove |
| `no-provider-type-in-signature` | a signature in the hexagon naming a type a package dependency or the transport owns |
| `port-behaviour-is-an-interface` | a port's behaviour declared as a type alias instead of an interface |
| `published-contract-publishes-no-mutable-value` | a published contract exposing a value any context can mutate |
| `read-port-returns-an-answer` | a read port method returning nothing instead of the answer |
| `stored-state-switch-has-a-throwing-default` | a driven-adapter switch on a persistence row property with no throwing `default` — the row type is a claim about the database, not a fact about it |
| `use-case-result-is-discriminated` | a use-case result a caller cannot narrow on |
| `use-case-throws-a-domain-error` | throwing an error type this repository does not declare inside the hexagon |

## Develop

One package per rule under `.lintcn/`, sharing `.lintcn/archkit/`.

```bash
cd .lintcn
go build ./...
TSGOLINT_SNAPSHOT_CWD=true go test ./...   # required: resolves __snapshots__ beside the rule
TSGOLINT_SNAPSHOT_CWD=true UPDATE_SNAPS=true go test ./...
gremlins unleash ./<rule>/
```

`docs/in-progress/port-old-rules-to-lintcn.md` holds the conventions a new
rule must follow; they are not in the vendored lintcn skill.
