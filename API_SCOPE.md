# API_SCOPE.md — the HTTP API the frontend needs

**Audience:** whoever writes the Go HTTP layer (the handlers in `Backend/cmd/server`).
**Status:** the frontend is already wired to every endpoint below. Nothing else in the frontend blocks on it.

## 1. Where we stand

- `Backend/cmd/server/main.go` is a stub: it registers `/` and answers `"Hi there!"` on every path.
- All the domain work already exists as packages: `person`, `marriage`, `children`, `integration`, `db`.
- The frontend (`Frontend/src/api/*`) calls the endpoints below through one `fetch` wrapper. Until those routes exist it shows **"Could not reach the API."** and nothing else.
- **No frontend change is needed** once the routes match this document. Endpoint paths live in one file: `Frontend/src/api/contracts.ts` (`ENDPOINTS`), so a rename is a one-line change there.

## 2. Conventions

| Topic | Rule |
| --- | --- |
| Host | Go server on `:8080` (as `main.go` already listens). |
| Base URL | The frontend calls `/api/*`. Vite proxies it to `http://localhost:8080` **and strips the `/api` prefix** (`Frontend/vite.config.ts`). So `/api/people` arrives at your mux as **`/people`** — register the routes *without* `/api`. |
| CORS | Not needed in development because of that proxy. Only add it if the frontend is served from another origin (`VITE_API_BASE_URL` set to an absolute URL). |
| Content type | `application/json` for every request with a body and every response. |
| Field names | The Go structs have **no json tags**, so the field names are the Go names: `Id`, `Name`, `PersonName`, `DateOfBirth`, `DeateOfDeath`, `Gender`, `Alive`, `Spouse`, `Chidren`, `StartOfFamily`, `EndOfFamily`. Keep the two Go typos — the frontend expects them. |
| Dates | `person.DateProper`: `DD-MM-YYYY`, or `""` when unknown. `*DateProper` marshals to `null` when nil; the frontend accepts both. |
| Nil slices | `*[]User` marshals to `null` when nil. The frontend treats a null `Spouse`/`Chidren` as empty, so either `null` or `[]` is fine. |
| `PersonName` vs `Name` | `person.NewPerson` calls it `PersonName`; `integration.User` and `person.UpdateUser` call it `Name`. Create takes `PersonName`, the PATCH body takes `Name`. That is deliberate: both bodies decode straight into the Go structs. |
| Errors | JSON object with a message: `{"error": "both spouses are male"}`. `message`, `Error` and `Message` are also read. **The string is displayed to the user as-is**, so return user-facing sentences (the existing `errors.New(...)` texts are already suitable). |
| Status codes | `200` ok · `201` created · `400` rejected write / bad body · `404` unknown person or marriage · `405` wrong method · `500` failure. A non-JSON body is reported to the user as an error, so always answer JSON. |

## 3. Server wiring

One place to hold the driver, one helper for JSON, then a route per endpoint. Go 1.22+ method patterns keep this readable:

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/Muhammad9922/Generations/internal/db"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type api struct {
	ctx    context.Context
	driver neo4j.Driver
}

var errNotFound = errors.New("not found")

func (a *api) write(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func (a *api) fail(w http.ResponseWriter, status int, err error) {
	a.write(w, status, map[string]string{"error": err.Error()})
}

func main() {
	ctx, driver := db.ConnectDatabase(os.Getenv("NEO4J_URI")) // same call the tests use
	defer driver.Close(ctx)
	a := &api{ctx: ctx, driver: driver}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /people", a.listPeople)
	mux.HandleFunc("POST /people", a.createPerson)
	mux.HandleFunc("GET /people/{id}", a.personDetails)
	mux.HandleFunc("PATCH /people/{id}", a.updatePerson)
	mux.HandleFunc("POST /marriages", a.createMarriage)
	mux.HandleFunc("PATCH /marriages/{id}", a.updateMarriage)
	mux.HandleFunc("DELETE /marriages/{id}", a.deleteMarriage)
	mux.HandleFunc("POST /marriages/{id}/children", a.addChild)
	mux.HandleFunc("DELETE /marriages/{id}/children/{childId}", a.removeChild)

	log.Println("API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

`db.ConnectDatabase` uses `neo4j.BasicAuth("", "", "")` and `log.Fatalf`s on failure, so pass the URI from an environment variable of your choice (`NEO4J_URI`) and keep it out of the source.

## 4. Endpoints

Each section says what the UI does, the exact contract, and **where to connect it** in the existing packages.

---

### 4.1 `GET /people` — list for the palette and every selector

**Needs a new query.** No package function lists people today.

- **Frontend call site:** `src/api/people.ts` → `getAllPeople()`, used by `App.tsx` for the command palette (`Ctrl+K`) and by every dialog's person selector.
- **Request:** none.
- **Response `200`:**

```json
{
  "people": [
    { "id": "8f1c…", "name": "Mahamaud Muhayodin", "gender": "Male", "alive": true,
      "dateOfBirth": "14-03-1980", "dateOfDeath": "" }
  ]
}
```

Note the **lowercase keys** on this one shape (it is the frontend's lean list model, not `integration.User`).

- **`personFatherName` is optional and you should omit it.** The `Person` node has no such property (see the `MERGE` in `internal/person/create.go`: `name, gender, date_of_birth, id, alive, date_of_death`). The palette shows a "Father: …" subtitle only when it is present.
- **Connect it:** a small query in `internal/person/query.go`:

```cypher
MATCH (p:Person)
RETURN p.id AS id, p.name AS name, p.gender AS gender, p.alive AS alive,
       p.date_of_birth AS dob, p.date_of_death AS dod
ORDER BY p.name
```

Map dates through the existing `person.GetProperDate(...)` so Neo4j dates become `DD-MM-YYYY`, and return `""` for a missing date rather than `null` (the frontend's `Person` uses optional strings).

---

### 4.2 `GET /people/{id}` — everything the details page renders

- **Frontend call site:** `src/api/family.ts` → `getPersonDetails()`, called by `FamilyView.tsx` on every open and after **every** write. Also used when a card is opened as the primary person.
- **Request:** none.
- **Response `200`:**

```json
{
  "Person": { "Id": "…", "Name": "…", "DateOfBirth": "14-03-1980", "DeateOfDeath": "", "Gender": "Male", "Alive": true },
  "ParentsMarriage": {
    "Id": "…", "Spouse": [ { "…": "User" }, { "…": "User" } ],
    "Chidren": [ { "…": "User" } ], "StartOfFamily": null, "EndOfFamily": null
  },
  "Marriages": [ { "Id": "…", "Spouse": [ … ], "Chidren": [ … ], "StartOfFamily": "12-06-2005", "EndOfFamily": null } ]
}
```

`ParentsMarriage` is `null` when the person is in no marriage as a child, and `Marriages` is `[]` (never `null`, though the frontend tolerates either) when they are a spouse in none.

- **Errors:** `404` if the ID belongs to nobody — or to a marriage rather than a person. The page then shows "Person not found". This is a normal answer, not a failure.
- **Connect it:** compose from the packages that already exist. `integration.GetFamilyCertificate` is close but **not sufficient**: it returns the single certificate found from a spouse, so it covers neither the parents' marriage nor a person with several marriages.

```go
func (a *api) personDetails(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	exists, err := person.CheckPersonExistence(a.ctx, a.driver, person.PersonQuery{ID: id})
	if err != nil { a.fail(w, 500, err); return }
	if !exists { a.fail(w, 404, errNotFound); return }

	record, err := person.GetPerson(a.ctx, a.driver, id)
	if err != nil { a.fail(w, 500, err); return }

	out := PersonDetails{Person: toUser(id, record), Marriages: []FamilyCertificate{}}

	if parents, err := children.GetMarriageThatOfChild(a.ctx, a.driver, id); err != nil {
		a.fail(w, 500, err); return
	} else if parents != nil {
		cert, err := a.certificate(parents.Id)
		if err != nil { a.fail(w, 500, err); return }
		out.ParentsMarriage = cert
	}

	spouseMarriages, err := children.GetMarriageThatOfSpouse(a.ctx, a.driver, id)
	if err != nil { a.fail(w, 500, err); return }
	if spouseMarriages != nil {
		for _, m := range *spouseMarriages {
			cert, err := a.certificate(m.Id)
			if err != nil { a.fail(w, 500, err); return }
			out.Marriages = append(out.Marriages, *cert)
		}
	}

	a.write(w, 200, out)
}

// certificate builds one FamilyCertificate; get_family.go already does this for
// a single marriage, so most of this is a refactor of that code.
func (a *api) certificate(marriageId string) (*FamilyCertificate, error) {
	oneId, twoId, _, err := marriage.GetSpouses(a.ctx, a.driver, marriage.SpouseQueryParams{MarriageId: marriageId})
	if err != nil { return nil, err }
	rows, err := marriage.GetMarriageFromMarriageId(a.ctx, a.driver, marriageId)
	if err != nil || len(rows) == 0 { return nil, err }

	one, err := person.GetPerson(a.ctx, a.driver, oneId)
	if err != nil { return nil, err }
	two, err := person.GetPerson(a.ctx, a.driver, twoId)
	if err != nil { return nil, err }
	kids, err := children.GetChildren(a.ctx, a.driver, marriageId)
	if err != nil { return nil, err }

	spouses := []User{toUser(oneId, one), toUser(twoId, two)}
	chidren := []User{}
	for _, kid := range kids { chidren = append(chidren, toUser(kid.Id, &kid)) }

	start, end := rows[0].Start, rows[0].End
	return &FamilyCertificate{Id: marriageId, Spouse: &spouses, Chidren: &chidren, StartOfFamily: &start, EndOfFamily: &end}, nil
}

func toUser(id string, p *person.NewPerson) User {
	return User{Id: id, Name: p.PersonName, DateOfBirth: p.DateOfBirth, DeateOfDeath: p.DateOfDeath, Gender: p.Gender, Alive: p.Alive}
}
```

`integration/get_family.go` already contains this `NewPerson` → `User` mapping, so copying its field assignment keeps things consistent. `*[]User` and `*DateProper` give you `null` for a marriage with no children or no end date, which the frontend handles.

---

### 4.3 `POST /people` — create a person

- **UI actions:** "Create new person" inside Add spouse / Add child / Add parents.
- **Frontend call site:** `src/api/people.ts` → `createPerson()`, reached through `src/helpers/personDrafts.ts` (`createPersonFromDraft`).
- **Request:**

```json
{ "PersonName": "Ada Lovelace", "Gender": "Female", "DateOfBirth": "10-12-1815", "DateOfDeath": "", "Alive": true }
```

- **Response `201`:** the **saved person**, so the client never invents an ID:

```json
{ "Id": "a1b2…", "Name": "Ada Lovelace", "DateOfBirth": "10-12-1815", "DeateOfDeath": "", "Gender": "Female", "Alive": true }
```

- **Errors:** `400` with a message for each rule `CreateNewPerson` already enforces — missing name, invalid gender, malformed or impossible dates, `Alive: true` together with a death date.
- **Connect it:**

```go
name, id, err := person.CreateNewPerson(a.ctx, a.driver, person.NewPerson{
	PersonName:  body.PersonName,
	Gender:      person.Gender(body.Gender),
	DateOfBirth: person.DateProper(body.DateOfBirth),
	DateOfDeath: person.DateProper(body.DateOfDeath),
	Alive:       body.Alive,
})
if err != nil { a.fail(w, 400, err); return }
a.write(w, 201, User{Id: id, Name: name, DateOfBirth: body.DateOfBirth, DeateOfDeath: body.DateOfDeath, Gender: body.Gender, Alive: body.Alive})
```

`params.Id` is left empty so the package generates a UUID.

---

### 4.4 `PATCH /people/{id}` — edit any person on the page

Covers the selected person, a parent, a spouse and a child.

- **Frontend call site:** `src/api/people.ts` → `updateUserDetails()` (the "Edit person" dialog in `FamilyView.tsx`).
- **Request** (only the editable profile; omitted fields keep their value, `null` clears):

```json
{ "Name": "Renamed", "Gender": "Male", "Alive": true, "DateOfBirth": "01-01-1990", "DateOfDeath": null }
```

- **Response `200`:** the saved `User` (the frontend reloads the page afterwards anyway, so echoing the record is enough).
- **Errors:** `404` unknown person, `400` on a rejected update.
- **Connect it:** decode straight into `person.UpdateUser` — its pointer fields already mean "absent = unchanged":

```go
var body person.UpdateUser
if err := json.NewDecoder(r.Body).Decode(&body); err != nil { a.fail(w, 400, err); return }

id := r.PathValue("id")
if _, _, err := person.UpdatePerson(a.ctx, a.driver, id, body); err != nil { a.fail(w, 400, err); return }

saved, err := person.GetPerson(a.ctx, a.driver, id)
if err != nil { a.fail(w, 500, err); return }
a.write(w, 200, toUser(id, saved))
```

---

### 4.5 `POST /marriages` — add a spouse, or add parents

One endpoint serves both, because both are "create a marriage, optionally with children".

- **UI actions:** **Add spouse** (`SpouseCreatorDialog`), **Add parents** (`ParentCreatorDialog` → `FamilyView`), and "Add child" when the other parent is not yet a spouse (`ChildCreatorDialog`).
- **Frontend call site:** `src/api/marriages.ts` → `createMarriage()`.
- **Request:**

```json
{ "SpouseOne": "id-a", "SpouseTwo": "id-b", "DateStart": "01-01-2005", "DateEnd": "", "childrenIds": ["id-child"] }
```

`DateStart`/`DateEnd` are `""` when the UI left them blank; `childrenIds` is optional and absent when empty.

- **Response `201`:** `{"id": "marriage-uuid"}`
- **Errors:** `400` for every rule `CreateNewMarriage` enforces (two spouses of the same gender, a marriage before either birth, an end after either death, malformed dates) **and** for the integration-level failures (unknown spouse, unknown child). Those `errors.New` texts already read well in the UI.
- **Connect it:** `integration.CreateFamily` does exactly this in one call, including rollback. Put the incoming **existing** IDs into `person.NewPerson{Id: …}` — the function checks existence first and reuses them instead of creating anyone:

```go
detail, err := integration.CreateFamily(a.ctx, a.driver,
	person.NewPerson{Id: body.SpouseOne},
	person.NewPerson{Id: body.SpouseTwo},
	childrenAsNewPeople(body.ChildrenIds),
	integration.MarriageOptionalParams{
		DateStart: body.DateStart,
		DateEnd:   body.DateEnd,
		Id:        "", // let the package generate one
	},
)
if err != nil { a.fail(w, 400, err); return }
a.write(w, 201, map[string]string{"id": detail.MarriageID})
```

`integration.CreateFamily` returns `*FamilyDetail{MarriageID, SpouseOneId, SpouseTwoId, ChildrenIds}`. Its test suite already covers new spouses, existing spouses, mixed, children, and attaching to an existing marriage.

---

### 4.6 `PATCH /marriages/{id}` — change only the dates

- **UI action:** **Actions → Change dates…** beside a spouse.
- **Frontend call site:** `src/api/marriages.ts` → `updateMarriage()`.
- **Request:** `{ "DateStart": "01-01-2010", "DateEnd": "" }` — an empty string clears that date. **Never** sends spouses or children.
- **Response `200`:** `true`
- **Errors:** `404` unknown marriage, `400` invalid or reversed dates.
- **Connect it:**

```go
if _, err := marriage.UpdateMarriageDates(a.ctx, a.driver, r.PathValue("id"),
	marriage.MarriageUpdate{DateStart: person.DateProper(body.DateStart), DateEnd: person.DateProper(body.DateEnd)}); err != nil {
	a.fail(w, 400, err); return
}
a.write(w, 200, true)
```

---

### 4.7 `DELETE /marriages/{id}` — delink a spouse (disband the marriage)

- **UI action:** **Actions → Delink spouse…** Next to the confirm button the app states how many children will be orphaned, so the copy and the behaviour must agree.
- **Frontend call site:** `src/api/marriages.ts` → `deleteMarriage()`.
- **Request:** none.
- **Response `200`:** `true`
- **Errors:** `404` unknown marriage (the UI reports the message as-is).
- **Connect it:** `marriage.DeleteMarriage(a.ctx, a.driver, r.PathValue("id"))`. The children remain people but lose their `PRODUCED` link with the marriage — which is what the UI's "left without parents" wording promises.

---

### 4.8 `POST /marriages/{id}/children` — add a child to an existing marriage

- **UI action:** **Actions → Add child…** when the other parent is already a spouse; the child may be an existing person or one just created.
- **Frontend call site:** `src/api/marriages.ts` → `addChild()`.
- **Request:** `{ "childId": "id-c" }`
- **Response `200`:** `true` (the frontend re-reads the person afterwards, so no body content is needed).
- **Errors:** `400` when the child or the marriage does not exist, `404` if you prefer that for both.
- **Connect it:** `children.CreateNewChild(a.ctx, a.driver, r.PathValue("id"), body.ChildId)`.

---

### 4.9 `DELETE /marriages/{id}/children/{childId}` — remove a child, or delete parents

- **UI actions:** **Remove child** on a child card, and **Delete parents** (which just delinks this person from their parents' marriage — the two parents stay married).
- **Frontend call site:** `src/api/marriages.ts` → `removeChild()`.
- **Request:** none.
- **Response `200`:** `true`
- **Errors:** the package returns an error when no relationship was deleted; answer `400` with that message (the UI surfaces it).
- **Connect it — mind the argument order:** the route reads `{id}` as the marriage and `{childId}` as the person, but the package takes the person first:

```go
if _, err := children.DeleteChildren(a.ctx, a.driver, r.PathValue("childId"), r.PathValue("id")); err != nil {
	a.fail(w, 400, err); return
}
a.write(w, 200, true)
```

## 5. Which UI action calls which endpoint

| UI action | Endpoint | Client function |
| --- | --- | --- |
| Open any person, reload after every write | `GET /people/{id}` | `getPersonDetails` |
| Palette, spouse/child/parent selectors | `GET /people` | `getAllPeople` |
| Create new person (in any dialog) | `POST /people` | `createPerson` |
| Edit person (name, dates, gender, alive) | `PATCH /people/{id}` | `updateUserDetails` |
| Add spouse | `POST /marriages` | `createMarriage` |
| Add parents | `POST /marriages` with `childrenIds: [personId]` | `createMarriage` |
| Add child, other parent is a spouse | `POST /marriages/{id}/children` | `addChild` |
| Add child, other parent is not a spouse yet | `POST /marriages` with `childrenIds` | `createMarriage` |
| Change dates | `PATCH /marriages/{id}` | `updateMarriage` |
| Delink spouse | `DELETE /marriages/{id}` | `deleteMarriage` |
| Remove child | `DELETE /marriages/{id}/children/{childId}` | `removeChild` |
| Delete parents | `DELETE /marriages/{id}/children/{childId}` (the parents' marriage, this person as the child) | `removeChild` |

## 6. Rules the server must own

These are enforced in the packages today and must stay server-side. The client no longer re-implements any of them:

- A person needs a name, a valid gender, and dates in `DD-MM-YYYY`; death cannot precede birth; nobody alive may carry a death date.
- A marriage needs two distinct people of **opposite** gender (`CreateNewMarriage` rejects two males or two females).
- A marriage starts after both spouses were born and ends on or before either death.
- The people in a marriage are immutable — there is deliberately no endpoint that changes them.
- `CreateFamily` rolls back people it created if the marriage or a child link fails.

## 7. Suggested order of work

1. `db.ConnectDatabase` + the `api` struct + `write`/`fail` helpers, then `GET /people` — this unblocks the palette.
2. `GET /people/{id}` (the composition above) — this unblocks the details page, which is the whole app.
3. `POST /people`, `PATCH /people/{id}`.
4. `POST /marriages`, then `POST /marriages/{id}/children` and `DELETE /marriages/{id}/children/{childId}`.
5. `PATCH /marriages/{id}`, `DELETE /marriages/{id}`.

After each endpoint, open the app, filter DevTools on `api-contract`, and compare the logged request with this document.

## 8. Gaps and risks

| Item | Note |
| --- | --- |
| `GET /people` has no backing function | New Cypher query, specified in §4.1. |
| `GET /people/{id}` has no backing function | Compose it, specified in §4.2. `integration.GetFamilyCertificate` alone is not enough. |
| `personFatherName` does not exist in the graph | Omit it from the list response; the palette hides the subtitle when absent. |
| `children.DeleteChildren` argument order | Person first, marriage second — the reverse of the route. |
| Route paths | Register `/people`, `/marriages`, … **without** `/api`: the Vite proxy strips that prefix. |
| `GetSpouses` / `GetChildren` per marriage | `GET /people/{id}` calls them once per marriage. Fine at this scale; batch them if a person has many marriages. |
| Dates | Always `DateProper` (`DD-MM-YYYY`). A raw ISO string is only converted by `GetProperDate` on input. |

## 9. Frontend file map

| Path | Role |
| --- | --- |
| `Frontend/src/api/contracts.ts` | Every request/response interface and the `ENDPOINTS` paths. Change a path here, not at the call site. |
| `Frontend/src/api/http.ts` | The only `fetch` call: base URL, JSON, `ApiError` with the API's message and status, cancellation. |
| `Frontend/src/api/people.ts` | `getAllPeople`, `createPerson`, `updateUserDetails`. |
| `Frontend/src/api/family.ts` | `getPersonDetails` (404 resolves to `null`). |
| `Frontend/src/api/marriages.ts` | `createMarriage`, `addChild`, `removeChild`, `deleteMarriage`, `updateMarriage`. |
| `Frontend/src/api/contractLog.ts` | The `[api-contract]` DevTools logging. |
| `Frontend/api-contracts/README.md` | The same contract, from the frontend's point of view, with the marriage model. |
| `Frontend/docs/manual-qa.md` | The checklist to run against a live API. |
