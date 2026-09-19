/// <reference types="vite/client" />

/**
 * The environment variables this app reads.
 *
 * Declaring them here gives each name a real type and one place to look for the
 * full list; the values themselves are documented in `.env`. Vite replaces
 * every `import.meta.env.VITE_*` reference with a literal while building, so
 * everything listed below ends up public in the bundle and none of it may hold
 * a secret.
 *
 * Vite's own `ImportMetaEnv` already declares `DEV`, `PROD`, `MODE` and
 * `BASE_URL`; this file adds the names this project introduces.
 */
interface ImportMetaEnv {
  /**
   * Where the browser sends API calls. `/api` (the default) is proxied to the
   * Go server by Vite; an absolute URL calls the API directly instead.
   */
  readonly VITE_API_BASE_URL?: string;

  /**
   * `[api-contract]` logging in DevTools: `"true"` to always log, `"false"` to
   * never log, unset to log while developing only.
   */
  readonly VITE_API_CONTRACT_DEBUG?: string;
}
