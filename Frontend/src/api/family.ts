import { ENDPOINTS, type PersonDetailsResponse } from "./contracts.ts";
import { ApiError, request } from "./http.ts";
import { withContractLog } from "./contractLog.ts";
import type { PersonDetailsData } from "../helpers/personModel.ts";

/**
 * `GET /people/:id` — everything the details page renders: the marriage this
 * person is a child of, plus every marriage they are a spouse in. One payload,
 * so the page can open any person in the tree as the primary person.
 *
 * An ID that belongs to nobody (or to a marriage) answers 404 and resolves to
 * `null`, which the page shows as "Person not found".
 */
export async function getPersonDetails(personId: string, signal?: AbortSignal): Promise<PersonDetailsData | null> {
  return withContractLog("GET /people/:id", { personId, payload: { personId } }, async () => {
    try {
      return await request<PersonDetailsResponse>(ENDPOINTS.person(personId), { signal });
    } catch (error) {
      if (error instanceof ApiError && error.status === 404) return null;
      throw error;
    }
  }, (result) => ({ found: result !== null, marriageCount: result?.Marriages.length ?? 0 }));
}
