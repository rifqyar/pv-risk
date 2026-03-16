package seeder

import "database/sql"

func PHCategory(db *sql.DB) error {

	data := []struct {
		Range string
		Index int
		LIC   int
	}{
		{"<= 5.5", 1, 5},
		{"5.5 < PH <= 7.5", 2, 10},
		{"7.5 < PH <= 8.3", 3, 10},
		{"8.3 < PH <= 8.9", 4, 15},
		{"> 9.0", 5, 15},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
	INSERT OR IGNORE INTO ph_category (ph_range, ph_index, lic_value)
	VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, d := range data {
		_, err := stmt.Exec(d.Range, d.Index, d.LIC)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
