package orm

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"

	"sumeru/core/applog"
)

var DB DBWrapper

func InitDB(connStr string) {
	rawDB, err := sql.Open("postgres", connStr)
	if err != nil {
		applog.Fatal(context.Background(), "db_open", "err", err)
	}

	DB = NewDBWrapper(rawDB)

	err = DB.Ping()
	if err != nil {
		applog.Fatal(context.Background(), "db_ping", "err", err)
	}

	applog.InfoMsg(context.Background(), "orm", "connect", "Successfully connected to the database", nil)
}
