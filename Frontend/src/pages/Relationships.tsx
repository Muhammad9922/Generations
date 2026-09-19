import { useEffect, useMemo, useState } from "react";
import { Button } from "@radix-ui/themes";
import { ArrowLeft, UserRound } from "lucide-react";
import { Link, useNavigate, useParams } from "react-router";
import FamilyPersonCard from "../components/FamilyPersonCard";
import PersonPicker from "../components/PersonPicker";
import { loadRelatives, type RelationReport } from "../api/relations.ts";
import { isAbort } from "../api/http.ts";
import { personToUser, type User } from "../helpers/personModel.ts";
import type { Person } from "../api/contracts.ts";
import {
  RELATION_LABELS, RELATION_ORDER, RELATION_PLURALS, sortRelatives,
  type RelationKind, type RelationSort, type Relative,
} from "../helpers/relations.ts";

interface Props {
  /** The shared people list, used as the picker's options. */
  people: Person[];
  /** True while that list is loading for the first time. */
  loading: boolean;
  /** A user-facing failure from the shared people load, or null. */
  error: string | null;
}

type GenderFilter = "all" | User["gender"];
type StatusFilter = "all" | "alive" | "deceased";

/**
 * The extended family of one person: parents, siblings, grandparents, uncles,
 * aunts, cousins, nieces and nephews, with filters over the whole set.
 *
 * Nothing here is stored server-side — the relationships are derived from the
 * payloads `GET /people/:id` already returns (see `src/api/relations.ts`), so
 * the page says which person it is reading from and can be pointed at anyone.
 */
export default function Relationships({ people, loading, error }: Props) {
  const { id } = useParams();
  const navigate = useNavigate();
  const [report, setReport] = useState<RelationReport | null>(null);
  const [loaded, setLoaded] = useState(false);
  const [loadError, setLoadError] = useState("");
  // Bumped by Retry, and after nothing else: this page never writes.
  const [attempt, setAttempt] = useState(0);
  // Filters are multi-select over the relations, and deliberately stay put when
  // the reader switches to another person to compare the two.
  const [kinds, setKinds] = useState<Set<RelationKind>>(() => new Set(RELATION_ORDER));
  const [gender, setGender] = useState<GenderFilter>("all");
  const [status, setStatus] = useState<StatusFilter>("all");
  const [query, setQuery] = useState("");
  const [sort, setSort] = useState<RelationSort>("closest");

  useEffect(() => {
    if (!id) return;
    // A superseded read is cancelled so a slow answer cannot replace a newer
    // one. `loaded` stays true across reloads to avoid a loading flash.
    const controller = new AbortController();
    loadRelatives(id, controller.signal).then((result) => {
      setReport(result);
      setLoadError("");
      setLoaded(true);
    }).catch((cause) => {
      if (isAbort(cause)) return;
      setLoadError(cause instanceof Error ? cause.message : "Could not read this family.");
      setLoaded(true);
    });
    return () => controller.abort();
  }, [id, attempt]);

  const candidates = useMemo<User[]>(() => people.map(personToUser), [people]);
  // Memoized so the two filter passes below depend on a stable reference rather
  // than re-running on every render of this page.
  const relatives = useMemo<Relative[]>(() => report?.relatives ?? [], [report]);

  /** How many of each relation exist, so the filters can show what is there. */
  const counts = useMemo(() => {
    const map = new Map<RelationKind, number>();
    for (const relative of relatives) map.set(relative.kind, (map.get(relative.kind) ?? 0) + 1);
    return map;
  }, [relatives]);

  const visible = useMemo(() => {
    const needle = query.trim().toLowerCase();
    return sortRelatives(relatives.filter((relative) =>
      kinds.has(relative.kind)
      && (gender === "all" || relative.user.gender === gender)
      && (status === "all" || (status === "alive" ? relative.user.alive : !relative.user.alive))
      && (!needle || relative.user.name.toLowerCase().includes(needle) || relative.user.id.toLowerCase().includes(needle))
    ), sort);
  }, [relatives, kinds, gender, status, query, sort]);

  // Only the relations this person actually has become sections, in kinship order.
  const groups = useMemo(() => RELATION_ORDER
    .map((kind) => ({ kind, items: visible.filter((relative) => relative.kind === kind) }))
    .filter((group) => group.items.length > 0), [visible]);

  const filtered = kinds.size !== RELATION_ORDER.length || gender !== "all" || status !== "all" || !!query.trim();
  const clearFilters = () => {
    setKinds(new Set(RELATION_ORDER));
    setGender("all");
    setStatus("all");
    setQuery("");
    setSort("closest");
  };
  const toggleKind = (kind: RelationKind) => setKinds((current) => {
    const next = new Set(current);
    if (next.has(kind)) next.delete(kind); else next.add(kind);
    return next;
  });
  const openPerson = (person: User) => navigate(`/people/${encodeURIComponent(person.id)}`);
  const selected = candidates.find((person) => person.id === id);
  const title = selected?.name ?? report?.person.name ?? id;

  return (
    <main className="person-details">
      <div className="page-toolbar">
        {/* Router links navigate without reloading App or its shared people data. */}
        <Button asChild variant="soft" color="gray" highContrast size="3" radius="full">
          <Link to="/"><ArrowLeft size={18} aria-hidden="true" />Back home</Link>
        </Button>
        {id && <Button asChild variant="soft" color="gray" highContrast size="3" radius="full">
          <Link to={`/people/${encodeURIComponent(id)}`}><UserRound size={18} aria-hidden="true" />Person page</Link>
        </Button>}
      </div>

      <header className="family-heading">
        <span className="family-eyebrow">EXTENDED FAMILY</span>
        <h1>{title ? `Family of ${title}` : "Relationships"}</h1>
        <p>Parents, siblings, grandparents, uncles, aunts, cousins, nieces and nephews — read from the same family records as the person pages.</p>
        <p className="family-notice">Pick anyone to see how they are related. Relations by marriage, such as an uncle's wife, are not counted as blood relatives.</p>
      </header>

      {/* .family-form so the picker's label and trigger keep the field styling
          they have in the dialogs. */}
      <section className="family-form relations-picker">
        <label>Whose relationships?
          <PersonPicker label="Person" value={selected ? selected.id : ""} placeholder="Choose a person"
            groups={[{ label: "Everyone", people: candidates }]}
            onChange={(next) => navigate(`/relationships/${encodeURIComponent(next)}`)} />
        </label>
      </section>

      {/* The picker above needs the shared list, so its load failure is reported here. */}
      {loading && <p role="status">Loading people…</p>}
      {error && <p role="alert">{error}</p>}

      {id && !loading && !error && (loaded ? (
        loadError ? <p role="alert">{loadError} <button type="button" className="filter-clear" onClick={() => setAttempt((value) => value + 1)}>Retry</button></p>
          : report === null ? <h2 className="relations-empty">Person not found</h2>
            : <>
              <div className="family-form filter-bar">
                <div className="relations-kinds" role="group" aria-label="Relations to show">
                  {RELATION_ORDER.map((kind) => (
                    <button key={kind} type="button" className="relations-kind" aria-pressed={kinds.has(kind)} onClick={() => toggleKind(kind)}>
                      {RELATION_PLURALS[kind]} <span>{counts.get(kind) ?? 0}</span>
                    </button>
                  ))}
                </div>
                <label>Gender
                  <select value={gender} onChange={(event) => setGender(event.target.value as GenderFilter)}>
                    <option value="all">All</option><option value="Male">Male</option><option value="Female">Female</option>
                  </select>
                </label>
                <label>Living
                  <select value={status} onChange={(event) => setStatus(event.target.value as StatusFilter)}>
                    <option value="all">All</option><option value="alive">Alive</option><option value="deceased">Deceased</option>
                  </select>
                </label>
                <label>Sort by
                  <select value={sort} onChange={(event) => setSort(event.target.value as RelationSort)}>
                    <option value="closest">Closest relation</option><option value="name">Name</option><option value="oldest">Oldest first</option><option value="youngest">Youngest first</option>
                  </select>
                </label>
                <label>Name
                  <input value={query} placeholder="Filter by name" onChange={(event) => setQuery(event.target.value)} />
                </label>
                {filtered && <button type="button" className="filter-clear" onClick={clearFilters}>Clear filters</button>}
              </div>

              <p className="filter-summary" role="status">
                {visible.length} of {relatives.length} {relatives.length === 1 ? "relative" : "relatives"}
                {filtered ? " shown" : ""}
              </p>

              {relatives.length === 0 ? (
                <p className="family-empty">No relatives are recorded for {report.person.name} yet. Add parents, a spouse or children from their person page and they appear here.</p>
              ) : groups.length === 0 ? (
                <p className="family-empty">No relatives match these filters.</p>
              ) : groups.map(({ kind, items }) => (
                <section className="person-grid-section" key={kind}>
                  <h2>{RELATION_PLURALS[kind]} <small>{items.length}</small></h2>
                  <div className="person-grid">
                    {items.map((relative: Relative) => <FamilyPersonCard key={relative.user.id} user={relative.user}
                      label={relative.via ? `${RELATION_LABELS[kind]} · via ${relative.via}` : RELATION_LABELS[kind]}
                      primaryAction="open" onPrimary={openPerson} />)}
                  </div>
                </section>
              ))}
            </>
      ) : <p role="status">Reading the family tree…</p>)}
    </main>
  );
}
