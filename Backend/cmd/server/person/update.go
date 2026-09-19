package personRouter

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/Muhammad9922/Generations/cmd/server/apiresponse"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// updatePersonRequest is the body of PATCH /people/{id}. It exists for one
// reason: telling "the field was not sent" apart from "the field was sent as
// null". Both decode to a nil pointer in person.UpdateUser, but the frontend
// clears a birth or death date with an explicit null (API_SCOPE.md §4.4), and
// those two cases must not be the same.
//
// The dates are json.RawMessage because it keeps the literal `null`, which a
// *string would swallow into the same nil as an absent field.
type updatePersonRequest struct {
	Name        *string         `json:"Name"`
	Gender      *string         `json:"Gender"`
	Alive       *bool           `json:"Alive"`
	DateOfBirth json.RawMessage `json:"DateOfBirth"`
	DateOfDeath json.RawMessage `json:"DateOfDeath"`
}

// optionalDate reads one date out of the decoded body:
//
//	absent  → nil     leave the stored date alone
//	null    → &""     clear the date
//	"value" → &value  store it (validated by person.UpdatePerson)
func optionalDate(raw json.RawMessage) (*person.DateProper, error) {
	switch trimmed := strings.TrimSpace(string(raw)); trimmed {
	case "":
		return nil, nil
	case "null":
		return person.Ptr(person.DateProper("")), nil
	default:
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, errors.New("a date must be a DD-MM-YYYY string or null")
		}
		return person.Ptr(person.DateProper(value)), nil
	}
}

func handleUpdatePerson(driver neo4j.Driver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// The route is registered as "PATCH /people/{id}", so ServeMux already
		// answers any other method with 405.
		uid := r.PathValue("id")

		if uid == "" {
			writeError(w, http.StatusBadRequest, "Id Must Be Provided")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		var req updatePersonRequest
		decoder := json.NewDecoder(r.Body)
		// Mirrors POST /people: a body with unrecognised fields is a mistake
		// (for example snake_case "date_of_birth"), and silently ignoring it
		// would look like a successful update that changed nothing.
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		body := person.UpdateUser{Name: req.Name, Alive: req.Alive}

		if req.Gender != nil {
			gender := person.Gender(*req.Gender)
			body.Gender = &gender
		}

		var err error

		body.DateOfBirth, err = optionalDate(req.DateOfBirth)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		body.DateOfDeath, err = optionalDate(req.DateOfDeath)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		_, _, err = person.UpdatePerson(ctx, driver, uid, body)

		if err != nil {
			log.Printf("Error While Updating User: %q", err)

			switch {
			case errors.Is(err, person.ErrInvalidUpdate):
				// The package names the field it rejected, and that sentence is
				// what the dialog shows, so it is forwarded rather than replaced.
				writeError(w, http.StatusBadRequest, apiresponse.UserMessage(err, person.ErrInvalidUpdate))
			case errors.Is(err, person.ErrNotFound):
				writeError(w, http.StatusNotFound, "Person not found")
			default:
				writeError(w, http.StatusInternalServerError, "Error While Updating")
			}
			return
		}

		// API_SCOPE.md §4.4: answer with the saved person, which is what the
		// frontend reads the id from.
		saved, err := person.GetPerson(ctx, driver, uid)
		if err != nil || saved == nil {
			log.Printf("Error While Reading Back Updated User %q: %v", uid, err)
			writeError(w, http.StatusInternalServerError, "Unable To Get The Person")
			return
		}

		writeJSON(w, http.StatusOK, finalPeopleDesign{
			ID:          saved.Id,
			Name:        saved.PersonName,
			Gender:      string(saved.Gender),
			Alive:       saved.Alive,
			DateOfBirth: string(saved.DateOfBirth),
			DateOfDeath: string(saved.DateOfDeath),
		})
	}
}
