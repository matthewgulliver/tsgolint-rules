// Outside the domain tree, and silent: naming its own port is the
// application's business.
import type { PledgePersistence } from "./ports/pledge-persistence"
export const run = (persistence: PledgePersistence): void => {}
