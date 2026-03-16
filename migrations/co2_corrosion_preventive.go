package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func CO2CorrosionPreventive(db *sql.DB) {

	query := `
	CREATE TABLE IF NOT EXISTS co2_corrosion_preventive (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		method TEXT NOT NULL UNIQUE,
		value INTEGER,
		level TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := config.DB.Exec(query); err != nil {
		log.Fatal(err)
	}
}
