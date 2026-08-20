import { describe, it } from "node:test"
import assert from "node:assert/strict"
import { execFileSync } from "node:child_process"
import { existsSync } from "node:fs"
import path from "node:path"

const root = import.meta.dirname
const binary = path.join(root, ".tsgolint-build", "archlint")

// Upstream's rules are documented by upstream; only this repository's own have
// a page here.
const registeredRules = () =>
  execFileSync(binary, ["--list-rules"], { encoding: "utf8" })
    .trim()
    .split("\n")
    .filter(Boolean)
    .map((line) => line.split("\t"))
    .filter(([, origin]) => origin === "repository")
    .map(([rule]) => rule)

describe("every rule can be read about", () => {
  it("gives each rule a page", () => {
    const rules = registeredRules()
    assert.ok(rules.length > 0, "the binary registered no rules")
    const missing = rules.filter((rule) => !existsSync(path.join(root, "docs", "rules", `${rule}.md`)))
    assert.deepEqual(missing, [], "registered with no docs/rules page")
  })

  // `meta.docs.url` has no field on upstream's `rule.Rule`, so the page is put
  // on the diagnostic instead.
  it("names the page in a diagnostic", () => {
    const project = path.join(root, "fixtures", "port-behaviour-is-an-interface")
    let stdout = ""
    try {
      stdout = execFileSync(binary, ["--tsconfig", "tsconfig.json"], { cwd: project, encoding: "utf8" })
    } catch (error) {
      stdout = error.stdout ?? ""
    }
    assert.match(
      stdout,
      /docs\/rules\/port-behaviour-is-an-interface\.md/,
      "a diagnostic did not say where to read about the rule"
    )
  })
})
