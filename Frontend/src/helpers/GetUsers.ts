import type { Action } from "kbar";


export async function GetAllUsers() : Promise<Action[]> {

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

    const nameCounts = new Map<string, number>();
    for (const person of people) {
        nameCounts.set(person.name, (nameCounts.get(person.name) ?? 0) + 1);
    }

    return people.map((person): Action => ({
        perform: () => {
            console.error("TODO: Redirect Logic");
        },
        id: person.id,
        name: (nameCounts.get(person.name) ?? 0) > 1
            ? `${person.name} son of ${person.personFatherName}`
            : person.name
    }));
}