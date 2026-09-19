import assert from "node:assert/strict";
import test from "node:test";
import { newUser } from "../src/helpers/personModel.ts";
import {
  buildRelatives, parentIds, RELATION_LABELS, RELATION_ORDER, RELATION_PLURALS, secondStepIds, sortRelatives,
} from "../src/helpers/relations.ts";

const user = (id, name, dateOfBirth = "", gender = "Male") => newUser(id, name, dateOfBirth, gender);

/**
 * A three-generation family built by hand, because the shipped sample family has
 * no aunts, uncles or cousins in it:
 *
 *   Dada + Dadi -> Baba, Uncle1, Aunt1
 *   Nana + Nani -> Mama, Aunt2
 *   Baba + Mama -> Ali, Sana, Bilal          (Ali is the root)
 *   Baba + SecondWife -> HalfSib
 *   Ali + Wife -> Kid
 *   Sana + Husband -> Niece1
 *   Bilal + Wife2 -> Nephew1
 *   Uncle1 + AuntInLaw -> Cousin1, Cousin2
 *   Aunt1 + UncleInLaw -> Cousin3
 */
function family() {
  const marriage = (Id, spouse, children) => ({ Id, Spouse: spouse, Children: children, StartOfFamily: null, EndOfFamily: null });
  const ali = user("ali", "Ali Ahmed", "10-05-1990");
  const sana = user("sana", "Sana Ahmed", "02-03-1993", "Female");
  const bilal = user("bilal", "Bilal Ahmed", "07-07-1995");
  const baba = user("baba", "Baba Ahmed", "01-01-1960");
  const mama = user("mama", "Mama Begum", "04-04-1962", "Female");
  const dada = user("dada", "Dada Ahmed", "01-01-1930");
  const dadi = user("dadi", "Dadi Begum", "01-01-1935", "Female");
  const nana = user("nana", "Nana Khan", "01-01-1938");
  const nani = user("nani", "Nani Khan", "01-01-1940", "Female");
  const uncle1 = user("uncle1", "Uncle One", "01-01-1963");
  const aunt1 = user("aunt1", "Aunt One", "01-01-1965", "Female");
  const aunt2 = user("aunt2", "Aunt Two", "01-01-1968", "Female");
  const halfSib = user("half", "Half Sibling", "01-01-1975");
  const wife = user("wife", "Ali Wife", "01-01-1992", "Female");
  const kid = user("kid", "Ali Child", "01-01-2015");
  const cousin1 = user("cousin1", "Cousin One", "01-01-1990");
  const cousin2 = user("cousin2", "Cousin Two", "01-01-1992", "Female");
  const cousin3 = user("cousin3", "Cousin Three", "01-01-1995", "Female");
  const niece1 = user("niece1", "Niece One", "01-01-2018", "Female");
  const nephew1 = user("nephew1", "Nephew One", "01-01-2020");

  const root = {
    Person: ali,
    ParentsMarriage: marriage("ali-parents", [baba, mama], [ali, sana, bilal]),
    Marriages: [marriage("ali-marriage", [ali, wife], [kid])],
  };
  const fetched = new Map([
    ["baba", {
      Person: baba,
      ParentsMarriage: marriage("baba-parents", [dada, dadi], [baba, uncle1, aunt1]),
      Marriages: [marriage("ali-parents", [baba, mama], [ali, sana, bilal]), marriage("baba-second", [baba, user("second", "Second Wife", "01-01-1970", "Female")], [halfSib])],
    }],
    ["mama", {
      Person: mama,
      ParentsMarriage: marriage("mama-parents", [nana, nani], [mama, aunt2]),
      Marriages: [marriage("ali-parents", [mama, baba], [ali, sana, bilal])],
    }],
    ["uncle1", { Person: uncle1, ParentsMarriage: null, Marriages: [marriage("uncle1-marriage", [uncle1, user("inlaw1", "Aunt In Law", "", "Female")], [cousin1, cousin2])] }],
    ["aunt1", { Person: aunt1, ParentsMarriage: null, Marriages: [marriage("aunt1-marriage", [aunt1, user("inlaw2", "Uncle In Law")], [cousin3])] }],
    ["aunt2", { Person: aunt2, ParentsMarriage: null, Marriages: [] }],
    ["sana", { Person: sana, ParentsMarriage: null, Marriages: [marriage("sana-marriage", [sana, user("husband", "Sana Husband")], [niece1])] }],
    ["bilal", { Person: bilal, ParentsMarriage: null, Marriages: [marriage("bilal-marriage", [bilal, user("wife2", "Bilal Wife", "", "Female")], [nephew1])] }],
    ["half", { Person: halfSib, ParentsMarriage: null, Marriages: [] }],
  ]);
  return { root, fetched, uncle1, aunt1 };
}

const kindsOf = (relatives, kind) => relatives.filter((relative) => relative.kind === kind).map((relative) => relative.user.id).sort();

test("the family is classified into the expected relations", () => {
  const { root, fetched } = family();
  const relatives = buildRelatives(root, fetched);
  assert.deepEqual(kindsOf(relatives, "parent"), ["baba", "mama"]);
  assert.deepEqual(kindsOf(relatives, "spouse"), ["wife"]);
  assert.deepEqual(kindsOf(relatives, "child"), ["kid"]);
  assert.deepEqual(kindsOf(relatives, "sibling"), ["bilal", "sana"]);
  assert.deepEqual(kindsOf(relatives, "halfSibling"), ["half"]);
  assert.deepEqual(kindsOf(relatives, "grandparent"), ["dada", "dadi", "nana", "nani"]);
  assert.deepEqual(kindsOf(relatives, "uncle"), ["uncle1"]);
  assert.deepEqual(kindsOf(relatives, "aunt"), ["aunt1", "aunt2"]);
  assert.deepEqual(kindsOf(relatives, "cousin"), ["cousin1", "cousin2", "cousin3"]);
  assert.deepEqual(kindsOf(relatives, "niece"), ["niece1"]);
  assert.deepEqual(kindsOf(relatives, "nephew"), ["nephew1"]);
  // Nobody is invented, and the person themselves is never their own relative.
  // The 19 are blood relatives: a spouse of an uncle or sibling is a relation
  // by marriage, which this report deliberately does not claim.
  assert.equal(relatives.length, 19);
  assert.ok(!relatives.some((relative) => relative.user.id === "ali"));
  assert.ok(!relatives.some((relative) => ["inlaw1", "inlaw2", "husband", "wife2", "second"].includes(relative.user.id)));
});

test("every relative carries the label the page renders", () => {
  const { root, fetched } = family();
  for (const relative of buildRelatives(root, fetched)) {
    assert.ok(RELATION_LABELS[relative.kind], `no label for ${relative.kind}`);
    assert.ok(RELATION_PLURALS[relative.kind], `no plural for ${relative.kind}`);
    assert.ok(RELATION_ORDER.includes(relative.kind));
  }
  // Every kind the order names must be renderable, even if this family lacks it.
  for (const kind of RELATION_ORDER) {
    assert.ok(RELATION_LABELS[kind] && RELATION_PLURALS[kind], `unrenderable kind ${kind}`);
  }
});

test("relatives are named only by their closest relation to the root", () => {
  const { root, fetched } = family();
  // Messy data: the root's own child is also listed under a cousin's marriage.
  const uncle1 = fetched.get("uncle1");
  uncle1.Marriages[0].Children.push(root.Marriages[0].Children[0]);
  // And a sibling is also listed as their parent's sibling.
  fetched.get("baba").ParentsMarriage.Children.push(root.ParentsMarriage.Children[1]);
  const relatives = buildRelatives(root, fetched);
  assert.equal(kindsOf(relatives, "child").includes("kid"), true);
  assert.equal(kindsOf(relatives, "cousin").includes("kid"), false);
  assert.equal(kindsOf(relatives, "sibling").includes("sana"), true);
  assert.equal(kindsOf(relatives, "aunt").includes("sana"), false);
  assert.equal(relatives.filter((relative) => relative.user.id === "kid").length, 1);
});

test("a person listed as their own parent is not reported as their own grandparent", () => {
  // The live database contains marriages whose spouse list includes one of
  // their own children, which used to make a parent their own parent.
  const { root, fetched } = family();
  fetched.get("baba").ParentsMarriage.Spouse = [fetched.get("baba").Person, user("dadi2", "Dadi Two", "", "Female")];
  const relatives = buildRelatives(root, fetched);
  assert.equal(kindsOf(relatives, "grandparent").includes("baba"), false);
  assert.equal(relatives.find((relative) => relative.user.id === "baba").kind, "parent");
});

test("the traversal asks for parents first, then siblings and uncles", () => {
  const { root, fetched } = family();
  assert.deepEqual(parentIds(root), ["baba", "mama"]);
  const second = secondStepIds(root, [fetched.get("baba"), fetched.get("mama")]).sort();
  assert.deepEqual(second, ["aunt1", "aunt2", "bilal", "sana", "uncle1"]);
  // A person with no parents needs no second round beyond their own records.
  const orphan = { Person: user("orphan", "Orphan"), ParentsMarriage: null, Marriages: [] };
  assert.deepEqual(parentIds(orphan), []);
  assert.deepEqual(secondStepIds(orphan, []), []);
});

test("sorting keeps unknown birth dates last in both directions", () => {
  const unknown = { user: user("unknown", "Zed Unknown"), kind: "cousin" };
  const relatives = [
    { user: user("old", "Old Person", "01-01-1950"), kind: "cousin" },
    { user: user("new", "New Person", "01-01-2020"), kind: "cousin" },
    unknown,
  ];
  assert.deepEqual(sortRelatives(relatives, "oldest").map((r) => r.user.id), ["old", "new", "unknown"]);
  assert.deepEqual(sortRelatives(relatives, "youngest").map((r) => r.user.id), ["new", "old", "unknown"]);
  assert.deepEqual(sortRelatives(relatives, "name").map((r) => r.user.id), ["new", "old", "unknown"]);
  assert.deepEqual(sortRelatives(relatives, "closest").map((r) => r.user.id), ["old", "new", "unknown"]);
});

test("the closest relation sorts first, whatever people are grouped by", () => {
  const { root, fetched } = family();
  const order = buildRelatives(root, fetched).map((relative) => RELATION_ORDER.indexOf(relative.kind));
  assert.deepEqual(order, [...order].sort((a, b) => a - b));
});
