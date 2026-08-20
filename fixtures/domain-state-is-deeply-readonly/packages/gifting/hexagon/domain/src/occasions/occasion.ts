// expect: domain-state-is-deeply-readonly
type FundingState = { pledges: ReadonlyArray<string>; readonly closed: boolean }
export type Occasion = { readonly id: string; readonly funding: FundingState }
