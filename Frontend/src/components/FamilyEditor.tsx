import { useState } from "react";
import { Button, Dialog, Flex } from "@radix-ui/themes";
import { dateValue, fromInputDate, toInputDate, type User, type FamilyCertificate } from "../helpers/GetPersonDetails";

export type EditorRequest =
  | { kind: "users"; title: string; users: User[]; save: (users: User[]) => void }
  | { kind: "marriage"; family: FamilyCertificate; save: (family: FamilyCertificate) => void }
  | { kind: "delete"; family: FamilyCertificate; save: () => void };

/** Radix supplies focus trapping, Escape dismissal and accessible dialog naming. */
export default function FamilyEditor({ request, close }: { request: EditorRequest; close: () => void }) {
  const [users, setUsers] = useState(request.kind === "users" ? request.users : []);
  const [start, setStart] = useState(request.kind === "marriage" ? toInputDate(request.family.StartOfFamily) : "");
  const [end, setEnd] = useState(request.kind === "marriage" ? toInputDate(request.family.EndOfFamily) : "");
  const [error, setError] = useState("");
  const title = request.kind === "users" ? request.title : request.kind === "marriage" ? "Marriage dates" : "Delete marriage?";
  const update = (index: number, patch: Partial<User>) => setUsers((items) => items.map((user, i) => i === index ? { ...user, ...patch } : user));
  function save() {
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
        request.save(users.map((user) => ({ ...user, Name: user.Name.trim(), DeateOfDeath: user.Alive ? "" : user.DeateOfDeath })));
      } else if (request.kind === "marriage") {
        const first = fromInputDate(start);
        const last = fromInputDate(end);
        if ((first && dateValue(first) === null) || (last && dateValue(last) === null)) throw new Error("Enter valid calendar dates.");
        if (first && last && dateValue(last)! < dateValue(first)!) throw new Error("End date cannot be before start date.");
        request.save({ ...request.family, StartOfFamily: first || null, EndOfFamily: last || null });
      } else request.save();
      close();
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Could not save changes."); }
  }
  return (
    <Dialog.Root open onOpenChange={(open) => { if (!open) close(); }}>
      <Dialog.Content maxWidth="560px">
        <Dialog.Title>{title}</Dialog.Title>
        <Dialog.Description size="2" mb="4">
          {request.kind === "users" ? "Rename or change birth date to update age. IDs remain unchanged. These edits are demo-only." :
            request.kind === "marriage" ? "Both dates are optional. Clear a field to remove its date." :
              "Only this marriage record will be removed. Deletion is blocked while children are linked; unlinking and reassignment are not part of this prototype."}
        </Dialog.Description>
        <form className="family-form" onSubmit={(event) => { event.preventDefault(); save(); }}>
          {users.map((user, index) => (
            <fieldset key={user.Id}>
              <legend>{users.length === 2 ? `Parent ${index + 1}` : "Person details"}</legend>
              <label>Name<input required value={user.Name} onChange={(event) => update(index, { Name: event.target.value })} /></label>
              <label>Date of birth<input type="date" value={toInputDate(user.DateOfBirth)} onChange={(event) => update(index, { DateOfBirth: fromInputDate(event.target.value) })} /></label>
              <label>Gender<select value={user.Gender} onChange={(event) => update(index, { Gender: event.target.value as User["Gender"] })}><option>Male</option><option>Female</option></select></label>
              <label className="family-check"><input type="checkbox" checked={user.Alive} onChange={(event) => update(index, { Alive: event.target.checked })} />Alive</label>
              {!user.Alive && <label>Date of death<input type="date" value={toInputDate(user.DeateOfDeath)} onChange={(event) => update(index, { DeateOfDeath: fromInputDate(event.target.value) })} /></label>}
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
            <Button type="submit" color={request.kind === "delete" ? "red" : undefined}>{request.kind === "delete" ? "Delete marriage" : "Save changes"}</Button>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}
