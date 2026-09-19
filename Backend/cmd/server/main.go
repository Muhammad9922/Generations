package main

import (
	"log"
	"net/http"
	"os"
	"time"

	marriageRouter "github.com/Muhammad9922/Generations/cmd/server/marriage"
	personRouter "github.com/Muhammad9922/Generations/cmd/server/person"
	"github.com/Muhammad9922/Generations/internal/db"
)

// defaultDatabaseURI is the Neo4j instance the packages' own tests point at.
// NEO4J_URI overrides it, so the credential-free instance used in development
// stays the zero-config default while every other deployment keeps its address
// out of the source.
const defaultDatabaseURI = "bolt://192.168.0.133:7687"

// defaultListenAddress is the port the Vite dev proxy expects (vite.config.ts
// forwards /api to http://localhost:8080). API_ADDR overrides it, which is what
// lets a second instance run beside one already holding the port.
const defaultListenAddress = ":8080"

func main() {
	databaseURI := os.Getenv("NEO4J_URI")
	if databaseURI == "" {
		databaseURI = defaultDatabaseURI
	}

	listenAddress := os.Getenv("API_ADDR")
	if listenAddress == "" {
		listenAddress = defaultListenAddress
	}

	ctx, driver := db.ConnectDatabase(databaseURI)
	defer driver.Close(ctx)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth)

	// Both routers register bare paths — /people, /marriages, … — because the
	// Vite dev proxy strips the /api prefix the frontend calls before the request
	// arrives (API_SCOPE.md §2).
	personRouter.RegisterPersonRoutes(mux, driver)
	marriageRouter.RegisterMarriageRoutes(mux, driver)

	// Timeouts are set explicitly: http.ListenAndServe's defaults are unlimited,
	// which leaves a stalled client holding a connection for as long as it likes.
	// The write timeout is generous because a person's certificate is composed
	// from several graph round trips.
	server := &http.Server{
		Addr:              listenAddress,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	log.Printf("API listening on %s (database %s)", listenAddress, databaseURI)
	log.Fatal(server.ListenAndServe())
}
