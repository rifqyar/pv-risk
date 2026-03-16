package models

import "database/sql"

type CO2CorrosionPreventive struct {
	ID     int    `json:"id"`
	Method string `json:"method"`
	Value  int    `json:"value"`
	Level  string `json:"level"`
}

func GetCO2CorrosionPreventives(db *sql.DB) ([]CO2CorrosionPreventive, error) {

	rows, err := db.Query(`
	SELECT 
		id,
		method,
		value,
		level
	FROM co2_corrosion_preventive
	ORDER BY id
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var data []CO2CorrosionPreventive

	for rows.Next() {

		var r CO2CorrosionPreventive

		err := rows.Scan(
			&r.ID,
			&r.Method,
			&r.Value,
			&r.Level,
		)

		if err != nil {
			return nil, err
		}

		data = append(data, r)
	}

	return data, nil
}
