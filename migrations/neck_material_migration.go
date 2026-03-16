package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func NeckMaterial(db *sql.DB) {
	var err error

	createNeckMaterial := `
	CREATE TABLE IF NOT EXISTS neck_material (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		clad_status TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err = config.DB.Exec(createNeckMaterial); err != nil {
		log.Fatal(err)
	}
}
