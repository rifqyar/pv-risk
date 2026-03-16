package seeder

import "database/sql"

func Pollution(db *sql.DB) error {

	pollutions := []struct {
		Code string
		Name string
	}{
		{"HCGasNonPol", "Hydrocarbon Gas + Non Polluting Product"},
		{"OilyWtr", "Oily Water"},
		{"LiqHC", "Liquid Hydrocarbon"},
		{"Chemical", "Chemical"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO pollution (code, name)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range pollutions {
		_, err := stmt.Exec(
			p.Code,
			p.Name,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
