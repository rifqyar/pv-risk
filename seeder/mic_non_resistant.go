package seeder

import "database/sql"

func MicNonResistantSeeder(db *sql.DB) error {
	data := map[string]struct {
		MIC             string
		AmineCracking   string
		SulfideCracking string
	}{
		"Alloy Gr 70N+Alloy 825 cladded": {"MICNR", "Res", "Res"},
		"Carbon Steel":                   {"MICNR", "NonRes", "NonRes"},
		"Carbon Steel + Alloy":           {"MICNR", "NonRes", "NonRes"},
		"Copper Alloyed Aluminum":        {"MICNR", "Res", "NonRes"},
		"Duplex SS":                      {"MICNR", "Res", "NonRes"},
		"Low Alloy Steel":                {"MICNR", "NonRes", "NonRes"},
		"Nickel Alloy SS":                {"MICNR", "Res", "NonRes"},
		"SA-283-A":                       {"MICNR", "NonRes", "NonRes"},
		"SA-283-B":                       {"MICNR", "NonRes", "NonRes"},
		"SA-283-C":                       {"MICNR", "NonRes", "NonRes"},
		"SA-36":                          {"MICNR", "NonRes", "NonRes"},
		"SA-515-60":                      {"MICNR", "NonRes", "NonRes"},
		"SA-515-65":                      {"MICNR", "NonRes", "NonRes"},
		"SA-515-70":                      {"MICNR", "NonRes", "NonRes"},
		"SA-516-55":                      {"MICNR", "NonRes", "NonRes"},
		"SA-516-60":                      {"MICNR", "NonRes", "NonRes"},
		"SA-516-70":                      {"MICNR", "NonRes", "NonRes"},
		"SA-516-65":                      {"MICNR", "NonRes", "NonRes"},
		"SS 304":                         {"MICNR", "Res", "NonRes"},
		"Stainless 300 series":           {"MICNR", "Res", "NonRes"},
		"Stainless 400 series":           {"MICNR", "Res", "NonRes"},
		"A240 316":                       {"MICNR", "Res", "NonRes"},
		"SA - 106 Gr B":                  {"MICNR", "NonRes", "NonRes"},
		"SA-234 WPB":                     {"MICNR", "NonRes", "NonRes"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO mic_resistant
		(id_shell_material, mic, amine_cracking, sulfide_cracking)
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
		_, err := stmt.Exec(name, value.MIC, value.AmineCracking, value.SulfideCracking)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
