package migrations

import (
	"database/sql"
	"log"
)

func CorrosionResistantMigration(db *sql.DB) {
	var err error

	creatCorrosionResistant := `
	CREATE TABLE IF NOT EXISTS corrosion_resistant (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		id_shell_material INTEGER NOT NULL UNIQUE,
		external TEXT,
		internal TEXT,
		co2_corr TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (id_shell_material) 
			REFERENCES shell_material(id)
			ON DELETE CASCADE
	);`

	if _, err = db.Exec(creatCorrosionResistant); err != nil {
		log.Fatal(err)
	}
}
