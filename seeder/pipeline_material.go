package seeder

import "database/sql"

func PipelineMaterial(db *sql.DB) error {
	pipelineMaterial := []struct {
		Name string
		SMYS int
	}{
		{Name: "API 5L Gr A", SMYS: 30000},
		{Name: "API 5L Gr B", SMYS: 35000},
		{Name: "API 5L X 42", SMYS: 42000},
		{Name: "API 5L X 46", SMYS: 46000},
		{Name: "API 5L X 52", SMYS: 52000},
		{Name: "API 5L X 56", SMYS: 56000},
		{Name: "API 5L X 65", SMYS: 65000},
		{Name: "API 5L X 70", SMYS: 70000},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO pipeline_materials (name, smys)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, eq := range pipelineMaterial {
		_, err := stmt.Exec(eq.Name, eq.SMYS)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
