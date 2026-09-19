package personRouter

import (
	"encoding/json"
	"net/http"

	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

func handleGetAllPeople(driver neo4j.Driver) http.HandlerFunc {

	// 1. Capitalize fields so the JSON encoder can read them.
	// Use `json:"..."` tags to dictate exactly how they appear in the JSON output.
	type finalPeopleDesign struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Gender      string `json:"gender"`
		Alive       bool   `json:"alive"`
		DateOfBirth string `json:"dateOfBirth"`
		DateOfDeath string `json:"dateOfDeath"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		people, err := person.GetPersonList(ctx, driver)

		// 2. Move error handling UP. Check this immediately after the database call.
		if err != nil {
			http.Error(w, "Failed to retrieve people", http.StatusInternalServerError)
			return
		}

		// 3. Performance optimization: Pre-allocate the slice's capacity
		// since you already know how many items are in 'people'.
		allPeople := make([]finalPeopleDesign, 0, len(people))

		for _, p := range people {

			// 4. Simplified mapping. If DateOfBirth is "", converting it to a string
			// safely results in "", so you don't need the extra if-statements.
			allPeople = append(allPeople, finalPeopleDesign{
				ID:          p.Id,
				Name:        p.PersonName,
				Gender:      string(p.Gender),
				Alive:       p.Alive,
				DateOfBirth: string(p.DateOfBirth),
				DateOfDeath: string(p.DateOfDeath),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(allPeople)
	}
}
