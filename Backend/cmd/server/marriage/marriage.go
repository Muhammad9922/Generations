// Package marriageRouter serves the marriage endpoints the family page calls:
// creating a marriage, moving its dates, disbanding it, and adding or removing
// the child links that hang off it.
//
// It is the sibling of cmd/server/person: one file per operation, one
// registration function, and the same two JSON envelopes.
package marriageRouter

import (
	"net/http"

	"github.com/Muhammad9922/Generations/cmd/server/apiresponse"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// RegisterMarriageRoutes hangs every marriage endpoint off the shared mux.
//
// The paths carry no "/api" prefix: the Vite dev proxy strips it before the
// request arrives (API_SCOPE.md §2). The matching client-side paths live in
// Frontend/src/api/contracts.ts, so a rename is one edit on each side.
func RegisterMarriageRoutes(mux *http.ServeMux, driver neo4j.Driver) {
	mux.HandleFunc("POST /marriages", handleCreateMarriage(driver))
	mux.HandleFunc("PATCH /marriages/{id}", handleUpdateMarriage(driver))
	mux.HandleFunc("DELETE /marriages/{id}", handleDeleteMarriage(driver))
	mux.HandleFunc("POST /marriages/{id}/children", handleAddChild(driver))
	mux.HandleFunc("DELETE /marriages/{id}/children/{childId}", handleRemoveChild(driver))
}

// writeJSON and writeError are the two envelopes personRouter answers with as
// well. Both defer to apiresponse so the routers cannot drift apart: the
// frontend reads the failure message out of {"error": "..."}, and http.Error's
// text/plain body would be reported as "did not answer with JSON" instead.
func writeJSON(w http.ResponseWriter, status int, body any) {
	apiresponse.Write(w, status, body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	apiresponse.Error(w, status, message)
}
