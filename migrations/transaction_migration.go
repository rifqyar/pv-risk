package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
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

	// helper function cek kolom
	columnExists := func(columnName string) bool {
		rows, err := db.Query(`PRAGMA table_info(trx_equipments);`)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue interface{}

		for rows.Next() {
			err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
			if err != nil {
				log.Fatal(err)
			}
			if name == columnName {
				return true
			}
		}
		return false
	}

	columns := []string{
		// sebelumnya
		"location REAL DEFAULT ''",
		"diameter_type REAL DEFAULT 'inside'",
		"diameter_unit REAL DEFAULT 'inch'",
		"diameter_tube_type REAL DEFAULT 'inside'",
		"diameter_tube_unit REAL DEFAULT 'inch'",
		"length FLOAT DEFAULT 0",
		"length_unit REAL DEFAULT 'ft'",
		"volume_unit REAL DEFAULT 'm'",
		"temp_design_unit REAL DEFAULT 'c'",
		"temp_design_tube_unit REAL DEFAULT 'c'",
		"pwht REAL DEFAULT 'No'",
		"certificate REAL DEFAULT '-'",
		"data_reference REAL DEFAULT '-'",
		"nozzle FLOAT DEFAULT 0",
		"nozzle_unit REAL DEFAULT 'inch'",
		"phase_type REAL DEFAULT 'multi phase'",
		"internal_lining REAL DEFAULT 'None'",
		"insulation REAL DEFAULT 'No'",
		"special_service REAL DEFAULT '-'",
		"protection REAL DEFAULT '-'",
		"cathodic_protection REAL DEFAULT 'No'",
	}

	// add column kalau belum ada
	for _, col := range columns {
		colName := strings.Split(col, " ")[0]

		if !columnExists(colName) {
			alterQuery := fmt.Sprintf("ALTER TABLE trx_equipments ADD COLUMN %s;", col)
			if _, err := db.Exec(alterQuery); err != nil {
				log.Fatalf("Error adding column %s: %v", colName, err)
			}
		}
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

	// helper function cek kolom
	columnExists := func(columnName string) bool {
		rows, err := db.Query(`PRAGMA table_info(assessments);`)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue interface{}

		for rows.Next() {
			err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
			if err != nil {
				log.Fatal(err)
			}
			if name == columnName {
				return true
			}
		}
		return false
	}

	columns := []string{
		"temp_op_unit REAL DEFAULT 'c'",
		"temp_op_tube_unit REAL DEFAULT 'c'",
	}

	// add column kalau belum ada
	for _, col := range columns {
		colName := strings.Split(col, " ")[0]

		if !columnExists(colName) {
			alterQuery := fmt.Sprintf("ALTER TABLE assessments ADD COLUMN %s;", col)
			if _, err := db.Exec(alterQuery); err != nil {
				log.Fatalf("Error adding column %s: %v", colName, err)
			}
		}
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
