import { describe, it } from "node:test"
import assert from "node:assert/strict"
import { execFileSync } from "node:child_process"
import { readFileSync, statSync, writeFileSync } from "node:fs"
import path from "node:path"

const root = path.resolve(import.meta.dirname, "..")
const binary = path.join(root, ".tsgolint-build", "archlint")

const build = () => execFileSync("./archlint/build.sh", { cwd: root, encoding: "utf8" })
const builtAt = () => statSync(binary, { bigint: true }).mtimeNs

const whileChanged = (file, append, check) => {
  const original = readFileSync(file, "utf8")
  writeFileSync(file, original + append)
  try {
    check()
  } finally {
    writeFileSync(file, original)
    build()
  }
}

describe("archlint build", () => {
  it("leaves the binary alone when nothing that reaches it has changed", () => {
    build()
    const before = builtAt()
    build()
    assert.equal(builtAt(), before, "an unchanged tree rebuilt anyway")
  })

  it("rebuilds when a rule changes, so the linter never runs the previous rule", () => {
    build()
    const before = builtAt()
    whileChanged(
      path.join(root, "rules/domain_probe_returns_void/domain_probe_returns_void.go"),
      "\n// changed by a test\n",
      () => {
        build()
        assert.notEqual(builtAt(), before, "a changed rule did not reach the binary")
      }
    )
  })

  it("rebuilds when the registry changes", () => {
    build()
    const before = builtAt()
    whileChanged(path.join(root, "archlint/rules.go"), "\n// changed by a test\n", () => {
      build()
      assert.notEqual(builtAt(), before, "a changed registry did not reach the binary")
    })
  })

  it("leaves the binary alone when only a fixture changes", () => {
    build()
    const before = builtAt()
    whileChanged(
      path.join(root, "fixtures/domain-probe-returns-void/tsconfig.json"),
      "\n",
      () => {
        build()
        assert.equal(builtAt(), before, "a fixture is not compiled into anything")
      }
    )
  })
})
