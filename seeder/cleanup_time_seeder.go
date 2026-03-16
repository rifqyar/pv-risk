package seeder

import "database/sql"

func CleanupTime(db *sql.DB) error {

	data := []struct {
		Description string
		Score       int
		Category    string
	}{
		{"CT < 1 day", 1, "A"},
		{"1 <= CT <=7 days", 1, "A"},
		{"8 >= CT <= 14 days", 2, "B"},
		{"15<= CT <= 30 days", 3, "C"},
		{"> 1 month", 4, "D"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
	INSERT OR IGNORE INTO cleanup_time (description, score, category)
	VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, d := range data {

		_, err := stmt.Exec(
			d.Description,
			d.Score,
			d.Category,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
