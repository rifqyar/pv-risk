package seeder

import "database/sql"

func CO2CorrosionPreventive(db *sql.DB) error {

	data := []struct {
		Method string
		Value  int
		Level  string
	}{
		{"Corrosion Inhibitor", -20, "LOW"},
		{"PH Stabilization", -20, "LOW"},
		{"No Corrosion Control", 20, "HIGH"},
		{"Not required", -20, "LOW"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
	INSERT OR IGNORE INTO co2_corrosion_preventive (method, value, level)
	VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, d := range data {

		_, err := stmt.Exec(
			d.Method,
			d.Value,
			d.Level,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
