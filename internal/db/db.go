package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func Init() *sql.DB {
	connStr := "host=postgres port=5432 user=postgres password=password dbname=attendance_db sslmode=disable"

	database, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = database.Ping()
	if err != nil {
		log.Fatal("DB not connected:", err)
	}

	log.Println("DB connected")
	return database
}
