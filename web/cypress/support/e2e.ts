// Shared commands for the end-to-end tests (SCRUM-108) and the accessibility
// checks (SCRUM-110).

import 'cypress-axe'

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
      /** Uploads a photo the way the web app does (multipart through /api). */
      uploadPhoto(token: string, bytes: Uint8Array, type: string, name: string): Chainable<UploadResult>
      /** Signs in through the login page, the way a person does. */
      uiLogin(email: string, password: string): Chainable<void>
    }
  }
  interface UploadResult {
    status: number
    body: { photoId?: string; error?: string }
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

// The browser's own fetch and FormData, through the dev server's /api proxy,
// so the multipart body is exactly what the report form will send.
Cypress.Commands.add('uploadPhoto', (token: string, bytes: Uint8Array, type: string, name: string) => {
  cy.visit('/login')
  return cy.window().then((win) => {
    const form = new win.FormData()
    form.append('photo', new win.Blob([new Uint8Array(bytes)], { type }), name)
    return win
      .fetch('/api/photos', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: form })
      .then(async (res) => ({ status: res.status, body: await res.json() }))
  })
})

Cypress.Commands.add('uiLogin', (email: string, password: string) => {
  cy.intercept('POST', '/api/auth/login').as('uiLogin')
  cy.visit('/login')
  cy.findLabel('Email').type(email)
  cy.findLabel('Password').type(password, { log: false })
  cy.contains('button', 'Sign in').click()
  cy.wait('@uiLogin').its('response.statusCode').should('eq', 200)
  cy.location('pathname').should('not.eq', '/login')
})
