package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func Phase(db *sql.DB) {
	var err error

	createPhase := `
	CREATE TABLE IF NOT EXISTS phase (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		abbr TEXT NOT NULL,
		factor INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err = config.DB.Exec(createPhase); err != nil {
		log.Fatal(err)
	}
}
