import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv } from 'vite'
import tailwindcss from '@tailwindcss/vite'

/**
 * Where the dev server forwards /api when the environment says nothing.
 *
 * This is the other half of the Go server's `API_ADDR`: change one and the
 * other has to move with it, which is why both are settings rather than
 * constants buried in the two files.
 */
const DEFAULT_API_PROXY_TARGET = 'http://localhost:8080'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  // The environment has to be read here in Node: `import.meta.env` is only
  // substituted inside the app bundle. `loadEnv` layers .env and .env.local the
  // same way the app sees them, and the 'VITE_' prefix keeps every setting in
  // this project under one prefix.
  const environment = loadEnv(mode, process.cwd(), 'VITE_')
  const apiTarget = environment.VITE_API_PROXY_TARGET || DEFAULT_API_PROXY_TARGET

  return {
    plugins: [
      tailwindcss(),
      react()
    ],
    // Proxying keeps the browser on a single origin, so the Go API needs no
    // CORS handling (API_SCOPE.md §2). Point VITE_API_BASE_URL at an absolute
    // URL instead, and this proxy is bypassed.
    server: {
      proxy: {
        "/api": {
          target: apiTarget,
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, "")
        }
      }
    }
  }
})
