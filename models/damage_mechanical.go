package models

import "database/sql"

type DamageMechanicalImpact struct {
	ID     int
	Code   string
	Impact string
}

func GetDamageMechanical(db *sql.DB) ([]DamageMechanicalImpact, error) {
	rows, err := db.Query("SELECT id, code, impact from damage_mechanical_impact")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var damageMechanical []DamageMechanicalImpact

	for rows.Next() {
		var dm DamageMechanicalImpact
		err := rows.Scan(&dm.ID, &dm.Code, &dm.Impact)
		if err != nil {
			return nil, err
		}

		damageMechanical = append(damageMechanical, dm)
	}

	return damageMechanical, nil
}
