import type { PersonDetailsData, User } from "../helpers/personModel.ts";

/**
 * Where the Go API lives. In development `/api` is proxied to
 * http://localhost:8080 (see vite.config.ts), so the browser stays on one
 * origin and the server needs no CORS handling. Set VITE_API_BASE_URL to call
 * the API directly.
 */
const environment = (import.meta as ImportMeta & { env?: { VITE_API_BASE_URL?: string } }).env;
export const API_BASE_URL: string = environment?.VITE_API_BASE_URL || "/api";

/** Every path the frontend calls, so moving an endpoint is one edit here. */
export const ENDPOINTS = {
  people: "/people",
  person: (id: string) => `/people/${encodeURIComponent(id)}`,
  marriages: "/marriages",
  marriage: (id: string) => `/marriages/${encodeURIComponent(id)}`,
  marriageChildren: (id: string) => `/marriages/${encodeURIComponent(id)}/children`,
  marriageChild: (marriageId: string, childId: string) =>
    `/marriages/${encodeURIComponent(marriageId)}/children/${encodeURIComponent(childId)}`,
} as const;

/**
 * The list/search model the palette and every selector render. It uses the same
 * lowercase keys as `User`, the person shape carried inside marriage
 * certificates.
 */
export interface Person {
  id: string;
  name: string;
  personFatherName?: string;
  gender: User["gender"];
  alive: boolean;
  dateOfBirth?: string;
  dateOfDeath?: string;
}

/** `GET /people` — every person the palette and the selectors may offer. */
export interface ListPeopleResponse {
  people: Person[];
}

/** Body of `CreatePerson`, matching the Go `person.NewPerson`. */
export interface CreatePersonRequest {
  PersonName: string;
  Gender: User["gender"];
  /** `DateProper`: `DD-MM-YYYY`, or `""` when unknown. */
  DateOfBirth: string;
  DateOfDeath: string;
  Alive: boolean;
}

/** `CreatePerson` resolves the saved person, including the ID the API assigned. */
export type CreatePersonResponse = User;

/**
 * Body of `PATCH /people/:id`, matching the Go `person.UpdateUser`, whose fields
 * are pointers: an omitted field keeps its stored value and an explicit `null`
 * clears it. Note the name is `Name` here but `PersonName` on create — those are
 * the Go struct field names.
 */
export interface UpdatePersonRequest {
  Name?: string;
  Gender?: User["gender"];
  Alive?: boolean;
  DateOfBirth?: string | null;
  DateOfDeath?: string | null;
}

/** `PATCH /people/:id` resolves the saved person. */
export type UpdatePersonResponse = User;

/** Body of `CreateMarriage`, matching the Go `marriage.NewMarriage`. */
export interface CreateMarriageRequest {
  SpouseOne: string;
  SpouseTwo: string;
  DateStart: string;
  DateEnd: string;
  /**
   * Children the marriage starts with. The Go package links these one at a time
   * (`children.CreateNewChild`); the handler can loop over them.
   */
  childrenIds?: string[];
}

/** `CreateMarriage` resolves the marriage ID the API assigned. */
export interface CreateMarriageResponse {
  id: string;
}

/** Body of `PATCH /marriages/:id` — dates only (`marriage.UpdateMarriageDates`). */
export interface UpdateMarriageRequest {
  DateStart: string;
  DateEnd: string;
}

/** Body of `AddChildren(marriageID, userId)`. */
export interface AddChildRequest {
  childId: string;
}

/** `GET /people/:id` — everything the details page renders. */
export type PersonDetailsResponse = PersonDetailsData;

/** What the `RemoveChild`, `DeleteMarriage` and `AddChildren` endpoints answer. */
export type WriteResponse = boolean;
