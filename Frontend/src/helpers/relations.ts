import { dateValue, type PersonDetailsData, type User } from "./personModel.ts";

/**
 * The extended family of one person, derived from payloads the API already
 * returns. There is no relationship endpoint: `GET /people/:id` gives a person's
 * parents' marriage (which lists their siblings) and their own marriages, so
 * every other relation is one or two payloads away. This module holds only the
 * classification — no requests, no state — so it can be tested directly.
 */
export type RelationKind =
  | "parent" | "spouse" | "child" | "sibling" | "halfSibling"
  | "grandparent" | "uncle" | "aunt" | "niece" | "nephew" | "cousin";

export interface Relative {
  user: User;
  kind: RelationKind;
  /** The relative the link runs through, when that tells the two sides apart. */
  via?: string;
}

/** Display order, closest relation first, and the precedence used for deduping. */
export const RELATION_ORDER: RelationKind[] = [
  "parent", "spouse", "child", "sibling", "halfSibling",
  "grandparent", "uncle", "aunt", "niece", "nephew", "cousin",
];

export const RELATION_LABELS: Record<RelationKind, string> = {
  parent: "Parent",
  spouse: "Spouse",
  child: "Child",
  sibling: "Sibling",
  halfSibling: "Half-sibling",
  grandparent: "Grandparent",
  uncle: "Uncle",
  aunt: "Aunt",
  niece: "Niece",
  nephew: "Nephew",
  cousin: "Cousin",
};

/** Group headings need the plural, and three of these are irregular. */
export const RELATION_PLURALS: Record<RelationKind, string> = {
  parent: "Parents",
  spouse: "Spouses",
  child: "Children",
  sibling: "Siblings",
  halfSibling: "Half-siblings",
  grandparent: "Grandparents",
  uncle: "Uncles",
  aunt: "Aunts",
  niece: "Nieces",
  nephew: "Nephews",
  cousin: "Cousins",
};

const unique = (ids: string[]) => [...new Set(ids)];

/**
 * The payloads needed before anything beyond the first degree can be named: a
 * parent's own parents' marriage is what holds this person's grandparents,
 * uncles and aunts.
 */
export function parentIds(root: PersonDetailsData): string[] {
  const rootId = root.Person.id;
  return unique((root.ParentsMarriage?.Spouse ?? []).map((person) => person.id).filter((id) => id !== rootId));
}

/**
 * The payloads needed once the parents are known: each sibling's marriages hold
 * this person's nieces and nephews, and each uncle's or aunt's marriages hold
 * their cousins. Half-siblings come free with the parents already fetched.
 */
export function secondStepIds(root: PersonDetailsData, parents: (PersonDetailsData | null)[]): string[] {
  const rootId = root.Person.id;
  const ids = new Set<string>();
  for (const child of root.ParentsMarriage?.Children ?? []) if (child.id !== rootId) ids.add(child.id);
  for (const parent of parents) {
    if (!parent) continue;
    const parentId = parent.Person.id;
    for (const sibling of parent.ParentsMarriage?.Children ?? []) if (sibling.id !== parentId) ids.add(sibling.id);
  }
  return [...ids].filter((id) => id !== rootId);
}

/**
 * Names every relative the fetched payloads can prove, closest relation first.
 *
 * The database is not perfectly consistent — a marriage has been seen listing
 * one of its own children as a spouse — so a person is only ever classified
 * once, by their closest relation to the root, and the root is never listed as
 * their own relative.
 */
export function buildRelatives(root: PersonDetailsData, fetched: Map<string, PersonDetailsData>): Relative[] {
  const rootId = root.Person.id;
  const found = new Map<string, Relative>();

  const add = (person: User | null | undefined, kind: RelationKind, via?: string) => {
    if (!person || person.id === rootId) return;
    const existing = found.get(person.id);
    // A closer relation already claimed this person, so keep that one.
    if (existing && RELATION_ORDER.indexOf(existing.kind) <= RELATION_ORDER.indexOf(kind)) return;
    found.set(person.id, { user: person, kind, via });
  };

  const parents = (root.ParentsMarriage?.Spouse ?? []).filter((person) => person.id !== rootId);
  for (const parent of parents) add(parent, "parent");
  for (const sibling of root.ParentsMarriage?.Children ?? []) add(sibling, "sibling");
  for (const marriage of root.Marriages ?? []) {
    for (const spouse of marriage.Spouse ?? []) add(spouse, "spouse");
    for (const child of marriage.Children ?? []) add(child, "child");
  }

  for (const parent of parents) {
    const parentData = fetched.get(parent.id);
    if (!parentData) continue;
    const grandparents = parentData.ParentsMarriage;
    for (const grandparent of grandparents?.Spouse ?? []) add(grandparent, "grandparent", parent.name);
    // A parent's siblings are this person's uncles and aunts.
    for (const sibling of grandparents?.Children ?? []) {
      if (sibling.id === parent.id) continue;
      add(sibling, sibling.gender === "Male" ? "uncle" : "aunt", parent.name);
    }
    // Children of the parent's other marriages are half-siblings. The children
    // of the root's own parents' marriage are already closer as siblings.
    for (const marriage of parentData.Marriages ?? []) {
      for (const child of marriage.Children ?? []) add(child, "halfSibling", parent.name);
    }
  }

  // An uncle's or aunt's children are cousins; a sibling's children are
  // nieces and nephews. Both need the relative's own payload.
  for (const relative of [...found.values()]) {
    const data = fetched.get(relative.user.id);
    if (!data) continue;
    const isCousinSource = relative.kind === "uncle" || relative.kind === "aunt";
    const isNieceSource = relative.kind === "sibling" || relative.kind === "halfSibling";
    if (!isCousinSource && !isNieceSource) continue;
    for (const marriage of data.Marriages ?? []) {
      for (const child of marriage.Children ?? []) {
        if (isCousinSource) add(child, "cousin", relative.user.name);
        else add(child, child.gender === "Male" ? "nephew" : "niece", relative.user.name);
      }
    }
  }

  return [...found.values()].sort(compareRelatives);
}

/** Closest relation first; within a relation, oldest first, then by name. */
export function compareRelatives(a: Relative, b: Relative): number {
  const kind = RELATION_ORDER.indexOf(a.kind) - RELATION_ORDER.indexOf(b.kind);
  if (kind !== 0) return kind;
  const first = dateValue(a.user.dateOfBirth) ?? Infinity;
  const second = dateValue(b.user.dateOfBirth) ?? Infinity;
  if (first !== second) return first < second ? -1 : 1;
  return a.user.name.localeCompare(b.user.name) || a.user.id.localeCompare(b.user.id);
}

/** How the page orders the relatives it has already filtered. */
export type RelationSort = "closest" | "name" | "oldest" | "youngest";

export function sortRelatives(relatives: Relative[], sort: RelationSort): Relative[] {
  const byName = (a: Relative, b: Relative) => a.user.name.localeCompare(b.user.name) || a.user.id.localeCompare(b.user.id);
  const byBirth = (a: Relative, b: Relative) => (dateValue(a.user.dateOfBirth) ?? Infinity) - (dateValue(b.user.dateOfBirth) ?? Infinity) || byName(a, b);
  const sorted = [...relatives];
  if (sort === "name") sorted.sort(byName);
  else if (sort === "oldest") sorted.sort(byBirth);
  // Youngest first is oldest-first reversed, but an unknown birth date has to
  // stay last in both directions rather than leading the list.
  else if (sort === "youngest") sorted.sort((a, b) => {
    const first = dateValue(a.user.dateOfBirth);
    const second = dateValue(b.user.dateOfBirth);
    if (first === null || second === null) return first === second ? byName(a, b) : first === null ? 1 : -1;
    return second - first || byName(a, b);
  });
  else sorted.sort(compareRelatives);
  return sorted;
}
