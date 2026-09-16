package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

func ConnectPostgres(dbURI string) (*sql.DB, func(), error) {
	db, err := sql.Open("pgx", dbURI)

	if err != nil {
		return nil, nil, fmt.Errorf("Unable to open connection: %q", err)
	}

	db.SetMaxIdleConns(5)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(15 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("Postgres Ping Failed: %v", err)
	}

	cleanup := func() {
		if err := db.Close(); err != nil {
			log.Printf("Error Closing Database: %v", err)
		} else {
			log.Println("database connection pool closed")
		}
	}

	return db, cleanup, nil

}
