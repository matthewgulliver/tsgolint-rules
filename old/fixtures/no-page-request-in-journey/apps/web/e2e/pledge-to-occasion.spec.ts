// expect: no-page-request-in-journey
import type { Page } from "@playwright/test"
declare const page: Page
page.request.post("/api/occasions")
