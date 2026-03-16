package seeder

import "database/sql"

func PhaseSeeder(db *sql.DB) error {

	phase := []struct {
		Code   string
		Name   string
		Abbr   string
		Factor int
	}{
		{"Ph1", "Dry Gas", "drygas", 15},
		{"Ph2", "Wet Gas", "wetgas", 10},
		{"Ph3", "Liquid", "liquid", 15},
		{"Ph4", "Oily Water", "oilywtr", 15},
		{"Ph5", "Vapour", "vapour", 15},
		{"Ph6", "Processed Water", "processwtr", 15},
		{"Ph7", "Produced Water", "producedwtr", 15},
		{"Ph8", "Multiphase", "multiphase", 15},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO phase
		(code, name, abbr, factor)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range phase {
		_, err := stmt.Exec(
			p.Code,
			p.Name,
			p.Abbr,
			p.Factor,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
