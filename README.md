# Generations

A family tree you can actually edit. A Go HTTP API keeps every person and
marriage in a Bolt graph database, and a React single-page app reads and writes
the whole tree through it.

The data model is deliberately narrow: **every family group is a marriage**. A
child exists only inside the marriage that produced them, so there is no
standalone "parent" link to keep in sync. That one rule is what makes the API
small and the UI predictable — see [The graph model](#the-graph-model).

```
Generations/
├── Backend/                 Go API — one package per domain, one file per operation
│   ├── cmd/server/          main, the routers, the health probe
│   ├── internal/            person, marriage, children, db
│   ├── integration/         operations spanning packages (create a whole family)
│   └── Dockerfile           static binary on alpine, non-root
├── Frontend/                React 19 + Vite + Tailwind SPA
│   ├── src/api/             the only modules that speak HTTP
│   ├── src/pages/           Home, PersonDetails, Relationships, Singles, NewFamily
│   ├── src/helpers/         pure model + relation helpers (no state, no network)
│   ├── tests/               node:test suites for the helpers and the client
│   ├── nginx.conf.template  serves the SPA and proxies /api to the API
│   └── Dockerfile           Vite build, served by nginx
├── Generations/             the same calls as a Bruno collection
├── API_SCOPE.md             the HTTP contract, written from the server's side
├── Frontend/api-contracts/README.md   the same contract, client's side
├── Frontend/docs/manual-qa.md         the manual checklist for the UI
├── docker-compose.yml       builds and runs both halves
└── IDEAS.md                 backlog
```

---

## Table of contents

1. [Quick start](#quick-start)
2. [Prerequisites](#prerequisites)
3. [The graph database](#the-graph-database)
4. [Running in development](#running-in-development)
5. [Configuration reference](#configuration-reference)
6. [The graph model](#the-graph-model)
7. [API reference](#api-reference)
8. [Testing](#testing)
9. [Deploying with Docker](#deploying-with-docker)
10. [Deployment verification](#deployment-verification)
11. [Troubleshooting](#troubleshooting)
12. [Project conventions](#project-conventions)

---

## Quick start

Two terminals, assuming a graph database is already reachable:

```powershell
# Terminal 1 — the API on :8080
cd Backend
go run ./cmd/server

# Terminal 2 — the SPA on :5173
cd Frontend
npm install
npm run dev            # open http://localhost:5173
```

The Vite dev server proxies `/api` to the Go server, so the browser stays on one
origin and the API needs no CORS handling.

Or ship both halves as containers:

```powershell
$env:NEO4J_URI = "bolt://host.docker.internal:7687"
docker compose up --build     # site on http://localhost:8081
```

---

## Prerequisites

| Tool | Version used here | Needed for |
| --- | --- | --- |
| Go | 1.27.1 | the API (`go.mod` requires 1.27.1) |
| Node.js | 24.21.0 | the SPA (image uses `node:24-alpine`) |
| npm | 11.19.0 | installing and building the SPA |
| Docker + Compose | Engine 29.8.0, Compose v5.5.1 | the container path |
| A Bolt graph store | Neo4j or Memgraph | storage for everything |

Nothing else is required. The Go module depends only on
`neo4j-go-driver/v6` (plus `go-spew` for tests); the frontend toolchain is Vite,
TypeScript, Tailwind, Radix Themes and kbar.

---

## The graph database

The API is **the only** component that talks to the database. It expects a
Bolt-speaking graph store with **no authentication configured** — it connects
with `neo4j.BasicAuth("", "", "")`.

> The store in use during development is **Memgraph**. That is not incidental:
> one query is written specifically to avoid a Memgraph limitation (see
> [`GET /people/singles`](#get-peoplesingles)). Neo4j works too.

### Default instance

There is a hard-coded development default:

```go
// Backend/cmd/server/main.go
const defaultDatabaseURI = "bolt://192.168.0.133:7687"
```

Set `NEO4J_URI` to use anything else. **Change this default before deploying
anywhere real** — a hard-coded LAN address is a development convenience, not a
deployment strategy.

### Schema

The API creates nodes and relationships on demand; there is no migration step
and no index is created. A person is:

```
(:Person { id, name, gender, date_of_birth, alive, date_of_death })
```

where the two dates are native graph dates or absent when unknown. A marriage is
`(:Marriage { id, start, end })`, and the edges are described under
[The graph model](#the-graph-model).

Because nothing creates constraints, the fields the API *validates* (name,
gender, well-formed dates) are the only guarantees. Test fixtures written
directly into the graph can violate them — the development database currently
contains people with empty names and empty genders as a result.

---

## Running in development

### 1. The API

```powershell
cd Backend
go run ./cmd/server
```

```
Connected To Server!
2026/09/19 22:52:08 API listening on :8080 (database bolt://192.168.0.133:7687)
```

`Connected To Server!` is printed by `db.ConnectDatabase` via `fmt.Printf`, with
no trailing newline, so the startup log appears on the same line. The API
**verifies connectivity before it serves** and calls `log.Fatalf` if the graph
is unreachable, exiting with status 1 — see
[the API will not start without a graph](#the-api-will-not-start-without-a-graph).

Override the listen address to run a second instance beside one already holding
the port:

```powershell
$env:API_ADDR = ":8099"
go run ./cmd/server
```

### 2. The SPA

```powershell
cd Frontend
npm install
npm run dev
```

```
  VITE v8.2.2  ready in ...

  ➜  Local:   http://localhost:5173/
```

The dev server forwards `/api` to `VITE_API_PROXY_TARGET` (default
`http://localhost:8080`) **and strips the `/api` prefix**, which is why the Go
mux registers bare `/people` and `/marriages` paths rather than `/api/people`.

### 3. Point the two at each other

The pair is `API_ADDR` (Go) and `VITE_API_PROXY_TARGET` (Vite). Change one and
you must move the other:

```powershell
# API on 8099
$env:API_ADDR = ":8099"; go run ./cmd/server

# in Frontend/.env.local
VITE_API_PROXY_TARGET=http://localhost:8099
```

### Frontend scripts

| Script | Command | What it does |
| --- | --- | --- |
| `npm run dev` | `vite` | dev server with HMR on `:5173` |
| `npm run build` | `tsc -b && vite build` | type-check, then emit `dist/` |
| `npm run preview` | `vite preview` | serve the built `dist/` locally |
| `npm test` | `node --test tests/*.test.mjs` | the unit suites (no network) |
| `npm run lint` | `oxlint` | lint `src` with the repo's rule set |

`npm run build` runs `tsc -b` first, so **a type error fails the build** rather
than shipping a broken bundle. The Docker build runs the same script.

---

## Configuration reference

Nothing in this project holds a secret. Every `VITE_*` value is substituted into
the JavaScript bundle as a literal, so it configures the app and protects
nothing.

### Backend

| Variable | Default | What it does |
| --- | --- | --- |
| `NEO4J_URI` | `bolt://192.168.0.133:7687` | The Bolt graph to read and write |
| `API_ADDR` | `:8080` | The address the HTTP server listens on |

Both are read once in `Backend/cmd/server/main.go`; an empty value falls back to
the default.

### Frontend

These live in `Frontend/.env` (committed, holds the defaults) and are overridden
per machine by `Frontend/.env.local` (git-ignored).

| Variable | Default | What it does |
| --- | --- | --- |
| `VITE_API_BASE_URL` | `/api` | Where the browser sends API calls. `/api` keeps one origin so the proxy can forward it; an absolute URL calls the API directly |
| `VITE_API_PROXY_TARGET` | `http://localhost:8080` | Where the dev server forwards `/api`. Read by `vite.config.ts` in Node; pairs with `API_ADDR` |
| `VITE_API_CONTRACT_DEBUG` | unset | `[api-contract]` logging: `true` always, `false` never, unset means "while developing" |

`src/config/env.ts` is the only module that reads `import.meta.env`, so the
defaults and the trailing-slash trimming live in one place. Outside Vite — in
`npm test` — every value falls back to its default, which keeps the API tests
honest about the `/api` base URL.

### Docker Compose

| Variable | Default | What it does |
| --- | --- | --- |
| `NEO4J_URI` | `bolt://192.168.0.133:7687` | Passed to the API container |
| `API_PORT` | `8080` | Host port publishing the API |
| `WEB_PORT` | `8081` | Host port publishing the site |
| `API_UPSTREAM` | `http://api:8080` | Where nginx proxies `/api` (set in the compose file, not the host) |

> The bundled default `bolt://192.168.0.133:7687` is a LAN address. From inside
> a container that address is *not* the host — use
> `bolt://host.docker.internal:7687` for a graph running on your machine.

---

## The graph model

```
(p:Person)-[:MARRIED]->(m:Marriage)<-[:MARRIED]-(p:Person)
(m:Marriage)-[:PRODUCED]->(p:Person)
```

Four consequences follow, and the whole application is built on them:

1. **A marriage holds exactly two spouses**, reached by two `MARRIED` edges.
   The people in a marriage are otherwise immutable: swapping a spouse would
   break every child link hanging off it, so a marriage can only have its
   **dates** changed, or be deleted and recreated.
2. **A child exists only inside the marriage that produced them.** The
   `PRODUCED` edge is the only thing making two spouses somebody's parents;
   there is no standalone parent link.
3. **Deleting a marriage therefore orphans its children.** They remain people,
   but with no parents. This is exactly what the UI's "Delink spouse" copy
   promises, and why it warns how many children will be left behind.
4. **Being single is a graph fact, not a property.** It is the absence of a
   `MARRIED` edge, so it needs its own query — and a widow or widower is *not*
   single, because the marriage that made them a spouse is still recorded.

Relationships beyond the immediate family (siblings, grandparents, uncles,
aunts, cousins, nieces, nephews) are **derived from these payloads, never
stored**. The client walks the graph in three rounds rather than one request per
relative.

### Date handling

Everything crossing the wire is `DD-MM-YYYY`, or `""` when unknown. This is the
`person.DateProper` type.

Inside the packages a date being *changed* is a pointer, because one value has
to express three states:

| Value | Meaning |
| --- | --- |
| `nil` | the field was not sent — keep what is stored |
| `&""` | the field was sent empty — remove the date |
| `&date` | store this date |

The two PATCH endpoints spell "clear" differently on the wire —
`PATCH /people/{id}` uses JSON `null`, `PATCH /marriages/{id}` uses `""` — and
each handler translates its spelling into `&""` before calling the package.

---

## API reference

Full detail lives in [`API_SCOPE.md`](API_SCOPE.md) (server's view) and
[`Frontend/api-contracts/README.md`](Frontend/api-contracts/README.md) (client's
view). This is the working summary.

**Conventions.** JSON in and out. Routes are registered **without** an `/api`
prefix — the proxy strips it. Errors are `{"error": "sentence shown to the
user"}`. Status codes: `200` ok, `201` created, `400` rejected write or bad
body, `404` unknown person or marriage, `405` wrong method, `500` failure.

### Endpoints

| Method | Path | Body | Answers |
| --- | --- | --- | --- |
| `GET` | `/health` | — | `{"status":"ok"}`, never touches the graph |
| `GET` | `/people` | — | `{"people": Person[]}` |
| `GET` | `/people/singles` | — | `{"people": Person[]}` |
| `GET` | `/people/{id}` | — | the full certificate, or `404` |
| `POST` | `/people` | `CreatePerson` | `201` + the saved person |
| `PATCH` | `/people/{id}` | `UpdatePerson` | `200` + the saved person |
| `POST` | `/marriages` | `CreateMarriage` | `201` + `{"id": "…"}` |
| `PATCH` | `/marriages/{id}` | dates only | `200` `true` |
| `DELETE` | `/marriages/{id}` | — | `200` `true` |
| `POST` | `/marriages/{id}/children` | `{"childId": "…"}` | `200` `true` |
| `DELETE` | `/marriages/{id}/children/{childId}` | — | `200` `true` |

### Field casing — the one thing to get right

Only the **person shape** carries json tags. A person is always:

```ts
{ id, name, gender, alive, dateOfBirth, dateOfDeath }
```

in `GET /people`, in `GET /people/singles`, and inside every certificate from
`GET /people/{id}`. The **marriage container** around them and every **request
body** keep the Go field names: `Id`, `Spouse`, `Children`, `StartOfFamily`,
`EndOfFamily`, `PersonName`, `Name`, `DateOfBirth`, `DateOfDeath`, `Alive`,
`Gender`.

Note `PersonName` on create but `Name` on the PATCH — both bodies decode
straight into their Go structs, so the difference is deliberate.

```jsonc
// GET /people/{id}
{
  "Person": { "id": "…", "name": "…", "gender": "…", "alive": true,
              "dateOfBirth": "14-03-1980", "dateOfDeath": "" },
  "ParentsMarriage": {                    // null when they are in none
    "Id": "…",
    "Spouse":   [ /* person */, /* person */ ],
    "Children": [ /* person */ ],
    "StartOfFamily": "10-10-2015",
    "EndOfFamily": null
  },
  "Marriages": [ /* same container shape; [] when a spouse in none */ ]
}
```

### Endpoint notes

- **`GET /people/singles`** — everyone holding no `MARRIED` edge. The query uses
  an `OPTIONAL MATCH` and a count rather than
  `WHERE NOT (p)-[:MARRIED]->(:Marriage)`, because **Memgraph rejects a pattern
  used as an atom expression** ("Not yet implemented") and then retries until
  the driver's budget runs out, turning a syntax problem into a 30-second `500`.
  This is the single most Memgraph-specific thing in the codebase.
- **`GET /people/{id}`** — composed from `person.GetPerson`,
  `children.GetMarriageThatOfChild`, `children.GetMarriageThatOfSpouse`,
  `marriage.GetSpouses` and `children.GetChildren`. The pre-existing
  `integration.GetFamilyCertificate` was **not** reused: it returns the single
  certificate found from a spouse, so it covers neither the parents' marriage
  nor a person with several marriages. `ID` values belonging to a marriage
  rather than a person answer `404`.
- **`POST /marriages`** serves both "add spouse" and "add parents", because both
  are "create a marriage, optionally with children". It delegates to
  `integration.CreateFamily`, which checks existence and reuses existing people
  rather than duplicating them, and **rolls back any people it created** if the
  marriage or a child link fails.
- **`DELETE /marriages/{id}/children/{childId}`** covers both "remove child" and
  "delete parents" (delink this person from their parents' marriage, leaving the
  parents married). Mind the argument order: the route reads the marriage as
  `{id}` and the person as `{childId}`, but `children.DeleteChildren` takes the
  **person first**.

### Rules the server owns

The client re-implements none of these; the message is shown to the user as-is:

- A person needs a name, a valid gender, and dates in `DD-MM-YYYY`; death cannot
  precede birth; nobody alive may carry a death date; an impossible calendar
  date is rejected (`31-02-2001`); ISO dates are rejected outright.
- A marriage needs two **distinct** people of **opposite** gender.
- A marriage starts after both spouses were born and ends on or before either
  death.
- The spouses in a marriage can never be edited — only its dates.

Verified responses:

```
POST /people {"PersonName":""}               400 {"error":"The Person's Name Must Be Given"}
POST /people Gender "Other"                  400 {"error":"A Gender Must Be Supplied For New Person"}
POST /people DateOfBirth "31-02-2001"        400 {"error":"The Date Of Birth Is Not A Day That Exists: 31-02-2001"}
POST /people DateOfBirth "2000-01-01"        400 {"error":"The Date Of Birth Must Be In DD-MM-YYYY Format"}
POST /people alive + DateOfDeath             400 {"error":"The Person Can't Both Be Alive And Have A Death Date!"}
POST /people death before birth              400 {"error":"Time Of Death Must Be After Time Of Birth"}
GET  /people/does-not-exist                  404 {"error":"Person not found"}
DELETE /marriages/does-not-exist             404 {"error":"Marriage not found"}
```

A person cannot marry themselves, but there is no dedicated message for it — the
two-spouse gender rule rejects it as "both spouses are male/female".

### The Bruno collection

`Generations/` holds the same calls as a [Bruno](https://www.usebruno.com/)
collection (`People/` and `Marriages/`) pointing at `http://localhost:8080`.
Useful for exercising the API on its own, without the UI. Note the collection's
requests send **bare** paths (`/people`), matching the server's registration, not
the proxied `/api/people`.

---

## Testing

```powershell
cd Backend  && go build ./... && go vet ./... && go test -count=1 ./...
cd Frontend && npm run lint && npm test && npm run build
```

### What passes today

| Suite | Result | Notes |
| --- | --- | --- |
| `go build ./...` | pass | |
| `go vet ./...` | pass | |
| `go test ./...` | pass | 6 packages: `internal/{person,marriage,children,db}`, `integration`, `cmd/server/person` |
| `npm test` | **33 pass, 0 fail** | client contract, env defaults, family/relation helpers |
| `npm run lint` | **0 warnings, 0 errors** | 38 files, 116 rules |
| `npm run build` | pass | one warning: a chunk > 500 kB |

`cmd/server`, `cmd/server/apiresponse` and `cmd/server/marriage` have no test
files.

### The Go tests need a live graph

These are **not** hermetic. They connect to a real store and **create fixtures
without deleting them afterwards**:

```
go test ./...    # drives a live graph
```

Two cautions before running them anywhere shared:

- The test files connect to a **hard-coded URI** rather than reading
  `NEO4J_URI`. Point the constant in the file you are running at another
  instance first.
- They **leave data behind by design**. The development database is now full of
  fixtures such as `Some Name Here`, `Same Name`, and `String YYYY`, plus people
  with empty names and genders. Run them against a scratch graph, not one you
  care about.

### The frontend tests do not need a server

`npm test` runs the `node:test` suites in `tests/`, which exercise the API
client, the env defaults, and the pure helpers (`personModel`, `relations`,
`PersonActions`, `personDrafts`) against hand-built data. No network, no DOM.

### Manual QA

`Frontend/docs/manual-qa.md` is a 19-section checklist covering the UI flows —
opening people, adding spouses/children/parents, destructive confirmations,
keyboard and screen-reader behaviour, and the shipping checks. Several items
still reference the old `id-1 … id-9` fixture family from
`tests/sampleFamily.mjs`; substitute IDs from your own database.

### A chunk-size warning is expected

`npm run build` prints:

```
(!) Some chunks are larger than 500 kB after minification.
```

The bundle is ~542 kB JS / ~703 kB CSS (165 kB / 87 kB gzipped). This is a
warning, not a failure. Code-splitting the routes would remove it.

---

## Deploying with Docker

### Two images

| Image | Base | Size here | Runs as |
| --- | --- | --- | --- |
| `generations-api` | `alpine:3.21` | ~24.7 MB | `app` (uid 10001, non-root) |
| `generations-web` | `nginx:1.27-alpine` | ~75.2 MB | nginx default |

**Backend** — a multi-stage build compiles with `CGO_ENABLED=0 go build
-trimpath -ldflags="-s -w"`, which makes the binary static so it can run on bare
alpine rather than the build image that carries a libc. Only the 7.8 MB binary is
copied forward. `ca-certificates` is installed for a TLS `bolt://` URI, and the
container runs as an unprivileged user because the API needs no privileges.

**Frontend** — `npm ci` (exact lockfile, fails if it and `package.json`
disagree), then `npm run build` so a type error fails the image, then the static
`dist/` is served by nginx.

### Run it

```powershell
# Graph on this machine
$env:NEO4J_URI = "bolt://host.docker.internal:7687"
docker compose up --build -d
docker compose ps
```

```
NAME                STATUS                    PORTS
generations-api-1   Up (healthy)              0.0.0.0:8080->8080/tcp
generations-web-1   Up (healthy)              0.0.0.0:8081->80/tcp
```

Open <http://localhost:8081>.

### What compose does

- Builds both images from their Dockerfiles.
- Publishes the API on `8080` (`API_PORT`) and the site on `8081` (`WEB_PORT`).
- Gives the API a healthcheck hitting `/health`; `web` waits for
  `condition: service_healthy` before it starts, so the first page load cannot
  race the server's startup.
- **Does not start a database.** Point `NEO4J_URI` at the instance you run.
- `restart: unless-stopped` on both services.

Validate the file without starting anything:

```powershell
docker compose config
```

### What nginx does

`Frontend/nginx.conf.template` is rendered by the nginx entrypoint at container
start, substituting `${API_UPSTREAM}` (default `http://api:8080`) from the
environment — so the upstream can change without rebuilding the image.

| Concern | Behaviour |
| --- | --- |
| `/api/` | Proxied to `API_UPSTREAM`. The **trailing slash on `proxy_pass` strips the `/api` prefix** — the same rewrite Vite's dev proxy does, so the app calls `/api/people` in both places |
| Proxy timeout | `proxy_read_timeout 60s` — a person's certificate is composed from several graph round trips |
| `/assets/` | `Cache-Control: public, immutable` + `expires 1y`; Vite fingerprints these filenames, so a change produces a new URL |
| `/index.html` | `Cache-Control: no-cache`, so a deploy cannot leave open tabs on the old build |
| Everything else | `try_files … /index.html` — client-side routing survives a refresh or a pasted deep link |
| gzip | On, above 1024 bytes, for CSS/JS/JSON/SVG |

`API_UPSTREAM` must **not** end with a slash.

### Verify it by hand

```powershell
docker compose ps                       # both (healthy)
curl.exe http://localhost:8080/health   # {"status":"ok"}
curl.exe http://localhost:8081/api/people
```

### Tear down

```powershell
docker compose down          # stop and remove containers + network
docker compose down --rmi local --volumes   # also drop images
```

---

## Deployment verification

Everything below was executed against this working tree. The graph was the
default Memgraph instance at `bolt://192.168.0.133:7687`.

### Build and tests

| Step | Command | Result |
| --- | --- | --- |
| Backend build | `go build ./...` | exit 0 |
| Backend vet | `go vet ./...` | exit 0 |
| Backend tests | `go test -count=1 ./...` | exit 0, 6 packages ok (**live graph**) |
| Frontend lint | `npm run lint` | 0 warnings, 0 errors |
| Frontend tests | `npm test` | 33 pass, 0 fail |
| Frontend build | `npm run build` | exit 0, 2261 modules, chunk-size warning |
| Compose validity | `docker compose config` | exit 0 |
| Image builds | `docker compose build` | both images built |

### API surface (server run directly on `:8099`)

| Check | Result |
| --- | --- |
| `GET /health` | `200 {"status":"ok"}` |
| `GET /people` | `200`, 1698 people, `application/json` |
| `GET /people/{id}` — list keys | `id, name, gender, alive, dateOfBirth, dateOfDeath` |
| `GET /people/{id}` — container keys | `Person, ParentsMarriage, Marriages` |
| Certificate casing | container `Id/Spouse/Children/StartOfFamily/EndOfFamily`, people lowercase |
| `GET /people/bogus` | `404 {"error":"Person not found"}` |
| `DELETE /people` | `405` |
| Bad JSON body | `400 {"error":"Invalid JSON payload"}` |
| Six validation rules | all `400` with the user-facing sentences quoted above |

A full write cycle was exercised and verified: create father (`201`) → create
mother (`201`) → rename father (`200`) → create marriage (`201`) → create child
→ add child (`200`) → `GET` the child and confirm **both parents** and the
marriage are returned → change marriage dates (`200`) → remove child (`200`, and
the child's `ParentsMarriage` becomes `null`) → delete the marriage (`200`, and
the father's `Marriages` drops to 0). The test people were then removed from the
graph.

### Containerised stack

| Check | Result |
| --- | --- |
| `docker compose up -d` | `api` reached **healthy**, then `web` started |
| `docker compose ps` | both `Up (healthy)` |
| API container user | `uid=10001(app) gid=10001(app)` — non-root, as designed |
| API binary | 7.8 MB static |
| `GET :8080/health` | `200 {"status":"ok"}` |
| `GET :8081/` | `200`, SPA shell, `Cache-Control: no-cache` |
| `GET :8081/api/health` | `200 {"status":"ok"}` — proxy reaches the API |
| `GET :8081/api/people` | `200`, 1701 people — **`/api` prefix stripped correctly** |
| `GET :8081/relationships/some-id` | `200`, serves the SPA shell — deep links survive refresh |
| Person through the proxy | `200`, correct lowercase keys |
| `POST :8081/api/people` | `201` + the saved person — **writes work through nginx** |
| `/assets/index-*.js` | `Cache-Control: max-age=31536000, public, immutable` |
| `/icons.svg` with gzip | `Content-Encoding: gzip`, `Vary: Accept-Encoding` |
| Port overrides | `API_PORT=8299 WEB_PORT=8298` published correctly on those ports |

### A behaviour worth knowing

Pointing the API at an unreachable graph
(`NEO4J_URI=bolt://127.0.0.1:59999`) made the container **crash-loop**:

```
Failed to connect: ConnectivityError: dial tcp 127.0.0.1:59999: connect: connection refused
... RESTARTING
```

This is not a compose bug. `db.ConnectDatabase` calls
`driver.VerifyConnectivity` at startup and `log.Fatalf`s on failure, so a
container that cannot reach its graph exits rather than serving errors;
`restart: unless-stopped` then restarts it until the graph appears or the
retry budget is exhausted. The `/health` endpoint *is* graph-independent by
design, but it is unreachable while the process is down — so **`/health` proves
the HTTP server is up, not that the database is up**. There is no readiness
probe that distinguishes the two.

---

## Troubleshooting

### The API will not start without a graph

```
Failed to connect: ConnectivityError: dial tcp 127.0.0.1:59999: connect: connection refused
exit status 1
```

`ConnectDatabase` verifies connectivity and `log.Fatalf`s. Start the graph, or
fix `NEO4J_URI`. Under compose the container restart-loops until it succeeds.

### `bind: Only one usage of each socket address`

Port `8080` is taken — often by a `go run` or a stale `main.exe` from an earlier
session. Find and stop it:

```powershell
Get-NetTCPConnection -LocalPort 8080 -State Listen |
  ForEach-Object { Get-Process -Id $_.OwningProcess }
Stop-Process -Id <pid> -Force
```

Or move one side: `API_ADDR=:8099` for Go, `API_PORT=8099` for compose.

### A container is up but its `/health` returns 404

You are almost certainly talking to something else that already held the port —
not the API. This project **does** answer `/health`. Probe `/people` to tell the
two apart: an old or foreign server returns `404` on `/health` while `/people`
may still answer.

### `docker compose up` fails on a port that is not yours

`ports are not available: exposing port TCP 0.0.0.0:8080: listen tcp … bind:
Only one usage of each socket address`. See above. Note the check is against
**all** interfaces.

### The graph is unreachable from inside a container

`bolt://192.168.0.133:7687` is a LAN address, not the host. From a container that
is a different machine:

```powershell
$env:NEO4J_URI = "bolt://host.docker.internal:7687"
```

### The page loads but every person is blank

The certificate keys are the usual culprit. `Person` and every entry of
`Spouse`/`Children` must use the **lowercase** keys. If a handler starts sending
Go field names (`Id`, `Name`, `DateOfBirth`), the fields read `undefined` and the
page renders blank without erroring. `IDEAS.md` describes exactly this bug
hitting the details page once already. Compare what the API returns:

```powershell
curl.exe http://localhost:8080/people/<id>
```

### `The API did not answer with JSON`

A handler reached `http.Error`, whose body is `text/plain`. Use the
`apiresponse` helpers so every answer is JSON.

### `spawn EPERM` / `Access is denied` in a sandbox

Symptom of running `npm test`, `npm run build` or `docker compose build` under a
restricted sandbox rather than a problem with the project. Node's test runner
and Vite spawn subprocesses with piped stdio, and Docker writes to
`~/.docker/buildx/.lock`; both are outside the workspace. Re-run with full
access.

### Tests "pass" without running

`go test ./...` prints `(cached)` and reuses a previous result. Use
`go test -count=1 ./...` when you want certainty that it actually ran.

### `npm ci` fails in the Docker build

`.dockerignore` deliberately excludes `node_modules`, so the image installs its
own. If `npm ci` fails, `package.json` and `package-lock.json` have drifted —
run `npm install` locally and commit the updated lockfile.

---

## Project conventions

### Backend

- **One package per domain, one file per operation.** `cmd/server/person` holds
  `create.go`, `update.go`, `query.go`, `singles.go`; each router package exposes
  a single `Register…Routes(mux, driver)`.
- **Never `http.Error`.** Every response goes through `cmd/server/apiresponse`,
  which owns the two JSON envelopes `Write(w, status, body)` and
  `Error(w, status, message)`.
- **Routes are registered without `/api`.** The proxy strips that prefix; adding
  it would break both the dev server and nginx.
- **Validation lives in the packages**, not the handlers, so the rules hold no
  matter who calls them.
- **Multi-step writes roll back.** `integration.CreateFamily` deletes any people
  it created if a later step fails.
- **Timeouts are always explicit.** `main.go` sets `ReadHeaderTimeout`,
  `ReadTimeout`, `WriteTimeout` and `IdleTimeout`, because
  `http.ListenAndServe`'s defaults are unlimited and leave a stalled client
  holding a connection indefinitely.

### Frontend

- **`src/api/http.ts` is the only module that calls `fetch`.** Base URL, JSON
  encoding, status codes turned into `ApiError`, cancellation passed through.
- **`src/config/env.ts` is the only module that reads `import.meta.env`.**
- **Endpoint paths live only in `src/api/contracts.ts`** (`ENDPOINTS`), so
  moving a route is a one-line change.
- **The client never patches its own copy.** Every write ends by re-reading, so
  what is on screen is what the API holds.
- **Helpers are pure.** `src/helpers/*` hold no state and make no requests,
  which is what makes them testable under `node:test`.

### General

- Comments explain **why**, not what. The codebase is heavily commented in this
  style; match it.
- `.gitattributes` normalises line endings to LF for every text file.
- `.gitignore` excludes build output (`dist/`, `node_modules/`) and per-machine
  env files (`*.local`), but `Frontend/.env` **is** committed because it holds
  the shared defaults.

---

## Related documents

| Document | What it covers |
| --- | --- |
| [`API_SCOPE.md`](API_SCOPE.md) | The HTTP contract in full, from the server's side, with the schema sketches and the reasoning behind each endpoint |
| [`Frontend/api-contracts/README.md`](Frontend/api-contracts/README.md) | The same contract from the client's side, plus the marriage model and how relations are derived |
| [`Backend/README.md`](Backend/README.md) | Backend layout, graph model and the three-state date table |
| [`Frontend/README.md`](Frontend/README.md) | Frontend configuration and scripts |
| [`Frontend/docs/manual-qa.md`](Frontend/docs/manual-qa.md) | The 19-section manual UI checklist |
| [`IDEAS.md`](IDEAS.md) | Backlog: profile pictures, relationship finder, name search |
