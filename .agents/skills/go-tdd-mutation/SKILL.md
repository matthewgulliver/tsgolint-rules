---
name: go-tdd-mutation
description: |
  TDD and mutation-testing policy for Go projects. Load when writing or reviewing Go tests, running the RED-GREEN-REFACTOR loop in Go, or verifying test effectiveness with Gremlins. Covers behavior-testing through public interfaces (no internal/private testing), table-driven business-language tests, and the Gremlins mutation gate. For the general (language-agnostic) workflow load `tdd`, `testing`, or `mutation-testing` instead.
---

# Go TDD + Mutation Testing

Go synthesis of the `tdd`, `testing`, and `mutation-testing` skills. Same
ideas, Go-native mechanics: the standard library `testing` package,
table-driven tests, package-boundary selection, and Gremlins
(https://gremlins.dev) as the mutation harness.

## Core principle

**Test behavior, not implementation.** In Go, a package's public interface is
its exported API. Tests live in the same package (`package foo` internals
tests) or in `package foo_test` (black-box tests) — prefer `foo_test` when the
claim is about exported behavior; use same-package tests only when the public
interface genuinely cannot express the claim (unexported error paths worth
pinning), never as a license to test internals.

- Test exported functions and their observable results: return values,
  errors, emitted events, written state, calls to collaborators the caller
  supplied.
- Never test private helpers, internal state, or execution mechanics
  (goroutine counts, internal caching, call order the caller can't observe).
- A refactor that preserves exported behavior must not break tests. If it
  does, either the refactor changed behavior or the test was testing
  implementation.

## Behavior tests in Go

**Table-driven tests are the factory pattern.** Each case names a behavior in
business language; the code demonstrates it.

```go
func TestProcessPayment(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name string
        // inputs, expected outputs — whole values, not fragments
    }{
        // name states the business rule, not the call shape
        {name: "rejects negative amounts", ...},
        {name: "rejects amounts over the limit", ...},
        {name: "processes valid payments", ...},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // one whole-value assertion: got, want
        })
    }
}
```

Rules:

- **Whole-value assertions.** Compare the complete result (struct, error,
  message) — `cmp.Diff(got, want)` — not a field plucked out. Whole-value
  assertions kill more mutants and read as behavior.
- **Case names in business language**: "rejects expired tokens", not "case 3".
  On failure the test output documents the broken rule.
- **Edge cases as cases**: boundaries (`0`, `-1`, empty, max, exact-limit),
  error paths, nil handling — each an explicit table row.
- **No mocking of the subject.** Fakes/stubs only for collaborators the
  subject consumes, ideally supplied through the public API. Assert outcomes,
  never "was called" for internal collaborators the caller never provided.
- **No extraction just to mirror tests.** Don't split a function into helpers
  to give each a test file. Inline logic is covered by its consumer's
  behavior tests; extract only for a coherent reusable contract. No 1:1
  file/test mapping by reflex — organize tests around stable behavior.
- **Fresh state per test.** Table rows must not depend on each other;
  `t.Parallel()` by default; build inputs per-case, share only immutable
  values.
- **Reuse production constructors/schemas** for test data rather than
  redefining shapes in tests.

## TDD loop in Go

```
RED:      write the failing table case (or first case of a new table)
          go test ./<pkg>/ -run 'TestName/case_name'   # exact selector, prove it fails
GREEN:    minimum code to pass
          go test ./<pkg>/                             # owning package
REFACTOR: improve structure while green
GATE:     go test ./...                                # full module at phase end
MUTATE:   gremlins unleash ./<pkg>/                    # once per phase
```

Selector policy (from the testing skill, Go column): exact `-run` selectors
are for proving RED or debugging. GREEN/REFACTOR scope is the owning package
plus known consumers; the full `./...` run is the non-watch final gate. Go's
package graph gives the affected scope for free: `go test ./...` is cheap
enough to be the default gate in most modules.

During RED, scan the planned code against the mutator rules (below) to choose
strong cases cheaply — every conditional you're about to write should have
both sides represented in the table.

## Mutation gate: Gremlins

Gremlins is the Go mutation harness (install:
`go install github.com/go-gremlins/gremlins/cmd/gremlins@latest`). It gathers
coverage, mutates covered code only, and runs the package tests per mutant.

**Not in the inner loop.** Run once per phase/PR after implementation and
refactoring are complete:

```bash
gremlins unleash ./<pkg>/          # per package
gremlins unleash ./...             # whole module (rare, slow)
```

Optional config in `.gremlins.yml` at the module root (`unleash:` key):
`workers: 0` (auto), `timeout-coefficient: 2`, plus mutant-set toggles
(e.g. `invert-logical: true` for boolean-heavy code).

### Triage

| Status | Meaning | Action |
|---|---|---|
| KILLED | tests detect the mutant | none |
| LIVED | tests pass with a bug injected | missing/weak behavior test — add a table row that fails for the right reason |
| NOT COVERED | no test exercises the code | missing behavior case, or dead code |
| NOT VIABLE | mutant doesn't compile | none |
| TIMED OUT | mutant causes hang | often a killed-equivalent (missing loop/liveness assertion) — usually fine |

Survivor policy:

- **LIVED** on a conditional → a branch behavior is untested: add the
  missing table row (boundary, both boolean sides, non-identity values).
- **LIVED** on a return/call → weak assertion: strengthen to a whole-value
  comparison.
- **NOT COVERED** after behavior tests are complete → suspect **dead code**:
  remove it in a refactor step instead of writing coverage-theater tests.
- **Survivor no behavior test can kill, but the code must stay** (it exists
  to enable other behaviors — e.g. a defensive branch required by a caller's
  contract, an interface-satisfying stub, framework-required plumbing):
  **refactor so the untestable mechanism disappears into testable behavior.**
  Move the forced code behind a coherent exported contract, narrow its scope
  to the one place that requires it, or reshape it until every remaining line
  is observable through the public API. The resolution is always "make it
  testable or make it smaller" — never "leave it, it's load-bearing". If after
  honest reshaping a minimal residue genuinely cannot be observed, comment
  the constraint at the site and record it as an equivalent-mutant exception.
- **Equivalent mutants** (behavior genuinely unchanged) → note and move on;
  don't distort tests to chase them. When uncertain whether a survivor is
  equivalent, ask one concise question with the mutation and tradeoff stated.
- Priority goes to high-value behavior: money, permissions, eligibility,
  safety, data loss, and anything gating externally visible decisions.

Thresholds: none initially; establish a baseline first, then consider
`--threshold-efficacy` / `--threshold-mcover`. A score is a signal, not a
vanity metric.

## Proportionate evidence

Not every change needs RED or mutants (same carve-outs as the language-
agnostic policy):

- Pure refactor/reduction: starts from passing preservation tests, stays
  green; mutation gate optional.
- Configuration, wiring, generated, CI, or unreachable-by-design code:
  record reachability/operational evidence and `N/A` instead of fabricating
  RED or structural mutants.

## Checklist

- [ ] Tests target the exported API; no private-method or internal-state tests
- [ ] Table-driven; case names state the business rule
- [ ] Whole-value assertions (`cmp.Diff`), exact error expectations
- [ ] Boundaries and error paths are explicit table rows
- [ ] No mocks of the subject; fakes only for caller-supplied seams
- [ ] `t.Parallel()` unless state forbids it; rows independent
- [ ] RED proven with an exact `-run` selector; GREEN at package scope;
      gate at `go test ./...`
- [ ] `gremlins unleash` run at the phase gate; survivors triaged
      (row added / dead code removed / load-bearing code refactored to be
      testable or smaller / equivalent noted)
