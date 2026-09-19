import { useState } from "react";
import { Button, Dialog, Flex } from "@radix-ui/themes";
import { dateValue, reverseDate, toInputDate, type User, type FamilyCertificate } from "../helpers/personModel.ts";

export type EditorRequest =
  | { kind: "users"; title: string; users: User[]; save: (users: User[]) => void | Promise<void> }
  | { kind: "marriage"; family: FamilyCertificate; save: (family: FamilyCertificate) => void | Promise<void> }
  /** Destructive operations share one confirmation dialog. */
  | { kind: "confirm"; title: string; description: string; confirmLabel: string; save: () => void | Promise<void> };

/** Radix supplies focus trapping, Escape dismissal and accessible dialog naming. */
export default function FamilyEditor({ request, close, onCloseFocus }: { request: EditorRequest; close: () => void; onCloseFocus?: () => void }) {
  const [users, setUsers] = useState(request.kind === "users" ? request.users : []);
  const [start, setStart] = useState(request.kind === "marriage" ? toInputDate(request.family.StartOfFamily) : "");
  const [end, setEnd] = useState(request.kind === "marriage" ? toInputDate(request.family.EndOfFamily) : "");
  const [error, setError] = useState("");
  // Saving disables the submit button, so a slow write cannot be sent twice.
  const [saving, setSaving] = useState(false);
  const title = request.kind === "confirm" ? request.title : request.kind === "marriage" ? "Marriage dates" : request.title;
  const description = request.kind === "users"
    ? "Rename or change birth date to update age. IDs remain unchanged."
    : request.kind === "marriage"
      ? "Both dates are optional. Clear a field to remove its date. The people in a marriage never change."
      : request.description;
  const update = (index: number, patch: Partial<User>) => setUsers((items) => items.map((user, i) => i === index ? { ...user, ...patch } : user));
  async function save() {
    setSaving(true);
    try {
      if (request.kind === "users") {
        for (const user of users) {
          if (!user.name.trim()) throw new Error("Every person needs a name.");
          for (const date of [user.dateOfBirth, user.dateOfDeath]) {
            if (date && dateValue(date) === null) throw new Error("Enter a valid calendar date.");
          }
          const birth = dateValue(user.dateOfBirth);
          const death = dateValue(user.dateOfDeath);
          if (birth !== null && birth > Date.now()) throw new Error("Birth date cannot be in the future.");
          if (!user.alive && death !== null && (death > Date.now() || (birth !== null && death < birth))) {
            throw new Error("Death date must be after birth and not in the future.");
          }
        }
        await request.save(users.map((user) => ({ ...user, name: user.name.trim(), dateOfDeath: user.alive ? "" : user.dateOfDeath })));
      } else if (request.kind === "marriage") {
        const first = reverseDate(start);
        const last = reverseDate(end);
        if ((first && dateValue(first) === null) || (last && dateValue(last) === null)) throw new Error("Enter valid calendar dates.");
        if (first && last && dateValue(last)! < dateValue(first)!) throw new Error("End date cannot be before start date.");
        await request.save({ ...request.family, StartOfFamily: first || null, EndOfFamily: last || null });
      } else await request.save();
      close();
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Could not save changes."); }
    finally { setSaving(false); }
  }
  return (
    <Dialog.Root open onOpenChange={(open) => { if (!open) close(); }}>
      <Dialog.Content maxWidth="560px" onCloseAutoFocus={onCloseFocus ? (event) => { event.preventDefault(); onCloseFocus(); } : undefined}>
        <Dialog.Title>{title}</Dialog.Title>
        <Dialog.Description size="2" mb="4">{description}</Dialog.Description>
        <form className="family-form" onSubmit={(event) => { event.preventDefault(); void save(); }}>
          {users.map((user, index) => (
            <fieldset key={user.id}>
              <legend>{users.length === 2 ? `Parent ${index + 1}` : "Person details"}</legend>
              <label>Name<input required value={user.name} onChange={(event) => update(index, { name: event.target.value })} /></label>
              <label>Date of birth<input type="date" value={toInputDate(user.dateOfBirth)} onChange={(event) => update(index, { dateOfBirth: reverseDate(event.target.value) })} /></label>
              <label>Gender<select value={user.gender} onChange={(event) => update(index, { gender: event.target.value as User["gender"] })}><option>Male</option><option>Female</option></select></label>
              <label className="family-check"><input type="checkbox" checked={user.alive} onChange={(event) => update(index, { alive: event.target.checked })} />Alive</label>
              {!user.alive && <label>Date of death<input type="date" value={toInputDate(user.dateOfDeath)} onChange={(event) => update(index, { dateOfDeath: reverseDate(event.target.value) })} /></label>}
              <small>ID: {user.id}</small>
            </fieldset>
          ))}
          {request.kind === "marriage" && <>
            <label>Start date<input type="date" value={start} onChange={(event) => setStart(event.target.value)} /></label>
            <label>End date<input type="date" value={end} onChange={(event) => setEnd(event.target.value)} /></label>
          </>}
          {error && <p role="alert" className="family-error">{error}</p>}
          <Flex justify="end" gap="3" mt="4">
            <Button type="button" variant="soft" color="gray" onClick={close}>Cancel</Button>
            <Button type="submit" disabled={saving} color={request.kind === "confirm" ? "red" : undefined}>
              {saving ? "Saving…" : request.kind === "confirm" ? request.confirmLabel : "Save changes"}
            </Button>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}
