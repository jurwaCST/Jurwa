/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** api-gateway address, for example http://localhost:8080 or /api (dev proxy). */
  readonly VITE_API_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
