package personRouter

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

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

		var body person.UpdateUser
		decoder := json.NewDecoder(r.Body)
		// Mirrors POST /people: a body with unrecognised fields is a mistake
		// (for example snake_case "date_of_birth"), and silently ignoring it
		// would look like a successful update that changed nothing.
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		_, _, err := person.UpdatePerson(ctx, driver, uid, body)

		if err != nil {
			log.Printf("Error While Updating User: %q", err)

			switch {
			case errors.Is(err, person.ErrInvalidUpdate):
				writeError(w, http.StatusBadRequest, "Unable To Update The Person")
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(finalPeopleDesign{
			ID:          saved.Id,
			Name:        saved.PersonName,
			Gender:      string(saved.Gender),
			Alive:       saved.Alive,
			DateOfBirth: string(saved.DateOfBirth),
			DateOfDeath: string(saved.DateOfDeath),
		})
	}
}
