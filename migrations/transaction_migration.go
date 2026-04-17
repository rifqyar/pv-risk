package migrations

import (
	"database/sql"
	"fmt"
	"log"
)

// ==========================================
// HELPER: SMART MIGRATION (Biar Kodingan Nggak Berulang)
// ==========================================
func checkAndAddColumn(db *sql.DB, tableName, columnName, columnType string) {
	query := fmt.Sprintf("PRAGMA table_info(%s);", tableName)
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Gagal cek tabel %s: %v", tableName, err)
		return
	}
	defer rows.Close()

	exists := false
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue interface{}

		err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err == nil && name == columnName {
			exists = true
			break
		}
	}

	if !exists {
		alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", tableName, columnName, columnType)
		_, err := db.Exec(alterQuery)
		if err != nil {
			log.Printf("Gagal nambahin kolom %s ke %s: %v", columnName, tableName, err)
		} else {
			log.Printf("✅ Auto-Migration: Berhasil nambah kolom '%s' ke tabel '%s'", columnName, tableName)
		}
	}
}

// ==========================================
// 1. Tabel Master Equipment (Header)
// ==========================================
func EquipmentsTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS trx_equipments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		equipment_id INTEGER UNIQUE NOT NULL,
		tag_number TEXT UNIQUE NOT NULL,
		year_built INTEGER,
		shell_material_id INTEGER,
		design_pressure REAL,
		design_pressure_tube REAL,
		design_temp REAL,
		design_temp_tube REAL,
		diameter REAL,
		diameter_tube REAL,
		volume REAL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (equipment_id) REFERENCES equipments(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Error creating trx_equipments table: %v", err)
	}

	// Add missing columns (DIUBAH KE TEXT AGAR SESUAI DENGAN VALUE STRING)
	checkAndAddColumn(db, "trx_equipments", "location", "TEXT DEFAULT ''")
	checkAndAddColumn(db, "trx_equipments", "diameter_type", "TEXT DEFAULT 'inside'")
	checkAndAddColumn(db, "trx_equipments", "diameter_unit", "TEXT DEFAULT 'inch'")
	checkAndAddColumn(db, "trx_equipments", "diameter_tube_type", "TEXT DEFAULT 'inside'")
	checkAndAddColumn(db, "trx_equipments", "diameter_tube_unit", "TEXT DEFAULT 'inch'")
	checkAndAddColumn(db, "trx_equipments", "length", "REAL DEFAULT 0")
	checkAndAddColumn(db, "trx_equipments", "length_unit", "TEXT DEFAULT 'ft'")
	checkAndAddColumn(db, "trx_equipments", "volume_unit", "TEXT DEFAULT 'm'")
	checkAndAddColumn(db, "trx_equipments", "temp_design_unit", "TEXT DEFAULT 'c'")
	checkAndAddColumn(db, "trx_equipments", "temp_design_tube_unit", "TEXT DEFAULT 'c'")
	checkAndAddColumn(db, "trx_equipments", "pwht", "TEXT DEFAULT 'No'")
	checkAndAddColumn(db, "trx_equipments", "certificate", "TEXT DEFAULT '-'")
	checkAndAddColumn(db, "trx_equipments", "data_reference", "TEXT DEFAULT '-'")
	checkAndAddColumn(db, "trx_equipments", "nozzle", "REAL DEFAULT 0")
	checkAndAddColumn(db, "trx_equipments", "nozzle_unit", "TEXT DEFAULT 'inch'")
	checkAndAddColumn(db, "trx_equipments", "phase_type", "TEXT DEFAULT 'multi phase'")
	checkAndAddColumn(db, "trx_equipments", "internal_lining", "TEXT DEFAULT 'None'")
	checkAndAddColumn(db, "trx_equipments", "insulation", "TEXT DEFAULT 'No'")
	checkAndAddColumn(db, "trx_equipments", "special_service", "TEXT DEFAULT '-'")
	checkAndAddColumn(db, "trx_equipments", "protection", "TEXT DEFAULT '-'")
	checkAndAddColumn(db, "trx_equipments", "cathodic_protection", "TEXT DEFAULT 'No'")
}

// ==========================================
// 2. Tabel General Assessment (Detail Utama)
// ==========================================
func AssessmentsTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS assessments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		equipment_id INTEGER,
		assessment_date DATE,
		prev_inspection_date DATE,
		act_inspection_date DATE,
		operating_pressure REAL,     
		operating_temp REAL,         
		operating_pressure_tube REAL,
		operating_temp_tube REAL,    
		phase TEXT,
		h2s_content REAL,
		co2_content REAL,
		h2o_content REAL,
		chloride_index INTEGER,
		ph_index INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (equipment_id) REFERENCES trx_equipments(id) ON DELETE CASCADE
	);`
	db.Exec(query)

	// Add missing columns via Smart Migration
	checkAndAddColumn(db, "assessments", "temp_op_unit", "TEXT DEFAULT 'c'")
	checkAndAddColumn(db, "assessments", "temp_op_tube_unit", "TEXT DEFAULT 'c'")

	// TAMBAHAN STEP 3:
	checkAndAddColumn(db, "assessments", "impact_production", "TEXT DEFAULT ''")
	checkAndAddColumn(db, "assessments", "insulation_condition", "TEXT DEFAULT ''")
	checkAndAddColumn(db, "assessments", "insulation_damage_level", "TEXT DEFAULT ''")
	checkAndAddColumn(db, "assessments", "coating_condition", "TEXT DEFAULT ''")
	checkAndAddColumn(db, "assessments", "coating_damage_level", "TEXT DEFAULT ''")
	checkAndAddColumn(db, "assessments", "corrective_description", "TEXT DEFAULT ''")
	checkAndAddColumn(db, "assessments", "corrective_action", "TEXT DEFAULT ''")
	checkAndAddColumn(db, "assessments", "corrective_date", "DATE")
}

// ==========================================
// 3. Tabel Data Ketebalan (Sub-Detail)
// ==========================================
func AssessmentThicknessesTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS assessment_thicknesses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		assessment_id INTEGER,
		component_type TEXT, 
		prev_thick REAL,
		act_thick REAL,
		t_req REAL,
		corrosion_rate REAL,
		remaining_life REAL,
		FOREIGN KEY (assessment_id) REFERENCES assessments(id) ON DELETE CASCADE
	);`
	db.Exec(query)
}

// ==========================================
// 4. Tabel Damage Mechanism (Sub-Detail)
// ==========================================
func AssessmentDamageMechanismsTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS assessment_damage_mechanisms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		assessment_id INTEGER,
		atmospheric TEXT,
		cui TEXT,
		ext_cracking TEXT,
		co2 TEXT,
		mic TEXT,
		ssc TEXT,
		amine_scc TEXT,
		hic TEXT,
		ciscc TEXT,
		FOREIGN KEY (assessment_id) REFERENCES assessments(id) ON DELETE CASCADE
	);`
	db.Exec(query)

	// Tambahan Step 4
	checkAndAddColumn(db, "assessment_damage_mechanisms", "galvanic", "TEXT")
	checkAndAddColumn(db, "assessment_damage_mechanisms", "lof_score", "TEXT")
}

// ==========================================
// 5. Tabel Final Result & Strategy (Sub-Detail)
// ==========================================
func AssessmentResultsTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS assessment_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		assessment_id INTEGER,
		lof_category INTEGER,
		risk_level TEXT,
		risk_index INTEGER,
		governing_component TEXT, 
		max_interval_years REAL,
		next_inspection_year INTEGER,
		recommended_method TEXT,
		FOREIGN KEY (assessment_id) REFERENCES assessments(id) ON DELETE CASCADE
	);`
	db.Exec(query)

	// Tambahan Step 5 (Termasuk kolom JSON baru)
	checkAndAddColumn(db, "assessment_results", "cof_financial", "TEXT")
	checkAndAddColumn(db, "assessment_results", "cof_safety", "TEXT")
	checkAndAddColumn(db, "assessment_results", "cof_category", "TEXT")

	// FIX TERPENTING: JSON untuk nampung semua form inspection yg kompleks
	checkAndAddColumn(db, "assessment_results", "inspection_plan_json", "TEXT")
	checkAndAddColumn(db, "assessment_results", "cladding_json", "TEXT")
}

func RunAllAssessmentMigrations(db *sql.DB) {
	// PENTING: Untuk mengaktifkan fitur ON DELETE CASCADE di SQLite
	db.Exec("PRAGMA foreign_keys = ON;")

	EquipmentsTable(db)
	AssessmentsTable(db)
	AssessmentThicknessesTable(db)
	AssessmentDamageMechanismsTable(db)
	AssessmentResultsTable(db)

	log.Println("✅ All Assessment Tables Migrated Successfully!")
}
