// expect: read-port-returns-an-answer
export interface OccasionDashboardRows {
  readonly recordDashboardView: (id: string) => Promise<void>
}
