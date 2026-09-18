import { Button } from "@radix-ui/themes";
import { ArrowLeft } from "lucide-react";
import FamilyView from "../components/FamilyView";
import { Link, useParams } from "react-router";
import type { Person } from "../api/people.ts";

/** Data is owned by App and shared with the palette to avoid duplicate requests. */
interface PersonDetailsProps {
  /** The current people collection, used by the palette and the selectors. */
  people: Person[];
  /** Asks App to re-read the people list after a create or a rename. */
  onPeopleChanged: () => void;
  /** True while the shared people list is loading for the first time. */
  loading: boolean;
  /** A user-facing failure from the shared people load, or null. */
  error: string | null;
}

/**
 * Displays the person identified by the /people/:id route. Any person in the
 * tree — the one from search, a spouse, a child or a parent — is loaded the same
 * way, and FamilyView owns the not-found state because it holds the payload.
 */
export default function PersonDetails({ people, loading, error, onPeopleChanged }: PersonDetailsProps) {
  const { id } = useParams();

  return (
    <main className="person-details">
      {/* Router links navigate without reloading App or its shared people data. */}
      <Button asChild variant="soft" color="gray" highContrast size="3" radius="full">
        <Link to="/"><ArrowLeft size={18} aria-hidden="true" />Back home</Link>
      </Button>
      {/* Resolve the shared people load before rendering: an empty pending list
          is not evidence that the requested person does not exist. Live-region
          roles announce asynchronous status changes to assistive technology. */}
      {loading ? <p role="status">Loading person…</p> : error ? <p role="alert">{error}</p> : id ? (
        <FamilyView key={id} id={id} people={people} onPeopleChanged={onPeopleChanged} />
      ) : <h1 className="mt-6 text-2xl">Person not found</h1>}
    </main>
  );
}
