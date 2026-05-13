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

	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *sqlx.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS nodes (
		id TEXT PRIMARY KEY,
		name TEXT,
		cpu_cores REAL,
		ram_gb REAL,
		safe_boundary REAL DEFAULT 75,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		node_id TEXT,
		timestamp DATETIME,
		cpu_usage_percent REAL,
		ram_usage_bytes INTEGER,
		ram_usage_percent REAL,
		UNIQUE(node_id, timestamp)
	);

	CREATE INDEX IF NOT EXISTS idx_metrics_node_time ON metrics(node_id, timestamp);
	`
	_, err := db.Exec(schema)
	return err
}
