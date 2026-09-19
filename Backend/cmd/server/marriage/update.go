package marriageRouter

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Muhammad9922/Generations/cmd/server/apiresponse"
	"github.com/Muhammad9922/Generations/internal/marriage"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// updateMarriageRequest is the body of PATCH /marriages/{id}. It never carries
// spouses or children: the people in a marriage are fixed once it exists,
// because swapping a spouse would break every child link hanging off it.
//
// An empty string clears that date, which is the promise the dates dialog makes
// when it says "Clear a field to remove its date". The dialog always submits
// both fields, so there is no third "leave it alone" case to spell here.
type updateMarriageRequest struct {
	DateStart string
	DateEnd   string
}

func handleUpdateMarriage(driver neo4j.Driver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("id")

		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		var req updateMarriageRequest
		decoder := json.NewDecoder(r.Body)
		// Mirrors the other writes: "date_start" or a spouse field is a mistake
		// the client should hear about, not something to silently drop.
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		if err := requireMarriage(ctx, driver, id); err != nil {
			respondLookupError(w, err, false)
			return
		}

		_, err := marriage.UpdateMarriageDates(ctx, driver, id, marriage.MarriageUpdate{
			DateStart: person.Ptr(person.DateProper(req.DateStart)),
			DateEnd:   person.Ptr(person.DateProper(req.DateEnd)),
		})

		if err != nil {
			log.Printf("Unable To Update The Marriage %q: %v", id, err)

			switch {
			case errors.Is(err, marriage.ErrInvalidUpdate):
				// The package names the date it rejected and why, and that
				// sentence is what the dialog shows.
				writeError(w, http.StatusBadRequest, apiresponse.UserMessage(err, marriage.ErrInvalidUpdate))
			case errors.Is(err, marriage.ErrNotFound):
				writeError(w, http.StatusNotFound, "Marriage not found")
			default:
				writeError(w, http.StatusInternalServerError, "Unable To Update The Marriage")
			}
			return
		}

		// The page re-reads the person after every write, so no body content is
		// needed here.
		writeJSON(w, http.StatusOK, true)
	}
}
