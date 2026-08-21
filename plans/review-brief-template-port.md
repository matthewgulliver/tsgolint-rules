# Review Brief: lintcn Rule Port — `.lintcn/no_page_request_in_journey` + `.lintcn/archkit`

You are reviewing the first ported rule of a 19-rule migration from
`old/rules` (a tsgolint fork with internal `archrule`/`archscope`/`archtypes`
packages) to lintcn-owned rules under `.lintcn/`. This port is the **template**
for the remaining 18 rules, so findings here multiply. Be adversarial; the
goal is the perfect example, not approval.

## What to read, in order

1. `.agents/skills/lintcn/SKILL.md` — the authoritative lintcn rule-writing
   reference (rule anatomy, metadata directives, testing, snapshots).
2. `.agents/skills/go-tdd-mutation/SKILL.md` — this repo's Go TDD + mutation
   policy (behavior tests, no internals, Gremlins triage).
3. `plans/port-old-rules-to-lintcn.md` — the port plan and commit discipline
   (one rule per commit, old rule deleted in the same commit).
4. The diff under review: `git show 3c11390`.
5. Sources: `old/rules/` (remaining rules + their tests), `old/internal/`
   (original helper implementations), `old/fixtures/` (per-rule fixture
   trees), and the ported `.lintcn/archkit/` +
   `.lintcn/no_page_request_in_journey/`.

## The port, in short

- `old/internal/archscope` + `archtypes` → `.lintcn/archkit` (scope.go,
  types.go, gate.go), tests ported (`scope_test.go`).
- `archrule.Rule{Files: ...}` (fork-only wrapper; lintcn registers plain
  `rule.Rule`) → `archkit.Gated(files, run)` which returns nil listeners for
  non-matching files.
- First rule `no_page_request_in_journey` ported: rule + test copied with
  `archkit` imports, `lintcn:name/severity warn/description` metadata added,
  snapshot captured and reviewed, old rule folder deleted in the same commit.

## Questions to critique, specifically

1. **Gating semantics.** `archkit.Gated` checks `ctx.SourceFile.FileName()`
   against `**/e2e/**`-style patterns. Is matching on the absolute path the
   right call vs the old archlint entrypoint's behavior? Any edge cases
   (Windows separators, fixture root prefix, `Files` map paths in tests)
   that make Includes mis-judge? Is returning `nil` listeners the correct
   no-op, per the lintcn/tsgolint contract?
2. **archkit placement.** Shared package inside `.lintcn/` — does anything in
   lintcn (discovery, codegen, `lintcn list/build/lint`) break or misbehave
   with a non-rule subfolder present? Is the package doc/comments quality
   consistent with the fork originals?
3. **Port fidelity.** Diff the rule logic against
   `old/rules/no_page_request_in_journey/` (in git history at `3c11390^`).
   Any behavior dropped or subtly changed beyond the mechanical
   archrule→Gated and import-path changes? Are the tests byte-equivalent
   apart from the `&Rule.Rule` → `&Rule` change?
4. **Metadata choices.** `lintcn:severity warn` for this rule: right or
   should it be error? Name/description wording? Missing directives
   (`lintcn:source`?)?
5. **Snapshot handling.** Snapshot was regenerated into the cached tsgolint
   source and copied to `.lintcn/<rule>/__snapshots__/`. Is copying it into
   the repo the right move, or should it stay untracked (does the test
   re-verify against the committed copy, or does it always compare against
   the cache)? Verify how `rule_tester` resolves snapshots and whether the
   committed file is load-bearing.
6. **Mutation evidence.** `gremlins unleash ./archkit` showed 30 killed,
   0 lived, 23 not-covered (all in types.go, covered only via rule-package
   tests Gremlins can't see cross-package). Is that triage sound per the
   go-tdd-mutation skill, or should archkit tests cover types.go directly?
   Was Gremlins run on the rule package itself? (It was not — check whether
   it should have been before this commit, given rule_tester runs in the
   same package.)
7. **Commit discipline.** Does `3c11390` satisfy the plan's one-rule-per-
   commit rule (complete port, old rule deleted, nothing unrelated mixed
   in)? Is the message adequate per the commit-rules notion of a labelled
   example (subject as index, why-body for non-pure moves)?
8. **go.work.sum / generated files.** Was anything generated committed that
   shouldn't be, or anything needed left uncommitted? (Check `.lintcn/.gitignore`
   coverage.)
9. **Template fitness.** Would you hand the remaining 18 rules to agents
   using this commit as the worked example? What would you change first?

## Output format

A ranked list of findings, each with: severity (blocking / should-fix /
note), file:line evidence, why it matters for the 18-rule fan-out, and the
minimal corrective change. End with a verdict: template as-is, template with
listed fixes, or rework.
