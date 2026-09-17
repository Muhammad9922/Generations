import assert from "node:assert/strict";
import test from "node:test";
import {
  GetPersonDetails, ageLabel, dateValue, deleteMarriage, familyColor,
  fromInputDate, newUser, otherSpouses, replaceUser, sortedChildren, toInputDate, withParents,
} from "../src/helpers/GetPersonDetails.ts";

test("ID lookup supplies complete people and separate marriage certificates", async () => {
  const data = await GetPersonDetails("id-1");
  assert.equal(data.Person.Id, "id-1");
  assert.equal(data.Parents.length, 2);
  assert.equal(data.Families.length, 2);
  for (const family of data.Families) {
    assert.equal(otherSpouses(family, data.Person.Id).length, 1);
    for (const user of [...family.Spouse, ...family.Chidren, ...data.Parents]) {
      for (const key of ["Id", "Name", "DateOfBirth", "DeateOfDeath", "Gender", "Alive"]) assert.ok(key in user);
    }
  }
  assert.equal(await GetPersonDetails("unknown"), null);
  data.Person.Name = "Changed";
  assert.notEqual((await GetPersonDetails("id-1")).Person.Name, "Changed");
});

test("no-parent, no-spouse and null-child fixtures are supported", async () => {
  const single = await GetPersonDetails("id-2");
  assert.equal(single.Parents, null);
  assert.deepEqual(single.Families, []);
  assert.deepEqual(sortedChildren((await GetPersonDetails("id-3")).Families), []);
  assert.deepEqual(sortedChildren([{ Id: "missing" }, { Id: "null", Chidren: null }]), []);
  assert.deepEqual(otherSpouses({ Id: "none", Spouse: null }, "person"), []);
});

test("children sort globally across families, with invalid dates last and stable ties", () => {
  const families = [
    { Id: "a", Chidren: [newUser("young", "Young", "01-01-2020"), newUser("old", "Old", "31-12-1999")] },
    { Id: "b", Chidren: [newUser("middle", "Middle", "01-02-2010"), newUser("unknown", "Zed"), newUser("invalid", "Amy", "31-02-2000")] },
  ];
  const before = structuredClone(families);
  assert.deepEqual(sortedChildren(families).map(({ user }) => user.Id), ["old", "middle", "young", "invalid", "unknown"]);
  assert.deepEqual(families, before);
  assert.equal(sortedChildren(families)[1].familyId, "b");
  const ties = [{ Id: "a", Chidren: [newUser("2", "Same", "01-01-2000"), newUser("1", "Same", "01-01-2000")] }];
  assert.deepEqual(sortedChildren(ties).map(({ user }) => user.Id), ["1", "2"]);
});

test("dates validate leap years and pre-1970 values; age respects birthdays and death", () => {
  assert.notEqual(dateValue("29-02-2000"), null);
  assert.equal(dateValue("29-02-1900"), null);
  assert.equal(dateValue("31-04-2000"), null);
  assert.equal(dateValue("01-01-0000"), null);
  assert.ok(dateValue("01-01-1900") < 0);
  assert.equal(fromInputDate(toInputDate("03-02-1900")), "03-02-1900");
  assert.equal(toInputDate(null), "");
  const user = newUser("p", "Person", "18-09-2000");
  assert.equal(ageLabel(user, new Date(2026, 8, 17)), "Age 25");
  assert.equal(ageLabel(user, new Date(2026, 8, 18)), "Age 26");
  assert.equal(ageLabel({ ...user, Alive: false, DeateOfDeath: "17-09-2020" }), "Died aged 19");
});

test("updates replace every matching ID and birth-date changes reorder children", async () => {
  const data = await GetPersonDetails("id-1");
  const renamed = replaceUser(data, { ...data.Person, Name: "New name" });
  assert.equal(renamed.Person.Name, "New name");
  assert.ok(renamed.Families.every((family) => family.Spouse[0].Name === "New name"));
  assert.notEqual(data.Person.Name, "New name");
  const child = sortedChildren(data.Families).at(-1).user;
  const updated = replaceUser(data, { ...child, DateOfBirth: "01-01-2006" });
  assert.equal(sortedChildren(updated.Families)[0].user.Id, child.Id);
});

test("parents require two distinct identities and deletion protects linked children", async () => {
  const data = await GetPersonDetails("id-2");
  const first = newUser("parent-1", "First");
  const second = newUser("parent-2", "Second");
  assert.throws(() => withParents(data, [first]));
  assert.throws(() => withParents(data, [first, first]));
  assert.throws(() => withParents(data, [data.Person, first]));
  assert.equal(withParents(data, [first, second]).Parents.length, 2);
  const married = await GetPersonDetails("id-1");
  assert.throws(() => deleteMarriage(married, married.Families[0].Id), /linked children/);
  const empty = await GetPersonDetails("id-3");
  assert.deepEqual(deleteMarriage(empty, empty.Families[0].Id).Families, []);
});

test("family colors use stable marriage identity, not ordering", () => {
  assert.equal(familyColor("id-1-marriage-1"), familyColor("id-1-marriage-1"));
  assert.notEqual(familyColor("id-1-marriage-1"), familyColor("id-1-marriage-2"));
});
