package migrations

import (
	"database/sql"
	"log"
)

func PipelineOilReportingTables(db *sql.DB) {
	versionQuery := `
	CREATE TABLE IF NOT EXISTS pipeline_oil_assessment_versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		assessment_id INTEGER NOT NULL,
		version_number INTEGER NOT NULL,
		status TEXT NOT NULL,
		formula_version TEXT NOT NULL,
		input_json TEXT NOT NULL,
		result_json TEXT,
		formula_trace_json TEXT,
		snapshot_json TEXT NOT NULL,
		created_by TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(assessment_id, version_number),
		FOREIGN KEY (assessment_id) REFERENCES pipeline_oil_assessments(id)
	);`
	if _, err := db.Exec(versionQuery); err != nil {
		log.Fatalf("Error creating pipeline_oil_assessment_versions table: %v", err)
	}

	auditQuery := `
	CREATE TABLE IF NOT EXISTS pipeline_oil_audit_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		assessment_id INTEGER NOT NULL,
		version_id INTEGER,
		action TEXT NOT NULL,
		actor TEXT,
		affected_fields_json TEXT,
		old_values_json TEXT,
		new_values_json TEXT,
		note TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (assessment_id) REFERENCES pipeline_oil_assessments(id),
		FOREIGN KEY (version_id) REFERENCES pipeline_oil_assessment_versions(id)
	);`
	if _, err := db.Exec(auditQuery); err != nil {
		log.Fatalf("Error creating pipeline_oil_audit_events table: %v", err)
	}

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_pipeline_oil_versions_assessment ON pipeline_oil_assessment_versions(assessment_id, version_number);`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_pipeline_oil_versions_created ON pipeline_oil_assessment_versions(assessment_id, created_at);`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_pipeline_oil_audit_assessment ON pipeline_oil_audit_events(assessment_id, created_at);`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_pipeline_oil_audit_action ON pipeline_oil_audit_events(action);`)

	log.Println("Pipeline Oil reporting tables migrated successfully")
}
