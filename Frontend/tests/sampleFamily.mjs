import { newUser } from "../src/helpers/personModel.ts";

/**
 * The sample family the UI used to ship with, kept here as test data. The app now
 * reads every relationship from the API, so this exists only to give the pure
 * helpers and the API mapping realistic input.
 */

/** The nine sample people, as `GET /people` would list them. */
export const samplePeople = [
  { personFatherName: "Muhammad Saleem", name: "Mahammad Muhayodin", id: "id-1", gender: "Male", alive: true, dateOfBirth: "14-03-1980" },
  { personFatherName: "Tariq Mahmood", name: "Hamza Tariq", id: "id-2", gender: "Male", alive: true, dateOfBirth: "14-03-1980" },
  { personFatherName: "Abdul Rahman", name: "Usman Abdul", id: "id-3", gender: "Male", alive: true, dateOfBirth: "14-03-1980" },
  { personFatherName: "Bilal Ahmed", name: "Zaid Bilal", id: "id-4", gender: "Female", alive: true, dateOfBirth: "14-03-1980" },
  { personFatherName: "Rashid Khan", name: "Omar Rashid", id: "id-5", gender: "Male", alive: true, dateOfBirth: "14-03-1980" },
  { personFatherName: "Abdul Rauf", name: "Usman Abdul", id: "id-6", gender: "Male", alive: true, dateOfBirth: "14-03-1980" },
  { name: "Fatima Noor", id: "id-7", gender: "Female", alive: true },
  { personFatherName: "Khalid Hassan", name: "Ibrahim Khalid", id: "id-8", gender: "Male", alive: true, dateOfBirth: "09-07-1992" },
  { personFatherName: "Nadia Omar", name: "Layla Omar", id: "id-9", gender: "Female", alive: false, dateOfBirth: "12-01-1978", dateOfDeath: "04-06-2021" },
];

const WITHOUT_PARENTS = new Set(["id-2", "id-6", "id-7"]);
const WITHOUT_MARRIAGE = new Set([...WITHOUT_PARENTS, "id-5", "id-8"]);

/** Builds every person and every marriage in the sample family. */
export function sampleFamily() {
  const users = new Map();
  for (const person of samplePeople) {
    users.set(person.id, { ...newUser(person.id, person.name, person.dateOfBirth ?? "14-03-1980", person.gender), Alive: person.alive, DeateOfDeath: person.dateOfDeath ?? "" });
  }
  const relative = (id, name, dateOfBirth = "", gender = "Male") => {
    if (!users.has(id)) users.set(id, newUser(id, name, dateOfBirth, gender));
    return users.get(id);
  };
  const marriages = [];
  for (const person of samplePeople) {
    const self = users.get(person.id);
    if (!WITHOUT_PARENTS.has(person.id)) {
      marriages.push({
        Id: `${person.id}-parents-marriage`,
        Spouse: [
          relative(`${person.id}-parent-1`, person.personFatherName ?? "Unknown parent", "02-06-1952"),
          relative(`${person.id}-parent-2`, "Amina Saleem", "19-11-1956", "Female"),
        ],
        Chidren: [self],
        StartOfFamily: null,
        EndOfFamily: null,
      });
    }
    if (WITHOUT_MARRIAGE.has(person.id)) continue;
    const marriage = {
      Id: `${person.id}-marriage-1`,
      Spouse: [self, relative(`${person.id}-spouse-1`, "Sara Ahmed", "08-09-1983", "Female")],
      StartOfFamily: "12-06-2005",
      Chidren: person.id === "id-3" ? null : [
        relative(`${person.id}-child-2`, "Yusuf Ahmed", "23-08-2014"),
        relative(`${person.id}-child-1`, "Maryam Ahmed", "05-02-2007", "Female"),
      ],
    };
    marriages.push(marriage);
    if (person.id === "id-1") {
      marriage.EndOfFamily = "15-10-2015";
      marriages.push({
        Id: `${person.id}-marriage-2`,
        Spouse: [self, relative(`${person.id}-spouse-2`, "Hana Ali", "21-04-1985", "Female")],
        StartOfFamily: "10-02-2017", EndOfFamily: null,
        Chidren: [
          relative(`${person.id}-child-4`, "Adam Ahmed", "11-07-2021"),
          relative(`${person.id}-child-3`, "Noor Ahmed", "18-12-2018", "Female"),
        ],
      });
    }
  }
  return { users, marriages };
}

/** The payload `GET /people/:id` answers with, derived from the sample family. */
export function samplePersonDetails(id) {
  const { users, marriages } = sampleFamily();
  const person = users.get(id);
  if (!person) return null;
  return {
    Person: person,
    ParentsMarriage: marriages.find((marriage) => (marriage.Chidren ?? []).some((child) => child.Id === id)) ?? null,
    Marriages: marriages.filter((marriage) => (marriage.Spouse ?? []).some((spouse) => spouse.Id === id)),
  };
}
