# Backend

The Go API behind the Generations family tree. It speaks JSON over HTTP, owns
every rule about people and marriages, and is the only piece that talks to
Neo4j — the React app reaches it through `Frontend/src/api/*`.

`API_SCOPE.md` in the repository root is the written contract between the two.

## Running it

```
go run ./cmd/server
```

It listens on `:8080`, which is where the Vite dev proxy sends `/api`
(`Frontend/vite.config.ts`). Two environment variables override the defaults:

| Variable | Default | What it is |
| --- | --- | --- |
| `NEO4J_URI` | `bolt://192.168.0.133:7687` | The database to use. The default instance takes no credentials (`neo4j.BasicAuth("", "", "")`). |
| `API_ADDR` | `:8080` | The listen address, so a second instance can run beside one already holding the port. |

## Tests

```
go test ./...
```

The tests drive a **live Neo4j**: they connect straight to
`bolt://192.168.0.133:7687` rather than reading `NEO4J_URI`, and they create
their fixtures without deleting them afterwards. Point the URI constant in the
test file you are running at another instance before running the suite anywhere
shared.

## Layout

| Path | What lives there |
| --- | --- |
| `cmd/server` | `main`: opens the driver, registers both routers, listens. |
| `cmd/server/person` | The people endpoints, one file per operation. |
| `cmd/server/marriage` | The marriage endpoints, one file per operation. |
| `cmd/server/apiresponse` | The two JSON envelopes every handler answers with. |
| `internal/person` | Person creation, validation, lookup and update. |
| `internal/marriage` | Marriages and the dates they ran between. |
| `internal/children` | The `PRODUCED` link between a marriage and a child. |
| `internal/db` | Opening the Neo4j driver. |
| `integration` | Operations that span the packages, such as creating a whole family with rollback. |

## Graph model

```
(p:Person)-[:MARRIED]->(m:Marriage)<-[:MARRIED]-(p:Person)
(m:Marriage)-[:PRODUCED]->(p:Person)
```

- A person is `{id, name, gender, date_of_birth, alive, date_of_death}`, where
  the dates are native Neo4j dates or absent when unknown.
- A marriage is `{id, start, end}` and holds exactly two spouses.
- A child exists only inside the marriage that produced them: the `PRODUCED`
  link is the only thing that makes two spouses somebody's parents. There is no
  standalone parent link, which is why deleting a marriage leaves its children
  without parents.
- The people in a marriage are fixed once it exists, because swapping a spouse
  would break every child link hanging off it.

## Date handling

Everything crossing the wire is `person.DateProper`: `DD-MM-YYYY`, or `""` when
the date is unknown.

Inside the packages a date is always a pointer when it is being changed, because
one value has to express three states:

| Value | Meaning |
| --- | --- |
| `nil` | the field was not sent — keep what is stored |
| `&""` | the field was sent empty — remove the date |
| `&date` | store this date |

`PATCH /people/{id}` spells "clear" as JSON `null` (the frontend's wire shape) and
`PATCH /marriages/{id}` spells it as `""` (the dates dialog's shape). Both
handlers turn their spelling into `&""` before calling the package.
