import { defineConfig, devices } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, resolve } from 'path'

// ES module compatible __dirname
const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

// Use the existing test database from pulse/docker-compose.test.yml
const TEST_DB_URL = process.env.TEST_DB_URL || 'postgresql://testuser:testpass123@localhost:5432/nodepulse_test'
const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:6532'
const FRONTEND_BASE_URL = process.env.FRONTEND_BASE_URL || 'http://127.0.0.1:5173'

// Set environment variables for tests
process.env.TEST_DB_URL = TEST_DB_URL
process.env.API_BASE_URL = API_BASE_URL
process.env.FRONTEND_BASE_URL = FRONTEND_BASE_URL

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: 2,
  workers: 1, // Single worker to avoid backend overload and auth timeouts
  timeout: 45000, // Increased from 30000 for slower environments
  expect: {
    timeout: 10000,
    toHaveScreenshot: {
      maxDiffPixels: 100,
      threshold: 0.2,
    },
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
    // Chromium tests - globalSetup handles seeding and auth state
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },

    // Firefox tests - cross-browser compatibility
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },

    // WebKit (Safari) tests - cross-browser compatibility
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },

    // Visual regression tests (Chromium only for consistency)
    {
      name: 'chromium-visual',
      use: { ...devices['Desktop Chrome'] },
      testMatch: '**/tests/visual/**/*.spec.ts',
      expect: {
        toHaveScreenshot: {
          maxDiffPixels: 100,
          threshold: 0.2,
        },
      },
    },
  ],
  // Global setup for database seeding
  globalSetup: resolve(__dirname, 'global-setup.ts'),
  // Global teardown for cleanup
  globalTeardown: resolve(__dirname, 'global-teardown.ts'),
})
