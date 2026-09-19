import { useState } from "react";
import { Button } from "@radix-ui/themes";
import { ArrowLeft, Plus, UserRound } from "lucide-react";
import { Link, useNavigate } from "react-router";
import PersonCreatorDialog from "../components/PersonCreatorDialog";
import type { User } from "../helpers/personModel.ts";

interface Props {
  /** Asks App to re-read the people list so the new person reaches the palette. */
  onPeopleChanged: () => void;
}

/**
 * The "New Family" entry point. It is deliberately the person page's layout —
 * the same header, three sections and controls — rendered before any person
 * exists: every control is present but disabled, and the focal person's slot is
 * a placeholder whose button opens the create dialog. Creating there navigates
 * to the new person's page, which is the same layout with everything enabled.
 */
export default function NewFamily({ onPeopleChanged }: Props) {
  const navigate = useNavigate();
  const [creating, setCreating] = useState(false);
  // Every disabled control says what unlocks it, so a greyed-out page is not a mystery.
  const locked = "Available once the first person exists";
  const openPersonPage = (person: User) => {
    // Refresh the shared list first, so the palette and every selector can offer
    // the person the page is about to load.
    onPeopleChanged();
    navigate(`/people/${encodeURIComponent(person.id)}`);
  };
  return (
    <main className="person-details">
      {/* Kept enabled: leaving the page is not part of the disabled family view. */}
      <Button asChild variant="soft" color="gray" highContrast size="3" radius="full">
        <Link to="/"><ArrowLeft size={18} aria-hidden="true" />Back home</Link>
      </Button>
      <header className="family-heading">
        <span className="family-eyebrow">NEW FAMILY, NOT STARTED</span>
        <h1>New family</h1>
        <p>This is the person page before anyone exists. Every control is in place but disabled.</p>
        <p className="family-notice">Create the first person to start the family; the page then opens as their person page.</p>
      </header>
      <div className="family-layout">
        <section>
          <h2>Parents <small>the marriage they are a child of</small></h2>
          <div className="parents-grid">
            <button className="family-add" type="button" disabled><Plus aria-hidden="true" />Add parents<span>{locked}</span></button>
          </div>
        </section>
        <section>
          <h2>Person &amp; spouses</h2>
          {/* The focal card's slot: a dashed placeholder instead of a person. */}
          <div className="family-card">
            <div className="family-person-placeholder">
              <span className="family-card-top"><span className="family-avatar"><UserRound size={22} aria-hidden="true" /></span></span>
              <span className="family-eyebrow">Selected person</span>
              <strong>No person yet</strong>
              <span>Create the first person to start this family.</span>
              <Button type="button" onClick={() => setCreating(true)}><Plus size={16} aria-hidden="true" />Create new person</Button>
            </div>
          </div>
          <button className="family-add" type="button" disabled><Plus aria-hidden="true" />Add spouse<span>{locked}</span></button>
          <button className="family-add" type="button" disabled><Plus aria-hidden="true" />Add child<span>{locked}</span></button>
        </section>
        <section>
          <h2>Children <small>0 · Oldest first</small></h2>
          <div className="children-grid" />
          <p className="family-empty">Create the first person, then add a spouse before adding children.</p>
        </section>
      </div>
      {creating && <PersonCreatorDialog close={() => setCreating(false)} onCreated={openPersonPage} />}
    </main>
  );
}
