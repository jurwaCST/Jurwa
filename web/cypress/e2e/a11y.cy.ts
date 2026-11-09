// Accessibility checks with axe-core (SCRUM-110, NFR-06). Every page is checked
// against WCAG 2.0 and 2.1, levels A and AA. A test fails on any violation, and
// each violation is printed in the terminal with its rule, impact and element,
// so it can be sent to the page's owner. See docs/testing/accessibility.md.

import type { Result } from 'axe-core'

const WCAG = { runOnly: { type: 'tag' as const, values: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'] } }

// Prints each violation in the terminal (cypress run) and in the command log.
function report(page: string) {
  return (violations: Result[]) => {
    const lines = violations.flatMap((v) => [
      `[${page}] ${v.id} (${v.impact}): ${v.help}`,
      ...v.nodes.map((n) => `    element: ${n.target.join(' ')}`),
      `    rule: ${v.helpUrl}`,
    ])
    cy.task('log', lines.join('\n'))
  }
}

function checkPage(page: string) {
  cy.injectAxe()
  cy.checkA11y(undefined, WCAG, report(page))
}

const demoPassword = () => {
  const p = Cypress.env('demoPassword') as string
  if (!p) {
    throw new Error('Set CYPRESS_demoPassword to the SEED_DEMO_PASSWORD used to start identity')
  }
  return p
}

describe('accessibility (axe-core, WCAG 2.1 A and AA)', () => {
  it('sign-in page', () => {
    cy.visit('/login')
    cy.contains('button', 'Sign in').should('be.visible')
    checkPage('/login')
  })

  it('sign-in page showing a wrong password error', () => {
    cy.intercept('POST', '/api/auth/login').as('login')
    cy.visit('/login')
    cy.findLabel('Email').type('nobody@e2e.jurwa.example')
    cy.findLabel('Password').type('not-the-password', { log: false })
    cy.contains('button', 'Sign in').click()
    cy.wait('@login')
    cy.get('[role=alert]').should('be.visible')
    checkPage('/login with an error')
  })

  it('home page, signed out', () => {
    cy.visit('/')
    checkPage('/ signed out')
  })

  it('page not found', () => {
    cy.visit('/no-such-page')
    checkPage('404')
  })

  describe('as a student', () => {
    beforeEach(() => {
      cy.fixture('users.json').then((users) => {
        cy.registerStudent().then((email) => cy.uiLogin(email, users.studentPassword))
      })
    })

    it('home page', () => {
      cy.visit('/')
      checkPage('/ student')
    })

    it('report form', () => {
      cy.visit('/report/new')
      cy.contains('h1', 'Report an issue').should('be.visible')
      checkPage('/report/new')
    })

    it('report form with a category and building chosen', () => {
      cy.intercept('GET', '/api/categories/*/locations').as('locations')
      cy.visit('/report/new')
      cy.findLabel('Category').select('Hostel')
      cy.wait('@locations')
      cy.findLabel('Building').should('be.visible')
      checkPage('/report/new with category chosen')
    })

    it('report form showing its error messages', () => {
      cy.visit('/report/new')
      cy.contains('button', 'Send report').click()
      cy.get('.error').should('be.visible')
      checkPage('/report/new with errors')
    })

    it('a page the student may not open (403)', () => {
      cy.visit('/staff')
      checkPage('403')
    })
  })

  describe('as staff and admins (demo accounts, CYPRESS_demoPassword)', () => {
    it('Category Head view', () => {
      cy.fixture('users.json').then((users) => cy.uiLogin(users.categoryHeadDSA, demoPassword()))
      cy.visit('/staff')
      checkPage('/staff')
    })

    it('System Administrator page', () => {
      cy.fixture('users.json').then((users) => cy.uiLogin(users.admin, demoPassword()))
      cy.visit('/admin')
      checkPage('/admin')
    })

    it('College President dashboard', () => {
      cy.fixture('users.json').then((users) => cy.uiLogin(users.president, demoPassword()))
      cy.visit('/president')
      checkPage('/president')
    })
  })
})
