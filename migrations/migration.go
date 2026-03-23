package migrations

import (
	"database/sql"
	"log"
	"pv-risk/config"
)

func Migrate(db *sql.DB) {
	var err error

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

	// Transaction Migration
	RunAllAssessmentMigrations(config.DB)
}
