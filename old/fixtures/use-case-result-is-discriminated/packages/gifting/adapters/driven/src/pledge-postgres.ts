// Outside the judged trees, and silent: an adapter's own union answers to
// nobody's discriminant.
export const decide = (): { readonly a: string } | { readonly b: string } => ({ a: "" })
