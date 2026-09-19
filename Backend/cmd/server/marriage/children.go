package marriageRouter

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Muhammad9922/Generations/internal/children"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// addChildRequest is the body of POST /marriages/{id}/children.
type addChildRequest struct {
	ChildId string `json:"childId"`
}

// handleAddChild links an existing person as a child of a marriage. It backs
// "Actions → Add child" once the other parent is already a spouse; when they are
// not, the dialog calls POST /marriages instead and passes the child there.
func handleAddChild(driver neo4j.Driver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		marriageId := r.PathValue("id")

		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		var req addChildRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		if req.ChildId == "" {
			writeError(w, http.StatusBadRequest, "A Child Must Be Provided")
			return
		}

		// Both ids sit in the request itself, so an unknown one is a 404 rather
		// than a 500 — and the client then knows which record went stale.
		if err := requireMarriage(ctx, driver, marriageId); err != nil {
			respondLookupError(w, err, false)
			return
		}

		if err := requirePeople(ctx, driver, req.ChildId); err != nil {
			respondLookupError(w, err, false)
			return
		}

		if _, err := children.CreateNewChild(ctx, driver, marriageId, req.ChildId); err != nil {
			log.Printf("Unable To Add Child %q To Marriage %q: %v", req.ChildId, marriageId, err)
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// The frontend re-reads the person afterwards, so a bare true is all the
		// caller needs.
		writeJSON(w, http.StatusOK, true)
	}
}

// handleRemoveChild removes one child link from a marriage. It also backs
// "Delete parents", which only delinks this person from their parents' marriage:
// the two parents stay married to each other.
func handleRemoveChild(driver neo4j.Driver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		marriageId := r.PathValue("id")
		childId := r.PathValue("childId")

		if marriageId == "" || childId == "" {
			writeError(w, http.StatusBadRequest, "A Marriage And A Child Must Be Provided")
			return
		}

		// Mind the argument order: the route reads {id} as the marriage and
		// {childId} as the person, but the package takes the person first.
		//
		// A link that is not there comes back as an error, and it is worth
		// reporting rather than swallowing: it means the page is showing a
		// relationship the database no longer holds.
		if _, err := children.DeleteChildren(ctx, driver, childId, marriageId); err != nil {
			log.Printf("Unable To Remove Child %q From Marriage %q: %v", childId, marriageId, err)
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, true)
	}
}
