package seeder

import "database/sql"

func H2SContent(db *sql.DB) error {

	data := []struct {
		Range string
		Index int
		Code  string
	}{
		{"<= 50 PPM", 1, "A"},
		{"50 < PPM <= 1000", 2, "B"},
		{"1000 < PPM <= 10000", 3, "C"},
		{"> 10000 PPM", 4, "D"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
	INSERT OR IGNORE INTO h2s_content (range, h2s_index, code)
	VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, d := range data {
		_, err := stmt.Exec(d.Range, d.Index, d.Code)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
