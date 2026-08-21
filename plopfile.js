import { execFileSync } from "node:child_process";

const RULE_ID = /^[a-z][a-z0-9]*(-[a-z0-9]+)*$/;
const DESCRIPTION_PREFIX = /^(Disallow|Require|Enforce)\b/;

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
