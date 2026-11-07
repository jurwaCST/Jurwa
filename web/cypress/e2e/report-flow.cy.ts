// The report flow from a student's point of view (SCRUM-108). Steps that need
// parts not merged yet are listed in docs/testing/e2e-tests.md, "Known gaps".

const api = () => Cypress.env('apiUrl') as string

describe('report flow', () => {
  let studentEmail: string
  let studentPassword: string

  beforeEach(() => {
    cy.fixture('users.json').then((users) => {
      studentPassword = users.studentPassword
    })
    cy.registerStudent().then((email) => {
      studentEmail = email
    })
  })

  it('a student signs in through the UI', () => {
    cy.intercept('POST', '/api/auth/login').as('login')
    cy.visit('/login')

    cy.findLabel('Email').type(studentEmail)
    cy.findLabel('Password').type(studentPassword, { log: false })
    cy.contains('button', 'Sign in').click()

    cy.wait('@login').its('response.statusCode').should('eq', 200)
    cy.contains('Signed in as E2E Student').should('be.visible')
  })

  it('a wrong password shows an error and stays on the sign-in page', () => {
    cy.intercept('POST', '/api/auth/login').as('login')
    cy.visit('/login')

    cy.findLabel('Email').type(studentEmail)
    cy.findLabel('Password').type('not-the-password', { log: false })
    cy.contains('button', 'Sign in').click()

    cy.wait('@login').its('response.statusCode').should('eq', 401)
    cy.get('[role=alert]').should('be.visible')
    cy.location('pathname').should('eq', '/login')
  })

  it('a student submits a Hostel report and can read it back as OPEN', () => {
    cy.apiLogin(studentEmail, studentPassword).then((token) => {
      const auth = { Authorization: `Bearer ${token}` }
      cy.request({
        method: 'POST',
        url: `${api()}/reports`,
        headers: auth,
        body: {
          category: 'HOSTEL',
          location: { building: 'Hostel A', wing: 'Wing 1', area: 'Bathroom 1' },
          description: 'E2E test: the tap in bathroom 1 is leaking',
        },
      }).then((res) => {
        expect(res.status).to.eq(201)
        const report = res.body.report
        expect(report.id).to.match(/^[0-9a-f-]{36}$/)
        expect(report.status).to.eq('OPEN')

        cy.request({ url: `${api()}/reports/${report.id}`, headers: auth }).then((got) => {
          expect(got.status).to.eq(200)
          expect(got.body.id).to.eq(report.id)
          expect(got.body.category).to.eq('HOSTEL')
          expect(got.body.location.building).to.eq('Hostel A')
        })
      })
    })
  })

  it('a description of 501 characters is refused by the server', () => {
    cy.apiLogin(studentEmail, studentPassword).then((token) => {
      cy.request({
        method: 'POST',
        url: `${api()}/reports`,
        headers: { Authorization: `Bearer ${token}` },
        body: {
          category: 'HOSTEL',
          location: { building: 'Hostel A', wing: 'Wing 1', area: 'Bathroom 1' },
          description: 'x'.repeat(501),
        },
        failOnStatusCode: false,
      }).its('status').should('eq', 400)
    })
  })

  it('another student cannot read the report', () => {
    cy.apiLogin(studentEmail, studentPassword).then((token) => {
      cy.request({
        method: 'POST',
        url: `${api()}/reports`,
        headers: { Authorization: `Bearer ${token}` },
        body: {
          category: 'HOSTEL',
          location: { building: 'Hostel A', wing: 'Wing 1', area: 'Bathroom 1' },
          description: 'E2E test: private to its reporter',
        },
      }).then((res) => {
        const id = res.body.report.id
        cy.registerStudent().then((other) => {
          cy.apiLogin(other, studentPassword).then((otherToken) => {
            cy.request({
              url: `${api()}/reports/${id}`,
              headers: { Authorization: `Bearer ${otherToken}` },
              failOnStatusCode: false,
            }).its('status').should('eq', 403)
          })
        })
      })
    })
  })
})
