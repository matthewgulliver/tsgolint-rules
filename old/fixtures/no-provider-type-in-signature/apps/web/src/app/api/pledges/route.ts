// Outside the inside, and silent: holding the transport is what a driving
// adapter is for.
import type { Context } from "hono"
export const handle = (context: Context) => context.req.url
