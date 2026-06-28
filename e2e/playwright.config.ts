import { defineConfig, devices } from '@playwright/test'

const baseURL = process.env.E2E_BASE_URL || process.env.BASE_URL || 'http://localhost:5173'
const startFrontend = process.env.E2E_START_FRONTEND === '1'

export default defineConfig({
  testDir: './tests',
  outputDir: './test-results',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  // CI uses 2 workers for throughput. Locally default to 1: the bundled
  // docker-compose stack runs pulse+postgres+frontend in containers with
  // limited resources, and serial execution avoids flaky timeouts from
  // concurrent page loads overwhelming the local stack.
  workers: process.env.CI ? 2 : 1,
  timeout: 30_000,
  expect: {
    timeout: 10_000,
  },
  reporter: process.env.CI
    ? [['list'], ['html', { open: 'never' }], ['json', { outputFile: 'test-results/results.json' }]]
    : [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL,
    actionTimeout: 10_000,
    navigationTimeout: 15_000,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  webServer: startFrontend
    ? {
        command: 'npm run dev -- --host 127.0.0.1',
        cwd: '../frontend',
        url: baseURL,
        reuseExistingServer: !process.env.CI,
        timeout: 120_000,
      }
    : undefined,
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    {
      name: 'mobile-chromium',
      testMatch: /.*smoke.*\.spec\.ts/,
      use: { ...devices['Pixel 7'] },
    },
  ],
})
