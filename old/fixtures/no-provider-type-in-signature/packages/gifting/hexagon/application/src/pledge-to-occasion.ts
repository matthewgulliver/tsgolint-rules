// expect: no-provider-type-in-signature
import type { Context } from "hono"
export const pledge = (context: Context) => context.req.url
