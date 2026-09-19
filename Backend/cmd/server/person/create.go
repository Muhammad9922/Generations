package personRouter

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// createUserRequest is the body of POST /people, spelled with the Go field names
// of person.NewPerson (API_SCOPE.md §2) so it decodes straight into it.
type createUserRequest struct {
	PersonName  string
	Gender      string
	DateOfBirth string
	DateOfDeath string
	Alive       bool
}

func handleCreatePerson(driver neo4j.Driver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// The route is registered as "POST /people", so ServeMux already answers
		// any other method with 405.
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		var req createUserRequest
		decoder := json.NewDecoder(r.Body)
		// A body carrying unrecognised fields is a mistake — for example the
		// lowercase keys of a certificate — and ignoring them silently would look
		// like a create that dropped half the person.
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			log.Printf("JSON decode error: %v", err)
			writeError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		ctx := r.Context()

		_, uid, err := person.CreateNewPerson(ctx, driver, person.NewPerson{
			PersonName:  req.PersonName,
			Gender:      person.Gender(req.Gender),
			DateOfBirth: person.DateProper(req.DateOfBirth),
			DateOfDeath: person.DateProper(req.DateOfDeath),
			Alive:       req.Alive,
		})

		if err != nil {
			// CreateNewPerson names the rule it rejected the person over — a
			// missing name, an invalid gender, an impossible date. API_SCOPE.md
			// §4.3 asks for that sentence rather than a generic one, because the
			// dialog shows whatever it is given.
			log.Printf("Unable To Create A New Person: %v", err)
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Read the record back rather than echoing the request: the stored dates
		// are the canonical DD-MM-YYYY ones, and the client must never invent the
		// ID it is about to use for every later call.
		personDetail, err := person.GetPerson(ctx, driver, uid)
		if err != nil || personDetail == nil {
			log.Printf("Unable To Read Back The Created Person %q: %v", uid, err)
			// The person exists but cannot be read, so leaving them behind would
			// hand the client an ID it never learned about.
			if _, _, deleteErr := person.DeleteUser(ctx, driver, uid); deleteErr != nil {
				log.Printf("Unable To Roll Back The Created Person %q: %v", uid, deleteErr)
			}
			writeError(w, http.StatusInternalServerError, "Unable To Read Back The New Person")
			return
		}

		writeJSON(w, http.StatusCreated, finalPeopleDesign{
			ID:          personDetail.Id,
			Name:        personDetail.PersonName,
			Gender:      string(personDetail.Gender),
			Alive:       personDetail.Alive,
			DateOfBirth: string(personDetail.DateOfBirth),
			DateOfDeath: string(personDetail.DateOfDeath),
		})
	}
}
