// Shared commands for the end-to-end tests (SCRUM-108).

declare global {
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace Cypress {
    interface Chainable {
      /** Registers a new Student through the gateway and yields its email. */
      registerStudent(): Chainable<string>
      /** Logs in through the gateway and yields the access token. */
      apiLogin(email: string, password: string): Chainable<string>
      /** The input inside the <label> with this text, the way a user finds it. */
      findLabel(text: string): Chainable<JQuery<HTMLElement>>
    }
  }
}

const api = () => Cypress.env('apiUrl') as string

Cypress.Commands.add('registerStudent', () => {
  return cy.fixture('users.json').then((users) => {
    // A new address each run, so the test never depends on earlier runs.
    const email = `student-${Date.now()}@${users.studentEmailDomain}`
    return cy
      .request('POST', `${api()}/auth/register`, {
        email,
        fullName: 'E2E Student',
        password: users.studentPassword,
      })
      .then((res) => {
        expect(res.status).to.eq(201)
        return email
      })
  })
})

Cypress.Commands.add('apiLogin', (email: string, password: string) => {
  return cy.request('POST', `${api()}/auth/login`, { email, password }).then((res) => {
    expect(res.status).to.eq(200)
    return res.body.accessToken as string
  })
})

Cypress.Commands.add('findLabel', (text: string) => {
  return cy.contains('label', text).find('input, select, textarea')
})

export {}
