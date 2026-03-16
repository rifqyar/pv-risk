package models

import "database/sql"

type CISCCMatrix struct {
	ID             int    `json:"id"`
	PHIndex        int    `json:"ph_index"`
	H2SIndex       int    `json:"h2s_index"`
	Susceptibility string `json:"susceptibility"`
}

func GetCISCCMatrix(db *sql.DB) ([]CISCCMatrix, error) {

	rows, err := db.Query(`
		SELECT 
			id,
			ph_index,
			h2s_index,
			susceptibility
		FROM ciscc_matrix
		ORDER BY ph_index, h2s_index
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []CISCCMatrix

	for rows.Next() {

		var r CISCCMatrix

		err := rows.Scan(
			&r.ID,
			&r.PHIndex,
			&r.H2SIndex,
			&r.Susceptibility,
		)

		if err != nil {
			return nil, err
		}

		data = append(data, r)
	}
	return data, nil
}
