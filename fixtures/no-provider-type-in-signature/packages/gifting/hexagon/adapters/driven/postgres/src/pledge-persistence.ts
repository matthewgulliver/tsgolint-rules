// Inside `hexagon/`, outside the inside, and silent: `**/hexagon/**` would have
// caught this adapter, and the tree this rule names is the inside as a whole.
import type { Pool } from "pg"
export const save = (pool: Pool) => pool.query("select 1")
