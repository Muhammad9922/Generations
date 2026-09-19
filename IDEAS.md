# Ideas For *Version One*
## API Requiremnts 
1. List of users should be fetchable

# Ideas For *Version Two*

- Adding profile picture support for Persons
- Add Relationship Finder
- Add Search For Person Name

# Frontend Fixes

## Issue 1 — make the details page read the keys the API actually sends

`GET /people/{id}` answers with the *list* casing inside the certificate:
`Person`, `ParentsMarriage.Spouse` and the children all use `id`, `name`,
`gender`, `alive`, `dateOfBirth`, `dateOfDeath` (the `finalPeopleDesign` tags in
`Backend/cmd/server/person/query.go`). The frontend's `User` model still expects
the Go field names (`Id`, `Name`, `DateOfBirth`, `DeateOfDeath`, `Gender`,
`Alive`), so `data.Person.Name` and every card field come back `undefined` and
the family page renders blank.

Fix it on the frontend, not the server:

- `Frontend/src/helpers/personModel.ts` — change `User` (and the marriage
  `FamilyCertificate`) to the lowercase keys, or map the response once in
  `Frontend/src/api/family.ts` before the page sees it.
- Update every reader of those fields: `FamilyView.tsx`, `FamilyPersonCard`,
  `ChildCreatorDialog`, `SpouseCreatorDialog`, `ParentCreatorDialog`, and the
  helpers `personToUser`, `sortedChildren`, `otherSpouses`, `isYoungerThan`,
  `ageLabel`.
- `Frontend/src/api/contracts.ts` — `PersonDetailsResponse`,
  `CreatePersonResponse` and `UpdatePersonResponse` all resolve to `User`, so
  they follow whichever shape `User` ends up with.
- Also update `API_SCOPE.md` §4.2 (and the §4.1/§4.2 table), which currently
  documents the capitalized certificate keys. If the doc is left as-is the next
  backend change will put the capitalized keys back.

Done when: opening a person with parents, several marriages and children shows
their name, dates and every relative, with no `undefined` text.
