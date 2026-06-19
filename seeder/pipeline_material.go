package seeder

import "database/sql"

func PipelineMaterial(db *sql.DB) error {
	pipelineMaterial := []struct {
		Name            string
		Grade           string
		SMYS            int
		AllowableStress float64
	}{
		{Name: "API 5L Gr A", Grade: "Gr A", SMYS: 30000, AllowableStress: 20000},
		{Name: "API 5L Gr B", Grade: "Gr B", SMYS: 35000, AllowableStress: 20000},
		{Name: "API 5L X 42", Grade: "X42", SMYS: 42000, AllowableStress: 20000},
		{Name: "API 5L X 46", Grade: "X46", SMYS: 46000, AllowableStress: 20000},
		{Name: "API 5L X 52", Grade: "X52", SMYS: 52000, AllowableStress: 20000},
		{Name: "API 5L X 56", Grade: "X56", SMYS: 56000, AllowableStress: 20000},
		{Name: "API 5L X 65", Grade: "X65", SMYS: 65000, AllowableStress: 20000},
		{Name: "API 5L X 70", Grade: "X70", SMYS: 70000, AllowableStress: 20000},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO pipeline_materials (name, material_specification, grade, smys, allowable_stress, notes, is_active)
		VALUES (?, ?, ?, ?, ?, ?, 1)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, eq := range pipelineMaterial {
		_, err := stmt.Exec(eq.Name, eq.Name, eq.Grade, eq.SMYS, eq.AllowableStress, "Seeded default; verify allowable stress against approved project stress table.")
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
