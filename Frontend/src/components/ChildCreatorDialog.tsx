import { useMemo, useState } from "react";
import { Button, Dialog, Flex, Select } from "@radix-ui/themes";
import { createMarriage } from "../api/marriages.ts";
import { createPerson, registerPeopleForMarriageValidation } from "../api/people.ts";
import { fromInputDate, toInputDate, type FamilyCertificate, type User } from "../helpers/GetPersonDetails";
import type { Person } from "../helpers/GetUsers";

type SpouseMode = "none" | "existing" | "new";
const oppositeGender = (gender: User["Gender"]): User["Gender"] => gender === "Male" ? "Female" : "Male";
const blankPerson = (gender: User["Gender"]): User => ({ Id: "", Name: "", Gender: gender, Alive: true, DateOfBirth: "", DeateOfDeath: "" });
const asUser = (person: Person): User => ({ Id: person.id, Name: person.name, Gender: person.gender, Alive: person.alive, DateOfBirth: person.dateOfBirth ?? "", DeateOfDeath: person.dateOfDeath ?? "" });

interface Props {
  primary: User;
  families: FamilyCertificate[];
  people: Person[];
  close: () => void;
  save: (family: FamilyCertificate, child: User, childMarriage?: FamilyCertificate) => void;
}

/** Creates a child only after a second parent is selected; all writes use API-shaped helpers. */
export default function ChildCreatorDialog({ primary, families, people, close, save }: Props) {
  const existingSpouses = useMemo(() => families.flatMap((family) => family.Spouse ?? []).filter((user) => user.Id !== primary.Id), [families, primary.Id]);
  const selectablePeople = useMemo(() => people.map(asUser).filter((user) => user.Id !== primary.Id && user.Gender === oppositeGender(primary.Gender) && !existingSpouses.some((spouse) => spouse.Id === user.Id)), [people, primary, existingSpouses]);
  const [parentId, setParentId] = useState("");
  const [child, setChild] = useState<User>(blankPerson("Male"));
  const [spouseMode, setSpouseMode] = useState<SpouseMode>("none");
  const [existingChildSpouseId, setExistingChildSpouseId] = useState("");
  const [newChildSpouse, setNewChildSpouse] = useState<User>(blankPerson("Female"));
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const selectedParent = [...existingSpouses, ...selectablePeople].find((person) => person.Id === parentId);
  const childSpouseOptions = useMemo(() => people.map(asUser).filter((person) => person.Id !== primary.Id && person.Id !== parentId && person.Gender === oppositeGender(child.Gender)), [people, primary.Id, parentId, child.Gender]);
  const updateChild = (patch: Partial<User>) => setChild((current) => ({ ...current, ...patch }));
  const updateNewSpouse = (patch: Partial<User>) => setNewChildSpouse((current) => ({ ...current, ...patch }));

  async function submit() {
    try {
      if (!selectedParent) throw new Error("Select the child's other parent first.");
      if (!child.Name.trim()) throw new Error("The child needs a name.");
      setSaving(true);
      registerPeopleForMarriageValidation([primary, selectedParent, ...existingSpouses, ...people.map(asUser)]);
      const childId = await createPerson({ name: child.Name.trim(), gender: child.Gender, alive: child.Alive, date_of_birth: child.DateOfBirth || null, date_of_death: child.Alive ? null : child.DeateOfDeath || null });
      const createdChild = { ...child, Id: childId, Name: child.Name.trim(), DeateOfDeath: child.Alive ? "" : child.DeateOfDeath };
      let childMarriage: FamilyCertificate | undefined;
      if (spouseMode === "existing") {
        const spouse = childSpouseOptions.find((person) => person.Id === existingChildSpouseId);
        if (!spouse) throw new Error("Select the child's spouse.");
        const id = await createMarriage({ spouseOneId: childId, spouseTwoId: spouse.Id });
        childMarriage = { Id: id, Spouse: [createdChild, spouse], Chidren: [] };
      } else if (spouseMode === "new") {
        if (!newChildSpouse.Name.trim()) throw new Error("The new spouse needs a name.");
        const spouseId = await createPerson({ name: newChildSpouse.Name.trim(), gender: oppositeGender(createdChild.Gender), alive: newChildSpouse.Alive, date_of_birth: newChildSpouse.DateOfBirth || null, date_of_death: newChildSpouse.Alive ? null : newChildSpouse.DeateOfDeath || null });
        const spouse = { ...newChildSpouse, Id: spouseId, Name: newChildSpouse.Name.trim(), Gender: oppositeGender(createdChild.Gender), DeateOfDeath: newChildSpouse.Alive ? "" : newChildSpouse.DeateOfDeath };
        const id = await createMarriage({ spouseOneId: childId, spouseTwoId: spouseId });
        childMarriage = { Id: id, Spouse: [createdChild, spouse], Chidren: [] };
      }
      const family = families.find((item) => (item.Spouse ?? []).some((person) => person.Id === selectedParent.Id));
      save(family ?? { Id: await createMarriage({ spouseOneId: primary.Id, spouseTwoId: selectedParent.Id }), Spouse: [primary, selectedParent], Chidren: [] }, createdChild, childMarriage);
      close();
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Could not create child."); }
    finally { setSaving(false); }
  }

  return <Dialog.Root open onOpenChange={(open) => { if (!open) close(); }}>
    <Dialog.Content maxWidth="620px" className="family-dialog">
      <span className="family-eyebrow">FAMILY CONNECTION</span>
      <Dialog.Title>Add child</Dialog.Title>
      <Dialog.Description size="2" mb="4">Choose {primary.Name}'s spouse, then enter the child’s details.</Dialog.Description>
      <form className="family-form" onSubmit={(event) => { event.preventDefault(); void submit(); }}>
        <label>Other parent
          <Select.Root value={parentId} onValueChange={setParentId}><Select.Trigger placeholder="Select a spouse" />
            <Select.Content>{existingSpouses.length > 0 && <Select.Group><Select.Label>Existing spouses</Select.Label>{existingSpouses.map((person) => <Select.Item key={person.Id} value={person.Id}>{person.Name}</Select.Item>)}</Select.Group>}
              {existingSpouses.length > 0 && selectablePeople.length > 0 && <Select.Separator />}
              {selectablePeople.length > 0 && <Select.Group><Select.Label>Other eligible people</Select.Label>{selectablePeople.map((person) => <Select.Item key={person.Id} value={person.Id}>{person.Name}</Select.Item>)}</Select.Group>}
            </Select.Content>
          </Select.Root>
        </label>
        {selectedParent && <fieldset><legend>New child</legend>
          <label>Name<input required value={child.Name} onChange={(event) => updateChild({ Name: event.target.value })} /></label>
          <label>Gender<Select.Root value={child.Gender} onValueChange={(gender) => { const next = gender as User["Gender"]; updateChild({ Gender: next }); setNewChildSpouse((current) => ({ ...current, Gender: oppositeGender(next) })); }}><Select.Trigger /> <Select.Content><Select.Item value="Male">Male</Select.Item><Select.Item value="Female">Female</Select.Item></Select.Content></Select.Root></label>
          <label>Date of birth (optional)<input type="date" value={toInputDate(child.DateOfBirth)} onChange={(event) => updateChild({ DateOfBirth: fromInputDate(event.target.value) })} /></label>
          <label className="family-check"><input type="checkbox" checked={child.Alive} onChange={(event) => updateChild({ Alive: event.target.checked, DeateOfDeath: event.target.checked ? "" : child.DeateOfDeath })} />Alive</label>
          <label>Date of death (optional)<input type="date" value={toInputDate(child.DeateOfDeath)} onChange={(event) => updateChild({ DeateOfDeath: fromInputDate(event.target.value), Alive: event.target.value ? false : child.Alive })} /></label>
          <label>Add a spouse for this child<Select.Root value={spouseMode} onValueChange={(value) => setSpouseMode(value as SpouseMode)}><Select.Trigger /> <Select.Content><Select.Item value="none">No spouse</Select.Item><Select.Item value="existing">Choose existing person</Select.Item><Select.Item value="new">Create new person</Select.Item></Select.Content></Select.Root></label>
          {spouseMode === "existing" && <label>Existing spouse<Select.Root value={existingChildSpouseId} onValueChange={setExistingChildSpouseId}><Select.Trigger placeholder="Select a person" /><Select.Content>{childSpouseOptions.map((person) => <Select.Item key={person.Id} value={person.Id}>{person.Name}</Select.Item>)}</Select.Content></Select.Root></label>}
          {spouseMode === "new" && <div className="child-spouse-fields"><strong>New spouse</strong><label>Gender<input value={oppositeGender(child.Gender)} disabled /></label><label>Name<input required value={newChildSpouse.Name} onChange={(event) => updateNewSpouse({ Name: event.target.value })} /></label><label>Date of birth (optional)<input type="date" value={toInputDate(newChildSpouse.DateOfBirth)} onChange={(event) => updateNewSpouse({ DateOfBirth: fromInputDate(event.target.value) })} /></label><label className="family-check"><input type="checkbox" checked={newChildSpouse.Alive} onChange={(event) => updateNewSpouse({ Alive: event.target.checked, DeateOfDeath: event.target.checked ? "" : newChildSpouse.DeateOfDeath })} />Alive</label><label>Date of death (optional)<input type="date" value={toInputDate(newChildSpouse.DeateOfDeath)} onChange={(event) => updateNewSpouse({ DeateOfDeath: fromInputDate(event.target.value), Alive: event.target.value ? false : newChildSpouse.Alive })} /></label></div>}
        </fieldset>}
        {error && <p role="alert" className="family-error">{error}</p>}
        <Flex justify="end" gap="3" mt="4"><Button type="button" variant="soft" color="gray" onClick={close}>Cancel</Button><Button type="submit" disabled={!selectedParent || saving}>{saving ? "Creating…" : "Create child"}</Button></Flex>
      </form>
    </Dialog.Content>
  </Dialog.Root>;
}
