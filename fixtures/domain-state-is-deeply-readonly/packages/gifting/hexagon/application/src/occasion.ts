// Outside the domain and the kernel, and silent: the same mutable list behind
// a named type.
type PledgeList = string[]
export type ApplicationOccasion = { readonly pledges: PledgeList }
