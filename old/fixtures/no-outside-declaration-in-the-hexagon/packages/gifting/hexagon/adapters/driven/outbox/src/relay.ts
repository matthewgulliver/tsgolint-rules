// Outside the hexagon, and silent: one adapter importing another is outside
// talking to outside.
import { PostgresPledgeRepository } from "../../postgres/src/repo"
export const outbox = new PostgresPledgeRepository()
