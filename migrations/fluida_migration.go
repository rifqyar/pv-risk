package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func Fluida(db *sql.DB) {
	var err error

	createFluida := `
	CREATE TABLE IF NOT EXISTS fluida (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		abbr TEXT NOT NULL,
		localised_internal_corrosion INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_fluida_code ON fluida(code);`

	if _, err = config.DB.Exec(createFluida); err != nil {
		log.Fatal(err)
	}
}
