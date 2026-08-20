// Outside the domain tree, and silent: the same answerless export.
type Occasion = { readonly id: string; readonly settled: boolean }
declare const audit: (o: Occasion) => void
export const settle = (o: Occasion) => { audit(o) }
