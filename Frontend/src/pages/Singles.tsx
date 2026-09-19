import { useEffect, useMemo, useState } from "react";
import { Button } from "@radix-ui/themes";
import { ArrowLeft } from "lucide-react";
import { Link, useNavigate } from "react-router";
import FamilyPersonCard from "../components/FamilyPersonCard";
import { getSingles } from "../api/people.ts";
import { isAbort } from "../api/http.ts";
import {
  compareByBirthAsc, compareByBirthDesc, compareNames, personToUser, type User,
} from "../helpers/personModel.ts";

type GenderFilter = "all" | User["gender"];
type StatusFilter = "all" | "alive" | "deceased";
type SortOrder = "name" | "oldest" | "youngest";

/**
 * Everyone without a marriage, with filters over the list.
 *
 * "Single" is a graph question — whether a MARRIED edge exists — so the API
 * answers it (see `getSingles`); this page only filters what it is given. A
 * widowed spouse is not listed, because the marriage that made them a spouse is
 * still recorded.
 */
export default function Singles() {
  const navigate = useNavigate();
  const [singles, setSingles] = useState<User[]>([]);
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState("");
  // Bumped by Retry. The list is read-only, so nothing else reloads it.
  const [attempt, setAttempt] = useState(0);
  const [gender, setGender] = useState<GenderFilter>("all");
  const [status, setStatus] = useState<StatusFilter>("all");
  const [query, setQuery] = useState("");
  const [sort, setSort] = useState<SortOrder>("name");

  useEffect(() => {
    // A superseded read is cancelled so a slow answer cannot replace a newer one.
    const controller = new AbortController();
    getSingles(controller.signal).then((people) => {
      setSingles(people.map(personToUser));
      setError("");
      setLoaded(true);
    }).catch((cause) => {
      if (isAbort(cause)) return;
      setError(cause instanceof Error ? cause.message : "Could not load the singles list.");
      setLoaded(true);
    });
    return () => controller.abort();
  }, [attempt]);

  const visible = useMemo(() => {
    const needle = query.trim().toLowerCase();
    const filtered = singles.filter((person) =>
      (gender === "all" || person.gender === gender)
      && (status === "all" || (status === "alive" ? person.alive : !person.alive))
      && (!needle || person.name.toLowerCase().includes(needle) || person.id.toLowerCase().includes(needle)));
    const compare = sort === "oldest" ? compareByBirthAsc : sort === "youngest" ? compareByBirthDesc : compareNames;
    return [...filtered].sort(compare);
  }, [singles, gender, status, query, sort]);

  const filtered = gender !== "all" || status !== "all" || !!query.trim();
  const clearFilters = () => {
    setGender("all");
    setStatus("all");
    setQuery("");
    setSort("name");
  };

  return (
    <main className="person-details">
      <div className="page-toolbar">
        {/* Router links navigate without reloading App or its shared people data. */}
        <Button asChild variant="soft" color="gray" highContrast size="3" radius="full">
          <Link to="/"><ArrowLeft size={18} aria-hidden="true" />Back home</Link>
        </Button>
      </div>

      <header className="family-heading">
        <span className="family-eyebrow">WITHOUT A MARRIAGE</span>
        <h1>List Of Singles</h1>
        <p>Everyone the family records give no marriage to, read from the same records as the person pages.</p>
        <p className="family-notice">Someone whose marriage was disbanded appears here; a widow or widower does not, because their marriage is still recorded.</p>
      </header>

      {!loaded ? <p role="status">Reading the family records…</p>
        : error ? <p role="alert">{error} <button type="button" className="filter-clear" onClick={() => setAttempt((value) => value + 1)}>Retry</button></p>
          : singles.length === 0 ? <p className="family-empty">No single people are recorded yet.</p>
            : <>
              <div className="family-form filter-bar">
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
                  <select value={sort} onChange={(event) => setSort(event.target.value as SortOrder)}>
                    <option value="name">Name</option><option value="oldest">Oldest first</option><option value="youngest">Youngest first</option>
                  </select>
                </label>
                <label>Name
                  <input value={query} placeholder="Filter by name" onChange={(event) => setQuery(event.target.value)} />
                </label>
                {filtered && <button type="button" className="filter-clear" onClick={clearFilters}>Clear filters</button>}
              </div>

              <p className="filter-summary" role="status">
                {visible.length} of {singles.length} {singles.length === 1 ? "single" : "singles"}
                {filtered ? " shown" : ""}
              </p>

              {visible.length === 0 ? <p className="family-empty">No single people match these filters.</p> : (
                <div className="person-grid">
                  {visible.map((person) => <FamilyPersonCard key={person.id} user={person} label="Single"
                    primaryAction="open" onPrimary={(chosen) => navigate(`/people/${encodeURIComponent(chosen.id)}`)} />)}
                </div>
              )}
            </>}
    </main>
  );
}
