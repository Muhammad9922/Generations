import { Link, useParams } from "react-router";
import type { Person } from "../helpers/GetUsers";

/** Data is owned by App and shared with the palette to avoid duplicate requests. */
interface PersonDetailsProps {
  /** The current people collection, including the person selected in search. */
  people: Person[];
  /** True while the initial load or a retry is pending. */
  loading: boolean;
  /** A user-facing load failure, or null when no failure has occurred. */
  error: string | null;
}

/**
 * Displays the person identified by the /people/:id route.
 * Selecting a palette result and opening a details URL directly use the same
 * lookup. This page does not fetch data or own the global search/theme state.
 */
export default function PersonDetails({ people, loading, error }: PersonDetailsProps) {
  // Match stable IDs rather than names, since different people can share a name.
  // An absent or unknown route ID safely produces the not-found state below.
  const { id } = useParams();
  const person = people.find((item) => item.id === id);

  return (
    <main className="mx-auto max-w-xl p-8">
      {/* Router links navigate without reloading App or its shared people data. */}
      <Link to="/">← Back home</Link>
      {/* Resolve loading and errors before not-found: an empty pending list is
          not evidence that the requested person does not exist. Live-region
          roles announce asynchronous status changes to assistive technology. */}
      {loading ? <p role="status">Loading person…</p> : error ? <p role="alert">{error}</p> : person ? (
        <section className="mt-6">
          <h1 className="text-3xl font-bold">{person.name}</h1>
          <p className="mt-4">Father: {person.personFatherName}</p>
          <p>ID: {person.id}</p>
          {/* Keep the current data-source limitation visible to users. */}
          <p className="mt-4 text-sm">Sample person — backend data is not connected yet.</p>
        </section>
      ) : <h1 className="mt-6 text-2xl">Person not found</h1>}
    </main>
  );
}
