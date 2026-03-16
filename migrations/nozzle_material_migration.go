package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func NozzleMaterial(db *sql.DB) {
	var err error

	createNozzleMaterial := `
	CREATE TABLE IF NOT EXISTS nozzle_material (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		clad_status TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err = config.DB.Exec(createNozzleMaterial); err != nil {
		log.Fatal(err)
	}
}
