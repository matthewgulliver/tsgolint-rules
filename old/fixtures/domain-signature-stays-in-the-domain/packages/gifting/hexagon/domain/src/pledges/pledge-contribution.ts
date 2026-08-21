// expect: domain-signature-stays-in-the-domain
import type { PledgePersistence } from "../../../application/src/ports/pledge-persistence"
import type { Occasion } from "../occasions/occasion"
export const pledgeContribution = (
  occasion: Occasion,
  persistence: PledgePersistence,
): Occasion => occasion
