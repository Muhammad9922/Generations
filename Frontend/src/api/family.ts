import { GetPersonDetails, otherSpouses, type User } from "../helpers/GetPersonDetails.ts";

/** One spouse together with the children they share with the requested person. */
export type SpouseWithChildren = User & { children: User[] };

/** Initial family data for a person, ready to be replaced by a single API response. */
export interface InitialFamilyData {
  parents: User[];
  spouses: SpouseWithChildren[];
}

/** Query skeleton: replace the fixture lookup with GET /people/:id/parents. */
export async function getParents(personId: string): Promise<User[]> {
  const family = await GetPersonDetails(personId);
  return family?.Parents ? [...family.Parents] : [];
}

/** Query skeleton: replace the fixture lookup with GET /people/:id/spouses. */
export async function getSpousesWithChildren(personId: string): Promise<SpouseWithChildren[]> {
  const family = await GetPersonDetails(personId);
  if (!family) return [];
  return family.Families.flatMap((marriage) => otherSpouses(marriage, personId).map((spouse) => ({
    ...spouse,
    children: [...(marriage.Chidren ?? [])],
  })));
}

/**
 * Query skeleton: replace the two local calls with GET /people/:id/family.
 * Returned shape: { parents: User[], spouses: [{ ...User, children: User[] }] }.
 */
export async function getInitialFamilyData(personId: string): Promise<InitialFamilyData> {
  const [parents, spouses] = await Promise.all([getParents(personId), getSpousesWithChildren(personId)]);
  return { parents, spouses };
}
