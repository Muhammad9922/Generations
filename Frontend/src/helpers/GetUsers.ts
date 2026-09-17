import type { Action } from "kbar";

/** Shared display model used by search actions and the person details page. */
export interface Person {
    /** Stable identity for route lookup and action registration; names may repeat. */
    id: string;
    /** Display name, left unchanged in the underlying data. */
    name: string;
    /** Additional search text and context for distinguishing identical names. */
    personFatherName: string;
}

/**
 * Returns sample data until the backend exposes a people endpoint.
 * The asynchronous contract lets App handle loading, failures and retries without
 * coupling UI components to the eventual transport. No HTTP request is made yet.
 */
export async function GetAllUsers(): Promise<Person[]> {

    const people = [
        {
            personFatherName: "Muhammad Saleem",
            name: "Mahammad Muhayodin",
            id: "id-1"
        },
        {
            personFatherName: "Tariq Mahmood",
            name: "Hamza Tariq",
            id: "id-2"
        },
        {
            personFatherName: "Abdul Rahman",
            name: "Usman Abdul",
            id: "id-3"
        },
        {
            personFatherName: "Bilal Ahmed",
            name: "Zaid Bilal",
            id: "id-4"
        },
        {
            personFatherName: "Rashid Khan",
            name: "Omar Rashid",
            id: "id-5"
        },
        {
            personFatherName: "Abdul Rauf",
            name: "Usman Abdul",
            id: "id-6"
        },

    ];

    return people;
}

/**
 * Converts data into Kbar actions without mutating the people collection.
 * Selection is injected so this helper stays independent of React Router and can
 * be tested directly. Empty collections naturally produce an empty action list.
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
        // Always show father context and include it in Kbar's searchable metadata.
        subtitle: `Father: ${person.personFatherName}`,
        keywords: `${person.name} ${person.personFatherName} ${person.id}`,
        // Expand only duplicate labels; the original person.name is unchanged.
        name: (nameCounts.get(person.name) ?? 0) > 1
            ? `${person.name} son of ${person.personFatherName}`
            : person.name
    }));
}