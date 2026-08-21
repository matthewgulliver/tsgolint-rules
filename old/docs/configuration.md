# Configuring archlint

`.archtypesrc.json`, beside the target project's tsconfig, in `.oxlintrc.json`'s
grammar — a `rules` map of name to `"error"`/`"off"`, and `overrides` blocks
that re-decide a rule for the files they name. A repository with no such file
gets every rule on the defaults its own Go source declares.

```json
{
  "rules": { "port-behaviour-is-an-interface": "off" },
  "overrides": [
    {
      "files": ["**/contracts/**"],
      "rules": {
        "domain-state-is-deeply-readonly": ["error", { "mutableCollectionTypeNames": ["Array"] }]
      }
    }
  ]
}
```

A block's `files` are the tree its rules judge, so an override can widen a rule
past the tree its source names; a rule that sets `files` itself keeps it. Later
blocks beat earlier ones. A name no rule answers to fails the run rather than
silently doing nothing. There is no `warn`: this binary either fails a build or
does not.

**Scope is decided once, in `archlint`.** Each rule declares the tree it judges
beside itself, the configuration may replace it, and the entrypoint matches the
file against whichever won before the rule runs — so a rule contains no scope
check of its own. A rule that declared no tree would report on every file in
the tsconfig; `archlint`'s own tests fail if a rule of ours names no tree. The
shared matcher is [`internal/archscope`](../internal/archscope/archscope.go);
the shared checker questions are in
[`archtypes`](../internal/archtypes/archtypes.go).

**Upstream's ~58 rules are compiled in but off.** They declare no tree — each
one judges every file the tsconfig includes — so none run unless
`.archtypesrc.json` names them:

```json
{ "rules": { "switch-exhaustiveness-check": "error" } }
```
