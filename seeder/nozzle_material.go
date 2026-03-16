package seeder

import "database/sql"

func NozzleMaterial(db *sql.DB) error {
	shellMaterial := []struct {
		Name       string
		CladStatus string
	}{
		{Name: "SA -106-A", CladStatus: "noclad"},
		{Name: "SA -106-B", CladStatus: "noclad"},
		{Name: "SA -106-C", CladStatus: "noclad"},
		{Name: "SA-53", CladStatus: "noclad"},
		{Name: "A 105 N+Alloy 825 cladded", CladStatus: "cladded"},
		{Name: "A 106 Gr. B +Alloy 825 Cladded", CladStatus: "cladded"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO nozzle_material (name, clad_status)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, eq := range shellMaterial {
		_, err := stmt.Exec(eq.Name, eq.CladStatus)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
