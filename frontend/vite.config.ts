import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/api': {
        target: process.env.VITE_API_PROXY_TARGET || 'http://localhost:6532',
        changeOrigin: true,
        cookieDomainRewrite: '',
        cookiePathRewrite: '',
        configure: (proxy) => {
            proxy.on('proxyReq', (_proxyReq, req) => {
            if (req.url?.includes('/auth/')) {
              console.log('[Proxy] Request:', req.method, req.url)
            }
          })
          proxy.on('proxyRes', (proxyRes, req) => {
            if (req.url?.includes('/auth/')) {
              console.log('[Proxy] Response:', proxyRes.statusCode, req.url)
              const setCookie = proxyRes.headers['set-cookie']
              if (setCookie) {
                console.log('[Proxy] Set-Cookie:', setCookie)
              }
            }
          })
        },
      },
    },
  },
  build: {
    chunkSizeWarningLimit: 700,
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom', 'react-router-dom'],
          'vendor-i18n': ['i18next', 'react-i18next'],
          'vendor-state': ['zustand'],
          'vendor-tanstack-query': ['@tanstack/react-query'],
          'vendor-recharts': ['recharts'],
        },
      },
    },
  },
})
