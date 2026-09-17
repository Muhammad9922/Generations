
interface User {
    name: string,
    id: string,
}

export async function GetAllUsers() : Promise<User[]> {

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
            personFatherName: "Abdul Rahman",
            name: "Umar Aslam",
            id: "id-6"
        },

    ];

    const duplicatePeople: string[] = people.filter((p, i) => people.findIndex(x => x.name === p.name) !== i).map(v => v.id);
    const finalPeople = people.map(px => {
        const returnableObject = px
        if (duplicatePeople.includes(px.id)){
            returnableObject["name"] = returnableObject.name + " son of " + returnableObject.personFatherName
        }
        return returnableObject 
    })

    return finalPeople

    const response = await fetch("http://localhost:8080/api/v1/person/list", {
        method:"GET",
        headers: {
            "accept": "application/json"
        }
    }).then(res => res.json())
    .then(res => res["users"])
    .catch(res => {
        console.error(res)
        return null
    })

    return response
}