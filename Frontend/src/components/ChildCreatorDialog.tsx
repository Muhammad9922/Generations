import { useMemo, useState, type Dispatch, type SetStateAction } from "react";
import { Button, Dialog, Flex, Select } from "@radix-ui/themes";
import { addChild, createMarriage } from "../api/marriages.ts";
import type { Person } from "../api/contracts.ts";
import { ageLabel, isYoungerThan, newUser, oppositeGender, personToUser, reverseDate, toInputDate, type FamilyCertificate, type User } from "../helpers/personModel.ts";
import { createPersonFromDraft } from "../helpers/personDrafts.ts";

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
  const existingSpouses = useMemo(() => marriages.flatMap((marriage) => marriage.Spouse ?? []).filter((user) => user.Id !== primary.Id), [marriages, primary.Id]);
  const selectablePeople = useMemo(() => people.map(personToUser).filter((user) => user.Id !== primary.Id && user.Gender === oppositeGender(primary.Gender) && !existingSpouses.some((spouse) => spouse.Id === user.Id)), [people, primary, existingSpouses]);
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
  const selectedParent = [...existingSpouses, ...selectablePeople].find((person) => person.Id === parentId);
  /** The marriage the child joins, once the other parent is known to be a spouse. */
  const targetMarriage = useMemo(() => selectedParent ? marriages.find((marriage) => (marriage.Spouse ?? []).some((person) => person.Id === selectedParent.Id)) : undefined, [marriages, selectedParent]);
  /**
   * A child cannot be either spouse, anyone already married to this person, or
   * anyone already linked as a child of these marriages, and must be younger
   * than both spouses wherever their birth dates are known.
   */
  const childChoices = useMemo(() => {
    if (!selectedParent) return [];
    const taken = new Set([
      primary.Id,
      selectedParent.Id,
      ...existingSpouses.map((spouse) => spouse.Id),
      ...marriages.flatMap((marriage) => (marriage.Chidren ?? []).map((child) => child.Id)),
    ]);
    return people.map(personToUser).filter((person) => !taken.has(person.Id) && isYoungerThan([primary, selectedParent], person));
  }, [people, primary, selectedParent, existingSpouses, marriages]);
  const creatingChild = childSelection === CREATE_NEW;
  const selectedChild = creatingChild ? undefined : childChoices.find((person) => person.Id === childSelection);
  const childGender = creatingChild ? child.Gender : selectedChild?.Gender;
  // Cheap enough to derive each render, and the gender only exists once a child
  // has been picked, so there is nothing worth memoizing here.
  const childSpouseOptions = childGender ? people.map(personToUser).filter((person) => person.Id !== primary.Id && person.Id !== parentId && person.Gender === oppositeGender(childGender)) : [];
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
        childUser = await createPersonFromDraft({ name: child.Name, birth: child.DateOfBirth, alive: child.Alive, death: child.DeateOfDeath }, child.Gender, "child");
      } else {
        if (!selectedChild) throw new Error("Select an existing person or create a new one.");
        childUser = selectedChild;
      }
      // A spouse chosen for the child creates the child's own marriage. That
      // marriage belongs to the child's payload, never to this person's.
      if (spouseMode === "existing") {
        const spouse = childSpouseOptions.find((person) => person.Id === existingChildSpouseId);
        if (!spouse) throw new Error("Select the child's spouse.");
        await createMarriage({ SpouseOne: childUser.Id, SpouseTwo: spouse.Id, DateStart: childMarriageStart, DateEnd: childMarriageEnd });
      } else if (spouseMode === "new") {
        const spouse = await createPersonFromDraft({ name: newChildSpouse.Name, birth: newChildSpouse.DateOfBirth, alive: newChildSpouse.Alive, death: newChildSpouse.DeateOfDeath }, oppositeGender(childUser.Gender), "new spouse");
        await createMarriage({ SpouseOne: childUser.Id, SpouseTwo: spouse.Id, DateStart: childMarriageStart, DateEnd: childMarriageEnd });
      }
      // The other parent's marriage already exists, or is created with the child
      // as its first child; either way the child ends up linked to it.
      if (targetMarriage) await addChild(targetMarriage.Id, childUser.Id);
      else await createMarriage({ SpouseOne: primary.Id, SpouseTwo: selectedParent.Id, DateStart: parentMarriageStart, DateEnd: parentMarriageEnd, childrenIds: [childUser.Id] });
      onSaved();
      close();
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Could not add child."); }
    finally { setSaving(false); }
  }

  return <Dialog.Root open onOpenChange={(open) => { if (!open) close(); }}>
    <Dialog.Content maxWidth="620px" className="family-dialog">
      <span className="family-eyebrow">FAMILY CONNECTION</span>
      <Dialog.Title>Add child</Dialog.Title>
      <Dialog.Description size="2" mb="4">Choose {primary.Name}'s spouse, then the child: pick someone younger than both of them, or create a new person.</Dialog.Description>
      <form className="family-form" onSubmit={(event) => { event.preventDefault(); void submit(); }}>
        <label>Other parent
          <Select.Root value={parentId} onValueChange={setParentId}><Select.Trigger placeholder="Select a spouse" />
            <Select.Content>{existingSpouses.length > 0 && <Select.Group><Select.Label>Existing spouses</Select.Label>{existingSpouses.map((person) => <Select.Item key={person.Id} value={person.Id}>{person.Name}</Select.Item>)}</Select.Group>}
              {existingSpouses.length > 0 && selectablePeople.length > 0 && <Select.Separator />}
              {selectablePeople.length > 0 && <Select.Group><Select.Label>Other eligible people</Select.Label>{selectablePeople.map((person) => <Select.Item key={person.Id} value={person.Id}>{person.Name}</Select.Item>)}</Select.Group>}
            </Select.Content>
          </Select.Root>
        </label>
        {selectedParent && <fieldset><legend>Child</legend>
          <label>Child
            <Select.Root value={childSelection} onValueChange={setChildSelection}><Select.Trigger placeholder="Select a person" />
              <Select.Content>
                {childChoices.length > 0 && <Select.Group><Select.Label>Younger than both spouses</Select.Label>{childChoices.map((person) => <Select.Item key={person.Id} value={person.Id}>{person.Name}</Select.Item>)}</Select.Group>}
                <Select.Separator />
                <Select.Item value={CREATE_NEW}>Create new person</Select.Item>
              </Select.Content>
            </Select.Root>
          </label>
          {selectedChild && <small>{selectedChild.Name} · {ageLabel(selectedChild)} · born {selectedChild.DateOfBirth || "unknown"}</small>}
          {creatingChild && <>
            <label>Name<input required value={child.Name} onChange={(event) => updateChild({ Name: event.target.value })} /></label>
            <label>Gender<Select.Root value={child.Gender} onValueChange={(gender) => { const next = gender as User["Gender"]; updateChild({ Gender: next }); updateNewSpouse({ Gender: oppositeGender(next) }); }}><Select.Trigger /> <Select.Content><Select.Item value="Male">Male</Select.Item><Select.Item value="Female">Female</Select.Item></Select.Content></Select.Root></label>
            <label>Date of birth (optional)<input type="date" value={toInputDate(child.DateOfBirth)} onChange={(event) => updateChild({ DateOfBirth: reverseDate(event.target.value) })} /></label>
            <label className="family-check"><input type="checkbox" checked={child.Alive} onChange={(event) => updateChild({ Alive: event.target.checked, DeateOfDeath: event.target.checked ? "" : child.DeateOfDeath })} />Alive</label>
            <label>Date of death (optional)<input type="date" value={toInputDate(child.DeateOfDeath)} onChange={(event) => updateChild({ DeateOfDeath: reverseDate(event.target.value), Alive: event.target.value ? false : child.Alive })} /></label>
          </>}
        </fieldset>}
        {childGender && <fieldset><legend>Spouse for this child</legend>
          <label>Add a spouse for this child<Select.Root value={spouseMode} onValueChange={(value) => setSpouseMode(value as SpouseMode)}><Select.Trigger /> <Select.Content><Select.Item value="none">No spouse</Select.Item><Select.Item value="existing">Choose existing person</Select.Item><Select.Item value="new">Create new person</Select.Item></Select.Content></Select.Root></label>
          {spouseMode === "existing" && <label>Existing spouse<Select.Root value={existingChildSpouseId} onValueChange={setExistingChildSpouseId}><Select.Trigger placeholder="Select a person" /><Select.Content>{childSpouseOptions.map((person) => <Select.Item key={person.Id} value={person.Id}>{person.Name}</Select.Item>)}</Select.Content></Select.Root></label>}
          {spouseMode === "new" && <div className="child-spouse-fields"><strong>New spouse</strong><label>Gender<input value={oppositeGender(childGender)} disabled /></label><label>Name<input required value={newChildSpouse.Name} onChange={(event) => updateNewSpouse({ Name: event.target.value })} /></label><label>Date of birth (optional)<input type="date" value={toInputDate(newChildSpouse.DateOfBirth)} onChange={(event) => updateNewSpouse({ DateOfBirth: reverseDate(event.target.value) })} /></label><label className="family-check"><input type="checkbox" checked={newChildSpouse.Alive} onChange={(event) => updateNewSpouse({ Alive: event.target.checked, DeateOfDeath: event.target.checked ? "" : newChildSpouse.DeateOfDeath })} />Alive</label><label>Date of death (optional)<input type="date" value={toInputDate(newChildSpouse.DeateOfDeath)} onChange={(event) => updateNewSpouse({ DeateOfDeath: reverseDate(event.target.value), Alive: event.target.value ? false : newChildSpouse.Alive })} /></label></div>}
        </fieldset>}
        {childGender && spouseMode !== "none" && <fieldset><legend>Child's marriage dates</legend><label>Start date (optional)<input type="date" value={toInputDate(childMarriageStart)} onChange={(event) => setChildMarriageStart(reverseDate(event.target.value))} /></label><label>End date (optional)<input type="date" value={toInputDate(childMarriageEnd)} onChange={(event) => setChildMarriageEnd(reverseDate(event.target.value))} /></label></fieldset>}
        {selectedParent && !targetMarriage && <fieldset><legend>Parents' marriage dates</legend><label>Start date (optional)<input type="date" value={toInputDate(parentMarriageStart)} onChange={(event) => setParentMarriageStart(reverseDate(event.target.value))} /></label><label>End date (optional)<input type="date" value={toInputDate(parentMarriageEnd)} onChange={(event) => setParentMarriageEnd(reverseDate(event.target.value))} /></label></fieldset>}
        {error && <p role="alert" className="family-error">{error}</p>}
        <Flex justify="end" gap="3" mt="4"><Button type="button" variant="soft" color="gray" onClick={close}>Cancel</Button><Button type="submit" disabled={!selectedParent || !childSelection || saving}>{saving ? "Saving…" : creatingChild ? "Create child" : "Add child"}</Button></Flex>
      </form>
    </Dialog.Content>
  </Dialog.Root>;
}
