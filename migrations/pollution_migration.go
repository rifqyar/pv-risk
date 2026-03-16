package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func PollutioMigration(db *sql.DB) {
	var err error

	createPollution := `
	CREATE TABLE IF NOT EXISTS pollution (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err = config.DB.Exec(createPollution); err != nil {
		log.Fatal(err)
	}
}
