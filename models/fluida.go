package models

import "database/sql"

type FluidaData struct {
	ID                         int
	Code                       string
	Name                       string
	Abbr                       string
	LocalisedInternalCorrosion *int
}

func GetFluida(db *sql.DB) ([]FluidaData, error) {
	rows, err := db.Query("SELECT id, code, name, abbr, localised_internal_corrosion from fluida")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var fluida []FluidaData

	for rows.Next() {
		var f FluidaData
		err := rows.Scan(&f.ID, &f.Code, &f.Name, &f.Abbr, &f.LocalisedInternalCorrosion)
		if err != nil {
			return nil, err
		}

		fluida = append(fluida, f)
	}

	return fluida, nil
}
