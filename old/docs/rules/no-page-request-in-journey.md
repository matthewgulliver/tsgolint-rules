# no-page-request-in-journey

Type-aware. Written in Go, run by `archlint`, **not** by
`oxlint`.

### What it does

Reports `page.request` in an E2E tree only when the receiver resolves to
Playwright's `Page` type. An unrelated local object with a `request` member,
and a test's standalone API `request` fixture, are not reported.

### Why is this bad?

[`e2e-test.md`](https://github.com/matthewgulliver/typescript-examples/blob/main/docs/examples/tests/e2e-test.md) reserves a journey claim for
work the browser caused through accessible UI. Playwright's `page.request` is
Node-side HTTP: it can pass while the UI handler, browser cookies, CSRF/Fetch
Metadata, redirects, or rendering are broken.

### Why it is type-aware

`page.request` is not enough: applications can have an unrelated `page`
object with a `request` member. The checker resolves the receiver's declaration
and only reports types declared by `@playwright/*` under `node_modules`.

### Examples

Incorrect:

```ts
import type { Page } from "@playwright/test"
declare const page: Page
await page.request.post("/api/occasions/occasion-1/pledges")
```

Correct:

```ts
await page.getByRole("button", { name: /pledge/i }).click()
```

### Options

| Option | Type | Default | What it does |
|---|---|---|---|
| `files` | `string[]` | `[`"**/e2e/**"`]` | E2E trees the rule judges. |

### Limitations

The rule cannot prove that a locator represents an accessible user action, that
a request was observed rather than manufactured, or that a standalone API test
is named honestly. It deliberately only prevents the documented `page.request`
shortcut.
