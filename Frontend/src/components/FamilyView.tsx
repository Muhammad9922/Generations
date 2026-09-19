import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router";
import MarriageMenubar from "./MarriageMenubar";
import { Plus, Trash2 } from "lucide-react";
import type { Person } from "../api/contracts.ts";
import {
  otherSpouses, sortedChildren, type FamilyCertificate, type PersonDetailsData, type User,
} from "../helpers/personModel.ts";
import FamilyPersonCard from "./FamilyPersonCard";
import FamilyEditor, { type EditorRequest } from "./FamilyEditor";
import ChildCreatorDialog from "./ChildCreatorDialog";
import SpouseCreatorDialog from "./SpouseCreatorDialog";
import ParentCreatorDialog from "./ParentCreatorDialog";
import { updateUserDetails } from "../api/people.ts";
import { getPersonDetails } from "../api/family.ts";
import { isAbort } from "../api/http.ts";
import { createMarriage, deleteMarriage, removeChild, updateMarriage } from "../api/marriages.ts";

interface Props {
  id: string;
  people: Person[];
  /** Asks App to re-read the people list so the palette stays current. */
  onPeopleChanged: () => void;
}

/**
 * Person details: the parents' marriage plus every marriage this person is a
 * spouse in. Every write ends by re-reading this payload, so what is on screen
 * is what the API holds rather than a locally patched guess.
 */
export default function FamilyView({ id, people, onPeopleChanged }: Props) {
  const navigate = useNavigate();
  const [data, setData] = useState<PersonDetailsData | null>(null);
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState("");
  // Bumped after every successful write, and by the retry button on failure.
  const [reload, setReload] = useState(0);
  const refresh = useCallback(() => setReload((value) => value + 1), []);
  const [hovered, setHovered] = useState<string | null>(null);
  const [focused, setFocused] = useState<string | null>(null);
  const [editor, setEditor] = useState<EditorRequest | null>(null);
  const [creatingChild, setCreatingChild] = useState(false);
  const [creatingSpouse, setCreatingSpouse] = useState(false);
  const [creatingParents, setCreatingParents] = useState(false);
  const [restoreMenuFocus, setRestoreMenuFocus] = useState(false);
  const returnFocus = useRef<HTMLButtonElement | null>(null);
  const spousesHeading = useRef<HTMLHeadingElement>(null);
  const rememberTrigger = (trigger: HTMLButtonElement | null) => {
    returnFocus.current = trigger;
    setRestoreMenuFocus(true);
  };
  const openEditor = (request: EditorRequest, trigger: HTMLButtonElement | null) => {
    rememberTrigger(trigger);
    setEditor(request);
  };
  const openConfirm = (request: { title: string; description: string; confirmLabel: string; save: () => void | Promise<void> }, trigger: HTMLButtonElement | null) =>
    openEditor({ kind: "confirm", ...request }, trigger);
  const openChildCreator = (trigger: HTMLButtonElement | null) => {
    rememberTrigger(trigger);
    setCreatingChild(true);
  };
  /** Opens any card as the primary person, through the same route the palette uses. */
  const openPerson = (person: User) => navigate(`/people/${encodeURIComponent(person.id)}`);
  useEffect(() => {
    // A superseded request is cancelled, so a slow answer cannot replace a newer
    // one. `loaded` stays true across refreshes to avoid a loading flash.
    const controller = new AbortController();
    getPersonDetails(id, controller.signal).then((result) => {
      setData(result);
      setError("");
      setLoaded(true);
    }).catch((cause) => {
      if (isAbort(cause)) return;
      setError(cause instanceof Error ? cause.message : "Could not load this family.");
      setLoaded(true);
    });
    return () => controller.abort();
  }, [id, reload]);
  if (!loaded) return <p role="status">Loading family…</p>;
  if (error) return <p role="alert">{error}</p>;
  if (!data) return <h1>Person not found</h1>;
  const active = hovered ?? focused;
  const children = sortedChildren(data.Marriages);
  const parentsMarriage = data.ParentsMarriage;
  const editUser = (user: User) => setEditor({ kind: "users", title: "Edit person", users: [user], save: async ([updated]) => {
    await updateUserDetails(updated);
    onPeopleChanged();
    refresh();
  } });
  const card = (user: User, label: string, familyId?: string, focal = false, removable = false) => <FamilyPersonCard key={`${focal ? "focal" : familyId ?? "person"}-${user.id}`}
    user={user} label={label} familyId={familyId} highlighted={!!familyId && active === familyId}
    onHover={focal ? () => setHovered(null) : setHovered} onFocus={focal ? () => setFocused(null) : setFocused} onPrimary={editUser}
    onOpen={focal ? undefined : openPerson}
    onRemoveChild={removable && familyId ? (child, trigger) => openConfirm({
      title: "Remove child?",
      description: `${child.name} is removed from ${label}. The marriage and both spouses stay; this person is simply no longer their child.`,
      confirmLabel: "Remove child",
      save: async () => { await removeChild(familyId, child.id); refresh(); },
    }, trigger) : undefined} />;
  const label = (marriageId: string) => `Family ${data.Marriages.findIndex((item) => item.Id === marriageId) + 1}`;
  /** Disbanding a marriage orphans the children it produced, so say how many. */
  const delinkDescription = (marriage: FamilyCertificate) => {
    const count = (marriage.Children ?? []).length;
    const orphaned = count === 0 ? "" : count === 1 ? ", and its child is left without parents" : `, and its ${count} children are left without parents`;
    return `${label(marriage.Id)} is disbanded: the spouse is unlinked${orphaned}. The people themselves are kept.`;
  };
  const closeAndRestoreFocus = () => {
    const trigger = returnFocus.current;
    if (trigger?.isConnected) trigger.focus(); else spousesHeading.current?.focus();
    returnFocus.current = null;
    setRestoreMenuFocus(false);
  };
  return <>
    <header className="family-heading"><span className="family-eyebrow">YOUR FAMILY, CONNECTED</span><h1>{data.Person.name}</h1>
      <p>Hover or focus a spouse or child to illuminate their family.</p>
      <p className="family-notice">Every group here is a marriage: the marriage {data.Person.name} is a child of, plus every marriage they are a spouse in.</p>
    </header>
    <div className="family-layout">
      <section><h2>Parents <small>the marriage they are a child of</small></h2>
        <div className="parents-grid">
          {parentsMarriage ? (parentsMarriage.Spouse ?? []).map((user, index) => card(user, `Parent ${index + 1}`)) :
            <button className="family-add" onClick={() => setCreatingParents(true)}><Plus />Add parents<span>Select people or create both parents</span></button>}
        </div>
        {parentsMarriage && <button className="family-add family-unlink" onClick={(event) => openConfirm({
          title: "Delete parents?",
          description: `${data.Person.name} is removed as a child of their parents' marriage. The two parents stay married to each other; they are simply no longer linked as ${data.Person.name}'s parents.`,
          confirmLabel: "Delete parents",
          save: async () => { await removeChild(parentsMarriage.Id, data.Person.id); refresh(); },
        }, event.currentTarget)}><Trash2 size={16} aria-hidden="true" />Delete parents<span>Delinks this person; the parents stay married</span></button>}
      </section>
      <section><h2 ref={spousesHeading} tabIndex={-1}>Person & spouses</h2>{card(data.Person, "Selected person", active ?? undefined, true)}
        <div className="spouses-grid">{data.Marriages.map((marriage) => <div className="family-marriage" key={marriage.Id}>
          {otherSpouses(marriage, id).map((user) => card(user, `${label(marriage.Id)} · Spouse`, marriage.Id))}
          <MarriageMenubar label={label(marriage.Id)} canAddChild={otherSpouses(marriage, id).length > 0}
            onAddChild={openChildCreator}
            onEditDates={(trigger) => openEditor({ kind: "marriage", family: marriage, save: async (updated) => {
              await updateMarriage(updated.Id, { DateStart: updated.StartOfFamily ?? "", DateEnd: updated.EndOfFamily ?? "" });
              refresh();
            } }, trigger)}
            onDelinkSpouse={(trigger) => openConfirm({
              title: "Delink spouse?",
              description: delinkDescription(marriage),
              confirmLabel: "Delink spouse",
              save: async () => { await deleteMarriage(marriage.Id); refresh(); },
            }, trigger)} />
        </div>)}</div>
        <button className="family-add" onClick={() => setCreatingSpouse(true)}><Plus />Add spouse</button>
        <button className="family-add" onClick={() => openChildCreator(null)} disabled={!data.Marriages.some((marriage) => otherSpouses(marriage, id).length) && !people.some((person) => person.id !== id && person.gender !== data.Person.gender)}><Plus />Add child<span>Choose an existing spouse or another eligible person</span></button>
      </section>
      <section><h2>Children <small>{children.length} · Oldest first</small></h2>
        <div className="children-grid">{children.map(({ user, familyId }) => card(user, label(familyId), familyId, false, true))}</div>
        {!children.length && <p className="family-empty">{data.Marriages.some((marriage) => otherSpouses(marriage, id).length) ? "No children added yet. Use Actions beside a spouse to add a child." : "Add a spouse before adding children."}</p>}
      </section>
    </div>
    {editor && <FamilyEditor request={editor} close={() => setEditor(null)} onCloseFocus={restoreMenuFocus ? closeAndRestoreFocus : undefined} />}
    {creatingChild && <ChildCreatorDialog primary={data.Person} marriages={data.Marriages} people={people} close={() => {
      setCreatingChild(false);
      closeAndRestoreFocus();
    }} onSaved={() => { onPeopleChanged(); refresh(); }} />}
    {creatingSpouse && <SpouseCreatorDialog primary={data.Person} marriages={data.Marriages} people={people} close={() => setCreatingSpouse(false)} onSaved={() => { onPeopleChanged(); refresh(); }} />}
    {creatingParents && <ParentCreatorDialog child={data.Person} people={people} close={() => setCreatingParents(false)} save={async (parents) => {
      // Adding parents is creating the marriage they are the spouses of, with
      // this person linked as its child.
      await createMarriage({ SpouseOne: parents[0].id, SpouseTwo: parents[1].id, DateStart: "", DateEnd: "", childrenIds: [data.Person.id] });
      onPeopleChanged();
      refresh();
    }} />}
  </>;
}
