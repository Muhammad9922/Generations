import { useState } from "react";
import { Button, Dialog, Flex, Select } from "@radix-ui/themes";
import { dateValue, newUser, reverseDate, toInputDate, type User } from "../helpers/personModel.ts";
import { createPersonFromDraft } from "../helpers/personDrafts.ts";

interface Props {
  close: () => void;
  /** Receives the saved person, so the caller can open the page the API assigned. */
  onCreated: (person: User) => void | Promise<void>;
}

/**
 * Creates the very first person of a family. Unlike the spouse, parent and
 * child dialogs there is no existing person to derive a gender or a name from,
 * so every field is entered here and `CreatePerson` assigns the ID the page
 * then navigates to.
 */
export default function PersonCreatorDialog({ close, onCreated }: Props) {
  const [person, setPerson] = useState<User>(newUser(""));
  const [error, setError] = useState("");
  // Saving disables the submit button, so a slow write cannot be sent twice.
  const [saving, setSaving] = useState(false);
  const update = (changes: Partial<User>) => setPerson((current) => ({ ...current, ...changes }));

  async function submit() {
    setSaving(true);
    try {
      // `type="date"` guarantees the shape, but a future or inverted date still
      // needs rejecting before it reaches the API.
      const birth = dateValue(person.dateOfBirth);
      const death = dateValue(person.dateOfDeath);
      if (person.dateOfBirth && birth === null) throw new Error("Enter a valid calendar date.");
      if (birth !== null && birth > Date.now()) throw new Error("Birth date cannot be in the future.");
      if (!person.alive && person.dateOfDeath && (death === null || death > Date.now() || (birth !== null && death < birth))) {
        throw new Error("Death date must be after birth and not in the future.");
      }
      const created = await createPersonFromDraft(
        { name: person.name, birth: person.dateOfBirth, alive: person.alive, death: person.dateOfDeath },
        person.gender,
        "person",
      );
      await onCreated(created);
      close();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Could not create this person.");
    } finally { setSaving(false); }
  }

  return <Dialog.Root open onOpenChange={(open) => { if (!open) close(); }}>
    <Dialog.Content maxWidth="560px" className="family-dialog">
      <span className="family-eyebrow">NEW FAMILY</span>
      <Dialog.Title>Create new person</Dialog.Title>
      <Dialog.Description size="2" mb="4">This person starts the family. Once created, the page opens as their person page.</Dialog.Description>
      <form className="family-form" onSubmit={(event) => { event.preventDefault(); void submit(); }}>
        <fieldset>
          <legend>Person details</legend>
          <label>Name<input required value={person.name} onChange={(event) => update({ name: event.target.value })} /></label>
          <label>Gender<Select.Root value={person.gender} onValueChange={(gender) => update({ gender: gender as User["gender"] })}><Select.Trigger /><Select.Content><Select.Item value="Male">Male</Select.Item><Select.Item value="Female">Female</Select.Item></Select.Content></Select.Root></label>
          <label>Date of birth (optional)<input type="date" value={toInputDate(person.dateOfBirth)} onChange={(event) => update({ dateOfBirth: reverseDate(event.target.value) })} /></label>
          <label className="family-check"><input type="checkbox" checked={person.alive} onChange={(event) => update({ alive: event.target.checked, dateOfDeath: event.target.checked ? "" : person.dateOfDeath })} />Alive</label>
          {!person.alive && <label>Date of death (optional)<input type="date" value={toInputDate(person.dateOfDeath)} onChange={(event) => update({ dateOfDeath: reverseDate(event.target.value) })} /></label>}
        </fieldset>
        {error && <p role="alert" className="family-error">{error}</p>}
        <Flex justify="end" gap="3" mt="4">
          <Button type="button" variant="soft" color="gray" onClick={close}>Cancel</Button>
          <Button type="submit" disabled={saving}>{saving ? "Saving…" : "Create person"}</Button>
        </Flex>
      </form>
    </Dialog.Content>
  </Dialog.Root>;
}
