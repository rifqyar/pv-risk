package migrations

import (
	"database/sql"
	"log"
)

func MicNonResistantMigration(db *sql.DB) {
	var err error

	createMicNonResistant := `
	CREATE TABLE IF NOT EXISTS mic_resistant (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		id_shell_material INTEGER NOT NULL UNIQUE,
		mic TEXT,
		amine_cracking TEXT,
		sulfide_cracking TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (id_shell_material) 
			REFERENCES shell_material(id)
			ON DELETE CASCADE
	);`

	if _, err = db.Exec(createMicNonResistant); err != nil {
		log.Fatal(err)
	}
}
