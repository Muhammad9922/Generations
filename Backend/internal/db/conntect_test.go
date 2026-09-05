package db

import (
	"fmt"
	"testing"
)

func TestConnect(t *testing.T) {
	dbUri := "bolt://localhost:7687"
	ctx, driver := ConnectDatabase(dbUri)
	defer driver.Close(ctx)

	if err := driver.VerifyConnectivity(ctx); err != nil {
		t.Errorf("Error While Connecting: %v", err)
	} else {
		fmt.Printf("Succesfully Connected")
	}
}
