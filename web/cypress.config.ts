import { defineConfig } from 'cypress'

// End-to-end tests (SCRUM-108). They need the web dev server (npm run dev) and
// the api-gateway with its services running. See docs/testing/e2e-tests.md.
export default defineConfig({
  e2e: {
    baseUrl: 'http://localhost:5173',
    specPattern: 'cypress/e2e/**/*.cy.ts',
    supportFile: 'cypress/support/e2e.ts',
  },
  video: true,
  screenshotOnRunFailure: true,
  env: {
    // cy.request calls the gateway directly; the browser goes through the
    // dev server's /api proxy like a real user.
    apiUrl: 'http://localhost:8080',
    // The demo accounts' password (SEED_DEMO_PASSWORD of identity). Never commit
    // it: pass it with CYPRESS_DEMO_PASSWORD=... npx cypress run
    demoPassword: '',
  },
})
