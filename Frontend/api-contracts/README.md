# HTTP contract

The person and marriage screens reach the app only through `src/api/*`. Every function there maps to one endpoint below, so the transport can change without touching a component.

The Go packages for all of this already exist. What is missing is the `net/http` layer: `Backend/cmd/server/main.go` currently answers `"Hi there!"` on every path, so the frontend has nothing to read until those handlers are written.

## Calling convention

- **Base URL** `/api`, proxied to `http://localhost:8080` by Vite (`vite.config.ts`), so the API needs no CORS handling in development. Set `VITE_API_BASE_URL` to call it directly.
- **Errors** must be JSON with a message: `{ "error": "both spouses are male" }`. That string is shown to the user as-is, so write user-facing sentences. `message`, `Error` and `Message` are accepted too. A non-JSON body is reported as an error instead of being silently accepted.
- **Status codes**: `200`/`201` for success, `404` for an unknown person, `400` for a rejected write, `500` for a failure.
- **Field names** are mixed, because only the person shape carries json tags. A person is always `id`, `name`, `gender`, `alive`, `dateOfBirth`, `dateOfDeath` — both in `GET /people` and inside a `GET /people/:id` certificate (the `finalPeopleDesign` tags in `Backend/cmd/server/person/query.go`). The marriage container around those people keeps its Go names: `Id`, `Spouse`, `Children`, `StartOfFamily`, `EndOfFamily`. Request bodies keep the Go names too: `PersonName` on create, `Name` on the PATCH, plus `Gender`, `DateOfBirth`, `DateOfDeath`, `Alive`.
- **Dates** are `person.DateProper`: `DD-MM-YYYY`, or `""` / `null` when unknown. `*[]User` and `*DateProper` marshal to `null` when nil; the UI treats a null spouse or child list as empty.

## Endpoints

| UI action | Method and path | Request body | Response | Go to call |
| --- | --- | --- | --- | --- |
| Palette, every selector | `GET /people` | — | `{ "people": Person[] }` | **none yet** — needs a list-all query |
| Open a person | `GET /people/:id` | — | `PersonDetailsData`, or `404` | compose, see below |
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

Build it from the packages that already exist: `person.GetPerson`, `children.GetMarriageThatOfChild(childId)` for the parents' marriage, `children.GetMarriageThatOfSpouse(spouseId)` for the rest, then `marriage.GetSpouses` and `children.GetChildren` per marriage to fill `Spouse` and `Children`. `integration.GetFamilyCertificate` is close, but it returns the single certificate found from a spouse, so it covers neither the parents' marriage nor several marriages.

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

## Where the pieces live

- `src/api/http.ts` — the only module that calls `fetch`: base URL, JSON, status codes turned into `ApiError` with the API's own message, cancellation passed through.
- `src/api/contracts.ts` — every request and response interface above, plus the endpoint paths.
- `src/api/people.ts`, `marriages.ts`, `family.ts` — one typed function per endpoint.
- `src/helpers/personModel.ts` — the wire model and the pure view helpers (`otherSpouses`, `sortedChildren`, `isYoungerThan`, date conversion, `familyColor`). It holds no state and makes no requests.
- `src/helpers/personDrafts.ts` — maps a dialog's draft onto `CreatePersonRequest`.

## Debug logging

Every contract call records a structured request, response (including duration), or failure in development. In DevTools, each entry is a collapsed `[api-contract]` group with the labeled payload and metadata; do not enable these logs in production if the payload includes personal data.

Logging is on for Vite development by default. Set `VITE_API_CONTRACT_DEBUG=false` to suppress it, or `VITE_API_CONTRACT_DEBUG=true` to enable it in another build mode.
