// Outside a published contract, and silent: a module's own mutable map is not
// state another context reads.
export const registry = new Map<string, string>()
