export interface PledgePersistence {
  readonly save: (id: string) => Promise<void>
}
export type PledgeResult = { readonly saved: boolean }
