package seeder

import "database/sql"

func CorrosionResistant(db *sql.DB) error {
	data := map[string]struct {
		External string
		Internal string
		CO2_CORR string
	}{
		"Alloy Gr 70N+Alloy 825 cladded": {"Res", "", "SS"},
		"Carbon Steel":                   {"NonRes", "", "CS"},
		"Carbon Steel + Alloy":           {"NonRes", "", "CS"},
		"Copper Alloyed Aluminum":        {"NonRes", "", "SS"},
		"Duplex SS":                      {"Res", "", "SS"},
		"Low Alloy Steel":                {"NonRes", "", "LA"},
		"Nickel Alloy SS":                {"Res", "", "SS"},
		"SA-283-A":                       {"NonRes", "", "CS"},
		"SA-283-B":                       {"NonRes", "", "CS"},
		"SA-283-C":                       {"NonRes", "", "CS"},
		"SA-36":                          {"NonRes", "", "CS"},
		"SA-515-60":                      {"NonRes", "", "CS"},
		"SA-515-65":                      {"NonRes", "", "CS"},
		"SA-515-70":                      {"NonRes", "", "CS"},
		"SA-516-55":                      {"NonRes", "", "CS"},
		"SA-516-60":                      {"NonRes", "", "CS"},
		"SA-516-70":                      {"NonRes", "", "CS"},
		"SA-516-65":                      {"NonRes", "", "CS"},
		"SS 304":                         {"NonRes", "", "SS"},
		"Stainless 300 series":           {"Res", "", "SS"},
		"Stainless 400 series":           {"Res", "", "SS"},
		"A240 316":                       {"Res", "", "SS"},
		"SA - 106 Gr B":                  {"NonRes", "", "CS"},
		"SA-234 WPB":                     {"NonRes", "", "CS"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO corrosion_resistant
		(id_shell_material, external, internal, co2_corr)
		VALUES (
			(SELECT id FROM shell_material WHERE name = ?),
			?,
			?,
			?
		)
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for name, value := range data {
		_, err := stmt.Exec(name, value.External, value.Internal, value.CO2_CORR)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
