package models

import "database/sql"

type ChlorideContent struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Level       int    `json:"level"`
}

func GetChlorideContents(db *sql.DB) ([]ChlorideContent, error) {

	rows, err := db.Query(`
	SELECT 
		id,
		description,
		level
	FROM chloride_content
	ORDER BY id
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var data []ChlorideContent

	for rows.Next() {

		var r ChlorideContent

		err := rows.Scan(
			&r.ID,
			&r.Description,
			&r.Level,
		)

		if err != nil {
			return nil, err
		}

		data = append(data, r)
	}

	return data, nil
}
