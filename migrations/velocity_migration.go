package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func VelocityMigration(db *sql.DB) {
	var err error

	createVelocity := `
	CREATE TABLE IF NOT EXISTS velocity (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT NOT NULL UNIQUE,
		velocity_range TEXT NOT NULL,
		velocity_value INTEGER,
		mech_damage INTEGER,
		lic_velocity INTEGER,
		mic_velocity INTEGER,
		co2_corr TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err = config.DB.Exec(createVelocity); err != nil {
		log.Fatal(err)
	}
}
