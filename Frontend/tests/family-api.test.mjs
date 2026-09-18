import assert from "node:assert/strict";
import test from "node:test";
import { getInitialFamilyData, getParents, getSpousesWithChildren } from "../src/api/family.ts";

test("initial family data exposes parents and spouses with their children", async () => {
  const family = await getInitialFamilyData("id-1");
  assert.equal(family.parents.length, 2);
  assert.equal(family.spouses.length, 2);
  assert.deepEqual(family.spouses.map((spouse) => spouse.children.length), [2, 2]);
  assert.equal((await getParents("id-2")).length, 0);
  assert.equal((await getSpousesWithChildren("unknown")).length, 0);
});
