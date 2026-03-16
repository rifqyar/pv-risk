package seeder

import "database/sql"

func DamageMechanicalImpact(db *sql.DB) error {
	damageImpact := []struct {
		Code   string
		Impact string
	}{
		{Code: "NoSD", Impact: "No Shutdown"},
		{Code: "LocSD", Impact: "Local Shutdown"},
		{Code: "TotSD", Impact: "Total Shutdown"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO damage_mechanical_impact (code, impact)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, eq := range damageImpact {
		_, err := stmt.Exec(eq.Code, eq.Impact)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
