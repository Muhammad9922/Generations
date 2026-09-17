import { Button } from "@radix-ui/themes";
import { ArrowLeft } from "lucide-react";
import FamilyView from "../components/FamilyView";
import { Link, useParams } from "react-router";
import type { Person } from "../helpers/GetUsers";

/** Data is owned by App and shared with the palette to avoid duplicate requests. */
interface PersonDetailsProps {
  /** The current people collection, including the person selected in search. */
  people: Person[];
  onRename: (id: string, name: string) => void;
  /** True while the initial load or a retry is pending. */
  loading: boolean;
  /** A user-facing load failure, or null when no failure has occurred. */
  error: string | null;
}

/**
 * Displays the person identified by the /people/:id route.
 * Selecting a palette result and opening a details URL directly use the same
 * lookup. FamilyView loads example relationships; App owns search/theme state.
 */
export default function PersonDetails({ people, loading, error, onRename }: PersonDetailsProps) {
  // Match stable IDs rather than names, since different people can share a name.
  // An absent or unknown route ID safely produces the not-found state below.
  const { id } = useParams();
  const person = people.find((item) => item.id === id);

  return (
    <main className="person-details">
      {/* Router links navigate without reloading App or its shared people data. */}
      <Button asChild variant="soft" color="gray" highContrast size="3" radius="full">
        <Link to="/"><ArrowLeft size={18} aria-hidden="true" />Back home</Link>
      </Button>
      {/* Resolve loading and errors before not-found: an empty pending list is
          not evidence that the requested person does not exist. Live-region
          roles announce asynchronous status changes to assistive technology. */}
      {loading ? <p role="status">Loading person…</p> : error ? <p role="alert">{error}</p> : person ? (
        <FamilyView key={person.id} id={person.id} people={people} onRename={onRename} />
      ) : <h1 className="mt-6 text-2xl">Person not found</h1>}
    </main>
  );
}
