package models

import "database/sql"

type InhibitorInjectionReliability struct {
	ID               int    `json:"id"`
	Description      string `json:"description"`
	MPYRange         string `json:"mpy_range"`
	ReliabilityRange string `json:"reliability_range"`
}

func GetInhibitorInjectionReliability(db *sql.DB) ([]InhibitorInjectionReliability, error) {

	rows, err := db.Query(`
	SELECT 
		id,
		description,
		mpy_range,
		reliability_range
	FROM inhibitor_injection_reliability
	ORDER BY id
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var data []InhibitorInjectionReliability

	for rows.Next() {

		var r InhibitorInjectionReliability

		err := rows.Scan(
			&r.ID,
			&r.Description,
			&r.MPYRange,
			&r.ReliabilityRange,
		)

		if err != nil {
			return nil, err
		}

		data = append(data, r)
	}

	return data, nil
}
