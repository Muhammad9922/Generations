import assert from "node:assert/strict";
import test from "node:test";
import {
  ageLabel, compareByBirthAsc, compareByBirthDesc, compareNames, dateValue, familyColor, isYoungerThan, newUser,
  oppositeGender, otherSpouses, personToUser, reverseDate, sortedChildren, toInputDate,
} from "../src/helpers/personModel.ts";
import { samplePeople, samplePersonDetails } from "./sampleFamily.mjs";

test("dates validate leap years and pre-1970 values; age respects birthdays and death", () => {
  assert.notEqual(dateValue("29-02-2000"), null);
  assert.equal(dateValue("29-02-1900"), null);
  assert.equal(dateValue("31-04-2000"), null);
  assert.equal(dateValue("01-01-0000"), null);
  assert.ok(dateValue("01-01-1900") < 0);
  assert.equal(reverseDate(toInputDate("03-02-1900")), "03-02-1900");
  assert.equal(toInputDate(null), "");
  assert.equal(toInputDate("31-02-2000"), "");
  const user = newUser("p", "Person", "18-09-2000");
  assert.equal(ageLabel(user, new Date(2026, 8, 17)), "Age 25");
  assert.equal(ageLabel(user, new Date(2026, 8, 18)), "Age 26");
  assert.equal(ageLabel({ ...user, alive: false, dateOfDeath: "17-09-2020" }), "Died aged 19");
  assert.equal(ageLabel({ ...user, alive: false, dateOfDeath: "" }), "Age unknown");
});

test("children sort globally across marriages, with invalid dates last and stable ties", () => {
  const marriages = [
    { Id: "a", Children: [newUser("young", "Young", "01-01-2020"), newUser("old", "Old", "31-12-1999")] },
    { Id: "b", Children: [newUser("middle", "Middle", "01-02-2010"), newUser("unknown", "Zed"), newUser("invalid", "Amy", "31-02-2000")] },
  ];
  const before = structuredClone(marriages);
  assert.deepEqual(sortedChildren(marriages).map(({ user }) => user.id), ["old", "middle", "young", "invalid", "unknown"]);
  assert.deepEqual(marriages, before);
  assert.equal(sortedChildren(marriages)[1].familyId, "b");
  const ties = [{ Id: "a", Children: [newUser("2", "Same", "01-01-2000"), newUser("1", "Same", "01-01-2000")] }];
  assert.deepEqual(sortedChildren(ties).map(({ user }) => user.id), ["1", "2"]);
  assert.deepEqual(sortedChildren([{ Id: "missing" }, { Id: "null", Children: null }]), []);
});

test("the payload the page renders keeps parents and marriages apart", () => {
  const data = samplePersonDetails("id-1");
  // The child link is what makes the two spouses this person's parents.
  assert.deepEqual(data.ParentsMarriage.Children.map((child) => child.id), ["id-1"]);
  assert.equal(data.Marriages.length, 2);
  for (const marriage of data.Marriages) assert.equal(otherSpouses(marriage, data.Person.id).length, 1);
  // A spouse, a child and a parent each resolve from the same relationships.
  assert.equal(samplePersonDetails("id-1-spouse-1").Marriages[0].Id, "id-1-marriage-1");
  assert.equal(samplePersonDetails("id-1-child-1").ParentsMarriage.Id, "id-1-marriage-1");
  assert.equal(samplePersonDetails("id-1-child-1").Marriages.length, 0);
  assert.equal(samplePersonDetails("id-1-parent-2").Marriages[0].Children[0].id, "id-1");
  // A marriage ID is not a person.
  assert.equal(samplePersonDetails("id-1-marriage-1"), null);
});

test("isYoungerThan only rejects someone proven to be the same age or older", () => {
  const father = newUser("f", "Father", "14-03-1980");
  const mother = newUser("m", "Mother", "08-09-1983", "Female");
  assert.equal(isYoungerThan([father, mother], newUser("c1", "Older than both", "02-06-1979")), false);
  assert.equal(isYoungerThan([father, mother], newUser("c2", "Between them", "01-01-1982")), false);
  assert.equal(isYoungerThan([father, mother], newUser("c3", "Younger than both", "05-02-2007", "Female")), true);
  assert.equal(isYoungerThan([father, mother], newUser("c4", "Same age as father", "14-03-1980")), false);
  // An unknown birth date cannot disqualify a candidate on either side.
  assert.equal(isYoungerThan([father, mother], newUser("c5", "No birth date")), true);
  assert.equal(isYoungerThan([newUser("u", "Unknown parent")], newUser("c6", "Younger", "01-01-2000")), true);
});

test("personToUser maps the list model onto the wire model", () => {
  assert.deepEqual(personToUser(samplePeople[0]), {
    id: "id-1", name: "Mahammad Muhayodin", gender: "Male", alive: true, dateOfBirth: "14-03-1980", dateOfDeath: "",
  });
  // Missing optional dates become empty strings, never undefined.
  const bare = personToUser({ id: "x", name: "No dates", gender: "Female", alive: true });
  assert.equal(bare.dateOfBirth, "");
  assert.equal(bare.dateOfDeath, "");
});

test("newUser drafts start alive and empty; oppositeGender pairs the accepted genders", () => {
  assert.deepEqual(newUser(""), { id: "", name: "", dateOfBirth: "", gender: "Male", alive: true, dateOfDeath: "" });
  assert.equal(oppositeGender("Male"), "Female");
  assert.equal(oppositeGender("Female"), "Male");
});

test("family colors use stable marriage identity, not ordering", () => {
  assert.equal(familyColor("id-1-marriage-1"), familyColor("id-1-marriage-1"));
  assert.notEqual(familyColor("id-1-marriage-1"), familyColor("id-1-marriage-2"));
});

// The singles and relationships pages both order people with these, so an
// unknown birth date must not jump to the front just because the order flipped.
test("people sort by name, and by age in either direction with unknown dates last", () => {
  const oldest = newUser("a", "Ada", "01-01-1950");
  const youngest = newUser("b", "Bob", "01-01-2020");
  const unknown = newUser("c", "Cy", "");
  const people = [youngest, unknown, oldest];

  assert.deepEqual([...people].sort(compareNames).map((p) => p.id), ["a", "b", "c"]);
  assert.deepEqual([...people].sort(compareByBirthAsc).map((p) => p.id), ["a", "b", "c"]);
  assert.deepEqual([...people].sort(compareByBirthDesc).map((p) => p.id), ["b", "a", "c"]);
  // Two people with no date at all fall back to their names, not to input order.
  const [secondUnknown] = [newUser("d", "Bea", "")];
  assert.deepEqual([...people, secondUnknown].sort(compareByBirthAsc).map((p) => p.id), ["a", "b", "d", "c"]);
});
