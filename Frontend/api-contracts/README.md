# HTTP contract

The person and marriage screens reach the app only through `src/api/*`. Every function there maps to one endpoint below, so the transport can change without touching a component.

The Go HTTP layer implements every call below: `Backend/cmd/server/person` and
`Backend/cmd/server/marriage` hold one file per operation, and `API_SCOPE.md` in
the repository root is the same contract written from the server's side.

## Calling convention

- **Base URL** comes from `VITE_API_BASE_URL` and defaults to `/api` (see `.env` and `src/config/env.ts`), which Vite proxies to `http://localhost:8080` (`vite.config.ts`), so the API needs no CORS handling in development. Point it at an absolute URL to call the API directly.
- **Errors** must be JSON with a message: `{ "error": "both spouses are male" }`. That string is shown to the user as-is, so write user-facing sentences. `message`, `Error` and `Message` are accepted too. A non-JSON body is reported as an error instead of being silently accepted.
- **Status codes**: `200`/`201` for success, `404` for an unknown person, `400` for a rejected write, `500` for a failure.
- **Field names** are mixed, because only the person shape carries json tags. A person is always `id`, `name`, `gender`, `alive`, `dateOfBirth`, `dateOfDeath` — both in `GET /people` and inside a `GET /people/:id` certificate (the `finalPeopleDesign` tags in `Backend/cmd/server/person/query.go`). The marriage container around those people keeps its Go names: `Id`, `Spouse`, `Children`, `StartOfFamily`, `EndOfFamily`. Request bodies keep the Go names too: `PersonName` on create, `Name` on the PATCH, plus `Gender`, `DateOfBirth`, `DateOfDeath`, `Alive`.
- **Dates** are `person.DateProper`: `DD-MM-YYYY`, or `""` / `null` when unknown. `*[]User` and `*DateProper` marshal to `null` when nil; the UI treats a null spouse or child list as empty.

## Endpoints

| UI action | Method and path | Request body | Response | Go to call |
| --- | --- | --- | --- | --- |
| Palette, every selector | `GET /people` | — | `{ "people": Person[] }` | `person.GetPersonList` |
| List Of Singles | `GET /people/singles` | — | `{ "people": Person[] }` | `person.GetSingleList` |
| Open a person | `GET /people/:id` | — | `PersonDetailsData`, or `404` | composed in `cmd/server/person/query.go`, see below |
| Create a person | `POST /people` | `CreatePersonRequest` | the created `User` | `person.CreateNewPerson` |
| Edit a person | `PATCH /people/:id` | `UpdatePersonRequest` | the saved `User` | `person.UpdatePerson`, then `person.GetPerson` |
| Add spouse, add parents | `POST /marriages` | `CreateMarriageRequest` | `{ "id": "…" }` | `marriage.CreateNewMarriage`, then `children.CreateNewChild` for each `childrenIds` entry |
| Change marriage dates | `PATCH /marriages/:id` | `{ "DateStart": "", "DateEnd": "" }` | `true` | `marriage.UpdateMarriageDates` |
| Delink spouse | `DELETE /marriages/:id` | — | `true` | `marriage.DeleteMarriage` |
| Add child to an existing marriage | `POST /marriages/:id/children` | `{ "childId": "…" }` | `true` | `children.CreateNewChild(marriageId, childId)` |
| Remove child, delete parents | `DELETE /marriages/:id/children/:childId` | — | `true` | `children.DeleteChildren(childId, marriageId)` — note the argument order |

## Reading a person

`GET /people/:id` returns everything the details page renders:

```ts
{
  Person: User,
  ParentsMarriage: FamilyCertificate | null,   // the marriage this person is a child of
  Marriages: FamilyCertificate[]               // every marriage this person is a spouse in
}
```

```ts
User {
  id: string,
  name: string,
  dateOfBirth: string,        // DD-MM-YYYY, or ""
  dateOfDeath: string,
  gender: "Male" | "Female",
  alive: boolean,
}

FamilyCertificate {
  Id: string,
  Spouse?: User[] | null,     // both partners
  Children?: User[] | null,   // holds this person on their parents' marriage
  StartOfFamily?: string | null,
  EndOfFamily?: string | null,
}
```

Build it from the packages that already exist: `person.GetPerson`, `children.GetMarriageThatOfChild(childId)` for the parents' marriage, `children.GetMarriageThatOfSpouse(spouseId)` for the rest, then `marriage.GetSpouses` and `children.GetChildren` per marriage to fill `Spouse` and `Children`. `integration.GetFamilyCertificate` is close, but it returns the single certificate found from a spouse, so it covers neither the parents' marriage nor several marriages. That composition is what `cmd/server/person/query.go` does — note that `marriage.GetMarriageFromMarriageId` returns one row per **marriage** (its dates), not one row per spouse.

An ID that belongs to nobody — or to a marriage — must answer `404`; the page then shows "Person not found".

## Request bodies

```ts
CreatePersonRequest   { PersonName, Gender, DateOfBirth, DateOfDeath, Alive }
UpdatePersonRequest   { Name?, Gender?, Alive?, DateOfBirth?, DateOfDeath? }  // omitted = unchanged, null = clear
CreateMarriageRequest { SpouseOne, SpouseTwo, DateStart, DateEnd, childrenIds? }
```

`POST /people` answers with the saved person, so the client never invents an ID:

```json
{ "id": "…", "name": "Ada", "gender": "Female", "dateOfBirth": "10-12-1815", "dateOfDeath": "", "alive": true }
```

`person.NewPerson` spells the field `PersonName` while `integration.User` spells it `Name`; the handler maps between the two exactly as `integration/get_family.go` already does.

## The list model

`GET /people` answers with the leaner shape the palette renders. Keys are lowercase, and a father's name is optional because the `Person` node stores no such property — omit it rather than sending an empty string:

```ts
{ id, name, personFatherName?, gender, alive, dateOfBirth?, dateOfDeath? }
```

## What the UI does after a write

Every write ends by re-reading. The details page re-runs `GET /people/:id`, and anything that created or renamed a person also re-runs `GET /people` so the palette updates. The client never patches its own copy of a certificate, so what is on screen is what the API holds.

## The marriage model

Every group in the tree is the result of a marriage. There is no standalone parent list and no parentless child:

- Opening a person returns the marriage they are a **child** of, plus every marriage they are a **spouse** in.
- A child exists only inside the marriage that produced them. That child link is the only thing that makes two spouses someone's parents.
- The people in a marriage are fixed once it exists. Swapping a spouse would break every child link hanging off the marriage, so a marriage can only have its dates changed, or be deleted and created again.
- Deleting a marriage is therefore how a spouse is removed, and it orphans that marriage's children: they stay people, but with undefined parents.

| User action | Operation | Effect |
| --- | --- | --- |
| Delete parents | `RemoveChild(parentsMarriage.Id, person.Id)` | This person stops being a child of that marriage. The two parents stay married to each other. |
| Delink spouse | `DeleteMarriage(marriage.Id)` | The marriage is disbanded, the spouse is unlinked, and every child of that marriage is left without parents. |
| Remove child | `RemoveChild(marriage.Id, child.Id)` | Only that child link goes; the marriage and both spouses stay. |
| Add spouse | `CreateMarriage(spouseOne, spouseTwo, start*, end*)` | Creates their marriage. |
| Add child | `AddChildren(marriage.Id, person.Id)`, or `CreateMarriage(…, childrenIds: [person.Id])` when the marriage does not exist yet | The person becomes a child of that marriage. |
| Add parents | `CreateMarriage(father, mother, start*, end*, childrenIds: [person.Id])` | The person becomes a child of the new marriage. |

`CreateMarriage`, `DeleteMarriage`, `RemoveChild`, `AddChildren` and `CreatePerson` are also the operation names shown in the contract log.

## Relationships are derived, not stored

There is no relationship endpoint, and the UI does not need one for the labels
it shows — they follow from the payloads above:

- a person's siblings are the other `Children` of their `ParentsMarriage`;
- each parent's `ParentsMarriage` carries the grandparents, and its `Children`
  carry the uncles and aunts;
- an uncle's or aunt's `Marriages` carry their cousins, and a sibling's carry
  their nieces and nephews.

`src/api/relations.ts` walks that in three rounds — the person, then the
parents, then the siblings and uncles together — rather than one request per
relative, and `src/helpers/relations.ts` turns the payloads into labelled
relatives. Someone is only ever labelled by their closest relation to the person
being read, so an inconsistent record cannot list the same person twice, or as
their own relative. Only blood relations are reported: the spouse of an uncle is
not an aunt.

## Who counts as single

`GET /people/singles` answers the same `{ "people": Person[] }` shape as
`GET /people`, narrowed to people who hold no `MARRIED` edge to a `Marriage`
node. It is a separate endpoint because being single is a graph fact, not a
property: answering it from the list endpoint would mean reading every person's
marriages first.

Two consequences are deliberate. Disbanding a marriage makes both spouses single
again without their nodes changing, and a widow or widower is **not** single,
because the marriage that made them a spouse is still recorded — the graph knows
a marriage happened, not whether it ended.

The query counts an `OPTIONAL MATCH` rather than writing
`WHERE NOT (p)-[:MARRIED]->(:Marriage)`. The store is Memgraph, which rejects a
pattern used as an atom expression ("Not yet implemented") and then retries until
the driver's budget runs out, turning a syntax problem into a 30-second 500.

## Where the pieces live

- `src/api/http.ts` — the only module that calls `fetch`: base URL, JSON, status codes turned into `ApiError` with the API's own message, cancellation passed through.
- `src/config/env.ts` — the only module that reads `import.meta.env`. Every setting is named in `src/vite-env.d.ts` and documented in `.env`.
- `src/api/contracts.ts` — every request and response interface above, plus the endpoint paths.
- `src/api/people.ts`, `marriages.ts`, `family.ts` — one typed function per endpoint.
- `src/api/relations.ts` — composes `GET /people/:id` into one person's extended family; it adds no endpoint of its own.
- `src/helpers/personModel.ts` — the wire model and the pure view helpers (`otherSpouses`, `sortedChildren`, `isYoungerThan`, date conversion, `familyColor`). It holds no state and makes no requests.
- `src/helpers/relations.ts` — labels those payloads as parents, siblings, uncles, aunts, cousins and the rest. Pure, and tested against a hand-built three-generation family.
- `src/helpers/personDrafts.ts` — maps a dialog's draft onto `CreatePersonRequest`.

## Debug logging

Every contract call records a structured request, response (including duration), or failure in development. In DevTools, each entry is a collapsed `[api-contract]` group with the labeled payload and metadata; do not enable these logs in production if the payload includes personal data.

Logging is on for Vite development by default. `VITE_API_CONTRACT_DEBUG` in `.env` overrides that: `false` suppresses it, `true` enables it in another build mode.
