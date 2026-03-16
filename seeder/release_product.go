package seeder

import "database/sql"

func ReleaseProduct(db *sql.DB) error {

	data := []struct {
		Condition string
		Level     int
		Code      string
	}{
		{"RP < 10 m³", 2, "B"},
		{"10 >= RP < 20 m³", 3, "C"},
		{"RP >= 20 m³", 4, "D"},
		{"no release product", 1, "A"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
	INSERT OR IGNORE INTO release_product (condition, level, code)
	VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, d := range data {

		_, err := stmt.Exec(
			d.Condition,
			d.Level,
			d.Code,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
