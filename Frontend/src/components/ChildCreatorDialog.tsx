import { useMemo, useState, type Dispatch, type SetStateAction } from "react";
import { Button, Dialog, Flex, Select } from "@radix-ui/themes";
import { addChild, createMarriage } from "../api/marriages.ts";
import type { Person } from "../api/contracts.ts";
import { ageLabel, isYoungerThan, newUser, oppositeGender, personToUser, reverseDate, toInputDate, type FamilyCertificate, type User } from "../helpers/personModel.ts";
import { createPersonFromDraft } from "../helpers/personDrafts.ts";
import PersonPicker from "./PersonPicker";

type SpouseMode = "none" | "existing" | "new";
const CREATE_NEW = "__create_new__";

interface Props {
  primary: User;
  marriages: FamilyCertificate[];
  people: Person[];
  close: () => void;
  /** Called after the writes succeed so the page can re-read itself. */
  onSaved: () => void;
}

/** Links a child to a marriage: pick an existing younger person or create one. */
export default function ChildCreatorDialog({ primary, marriages, people, close, onSaved }: Props) {
  const existingSpouses = useMemo(() => marriages.flatMap((marriage) => marriage.Spouse ?? []).filter((user) => user.id !== primary.id), [marriages, primary.id]);
  const selectablePeople = useMemo(() => people.map(personToUser).filter((user) => user.id !== primary.id && user.gender === oppositeGender(primary.gender) && !existingSpouses.some((spouse) => spouse.id === user.id)), [people, primary, existingSpouses]);
  const [parentId, setParentId] = useState("");
  const [childSelection, setChildSelection] = useState("");
  const [child, setChild] = useState<User>(newUser(""));
  const [spouseMode, setSpouseMode] = useState<SpouseMode>("none");
  const [existingChildSpouseId, setExistingChildSpouseId] = useState("");
  const [newChildSpouse, setNewChildSpouse] = useState<User>(newUser("", "", "", "Female"));
  const [parentMarriageStart, setParentMarriageStart] = useState("");
  const [parentMarriageEnd, setParentMarriageEnd] = useState("");
  const [childMarriageStart, setChildMarriageStart] = useState("");
  const [childMarriageEnd, setChildMarriageEnd] = useState("");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const selectedParent = [...existingSpouses, ...selectablePeople].find((person) => person.id === parentId);
  /** The marriage the child joins, once the other parent is known to be a spouse. */
  const targetMarriage = useMemo(() => selectedParent ? marriages.find((marriage) => (marriage.Spouse ?? []).some((person) => person.id === selectedParent.id)) : undefined, [marriages, selectedParent]);
  /**
   * A child cannot be either spouse, anyone already married to this person, or
   * anyone already linked as a child of these marriages, and must be younger
   * than both spouses wherever their birth dates are known.
   */
  const childChoices = useMemo(() => {
    if (!selectedParent) return [];
    const taken = new Set([
      primary.id,
      selectedParent.id,
      ...existingSpouses.map((spouse) => spouse.id),
      ...marriages.flatMap((marriage) => (marriage.Children ?? []).map((child) => child.id)),
    ]);
    return people.map(personToUser).filter((person) => !taken.has(person.id) && isYoungerThan([primary, selectedParent], person));
  }, [people, primary, selectedParent, existingSpouses, marriages]);
  const creatingChild = childSelection === CREATE_NEW;
  const selectedChild = creatingChild ? undefined : childChoices.find((person) => person.id === childSelection);
  const childGender = creatingChild ? child.gender : selectedChild?.gender;
  // Cheap enough to derive each render, and the gender only exists once a child
  // has been picked, so there is nothing worth memoizing here.
  const childSpouseOptions = childGender ? people.map(personToUser).filter((person) => person.id !== primary.id && person.id !== parentId && person.gender === oppositeGender(childGender)) : [];
  /** One patcher for both person drafts in this dialog. */
  const patchDraft = (set: Dispatch<SetStateAction<User>>) => (changes: Partial<User>) => set((current) => ({ ...current, ...changes }));
  const updateChild = patchDraft(setChild);
  const updateNewSpouse = patchDraft(setNewChildSpouse);

  async function submit() {
    try {
      if (!selectedParent) throw new Error("Select the child's other parent first.");
      if (!childSelection) throw new Error("Select an existing person or create a new one.");
      setSaving(true);
      let childUser: User;
      if (creatingChild) {
        childUser = await createPersonFromDraft({ name: child.name, birth: child.dateOfBirth, alive: child.alive, death: child.dateOfDeath }, child.gender, "child");
      } else {
        if (!selectedChild) throw new Error("Select an existing person or create a new one.");
        childUser = selectedChild;
      }
      // A spouse chosen for the child creates the child's own marriage. That
      // marriage belongs to the child's payload, never to this person's.
      if (spouseMode === "existing") {
        const spouse = childSpouseOptions.find((person) => person.id === existingChildSpouseId);
        if (!spouse) throw new Error("Select the child's spouse.");
        await createMarriage({ SpouseOne: childUser.id, SpouseTwo: spouse.id, DateStart: childMarriageStart, DateEnd: childMarriageEnd });
      } else if (spouseMode === "new") {
        const spouse = await createPersonFromDraft({ name: newChildSpouse.name, birth: newChildSpouse.dateOfBirth, alive: newChildSpouse.alive, death: newChildSpouse.dateOfDeath }, oppositeGender(childUser.gender), "new spouse");
        await createMarriage({ SpouseOne: childUser.id, SpouseTwo: spouse.id, DateStart: childMarriageStart, DateEnd: childMarriageEnd });
      }
      // The other parent's marriage already exists, or is created with the child
      // as its first child; either way the child ends up linked to it.
      if (targetMarriage) await addChild(targetMarriage.Id, childUser.id);
      else await createMarriage({ SpouseOne: primary.id, SpouseTwo: selectedParent.id, DateStart: parentMarriageStart, DateEnd: parentMarriageEnd, childrenIds: [childUser.id] });
      onSaved();
      close();
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Could not add child."); }
    finally { setSaving(false); }
  }

  return <Dialog.Root open onOpenChange={(open) => { if (!open) close(); }}>
    <Dialog.Content maxWidth="620px" className="family-dialog">
      <span className="family-eyebrow">FAMILY CONNECTION</span>
      <Dialog.Title>Add child</Dialog.Title>
      <Dialog.Description size="2" mb="4">Choose {primary.name}'s spouse, then the child: pick someone younger than both of them, or create a new person.</Dialog.Description>
      <form className="family-form" onSubmit={(event) => { event.preventDefault(); void submit(); }}>
        <label>Other parent
          <PersonPicker label="Other parent" value={parentId} onChange={setParentId} placeholder="Select a spouse"
            groups={[{ label: "Existing spouses", people: existingSpouses }, { label: "Other eligible people", people: selectablePeople }]} />
        </label>
        {selectedParent && <fieldset><legend>Child</legend>
          <label>Child
            <PersonPicker label="Child" value={childSelection} onChange={setChildSelection} placeholder="Select a person"
              groups={[{ label: "Younger than both spouses", people: childChoices }]} extras={[{ value: CREATE_NEW, label: "Create new person" }]} />
          </label>
          {selectedChild && <small>{selectedChild.name} · {ageLabel(selectedChild)} · born {selectedChild.dateOfBirth || "unknown"}</small>}
          {creatingChild && <>
            <label>Name<input required value={child.name} onChange={(event) => updateChild({ name: event.target.value })} /></label>
            <label>Gender<Select.Root value={child.gender} onValueChange={(gender) => { const next = gender as User["gender"]; updateChild({ gender: next }); updateNewSpouse({ gender: oppositeGender(next) }); }}><Select.Trigger /> <Select.Content><Select.Item value="Male">Male</Select.Item><Select.Item value="Female">Female</Select.Item></Select.Content></Select.Root></label>
            <label>Date of birth (optional)<input type="date" value={toInputDate(child.dateOfBirth)} onChange={(event) => updateChild({ dateOfBirth: reverseDate(event.target.value) })} /></label>
            <label className="family-check"><input type="checkbox" checked={child.alive} onChange={(event) => updateChild({ alive: event.target.checked, dateOfDeath: event.target.checked ? "" : child.dateOfDeath })} />Alive</label>
            <label>Date of death (optional)<input type="date" value={toInputDate(child.dateOfDeath)} onChange={(event) => updateChild({ dateOfDeath: reverseDate(event.target.value), alive: event.target.value ? false : child.alive })} /></label>
          </>}
        </fieldset>}
        {childGender && <fieldset><legend>Spouse for this child</legend>
          <label>Add a spouse for this child<Select.Root value={spouseMode} onValueChange={(value) => setSpouseMode(value as SpouseMode)}><Select.Trigger /> <Select.Content><Select.Item value="none">No spouse</Select.Item><Select.Item value="existing">Choose existing person</Select.Item><Select.Item value="new">Create new person</Select.Item></Select.Content></Select.Root></label>
          {spouseMode === "existing" && <label>Existing spouse<PersonPicker label="Existing spouse" value={existingChildSpouseId} onChange={setExistingChildSpouseId} placeholder="Select a person" groups={[{ people: childSpouseOptions }]} /></label>}
          {spouseMode === "new" && <div className="child-spouse-fields"><strong>New spouse</strong><label>Gender<input value={oppositeGender(childGender)} disabled /></label><label>Name<input required value={newChildSpouse.name} onChange={(event) => updateNewSpouse({ name: event.target.value })} /></label><label>Date of birth (optional)<input type="date" value={toInputDate(newChildSpouse.dateOfBirth)} onChange={(event) => updateNewSpouse({ dateOfBirth: reverseDate(event.target.value) })} /></label><label className="family-check"><input type="checkbox" checked={newChildSpouse.alive} onChange={(event) => updateNewSpouse({ alive: event.target.checked, dateOfDeath: event.target.checked ? "" : newChildSpouse.dateOfDeath })} />Alive</label><label>Date of death (optional)<input type="date" value={toInputDate(newChildSpouse.dateOfDeath)} onChange={(event) => updateNewSpouse({ dateOfDeath: reverseDate(event.target.value), alive: event.target.value ? false : newChildSpouse.alive })} /></label></div>}
        </fieldset>}
        {childGender && spouseMode !== "none" && <fieldset><legend>Child's marriage dates</legend><label>Start date (optional)<input type="date" value={toInputDate(childMarriageStart)} onChange={(event) => setChildMarriageStart(reverseDate(event.target.value))} /></label><label>End date (optional)<input type="date" value={toInputDate(childMarriageEnd)} onChange={(event) => setChildMarriageEnd(reverseDate(event.target.value))} /></label></fieldset>}
        {selectedParent && !targetMarriage && <fieldset><legend>Parents' marriage dates</legend><label>Start date (optional)<input type="date" value={toInputDate(parentMarriageStart)} onChange={(event) => setParentMarriageStart(reverseDate(event.target.value))} /></label><label>End date (optional)<input type="date" value={toInputDate(parentMarriageEnd)} onChange={(event) => setParentMarriageEnd(reverseDate(event.target.value))} /></label></fieldset>}
        {error && <p role="alert" className="family-error">{error}</p>}
        <Flex justify="end" gap="3" mt="4"><Button type="button" variant="soft" color="gray" onClick={close}>Cancel</Button><Button type="submit" disabled={!selectedParent || !childSelection || saving}>{saving ? "Saving…" : creatingChild ? "Create child" : "Add child"}</Button></Flex>
      </form>
    </Dialog.Content>
  </Dialog.Root>;
}
