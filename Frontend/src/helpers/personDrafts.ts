import type { CreatePersonRequest } from "../api/contracts.ts";
import { createPerson } from "../api/people.ts";
import type { User } from "./personModel.ts";

/**
 * The person fields the spouse and parent dialogs collect before creating
 * someone. Dates are already in wire form (DD-MM-YYYY) or empty.
 */
export interface PersonDraft {
  name: string;
  birth: string;
  alive: boolean;
  death: string;
}

/** Maps a dialog draft onto the `CreatePerson` body. */
export function toCreatePersonRequest(draft: PersonDraft, gender: User["Gender"]): CreatePersonRequest {
  return {
    PersonName: draft.name.trim(),
    Gender: gender,
    DateOfBirth: draft.birth,
    DateOfDeath: draft.alive ? "" : draft.death,
    Alive: draft.alive,
  };
}

/**
 * `CreatePerson` from a dialog draft. It rejects an unnamed person with a message
 * naming which person is missing, then resolves the saved record so the caller
 * works with the ID the API assigned rather than one it invented.
 */
export async function createPersonFromDraft(draft: PersonDraft, gender: User["Gender"], who: string): Promise<User> {
  if (!draft.name.trim()) throw new Error(`The ${who} needs a name.`);
  return createPerson(toCreatePersonRequest(draft, gender));
}
