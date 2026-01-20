import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for Trade Machine E2E tests.
 *
 * The tests run against a standalone HTTP server (not Wails) that serves
 * the same routes and templates as the production app.
 */
export default defineConfig({
  testDir: './tests',

  // Run tests in parallel
  fullyParallel: true,

  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,

  // Retry on CI only
  retries: process.env.CI ? 2 : 0,

  // Use single worker to avoid state conflicts between tests
  // Settings tests share a settings store and need to run serially
  workers: 1,

  // Reporter to use
  reporter: [
    ['html', { open: 'never' }],
    ['list']
  ],

  // Shared settings for all projects
  use: {
    // Base URL for the test server
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:9090',

    // Collect trace when retrying the failed test
    trace: 'on-first-retry',

    // Screenshot on failure
    screenshot: 'only-on-failure',

    // Video on failure
    video: 'on-first-retry',
  },

  // Configure projects for major browsers
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    // Uncomment to test on other browsers:
    // {
    //   name: 'firefox',
    //   use: { ...devices['Desktop Firefox'] },
    // },
    // {
    //   name: 'webkit',
    //   use: { ...devices['Desktop Safari'] },
    // },
  ],

  // Run the Go test server before starting the tests
  webServer: {
    command: 'cd ../.. && go run ./cmd/e2e-server',
    url: 'http://localhost:9090/api/health',
    reuseExistingServer: !process.env.CI,
    timeout: 30 * 1000,
    env: {
      E2E_DATABASE_URL: process.env.E2E_DATABASE_URL || 'postgres://trademachine_test:test_password@localhost:5433/trademachine_test?sslmode=disable',
      E2E_SERVER_PORT: '9090',
    },
  },
});
