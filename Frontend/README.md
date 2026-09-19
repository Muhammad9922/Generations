# Generations — frontend

The React app for the family tree. It reads and writes everything through the Go
API in `../Backend`, using the typed client in `src/api/` — see
`api-contracts/README.md` for that contract and `../API_SCOPE.md` for the
server's side of it.

```
npm install
npm run dev      # http://localhost:5173, proxying /api to the Go server on :8080
npm test         # the API contract, the view helpers and the env defaults
npm run build    # tsc -b && vite build
```

## Configuration

Every setting is an environment variable. They live in `.env`, which is
committed, and `.env.local` overrides it per machine without touching the
repository. Only names starting with `VITE_` reach the app, and all of them are
substituted into the bundle as literals, so none may hold a secret.

| Variable | Default | What it does |
| --- | --- | --- |
| `VITE_API_BASE_URL` | `/api` | Where the browser sends API calls. `/api` stays on one origin so the dev proxy can forward it; an absolute URL calls the API directly instead. |
| `VITE_API_PROXY_TARGET` | `http://localhost:8080` | Where the dev server forwards `/api`. Read by `vite.config.ts` in Node, so it pairs with the Go server's `API_ADDR`. |
| `VITE_API_CONTRACT_DEBUG` | unset | `[api-contract]` logging in DevTools: `true` always, `false` never, unset means "while developing". |

`src/config/env.ts` is the only module that reads `import.meta.env`, so the
defaults and the trimming rules sit in one place; `src/vite-env.d.ts` declares
the names. Outside Vite — in `npm test` — every value falls back to its default.

## The template this started from

# React + TypeScript + Vite

This template provides a minimal setup to get React working in Vite with HMR and some Oxlint rules.

Currently, two official plugins are available:

- [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react) uses [Oxc](https://oxc.rs)
- [@vitejs/plugin-react-swc](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react-swc) uses [SWC](https://swc.rs/)

## React Compiler

The React Compiler is not enabled on this template because of its impact on dev & build performances. To add it, see [this documentation](https://react.dev/learn/react-compiler/installation).

## Expanding the Oxlint configuration

If you are developing a production application, we recommend enabling type-aware lint rules by installing `oxlint-tsgolint` and editing `.oxlintrc.json`:

```json
{
  "$schema": "./node_modules/oxlint/configuration_schema.json",
  "plugins": ["react", "typescript", "oxc"],
  "options": {
    "typeAware": true
  },
  "rules": {
    "react/rules-of-hooks": "error",
    "react/only-export-components": ["warn", { "allowConstantExport": true }]
  }
}
```

See the [Oxlint rules documentation](https://oxc.rs/docs/guide/usage/linter/rules) for the full list of rules and categories.
