package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func CISCCMatrix(db *sql.DB) {

	query := `
	CREATE TABLE IF NOT EXISTS ciscc_matrix (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ph_index INTEGER,
		h2s_index INTEGER,
		susceptibility TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(ph_index, h2s_index)
	);`

	if _, err := config.DB.Exec(query); err != nil {
		log.Fatal(err)
	}
}
