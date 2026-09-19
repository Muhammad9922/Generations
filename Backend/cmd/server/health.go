package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// handleHealth answers the container's liveness probe.
//
// It deliberately does not touch the database. A slow or unreachable graph must
// not make the process look dead and get restarted — the API can still answer
// for itself — and a request that reaches this handler has already proved the
// HTTP server is accepting connections.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("health check could not write its response: %v", err)
	}
}
