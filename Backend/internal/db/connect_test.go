package db

import (
	"fmt"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestConnectNeo(t *testing.T) {
	dbUri := "bolt://192.168.0.133:7687"
	ctx, driver := ConnectDatabase(dbUri)
	defer driver.Close(ctx)

	if err := driver.VerifyConnectivity(ctx); err != nil {
		t.Errorf("Error While Connecting: %v", err)
	} else {
		fmt.Printf("Succesfully Connected")
	}
}

func TestConnectPostgres(t *testing.T) {
	dbUri := "postgresql://myuser:mypassword@192.168.0.133:7644/mydatabase"
	db, closeConnection, err := ConnectPostgres(dbUri)
	if err != nil {
		t.Fatalf("Error While Connecting To Postgres: %q", err)
	}
	defer closeConnection()

	// 1. Verify the connection is truly active and responsive
	if err := db.Ping(); err != nil {
		t.Fatalf("Error pinging Postgres: %q", err)
	}

	// 2. Query data using QueryRow since SELECT returns a row
	var isInRecovery bool
	err = db.QueryRow("SELECT pg_is_in_recovery();").Scan(&isInRecovery)
	if err != nil {
		t.Errorf("Error While Querying: %q", err)
	}

	t.Logf("Postgres Health - Is in recovery: %v", isInRecovery)
}
