export type SettleCommand = {
  readonly id: string
  readonly state: "Open" | "Settled"
}
