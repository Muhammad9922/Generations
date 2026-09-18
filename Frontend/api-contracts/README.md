# Backend integration functions

These are local, typed API skeletons. They currently use in-memory data so the interface can be used before a backend endpoint exists. Replace the marked local operations with the indicated HTTP requests without changing callers.

## People — `src/api/people.ts`

| Function | Description | Future endpoint |
| --- | --- | --- |
| `getAllPeople(): Promise<Person[]>` | Returns all people for search and every spouse, child, and parent selector. Father and date fields are optional when unknown. | `GET /people` |
| `createPerson(params: CreatePersonParams): Promise<string>` | Validates and creates a person; resolves their ID. Params: `name`, `gender`, `alive`, `date_of_birth`, `date_of_death`. | `POST /people` |
| `updatePerson(params: UpdatePersonParams): Promise<void>` | Applies a typed partial update to an existing person. | `PATCH /people/:id` |

## Marriages — `src/api/marriages.ts`

| Function | Description | Future endpoint |
| --- | --- | --- |
| `createMarriage(params: CreateMarriageParams): Promise<string>` | Creates a marriage and resolves its ID. It requires distinct spouses, validates dates, requires marriage after either known birth date, and keeps marriage end at or before either known death date. | `POST /marriages` |
| `updateMarriage(params: UpdateMarriageParams): Promise<void>` | Applies a typed partial update to an existing marriage and reruns the same date rules. | `PATCH /marriages/:id` |

## Initial family data — `src/api/family.ts`

| Function | Description | Future endpoint |
| --- | --- | --- |
| `getParents(personId): Promise<User[]>` | Returns the requested person's parents. | `GET /people/:id/parents` |
| `getSpousesWithChildren(personId): Promise<SpouseWithChildren[]>` | Returns each spouse with the children shared in that marriage. | `GET /people/:id/spouses` |
| `getInitialFamilyData(personId): Promise<InitialFamilyData>` | Returns all initial relationship data in one UI-ready object. | `GET /people/:id/family` |

`getInitialFamilyData` returns this exact structure:

```ts
{
  parents: User[],
  spouses: [{ ...User, children: User[] }],
}
```

`src/api/id.ts` supplies compatible local IDs for this prototype. It deliberately does not rely on `crypto.randomUUID`, which is unavailable in some browser and test environments.
