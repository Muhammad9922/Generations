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
          if (!user.Name.trim()) throw new Error("Every person needs a name.");
          for (const date of [user.DateOfBirth, user.DeateOfDeath]) {
            if (date && dateValue(date) === null) throw new Error("Enter a valid calendar date.");
          }
          const birth = dateValue(user.DateOfBirth);
          const death = dateValue(user.DeateOfDeath);
          if (birth !== null && birth > Date.now()) throw new Error("Birth date cannot be in the future.");
          if (!user.Alive && death !== null && (death > Date.now() || (birth !== null && death < birth))) {
            throw new Error("Death date must be after birth and not in the future.");
          }
        }
        await request.save(users.map((user) => ({ ...user, Name: user.Name.trim(), DeateOfDeath: user.Alive ? "" : user.DeateOfDeath })));
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
            <fieldset key={user.Id}>
              <legend>{users.length === 2 ? `Parent ${index + 1}` : "Person details"}</legend>
              <label>Name<input required value={user.Name} onChange={(event) => update(index, { Name: event.target.value })} /></label>
              <label>Date of birth<input type="date" value={toInputDate(user.DateOfBirth)} onChange={(event) => update(index, { DateOfBirth: reverseDate(event.target.value) })} /></label>
              <label>Gender<select value={user.Gender} onChange={(event) => update(index, { Gender: event.target.value as User["Gender"] })}><option>Male</option><option>Female</option></select></label>
              <label className="family-check"><input type="checkbox" checked={user.Alive} onChange={(event) => update(index, { Alive: event.target.checked })} />Alive</label>
              {!user.Alive && <label>Date of death<input type="date" value={toInputDate(user.DeateOfDeath)} onChange={(event) => update(index, { DeateOfDeath: reverseDate(event.target.value) })} /></label>}
              <small>ID: {user.Id}</small>
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
