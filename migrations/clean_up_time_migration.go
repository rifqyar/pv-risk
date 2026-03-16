package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func CleanupTime(db *sql.DB) {

	query := `
	CREATE TABLE IF NOT EXISTS cleanup_time (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		description TEXT NOT NULL UNIQUE,
		score INTEGER,
		category TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := config.DB.Exec(query); err != nil {
		log.Fatal(err)
	}
}