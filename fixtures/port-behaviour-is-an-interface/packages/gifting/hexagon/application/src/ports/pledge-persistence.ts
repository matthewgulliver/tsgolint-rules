// expect: port-behaviour-is-an-interface
// The callability is behind an alias, so the syntactic rule reads it as data.
type PledgeCommand = { readonly occasionId: string; readonly pledgeId: string }
type PledgeToOccasion = (command: PledgeCommand) => Promise<string>
export type ForPledgingToOccasions = { readonly pledgeToOccasion: PledgeToOccasion }
