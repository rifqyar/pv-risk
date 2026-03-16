package models

import "database/sql"

type ShellMaterial struct {
	ID   int
	Name string
	Type string
}

func GetShellMaterial(db *sql.DB) ([]ShellMaterial, error) {
	rows, err := db.Query("SELECT id, name FROM shell_material")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var shellMaterial []ShellMaterial

	for rows.Next() {
		var sm ShellMaterial
		err := rows.Scan(&sm.ID, &sm.Name)
		if err != nil {
			return nil, err
		}

		shellMaterial = append(shellMaterial, sm)
	}

	return shellMaterial, nil
}
