package models

import "database/sql"

type ReleaseProduct struct {
	ID        int    `json:"id"`
	Condition string `json:"condition"`
	Level     int    `json:"level"`
	Code      string `json:"code"`
}

func GetReleaseProducts(db *sql.DB) ([]ReleaseProduct, error) {

	rows, err := db.Query(`
	SELECT 
		id,
		condition,
		level,
		code
	FROM release_product
	ORDER BY level
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var data []ReleaseProduct

	for rows.Next() {

		var r ReleaseProduct

		err := rows.Scan(
			&r.ID,
			&r.Condition,
			&r.Level,
			&r.Code,
		)

		if err != nil {
			return nil, err
		}

		data = append(data, r)
	}

	return data, nil
}
