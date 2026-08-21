// Outside the journey tree, and silent: the shortcut is a support helper's
// business.
import type { Page } from "@playwright/test"
declare const page: Page
page.request.post("/api/occasions")
