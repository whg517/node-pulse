/// <reference types="vitest" />

import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    include: ['src/**/*.{test,spec}.{js,ts,jsx,tsx}'],
    setupFiles: ['./src/vitest-setup.ts'],
    typecheck: {
      enabled: false,
    },
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'clover', 'json'],
      exclude: [
        'src/locales/**',
        'src/**/*.types.ts',
        'src/api/types.ts',
        'src/**/index.ts',
        'src/main.tsx',
        'src/vite-env.d.ts',
        'src/test/**',
      ],
      // Ratchet floors: these are intentionally set a few points below the
      // current baseline so normal fluctuations pass, but a real regression is
      // caught. Bump these (never lower) as coverage improves.
      // Column order is lines | branches | functions | statements.
      thresholds: {
        lines: 68,
        branches: 62,
        functions: 56,
        statements: 68,
      },
    },
  },
  resolve: {
    alias: {
      '@': '/src',
    },
  },
})