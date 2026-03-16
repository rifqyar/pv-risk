package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func ReleaseProduct(db *sql.DB) {

	query := `
	CREATE TABLE IF NOT EXISTS release_product (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		condition TEXT NOT NULL UNIQUE,
		level INTEGER,
		code TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := config.DB.Exec(query); err != nil {
		log.Fatal(err)
	}
}
