#!/usr/bin/env bash
# Builds `archlint`: tsgolint's engine, this repo's rules, and our own
# entrypoint in place of upstream's.
# Nothing installs it over oxlint's binary: oxlint only requests rule names
# compiled into its own binding, so these rules are run by this binary directly.
set -euo pipefail

TSGOLINT_COMMIT=92e0f4dac8eb767bd6ed523c934f798bf44d755e

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
build="$root/.tsgolint-build"
src="$build/src"

# Our rule packages, read from the tree rather than matched by a name prefix.
rules=()
for dir in "$root"/rules/*/; do rules+=("$(basename "$dir")"); done

# Everything that ends up inside the binary, as one hash: the rules package and
# the entrypoint that registers them, which is an app and lives outside it.
# Go sources only — `fixtures/` is the binary's subject, and the gates beside
# the rules run it rather than compile into it.
#
# Without this, `lint:arch-types` ran whatever was compiled last — so editing a
# rule and running the linter reported the previous rule's behaviour, silently.
fingerprint() {
  {
    echo "$TSGOLINT_COMMIT"
    find "$root/rules" "$root/internal" -type f -name '*.go' -exec sha256sum {} +
    find "$root/archlint" -type f -name '*.go' -exec sha256sum {} +
    sha256sum "$root/archlint/build.sh"
  } | LC_ALL=C sort | sha256sum | cut -d " " -f 1
}

stamp="$build/archlint.fingerprint"
want="$(fingerprint)"

if [ $# -eq 0 ] && [ -x "$build/archlint" ] && [ "$(cat "$stamp" 2>/dev/null)" = "$want" ]; then
  exit 0
fi

if [ ! -e "$src/.arch-initialised" ]; then
  rm -rf "$src"
  git clone --filter=blob:none https://github.com/oxc-project/tsgolint "$src"
  git -C "$src" checkout --detach "$TSGOLINT_COMMIT"
  git -C "$src" submodule update --init --depth 1
  # The shim these rules compile against lives in the patches, not upstream.
  # By absolute path: a relative glob resolves against the invoking shell's
  # directory, not the `-C` target.
  git -C "$src/typescript-go" am --3way --no-gpg-sign "$src"/patches/*.patch
  mkdir -p "$src/internal/collections"
  find "$src/typescript-go/internal/collections" -type f ! -name '*_test.go' \
    -exec cp {} "$src/internal/collections/" \;
  touch "$src/.arch-initialised"
fi

for name in "${rules[@]}"; do
  rm -rf "${src:?}/internal/rules/$name"
done
cp -R "$root"/rules/*/ "$src/internal/rules/"

# Shared helpers, which are not rules. They lived under `rules/` until the tree
# came to mirror `packages/oxlint/`, so a clone initialised before that still
# holds them.
rm -rf "${src:?}/internal/rules/archscope" "${src:?}/internal/rules/archtypes"
for dir in "$root"/internal/*/; do
  rm -rf "${src:?}/internal/$(basename "$dir")"
done
cp -R "$root"/internal/*/ "$src/internal/"

# Our entrypoint, not upstream's: `cmd/tsgolint` registers every rule in the
# binary and exits 0 whatever it finds. These files compile only inside the clone
# and takes upstream's `cmd/` shape only inside the build tree.
rm -rf "${src:?}/cmd/archlint"
cp -R "$root/archlint" "$src/cmd/archlint"

if [ "${1:-}" = "--test" ]; then
  packages=("./cmd/archlint")
  for dir in "$root"/internal/*/; do packages+=("./internal/$(basename "$dir")"); done
  for name in "${rules[@]}"; do packages+=("./internal/rules/$name"); done
  (cd "$src" && go test -p "${GO_TEST_PARALLEL:-4}" "${packages[@]}")
  exit 0
fi

(cd "$src" && go build -ldflags="-s -w" -trimpath -o "$build/archlint" ./cmd/archlint)
echo "$want" > "$stamp"
echo "built archlint -> $build/archlint"
