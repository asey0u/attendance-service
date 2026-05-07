package db

import (
	"database/sql"
	"log"
	"os"
)

func RunMigrations(db *sql.DB) {

	sqlBytes, err := os.ReadFile("migrations/init.sql")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("migrations applied")
}
