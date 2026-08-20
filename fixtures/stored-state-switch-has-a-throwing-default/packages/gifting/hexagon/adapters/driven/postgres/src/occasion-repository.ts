// expect: stored-state-switch-has-a-throwing-default
import type { OccasionRow } from "./rows"
export const toOccasion = (row: OccasionRow): string => {
  switch (row.state) {
    case "Open":
      return "open"
    case "Settled":
      return "settled"
  }
  return "unreachable"
}
