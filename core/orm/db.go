package orm

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

var DB *sql.DB

func InitDB(connStr string) {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	fmt.Println("Successfully connected to the database")
}
