package models

import "database/sql"

type H2SContent struct {
	ID    int
	Range string
	Index int
	Code  string
}

func GetH2SContents(db *sql.DB) ([]H2SContent, error) {

	rows, err := db.Query(`
	SELECT id, range, h2s_index, code
	FROM h2s_content
	ORDER BY h2s_index
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var data []H2SContent

	for rows.Next() {

		var r H2SContent

		err := rows.Scan(
			&r.ID,
			&r.Range,
			&r.Index,
			&r.Code,
		)

		if err != nil {
			return nil, err
		}

		data = append(data, r)
	}

	return data, nil
}
