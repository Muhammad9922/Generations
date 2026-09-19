import { useState } from "react";
import { Button, Dialog, Flex, Select } from "@radix-ui/themes";
import type { Person } from "../api/contracts.ts";
import { dateValue, personToUser, reverseDate, toInputDate, type User } from "../helpers/personModel.ts";
import { createPersonFromDraft } from "../helpers/personDrafts.ts";

const CREATE_NEW = "__create_new__";
type ParentGender = User["gender"];
interface Draft { selection: string; name: string; birth: string; alive: boolean; death: string; }
const emptyDraft = (): Draft => ({ selection: "", name: "", birth: "", alive: true, death: "" });

interface Props { child: User; people: Person[]; close: () => void; save: (parents: [User, User]) => void | Promise<void>; }

/**
 * Selects or creates two gender-correct parents, neither of whom can be the
 * child. The caller turns the pair into the marriage the child belongs to.
 */
export default function ParentCreatorDialog({ child, people, close, save }: Props) {
  const [male, setMale] = useState<Draft>(emptyDraft);
  const [female, setFemale] = useState<Draft>(emptyDraft);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const childBirth = dateValue(child.dateOfBirth);
  const choices = (gender: ParentGender) => people.map(personToUser).filter((person) => person.id !== child.id && person.gender === gender && (childBirth === null || dateValue(person.dateOfBirth) === null || dateValue(person.dateOfBirth)! < childBirth));
  async function resolve(draft: Draft, gender: ParentGender): Promise<User> {
    const selected = choices(gender).find((person) => person.id === draft.selection);
    if (selected) return selected;
    if (draft.selection !== CREATE_NEW) throw new Error(`Select or create the ${gender.toLowerCase()} parent.`);
    const birth = dateValue(draft.birth);
    if (draft.birth && birth === null) throw new Error("Enter a valid parent birth date.");
    if (birth !== null && childBirth !== null && birth >= childBirth) throw new Error("A parent's birth date must be before the child's birth date.");
    return createPersonFromDraft(draft, gender, `${gender.toLowerCase()} parent`);
  }
  async function submit() {
    try { setSaving(true); const parents = await Promise.all([resolve(male, "Male"), resolve(female, "Female")]); await save([parents[0], parents[1]]); close(); }
    catch (cause) { setError(cause instanceof Error ? cause.message : "Could not add parents."); }
    finally { setSaving(false); }
  }
  const form = (label: string, gender: ParentGender, draft: Draft, setDraft: (draft: Draft) => void) => <fieldset><legend>{label}</legend><label>Person<Select.Root value={draft.selection} onValueChange={(selection) => setDraft({ ...draft, selection })}><Select.Trigger placeholder={`Select ${label.toLowerCase()}`} /><Select.Content>{choices(gender).length > 0 && <Select.Group><Select.Label>Eligible people</Select.Label>{choices(gender).map((person) => <Select.Item key={person.id} value={person.id}>{person.name}</Select.Item>)}</Select.Group>}<Select.Separator /><Select.Item value={CREATE_NEW}>Create new person</Select.Item></Select.Content></Select.Root></label>{draft.selection === CREATE_NEW && <><label>Name<input required value={draft.name} onChange={(event) => setDraft({ ...draft, name: event.target.value })} /></label><label>Gender<input value={gender} disabled /></label><label>Date of birth (optional)<input type="date" value={toInputDate(draft.birth)} onChange={(event) => setDraft({ ...draft, birth: reverseDate(event.target.value) })} /></label><label className="family-check"><input type="checkbox" checked={draft.alive} onChange={(event) => setDraft({ ...draft, alive: event.target.checked, death: event.target.checked ? "" : draft.death })} />Alive</label><label>Date of death (optional)<input type="date" value={toInputDate(draft.death)} onChange={(event) => { const death = reverseDate(event.target.value); setDraft({ ...draft, death, alive: death ? false : draft.alive }); }} /></label></>}</fieldset>;
  return <Dialog.Root open onOpenChange={(open) => { if (!open) close(); }}><Dialog.Content maxWidth="620px" className="family-dialog"><span className="family-eyebrow">FAMILY CONNECTION</span><Dialog.Title>Add parents</Dialog.Title><Dialog.Description size="2" mb="4">Select existing people or create parents with a birth date before {child.name}'s when known. Their marriage is created with {child.name} as its child.</Dialog.Description><form className="family-form" onSubmit={(event) => { event.preventDefault(); void submit(); }}>{form("Father", "Male", male, setMale)}{form("Mother", "Female", female, setFemale)}{error && <p role="alert" className="family-error">{error}</p>}<Flex justify="end" gap="3" mt="4"><Button type="button" variant="soft" color="gray" onClick={close}>Cancel</Button><Button type="submit" disabled={!male.selection || !female.selection || saving}>{saving ? "Saving…" : "Add parents"}</Button></Flex></form></Dialog.Content></Dialog.Root>;
}
