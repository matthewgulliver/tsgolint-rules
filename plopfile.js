/**
 * `npm run new:rule`
 *
 * Generates a lintcn rule package under `.lintcn/` — the rule and its test.
 *
 * What it emits is deliberately RED: the test asserts a violation the empty
 * `isViolation` does not yet report, so `go test` fails until you write the
 * logic. This repository's TDD policy wants the failing test first, so a
 * generator that emitted green scaffolding would be handing you a started
 * cycle with the RED step already skipped.
 *
 * There is no registry to wire into: lintcn discovers a rule by scanning
 * `.lintcn/*​/` for the inline `Name: "..."` literal.
 */

import { execFileSync } from "node:child_process";

const RULE_ID = /^[a-z][a-z0-9]*(-[a-z0-9]+)*$/;
const DESCRIPTION_PREFIX = /^(Disallow|Require|Enforce)\b/;

/**
 * A rule whose name ends in "-test" collides with Go's test conventions twice,
 * so it deviates from the usual `<snake>/<snake>.go` + `package <snake>` shape:
 *
 *   - `<snake>.go` would end in `_test.go`, which Go reads as a test file — the
 *     rule source disappears and the package has no non-test files. The rule
 *     file is named `rule.go` instead.
 *   - `package <snake>` would end in `_test`, which names an external test
 *     package. The package takes the trimmed name.
 *
 * Both double-library rules are this shape. The test file keeps `<snake>_test.go`
 * (so, doubled: `..._test_test.go`), which is correct and only looks wrong.
 */
const endsInTest = (snake) => snake.endsWith("_test");
const packageNameFor = (snake) => snake.replace(/_test$/, "");
const ruleFileFor = (snake) => (endsInTest(snake) ? "rule" : snake);

export default function (plop) {
  plop.setGenerator("rule", {
    description: "Create a new lintcn rule with its test",
    prompts: [
      {
        type: "input",
        name: "name",
        message: "Rule ID (kebab-case, as it appears on the CLI):",
        validate: (value) =>
          RULE_ID.test(value) || "Must be kebab-case: lowercase words separated by dashes.",
      },
      {
        type: "input",
        name: "description",
        message: "Description (starts with Disallow/Require/Enforce):",
        validate: (value) =>
          DESCRIPTION_PREFIX.test(value) ||
          "Must start with `Disallow`, `Require`, or `Enforce` — it is the one-liner `npx lintcn list` prints.",
      },
      {
        type: "list",
        name: "severity",
        message: "Severity:",
        choices: [
          {
            name: "error  — an architectural invariant; fails CI",
            value: "error",
          },
          {
            name: "warn   — advisory; changed and untracked files only, never fails CI",
            value: "warn",
          },
        ],
      },
      {
        // Gating is derived rather than asked: a rule scoped to a tree is
        // exactly one that has patterns. A separate confirm would let the two
        // answers contradict each other, and `when:` makes the generator
        // impossible to drive non-interactively.
        type: "input",
        name: "files",
        message: "Default glob(s) to scope it to, comma-separated (blank = every file):",
      },
      {
        type: "confirm",
        name: "hasOptions",
        message: "Configurable via options?",
        default: false,
      },
    ],

    actions(answers) {
      const snakeName = answers.name.replace(/-/g, "_");
      const globs = (answers.files ?? "")
        .split(",")
        .map((glob) => glob.trim())
        .filter(Boolean);

      Object.assign(answers, {
        snakeName,
        packageName: packageNameFor(snakeName),
        ruleFile: ruleFileFor(snakeName),
        ruleVar: `${plop.getHelper("pascalCase")(answers.name)}Rule`,
        messageId: plop.getHelper("camelCase")(answers.name),
        testName: plop.getHelper("pascalCase")(answers.name),
        isGated: globs.length > 0,
        globs: globs.map((glob) => `"${glob}"`).join(", "),
      });

      return [
        {
          type: "add",
          path: ".lintcn/{{snakeName}}/{{ruleFile}}.go",
          templateFile: "plop-templates/rule.go.hbs",
        },
        {
          type: "add",
          path: ".lintcn/{{snakeName}}/{{snakeName}}_test.go",
          templateFile: "plop-templates/rule_test.go.hbs",
        },
        // Handlebars conditionals cannot help but disturb Go's whitespace, and
        // hand-tuning the templates to survive gofmt would break again on the
        // next edit. Let the formatter be the authority instead.
        () => {
          execFileSync("gofmt", ["-w", snakeName], { cwd: ".lintcn" });
          return `gofmt -w .lintcn/${snakeName}`;
        },
        () =>
          [
            "",
            "Next:",
            "  cd .lintcn",
            `  TSGOLINT_SNAPSHOT_CWD=true go test ./${snakeName}/   # RED, as intended`,
            "  # write the logic, then:",
            `  TSGOLINT_SNAPSHOT_CWD=true UPDATE_SNAPS=true go test ./${snakeName}/`,
            `  gremlins unleash ./${snakeName}/`,
            "",
            "Read the snapshot before committing — it is the agent-facing output.",
          ].join("\n"),
      ];
    },
  });
}
