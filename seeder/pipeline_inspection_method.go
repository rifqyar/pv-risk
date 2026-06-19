package seeder

import "database/sql"

func PipelineInspectionMethod(db *sql.DB) error {
	methods := []struct {
		Name        string
		Scope       string
		Effectivity string
	}{
		{Name: "Visual + CP / Coating Survey", Scope: "nonintrusive", Effectivity: "Medium"},
		{Name: "ROW Patrol + Visual Survey", Scope: "nonintrusive", Effectivity: "Medium"},
		{Name: "Wall Thickness measurement by UT", Scope: "nonintrusive", Effectivity: "Medium"},
		{Name: "Shear Wave Ultrasonic Testing", Scope: "nonintrusive", Effectivity: "Medium"},
		{Name: "Radiographic Testing", Scope: "nonintrusive", Effectivity: "Low"},
		{Name: "Wet Fluorescent MPT / DPT", Scope: "intrusive", Effectivity: "High"},
		{Name: "Direct Examination", Scope: "intrusive", Effectivity: "High"},
		{Name: "VIE + Wall Thickness measurement by UT", Scope: "intrusive", Effectivity: "High"},
		{Name: "None", Scope: "nonintrusive", Effectivity: "None"},
		{Name: "None", Scope: "intrusive", Effectivity: "None"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO pipeline_inspection_methods (name, scope, effectivity, notes, is_active)
		VALUES (?, ?, ?, ?, 1)
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, method := range methods {
		if _, err := stmt.Exec(method.Name, method.Scope, method.Effectivity, "Seeded default inspection method; editable in Pipeline Master Data."); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
