package models

import (
	"database/sql"
	"time"
)

type Pollution struct {
	ID        int       `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func GetPollutions(db *sql.DB) ([]Pollution, error) {

	rows, err := db.Query(`
		SELECT id, code, name, created_at
		FROM pollution
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pollutions []Pollution

	for rows.Next() {
		var p Pollution

		err := rows.Scan(
			&p.ID,
			&p.Code,
			&p.Name,
			&p.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		pollutions = append(pollutions, p)
	}

	return pollutions, nil
}
