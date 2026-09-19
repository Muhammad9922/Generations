package personRouter

import (
	"net/http"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

func RegisterPersonRoutes(mux *http.ServeMux, driver neo4j.Driver) {
	mux.HandleFunc("GET /people", handleGetAllPeople(driver))
	// The literal path wins over the /people/{id} wildcard however they are
	// ordered, so no person id can shadow it.
	mux.HandleFunc("GET /people/singles", handleGetSingles(driver))
	mux.HandleFunc("GET /people/{id}", handleGetPerson(driver))
	mux.HandleFunc("POST /people", handleCreatePerson(driver))
	mux.HandleFunc("PATCH /people/{id}", handleUpdatePerson(driver))
}
