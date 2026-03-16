package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func InhibitorInjectionReliability(db *sql.DB) {

	query := `
	CREATE TABLE IF NOT EXISTS inhibitor_injection_reliability (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		description TEXT NOT NULL UNIQUE,
		mpy_range TEXT,
		reliability_range TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := config.DB.Exec(query); err != nil {
		log.Fatal(err)
	}
}
