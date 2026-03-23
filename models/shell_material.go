package models

import "database/sql"

type ShellMaterial struct {
	ID       int
	Name     string
	CO2Corr  string
	Internal string
	External string
}

func GetShellMaterial(db *sql.DB) ([]ShellMaterial, error) {
	query := `
		SELECT 
			sm.id, 
			sm.name, 
			COALESCE(cr.co2_corr, '') as co2_corr, 
			COALESCE(cr.internal, '') as internal, 
			COALESCE(cr."external", '') as external
		FROM shell_material sm 
		INNER JOIN corrosion_resistant cr ON sm.id = cr.id_shell_material
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shellMaterial []ShellMaterial

	for rows.Next() {
		var sm ShellMaterial
		err := rows.Scan(&sm.ID, &sm.Name, &sm.CO2Corr, &sm.Internal, &sm.External)
		if err != nil {
			return nil, err
		}

		shellMaterial = append(shellMaterial, sm)
	}

	return shellMaterial, nil
}
