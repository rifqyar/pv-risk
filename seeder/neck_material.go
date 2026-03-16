package seeder

import "database/sql"

func NeckMaterial(db *sql.DB) error {
	shellMaterial := []struct {
		Name       string
		CladStatus string
	}{
		{Name: "A 105", CladStatus: "noclad"},
		{Name: "A 105 N", CladStatus: "noclad"},
		{Name: "A 105 N+Alloy 825 cladded", CladStatus: "cladded"},
		{Name: "A 106 Gr. B +Alloy 825 Cladded", CladStatus: "cladded"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO neck_material (name, clad_status)
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
