package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func Migrate(db *sql.DB) {
	var err error

	createAssessments := `
	CREATE TABLE IF NOT EXISTS assessments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tag_number TEXT,
		year_built INTEGER,
		thickness_actual REAL,
		thickness_min REAL,
		corrosion_rate REAL,
		operating_pressure REAL,
		fluid_type TEXT,
		is_critical INTEGER,
		damage_mechanism TEXT,
		damage_factor INTEGER,
		inspection_score INTEGER,
		inspection_quality TEXT,
		remaining_life REAL,
		lof INTEGER,
		cof INTEGER,
		risk_index INTEGER,
		risk_level TEXT,
		next_inspection INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_tag_number 
	ON assessments(tag_number);`

	createHistory := `
	CREATE TABLE IF NOT EXISTS thickness_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tag_number TEXT,
		thickness REAL,
		measured_at DATETIME
	);
	
	CREATE INDEX IF NOT EXISTS idx_history_tag 
	ON thickness_history(tag_number);`

	createEquipment := `
	CREATE TABLE IF NOT EXISTS equipments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		type TEXT NOT NULL,
		group_name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createMaterial := `
	CREATE TABLE IF NOT EXISTS design_materials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    equipment_id INTEGER,
    tag_number TEXT,
    year_built INTEGER,
    design_code TEXT,
    design_pressure REAL,
    design_temperature REAL,
    operating_pressure REAL,
    operating_temperature REAL,
    mdmt REAL,
    shell_material TEXT,
    head_material TEXT,
    corrosion_allowance REAL,
    joint_efficiency REAL,
    weld_type TEXT,
    has_cladding INTEGER,
    cladding_material TEXT,
    nominal_thickness REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (equipment_id) REFERENCES equipments(id)
	);`

	createTypeHead := `
	CREATE TABLE IF NOT EXISTS head_types (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createShellMaterial := `
	CREATE TABLE IF NOT EXISTS shell_material (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err = config.DB.Exec(createAssessments); err != nil {
		log.Fatal(err)
	}

	if _, err = config.DB.Exec(createHistory); err != nil {
		log.Fatal(err)
	}

	if _, err = config.DB.Exec(createEquipment); err != nil {
		log.Fatal(err)
	}

	if _, err = config.DB.Exec(createMaterial); err != nil {
		log.Fatal(err)
	}

	if _, err = config.DB.Exec(createTypeHead); err != nil {
		log.Fatal(err)
	}

	if _, err = config.DB.Exec(createShellMaterial); err != nil {
		log.Fatal(err)
	}

	NozzleMaterial(config.DB)
	NeckMaterial(config.DB)
	CorrosionResistantMigration(config.DB)
	MicNonResistantMigration(config.DB)
	DamageMechanicalImpact(config.DB)
	Fluida(config.DB)
	Phase(config.DB)
	PollutioMigration(config.DB)
	VelocityMigration(config.DB)
	PHCategory(config.DB)
	H2SContent(config.DB)
	CISCCMatrix(config.DB)
	CO2CorrosionPreventive((config.DB))
	InhibitorInjectionReliability(config.DB)
	ReleaseProduct(config.DB)
	CleanupTime(config.DB)
	ChlorideContent(config.DB)
}
