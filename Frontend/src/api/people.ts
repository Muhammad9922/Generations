import {
  ENDPOINTS, type CreatePersonRequest, type CreatePersonResponse, type ListPeopleResponse, type Person,
  type UpdatePersonRequest, type UpdatePersonResponse,
} from "./contracts.ts";
import { request } from "./http.ts";
import { withContractLog } from "./contractLog.ts";
import type { User } from "../helpers/personModel.ts";

export type { Person } from "./contracts.ts";

/** `GET /people` — the single source for the palette and every selector. */
export async function getAllPeople(signal?: AbortSignal): Promise<Person[]> {
  return withContractLog("GET /people", { payload: {} }, async () => {
    const body = await request<ListPeopleResponse>(ENDPOINTS.people, { signal });
    return body.people;
  }, (result) => ({ count: result.length }));
}

/**
 * `CreatePerson(User)` creates a person — a spouse, a child, or a parent — and
 * resolves the saved record, including the ID the API assigned.
 */
export async function createPerson(person: CreatePersonRequest, signal?: AbortSignal): Promise<CreatePersonResponse> {
  return withContractLog("CreatePerson", { payload: person }, () =>
    request<CreatePersonResponse>(ENDPOINTS.people, { method: "POST", body: person, signal }),
    (result) => ({ id: result.id }));
}

/**
 * `PATCH /people/:id` saves the editable profile of any person on a family page:
 * the selected person, a parent, a spouse or a child. Only these fields are ever
 * sent, so a death date is cleared for anyone still alive.
 */
export async function updateUserDetails(user: User, signal?: AbortSignal): Promise<UpdatePersonResponse> {
  const update: UpdatePersonRequest = {
    Name: user.name,
    Gender: user.gender,
    Alive: user.alive,
    DateOfBirth: user.dateOfBirth || null,
    DateOfDeath: user.alive ? null : user.dateOfDeath || null,
  };
  return withContractLog("PATCH /people/:id", { id: user.id, fields: Object.keys(update), payload: update }, () =>
    request<UpdatePersonResponse>(ENDPOINTS.person(user.id), { method: "PATCH", body: update, signal }),
    (result) => ({ id: result.id }));
}
