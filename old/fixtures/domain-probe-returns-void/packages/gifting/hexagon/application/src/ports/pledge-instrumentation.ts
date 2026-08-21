// expect: domain-probe-returns-void
export interface PledgeInstrumentation {
  readonly pledgeAccepted: (facts: PledgeFacts) => PledgeReceipt
}
