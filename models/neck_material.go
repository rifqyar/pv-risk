package models

import "database/sql"

type NeckMaterial struct {
	ID         int
	Name       string
	CladStatus string
}

func GetNeckMaterial(db *sql.DB) ([]NeckMaterial, error) {
	rows, err := db.Query("SELECT id, name, clad_status FROM neck_material")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var neckMaterial []NeckMaterial

	for rows.Next() {
		var sm NeckMaterial
		err := rows.Scan(&sm.ID, &sm.Name, &sm.CladStatus)
		if err != nil {
			return nil, err
		}

		neckMaterial = append(neckMaterial, sm)
	}

	return neckMaterial, nil
}
