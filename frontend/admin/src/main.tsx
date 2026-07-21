// Application entry point.
// In GitHub Pages (production without backend), MSW intercepts API calls with mock data.
// In dev mode (localhost), Vite proxy forwards /api to the real backend.

import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './styles/global.css'

async function bootstrap() {
  // Start MSW in production (GitHub Pages demo — no backend available).
  // In dev, the Vite proxy handles /api → localhost:8080.
  if (import.meta.env.PROD) {
    const { worker, workerUrl } = await import('./mocks/browser')
    await worker.start({ onUnhandledRequest: 'bypass', serviceWorker: { url: workerUrl } })
  }

  ReactDOM.createRoot(document.getElementById('root')!).render(
    <React.StrictMode>
      <App />
    </React.StrictMode>,
  )
}

bootstrap()
