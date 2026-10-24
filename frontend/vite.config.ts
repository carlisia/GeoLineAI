import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

// The dev server forwards the API and the health check to the Go server, which
// listens on GEOLINEAI_HTTP_PORT (default 18080).
const backend = `http://127.0.0.1:${process.env.GEOLINEAI_HTTP_PORT ?? '18080'}`

export default defineConfig({
  plugins: [react()],
  server: { proxy: { '/api': backend, '/_': backend } },
  test: { environment: 'jsdom', setupFiles: ['./src/test/setup.ts'] },
})
