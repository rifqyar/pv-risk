package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func PHCategory(db *sql.DB) {

	query := `
	CREATE TABLE IF NOT EXISTS ph_category (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ph_range TEXT NOT NULL,
		ph_index INTEGER NOT NULL UNIQUE,
		lic_value INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := config.DB.Exec(query); err != nil {
		log.Fatal(err)
	}
}
