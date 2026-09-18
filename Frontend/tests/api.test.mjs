import assert from "node:assert/strict";
import test from "node:test";
import { createPerson, getAllPeople, updateUserDetails } from "../src/api/people.ts";
import { addChild, createMarriage, deleteMarriage, removeChild, updateMarriage } from "../src/api/marriages.ts";
import { getPersonDetails } from "../src/api/family.ts";
import { ApiError, isAbort } from "../src/api/http.ts";
import { samplePeople, samplePersonDetails } from "./sampleFamily.mjs";

// The contract layer is the only place that talks HTTP, so these tests stub
// fetch and assert the exact request each function sends and how it maps the
// answer back. `API_BASE_URL` defaults to "/api" outside Vite.
function stubFetch(handler) {
  const calls = [];
  const original = globalThis.fetch;
  globalThis.fetch = async (url, init = {}) => {
    calls.push({ url, method: init.method ?? "GET", body: init.body ? JSON.parse(init.body) : undefined, signal: init.signal });
    return handler(url, init);
  };
  return { calls, restore: () => { globalThis.fetch = original; } };
}
const json = (payload, status = 200) => ({ ok: status >= 200 && status < 300, status, text: async () => JSON.stringify(payload) });

test("getAllPeople reads the list from GET /people", async () => {
  const stub = stubFetch(() => json({ people: samplePeople }));
  try {
    const people = await getAllPeople();
    assert.equal(people.length, 9);
    assert.deepEqual(stub.calls[0], { url: "/api/people", method: "GET", body: undefined, signal: undefined });
  } finally { stub.restore(); }
});

test("createPerson posts the CreatePerson body and resolves the saved person", async () => {
  const created = { Id: "new-1", Name: "Ada", Gender: "Female", DateOfBirth: "10-12-1815", DeateOfDeath: "", Alive: true };
  const stub = stubFetch(() => json(created, 201));
  try {
    const person = await createPerson({ PersonName: "Ada", Gender: "Female", DateOfBirth: "10-12-1815", DateOfDeath: "", Alive: true });
    // The ID comes from the API, never from the client.
    assert.equal(person.Id, "new-1");
    assert.equal(stub.calls[0].method, "POST");
    assert.equal(stub.calls[0].url, "/api/people");
    assert.deepEqual(stub.calls[0].body, { PersonName: "Ada", Gender: "Female", DateOfBirth: "10-12-1815", DateOfDeath: "", Alive: true });
  } finally { stub.restore(); }
});

test("updateUserDetails patches the profile and clears the death date of anyone alive", async () => {
  const saved = { Id: "id-1", Name: "Renamed", Gender: "Male", DateOfBirth: "01-01-1990", DeateOfDeath: "", Alive: true };
  const stub = stubFetch(() => json(saved));
  try {
    const person = await updateUserDetails({ Id: "id-1", Name: "Renamed", Gender: "Male", DateOfBirth: "01-01-1990", DeateOfDeath: "05-05-2020", Alive: true });
    assert.equal(person.Name, "Renamed");
    assert.equal(stub.calls[0].method, "PATCH");
    assert.equal(stub.calls[0].url, "/api/people/id-1");
    assert.deepEqual(stub.calls[0].body, { Name: "Renamed", Gender: "Male", Alive: true, DateOfBirth: "01-01-1990", DateOfDeath: null });
  } finally { stub.restore(); }
});

test("getPersonDetails resolves the payload, and null for an unknown person", async () => {
  const stub = stubFetch((url) => url.endsWith("/id-1") ? json(samplePersonDetails("id-1")) : json({ error: "User Does Not Exists" }, 404));
  try {
    const person = await getPersonDetails("id-1");
    assert.equal(person.Person.Id, "id-1");
    assert.equal(person.Marriages.length, 2);
    assert.equal(stub.calls[0].url, "/api/people/id-1");
    // 404 is a normal answer: the page shows "Person not found".
    assert.equal(await getPersonDetails("nobody"), null);
  } finally { stub.restore(); }
});

test("createMarriage posts both spouses, the dates and any starting children", async () => {
  const stub = stubFetch(() => json({ id: "m-1" }, 201));
  try {
    const id = await createMarriage({ SpouseOne: "a", SpouseTwo: "b", DateStart: "01-01-2005", DateEnd: "", childrenIds: ["c"] });
    assert.equal(id, "m-1");
    assert.equal(stub.calls[0].url, "/api/marriages");
    assert.deepEqual(stub.calls[0].body, { SpouseOne: "a", SpouseTwo: "b", DateStart: "01-01-2005", DateEnd: "", childrenIds: ["c"] });
  } finally { stub.restore(); }
});

test("addChild posts a child link and removeChild deletes that link", async () => {
  const stub = stubFetch(() => json(true));
  try {
    assert.equal(await addChild("m-1", "c-1"), true);
    assert.deepEqual(stub.calls[0], { url: "/api/marriages/m-1/children", method: "POST", body: { childId: "c-1" }, signal: undefined });
    assert.equal(await removeChild("m-1", "c-1"), true);
    assert.equal(stub.calls[1].url, "/api/marriages/m-1/children/c-1");
    assert.equal(stub.calls[1].method, "DELETE");
    assert.equal(stub.calls[1].body, undefined);
  } finally { stub.restore(); }
});

test("deleteMarriage disbands one marriage and updateMarriage moves only its dates", async () => {
  const stub = stubFetch(() => json(true));
  try {
    assert.equal(await deleteMarriage("m-1"), true);
    assert.deepEqual(stub.calls[0], { url: "/api/marriages/m-1", method: "DELETE", body: undefined, signal: undefined });
    await updateMarriage("m-1", { DateStart: "01-01-2010", DateEnd: "" });
    assert.equal(stub.calls[1].method, "PATCH");
    assert.deepEqual(stub.calls[1].body, { DateStart: "01-01-2010", DateEnd: "" });
  } finally { stub.restore(); }
});

test("IDs are encoded so reserved characters survive the URL", async () => {
  const stub = stubFetch(() => json(true));
  try {
    await removeChild("m/1", "c 2");
    assert.equal(stub.calls[0].url, "/api/marriages/m%2F1/children/c%202");
  } finally { stub.restore(); }
});

test("the API's own error message reaches the caller with its status", async () => {
  const stub = stubFetch(() => json({ error: "both spouses are male" }, 400));
  try {
    await assert.rejects(() => createMarriage({ SpouseOne: "a", SpouseTwo: "b", DateStart: "", DateEnd: "" }), /both spouses are male/);
    await assert.rejects(() => createMarriage({ SpouseOne: "a", SpouseTwo: "b", DateStart: "", DateEnd: "" }),
      (error) => error instanceof ApiError && error.status === 400);
  } finally { stub.restore(); }
});

test("an unreachable API and a non-JSON answer both report clearly", async () => {
  const offline = stubFetch(() => { throw new TypeError("fetch failed"); });
  try {
    await assert.rejects(() => getAllPeople(), /Could not reach the API/);
  } finally { offline.restore(); }
  const notJson = stubFetch(() => ({ ok: true, status: 200, text: async () => "Hi there!" }));
  try {
    await assert.rejects(() => getAllPeople(), /did not answer with JSON/);
  } finally { notJson.restore(); }
});

test("a cancelled request stays an abort instead of becoming an error", async () => {
  const stub = stubFetch(() => { throw new DOMException("aborted", "AbortError"); });
  try {
    await assert.rejects(() => getPersonDetails("id-1"), (error) => isAbort(error));
  } finally { stub.restore(); }
});
