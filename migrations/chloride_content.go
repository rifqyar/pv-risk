package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func ChlorideContent(db *sql.DB) {

	query := `
	CREATE TABLE IF NOT EXISTS chloride_content (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		description TEXT NOT NULL UNIQUE,
		level INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := config.DB.Exec(query); err != nil {
		log.Fatal(err)
	}
}
