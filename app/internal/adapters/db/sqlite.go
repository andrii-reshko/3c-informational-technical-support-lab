package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteConnector struct {
	db *sqlx.DB
}

func NewDbConnection(path string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000&parseTime=true&loc=UTC")
	if err != nil {
		return nil, err
	}
	return db, nil
}
