package migrations

import (
	"database/sql"
	"log"
)

// 1. Tabel Master Equipment (Header)
func EquipmentsTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS trx_equipments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		equipment_id INTEGER UNIQUE NOT NULL, -- FK ke master equipments.id
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
		operating_pressure_tube REAL,
		operating_temp_tube REAL,    
		phase TEXT,
		h2s_content REAL,
		co2_content REAL,
		h2o_content REAL,
		chloride_index INTEGER,
		ph_index INTEGER,
		
		-- TAMBAHAN STEP 3:
		impact_production TEXT,
		insulation_condition TEXT,
		insulation_damage_level TEXT,
		coating_condition TEXT,
		coating_damage_level TEXT,
		corrective_description TEXT,
		corrective_action TEXT,
		corrective_date DATE,

		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (equipment_id) REFERENCES trx_equipments(id) ON DELETE CASCADE
	);`
	db.Exec(query)
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
		galvanic TEXT,     -- TAMBAHAN STEP 4
		lof_score TEXT,    -- DIUBAH DARI FLOAT KE TEXT (Biar bisa nampung "PoF: 1.5E-3 (DF: 50)")
		FOREIGN KEY (assessment_id) REFERENCES assessments(id) ON DELETE CASCADE
	);`
	db.Exec(query)
}

// 5. Tabel Final Result & Strategy (Sub-Detail)
func AssessmentResultsTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS assessment_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		assessment_id INTEGER,
		lof_category INTEGER,
		cof_financial TEXT, -- TAMBAHAN STEP 5
		cof_safety TEXT,    -- TAMBAHAN STEP 5
		cof_category TEXT,
		risk_level TEXT,
		risk_index INTEGER,

		insp_internal_thinning TEXT,  -- TAMBAHAN STEP 5 (Inspection Effectiveness)
		insp_external_corrosion TEXT, -- TAMBAHAN STEP 5
		insp_cracking TEXT,           -- TAMBAHAN STEP 5

		governing_component TEXT,
		max_interval_years REAL,
		next_inspection_year INTEGER,
		recommended_method TEXT,
		FOREIGN KEY (assessment_id) REFERENCES assessments(id) ON DELETE CASCADE
	);`
	db.Exec(query)
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
