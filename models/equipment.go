package models

import "database/sql"

type Equipment struct {
	ID        int
	Name      string
	Type      string
	GroupName string
}

func GetEquipments(db *sql.DB) ([]Equipment, error) {
	rows, err := db.Query("SELECT id, name, type, group_name type FROM equipments")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var equipments []Equipment

	for rows.Next() {
		var eq Equipment
		err := rows.Scan(&eq.ID, &eq.Name, &eq.Type, &eq.GroupName)
		if err != nil {
			return nil, err
		}
		equipments = append(equipments, eq)
	}

	return equipments, nil
}
