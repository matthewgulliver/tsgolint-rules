// expect: no-double-library-in-use-case-test
// The application test tree belongs to the sibling rule, which takes it here.
import { vi } from "vitest"
vi.fn()
