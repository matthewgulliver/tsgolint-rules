// expect: no-double-library-in-domain-test
// The domain test tree belongs to the sibling rule, which takes it here.
import { vi } from "vitest"
vi.fn()
