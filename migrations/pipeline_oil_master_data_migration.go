package migrations

import (
	"database/sql"
	"log"
)

func PipelineOilMasterDataTables(db *sql.DB) {
	addPipelineColumn(db, "pipeline_materials", "material_specification", "TEXT")
	addPipelineColumn(db, "pipeline_materials", "grade", "TEXT")
	addPipelineColumn(db, "pipeline_materials", "allowable_stress", "REAL NOT NULL DEFAULT 0")
	addPipelineColumn(db, "pipeline_materials", "notes", "TEXT")
	addPipelineColumn(db, "pipeline_materials", "is_active", "INTEGER NOT NULL DEFAULT 1")
	addPipelineColumn(db, "pipeline_materials", "updated_at", "DATETIME")
	if _, err := db.Exec(`UPDATE pipeline_materials SET material_specification = name WHERE material_specification IS NULL OR material_specification = ''`); err != nil {
		log.Fatalf("Error backfilling pipeline_materials material_specification: %v", err)
	}
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_pipeline_materials_active ON pipeline_materials(is_active)`); err != nil {
		log.Fatalf("Error indexing pipeline_materials active flag: %v", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS pipeline_inspection_methods (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		scope TEXT NOT NULL DEFAULT 'nonintrusive',
		effectivity TEXT NOT NULL DEFAULT 'Medium',
		notes TEXT,
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME,
		UNIQUE(name, scope)
	);`
	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Error creating pipeline_inspection_methods table: %v", err)
	}
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_pipeline_inspection_methods_active_scope ON pipeline_inspection_methods(is_active, scope)`); err != nil {
		log.Fatalf("Error indexing pipeline_inspection_methods: %v", err)
	}

	log.Println("Pipeline Oil master data tables migrated successfully")
}

func addPipelineColumn(db *sql.DB, table, column, definition string) {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		log.Fatalf("Error reading %s schema: %v", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull int
		var defaultValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			log.Fatalf("Error scanning %s schema: %v", table, err)
		}
		if name == column {
			return
		}
	}
	if _, err := db.Exec("ALTER TABLE " + table + " ADD COLUMN " + column + " " + definition); err != nil {
		log.Fatalf("Error adding %s.%s: %v", table, column, err)
	}
}
