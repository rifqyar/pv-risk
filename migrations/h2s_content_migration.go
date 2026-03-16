package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func H2SContent(db *sql.DB) {

	query := `
	CREATE TABLE IF NOT EXISTS h2s_content (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		range TEXT NOT NULL,
		h2s_index INTEGER NOT NULL UNIQUE,
		code TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := config.DB.Exec(query); err != nil {
		log.Fatal(err)
	}
}
