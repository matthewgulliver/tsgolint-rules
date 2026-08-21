# lintcn rule conformance checklist

Binary propositions for the Go architecture lint rules in `.lintcn/`.
**Every item is phrased so that `YES` = the intended state.** A `NO` is a defect or a deliberate,
documented exception. `N/A` is a legitimate answer where the *Applies to* column narrows scope.

Every item is checkable by reading one rule package, or by running one command named in the
Enforcement column. Items grounded in vendored source cite the file.

## How to read the Enforcement column

| Tag | Meaning |
| --- | --- |
| `COMPILE` | The Go compiler rejects it. `go build ./...` in `.lintcn/`, or the generated `wrapper/main.go` build during `lintcn build`. |
| `DISCOVERY` | `node_modules/lintcn/dist/discover.js` scans rule sources with regexes before anything is compiled. A miss here is silent: the rule is dropped, renamed, or re-severitied with no error. |
| `RUNTIME` | The linter or the built binary behaves wrongly. Nothing fails; the rule just does the wrong thing. |
| `TEST` | `go test` fails via `rule_tester`'s own assertions. |
| `SNAPSHOT` | `__snapshots__/<rule-name>.snap` mismatches, churns, or accumulates orphans. |
| `VET` | `go vet ./...`. |
| `GOFMT` | `gofmt -l .`. |
| `CONV` | This repository's convention, encoded in the plop generator and the 19 existing rules. No tool enforces it; violating it is a review finding. |
| `MUTATION` | The `gremlins unleash ./<rule>/` gate from `.agents/skills/go-tdd-mutation/SKILL.md`. |

Sources per section. `SKILL.md` means `.agents/skills/lintcn/SKILL.md`. Where it disagrees
with the 19 rules, the rules win and the disagreement is recorded in *Received wisdom that is
now wrong*.

---

## A. Rule identity and layout — *scope: one rule package*

Sources: `node_modules/lintcn/dist/discover.js`; `node_modules/lintcn/src/codegen.ts`;
`plopfile.js`; SKILL.md § Directory Layout, § Package Name.

| # | Proposition (YES = intended) | Applies to | Enforcement |
| --- | --- | --- | --- |
| A1 | The rule lives in its own subfolder of `.lintcn/`, not as a flat `.go` file at the root. | every rule | DISCOVERY (`discoverRules` only descends into directories; a flat file prints an error and is skipped) |
| A2 | The folder name is the snake_case form of the kebab-case CLI name. | every rule | CONV (the folder name is only a fallback display name; `codegen.ts` imports it under a sanitised alias) |
| A3 | Exactly one `var …Rule = rule.Rule{` appears per `.go` file. | every rule file | DISCOVERY (`RULE_VAR_RE` is matched non-globally — a second rule var in the same file is invisible) |
| A4 | The rule variable is exported (leading capital). | every rule | COMPILE (the generated `wrapper/main.go` writes `alias.VarName`) |
| A5 | The rule variable's name is PascalCase of the CLI name with a `Rule` suffix. | every rule | CONV (`discover.js` accepts any `\w+`; `plopfile.js` generates this shape) |
| A6 | `Name:` is an inline `"…"` string literal inside the `rule.Rule{…}` literal, never a const or a variable. | every rule | DISCOVERY + RUNTIME (`buildGoRuleNameRe` matches `Name:\s*"([^"]+)"`; on a miss `goRuleName` falls back to the display name, and `--warn <goRuleName>` then fails to match the runtime `RuleName`, silently promoting a `warn` rule to a CI-failing error) |
| A7 | The rule file is `<snake_name>.go`. | rules whose name does **not** end in `-test` | CONV |
| A8 | The rule file is `rule.go`, and the package is the snake name with the `_test` suffix trimmed. | rules whose name ends in `-test` | COMPILE, in two separate ways — both verified by construction, see the note below |
| A9 | The Go package name matches the folder name. | every rule except A8's exception | CONV (`codegen.ts` imports each folder under an explicit alias, so a mismatch builds fine — this is why A8 works) |
| A10 | The test file is `<snake_name>_test.go` — including the doubled `..._test_test.go` for A8's rules. | every rule | COMPILE (Go only runs `*_test.go`) |
| A11 | The test function is `Test<PascalCase of the CLI name>`. | every rule | SNAPSHOT (`t.Name()` is the snapshot key; renaming the function orphans every entry) + CONV |
| A12 | Rule options, if separated out, live in `options.go` beside the rule and contain no `rule.Rule` var. | rules with a separate options file | DISCOVERY (a second non-test `.go` file carrying a rule var registers as a *second* rule) |
| A13 | The generated `go.mod`, `go.work`, `go.work.sum`, `go.sum` and `.tsgolint/` are gitignored, not committed. | the repository | CONV (`.gitignore`; `codegen.ts` rewrites `go.work`/`go.mod` on every build) |

**A8, the two mechanisms.** `plopfile.js` explains this rule, but its second reason is imprecise.
Verified by building the broken shapes in a scratch module:

- **The filename.** A package whose only non-test file is `<snake>_test.go` reports
  `GoFiles=[] XTestGoFiles=[<snake>_test.go]` — Go classifies it as a test file and the package has
  no source at all. `go build ./...` inside `.lintcn/` still *succeeds*, silently, because there is
  nothing to build. The failure surfaces later, when the generated `wrapper/main.go` imports the
  package and finds no such symbol. Hence `rule.go`.
- **The package name.** `package <snake>_test` is **not** illegal on its own — a file named
  `rule.go` declaring `package foo_test` compiles fine. It breaks only once a `_test.go` file
  exists beside it: Go strips the `_test` suffix from a test file's package clause to derive the
  package under test, so a directory holding `rule.go` (`package foo_test`) and any
  `*_test.go` (`package foo_test`) fails with `found packages foo (…_test.go) and foo_test (rule.go)`.
  Since every rule has a test file, the trimmed package name is forced in practice.

---

## B. `lintcn:` metadata directives — *scope: the comment block above `package`*

Sources: `discover.js` (`METADATA_RE`, `parseMetadata`); `src/commands/list.ts`;
`src/commands/lint.ts`; SKILL.md § Metadata Comments.

Four directives exist and no others are read: `name`, `severity`, `description`, `source`.
All are optional. All are parsed by line regex `^// lintcn:(\w+) (.+)$` anywhere in the file.

| # | Proposition (YES = intended) | Applies to | Enforcement |
| --- | --- | --- | --- |
| B1 | `// lintcn:name` is present and its value is byte-identical to the Go `Name:` literal. | every rule | CONV (they are independent: `name` drives `lintcn list` and the fallback, `Name:` drives diagnostics and `--warn`; a mismatch shows two different names for one rule) |
| B2 | `// lintcn:severity` is written explicitly, even for `error`. | every rule | CONV (all 19 rules and the generator do this; `error` is the parsed default but is still stated) |
| B3 | The severity value is exactly `error` or `warn`, lowercase. | every rule | DISCOVERY (`meta.severity === 'warn' ? 'warn' : 'error'` — every other spelling, including `Warn` and `warning`, silently becomes `error`) |
| B4 | The severity is `error` because the rule states an architectural invariant that must fail CI. | invariant rules | CONV |
| B5 | The severity is `warn` because the rule is advisory guidance for newly written code. | advisory rules | CONV (16 rules are `error`; `no-page-request-in-journey`, `no-double-library-in-domain-test`, `no-double-library-in-use-case-test` are `warn`) |
| B6 | `// lintcn:description` is present, a single line, and starts with `Disallow`, `Require`, or `Enforce`. | every rule | CONV (`plopfile.js` validates this prefix at generation time; nothing checks it afterwards) |
| B7 | The description names the concrete thing disallowed, not the rule's own title restated. | every rule | CONV |
| B8 | `// lintcn:source` is absent. | rules written here | CONV (it exists for rules vendored via `lintcn add`; no rule in this repository uses it) |
| B9 | The directives sit above `package`, in the order name / severity / description. | every rule | CONV (`plop-templates/rule.go.hbs`; the parser accepts them anywhere) |

---

## C. Rule structure and file gating — *scope: the `rule.Rule` value and its `Run`*

Sources: `.lintcn/.tsgolint/internal/rule/rule.go`; `.lintcn/.tsgolint/internal/linter/linter.go`;
`.lintcn/archkit/gate.go`.

| # | Proposition (YES = intended) | Applies to | Enforcement |
| --- | --- | --- | --- |
| C1 | The rule is a `rule.Rule` value with exactly the two fields the struct has, `Name` and `Run`. | every rule | COMPILE (`rule.Rule` has no `meta`, no `Files`, no docs URL, no schema) |
| C2 | File scoping is done inside the rule, via `archkit.Gated(defaultFiles, run)`. | rules scoped to a tree | RUNTIME (there is no registration-time file scope; without a gate the rule judges every file in the program) |
| C3 | The default patterns live in `var defaultFiles []string` at the top of the rule file. | gated rules | CONV (all 19) |
| C4 | A rule whose scope is user-overridable parses its options first and gates on the resolved patterns, inlining the gate rather than calling `archkit.Gated` with the defaults. | rules whose options move their own scope | RUNTIME (`archkit.Gated` closes over the patterns before `Run` sees options — `domain_type_is_declared_once` is the one rule in this shape and inlines the `archkit.Includes` check for exactly this reason) |
| C5 | The gate returns `nil` listeners, not an empty map, for a file outside the tree. | gated rules | CONV (both are safe: the linter ranges over the returned map and a nil map yields nothing) |
| C6 | The rule tolerates `ctx.SourceFile == nil`. | rules reading the file name | RUNTIME (`archkit.Gated` checks it; an inlined gate must too) |
| C7 | Per-file state is created inside `Run`, not in a package-level `var`. | rules holding state | RUNTIME (`linter.go` calls `r.Run(ctx)` once per file, so a `Run`-local closure is already per-file; a package-level var leaks across files and across the parallel workers) |
| C8 | Whole-program state shared across files is keyed by `*compiler.Program` and built once. | rules needing cross-file facts | RUNTIME (`domain_type_is_declared_once` uses a `sync.Map` keyed by program; the rule tester builds a fresh program per test case, so a global cache would bleed between cases) |
| C9 | Everything the rule decides is reachable from `ctx.SourceFile`, `ctx.Program` and `ctx.TypeChecker` — the rule reads no files and no environment of its own. | every rule | CONV |
| C10 | The rule declares no fixes and no suggestions. | every rule here | CONV (`ReportNodeWithFixes`/`ReportNodeWithSuggestions` exist and work; no rule in this repository uses them, because an architectural violation has no mechanical correction) |

---

## D. Listeners and traversal — *scope: the returned `rule.RuleListeners`*

Sources: `linter.go` (`visitLintNodes`, `runListeners`); `rule.go`
(`ListenerOnExit`, `ListenerOnAllowPattern`, `ListenerOnNotAllowPattern`).

| # | Proposition (YES = intended) | Applies to | Enforcement |
| --- | --- | --- | --- |
| D1 | Every map key is an `ast.Kind*` constant (or a `rule.Listener*` wrapper of one). | every rule | RUNTIME (`RuleListeners` is `map[ast.Kind]func(*ast.Node)`; a wrong-but-valid kind compiles and simply never fires — the rule reports nothing and the tests go green only if the invalid cases are missing too) |
| D2 | The rule does not listen on `ast.KindSourceFile`. | every rule | RUNTIME (`visitLintNodes` calls `file.Node.ForEachChild(childVisitor)`, so the source-file node itself is never passed to `runListeners`) |
| D3 | The listener re-checks the node's shape before casting (`node.AsXxx()` matches the registered kind). | every listener | RUNTIME (a mismatched `AsXxx` reads the wrong union member) |
| D4 | `rule.ListenerOnExit(kind)` is used where the decision needs the subtree already walked; enter listeners otherwise. | rules tracking scope | CONV (no rule here needs it) |
| D5 | Where two listeners share a decision, they share one closure rather than duplicating it. | multi-kind rules | CONV (`domain_type_is_declared_once` registers one `judge` under both `KindInterfaceDeclaration` and `KindTypeAliasDeclaration`) |
| D6 | A declaration-level rule confirms the declaration is top-level rather than nested, where that is part of the claim. | rules about a file's published vocabulary | CONV (`node.Parent != ctx.SourceFile.AsNode()`) |
| D7 | Parentheses are skipped before an expression's content is judged (`ast.SkipParentheses`). | rules reading expressions | RUNTIME (SKILL.md § Skipping Parentheses) |
| D8 | The report targets the smallest node that identifies the problem — the name, the parameter, the clause — not the enclosing statement, unless the statement *is* the problem. | every report | CONV + SNAPSHOT (the underline in the `.snap` is the evidence) |
| D9 | The rule reports at most once per subject, returning after the first finding where a second would restate it. | rules that could report per member | CONV (`port_behaviour_is_an_interface` returns on the first callable member) |
| D10 | Where another rule already reports the same mistake, this rule detects the overlap and stays silent. | overlapping rules | CONV (`readableAsWritten` in `port_behaviour_is_an_interface`, documented at the function) |

---

## E. Type-checker use — *scope: every `ctx.TypeChecker` call*

Sources: `.lintcn/archkit/types.go` (and its recorded failure modes);
SKILL.md § Type Checker APIs; the 16 rules that use the checker.

| # | Proposition (YES = intended) | Applies to | Enforcement |
| --- | --- | --- | --- |
| E1 | The rule uses the type checker only for the fact syntax cannot supply — which file declares the thing, or what a name resolves to through aliases. | every typed rule | CONV (the stated reason these rules are in Go at all; `types.go` package doc) |
| E2 | A union type is flattened before its members are judged (`archkit.Constituents`, or `utils.UnionTypeParts`). | rules judging a type's shape | RUNTIME (recorded failure: a check that skips this passes `string \| undefined` and the rule stops enforcing anything — `types.go`, `Constituents`) |
| E3 | A resolved symbol is passed through `checker.SkipAlias` before its declarations are read. | rules resolving imported names | RUNTIME (an import alias declares itself in the importing file, so without this every imported symbol looks locally declared) |
| E4 | `Promise<T>` is unwrapped before a return type is judged (`archkit.Unwrapped`). | rules judging port/use-case returns | RUNTIME (recorded: every port here is async, so a rule without this judges `Promise` — `types.go`, `Unwrapped`) |
| E5 | Array and tuple element types are read through `Checker_getTypeArguments`, not from type flags. | rules judging collections | RUNTIME (recorded: `readonly string[]` is a reference type, not a string-flagged one — `types.go`, `ElementTypes`) |
| E6 | Every checker call that can return `nil` — type, symbol, declaration, source file — has an explicit arm, and that arm returns "not a violation" rather than guessing. | every typed rule | RUNTIME (panics otherwise; `types.go` centralises the common ones) |
| E7 | A defensive nil arm that no test can reach is commented at the site with why it exists. | unreachable arms | CONV + MUTATION (`DeclaredType`'s two arms are documented this way; the alternative is a survivor with no explanation) |
| E8 | Shared checker questions live in `archkit`, added when the first rule needs one. | every rule | CONV (porting helpers wholesale leaves unconsumed symbols that read as mutation `NOT COVERED` noise) |
| E9 | A helper added to `archkit` is consumed by at least one rule in the same change. | every archkit addition | MUTATION (an unconsumed exported helper is a `NOT COVERED` survivor by construction) |
| E10 | Type names put into a message come from `TypeChecker.TypeToString` or the written identifier, not from a hand-built string. | messages naming types | CONV |

---

## F. Options — *scope: the rule's `Options` struct and its accessors*

Sources: `.lintcn/.tsgolint/internal/utils/utils.go` (`UnmarshalOptions`);
`.lintcn/.tsgolint/internal/runner/runner.go:292`.

**The governing fact:** the shipped binary calls `r.Run(ctx, nil)`. lintcn has no rule-configuration
file, so in production every rule runs with `options == nil` and therefore on its defaults. Options
are reachable only from `rule_tester` cases. Defaults are the real behaviour; the option is the
tested-but-unreachable path.

| # | Proposition (YES = intended) | Applies to | Enforcement |
| --- | --- | --- | --- |
| F1 | Options are read with `utils.UnmarshalOptions[Options](options, "<rule-name>")`. | rules with options | RUNTIME (it marshals then unmarshals, which is what makes `nil` safe) |
| F2 | The rule-name argument to `UnmarshalOptions` is the CLI name. | rules with options | RUNTIME (it is the panic prefix and nothing else) |
| F3 | Every field is optional in JSON (`omitempty`) and the struct's zero value is a valid "nothing configured" state. | rules with options | RUNTIME (`options == nil` marshals to `null`, which leaves every field zero) |
| F4 | Every option has an accessor method that substitutes the package-level default when the field is unset. | rules with options | RUNTIME (without one the rule runs with empty patterns in production and enforces nothing) |
| F5 | A boolean option whose default is `true` is a `*bool`, so "unset" is distinguishable from "explicitly false". | boolean options | RUNTIME (`port_behaviour_is_an_interface`: `opts.ExportedOnly == nil \|\| *opts.ExportedOnly`) |
| F6 | A slice option uses `len(…) == 0` as the unset test rather than a pointer. | slice options | CONV (an explicitly empty list and an absent one mean the same thing for a pattern list) |
| F7 | Pattern options are compiled with `archkit.Compile`, which drops patterns that do not compile. | regex options | RUNTIME (a bad pattern identifies nothing rather than failing the whole run) |
| F8 | Each option that changes what reports has at least one test case exercising it. | rules with options | CONV + MUTATION (it is the only reachable path for that code) |
| F9 | The rule does not branch on whether options were supplied — it branches on resolved values only. | rules with options | CONV |

---

## G. Messages — *scope: each `rule.RuleMessage`*

Sources: `rule.go` (`RuleMessage{Id, Description, Help}`); the 19 rules' `build…Message`
functions and their snapshots; `plop-templates/rule.go.hbs`.

| # | Proposition (YES = intended) | Applies to | Enforcement |
| --- | --- | --- | --- |
| G1 | Each message is built by a named `build<Something>Message(…)` function, not inlined at the report site. | every message | CONV (all 19) |
| G2 | `Id` is camelCase and names the **violation**, not the rule. | every message | CONV (`pageRequest`, `bareFunctionTypeAlias`, `foreignReturnType` — note the plop template seeds `camelCase(rule name)` instead, which every real rule replaced) |
| G3 | `Id` is unique within the rule, and a rule with two distinct failures has two distinct ids. | every rule | TEST (the test asserts `MessageId` per error, so a shared id makes two failures indistinguishable) |
| G4 | `Description` says what is wrong **and what it costs**, in one sentence. | every message | CONV |
| G5 | `Help` says what to do instead, and appears in every message. | every message | CONV (`Help` is optional in the struct; all 19 supply it. `rule.Rule` has no docs-URL field, so nothing links a diagnostic to `docs/rules/` — `Help` is the only guidance the reader gets in place) |
| G6 | The message names the specific symbol it is about, backticked. | every message | CONV |
| G7 | Interpolated names come from the AST or the checker, never from a fixed placeholder string. | every message | CONV |
| G8 | `Description` and `Help` are complete sentences ending in a period. | every message | CONV |
| G9 | The message has been read as rendered in the `.snap`, not just as source. | every rule | CONV (the snapshot is the agent-facing output) |

---

## H. Tests — *scope: one rule's `_test.go`*

Sources: `.lintcn/.tsgolint/internal/rule_tester/rule_tester.go`; SKILL.md § Testing;
`plop-templates/rule_test.go.hbs`; `.agents/skills/go-tdd-mutation/SKILL.md`.

| # | Proposition (YES = intended) | Applies to | Enforcement |
| --- | --- | --- | --- |
| H1 | The test calls `rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &XxxRule, valid, invalid)`. | every rule | CONV (all 19) |
| H2 | The test function calls `t.Parallel()`. | every rule | CONV (`rule_tester` also parallelises each case) |
| H3 | Both the valid and the invalid slice are non-empty. | every rule | CONV (a rule with no invalid case proves nothing; one with no valid case has no evidence it is narrow) |
| H4 | Every invalid case lists exactly as many `Errors` entries as the rule reports, in report order. | invalid cases | TEST (`t.Fatalf` on a count mismatch; index-wise comparison thereafter) |
| H5 | Every `InvalidTestCaseError` sets `MessageId`. | invalid cases | TEST (an empty `MessageId` is compared against the real id and fails) |
| H6 | `Line`/`Column`/`EndLine`/`EndColumn` are asserted only where the position is the point of the case. | invalid cases | TEST (zero means "don't check"; the snapshot already pins the range) |
| H7 | `Output` is absent. | rules that declare no fixes | TEST (`len(testCase.Output) == len(outputs)`, both zero) |
| H8 | No committed case carries `Only: true`. | every rule | TEST-defeating (one `Only` anywhere makes `rule_tester` `t.SkipNow()` every other case in that rule — the suite passes green while testing almost nothing) |
| H9 | No committed case carries `Skip: true`. | every rule | CONV (same silent-green failure mode, per case) |
| H10 | Gated rules have at least one valid case whose file is outside the gated tree, carrying code that would otherwise report. | gated rules | CONV (all 19 have one, either by omitting `FileName` — the default is `file.ts` — or by naming an out-of-tree path such as `apps/web/src/page.ts`) |
| H11 | Every case that should report sets `FileName` to a path inside the gated tree. | gated rules | TEST (otherwise the gate suppresses the rule and the invalid case reports nothing) |
| H12 | `FileName` is written relative and is matched by `defaultFiles` through a leading `**/`. | gated rules | RUNTIME (`rule_tester` resolves it against `fixtures.GetRootDir()`, which is inside the tsgolint cache, so the absolute path has an unpredictable prefix) |
| H13 | The valid cases include near-misses — the shapes a careless implementation would flag. | every rule | CONV + MUTATION (a valid case that could never report kills no mutants) |
| H14 | Each case carries a comment naming the behaviour it proves, not restating its code. | every rule | CONV |
| H15 | Multi-file scenarios use `Files: map[string]string{…}` rather than a single concatenated source. | cross-file rules | CONV |
| H16 | Cases are ordered so that inserting a new one appends rather than renumbering. | every rule | SNAPSHOT (snapshot keys are `invalid-<index>`; inserting in the middle rewrites every later entry) |
| H17 | Option-bearing cases use one of the three supported shapes consistently within a rule. | rules with options | CONV — currently inconsistent across the repo: `map[string]any` (2 rules), a typed `Options{…}` literal (5), `rule_tester.OptionsFromJSON[Options](…)` (1). All three round-trip through JSON identically. The typed literal is the majority and the one that fails to compile when a field is renamed; prefer it. |
| H18 | Tests run with `TSGOLINT_SNAPSHOT_CWD=true`. | every rule | RUNTIME (without it snapshots resolve into the tsgolint cache directory, not beside the rule) |

---

## I. Snapshots — *scope: `__snapshots__/<rule-name>.snap`*

Sources: `.lintcn/.tsgolint/internal/rule_tester/snapshot.go`.

| # | Proposition (YES = intended) | Applies to | Enforcement |
| --- | --- | --- | --- |
| I1 | The snapshot file is committed. | every rule | CONV (it is the regression evidence and the agent-facing output) |
| I2 | Its name is the CLI rule name plus `.snap` — not the package or folder name. | every rule | RUNTIME (`newSnapshotter(r.Name)`) |
| I3 | It sits in `__snapshots__/` inside the rule package. | every rule | RUNTIME (only when `TSGOLINT_SNAPSHOT_CWD=true`; the variable is read once in `init()`) |
| I4 | It was generated by `TSGOLINT_SNAPSHOT_CWD=true UPDATE_SNAPS=true go test`, never hand-copied out of the cache. | every rule | CONV (a hand-copied snapshot silently drifts from what the rule actually emits) |
| I5 | Every entry corresponds to a live invalid case. | every rule | SNAPSHOT — **not enforced**: `write` merges over loaded entries and never deletes, so a removed or renumbered case leaves an orphan forever. Deleting the `.snap` and regenerating is the only cleanup. |
| I6 | A newly added invalid case's entry was read before committing. | new cases | CONV — **not enforced**: a missing key is written and passes silently (`if update \|\| !exists`), so an unreviewed snapshot is indistinguishable from a reviewed one. |
| I7 | Valid cases have no snapshot entries. | every rule | RUNTIME (`MatchSnapshot` is called only in the invalid loop) |
| I8 | The underlined range in each entry is the smallest range that identifies the violation. | every entry | CONV (this is the check D8 is verified by) |

---

## J. Go hygiene — *scope: the `.lintcn/` module*

| # | Proposition (YES = intended) | Applies to | Enforcement |
| --- | --- | --- | --- |
| J1 | `go build ./...` succeeds from `.lintcn/`. | every change | COMPILE |
| J2 | `go vet ./...` is clean. | every change | VET (verified clean at the time of writing) |
| J3 | `gofmt -l .` reports nothing. | every file | GOFMT — **currently NO**: three committed test files have unaligned struct-literal keys (`no_constructed_collaborators`, `port_behaviour_is_an_interface`, `published_contract_publishes_no_mutable_value`). No hook or CI enforces it; `plopfile.js` runs `gofmt -w` only on generated files. |
| J4 | The rule file has a package doc comment, or the `lintcn:` block plus per-function comments carry the explanation. | every rule | CONV (only `archkit` carries package docs; rule packages explain themselves at the message builders and the decision functions) |
| J5 | Every exported symbol in `archkit` is consumed by at least one rule. | `archkit` | MUTATION (see E9) |
| J6 | Comments record the constraint or the recorded failure, not a restatement of the code. | every file | CONV (`types.go` and `scope.go` are the model: each comment names a real bug the code exists to prevent) |
| J7 | No rule imports another rule package. | every rule | CONV (shared logic goes to `archkit`; the packages are siblings with no dependency order) |

---

## K. TDD and the mutation gate — *scope: one rule package's development cycle*

Source: `.agents/skills/go-tdd-mutation/SKILL.md`.

| # | Proposition (YES = intended) | Applies to | Enforcement |
| --- | --- | --- | --- |
| K1 | The failing invalid case was written before the logic that makes it report. | every behaviour change | CONV (`npm run new:rule` emits a deliberately RED package for this reason) |
| K2 | RED was proven with an exact `-run` selector. | every behaviour change | CONV |
| K3 | `gremlins unleash ./<rule>/` was run once at the end of the phase, not per increment. | every rule | MUTATION |
| K4 | Every `LIVED` mutant was resolved by adding the missing case, not by deleting or weakening a test. | every rule | MUTATION |
| K5 | Every `NOT COVERED` mutant is either newly covered or the code is deleted as dead. | every rule | MUTATION |
| K6 | A survivor that must stay was made testable or made smaller; an irreducible residue is commented at the site and recorded as an equivalent-mutant exception. | every rule | MUTATION |
| K7 | `archkit`'s cross-package `NOT COVERED` survivors remain documented in `types.go` with which rule tests cover them. | `archkit` | MUTATION (gremlins cannot see coverage originating in the rule packages; ten such mutants are recorded there) |
| K8 | The whole module was green (`TSGOLINT_SNAPSHOT_CWD=true go test ./...`) before the commit. | every commit | TEST |

---

## L. Per-file execution model — *scope: the `Run` body*

`RunLinterOnProgram` calls `r.Run(ctx)` once per file per rule, and every registered listener
fires on every matching node in every file of the program. There is no per-rule enablement and
no file filter outside the rule itself. Both items below follow from that; neither is a
performance claim, because nothing here has been profiled — see *Known gaps*.

| # | Proposition (YES = intended) | Applies to | Enforcement |
| --- | --- | --- | --- |
| L1 | The gate is the first thing `Run` does, so an out-of-scope file registers no listeners at all. | gated rules | RUNTIME |
| L2 | Whole-program work is done once and cached, not per file. | rules with cross-file facts | RUNTIME (`domain_type_is_declared_once`) |

---

## Received wisdom that is now wrong

| Stale advice | Current state |
| --- | --- |
| SKILL.md § Package Name: "The package name must match the folder name." | Not enforced anywhere. `codegen.ts` imports each folder under an explicit alias, which is why the two `-test` rules can use trimmed package names. Match it as a convention; do not treat a mismatch as a build error. |
| SKILL.md § Package Name: "The exported variable name must match the pattern `var XxxRule`." | `RULE_VAR_RE` accepts `var <anything> = <anything>.Rule{`. What is actually forced is that the variable be **exported** (`wrapper/main.go` references it) and that the type name end in `Rule`. |
| SKILL.md § Rule Options: rules "accept configuration via JSON". | Not through lintcn. `runner.go:292` passes `nil` for every rule on every file. Options are a test-only surface today; defaults are the shipped behaviour. |
| SKILL.md § Snapshots: "Copy them into your rule folder for reference." | Superseded by `TSGOLINT_SNAPSHOT_CWD=true`, which writes them there directly. Hand-copying is how a snapshot silently drifts from what the rule emits. |
| "A snapshot mismatch will catch it." | Only for an entry that already exists. A brand-new key is written and the test passes; a deleted case's key is never removed. |
| "`meta`/docs URL/schema" — anything carried over from the ESLint or oxlint rule model. | `rule.Rule` has two fields. There is no metadata object, no options schema, no docs URL, no `recommended` preset, and no per-project rule configuration. `docs/rules/` exists, but nothing links a diagnostic to it, so `RuleMessage.Help` has to stand on its own. |

---

## Known gaps

- **Nothing proves a rule fires through the built binary.** Every assertion in this repository goes
  through `rule_tester`, which constructs its own program and calls `r.Run` directly. The old fork's
  fixture-gate and its 19 fixture projects covered this and were deleted with `old/` as a deliberate
  trade. Section A's DISCOVERY items are therefore the least-verified in
  practice: a discovery-level break shows up as a rule that silently stops running.
- **No CI.** There is no `.github/` in this repository. Every `TEST`, `VET`, `GOFMT` and `MUTATION`
  tag names a command someone has to run, not a gate that blocks a merge.
- **There are no performance propositions, deliberately.** The linter has per-rule timing
  (`RuleTimingStore` in `linter.go`), but `runner.go` never sets it, so `npx lintcn lint` exposes
  no timing flag and no rule here has ever been profiled. Items asserting a cost model — guard
  before checker call, narrowest node kind, allocation in `Run` — were dropped rather than shipped
  as reasoning dressed up as findings. Add them when there is a measurement.
- **The mutation numbers are historical.** 100% test efficacy per rule package and for `archkit`
  was recorded at the time of the port; it was not re-run for this checklist.
- **`archkit`'s ten recorded cross-package survivors were not re-verified.** The comment in
  `types.go` names which rule-tester cases cover them; that claim is taken at face value here.
- **Upstream tsgolint behaviour is read from the vendored cache**, at
  `~/.cache/lintcn/tsgolint/c031d7264983dba00bae76eb03532fe3884e5667`, which `.tsgolint` symlinks to.
  Every `RUNTIME` and `COMPILE` claim about `rule`, `linter`, `runner` and `rule_tester` is accurate
  for that commit and should be re-read after a tsgolint bump.
- **Two conventions have no stated rationale and were left as CONV rather than promoted:** why
  `Help` is universal (inferable from the dropped docs URLs) and why no rule offers a fix. Both are
  uniform across 19 rules, which is evidence of a decision but not a record of one.
- **Not covered:** the `lintcn add` / vendoring path, the stale build-lock gotcha (see the README), and
  `--fix` semantics. None are rule-authoring propositions.
