package marriageRouter

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Muhammad9922/Generations/integration"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// createMarriageRequest is the body of POST /marriages (API_SCOPE.md §4.5).
//
// One endpoint serves "add spouse", "add parents" and "add child while the other
// parent is not a spouse yet", because all three are the same operation: create
// a marriage, optionally with children. `childrenIds` is absent when the dialog
// had none to pass, which is what the frontend does.
type createMarriageRequest struct {
	SpouseOne   string
	SpouseTwo   string
	DateStart   string
	DateEnd     string
	ChildrenIds []string `json:"childrenIds"`
}

// childrenAsNewPeople turns the child ids into the NewPerson values CreateFamily
// expects. Only Id is filled in: CreateFamily checks existence first and reuses
// the person instead of creating anyone.
func childrenAsNewPeople(ids []string) []person.NewPerson {
	children := make([]person.NewPerson, 0, len(ids))

	for _, id := range ids {
		children = append(children, person.NewPerson{Id: id})
	}

	return children
}

func handleCreateMarriage(driver neo4j.Driver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// The route is registered as "POST /marriages", so ServeMux already
		// answers any other method with 405.
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		var req createMarriageRequest
		decoder := json.NewDecoder(r.Body)
		// Mirrors POST /people: unrecognised fields are a mistake (for example
		// the lowercase "spouseOne" of a certificate) and ignoring them would
		// look like a marriage that was created with no spouses at all.
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			log.Printf("JSON decode error: %v", err)
			writeError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		ctx := r.Context()

		if req.SpouseOne == "" || req.SpouseTwo == "" {
			writeError(w, http.StatusBadRequest, "Both spouses must be provided")
			return
		}

		// Check every id before anything is written, so a marriage is never
		// created around a person who does not exist.
		if err := requirePeople(ctx, driver, append([]string{req.SpouseOne, req.SpouseTwo}, req.ChildrenIds...)...); err != nil {
			respondLookupError(w, err, true)
			return
		}

		// CreateFamily does the whole job in one call, including rolling back the
		// people it created if the marriage or a child link fails. Its test suite
		// already covers new spouses, existing spouses, children and attaching to
		// an existing marriage.
		detail, err := integration.CreateFamily(ctx, driver,
			person.NewPerson{Id: req.SpouseOne},
			person.NewPerson{Id: req.SpouseTwo},
			childrenAsNewPeople(req.ChildrenIds),
			integration.MarriageOptionalParams{
				DateStart: req.DateStart,
				DateEnd:   req.DateEnd,
				Id:        "", // let the package generate the id
			},
		)
		if err != nil {
			// CreateNewMarriage names the rule it rejected — two spouses of the
			// same gender, a marriage before either birth, an end after a death,
			// a malformed date — and that sentence is what the dialog shows.
			log.Printf("Unable To Create The Marriage: %v", err)
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, map[string]string{"id": detail.MarriageID})
	}
}
