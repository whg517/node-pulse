/// <reference types="vite/client" />

interface ImportMetaEnv {
  /**
   * Base URL of the Pulse API. Empty in development (uses the Vite proxy) and
   * in same-origin production builds; set to an absolute URL only when the SPA
   * is hosted on a different origin from the API.
   */
  readonly VITE_API_BASE_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
