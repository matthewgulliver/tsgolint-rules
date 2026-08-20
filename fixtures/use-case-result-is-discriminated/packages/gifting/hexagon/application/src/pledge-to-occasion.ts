// expect: use-case-result-is-discriminated
export const pledgeToOccasion = async (): Promise<
  { readonly occasionId: string } | { readonly reason: string }
> => ({ occasionId: "o-1" })
