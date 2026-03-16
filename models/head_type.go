package models

import "database/sql"

type HeadType struct {
	ID   int
	Name string
}

func GetHeadTypes(db *sql.DB) ([]HeadType, error) {
	rows, err := db.Query("SELECT id, name FROM head_types")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var headTypes []HeadType

	for rows.Next() {
		var ht HeadType
		err := rows.Scan(&ht.ID, &ht.Name)
		if err != nil {
			return nil, err
		}
		headTypes = append(headTypes, ht)
	}

	return headTypes, nil
}
