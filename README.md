# tsgolint-rules

Type-aware architecture lint rules for TypeScript, written in Go against
[tsgolint](https://github.com/oxc-project/tsgolint)'s checker and compiled into
`archlint`, a standalone binary. Extracted from
[typescript-examples](https://github.com/matthewgulliver/typescript-examples),
whose `docs/examples/` hold the reasoning the rules enforce.

- `rules/` — one package per rule; `docs/rules/` documents each one
- `internal/` — `archrule`, `archscope`, `archtypes`: shared helpers, not rules
- `archlint/` — the entrypoint and `build.sh`, which assembles a pinned tsgolint
  clone and compiles the rules into it
- `fixtures/` — one project per rule, proving it reports through the built binary

```bash
./archlint/build.sh                                  # -> .tsgolint-build/archlint
./archlint/build.sh --test                           # go test over every rule
node fixture-gate.mjs                                # rules report through the binary
.tsgolint-build/archlint --tsconfig tsconfig.json    # run it over a project
```

`docs/writing-rules.md` is the guide for adding or changing a rule.
