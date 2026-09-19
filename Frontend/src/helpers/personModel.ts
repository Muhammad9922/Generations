import type { Person } from "../api/contracts.ts";

/**
 * The domain model. `User` is the person shape `GET /people/{id}` actually
 * sends inside its certificate: the handler re-tags the person with the same
 * lowercase keys the list endpoint uses (`finalPeopleDesign` in
 * `Backend/cmd/server/person/query.go`), so the fields are `id`, `name`,
 * `gender`, `alive`, `dateOfBirth` and `dateOfDeath` — not the Go field names.
 *
 * The marriage container around those people keeps its Go field names (`Id`,
 * `Spouse`, `Children`, `StartOfFamily`, `EndOfFamily`), because only the
 * person shape carries json tags.
 *
 * Dates are `DD-MM-YYYY` strings (`person.DateProper`), or `""` / `null` when
 * unknown. Nothing in this module talks to the network.
 */
export interface User {
  id: string;
  name: string;
  dateOfBirth: string;
  dateOfDeath: string;
  gender: "Male" | "Female";
  alive: boolean;
}

/**
 * One marriage. Every group in the tree is the result of a marriage, so this
 * type is both "the marriage my parents had" and "a marriage I am a spouse in".
 * The people in a marriage are fixed once it exists; only its dates can change.
 */
export interface FamilyCertificate {
  Id: string;
  /** Both partners, matching the Go certificate. */
  Spouse?: User[] | null;
  /** The children this marriage produced; contains the person on their parents' marriage. */
  Children?: User[] | null;
  StartOfFamily?: string | null;
  EndOfFamily?: string | null;
}

/**
 * Opening a person returns the marriage they are a child of and every marriage
 * they are a spouse in. There is no standalone parent list because parents are
 * simply the spouses of `ParentsMarriage`, and a child only exists inside the
 * marriage that produced them.
 */
export interface PersonDetailsData {
  Person: User;
  /** The marriage that lists this person as a child; null once they are delinked. */
  ParentsMarriage: FamilyCertificate | null;
  /** Every marriage this person is a spouse in. */
  Marriages: FamilyCertificate[];
}

/** An empty person to edit in a dialog before it is sent to `CreatePerson`. */
export function newUser(id: string, name = "", dateOfBirth = "", gender: User["gender"] = "Male"): User {
  return { id, name, dateOfBirth, gender, alive: true, dateOfDeath: "" };
}

/** Marriages pair opposite genders; the Go API rejects two spouses of one gender. */
export function oppositeGender(gender: User["gender"]): User["gender"] {
  return gender === "Male" ? "Female" : "Male";
}

/**
 * Converts the list/search model into the certificate's person shape, so a
 * person picked from a selector can be stored as a spouse, parent or child
 * without a second lookup. The two models share their keys, so this only fills
 * the optional dates in with empty strings.
 */
export function personToUser(person: Person): User {
  return {
    id: person.id,
    name: person.name,
    gender: person.gender,
    alive: person.alive,
    dateOfBirth: person.dateOfBirth ?? "",
    dateOfDeath: person.dateOfDeath ?? "",
  };
}

/** Calendar validation avoids lexicographic sorting and date rollover. */
export function dateValue(value?: string | null): number | null {
  if (!value || !/^\d{2}-\d{2}-\d{4}$/.test(value)) return null;
  const [day, month, year] = value.split("-").map(Number);
  if (year < 1) return null;
  const date = new Date(0);
  date.setUTCFullYear(year, month - 1, day);
  date.setUTCHours(0, 0, 0, 0);
  return date.getUTCFullYear() === year && date.getUTCMonth() === month - 1 && date.getUTCDate() === day
    ? date.getTime() : null;
}
/** DD-MM-YYYY and YYYY-MM-DD are reverses of each other. */
export function reverseDate(value: string): string {
  return value ? value.split("-").reverse().join("-") : "";
}
export function toInputDate(value?: string | null): string {
  return dateValue(value) === null ? "" : reverseDate(value!);
}
export function ageLabel(user: User, now = new Date()): string {
  const birth = dateValue(user.dateOfBirth);
  const end = user.alive ? Date.UTC(now.getFullYear(), now.getMonth(), now.getDate()) : dateValue(user.dateOfDeath);
  if (birth === null || end === null || end < birth) return "Age unknown";
  const startDate = new Date(birth);
  const endDate = new Date(end);
  let age = endDate.getUTCFullYear() - startDate.getUTCFullYear();
  if (endDate.getUTCMonth() < startDate.getUTCMonth() ||
      (endDate.getUTCMonth() === startDate.getUTCMonth() && endDate.getUTCDate() < startDate.getUTCDate())) age--;
  return user.alive ? `Age ${age}` : `Died aged ${age}`;
}
/**
 * True when this person was born after every spouse whose birth date is known,
 * which is the rule for offering someone as a couple's child. An unknown birth
 * date on either side cannot disqualify a candidate, so only someone proven to
 * be the same age or older is rejected.
 */
export function isYoungerThan(spouses: User[], person: User): boolean {
  const birth = dateValue(person.dateOfBirth);
  if (birth === null) return true;
  return spouses.every((spouse) => {
    const spouseBirth = dateValue(spouse.dateOfBirth);
    return spouseBirth === null || birth > spouseBirth;
  });
}
export function otherSpouses(marriage: FamilyCertificate, personId: string): User[] {
  return (marriage.Spouse ?? []).filter((user) => user.id !== personId);
}

/** Orders people by name, falling back to the id so the order is always stable. */
export function compareNames(a: User, b: User): number {
  return a.name.localeCompare(b.name) || a.id.localeCompare(b.id);
}

/**
 * Orders by birth date, oldest first. A person with no recorded birth date
 * sorts last in either direction — an unknown date is missing information, not
 * evidence that someone was born earliest.
 */
export function compareByBirthAsc(a: User, b: User): number {
  const first = dateValue(a.dateOfBirth);
  const second = dateValue(b.dateOfBirth);
  if (first === null || second === null) {
    if (first === second) return compareNames(a, b);
    return first === null ? 1 : -1;
  }
  return first - second || compareNames(a, b);
}

/** Orders by birth date, youngest first, with the same rule for unknown dates. */
export function compareByBirthDesc(a: User, b: User): number {
  const first = dateValue(a.dateOfBirth);
  const second = dateValue(b.dateOfBirth);
  if (first === null || second === null) {
    if (first === second) return compareNames(a, b);
    return first === null ? 1 : -1;
  }
  return second - first || compareNames(a, b);
}
export function sortedChildren(marriages: FamilyCertificate[]) {
  return marriages.flatMap((marriage) => (marriage.Children ?? []).map((user) => ({ user, familyId: marriage.Id })))
    .sort((a, b) => {
      const first = dateValue(a.user.dateOfBirth) ?? Infinity;
      const second = dateValue(b.user.dateOfBirth) ?? Infinity;
      return (first === second ? 0 : first < second ? -1 : 1) ||
        a.user.name.localeCompare(b.user.name) || a.user.id.localeCompare(b.user.id) || a.familyId.localeCompare(b.familyId);
    });
}
/** Per-ID colors survive sorting and deleting other marriages. */
export function familyColor(id: string): string {
  let hash = 0;
  for (const char of id) hash = (hash * 31 + char.charCodeAt(0)) >>> 0;
  return `hsl(${(hash * 137.508) % 360} 65% 52%)`;
}
