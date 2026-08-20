// expect: driving-port-command-is-modelled
export interface ForPledgingToOccasions {
  readonly pledgeToOccasion: (occasionId: string) => Promise<void>
}
