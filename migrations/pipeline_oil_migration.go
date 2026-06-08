package migrations

import (
	"database/sql"
	"log"
)

func PipelineOilAssessmentTables(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS pipeline_oil_assessments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		status TEXT NOT NULL DEFAULT 'DRAFT',
		report_no TEXT,
		line_identification TEXT,
		owner_user TEXT,
		location TEXT,
		service TEXT,
		assessment_by TEXT,
		formula_version TEXT NOT NULL,
		input_json TEXT NOT NULL,
		result_json TEXT,
		formula_trace_json TEXT,
		snapshot_json TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_by TEXT,
		updated_by TEXT
	);`
	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Error creating pipeline_oil_assessments table: %v", err)
	}

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_pipeline_oil_status ON pipeline_oil_assessments(status);`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_pipeline_oil_report_no ON pipeline_oil_assessments(report_no);`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_pipeline_oil_line ON pipeline_oil_assessments(line_identification);`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_pipeline_oil_updated_at ON pipeline_oil_assessments(updated_at);`)

	log.Println("Pipeline Oil assessment tables migrated successfully")
}
