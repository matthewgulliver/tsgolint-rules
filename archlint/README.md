# archlint

Runs `packages/tsgolint`'s type-aware rules over a TypeScript project and fails when one reports.
`rules.go` is the registry — a rule absent from it does not run, however complete its package looks.

There is no `go build` here: these files compile only inside a cloned tsgolint checkout, which
`build.sh` assembles and where they take upstream's `cmd/archlint` shape.

```bash
./archlint/build.sh              # -> .tsgolint-build/archlint
./archlint/build.sh --test       # go test, this entrypoint and every rule
node fixture-gate.mjs            # every rule still reports through the built binary
.tsgolint-build/archlint --help
```
