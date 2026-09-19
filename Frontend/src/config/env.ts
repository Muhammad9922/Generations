/**
 * The one place that reads `import.meta.env`.
 *
 * Vite substitutes each `import.meta.env.VITE_*` reference with a literal while
 * building, so these values configure how the app is wired — they are not a
 * place to keep a secret. The names are declared in `src/vite-env.d.ts` and
 * documented in `.env`.
 *
 * Nothing here touches the DOM, React or the network, so any layer may import
 * it — including the `node:test` suite, where `import.meta.env` does not exist
 * and every value falls back to the default below. That fallback is what keeps
 * the API tests honest about the `/api` base URL.
 */

/** Vite's injected environment. Absent when this module runs outside Vite. */
const environment = import.meta.env as Partial<ImportMetaEnv> | undefined;

/** True under `vite dev` and a dev-mode build; false under `vite build`. */
export const isDevelopment: boolean = environment?.DEV === true;

/**
 * Turns a configured base URL into the prefix paths are appended to, or `""`
 * when there is nothing usable to append to.
 *
 * Trimming the trailing slashes matters: callers build `${base}${path}` from a
 * path that already starts with `/`, so `http://host:8080/` would otherwise ask
 * for `http://host:8080//people`.
 */
export function normalizeApiBaseUrl(value?: string): string {
  return (value ?? "").trim().replace(/\/+$/, "");
}

/**
 * Where every API call is sent.
 *
 * `/api` — the default — keeps the browser on a single origin, so Vite's dev
 * proxy forwards to the Go server and neither side needs CORS. An absolute URL
 * calls the API directly instead, which only works when the API is served from
 * the origin the app was loaded from, or sits behind something that adds CORS
 * headers; the Go server sends none (API_SCOPE.md §2).
 */
export const apiBaseUrl: string = normalizeApiBaseUrl(environment?.VITE_API_BASE_URL) || "/api";

/**
 * Whether every contract call is logged to DevTools.
 *
 * `VITE_API_CONTRACT_DEBUG` wins whenever it is set, so logging can be forced on
 * outside development or silenced inside it. Left unset it follows the build
 * mode: on while developing, off in a production build that would otherwise
 * print personal data to the console.
 */
export const contractDebugEnabled: boolean =
  environment?.VITE_API_CONTRACT_DEBUG === "true" ? true :
  environment?.VITE_API_CONTRACT_DEBUG === "false" ? false :
  isDevelopment;
