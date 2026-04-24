package seeder

import (
	"database/sql"
)

func SeedEquipment(db *sql.DB) error {
	equipments := []struct {
		Name  string
		Type  string
		Group string
	}{
		{Name: "Accumulator", Type: "EQT1", Group: "Pressure Vessel"},
		{Name: "Air Dryer Unit", Type: "EQT1", Group: "Package & Utility Unit"},
		{Name: "Amine Absorber", Type: "EQT3", Group: "Column / Tower"},
		{Name: "Amine Charcoal Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "Amine Flash", Type: "EQT2", Group: "Column / Tower"},
		{Name: "Amine Particulate After Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "Amine Particulate Pre Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "Amine Reboiler", Type: "EQT3", Group: "Heat Exchanger & Heater"},
		{Name: "Amine Regeneration", Type: "EQT2", Group: "Column / Tower"},
		{Name: "Amine Sump Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "Amine Sump Vessel", Type: "EQT1", Group: "Pressure Vessel"},
		{Name: "Anti Foam Tank", Type: "EQT1", Group: "Pressure Vessel"},
		{Name: "Carbon Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "Cold Glycol/Heat Exchanger", Type: "EQT3", Group: "Heat Exchanger & Heater"},
		{Name: "Diesel Fuel Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "First Stage Electric Heater", Type: "EQT1", Group: "Heat Exchanger & Heater"},
		{Name: "First Stage Filter Coalescer", Type: "EQT1", Group: "Filtration System"},
		{Name: "First Stage Guard Bed", Type: "EQT1", Group: "Adsorber / Reactor"},
		{Name: "First Stage Particle Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "Flare K O Drum", Type: "EQT1", Group: "Pressure Vessel"},
		{Name: "Fuel Gas Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "Fuel gas spot", Type: "EQT1", Group: "Package & Utility Unit"},

		// --- UPDATED KE EQT1 ---
		{Name: "Gas Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "Gas Pig Launcher", Type: "EQT1", Group: "Pipeline Equipment"},
		// -----------------------

		{Name: "gas/gas exchanger", Type: "EQT3", Group: "Heat Exchanger & Heater"},
		{Name: "Gas/glycol Heat Exchanger", Type: "EQT3", Group: "Heat Exchanger & Heater"},
		{Name: "Glycol Carbon Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "Glycol Contractor w/integral", Type: "EQT1", Group: "Column / Tower"},
		{Name: "Glycol Flash Separator", Type: "EQT1", Group: "Pressure Vessel"},
		{Name: "Glycol Reflux condenser", Type: "EQT1", Group: "Heat Exchanger & Heater"},
		{Name: "Glycol sock filter", Type: "EQT1", Group: "Filtration System"},

		// --- UPDATED KE EQT1 ---
		{Name: "Glycol Still Colomn", Type: "EQT1", Group: "Column / Tower"},
		// -----------------------

		{Name: "Glycol Sump Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "Glycol Surge Tank", Type: "EQT1", Group: "Pressure Vessel"},

		// --- UPDATED KE EQT1 ---
		{Name: "H2S Scavenger", Type: "EQT1", Group: "Adsorber / Reactor"},
		// -----------------------

		{Name: "Hot Glycol/Heat Exchanger", Type: "EQT3", Group: "Heat Exchanger & Heater"},
		{Name: "Hot Oil Circulation Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "Hot Oil Expansion Vessel", Type: "EQT1", Group: "Pressure Vessel"},
		{Name: "Hot Oil Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "HP Fuel Gas Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "Hp Fuel Gas Scrubber", Type: "EQT1", Group: "Pressure Vessel"},
		{Name: "Inlet Separator", Type: "EQT1", Group: "Pressure Vessel"},
		{Name: "Instrument Air receiver", Type: "EQT1", Group: "Pressure Vessel"},
		{Name: "Lean Amine Cooler", Type: "EQT2", Group: "Heat Exchanger & Heater"},
		{Name: "LP Fuel Gas Scrubber", Type: "EQT1", Group: "Pressure Vessel"},

		// --- UPDATED KE EQT1 ---
		{Name: "Produced WTR H2S stripper", Type: "EQT1", Group: "Column / Tower"},
		// -----------------------

		{Name: "Production Separator", Type: "EQT1", Group: "Pressure Vessel"},
		{Name: "Sand Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "Second Stage Particle Filter", Type: "EQT1", Group: "Filtration System"},
		{Name: "Second Stage Electric Heater", Type: "EQT1", Group: "Heat Exchanger & Heater"},
		{Name: "Second Stage Filter Coalesser", Type: "EQT1", Group: "Filtration System"},
		{Name: "Second Stage Guard Bed", Type: "EQT1", Group: "Adsorber / Reactor"},

		// --- UPDATED KE EQT1 ---
		{Name: "Solvent amine lean/rich amine", Type: "EQT1", Group: "Package & Utility Unit"},
		// -----------------------

		{Name: "Solvent Recovery Drum", Type: "EQT1", Group: "Pressure Vessel"},
		{Name: "Sweet Gas Cooler", Type: "EQT1", Group: "Heat Exchanger & Heater"},
		{Name: "Sweet Gas KO Drum", Type: "EQT1", Group: "Pressure Vessel"},
		{Name: "Test Separator", Type: "EQT1", Group: "Pressure Vessel"},
		{Name: "Utility Air Receiver", Type: "EQT1", Group: "Pressure Vessel"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Loop dan eksekusi Update, lalu Insert jika belum ada
	for _, eq := range equipments {
		// 1. Coba lakukan UPDATE terlebih dahulu
		res, err := tx.Exec(`UPDATE equipments SET type = ?, group_name = ? WHERE name = ?`, eq.Type, eq.Group, eq.Name)
		if err != nil {
			tx.Rollback()
			return err
		}

		// 2. Cek apakah ada baris yang berhasil diupdate
		affected, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return err
		}

		// 3. Jika affected == 0, berarti equipment belum ada di database, maka INSERT
		if affected == 0 {
			_, err = tx.Exec(`INSERT INTO equipments (name, type, group_name) VALUES (?, ?, ?)`, eq.Name, eq.Type, eq.Group)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}
