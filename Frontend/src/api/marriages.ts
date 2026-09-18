import {
  ENDPOINTS, type AddChildRequest, type CreateMarriageRequest, type CreateMarriageResponse,
  type UpdateMarriageRequest, type WriteResponse,
} from "./contracts.ts";
import { request } from "./http.ts";
import { withContractLog } from "./contractLog.ts";

/**
 * `CreateMarriage(spouseOne, spouseTwo, startDate*, endDate*, childrenIds*)`
 * creates a marriage and resolves its ID. Used for a new spouse and for a new
 * parents' marriage, where the person is passed among `childrenIds`.
 *
 * The people in a marriage are fixed at creation — swapping a spouse would break
 * every child link hanging off it — so the only way to change them is
 * `DeleteMarriage` followed by another `CreateMarriage`.
 */
export async function createMarriage(marriage: CreateMarriageRequest, signal?: AbortSignal): Promise<string> {
  return withContractLog("CreateMarriage", {
    spouseOneId: marriage.SpouseOne,
    spouseTwoId: marriage.SpouseTwo,
    hasStartDate: Boolean(marriage.DateStart),
    hasEndDate: Boolean(marriage.DateEnd),
    childCount: marriage.childrenIds?.length ?? 0,
    payload: marriage,
  }, async () => {
    const created = await request<CreateMarriageResponse>(ENDPOINTS.marriages, { method: "POST", body: marriage, signal });
    return created.id;
  }, (id) => ({ id }));
}

/**
 * `AddChildren(marriageID, userId)` links an existing person as a child of a
 * marriage, which is how a child is added to a marriage that already exists.
 */
export async function addChild(marriageId: string, childId: string, signal?: AbortSignal): Promise<WriteResponse> {
  const body: AddChildRequest = { childId };
  return withContractLog("AddChildren", { marriageId, payload: body }, () =>
    request<WriteResponse>(ENDPOINTS.marriageChildren(marriageId), { method: "POST", body, signal }),
    (result) => ({ added: result }));
}

/**
 * `RemoveChild(marriageID, childId)` removes any child from a marriage. It also
 * backs "delete parents", which only delinks this person from their parents'
 * marriage: the two parents stay married to each other.
 */
export async function removeChild(marriageId: string, childId: string, signal?: AbortSignal): Promise<WriteResponse> {
  return withContractLog("RemoveChild", { marriageId, payload: { marriageId, childId } }, () =>
    request<WriteResponse>(ENDPOINTS.marriageChild(marriageId, childId), { method: "DELETE", signal }),
    (result) => ({ removed: result }));
}

/**
 * `DeleteMarriage(marriageID)` disbands a marriage, which is how a spouse or a
 * whole family is removed. The spouses and children remain people, but this
 * marriage's child links vanish with it, so those children are left with
 * undefined parents.
 */
export async function deleteMarriage(marriageId: string, signal?: AbortSignal): Promise<WriteResponse> {
  return withContractLog("DeleteMarriage", { marriageId, payload: { marriageId } }, () =>
    request<WriteResponse>(ENDPOINTS.marriage(marriageId), { method: "DELETE", signal }),
    (result) => ({ deleted: result }));
}

/**
 * `PATCH /marriages/:id` for dates only (`marriage.UpdateMarriageDates`). An
 * empty string clears a date; the spouses and children cannot be changed here.
 */
export async function updateMarriage(marriageId: string, dates: UpdateMarriageRequest, signal?: AbortSignal): Promise<void> {
  return withContractLog("PATCH /marriages/:id", { marriageId, fields: Object.keys(dates), payload: dates }, async () => {
    await request<unknown>(ENDPOINTS.marriage(marriageId), { method: "PATCH", body: dates, signal });
  });
}
