# Plan: Port `old/rules` (19 tsgolint rules) to lintcn `.lintcn/` rules

## Status

- Phase 0 done: `.lintcn/` workspace bootstrapped and verified end-to-end
  (`go build`, `go test`, `npx lintcn lint` all work). Root `tsconfig.json`
  added (scopes lint to `old/**/*.ts`).
- Seed rule `no_floating_promises` added for verification, then removed.
- Phase 2 in progress. Ported and committed: `no_page_request_in_journey`,
  `use_case_throws_a_domain_error`, `stored_state_switch_has_a_throwing_default`,
  `use_case_result_is_discriminated`, `domain_function_returns_an_answer`,
  `domain_probe_returns_void`. Remaining work is owned by two parallel porter
  agents (`.opencode/agents/porter-hexagon.md`, `.opencode/agents/porter-tests.md`).

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

## Decisions (from the template-port review)

1. **`old/` is a frozen reference until Phase 3.** Do not delete anything
   under `old/` per-rule — `old/archlint` registers every rule, so piecemeal
   deletion kills the old binary and the old-vs-new fixture comparison from
   the first commit. Ported rules coexist with their old copies; Phase 3
   deletes `old/` wholesale in one final commit. (Commit discipline note:
   "delete the old rule in the same commit" is superseded by this.)
2. **Severity `warn` for advisory rules is a deliberate reversal** of the old
   fork's parseSeverity refusal ("a severity that reports without failing is
   a gate nobody notices going out"). lintcn's warn semantics — changed/untracked
   files only, never fails CI — is exactly right for agent-facing guidance on
   new code, which is the audience here. Architectural invariants stay
   `error` (default).
3. **Snapshots are load-bearing regression tests**, not decoration: run all
   rule tests with `TSGOLINT_SNAPSHOT_CWD=true`, which resolves
   `__snapshots__` next to the rule package (`UPDATE_SNAPS=true` likewise).
   Never hand-copy snapshots out of the tsgolint cache.
4. **Docs URLs are dropped deliberately.** `archrule.DocumentedAt` +
   `docs/rules/*.md` enforcement has no lintcn equivalent and `lintcn:source`
   is provenance only. Message `Help` text carries the guidance.
5. **archkit grows on demand.** Helpers port from `old/internal/archtypes`
   only when the first rule needing them lands. Wholesale ports leave
   unconsumed symbols (mutation NOT COVERED noise).
6. **`Name:` must be an inline string literal**, not a const — lintcn's
   discovery binds the CLI name by matching `Name: "..."` in source.
7. **Gremlins runs on every rule package at its port commit** (sub-second,
   `gremlins unleash ./<rule>/`), plus archkit.
8. **Configurable scope:** rules with a user-overridable tree parse their
   options first and pass the resolved pattern list to `archkit.Gated`
   (signature already takes `files []string`); defaults stay `var defaultFiles`.
9. Every gated rule's test table carries at least one **out-of-tree valid
   case** proving the gate, and `archkit` has `TestGated`.

## Phase 1 — Rebuild shared helpers as `.lintcn/archkit/` (TDD) — DONE

Ported from `old/internal/` (exists in-tree): scope.go (glob matching,
dependency/stdlib classification), types.go (trimmed to consumed symbols,
grows per decision 5), gate.go (`Gated` replaces `archrule.Rule{Files}`).
Mutation: archkit 100% efficacy (2 recorded cross-package survivors), rule
package 100%/100%.

## Phase 2 — Port rules in dependency-risk order (TDD per rule)

### Porting is parallelized across two subgroups (commits stay serial)

The remaining rules are split into two subgroups, each owned by a porter
agent (`.opencode/agents/`), both on `deepseek/deepseek-v4-flash`:

- **porter-hexagon**: `no_outside_declaration_in_the_hexagon`,
  `domain_signature_stays_in_the_domain`, `domain_type_is_declared_once`,
  `domain_state_is_deeply_readonly`, `port_behaviour_is_an_interface`,
  `driving_port_command_is_modelled`, `read_port_returns_an_answer`,
  `no_provider_type_in_signature`,
  `published_contract_publishes_no_mutable_value`, `no_constructed_collaborators`,
  `context_model_does_not_cross_the_boundary`.
- **porter-tests** (glob-gated, last): `no_double_library_in_domain_test`,
  `no_double_library_in_use_case_test`.

Each agent ports one rule at a time and reports without committing. Commits
still land one rule per commit, serialized on the same branch (no parallel
commits — the agents prepare work, a human or coordinating agent commits).

### Commit discipline (one rule per commit)

Every commit is a labelled example (see ~/repos/personal/commit-rules):

- **One rule port per commit, and the commit is complete**: the new
  `.lintcn/<rule>/` (rule + tests + snapshots) and any archkit helpers it
  consumes. `old/` stays frozen (decision 1) — no per-rule deletions.
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
`TSGOLINT_SNAPSHOT_CWD=true UPDATE_SNAPS=true` → read snapshot to verify
message quality → add
`lintcn:name` / `lintcn:description` / `lintcn:severity` metadata → out-of-tree
gate case → `gremlins unleash ./<rule>/` → commit (old/ untouched).

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

1. `cd .lintcn && go build ./... && TSGOLINT_SNAPSHOT_CWD=true go test -v ./...` — all green.
2. `npx lintcn lint` against `old/fixtures/` trees — each rule fires on its
   fixture's violations; glob gating respected; severities correct.
3. Spot-check snapshots for message quality (they are the agent-facing output).
4. Final commit deletes `old/` wholesale (decision 1): all remaining rules,
   fixtures, docs, archlint — nothing of the old fork survives.
5. Decide: keep `tsconfig.json` scoped to fixtures, or add a demo lint target.

## Open questions (resolved by default)

- Where helpers live → `.lintcn/archkit/` in the lintcn workspace module.
- Files glob → in-rule gating via helper; no registration-level support.
- Severity → error for invariants, warn for advisories (table above).
