import { useEffect, useRef, useState } from "react";
import MarriageMenubar from "./MarriageMenubar";
import { Plus } from "lucide-react";
import type { Person } from "../helpers/GetUsers";
import { GetPersonDetails, deleteMarriage, otherSpouses, replaceUser, sortedChildren, withParents, type PersonDetailsData, type User } from "../helpers/GetPersonDetails";
import FamilyPersonCard from "./FamilyPersonCard";
import FamilyEditor, { type EditorRequest } from "./FamilyEditor";
import ChildCreatorDialog from "./ChildCreatorDialog";
import SpouseCreatorDialog from "./SpouseCreatorDialog";
import ParentCreatorDialog from "./ParentCreatorDialog";

/** Route-keyed prototype state. No mutations are sent to the backend. */
export default function FamilyView({ id, people, onRename }: { id: string; people: Person[]; onRename: (id: string, name: string) => void }) {
  // Keep the initial lookup stable when App updates a search label after a rename.
  const [initialPeople] = useState(people);
  const [data, setData] = useState<PersonDetailsData | null>(null);
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState("");
  const [hovered, setHovered] = useState<string | null>(null);
  const [focused, setFocused] = useState<string | null>(null);
  const [editor, setEditor] = useState<EditorRequest | null>(null);
  const [creatingChild, setCreatingChild] = useState(false);
  const [creatingSpouse, setCreatingSpouse] = useState(false);
  const [creatingParents, setCreatingParents] = useState(false);
  const [restoreMenuFocus, setRestoreMenuFocus] = useState(false);
  const returnFocus = useRef<HTMLButtonElement | null>(null);
  const spousesHeading = useRef<HTMLHeadingElement>(null);
  const openMarriageEditor = (request: EditorRequest, trigger: HTMLButtonElement | null) => {
    returnFocus.current = trigger;
    setRestoreMenuFocus(true);
    setEditor(request);
  };
  useEffect(() => {
    let cancelled = false;
    GetPersonDetails(id, initialPeople).then((result) => {
      if (!cancelled) { setData(result); setLoaded(true); }
    }).catch(() => { if (!cancelled) { setError("Could not load this family."); setLoaded(true); } });
    return () => { cancelled = true; };
  }, [id, initialPeople]);
  if (error) return <p role="alert">{error}</p>;
  if (!loaded) return <p role="status">Loading family…</p>;
  if (!data) return <h1>Person not found</h1>;
  const active = hovered ?? focused;
  const children = sortedChildren(data.Families);
  const editUser = (user: User) => setEditor({ kind: "users", title: "Edit person", users: [user], save: ([updated]) => { setData(replaceUser(data, updated)); onRename(updated.Id, updated.Name); } });
  const card = (user: User, label: string, familyId?: string, focal = false) => <FamilyPersonCard key={`${focal ? "focal" : familyId ?? "person"}-${user.Id}`}
    user={user} label={label} familyId={familyId} highlighted={!!familyId && active === familyId}
    onHover={focal ? () => setHovered(null) : setHovered} onFocus={focal ? () => setFocused(null) : setFocused} onEdit={editUser} />;
  const openChildCreator = (trigger: HTMLButtonElement | null) => {
    returnFocus.current = trigger;
    setRestoreMenuFocus(true);
    setCreatingChild(true);
  };
  const label = (familyId: string) => `Family ${data.Families.findIndex((item) => item.Id === familyId) + 1}`;
  return <>
    <header className="family-heading"><span className="family-eyebrow">YOUR FAMILY, CONNECTED</span><h1>{data.Person.Name}</h1>
      <p>Hover or focus a spouse or child to illuminate their family.</p>
      <p className="family-notice">Interactive sample · Family edits reset when you leave; search renames last until refresh. No backend writes.</p>
    </header>
    <div className="family-layout">
      <section><h2>Parents</h2><div className="parents-grid">
        {data.Parents ? data.Parents.map((user, index) => card(user, `Parent ${index + 1}`)) :
          <button className="family-add" onClick={() => setCreatingParents(true)}><Plus />Add parents<span>Select people or create both parents</span></button>}
      </div></section>
      <section><h2 ref={spousesHeading} tabIndex={-1}>Person & spouses</h2>{card(data.Person, "Selected person", active ?? undefined, true)}
        <div className="spouses-grid">{data.Families.map((family) => <div className="family-marriage" key={family.Id}>
          {otherSpouses(family, id).map((user) => card(user, `${label(family.Id)} · Spouse`, family.Id))}
          <MarriageMenubar label={label(family.Id)} canAddChild={otherSpouses(family, id).length > 0}
            onAddChild={openChildCreator}
            onEditDates={(trigger) => openMarriageEditor({ kind: "marriage", family, save: (updated) => setData({ ...data, Families: data.Families.map((item) => item.Id === updated.Id ? updated : item) }) }, trigger)}
            onDelete={(trigger) => openMarriageEditor({ kind: "delete", family, save: () => setData(deleteMarriage(data, family.Id)) }, trigger)} />
        </div>)}</div>
        <button className="family-add" onClick={() => setCreatingSpouse(true)}><Plus />Add spouse</button>
        <button className="family-add" onClick={() => openChildCreator(null)} disabled={!data.Families.some((family) => otherSpouses(family, id).length) && !people.some((person) => person.id !== id && person.gender !== data.Person.Gender)}><Plus />Add child<span>Choose an existing spouse or another eligible person</span></button>
      </section>
      <section><h2>Children <small>{children.length} · Oldest first</small></h2>
        <div className="children-grid">{children.map(({ user, familyId }) => card(user, label(familyId), familyId))}</div>
        {!children.length && <p className="family-empty">{data.Families.some((family) => otherSpouses(family, id).length) ? "No children added yet. Use Actions beside a spouse to add a child." : "Add a spouse before adding children."}</p>}
      </section>
    </div>
    {editor && <FamilyEditor request={editor} close={() => setEditor(null)} onCloseFocus={restoreMenuFocus ? () => {
      const trigger = returnFocus.current;
      if (trigger?.isConnected) trigger.focus();
      else spousesHeading.current?.focus();
      returnFocus.current = null;
      setRestoreMenuFocus(false);
    } : undefined} />}
    {creatingChild && <ChildCreatorDialog primary={data.Person} families={data.Families} people={people} close={() => {
      setCreatingChild(false);
      const trigger = returnFocus.current;
      if (trigger?.isConnected) trigger.focus(); else spousesHeading.current?.focus();
      returnFocus.current = null;
      setRestoreMenuFocus(false);
    }} save={(family, child, childMarriage) => setData((current) => {
      if (!current) return current;
      const exists = current.Families.some((item) => item.Id === family.Id);
      const updatedFamily = { ...family, Chidren: [...(family.Chidren ?? []), child] };
      return { ...current, Families: exists ? current.Families.map((item) => item.Id === family.Id ? updatedFamily : item) : [...current.Families, updatedFamily, ...(childMarriage ? [childMarriage] : [])] };
    })} />}
    {creatingSpouse && <SpouseCreatorDialog primary={data.Person} families={data.Families} people={people} close={() => setCreatingSpouse(false)} save={(family) => setData((current) => current ? { ...current, Families: [...current.Families, family] } : current)} />}
    {creatingParents && <ParentCreatorDialog child={data.Person} people={people} close={() => setCreatingParents(false)} save={(parents) => setData((current) => current ? withParents(current, parents) : current)} />}
  </>;
}
