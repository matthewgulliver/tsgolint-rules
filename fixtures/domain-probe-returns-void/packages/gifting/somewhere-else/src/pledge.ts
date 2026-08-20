// Outside a ports tree, and silent: the same probe, answering.
type PledgeFacts = { readonly id: string }
type PledgeReceipt = { readonly acknowledged: true }
export interface PledgeInstrumentation {
  readonly pledgeAccepted: (facts: PledgeFacts) => PledgeReceipt
}
