package seeder

import "database/sql"

func VelocitySeeder(db *sql.DB) error {

	velocity := []struct {
		Code        string
		Range       string
		Value       int
		MechDamage  int
		LicVelocity int
		MicVelocity int
		Co2Corr     string
	}{
		{"Vel1", "v < 5 ft/s", 35, 5, 1, 25, "Low"},
		{"Vel2", "5 ≤ v < 10 ft/s", 30, 5, 1, 25, "Low"},
		{"Vel3", "10 ≤ v < 20 ft/s", 25, 10, 2, 30, "Med"},
		{"Vel4", "20 ≤ v < 30 ft/s", 20, 15, 3, 35, "Med"},
		{"Vel5", "v ≥ 30 ft/s", 15, 15, 4, 35, "High"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO velocity
		(code, velocity_range, velocity_value, mech_damage, lic_velocity, mic_velocity, co2_corr)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, v := range velocity {
		_, err := stmt.Exec(
			v.Code,
			v.Range,
			v.Value,
			v.MechDamage,
			v.LicVelocity,
			v.MicVelocity,
			v.Co2Corr,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
