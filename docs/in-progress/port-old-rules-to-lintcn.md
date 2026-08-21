# Port `old/rules` to lintcn `.lintcn/` rules — DONE

All 19 rules live under `.lintcn/<snake_name>/`, sharing `.lintcn/archkit/`.
The `old/` tsgolint fork is deleted. `npx lintcn list` shows all 19 with the
intended name, description and severity; `go build ./...` and
`TSGOLINT_SNAPSHOT_CWD=true go test ./...` are green; gremlins reports 100%
test efficacy on every rule package and on archkit.

## Conventions this port established (they govern rule 20)

These are not in the vendored `.agents/skills/lintcn/SKILL.md`, which is
synced from `remorses/lintcn` and must not be edited locally.

1. **`Name:` must be an inline string literal**, never a const — lintcn's
   discovery binds the CLI name by matching `Name: "..."` in source.
2. **Run rule tests with `TSGOLINT_SNAPSHOT_CWD=true`**, which resolves
   `__snapshots__` next to the rule package instead of inside the tsgolint
   cache (`UPDATE_SNAPS=true` likewise). Never hand-copy a snapshot out of
   the cache. The committed snapshots are load-bearing regression tests and
   the agent-facing output — read them after writing tests.
3. **File gating lives in the rule.** lintcn registers a plain `rule.Rule`
   with no `Files` field, so the old fork's `archrule.Rule{Files: ...}`
   became `archkit.Gated(files, run)`, which returns nil listeners for a
   non-matching `ctx.SourceFile.FileName()`. Every gated rule's test table
   carries an out-of-tree valid case proving the gate.
4. **Configurable scope:** a rule with a user-overridable tree parses its
   options first and passes the resolved patterns to `archkit.Gated`;
   defaults stay in `var defaultFiles`.
5. **archkit grows on demand** — add a helper when the first rule needs it.
   Porting wholesale leaves unconsumed symbols that read as mutation
   NOT COVERED noise. Ten such mutants in `types.go` are covered only by
   rule-package tests gremlins cannot see cross-package; the file documents
   them.
6. **Severity:** architectural invariants are `error` (the default — omit
   the directive). Advisory rules are `warn`, which in lintcn reports on
   changed and untracked files only and never fails CI — the right shape for
   agent-facing guidance on new code. Currently `warn`:
   `no-page-request-in-journey`, `no-double-library-in-domain-test`,
   `no-double-library-in-use-case-test`.
7. **Docs URLs are dropped.** `archrule.DocumentedAt` and the `docs/rules/*.md`
   enforcement have no lintcn equivalent; message `Help` text carries the
   guidance instead.
8. **Gremlins runs per rule package** at its commit (`gremlins unleash
   ./<rule>/`, sub-second), plus archkit.
9. **Stale-lock gotcha:** aborting `lintcn lint`/`build` leaves a lock dir
   under `~/.cache/lintcn/locks/build/<hash>` that blocks later runs. Remove
   it, or `npx lintcn clean`.

## Known gap

Nothing now proves a rule fires through the built binary rather than through
`rule_tester`. The old fork's `fixture-gate.mjs` and its 19 fixture projects
did, and were deleted with `old/` as a deliberate trade.
