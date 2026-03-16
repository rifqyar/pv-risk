package seeder

import "database/sql"

func ShellMaterial(db *sql.DB) error {
	shellMaterial := []struct {
		Name string
	}{
		{Name: "Alloy Gr 70N+Alloy 825 cladded"},
		{Name: "Carbon Steel"},
		{Name: "Carbon Steel + Alloy"},
		{Name: "Copper Alloyed Aluminum"},
		{Name: "Duplex SS"},
		{Name: "Low Alloy Steel"},
		{Name: "Nickel Alloy SS"},
		{Name: "SA-283-A"},
		{Name: "SA-283-B"},
		{Name: "SA-283-C"},
		{Name: "SA-36"},
		{Name: "SA-515-60"},
		{Name: "SA-515-65"},
		{Name: "SA-515-70"},
		{Name: "SA-516-55"},
		{Name: "SA-516-60"},
		{Name: "SA-516-65"},
		{Name: "SA-516-70"},
		{Name: "SS 304"},
		{Name: "Stainless 300 series"},
		{Name: "Stainless 400 series"},
		{Name: "A240 316"},
		{Name: "SA - 106 Gr B"},
		{Name: "SA-234 WPB"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO shell_material (name)
		VALUES (?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, eq := range shellMaterial {
		_, err := stmt.Exec(eq.Name)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
