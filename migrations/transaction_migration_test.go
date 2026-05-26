package migrations

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestEquipmentsTableRemovesTagNumberUniqueConstraint(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "migration.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	setup := []string{
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE equipments (id INTEGER PRIMARY KEY)`,
		`CREATE TABLE trx_equipments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			equipment_id INTEGER UNIQUE NOT NULL,
			tag_number TEXT UNIQUE NOT NULL,
			FOREIGN KEY (equipment_id) REFERENCES equipments(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE assessments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			equipment_id INTEGER,
			FOREIGN KEY (equipment_id) REFERENCES trx_equipments(id) ON DELETE CASCADE
		)`,
		`INSERT INTO equipments (id) VALUES (1), (2)`,
		`INSERT INTO trx_equipments (equipment_id, tag_number) VALUES (1, 'PV-001')`,
		`INSERT INTO assessments (equipment_id) VALUES (1)`,
	}
	for _, statement := range setup {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	EquipmentsTable(db)

	if _, err := db.Exec(`INSERT INTO trx_equipments (equipment_id, tag_number) VALUES (2, 'PV-001')`); err != nil {
		t.Fatalf("tag_number should be reusable after migration: %v", err)
	}

	var assessmentCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM assessments`).Scan(&assessmentCount); err != nil {
		t.Fatal(err)
	}
	if assessmentCount != 1 {
		t.Fatalf("migration changed related assessments: got %d rows", assessmentCount)
	}

	var headEnclosureColumnCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('trx_equipments') WHERE name = 'head_enclosure'`).Scan(&headEnclosureColumnCount); err != nil {
		t.Fatal(err)
	}
	if headEnclosureColumnCount != 1 {
		t.Fatal("head_enclosure column was not added")
	}
}
