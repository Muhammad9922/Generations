package personRouter

import (
	"net/http"

	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// handleGetSingles answers GET /people/singles with everyone who is not a spouse
// in any marriage.
//
// It replies with the same envelope as GET /people, so the frontend reads both
// lists through one shape. Being single is a graph fact rather than a property
// on the node, which is why it needs its own query: answering it from the list
// endpoint would mean reading every person's marriages first.
func handleGetSingles(driver neo4j.Driver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		singles, err := person.GetSingleList(r.Context(), driver)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to retrieve singles")
			return
		}

		// Pre-allocate: the size is known, and the page renders the whole list.
		allSingles := make([]finalPeopleDesign, 0, len(singles))

		for _, single := range singles {
			allSingles = append(allSingles, toWirePerson(&single))
		}

		writeJSON(w, http.StatusOK, struct {
			People []finalPeopleDesign `json:"people"`
		}{People: allSingles})
	}
}
