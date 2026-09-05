package db

import (
	"context"
	"fmt"
	"log"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

func ConnectDatabase(dbUri string) (context.Context, neo4j.Driver) {
	ctx := context.Background()

	auth := neo4j.BasicAuth("", "", "")

	driver, err := neo4j.NewDriver(dbUri, auth)

	if err != nil {
		log.Fatalf("Failed To Connect: %v", err)
	}

	if err := driver.VerifyConnectivity(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Printf("Connected To Server!")

	return ctx, driver
}
