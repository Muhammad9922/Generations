import { GetAllUsers, type Person } from "./GetUsers.ts";

/** Go wire names preserved. Dates are DD-MM-YYYY or empty. */
export interface User {
  Id: string;
  Name: string;
  DateOfBirth: string;
  DeateOfDeath: string;
  Gender: "Male" | "Female";
  Alive: boolean;
}
export interface FamilyCertificate {
  Id: string;
  /** Both partners, matching the Go certificate. */
  Spouse?: User[] | null;
  Chidren?: User[] | null;
  StartOfFamily?: string | null;
  EndOfFamily?: string | null;
}
export interface PersonDetailsData {
  Person: User;
  Parents?: [User, User] | null;
  Families: FamilyCertificate[];
}
export function newUser(Id: string, Name = "", DateOfBirth = "", Gender: User["Gender"] = "Male"): User {
  return { Id, Name, DateOfBirth, Gender, Alive: true, DeateOfDeath: "" };
}

/** Sample ID lookup, with fresh records per call; no backend writes. */
export async function GetPersonDetails(id: string, people?: Person[]): Promise<PersonDetailsData | null> {
  const person = (people ?? await GetAllUsers()).find((item) => item.id === id);
  if (!person) return null;
  const user = { ...newUser(id, person.name, person.dateOfBirth ?? "14-03-1980", person.gender), Alive: person.alive, DeateOfDeath: person.dateOfDeath ?? "" };
  const result: PersonDetailsData = { Person: user, Parents: null, Families: [] };
  if (id === "id-2" || id === "id-6" || id === "id-7") return result;
  result.Parents = [
    newUser(`${id}-parent-1`, person.personFatherName ?? "Unknown parent", "02-06-1952"),
    newUser(`${id}-parent-2`, "Amina Saleem", "19-11-1956", "Female"),
  ];
  if (id === "id-5" || id === "id-8") return result;
  result.Families.push({
    Id: `${id}-marriage-1`,
    Spouse: [user, newUser(`${id}-spouse-1`, "Sara Ahmed", "08-09-1983", "Female")],
    StartOfFamily: "12-06-2005",
    Chidren: id === "id-3" ? null : [
      newUser(`${id}-child-2`, "Yusuf Ahmed", "23-08-2014"),
      newUser(`${id}-child-1`, "Maryam Ahmed", "05-02-2007", "Female"),
    ],
  });
  if (id === "id-1") {
    result.Families[0].EndOfFamily = "15-10-2015";
    result.Families.push({
      Id: `${id}-marriage-2`,
      Spouse: [user, newUser(`${id}-spouse-2`, "Hana Ali", "21-04-1985", "Female")],
      StartOfFamily: "10-02-2017", EndOfFamily: null,
      Chidren: [
        newUser(`${id}-child-4`, "Adam Ahmed", "11-07-2021"),
        newUser(`${id}-child-3`, "Noor Ahmed", "18-12-2018", "Female"),
      ],
    });
  }
  return result;
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
export function toInputDate(value?: string | null): string {
  return dateValue(value) === null ? "" : value!.split("-").reverse().join("-");
}
export function fromInputDate(value: string): string {
  return value ? value.split("-").reverse().join("-") : "";
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
export function otherSpouses(family: FamilyCertificate, personId: string): User[] {
  return (family.Spouse ?? []).filter((user) => user.Id !== personId);
}
export function sortedChildren(families: FamilyCertificate[]) {
  return families.flatMap((family) => (family.Chidren ?? []).map((user) => ({ user, familyId: family.Id })))
    .sort((a, b) => {
      const first = dateValue(a.user.DateOfBirth) ?? Infinity;
      const second = dateValue(b.user.DateOfBirth) ?? Infinity;
      return (first === second ? 0 : first < second ? -1 : 1) ||
        a.user.Name.localeCompare(b.user.Name) || a.user.Id.localeCompare(b.user.Id) || a.familyId.localeCompare(b.familyId);
    });
}
export function replaceUser(data: PersonDetailsData, user: User): PersonDetailsData {
  const replace = (item: User) => item.Id === user.Id ? user : item;
  return {
    ...data, Person: replace(data.Person),
    Parents: data.Parents ? [replace(data.Parents[0]), replace(data.Parents[1])] : data.Parents,
    Families: data.Families.map((family) => ({
      ...family, Spouse: family.Spouse?.map(replace), Chidren: family.Chidren?.map(replace),
    })),
  };
}
export function withParents(data: PersonDetailsData, parents: User[]): PersonDetailsData {
  if (parents.length !== 2 || parents[0].Id === parents[1].Id || parents.some((parent) => parent.Id === data.Person.Id)) {
    throw new Error("Add two distinct parents, neither of whom is the selected person.");
  }
  return { ...data, Parents: [parents[0], parents[1]] };
}
export function deleteMarriage(data: PersonDetailsData, id: string): PersonDetailsData {
  if (data.Families.find((family) => family.Id === id)?.Chidren?.length) {
    throw new Error("This marriage has linked children. Reassign or unlink them before deleting it.");
  }
  return { ...data, Families: data.Families.filter((family) => family.Id !== id) };
}
/** Per-ID colors survive sorting and deleting other marriages. */
export function familyColor(id: string): string {
  let hash = 0;
  for (const char of id) hash = (hash * 31 + char.charCodeAt(0)) >>> 0;
  return `hsl(${(hash * 137.508) % 360} 65% 52%)`;
}
