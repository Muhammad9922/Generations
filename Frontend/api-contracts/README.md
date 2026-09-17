# Backend integration functions

These are local, typed API skeletons. They currently use in-memory data so the interface can be used before a backend endpoint exists. Replace the marked local operations with the indicated HTTP requests without changing callers.

## People — `src/api/people.ts`

| Function | Description | Future endpoint |
| --- | --- | --- |
| `getAllPeople(): Promise<Person[]>` | Returns all people for search and relationship selectors. | `GET /people` |
| `createPerson(params: CreatePersonParams): Promise<string>` | Validates and creates a person; resolves their ID. Params: `name`, `gender`, `alive`, `date_of_birth`, `date_of_death`. | `POST /people` |
| `updatePerson(params: UpdatePersonParams): Promise<void>` | Applies a typed partial update to an existing person. | `PATCH /people/:id` |

## Marriages — `src/api/marriages.ts`

| Function | Description | Future endpoint |
| --- | --- | --- |
| `createMarriage(params: CreateMarriageParams): Promise<string>` | Creates a marriage and resolves its ID. It requires distinct spouses, validates dates, requires marriage after either known birth date, and keeps marriage end at or before either known death date. | `POST /marriages` |
| `updateMarriage(params: UpdateMarriageParams): Promise<void>` | Applies a typed partial update to an existing marriage and reruns the same date rules. | `PATCH /marriages/:id` |

`src/api/id.ts` supplies compatible local IDs for this prototype. It deliberately does not rely on `crypto.randomUUID`, which is unavailable in some browser and test environments.
