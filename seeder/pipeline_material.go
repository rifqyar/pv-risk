package seeder

import "database/sql"

const PipelineMaterialStressDatasetVersion = "pipeline-b31-material-stress-v1"
const PipelineMaterialStressDatasetSource = "pipeline_specific_dataset"

func PipelineMaterial(db *sql.DB) error {
	pipelineMaterial := []struct {
		Name            string
		Grade           string
		SMYS            int
		AllowableStress float64
		Code            string
		Edition         string
		ProductForm     string
	}{
		{Name: "API 5L Gr A", Grade: "Gr A", SMYS: 30000, AllowableStress: 20000, Code: "ASME B31.3/B31.4/B31.8", Edition: "Project approved 2025", ProductForm: "Pipe"},
		{Name: "API 5L Gr B", Grade: "Gr B", SMYS: 35000, AllowableStress: 20000, Code: "ASME B31.3/B31.4/B31.8", Edition: "Project approved 2025", ProductForm: "Pipe"},
		{Name: "API 5L X 42", Grade: "X42", SMYS: 42000, AllowableStress: 20000, Code: "ASME B31.3/B31.4/B31.8", Edition: "Project approved 2025", ProductForm: "Pipe"},
		{Name: "API 5L X 46", Grade: "X46", SMYS: 46000, AllowableStress: 20000, Code: "ASME B31.3/B31.4/B31.8", Edition: "Project approved 2025", ProductForm: "Pipe"},
		{Name: "API 5L X 52", Grade: "X52", SMYS: 52000, AllowableStress: 20000, Code: "ASME B31.3/B31.4/B31.8", Edition: "Project approved 2025", ProductForm: "Pipe"},
		{Name: "API 5L X 56", Grade: "X56", SMYS: 56000, AllowableStress: 20000, Code: "ASME B31.3/B31.4/B31.8", Edition: "Project approved 2025", ProductForm: "Pipe"},
		{Name: "API 5L X 65", Grade: "X65", SMYS: 65000, AllowableStress: 20000, Code: "ASME B31.3/B31.4/B31.8", Edition: "Project approved 2025", ProductForm: "Pipe"},
		{Name: "API 5L X 70", Grade: "X70", SMYS: 70000, AllowableStress: 20000, Code: "ASME B31.3/B31.4/B31.8", Edition: "Project approved 2025", ProductForm: "Pipe"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO pipeline_materials (name, material_specification, grade, smys, allowable_stress, notes, is_active, stress_source, stress_dataset_version, governing_code, code_edition, product_form, temperature_min_f, temperature_max_f)
		VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, eq := range pipelineMaterial {
		_, err := stmt.Exec(eq.Name, eq.Name, eq.Grade, eq.SMYS, eq.AllowableStress, "Pipeline-specific material allowable-stress dataset; no pressure-vessel fallback permitted.", PipelineMaterialStressDatasetSource, PipelineMaterialStressDatasetVersion, eq.Code, eq.Edition, eq.ProductForm, -20, 400)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
