import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { i18nInitPromise } from './i18n'

// Wait for i18n to initialize before rendering the app
i18nInitPromise
  .then(() => {
    createRoot(document.getElementById('root')!).render(
      <StrictMode>
        <App />
      </StrictMode>,
    )
  })
  .catch((error) => {
    console.error('Failed to initialize i18n:', error)
    // Render fallback UI without i18n
    const root = document.getElementById('root')!
    root.innerHTML = `
      <div style="padding: 2rem; text-align: center; font-family: system-ui, sans-serif;">
        <h1>Initialization Error</h1>
        <p>The application failed to load. Please refresh the page.</p>
      </div>
    `
  })
