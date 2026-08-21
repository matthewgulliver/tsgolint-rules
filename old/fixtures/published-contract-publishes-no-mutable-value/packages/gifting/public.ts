// expect: published-contract-publishes-no-mutable-value
const buildRegistry = (): Map<string, string> => new Map()
export const REGISTRY = buildRegistry()
