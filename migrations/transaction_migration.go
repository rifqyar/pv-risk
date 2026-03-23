package migrations

import (
	"database/sql"
	"log"
)

// 1. Tabel Master Equipment (Header)
func EquipmentsTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS equipments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tag_number TEXT UNIQUE NOT NULL,
		description TEXT,
		equipment_type_id TEXT,
		year_built INTEGER,
		shell_material_id INTEGER,
		design_pressure REAL,
		design_temp REAL,
		diameter REAL,
		volume REAL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Error creating equipments table: %v", err)
	}
}

// 2. Tabel General Assessment (Detail Utama)
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
		phase TEXT,
		h2s_content REAL,
		co2_content REAL,
		h2o_content REAL,
		chloride_index INTEGER,
		ph_index INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (equipment_id) REFERENCES equipments(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Error creating assessments table: %v", err)
	}
}

// 3. Tabel Data Ketebalan (Sub-Detail)
func AssessmentThicknessesTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS assessment_thicknesses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		assessment_id INTEGER,
		component_type TEXT, -- isinya: 'shell', 'head', atau 'nozzle'
		prev_thick REAL,
		act_thick REAL,
		t_req REAL,
		corrosion_rate REAL,
		remaining_life REAL,
		FOREIGN KEY (assessment_id) REFERENCES assessments(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Error creating assessment_thicknesses table: %v", err)
	}
}

// 4. Tabel Damage Mechanism (Sub-Detail)
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
		total_damage_factor REAL,
		FOREIGN KEY (assessment_id) REFERENCES assessments(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Error creating assessment_damage_mechanisms table: %v", err)
	}
}

// 5. Tabel Final Result & Strategy (Sub-Detail)
func AssessmentResultsTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS assessment_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		assessment_id INTEGER,
		lof_category INTEGER,
		cof_category TEXT,
		risk_level TEXT,
		risk_index INTEGER,
		governing_component TEXT,
		max_interval_years REAL,
		next_inspection_year INTEGER,
		recommended_method TEXT,
		FOREIGN KEY (assessment_id) REFERENCES assessments(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Error creating assessment_results table: %v", err)
	}
}

func RunAllAssessmentMigrations(db *sql.DB) {
	// PENTING: Untuk mengaktifkan fitur ON DELETE CASCADE di SQLite
	// harus eksekusi pragma ini setiap kali koneksi baru dibuka
	db.Exec("PRAGMA foreign_keys = ON;")

	EquipmentsTable(db)
	AssessmentsTable(db)
	AssessmentThicknessesTable(db)
	AssessmentDamageMechanismsTable(db)
	AssessmentResultsTable(db)

	log.Println("✅ All Assessment Tables Migrated Successfully!")
}
