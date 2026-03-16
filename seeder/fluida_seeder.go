package seeder

import "database/sql"

func Fluida(db *sql.DB) error {

	fluida := []struct {
		Code string
		Name string
		Abbr string
		LIC  *int
	}{
		{"Fl1", "air", "air", nil},
		{"Fl2", "air, utility", "airutil", nil},
		{"Fl3", "amine", "amine", nil},
		{"Fl4", "amine, lean", "aminelean", nil},
		{"Fl5", "amine, lean + feed gas", "amleanfgas", nil},
		{"Fl6", "amine, liquid", "amineliq", nil},
		{"Fl7", "amine, rich", "aminerich", nil},
		{"Fl8", "gas, acid", "gasacid", nil},
		{"Fl9", "gas, dry", "gasdry", nil},
		{"Fl10", "gas, fuel", "gasfuel", nil},
		{"Fl11", "gas, sales", "gasales", nil},
		{"Fl12", "gas, sweet", "gasweet", nil},
		{"Fl13", "glycol", "glycol", nil},
		{"Fl14", "glycol, lean", "glycolean", nil},
		{"Fl15", "glycol, lean + wet gas", "glycleanwetgas", intPtr(35)},
		{"Fl16", "glycol, rich", "glycolrich", nil},
		{"Fl17", "hydrocarbon, gas", "Hcgas", nil},
		{"Fl18", "hydrocarbon, liquid", "Hcliq", intPtr(30)},
		{"Fl19", "hydrocarbon, liquid + water", "Hcliqwtr", intPtr(25)},
		{"Fl20", "oil, hot", "oilhot", nil},
		{"Fl21", "vapour", "vapour", nil},
		{"Fl22", "water, free (KO drum)", "waterfreeKO", nil},
		{"Fl23", "water, process", "waterprocess", nil},
		{"Fl24", "water, produced", "waterproduced", nil},
		{"Fl25", "water, oily", "wateroily", nil},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO fluida 
		(code, name, abbr, localised_internal_corrosion)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, f := range fluida {
		_, err := stmt.Exec(
			f.Code,
			f.Name,
			f.Abbr,
			f.LIC,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func intPtr(i int) *int {
	return &i
}
