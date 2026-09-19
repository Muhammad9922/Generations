package marriageRouter

import (
	"log"
	"net/http"

	"github.com/Muhammad9922/Generations/internal/marriage"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// handleDeleteMarriage disbands one marriage, which is how a spouse is delinked.
//
// The spouses and the children all stay people; only the marriage node and its
// relationships go. A child of that marriage therefore loses the link that made
// those two their parents, which is exactly what the confirmation dialog
// promises when it counts how many children are "left without parents".
func handleDeleteMarriage(driver neo4j.Driver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("id")

		if err := requireMarriage(ctx, driver, id); err != nil {
			respondLookupError(w, err, false)
			return
		}

		if _, err := marriage.DeleteMarriage(ctx, driver, id); err != nil {
			// The id was there a moment ago, so a failure here is ours rather
			// than the client's.
			log.Printf("Unable To Delete The Marriage %q: %v", id, err)
			writeError(w, http.StatusInternalServerError, "Unable To Disband The Marriage")
			return
		}

		writeJSON(w, http.StatusOK, true)
	}
}
