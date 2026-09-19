import { getPersonDetails } from "./family.ts";
import type { PersonDetailsData, User } from "../helpers/personModel.ts";
import { buildRelatives, parentIds, secondStepIds, type Relative } from "../helpers/relations.ts";

export interface RelationReport {
  /** The person the report is about. */
  person: User;
  relatives: Relative[];
}

/**
 * Every relative of one person.
 *
 * There is no relationship endpoint, so this walks the graph the API exposes:
 * the root payload names the parents and siblings, each parent's payload names
 * the grandparents, uncles and aunts, and each of those payloads names the
 * cousins, nieces and nephews. That is three rounds of requests, the first
 * alone and the next two in parallel, rather than one call per relative.
 *
 * An unknown ID resolves to `null`, the same way the person page reports it.
 */
export async function loadRelatives(personId: string, signal?: AbortSignal): Promise<RelationReport | null> {
  const root = await getPersonDetails(personId, signal);
  if (!root) return null;

  // Fetched payloads are shared with the classifier, so the root is never read twice.
  const fetched = new Map<string, PersonDetailsData>([[personId, root]]);

  const parents = await Promise.all(parentIds(root).map((id) => getPersonDetails(id, signal)));
  for (const parent of parents) if (parent) fetched.set(parent.Person.id, parent);

  const others = await Promise.all(secondStepIds(root, parents).map((id) => getPersonDetails(id, signal)));
  for (const other of others) if (other) fetched.set(other.Person.id, other);

  return { person: root.Person, relatives: buildRelatives(root, fetched) };
}
