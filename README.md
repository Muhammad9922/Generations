# Generations

A family tree: a Go API over a Bolt graph database, and a React single-page app
that reads and writes every relationship through it.

```
Backend/              Go API — one package per domain, one file per operation
Frontend/             React + Vite SPA
API_SCOPE.md          the HTTP contract, written from the server's side
docker-compose.yml    builds and runs both halves
```

## Running it in development

The API needs a Bolt-compatible graph store — Neo4j or Memgraph. It defaults to
the development instance in `Backend/cmd/server/main.go`; point `NEO4J_URI`
somewhere else to use yours.

```powershell
cd Backend
go run ./cmd/server          # listens on :8080, or API_ADDR

cd ../Frontend
npm install
npm run dev                  # http://localhost:5173, proxying /api to :8080
```

The dev server forwards `/api` to `VITE_API_PROXY_TARGET`, so the browser stays
on one origin and the API needs no CORS handling. Set that variable to point the
app at a different API without touching source.

## Shipping with Docker

```powershell
docker compose up --build            # web on http://localhost:8081
```

The compose file builds two images:

- **Backend** — a static Go binary on plain alpine, running as a non-root user.
  It answers `GET /health` for the container healthcheck.
- **Frontend** — the built SPA served by nginx, which proxies `/api` to the API
  container and falls back to `index.html` for client-side routes.

The graph database is not started by compose. Point it at the instance you run:

```powershell
$env:NEO4J_URI = "bolt://host.docker.internal:7687"   # a graph on the host
docker compose up --build
```

Ports are overridable: `API_PORT` (default `8080`) publishes the API for direct
calls, and `WEB_PORT` (default `8081`) publishes the site. Change `API_PORT` if
the Vite dev server is already holding `8080`.

## Configuration

Nothing here holds a secret: every value starting with `VITE_` is substituted
into the bundle as a literal, so it configures the app rather than protecting it.

| Side | Variable | Default | What it does |
| --- | --- | --- | --- |
| Backend | `NEO4J_URI` | the development instance | The Bolt graph to read and write |
| Backend | `API_ADDR` | `:8080` | Where the API listens |
| Frontend | `VITE_API_BASE_URL` | `/api` | Where the browser sends API calls |
| Frontend | `VITE_API_PROXY_TARGET` | `http://localhost:8080` | Where the dev server forwards `/api` |
| Frontend | `VITE_API_CONTRACT_DEBUG` | on while developing | `[api-contract]` logging in DevTools |
| Frontend | `API_UPSTREAM` | `http://api:8080` | Where the shipped nginx proxies `/api` |

## Tests

```powershell
cd Backend  && go test ./...        # needs the graph reachable
cd Frontend && npm test             # pure helpers, no network
cd Frontend && npm run build        # tsc -b && vite build
```
