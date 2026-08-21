package no_page_request_in_journey

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const inE2E = "apps/web/e2e/pledge-to-occasion.spec.ts"

var playwright = map[string]string{
	"node_modules/@playwright/test/index.d.ts": `
    export interface APIRequestContext { post(path: string): Promise<void> }
    export interface Page {
      readonly request: APIRequestContext
      getByRole(role: string): { click(): Promise<void> }
    }
  `,
}

func TestNoPageRequestInJourney(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoPageRequestInJourneyRule, []rule_tester.ValidTestCase{
		{
			Code: `
        import type { Page } from "@playwright/test"
        declare const page: Page
        page.getByRole("button").click()
      `,
			FileName: inE2E,
			Files:    playwright,
		},
		{
			Code: `
        declare const page: { readonly request: { post(path: string): Promise<void> } }
        page.request.post("/api/occasions")
      `,
			FileName: inE2E,
		},
		{
			Code: `
        import type { Page } from "@playwright/test"
        declare const page: Page
        page.request.post("/api/occasions")
      `,
			Files: playwright,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        import type { Page } from "@playwright/test"
        declare const page: Page
        page.request.post("/api/occasions")
      `,
			FileName: inE2E,
			Files:    playwright,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "pageRequest"}},
		},
		{
			Code: `
        import type { Page } from "@playwright/test"
        type JourneyPage = Page
        declare const page: JourneyPage
        page.request.post("/api/occasions")
      `,
			FileName: inE2E,
			Files:    playwright,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "pageRequest"}},
		},
	})
}
