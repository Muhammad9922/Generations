import assert from "node:assert/strict";
import test from "node:test";
import { createPersonActions, GetAllUsers } from "../src/helpers/GetUsers.ts";

// Run with Node's built-in test runner on a version supporting TS type stripping.
// These tests validate action construction, not browser focus or Kbar matching.
// Calling perform directly checks that each closure selects its own record.
test("sample people have unique, searchable actions and working selection", async () => {
  const people = await GetAllUsers();
  let selected;
  const actions = createPersonActions(people, (person) => { selected = person; });
  assert.equal(actions.length, people.length);
  assert.equal(new Set(actions.map((action) => action.id)).size, people.length);
  for (const [index, action] of actions.entries()) {
    assert.ok(action.keywords.includes(people[index].name));
    assert.ok(action.keywords.includes(people[index].personFatherName));
    assert.equal(action.subtitle, `Father: ${people[index].personFatherName}`);
    action.perform();
    assert.equal(selected, people[index]);
  }
});

// Identical names must remain distinguishable while unique names stay concise.
// Stable action IDs let us assert the correct father/name association.
test("duplicate names include fathers while unique names remain unchanged", async () => {
  const people = await GetAllUsers();
  const actions = createPersonActions(people, () => {});
  assert.equal(actions.find((action) => action.id === "person-id-3").name, "Usman Abdul son of Abdul Rahman");
  assert.equal(actions.find((action) => action.id === "person-id-6").name, "Usman Abdul son of Abdul Rauf");
  assert.equal(actions[0].name, people[0].name);
});

// Before loading (or for an empty dataset), no person commands should be added.
// Static navigation/theme actions are created separately in App.
test("an empty people list has no person actions", () => {
  assert.deepEqual(createPersonActions([], () => {}), []);
});
