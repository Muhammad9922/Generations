import type { Action } from "kbar";
import type { Person } from "../api/contracts.ts";

/**
 * Converts data into Kbar actions without mutating the people collection.
 * Selection is injected so this helper stays independent of React Router and can
 * be tested directly. Empty collections naturally produce an empty action list.
 *
 * A father's name is optional on the list endpoint, so it is used only when the
 * API supplies one: a label never falls back to the string "undefined".
 */
export function createPersonActions(people: Person[], onSelect: (person: Person) => void): Action[] {
  // Count exact display-name duplicates once instead of rescanning per person.
  const nameCounts = new Map<string, number>();
  for (const person of people) {
    nameCounts.set(person.name, (nameCounts.get(person.name) ?? 0) + 1);
  }

  return people.map((person): Action => ({
    // Capture the full record so selection does not rely on a display label.
    perform: () => onSelect(person),
    // Namespace stable IDs to avoid collisions with built-in command IDs.
    id: `person-${person.id}`,
    section: "People",
    subtitle: person.personFatherName ? `Father: ${person.personFatherName}` : undefined,
    keywords: [person.name, person.personFatherName, person.id].filter(Boolean).join(" "),
    // Expand only duplicate labels, and only when there is a father to name.
    name: (nameCounts.get(person.name) ?? 0) > 1 && person.personFatherName
      ? `${person.name} son of ${person.personFatherName}`
      : person.name
  }));
}
