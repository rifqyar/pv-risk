package models

import "database/sql"

type NozzleMaterial struct {
	ID         int
	Name       string
	CladStatus string
}

func GetNozzleMaterial(db *sql.DB) ([]NozzleMaterial, error) {
	rows, err := db.Query("SELECT id, name, clad_status FROM nozzle_material")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var nozzleMaterial []NozzleMaterial

	for rows.Next() {
		var sm NozzleMaterial
		err := rows.Scan(&sm.ID, &sm.Name, &sm.CladStatus)
		if err != nil {
			return nil, err
		}

		nozzleMaterial = append(nozzleMaterial, sm)
	}

	return nozzleMaterial, nil
}
