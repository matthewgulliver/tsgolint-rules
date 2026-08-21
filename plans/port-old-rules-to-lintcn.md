# Plan: Port `old/rules` (19 tsgolint rules) to lintcn `.lintcn/` rules

## Status

- Phase 0 done: `.lintcn/` workspace bootstrapped and verified end-to-end
  (`go build`, `go test`, `npx lintcn lint` all work). Root `tsconfig.json`
  added (scopes lint to `old/**/*.ts`).
- Seed rule `no_floating_promises` added for verification, then removed.

## Findings driving the design

1. **No `Files` glob support in lintcn registration.** The generated
   `wrapper/main.go` collects `[]rule.Rule` only. `archrule.Rule{Files: ...}`
   from the old fork cannot be registered as-is. Idiomatic lintcn replacement:
   gate inside `Run` by matching `ctx.SourceFile.FileName()` against the rule's
   glob patterns (small shared helper), defaulting patterns via rule options.
2. **No `archrule`, `archscope`, `archtypes` upstream.** These internal helper
   packages from the old fork must be rebuilt as a shared package inside the
   lintcn workspace module (e.g. `.lintcn/archkit/`), importable by all rules
   since the module is a child path of tsgolint.
3. **Fixtures exist:** `old/fixtures/<rule-name>/` holds a multi-file TS
   fixture tree per rule; Go tests use `rule_tester` with
   `fixtures.GetRootDir()` + `tsconfig.minimal.json` (fixture infra inside the
   cached tsgolint source — usable as-is).
4. **Stale-lock gotcha:** aborting `lintcn lint/build` leaves a lock dir under
   `~/.cache/lintcn/locks/build/<hash>` that blocks all later runs. Remove the
   dir (or `npx lintcn clean`) when a run is aborted.

## Phase 1 — Rebuild shared helpers as `.lintcn/archkit/` (TDD)

One Go package, subfiles by concern; table-driven unit tests before rules use them.

### archtypes (type helpers over checker/utils) — ported signatures:
`Constituents`, `TypeReferenceNames`, `Members`, `CallSignatures`,
`ReturnType`, `IsCallable`, `DeclaringFiles`, `DeclaringFilesOfSymbol`,
`DeclaredType`, `Unwrapped`, `IsVoidLike`, `DeclaredUnder`, `WrittenName`,
`ElementTypes`

### archscope (file/module scoping)
`IsPackageDependency`, `IsStandardLibrary`, `Includes` (glob match),
`ContextOf`, `Compile`

### archrule replacement
`Gate(files []string) func(ctx) bool` — glob check on
`ctx.SourceFile.FileName()`; each gated rule early-returns in its listeners.

Where implementations are non-obvious, derive behavior from the old rules'
usage + their tests (the old `_test.go` files are the spec).

## Phase 2 — Port rules in dependency-risk order (TDD per rule)

### Commit discipline (one rule per commit)

Every commit is a labelled example (see ~/repos/personal/commit-rules):

- **One rule port per commit, and the commit is complete**: the new
  `.lintcn/<rule>/` (rule + tests + snapshots), the helper code it needs, and
  the **deletion of `old/rules/<rule>/` (and its fixture docs/tests that
  exist only for it) in the same commit**. After each commit, no trace of the
  old rule remains — the port is a move + adaptation, and the commit's diff
  shows exactly what changed while moving.
- Commit message states the rule name in the subject (the message is the
  index) and, where the port was not a pure move, a short body on what
  changed and why (e.g. "archrule.Files gating replaced with in-Run glob
  check because lintcn registration has no Files field").
- Never mix two rule ports or a rule port with unrelated tooling changes;
  helper-only refactors get their own commits.
- Order still follows the dependency-risk sequence below so every commit
  builds and tests green on its own (`go build ./... && go test ./...`).

Per rule (RED→GREEN): copy rule + test into `.lintcn/<snake>/` → run
`go test -run TestX` (fails on missing helpers) → port/adapt logic → green →
`UPDATE_SNAPS=true` → read snapshot to verify message quality → add
`lintcn:name` / `lintcn:description` / `lintcn:severity` metadata → commit
with the old rule deleted.

Order:
1. **Checker-only rules** (no archscope): `stored_state_switch_has_a_throwing_default`,
   `use_case_result_is_discriminated`, `use_case_throws_a_domain_error`,
   `no_page_request_in_journey` (needs only `DeclaringFiles` + glob gate).
2. **archscope-dependent hexagon rules**: `no_outside_declaration_in_the_hexagon`,
   `domain_signature_stays_in_the_domain`, `domain_type_is_declared_once`,
   `domain_state_is_deeply_readonly`, `domain_function_returns_an_answer`,
   `domain_probe_returns_void`, `port_behaviour_is_an_interface`,
   `driving_port_command_is_modelled`, `read_port_returns_an_answer`,
   `no_provider_type_in_signature`, `published_contract_publishes_no_mutable_value`,
   `no_constructed_collaborators`, `context_model_does_not_cross_the_boundary`.
3. **Glob-gated test rules last**: `no_double_library_in_domain_test`,
   `no_double_library_in_use_case_test`.

### Severity defaults (least surprise)
- Architectural invariants (hexagon/domain/port rules): `error` (default —
  omit the directive).
- Advisory/contextual rules (`no_page_request_in_journey`, double-library
  test rules): `// lintcn:severity warn`.
Revisit per rule after seeing violation volume on `old/fixtures/`.

## Phase 3 — End-to-end verification + final cleanup commit

1. `cd .lintcn && go build ./... && go test -v ./...` — all green.
2. `npx lintcn lint` against `old/fixtures/` trees — each rule fires on its
   fixture's violations; glob gating respected; severities correct.
3. Spot-check snapshots for message quality (they are the agent-facing output).
4. Final commit deletes whatever remains under `old/` once empty (fixtures
   trees move or go with their last consumer; `old/archlint`, `old/internal`
   go with the last rule that consumed them) — the tree ends with no dead
   `old/` residue.
5. Decide: keep `tsconfig.json` scoped to fixtures, or add a demo lint target.

## Open questions (resolved by default)

- Where helpers live → `.lintcn/archkit/` in the lintcn workspace module.
- Files glob → in-rule gating via helper; no registration-level support.
- Severity → error for invariants, warn for advisories (table above).
