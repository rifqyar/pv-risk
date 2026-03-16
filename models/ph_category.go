package models

import "database/sql"

type PHCategory struct {
	ID       int
	Range    string
	Index    int
	LICValue int
}

func GetPHCategories(db *sql.DB) ([]PHCategory, error) {

	rows, err := db.Query(`
	SELECT id, ph_range, ph_index, lic_value
	FROM ph_category
	ORDER BY ph_index
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var data []PHCategory

	for rows.Next() {

		var r PHCategory

		err := rows.Scan(
			&r.ID,
			&r.Range,
			&r.Index,
			&r.LICValue,
		)

		if err != nil {
			return nil, err
		}

		data = append(data, r)
	}

	return data, nil
}
