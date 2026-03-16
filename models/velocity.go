package models

import (
	"database/sql"
	"time"
)

type Velocity struct {
	ID            int       `json:"id"`
	Code          string    `json:"code"`
	VelocityRange string    `json:"velocity_range"`
	VelocityValue int       `json:"velocity_value"`
	MechDamage    int       `json:"mech_damage"`
	LicVelocity   int       `json:"lic_velocity"`
	MicVelocity   int       `json:"mic_velocity"`
	Co2Corr       string    `json:"co2_corr"`
	CreatedAt     time.Time `json:"created_at"`
}

func GetVelocities(db *sql.DB) ([]Velocity, error) {

	rows, err := db.Query(`
		SELECT 
			id,
			code,
			velocity_range,
			velocity_value,
			mech_damage,
			lic_velocity,
			mic_velocity,
			co2_corr,
			created_at
		FROM velocity
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var velocities []Velocity

	for rows.Next() {
		var v Velocity

		err := rows.Scan(
			&v.ID,
			&v.Code,
			&v.VelocityRange,
			&v.VelocityValue,
			&v.MechDamage,
			&v.LicVelocity,
			&v.MicVelocity,
			&v.Co2Corr,
			&v.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		velocities = append(velocities, v)
	}

	return velocities, nil
}
