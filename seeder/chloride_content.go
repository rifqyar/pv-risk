package seeder

import "database/sql"

func ChlorideContent(db *sql.DB) error {

	data := []struct {
		Description string
		Level       int
	}{
		{"1 <= PPM", 1},
		{"1 < PPM <= 10", 2},
		{"10 < PPM <= 100", 3},
		{"100 < PPM <= 1000", 4},
		{"> 1000 PPM", 5},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
	INSERT OR IGNORE INTO chloride_content (description, level)
	VALUES (?, ?)
	`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, d := range data {

		_, err := stmt.Exec(
			d.Description,
			d.Level,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
