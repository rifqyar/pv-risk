package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func DamageMechanicalImpact(db *sql.DB) {
	var err error

	createDamageMehcanicalImpact := `
	CREATE TABLE IF NOT EXISTS damage_mechanical_impact (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT NOT NULL UNIQUE,
		impact TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err = config.DB.Exec(createDamageMehcanicalImpact); err != nil {
		log.Fatal(err)
	}
}
