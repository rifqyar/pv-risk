package seeder

import "database/sql"

func InhibitorInjectionReliability(db *sql.DB) error {

	data := []struct {
		Description      string
		MPYRange         string
		ReliabilityRange string
	}{
		{"1 < MPY ; < 95%", "1 < MPY", "< 95%"},
		{"1.0 <= MPY < 4.9 ; 89-95%", "1.0 <= MPY < 4.9", "89-95%"},
		{"5.0 < MPY < 10 ; 50-88%", "5.0 < MPY < 10", "50-88%"},
		{"10 >= MPY ; <50%", "10 >= MPY", "<50%"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
	INSERT OR IGNORE INTO inhibitor_injection_reliability 
	(description, mpy_range, reliability_range)
	VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, d := range data {

		_, err := stmt.Exec(
			d.Description,
			d.MPYRange,
			d.ReliabilityRange,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
