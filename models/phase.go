package models

import (
	"database/sql"
	"time"
)

type Phase struct {
	ID        int       `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Abbr      string    `json:"abbr"`
	Factor    int       `json:"factor"`
	CreatedAt time.Time `json:"created_at"`
}

func GetPhases(db *sql.DB) ([]Phase, error) {

	rows, err := db.Query(`
		SELECT id, code, name, abbr, factor, created_at
		FROM phase
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var phases []Phase

	for rows.Next() {
		var p Phase

		err := rows.Scan(
			&p.ID,
			&p.Code,
			&p.Name,
			&p.Abbr,
			&p.Factor,
			&p.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		phases = append(phases, p)
	}

	return phases, nil
}
