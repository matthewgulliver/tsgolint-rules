export type Occasion = { readonly id: string }
export type PledgeDecision =
  | { readonly success: true; readonly occasion: Occasion }
  | { readonly success: false; readonly reason: "funding-closed" }
