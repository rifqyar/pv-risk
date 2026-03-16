package models

import "database/sql"

type CleanupTime struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Score       int    `json:"score"`
	Category    string `json:"category"`
}

func GetCleanupTimes(db *sql.DB) ([]CleanupTime, error) {

	rows, err := db.Query(`
	SELECT 
		id,
		description,
		score,
		category
	FROM cleanup_time
	ORDER BY id
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var data []CleanupTime

	for rows.Next() {

		var r CleanupTime

		err := rows.Scan(
			&r.ID,
			&r.Description,
			&r.Score,
			&r.Category,
		)

		if err != nil {
			return nil, err
		}

		data = append(data, r)
	}

	return data, nil
}
