import { useMemo, useState } from "react";
import { Button, Dialog, Flex, Select } from "@radix-ui/themes";
import { createMarriage } from "../api/marriages.ts";
import { createPerson, registerPeopleForMarriageValidation } from "../api/people.ts";
import { fromInputDate, toInputDate, type FamilyCertificate, type User } from "../helpers/GetPersonDetails";
import type { Person } from "../helpers/GetUsers";

const CREATE_NEW = "__create_new__";
const oppositeGender = (gender: User["Gender"]): User["Gender"] => gender === "Male" ? "Female" : "Male";
const asUser = (person: Person): User => ({ Id: person.id, Name: person.name, Gender: person.gender, Alive: person.alive, DateOfBirth: person.dateOfBirth ?? "", DeateOfDeath: person.dateOfDeath ?? "" });

interface Props { primary: User; families: FamilyCertificate[]; people: Person[]; close: () => void; save: (family: FamilyCertificate) => void; }

/** Adds a spouse from the shared people list or creates an API-ready new record. */
export default function SpouseCreatorDialog({ primary, families, people, close, save }: Props) {
  const currentSpouseIds = useMemo(() => new Set(families.flatMap((family) => family.Spouse ?? []).map((person) => person.Id)), [families]);
  const choices = useMemo(() => people.map(asUser).filter((person) => person.Gender === oppositeGender(primary.Gender) && !currentSpouseIds.has(person.Id)), [people, primary.Gender, currentSpouseIds]);
  const [selection, setSelection] = useState("");
  const [name, setName] = useState("");
  const [alive, setAlive] = useState(true);
  const [birth, setBirth] = useState("");
  const [death, setDeath] = useState("");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const isNew = selection === CREATE_NEW;

  async function submit() {
    try {
      let spouse = choices.find((person) => person.Id === selection);
      if (isNew) {
        if (!name.trim()) throw new Error("The new spouse needs a name.");
        const id = await createPerson({ name: name.trim(), gender: oppositeGender(primary.Gender), alive, date_of_birth: birth || null, date_of_death: alive ? null : death || null });
        spouse = { Id: id, Name: name.trim(), Gender: oppositeGender(primary.Gender), Alive: alive, DateOfBirth: birth, DeateOfDeath: alive ? "" : death };
      }
      if (!spouse) throw new Error("Choose an existing person or create a new one.");
      setSaving(true);
      registerPeopleForMarriageValidation([primary, spouse, ...choices]);
      const id = await createMarriage({ spouseOneId: primary.Id, spouseTwoId: spouse.Id });
      save({ Id: id, Spouse: [primary, spouse], Chidren: [] });
      close();
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Could not add spouse."); }
    finally { setSaving(false); }
  }

  return <Dialog.Root open onOpenChange={(open) => { if (!open) close(); }}>
    <Dialog.Content maxWidth="560px" className="family-dialog"><span className="family-eyebrow">FAMILY CONNECTION</span><Dialog.Title>Add spouse</Dialog.Title><Dialog.Description size="2" mb="4">Choose an existing eligible person or create a new one.</Dialog.Description>
      <form className="family-form" onSubmit={(event) => { event.preventDefault(); void submit(); }}>
        <label>Spouse<Select.Root value={selection} onValueChange={setSelection}><Select.Trigger placeholder="Select a person" /><Select.Content>{choices.length > 0 && <Select.Group><Select.Label>Eligible people</Select.Label>{choices.map((person) => <Select.Item key={person.Id} value={person.Id}>{person.Name}</Select.Item>)}</Select.Group>}<Select.Separator /><Select.Item value={CREATE_NEW}>Create new person</Select.Item></Select.Content></Select.Root></label>
        {isNew && <fieldset><legend>New spouse</legend><label>Name<input required value={name} onChange={(event) => setName(event.target.value)} /></label><label>Gender<input value={oppositeGender(primary.Gender)} disabled /></label><label>Date of birth (optional)<input type="date" value={toInputDate(birth)} onChange={(event) => setBirth(fromInputDate(event.target.value))} /></label><label className="family-check"><input type="checkbox" checked={alive} onChange={(event) => { setAlive(event.target.checked); if (event.target.checked) setDeath(""); }} />Alive</label><label>Date of death (optional)<input type="date" value={toInputDate(death)} onChange={(event) => { const next = fromInputDate(event.target.value); setDeath(next); if (next) setAlive(false); }} /></label></fieldset>}
        {error && <p role="alert" className="family-error">{error}</p>}<Flex justify="end" gap="3" mt="4"><Button type="button" variant="soft" color="gray" onClick={close}>Cancel</Button><Button type="submit" disabled={!selection || saving}>{saving ? "Creating…" : "Add spouse"}</Button></Flex>
      </form>
    </Dialog.Content>
  </Dialog.Root>;
}
