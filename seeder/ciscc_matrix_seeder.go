package seeder

import "database/sql"

func CISCCMatrix(db *sql.DB) error {

	data := []struct {
		PHIndex int
		H2S     int
		Result  string
	}{
		{1, 1, "Low"}, {1, 2, "Moderate"}, {1, 3, "High"}, {1, 4, "High"},
		{2, 1, "Low"}, {2, 2, "Low"}, {2, 3, "Low"}, {2, 4, "Moderate"},
		{3, 1, "Low"}, {3, 2, "Moderate"}, {3, 3, "Moderate"}, {3, 4, "Moderate"},
		{4, 1, "Low"}, {4, 2, "Moderate"}, {4, 3, "Moderate"}, {4, 4, "High"},
		{5, 1, "Low"}, {5, 2, "Moderate"}, {5, 3, "High"}, {5, 4, "High"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
	INSERT OR IGNORE INTO ciscc_matrix (ph_index, h2s_index, susceptibility)
	VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, d := range data {
		_, err := stmt.Exec(d.PHIndex, d.H2S, d.Result)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
