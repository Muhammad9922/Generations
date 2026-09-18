import assert from "node:assert/strict";
import test from "node:test";
import { createPersonActions } from "../src/helpers/PersonActions.ts";
import { samplePeople } from "./sampleFamily.mjs";

// These tests validate action construction, not browser focus or Kbar matching.
// Calling perform directly checks that each closure selects its own record.
test("sample people have unique, searchable actions and working selection", () => {
  let selected;
  const actions = createPersonActions(samplePeople, (person) => { selected = person; });
  assert.equal(actions.length, samplePeople.length);
  assert.equal(new Set(actions.map((action) => action.id)).size, samplePeople.length);
  for (const [index, action] of actions.entries()) {
    const person = samplePeople[index];
    assert.ok(action.keywords.includes(person.name));
    assert.ok(action.keywords.includes(person.id));
    assert.equal(action.subtitle, person.personFatherName ? `Father: ${person.personFatherName}` : undefined);
    action.perform();
    assert.equal(selected, person);
  }
});

// Identical names must remain distinguishable while unique names stay concise.
// Stable action IDs let us assert the correct father/name association.
test("duplicate names include fathers while unique names remain unchanged", () => {
  const actions = createPersonActions(samplePeople, () => {});
  assert.equal(actions.find((action) => action.id === "person-id-3").name, "Usman Abdul son of Abdul Rahman");
  assert.equal(actions.find((action) => action.id === "person-id-6").name, "Usman Abdul son of Abdul Rauf");
  assert.equal(actions[0].name, samplePeople[0].name);
});

// The list endpoint has no father's name to give for everyone, so an absent one
// must stay absent rather than becoming the word "undefined".
test("a person with no father's name keeps a clean label", () => {
  const [action] = createPersonActions([{ id: "id-7", name: "Fatima Noor", gender: "Female", alive: true }], () => {});
  assert.equal(action.subtitle, undefined);
  assert.equal(action.name, "Fatima Noor");
  assert.equal(action.keywords, "Fatima Noor id-7");
});

// Before loading (or for an empty dataset), no person commands should be added.
// Static navigation/theme actions are created separately in App.
test("an empty people list has no person actions", () => {
  assert.deepEqual(createPersonActions([], () => {}), []);
});
