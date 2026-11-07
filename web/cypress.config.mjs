// End-to-end tests (SCRUM-108). They need the web dev server (npm run dev) and
// the api-gateway with its services running. See docs/testing/e2e-tests.md.
//
// Plain JavaScript on purpose: Cypress loads a .ts config through its own tsx
// loader, which hangs on Node 26. The tests stay TypeScript; esbuild compiles
// them (Cypress's default webpack + ts-loader needs the TypeScript JS API,
// which TypeScript 7 no longer has).
import createBundler from '@bahmutov/cypress-esbuild-preprocessor'
import { defineConfig } from 'cypress'

export default defineConfig({
  e2e: {
    baseUrl: 'http://localhost:5173',
    specPattern: 'cypress/e2e/**/*.cy.ts',
    supportFile: 'cypress/support/e2e.ts',
    setupNodeEvents(on) {
      on('file:preprocessor', createBundler())
      // cy.task('log', text) prints in the terminal (used for axe violations).
      on('task', {
        log(text) {
          console.log(text)
          return null
        },
      })
    },
  },
  video: true,
  screenshotOnRunFailure: true,
  env: {
    // cy.request calls the gateway directly; the browser goes through the
    // dev server's /api proxy like a real user.
    apiUrl: 'http://localhost:8080',
    // The demo accounts' password (SEED_DEMO_PASSWORD of identity). Never commit
    // it: pass it with CYPRESS_demoPassword=... npx cypress run (this exact case)
    demoPassword: '',
  },
})
