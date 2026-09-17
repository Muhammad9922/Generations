import { dateValue } from "../helpers/GetPersonDetails.ts";
import { createLocalId } from "./id.ts";
import { personForMarriageValidation } from "./people.ts";

export interface CreateMarriageParams {
  spouseOneId: string;
  spouseTwoId: string;
  dateOfMarriage?: string | null;
  dateOfMarriageEnd?: string | null;
}
export interface UpdateMarriageParams extends Partial<Omit<CreateMarriageParams, "spouseOneId" | "spouseTwoId">> { id: string; }

const marriages = new Map<string, CreateMarriageParams>();
function validateMarriage(params: CreateMarriageParams): void {
  if (params.spouseOneId === params.spouseTwoId) throw new Error("A marriage requires two distinct people.");
  const start = params.dateOfMarriage ? dateValue(params.dateOfMarriage) : null;
  const end = params.dateOfMarriageEnd ? dateValue(params.dateOfMarriageEnd) : null;
  if ((params.dateOfMarriage && start === null) || (params.dateOfMarriageEnd && end === null)) throw new Error("Enter valid marriage dates.");
  if (start !== null && end !== null && end < start) throw new Error("Marriage end cannot be before its start.");
  for (const spouseId of [params.spouseOneId, params.spouseTwoId]) {
    const spouse = personForMarriageValidation(spouseId);
    const birth = dateValue(spouse?.dateOfBirth);
    const death = dateValue(spouse?.dateOfDeath);
    if (start !== null && birth !== null && start <= birth) throw new Error("Marriage date must be after both spouses' birth dates.");
    if (end !== null && death !== null && end > death) throw new Error("Marriage end must be on or before both spouses' death dates.");
  }
}

/** Creation skeleton: replace the local map write with POST /marriages. */
export async function createMarriage(params: CreateMarriageParams): Promise<string> {
  validateMarriage(params);
  const id = createLocalId("marriage");
  marriages.set(id, { ...params });
  return id;
}

/** Update skeleton: replace the local map write with PATCH /marriages/:id. */
export async function updateMarriage(params: UpdateMarriageParams): Promise<void> {
  const current = marriages.get(params.id);
  if (!current) throw new Error("Marriage not found.");
  const next = { ...current, ...params };
  validateMarriage(next);
  marriages.set(params.id, next);
}
