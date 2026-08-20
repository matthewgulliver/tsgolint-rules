// Runs the built binary over `fixtures/`, one isolated project per
// rule, and checks each against the `// expect:` markers in its own sources.
//
// A rule test proves a rule reports through `rule_tester`. This proves it
// still reports through the binary a developer actually runs — the gap once
// recorded as "a green run is not evidence a rule works".
import { readdirSync, readFileSync, statSync } from "node:fs"
import { execFileSync } from "node:child_process"
import path from "node:path"

const root = import.meta.dirname
const fixtures = path.join(import.meta.dirname, "fixtures")
const binary = path.join(root, ".tsgolint-build", "archlint")

const sources = (dir) =>
  readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const full = path.join(dir, entry.name)
    if (entry.isDirectory()) return entry.name === "node_modules" ? [] : sources(full)
    return /\.tsx?$/.test(entry.name) ? [full] : []
  })

const expectationsIn = (project) => {
  const expected = new Set()
  for (const file of sources(project))
    for (const [, rule] of readFileSync(file, "utf8").matchAll(/^\/\/ expect: (\S+)$/gm))
      expected.add(`${path.relative(project, file)} ${rule}`)
  return expected
}

// The exit code is part of what is under test: a binary that prints findings
// and exits 0 is the defect this whole slice exists to remove, and a gate that
// only reads stdout passes it.
const runOn = (project) => {
  try {
    const stdout = execFileSync(binary, ["--tsconfig", "tsconfig.json"], { cwd: project, encoding: "utf8" })
    return { status: 0, stdout }
  } catch (error) {
    if (error.status === undefined) throw new Error(`${path.basename(project)}: ${error.stderr || error.message}`)
    return { status: error.status, stdout: error.stdout ?? "", stderr: error.stderr ?? "" }
  }
}

const reportedIn = (stdout) => {
  const reported = new Set()
  for (const [, file, rule] of stdout.matchAll(/^(.+?):\d+:\d+: error arch-types\((\S+)\):/gm))
    reported.add(`${file} ${rule}`)
  return reported
}

const projects = readdirSync(fixtures).filter((name) => statSync(path.join(fixtures, name)).isDirectory())
const failures = []

for (const name of projects.sort()) {
  const project = path.join(fixtures, name)
  const expected = expectationsIn(project)
  if (expected.size === 0) failures.push(`${name}: no \`// expect:\` marker in the project`)

  const { status, stdout, stderr } = runOn(project)
  if (status === 2) failures.push(`${name}: the run failed — ${stderr.trim()}`)

  const reported = reportedIn(stdout)
  for (const pair of expected) if (!reported.has(pair)) failures.push(`${name}: expected but silent — ${pair}`)
  for (const pair of reported) if (!expected.has(pair)) failures.push(`${name}: reported but unexpected — ${pair}`)

  const wanted = reported.size > 0 ? 1 : 0
  if (status !== wanted && status !== 2)
    failures.push(`${name}: ${reported.size} findings printed but exited ${status}, wanted ${wanted}`)
}

const rules = new Set(projects)
console.log(`${projects.length} fixture projects, ${rules.size} rules, ${failures.length} problems`)
for (const failure of failures) console.log(`  ${failure}`)
process.exit(failures.length === 0 ? 0 : 1)
