import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: process.env.VITE_API_PROXY_TARGET || 'http://localhost:6532',
        changeOrigin: true,
        cookieDomainRewrite: '',
        cookiePathRewrite: '',
        configure: (proxy) => {
          proxy.on('proxyReq', (proxyReq, req) => {
            // Log request for debugging
            if (req.url?.includes('/auth/')) {
              console.log('[Proxy] Request:', req.method, req.url)
            }
          })
          proxy.on('proxyRes', (proxyRes, req) => {
            // Log auth responses for debugging
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
          'vendor-echarts': ['echarts'],
          'vendor-state': ['zustand'],
        },
      },
    },
  },
})
