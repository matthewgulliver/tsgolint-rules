// expect: no-double-library-in-use-case-test
import { vi } from "vitest"
const persistence = { save: vi.fn() }
