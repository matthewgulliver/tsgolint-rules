// expect: domain-function-returns-an-answer
export type Occasion = { readonly settled: boolean }
declare const audit: (o: Occasion) => void
export const settle = (o: Occasion) => { audit(o) }
