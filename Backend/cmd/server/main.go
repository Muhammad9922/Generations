package main

import (
	"log"
	"net/http"
	"os"

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

	// Both routers register bare paths — /people, /marriages, … — because the
	// Vite dev proxy strips the /api prefix the frontend calls before the request
	// arrives (API_SCOPE.md §2).
	personRouter.RegisterPersonRoutes(mux, driver)
	marriageRouter.RegisterMarriageRoutes(mux, driver)

	log.Printf("API listening on %s (database %s)", listenAddress, databaseURI)
	log.Fatal(http.ListenAndServe(listenAddress, mux))
}
