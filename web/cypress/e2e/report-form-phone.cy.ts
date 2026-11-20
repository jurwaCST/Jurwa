// The report form on a phone (SCRUM-35 layout). Owner: Nyendrak. Found at the
// rehearsal: on an iPhone the form was wider than the screen, so the category
// tiles and the Send report bar were cut off on the right. A select is as wide
// as its longest option, and the form's grids had no column limit on phones.
//
// Needs only the web app (vite preview): the API is stubbed and the session is
// put in sessionStorage, so no services have to run.

const PHONE = { width: 390, height: 844 } // iPhone 12 to 15

const LONG_BUILDING = 'Hostel Block C, the new boys hostel next to the football ground and the canteen'

const session = {
  token: 'e2e-stub-token',
  expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
  user: { id: '00000000-0000-4000-8000-0000000000aa', email: 'phone@e2e.jurwa.example', fullName: 'Phone Student', roles: ['STUDENT'] },
}

/** Fails when an element's right edge is past the right edge of the screen. */
function insideScreen(selector: string) {
  cy.get(selector)
    .filter(':visible')
    .each(($el) => {
      const right = $el[0].getBoundingClientRect().right
      expect(right, `${selector} right edge`).to.be.at.most(PHONE.width)
    })
}

describe('report form on a phone', () => {
  beforeEach(() => {
    cy.viewport(PHONE.width, PHONE.height)
    cy.intercept('GET', '/api/me/notifications*', { notifications: [] })
    cy.intercept('GET', '/api/categories', {
      categories: [
        { category: 'HOSTEL', name: 'Hostel', sortOrder: 1 },
        { category: 'ACADEMIC_BLOCK', name: 'Academic block', sortOrder: 2 },
        { category: 'OTHERS', name: 'Others', sortOrder: 3 },
      ],
    })
    cy.intercept('GET', '/api/categories/HOSTEL/locations', {
      locations: [
        { id: 'l1', category: 'HOSTEL', location: { building: LONG_BUILDING, wing: 'Ground floor', area: 'Common room' } },
        { id: 'l2', category: 'HOSTEL', location: { building: 'Hostel A', wing: 'First floor', area: 'Washroom' } },
      ],
    }).as('locations')

    cy.visit('/report/new', {
      onBeforeLoad(win) {
        win.sessionStorage.setItem('jurwa.session', JSON.stringify(session))
      },
    })
    cy.contains('h1', 'Report an issue').should('be.visible')
  })

  it('fits the screen before a category is chosen', () => {
    insideScreen('.report-tiles')
    insideScreen('.report-submit-bar')
    cy.document().its('documentElement.scrollWidth').should('be.at.most', PHONE.width)
  })

  it('still fits the screen with a long building name chosen', () => {
    cy.contains('.report-tile', 'Hostel').click()
    cy.wait('@locations')
    cy.findLabel('Building').select(LONG_BUILDING)

    insideScreen('.report-tiles')
    insideScreen('.report-page select')
    insideScreen('.report-page input[type=file]')
    insideScreen('.report-submit-bar')
    cy.document().its('documentElement.scrollWidth').should('be.at.most', PHONE.width)
  })
})
