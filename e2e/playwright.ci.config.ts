/**
 * CI-Optimized Playwright Configuration
 * 
 * This configuration is optimized for CI environments:
 * - Parallel execution with multiple workers
 * - Sharding support for distributed execution
 * - Only Chromium for faster execution (use full config for cross-browser testing)
 * - Reduced timeouts for faster feedback
 * 
 * Usage:
 *   npx playwright test --config=playwright.ci.config.ts
 *   npx playwright test --config=playwright.ci.config.ts --shard=1/3
 */
import { defineConfig, devices } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, resolve } from 'path'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

const TEST_DB_URL = process.env.TEST_DB_URL || 'postgresql://testuser:testpass123@localhost:5432/nodepulse_test'
const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:6532'
const FRONTEND_BASE_URL = process.env.FRONTEND_BASE_URL || 'http://localhost:5173'

process.env.TEST_DB_URL = TEST_DB_URL
process.env.API_BASE_URL = API_BASE_URL
process.env.FRONTEND_BASE_URL = FRONTEND_BASE_URL

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: true, // Fail if test.only is used
  retries: 1, // Single retry for flakiness
  workers: '50%', // Use 50% of available CPUs
  timeout: 30000, // Reduced timeout for faster feedback
  expect: {
    timeout: 5000,
    toHaveScreenshot: {
      maxDiffPixels: 100,
      threshold: 0.2,
    },
  },
  reporter: [
    ['github'], // GitHub Actions annotations
    ['html', { outputFolder: 'playwright-report', open: 'never' }],
    ['json', { outputFile: 'playwright-report/results.json' }],
    ['junit', { outputFile: 'playwright-report/junit.xml' }],
  ],
  use: {
    baseURL: FRONTEND_BASE_URL,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 10000,
    navigationTimeout: 20000,
  },
  projects: [
    // Only Chromium for CI speed
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },

    // Smoke tests with higher priority
    {
      name: 'smoke',
      use: { ...devices['Desktop Chrome'] },
      testMatch: '**/smoke/**/*.spec.ts',
      retries: 0, // No retries for smoke tests (fast feedback)
    },

    // Visual regression (optional, can be skipped in CI)
    {
      name: 'chromium-visual',
      use: { ...devices['Desktop Chrome'] },
      testMatch: '**/visual/**/*.spec.ts',
      expect: {
        toHaveScreenshot: {
          maxDiffPixels: 100,
          threshold: 0.2,
        },
      },
    },
  ],
  globalSetup: resolve(__dirname, 'global-setup.ts'),
  globalTeardown: resolve(__dirname, 'global-teardown.ts'),

  // Output directories
  outputDir: './test-results',
  snapshotDir: './tests/visual/__screenshots__',

  // Shard support for parallel CI execution
  // Usage: --shard=1/3 --shard=2/3 --shard=3/3
})
