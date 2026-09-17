import { dateValue, type User } from "../helpers/GetPersonDetails.ts";
import { createLocalId } from "./id.ts";

export interface Person {
  id: string;
  name: string;
  personFatherName: string;
  gender: User["Gender"];
  alive: boolean;
  dateOfBirth?: string;
  dateOfDeath?: string;
}
export interface CreatePersonParams {
  name: string;
  gender: User["Gender"];
  alive: boolean;
  date_of_birth?: string | null;
  date_of_death?: string | null;
}
export interface UpdatePersonParams extends Partial<CreatePersonParams> { id: string; }

const people: Person[] = [
  { personFatherName: "Muhammad Saleem", name: "Mahammad Muhayodin", id: "id-1", gender: "Male", alive: true, dateOfBirth: "14-03-1980" },
  { personFatherName: "Tariq Mahmood", name: "Hamza Tariq", id: "id-2", gender: "Male", alive: true, dateOfBirth: "14-03-1980" },
  { personFatherName: "Abdul Rahman", name: "Usman Abdul", id: "id-3", gender: "Male", alive: true, dateOfBirth: "14-03-1980" },
  { personFatherName: "Bilal Ahmed", name: "Zaid Bilal", id: "id-4", gender: "Female", alive: true, dateOfBirth: "14-03-1980" },
  { personFatherName: "Rashid Khan", name: "Omar Rashid", id: "id-5", gender: "Male", alive: true, dateOfBirth: "14-03-1980" },
  { personFatherName: "Abdul Rauf", name: "Usman Abdul", id: "id-6", gender: "Male", alive: true, dateOfBirth: "14-03-1980" },
];

function validatePerson(params: Omit<CreatePersonParams, "name"> & { name?: string }): void {
  const birth = params.date_of_birth ? dateValue(params.date_of_birth) : null;
  const death = params.date_of_death ? dateValue(params.date_of_death) : null;
  if ((params.date_of_birth && birth === null) || (params.date_of_death && death === null)) throw new Error("Enter valid calendar dates.");
  if (birth !== null && death !== null && death < birth) throw new Error("Death date cannot be before birth date.");
}

/** Query skeleton: replace its local return with GET /people. */
export async function getAllPeople(): Promise<Person[]> { return people.map((person) => ({ ...person })); }

/** Creation skeleton: replace the local append with POST /people; resolves the new ID. */
export async function createPerson(params: CreatePersonParams): Promise<string> {
  if (!params.name.trim()) throw new Error("A person needs a name.");
  validatePerson(params);
  const id = createLocalId("person");
  people.push({ id, name: params.name.trim(), personFatherName: "", gender: params.gender, alive: params.alive, dateOfBirth: params.date_of_birth ?? "", dateOfDeath: params.alive ? "" : params.date_of_death ?? "" });
  return id;
}

/** Update skeleton: replace the local patch with PATCH /people/:id. */
export async function updatePerson(params: UpdatePersonParams): Promise<void> {
  const index = people.findIndex((person) => person.id === params.id);
  if (index < 0) throw new Error("Person not found.");
  const current = people[index];
  const next = { ...current, ...params, dateOfBirth: params.date_of_birth ?? current.dateOfBirth, dateOfDeath: params.date_of_death ?? current.dateOfDeath };
  validatePerson({ name: next.name, gender: next.gender, alive: next.alive, date_of_birth: next.dateOfBirth, date_of_death: next.dateOfDeath });
  people[index] = next;
}

/** Internal lookup used by marriage validation until the backend owns this rule. */
export function registerPeopleForMarriageValidation(users: User[]): void {
  for (const user of users) {
    const existing = people.find((person) => person.id === user.Id);
    if (existing) Object.assign(existing, { dateOfBirth: user.DateOfBirth, dateOfDeath: user.DeateOfDeath, alive: user.Alive, gender: user.Gender, name: user.Name });
    else people.push({ id: user.Id, name: user.Name, personFatherName: "", gender: user.Gender, alive: user.Alive, dateOfBirth: user.DateOfBirth, dateOfDeath: user.DeateOfDeath });
  }
}

export function personForMarriageValidation(id: string): Person | undefined { return people.find((person) => person.id === id); }
