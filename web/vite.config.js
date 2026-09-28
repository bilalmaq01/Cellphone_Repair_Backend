import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// During development, proxy API calls to the Go backend on :8080 so the
// browser can use same-origin /api/* paths (no CORS setup needed).
// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
