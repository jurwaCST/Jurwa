// The student pages (SCRUM-152): My reports, the report page with its timeline,
// notifications and feedback stars. Owner: Nyendrak. The pages are Gayley's
// (SCRUM-143, 144, 145). Needs the stack with Redis (routing sets Assigned) and
// the notification service; see docs/testing/e2e-tests.md.

import { changeStatus, demoPassword, sendHostelReport, waitForNotification, waitForStatus } from '../support/student'

/** The main navigation that is on screen (top bar on a laptop, tab bar on a phone). */
const nav = () => cy.get('nav[aria-label=Main]').filter(':visible')

describe('student pages', () => {
  let email: string
  let password: string

  beforeEach(() => {
    cy.fixture('users.json').then((users) => {
      password = users.studentPassword
      cy.registerStudent().then((e) => {
        email = e
        cy.uiLogin(email, password)
      })
    })
  })

  it('a report sent in the form shows in My reports and its page shows Open, then Assigned', () => {
    const description = `E2E ${Date.now()}: the light in the corridor is not working`

    // The form (SCRUM-35): category, the first real place, a description.
    cy.intercept('GET', '/api/categories/*/locations').as('locations')
    cy.intercept('POST', '/api/reports').as('send')
    nav().contains('a', 'Report').click()
    cy.contains('.report-tile', 'Hostel').click()
    cy.wait('@locations')
    cy.findLabel('Building').select(1)
    cy.get('.report-place').then(($place) => {
      // Hostels keep the floor in the wing field (SCRUM-150, SCRUM-154).
      if ($place.find('label:contains("Floor")').length) cy.findLabel('Floor').select(1)
    })
    cy.findLabel('Area').select(1)
    cy.findLabel('Description').type(description)
    cy.contains('button', 'Send report').click()

    cy.wait('@send').then(({ response }) => {
      expect(response?.statusCode).to.eq(201)
      const id = response?.body.report.id as string
      cy.contains('[role=status]', 'Report sent').should('contain', id)

      // routing-engine sends report.routed and status-lifecycle sets Assigned.
      cy.apiLogin(email, password).then((token) => waitForStatus(token, id, 'ASSIGNED'))

      // My reports (SCRUM-143): the one report, with its status.
      nav().contains('a', 'My reports').click()
      cy.contains('h1', 'My reports').should('be.visible')
      cy.get('.myr-card').should('have.length', 1).first().as('card')
      cy.get('@card').find('.badge').should('contain', 'Assigned')

      // The report page: what was sent, and the timeline.
      cy.get('@card').find('a.myr-place').click()
      cy.location('pathname').should('eq', `/reports/${id}`)
      cy.contains('section', 'What you reported').should('contain', description)
      cy.contains('section', 'What happened').within(() => {
        cy.contains('Open').should('be.visible')
        cy.contains('Assigned').should('be.visible')
      })
    })
  })

  it('the bell opens Notifications, which shows the list or the empty state', () => {
    cy.intercept('GET', '/api/me/notifications*').as('notifications')
    cy.get('a[aria-label^="Notifications"]').filter(':visible').click()

    cy.location('pathname').should('eq', '/notifications')
    cy.contains('h1', 'Notifications').should('be.visible')
    cy.wait('@notifications').its('response.statusCode').should('eq', 200)
    cy.get('.nt-page').should(($page) => {
      const list = $page.find('.nt-item').length > 0
      const empty = $page.text().includes('No notifications yet')
      expect(list || empty, 'the list or the empty state').to.equal(true)
    })
    cy.get('.nt-page [role=alert]').should('not.exist')
  })

  it('after a report is sent, the bell shows new and Notifications lists "We received your report"', () => {
    // The notification service turns report.submitted into a message (SCRUM-155).
    cy.apiLogin(email, password).then((token) => {
      sendHostelReport(token, `E2E ${Date.now()}: the door of the common room is broken`)
      waitForNotification(token, 'We received your report')
    })

    cy.visit('/')
    cy.get('a[aria-label="Notifications, new"]').filter(':visible').click()
    cy.location('pathname').should('eq', '/notifications')
    cy.contains('.nt-item', 'We received your report').should('contain', 'You can follow it in My reports.')
    // On the page itself the dot is gone.
    cy.get('a[aria-label="Notifications"]').filter(':visible').should('exist')
  })

  it('on a resolved report the student gives four stars and sees the thank-you', () => {
    cy.apiLogin(email, password).then((token) => {
      sendHostelReport(token, `E2E ${Date.now()}: a broken window, now fixed`).then((id) => {
        // The DSA Category Head (Hostel rule) moves it to Resolved.
        cy.fixture('users.json').then((users) => {
          cy.apiLogin(users.categoryHeadDSA, demoPassword()).then((head) => {
            waitForStatus(head, id, 'ASSIGNED')
              .then((s) => changeStatus(head, id, s, 'IN_PROGRESS'))
              .then((s) => changeStatus(head, id, s, 'RESOLVED'))
          })
        })

        cy.intercept('POST', `/api/reports/${id}/feedback`).as('feedback')
        cy.visit(`/reports/${id}`)
        cy.contains('section', 'What happened').should('contain', 'Resolved')
        cy.contains('section', 'How was the fix?').within(() => {
          cy.get('.fb-pick label').eq(3).click()
          cy.contains('4 stars out of 5').should('be.visible')
          cy.findLabel('Comment (optional)').type('Fixed the same week, thank you.')
          cy.contains('button', 'Send feedback').click()
        })
        cy.wait('@feedback').its('response.statusCode').should('be.oneOf', [200, 201])
        cy.contains('You rated this 4 out of 5 stars.').should('be.visible')
      })
    })
  })
})
