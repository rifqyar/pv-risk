package seeder

import (
	"database/sql"
)

func SeedAll(db *sql.DB) error {
	if err := SeedEquipment(db); err != nil {
		return err
	}

	if err := SeedTypeHead(db); err != nil {
		return err
	}

	if err := ShellMaterial(db); err != nil {
		return err
	}

	if err := NozzleMaterial(db); err != nil {
		return err
	}

	if err := NeckMaterial(db); err != nil {
		return err
	}

	if err := CorrosionResistant(db); err != nil {
		return err
	}

	if err := MicNonResistantSeeder(db); err != nil {
		return err
	}

	if err := DamageMechanicalImpact(db); err != nil {
		return err
	}

	if err := Fluida(db); err != nil {
		return err
	}

	if err := PhaseSeeder(db); err != nil {
		return err
	}

	if err := Pollution(db); err != nil {
		return err
	}

	if err := VelocitySeeder(db); err != nil {
		return err
	}

	if err := PHCategory(db); err != nil {
		return err
	}

	if err := H2SContent(db); err != nil {
		return err
	}

	if err := CISCCMatrix(db); err != nil {
		return err
	}

	if err := CO2CorrosionPreventive(db); err != nil {
		return err
	}

	if err := InhibitorInjectionReliability(db); err != nil {
		return err
	}

	if err := ReleaseProduct(db); err != nil {
		return err
	}

	if err := CleanupTime(db); err != nil {
		return err
	}

	if err := ChlorideContent(db); err != nil {
		return err
	}

	return nil
}
