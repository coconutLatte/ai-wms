// MSW browser worker setup — enables API mocking in the GitHub Pages demo.
// In development, the Vite proxy forwards /api to the real backend.
// In production (GitHub Pages), MSW intercepts requests with mock data.

import { setupWorker } from 'msw/browser'
import { handlers } from './handlers'

export const worker = setupWorker(...handlers)

// Path where mockServiceWorker.js is served.
// In dev: /mockServiceWorker.js (Vite serves from public/ at root)
// In prod: /ai-wms/mockServiceWorker.js (GitHub Pages base path)
export const workerUrl = import.meta.env.DEV
  ? '/mockServiceWorker.js'
  : '/ai-wms/mockServiceWorker.js'
