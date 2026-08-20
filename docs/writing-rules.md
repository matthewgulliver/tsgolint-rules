
# Writing type-aware lint rules in Go

The `arch/*` rules in `typescript-examples' packages/oxlint/` cannot see types — that is settled, and
typescript-examples records why. The rules here can. They are
written in Go against tsgolint's real `checker.Checker`, and they are the only
place a type question gets asked in this repository.

## Which half you are in

| | oxlint (typescript-examples) | `archlint` |
|---|---|---|
| language | TypeScript | **Go** |
| sees | ESTree AST, file path, scope | AST **plus** the TypeScript checker |
| source | `typescript-examples' packages/oxlint/rules/` | `rules/` |
| runs under `oxlint` | yes | **no — see below** |

Write a Go rule when answering the question requires resolving a type: does
this alias describe absence, do these two annotations share a brand, is this
error channel a literal union or `string`, is this identifier a boolean.
Write a JS rule for everything else — it is faster to author, faster to run,
and it is the half `oxlint` will actually run.

**Never approximate a type question syntactically.** A rule that reads the
annotation as written stops enforcing anything the moment a codebase names its
widening, and a rule that stops matching makes the tree greener, not redder.

## These rules do not run under `oxlint --type-aware`

oxlint's native binding carries its own hardcoded list of tsgolint rule names
and only ever requests those. An unknown name is never sent to the tsgolint
process, so it does not matter that our binary can run it. Verified against
`@oxlint/binding-linux-x64-gnu@1.78.0`: `no-for-in-array` appears in the
binary, `arch-no-widened-reason-in-envelope` does not.

That is why `archlint` invokes the binary directly. Nothing is
installed over oxlint's own binary — the build script says so in its first
comment — so `oxlint --type-aware` is untouched and runs whatever it shipped
with.

Do not spend time looking for a config key, env var or plugin hook that fixes
this. There is none; the gate is compiled into oxlint.

## Why the rules live inside tsgolint's tree

`rule.Rule` is in `github.com/typescript-eslint/tsgolint/internal/rule`. Go's
`internal/` visibility means nothing outside that module can import it, so a
rule **cannot** be its own Go module, and `replace` does not help.

So `rules/<name>/` holds the source we own, and
`archlint/build.sh` copies it into a pinned clone under
`.tsgolint-build/` before compiling. Edit this repository, never `.tsgolint-build/`
— the latter is gitignored scratch and every build overwrites those
directories in it. The script reads the directory names from
`rules/`, so nothing depends on how a rule is spelled.

`archlint/` is our entrypoint, and `rules.go` in it is the whole
registry — the mirror of `typescript-examples' packages/oxlint/plugin.ts`. Upstream's `cmd/tsgolint` is not used:
it registers every rule in the binary and returns 0 whatever it finds.

`internal/` holds `archscope`, `archtypes` and `archrule`, which are
not rules; `typescript-examples' packages/oxlint/internal/` holds the JavaScript half's shared helpers for the
same reason. Every directory under `rules/` is a rule.

**A rule whose name ends in `-test` cannot be in a file named after itself.**
Go reads any `*_test.go` as a test file, and `package foo_test` as the external
test package for `foo`. So `no_double_library_in_domain_test/` holds `rule.go`,
and its package clause drops the suffix. Those two are the only exception to
naming the file after the rule.

## Anatomy of a rule

```go
var ThingIsSoRule = rule.Rule{
    Name: "thing-is-so",
    Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
        opts := utils.UnmarshalOptions[Options](options, "thing-is-so")

        return rule.RuleListeners{
            ast.KindPropertySignature: func(node *ast.Node) {
                ctx.ReportNode(node, buildThingIsSoMessage())
            },
        }
    },
}
```

- **`RuleContext`** carries `SourceFile`, `Program`, `TypeChecker`, and the
  `Report*` family (`ReportNode`, `ReportRange`, and the `WithFixes` /
  `WithSuggestions` variants, which take a thunk so the fix is only computed
  when it is needed).
- **`RuleMessage`** has `Id`, `Description` and `Help`. Keep this repository's
  convention: the description states what is wrong and what it costs, and
  `Help` ends on an imperative.
- **Options** are plain structs with `json` tags, decoded by
  `utils.UnmarshalOptions[T]`. Upstream generates these from a `schema.json`;
  ours are hand-written, so apply defaults yourself when the slice is empty.
- **Listeners** are keyed by `ast.Kind`. `rule.ListenerOnExit(kind)` fires on
  the way out.

## Resolving types

```go
t := checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, signature.Type)
utils.IsTypeFlagSet(t, checker.TypeFlagsString)
```

- `Checker_getTypeFromTypeNode` resolves a written annotation.
  `ctx.TypeChecker.GetTypeAtLocation` resolves an expression's type — a public
  method, not a shim function, along with `GetSymbolAtLocation`,
  `GetTypeOfSymbolAtLocation` and `TypeToString`. Reach for the first when the
  rule judges a declaration.
- **A symbol reaches its declaring file**, and that is the fact no syntactic
  rule can have: `checker.Type_symbol(t)` → `symbol.Declarations` →
  `ast.GetSourceFileOfNode(d).FileName()`. Whether a type was declared here, in
  `node_modules`, or in `lib.*.d.ts` is what separates `new Pool()` from
  `new Date()`, and a vendor `Context` from a domain command.
  `archtypes.DeclaringFiles` wraps it.
- **Recurse through unions yourself.** `TypeFlagsUnion` means `t.Types()` holds
  the constituents; a check that ignores this passes `string | undefined`.
- **`TypeFlagsString` is bare `string`; `TypeFlagsStringLiteral` is `"a"`.**
  Confusing them inverts the rule. Note `"a" | string` collapses to `string`
  before a rule ever sees it.
- `utils` has the ts-api-utils equivalents — `IsTypeFlagSet`,
  `GetConstrainedTypeAtLocation`, and friends. Look there before hand-rolling.

**The shim is an allowlist.** Rules import
`github.com/microsoft/typescript-go/shim/checker`, not typescript-go directly,
and `shim/checker/shim.go` is generated from `extra-shim.json`. A checker
method absent from that list is not callable, and adding one means regenerating
the shim in the pinned clone — which the build will then overwrite. Prefer an
exposed function; treat needing a new one as a signal to reconsider the rule.

## Scoping a rule to the tree it judges

**Scope is `archlint`'s, and a rule contains no scope check.** A rule declares
the tree it judges beside itself; `.archtypesrc.json` may replace it; the
entrypoint matches the file before the rule runs:

```go
var defaultFiles = []string{"**/hexagon/domain/**"}

var ThingIsSoRule = archrule.Rule{Files: defaultFiles, Rule: rule.Rule{
    Name: ruleName,
    Run:  func(ctx rule.RuleContext, options any) rule.RuleListeners { ... },
}}
```

Never the empty list: a rule that names no tree judges every file the tsconfig
includes, and `cmd/archlint`'s tests fail if one does. A rule reaches for
`archscope` only for a second path vocabulary of its own — the trees a domain
signature may name, where a row type is declared — never to decide whether it
should be looking at this file.

`internal/archscope` is the glob matcher (`**` spans segments) plus
`IsPackageDependency` / `IsStandardLibrary`. `internal/archtypes` holds
the repeated checker questions. Neither is a rule, which is why they sit in
`internal/` rather than beside the rules — `typescript-examples' packages/oxlint/internal/` holds the
JavaScript half's shared helpers for the same reason. Every directory under
`rules/` is a rule, and `cmd/archlint/rules.go` decides what registers.

## Testing

Every rule gets a `_test.go` beside it using upstream's harness:

```go
rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t,
    &ThingIsSoRule.Rule,
    []rule_tester.ValidTestCase{{Code: `...`}},
    []rule_tester.InvalidTestCase{
        {Code: `...`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thing"}}},
    })
```

Run with `./archlint/build.sh --test`. Assert a `MessageId`, never a count.

**Cover the cases that motivated writing a Go rule at all** — the alias, the
union, the generic. A Go rule whose tests only contain what a syntactic rule
would already catch has not been shown to earn its cost.

Confirm the test can fail. Breaking the type flag and watching the suite go red
takes a minute and is the only thing distinguishing a working rule from one
that reports nothing.

## Building

```bash
./archlint/build.sh   # clone if needed, inject rules, compile
./archlint/build.sh --test    # go test over the entrypoint, the helpers and every rule
node fixture-gate.mjs    # every rule still reports through the built binary
bun run test:examples-types      # in typescript-examples: no rule reports on a documented Perfect Example
.tsgolint-build/archlint --tsconfig tsconfig.json   # run it over a project
```

First build clones tsgolint and typescript-go and takes a few minutes; after
that it is incremental. The clone is pinned by `TSGOLINT_COMMIT` in the script,
with typescript-go pinned by tsgolint's own submodule.

`.tsgolint-build/src/.arch-initialised` marks a tree that has been cloned,
submoduled and patched. Delete it to force a clean re-init — a bare
`rm -rf .tsgolint-build` also works and costs the full rebuild.

The five patches in the clone's own `patches/` are applied to typescript-go with
`git am`, by absolute path — a relative `../patches/*.patch` is resolved against
the invoking shell's directory, not the `-C` target, and fails, and the shim depends on them. A tree that will not compile with
missing shim symbols is usually one where they did not apply.

## Footguns

- **Bumping `TSGOLINT_COMMIT` can break a rule silently.** The shim is
  generated, so an upstream rename removes a function; that is a compile error
  and fine. A changed *behaviour* is not. Run `./archlint/build.sh --test` after a bump.
- **Upstream's rules are compiled in but off.** They carry no `files` option,
  so each one judges every file the tsconfig includes; running them all is how
  `lint:arch-types` once printed 580 diagnostics, none of them ours, and exited
  0. `.archtypesrc.json` turns one on by name.
- **Check upstream before writing a rule.** All ~58 are still in the binary and
  can be enabled. `switch-exhaustiveness-check` already covers the exhaustiveness
  half of "returns a discriminated union"; writing a local version would have
  been a duplicate under a second name.
- **A `go test` snapshot is not evidence on a fresh clone.** `rule_tester`
  writes snapshots into `.tsgolint-build/src/internal/rule_tester/__snapshots__/`,
  which is gitignored, so the first run anywhere new writes them and passes.
  They catch a change you make locally; they do not catch one you push. Assert a
  `MessageId`, and let `node fixture-gate.mjs` be the check that survives.
- **A rule name must be unique across the whole binary**, which includes the ~58
  upstream rules, and nothing here namespaces ours apart from them. A duplicate
  silently overwrites the entry, so check before naming:

  ```bash
  grep -rhoE '^\s*Name:\s*"[a-z-]+"' .tsgolint-build/src/internal/rules/*/*.go
  ```

## Adding a rule: the checklist

1. Does it actually need the checker? If not, write a JS rule instead.
2. `rules/<snake_name>/<snake_name>.go` — `Name: "<kebab-name>"`,
   checked against the upstream names first; message with
   `Id`/`Description`/`Help`; an `archrule.Rule` whose `Files` names the tree it
   judges.
3. `<snake_name>_test.go` beside it, covering the alias and union cases.
   `rule_tester` runs the rule unscoped, so scope is not testable here — the
   fixture project is where it is proved. `rule_tester` takes `FileName`
   (resolved under the fixtures dir) and `Files` for extra sources — including
   `node_modules/<pkg>/index.d.ts`, which is how a vendor declaration is
   tested.
4. Register it in `archlint/rules.go`.
5. A fixture project at `fixtures/<kebab-name>/`, violating the rule at
   a path its `Files` matches, with `// expect: <kebab-name>` on the file, and a
   second file breaking the same rule just outside that tree with no marker.
   That pair is what proves the rule reports through the binary rather than only
   through `rule_tester`, and that it keeps its hands off its neighbours.
6. `./archlint/build.sh --test`, then break it once and confirm red.
7. `./archlint/build.sh && node fixture-gate.mjs`, then typescript-examples' `bun run test:examples-types`.
8. `docs/rules/<kebab-name>.md` — what, why, examples, options, limitations.
   Every rule has a page, and `node --test rule-docs.test.mjs` fails if one does not.
