import { defineConfig, devices } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, resolve } from 'path'

// ES module compatible __dirname
const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

// Use the existing test database from pulse/docker-compose.test.yml
const TEST_DB_URL = process.env.TEST_DB_URL || 'postgresql://testuser:testpass123@localhost:5432/nodepulse_test'
const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:6532'
const FRONTEND_BASE_URL = process.env.FRONTEND_BASE_URL || 'http://localhost:5173'

// Set environment variables for tests
process.env.TEST_DB_URL = TEST_DB_URL
process.env.API_BASE_URL = API_BASE_URL
process.env.FRONTEND_BASE_URL = FRONTEND_BASE_URL

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: 2,
  workers: 4,
  timeout: 30000,
  expect: {
    timeout: 10000,
  },
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['list'],
  ],
  use: {
    baseURL: FRONTEND_BASE_URL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 15000,
    navigationTimeout: 30000,
  },
  projects: [
    // Setup project for database seeding and auth state
    {
      name: 'setup',
      testMatch: /global-setup\.ts/,
    },
    // Chromium tests
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
      dependencies: ['setup'],
      testIgnore: /global-setup\.ts/,
    },
  ],
  // Global setup for database seeding
  globalSetup: resolve(__dirname, 'global-setup.ts'),
  // Global teardown for cleanup
  globalTeardown: resolve(__dirname, 'global-teardown.ts'),
})
