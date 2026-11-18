// Helpers for the student pages tests (SCRUM-152). Owner: Nyendrak.
// They talk to the gateway directly (cy.request), to set up a report and move
// it along, so each test only clicks through the page it is about.

const api = () => Cypress.env('apiUrl') as string

type Location = { building: string; wing?: string; area?: string }

/** The first real Hostel place from location-taxonomy (SCRUM-150 data). */
export function firstHostelPlace(token: string): Cypress.Chainable<Location> {
  return cy
    .request({ url: `${api()}/categories/HOSTEL/locations`, headers: { Authorization: `Bearer ${token}` } })
    .then((res) => {
      const first = res.body.locations?.[0]?.location as Location | undefined
      expect(first, 'a Hostel location in location-taxonomy').to.not.equal(undefined)
      return first as Location
    })
}

/** Sends a Hostel report as the student and yields its id. */
export function sendHostelReport(token: string, description: string): Cypress.Chainable<string> {
  return firstHostelPlace(token).then((location) =>
    cy
      .request({
        method: 'POST',
        url: `${api()}/reports`,
        headers: { Authorization: `Bearer ${token}` },
        body: { category: 'HOSTEL', location, description },
      })
      .then((res) => {
        expect(res.status).to.eq(201)
        return res.body.report.id as string
      }),
  )
}

type Status = { status: string; version?: number }

/**
 * Asks status-lifecycle until the report has the status it should reach. The
 * routing happens over the event bus, a moment after the report is sent, so
 * this polls (at most 20 times, half a second apart) instead of a fixed wait.
 */
export function waitForStatus(token: string, reportId: string, want: string, tries = 20): Cypress.Chainable<Status> {
  return cy
    .request({
      url: `${api()}/reports/${reportId}/status`,
      headers: { Authorization: `Bearer ${token}` },
      failOnStatusCode: false,
    })
    .then((res) => {
      const got = res.status === 200 ? (res.body.status as Status) : null
      if (got?.status === want) return cy.wrap(got, { log: false })
      if (tries <= 1) throw new Error(`report ${reportId} did not reach ${want} (last: ${got?.status ?? res.status})`)
      cy.wait(500, { log: false })
      return waitForStatus(token, reportId, want, tries - 1)
    })
}

/** As the Category Head: moves the report to the next status, with its version. */
export function changeStatus(token: string, reportId: string, from: Status, toStatus: string): Cypress.Chainable<Status> {
  return cy
    .request({
      method: 'POST',
      url: `${api()}/reports/${reportId}/status`,
      headers: { Authorization: `Bearer ${token}` },
      body: { toStatus, expectedVersion: from.version ?? 0, reason: `E2E: set ${toStatus}` },
    })
    .then((res) => {
      expect(res.status).to.eq(200)
      return res.body.status as Status
    })
}

/** The demo accounts' password, needed only to act as the Category Head. */
export function demoPassword(): string {
  const p = Cypress.env('demoPassword') as string
  if (!p) throw new Error('Set CYPRESS_demoPassword to the SEED_DEMO_PASSWORD used to start identity')
  return p
}
