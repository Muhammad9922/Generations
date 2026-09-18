import { useMemo, useState } from "react";
import { Button, Dialog, Flex, Select } from "@radix-ui/themes";
import { createMarriage } from "../api/marriages.ts";
import type { Person } from "../api/contracts.ts";
import { oppositeGender, personToUser, reverseDate, toInputDate, type FamilyCertificate, type User } from "../helpers/personModel.ts";
import { createPersonFromDraft } from "../helpers/personDrafts.ts";

const CREATE_NEW = "__create_new__";

interface Props {
  primary: User;
  marriages: FamilyCertificate[];
  people: Person[];
  close: () => void;
  /** Called after the writes succeed so the page can re-read itself. */
  onSaved: () => void;
}

/** Adds a spouse by creating a marriage; a marriage's people never change later. */
export default function SpouseCreatorDialog({ primary, marriages, people, close, onSaved }: Props) {
  const currentSpouseIds = useMemo(() => new Set(marriages.flatMap((marriage) => marriage.Spouse ?? []).map((person) => person.Id)), [marriages]);
  const choices = useMemo(() => people.map(personToUser).filter((person) => person.Gender === oppositeGender(primary.Gender) && !currentSpouseIds.has(person.Id)), [people, primary.Gender, currentSpouseIds]);
  const [selection, setSelection] = useState("");
  const [name, setName] = useState("");
  const [alive, setAlive] = useState(true);
  const [birth, setBirth] = useState("");
  const [death, setDeath] = useState("");
  const [marriageStart, setMarriageStart] = useState("");
  const [marriageEnd, setMarriageEnd] = useState("");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const isNew = selection === CREATE_NEW;

  async function submit() {
    try {
      let spouse = choices.find((person) => person.Id === selection);
      if (isNew) spouse = await createPersonFromDraft({ name, birth, alive, death }, oppositeGender(primary.Gender), "new spouse");
      if (!spouse) throw new Error("Choose an existing person or create a new one.");
      setSaving(true);
      await createMarriage({ SpouseOne: primary.Id, SpouseTwo: spouse.Id, DateStart: marriageStart, DateEnd: marriageEnd });
      onSaved();
      close();
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Could not add spouse."); }
    finally { setSaving(false); }
  }

  return <Dialog.Root open onOpenChange={(open) => { if (!open) close(); }}>
    <Dialog.Content maxWidth="560px" className="family-dialog"><span className="family-eyebrow">FAMILY CONNECTION</span><Dialog.Title>Add spouse</Dialog.Title><Dialog.Description size="2" mb="4">Choose an existing eligible person or create a new one. This creates their marriage with {primary.Name}.</Dialog.Description>
      <form className="family-form" onSubmit={(event) => { event.preventDefault(); void submit(); }}>
        <label>Spouse<Select.Root value={selection} onValueChange={setSelection}><Select.Trigger placeholder="Select a person" /><Select.Content>{choices.length > 0 && <Select.Group><Select.Label>Eligible people</Select.Label>{choices.map((person) => <Select.Item key={person.Id} value={person.Id}>{person.Name}</Select.Item>)}</Select.Group>}<Select.Separator /><Select.Item value={CREATE_NEW}>Create new person</Select.Item></Select.Content></Select.Root></label>
        {isNew && <fieldset><legend>New spouse</legend><label>Name<input required value={name} onChange={(event) => setName(event.target.value)} /></label><label>Gender<input value={oppositeGender(primary.Gender)} disabled /></label><label>Date of birth (optional)<input type="date" value={toInputDate(birth)} onChange={(event) => setBirth(reverseDate(event.target.value))} /></label><label className="family-check"><input type="checkbox" checked={alive} onChange={(event) => { setAlive(event.target.checked); if (event.target.checked) setDeath(""); }} />Alive</label><label>Date of death (optional)<input type="date" value={toInputDate(death)} onChange={(event) => { const next = reverseDate(event.target.value); setDeath(next); if (next) setAlive(false); }} /></label></fieldset>}
        {selection && <fieldset><legend>Marriage dates</legend><label>Start date (optional)<input type="date" value={toInputDate(marriageStart)} onChange={(event) => setMarriageStart(reverseDate(event.target.value))} /></label><label>End date (optional)<input type="date" value={toInputDate(marriageEnd)} onChange={(event) => setMarriageEnd(reverseDate(event.target.value))} /></label></fieldset>}
        {error && <p role="alert" className="family-error">{error}</p>}<Flex justify="end" gap="3" mt="4"><Button type="button" variant="soft" color="gray" onClick={close}>Cancel</Button><Button type="submit" disabled={!selection || saving}>{saving ? "Saving…" : "Add spouse"}</Button></Flex>
      </form>
    </Dialog.Content>
  </Dialog.Root>;
}
