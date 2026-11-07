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

  it('a student uploads a photo, attaches it to a report and can view it', () => {
    cy.apiLogin(studentEmail, studentPassword).then((token) => {
      const auth = { Authorization: `Bearer ${token}` }
      cy.fixture('photo.png', null).then((png: Uint8Array) => {
        cy.uploadPhoto(token, png, 'image/png', 'photo.png').then((up) => {
          expect(up.status).to.eq(201)
          const photoId = up.body.photoId as string
          expect(photoId).to.match(/^[0-9a-f-]{36}$/)

          cy.request({
            method: 'POST',
            url: `${api()}/reports`,
            headers: auth,
            body: {
              category: 'HOSTEL',
              location: { building: 'Hostel A', wing: 'Wing 1', area: 'Bathroom 1' },
              description: 'E2E test: broken window, photo attached',
              photoId,
            },
          }).then((res) => {
            expect(res.status).to.eq(201)
            expect(res.body.report.photoId).to.eq(photoId)
          })

          cy.request({ url: `${api()}/photos/${photoId}`, headers: auth }).then((got) => {
            expect(got.status).to.eq(200)
            expect(got.headers['content-type']).to.eq('image/png')
            expect(got.headers['x-content-type-options']).to.eq('nosniff')
          })
        })
      })
    })
  })

  it('a photo over 5 MB is refused with the size message', () => {
    // 5.5 MB: over the 5 MB rule, under the gateway's 6 MB body limit, so
    // photo-media itself answers. A PNG signature first, zeros after.
    const big = new Uint8Array(5.5 * 1024 * 1024)
    big.set([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a])
    cy.apiLogin(studentEmail, studentPassword).then((token) => {
      cy.uploadPhoto(token, big, 'image/png', 'big.png').then((up) => {
        expect(up.status).to.eq(400)
        expect(up.body.error).to.contain('larger than 5 MB')
      })
    })
  })

  it('another student cannot view the photo', () => {
    cy.apiLogin(studentEmail, studentPassword).then((token) => {
      cy.fixture('photo.png', null).then((png: Uint8Array) => {
        cy.uploadPhoto(token, png, 'image/png', 'photo.png').then((up) => {
          cy.registerStudent().then((other) => {
            cy.apiLogin(other, studentPassword).then((otherToken) => {
              cy.request({
                url: `${api()}/photos/${up.body.photoId}`,
                headers: { Authorization: `Bearer ${otherToken}` },
                failOnStatusCode: false,
              }).its('status').should('eq', 403)
            })
          })
        })
      })
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
