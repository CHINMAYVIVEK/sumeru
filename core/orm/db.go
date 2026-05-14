package orm

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

var DB DBWrapper

func InitDB(connStr string) {
	rawDB, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	DB = NewDBWrapper(rawDB)

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	fmt.Println("Successfully connected to the database")
}
