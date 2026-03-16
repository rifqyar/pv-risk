package seeder

import "database/sql"

func SeedTypeHead(db *sql.DB) error {
	typeHead := []struct {
		Name string
	}{
		{Name: "Bolted Cover"},
		{Name: "Conical/Toriconical"},
		{Name: "Ellipsoidal"},
		{Name: "Hemispherical"},
		{Name: "Torisherical"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO head_types (name)
		VALUES (?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, eq := range typeHead {
		_, err := stmt.Exec(eq.Name)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
