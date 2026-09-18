import type { Person } from "../api/contracts.ts";

/**
 * The domain model. These are the shapes the Go API sends on the wire: no json
 * tags are set on the Go structs, so the field names are the Go field names,
 * including the `DeateOfDeath` and `Chidren` spellings.
 *
 * Dates are `DD-MM-YYYY` strings (`person.DateProper`), or `""` / `null` when
 * unknown. Nothing in this module talks to the network.
 */
export interface User {
  Id: string;
  Name: string;
  DateOfBirth: string;
  DeateOfDeath: string;
  Gender: "Male" | "Female";
  Alive: boolean;
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
  Chidren?: User[] | null;
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
export function newUser(Id: string, Name = "", DateOfBirth = "", Gender: User["Gender"] = "Male"): User {
  return { Id, Name, DateOfBirth, Gender, Alive: true, DeateOfDeath: "" };
}

/** Marriages pair opposite genders; the Go API rejects two spouses of one gender. */
export function oppositeGender(gender: User["Gender"]): User["Gender"] {
  return gender === "Male" ? "Female" : "Male";
}

/**
 * Converts the list/search model into the wire model used inside marriage
 * certificates, so a person picked from a selector can be stored as a spouse,
 * parent or child without a second lookup.
 */
export function personToUser(person: Person): User {
  return {
    Id: person.id,
    Name: person.name,
    Gender: person.gender,
    Alive: person.alive,
    DateOfBirth: person.dateOfBirth ?? "",
    DeateOfDeath: person.dateOfDeath ?? "",
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
  const birth = dateValue(user.DateOfBirth);
  const end = user.Alive ? Date.UTC(now.getFullYear(), now.getMonth(), now.getDate()) : dateValue(user.DeateOfDeath);
  if (birth === null || end === null || end < birth) return "Age unknown";
  const startDate = new Date(birth);
  const endDate = new Date(end);
  let age = endDate.getUTCFullYear() - startDate.getUTCFullYear();
  if (endDate.getUTCMonth() < startDate.getUTCMonth() ||
      (endDate.getUTCMonth() === startDate.getUTCMonth() && endDate.getUTCDate() < startDate.getUTCDate())) age--;
  return user.Alive ? `Age ${age}` : `Died aged ${age}`;
}
/**
 * True when this person was born after every spouse whose birth date is known,
 * which is the rule for offering someone as a couple's child. An unknown birth
 * date on either side cannot disqualify a candidate, so only someone proven to
 * be the same age or older is rejected.
 */
export function isYoungerThan(spouses: User[], person: User): boolean {
  const birth = dateValue(person.DateOfBirth);
  if (birth === null) return true;
  return spouses.every((spouse) => {
    const spouseBirth = dateValue(spouse.DateOfBirth);
    return spouseBirth === null || birth > spouseBirth;
  });
}
export function otherSpouses(marriage: FamilyCertificate, personId: string): User[] {
  return (marriage.Spouse ?? []).filter((user) => user.Id !== personId);
}
export function sortedChildren(marriages: FamilyCertificate[]) {
  return marriages.flatMap((marriage) => (marriage.Chidren ?? []).map((user) => ({ user, familyId: marriage.Id })))
    .sort((a, b) => {
      const first = dateValue(a.user.DateOfBirth) ?? Infinity;
      const second = dateValue(b.user.DateOfBirth) ?? Infinity;
      return (first === second ? 0 : first < second ? -1 : 1) ||
        a.user.Name.localeCompare(b.user.Name) || a.user.Id.localeCompare(b.user.Id) || a.familyId.localeCompare(b.familyId);
    });
}
/** Per-ID colors survive sorting and deleting other marriages. */
export function familyColor(id: string): string {
  let hash = 0;
  for (const char of id) hash = (hash * 31 + char.charCodeAt(0)) >>> 0;
  return `hsl(${(hash * 137.508) % 360} 65% 52%)`;
}
