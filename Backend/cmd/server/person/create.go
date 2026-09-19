package personRouter

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type createUserRequest struct {
	PersonName  string
	Gender      string
	DateOfBirth string
	DateOfDeath string
	Alive       bool
}

type createUserResponse struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	DateOfBirth string `json:"dateOfBirth"`
	DateOfDeath string `json:"dateOfDeath"`
	Gender      string `json:"gender"`
	Alive       bool   `json:"alive"`
}

func handleCreatePerson(driver neo4j.Driver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Must Be POST", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		var req createUserRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			log.Printf("JSON decode error: %v", err)
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
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
			http.Error(w, "Unable To Create New User", http.StatusBadRequest)
			return
		}

		personDetail, err := person.GetPerson(ctx, driver, uid)
		if err != nil || personDetail == nil {
			http.Error(w, "Unable To Get The Person", http.StatusBadRequest)
			person.DeleteUser(ctx, driver, uid)
			return
		}

		response := createUserResponse{
			Alive:       personDetail.Alive,
			Gender:      string(personDetail.Gender),
			DateOfDeath: string(personDetail.DateOfDeath),
			DateOfBirth: string(personDetail.DateOfBirth),
			Name:        personDetail.PersonName,
			Id:          personDetail.Id,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}
