package controller

import (
	"database/sql"
	"net/http"
	"net/url"
	"pv-risk/config"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type MasterEquipmentRow struct {
	ID         int
	Name       string
	Type       string
	GroupName  string
	UsageCount int
}

type MasterShellMaterialRow struct {
	ID              int
	Name            string
	External        string
	Internal        string
	CO2Corr         string
	MIC             string
	AmineCracking   string
	SulfideCracking string
	UsageCount      int
}

type MasterSimpleMaterialRow struct {
	ID         int
	Name       string
	CladStatus string
	UsageCount int
}

func ShowMasterData(c *gin.Context) {
	equipments, err := getMasterEquipments(config.DB)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching equipments: %v", err)
		return
	}

	shellMaterials, err := getMasterShellMaterials(config.DB)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching shell materials: %v", err)
		return
	}

	neckMaterials, err := getSimpleMaterials(config.DB, "neck_material", "neck_material_id")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching neck materials: %v", err)
		return
	}

	nozzleMaterials, err := getSimpleMaterials(config.DB, "nozzle_material", "nozzle_material_id")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching nozzle materials: %v", err)
		return
	}

	c.HTML(http.StatusOK, "master_data.html", gin.H{
		"ActiveMenu":      "master-data",
		"FlashStatus":     c.Query("status"),
		"FlashMessage":    c.Query("message"),
		"Equipments":      equipments,
		"ShellMaterials":  shellMaterials,
		"NeckMaterials":   neckMaterials,
		"NozzleMaterials": nozzleMaterials,
	})
}

func SaveMasterEquipment(c *gin.Context) {
	id := parseFormID(c.PostForm("id"))
	name := strings.TrimSpace(c.PostForm("name"))
	eqType := strings.TrimSpace(c.PostForm("type"))
	groupName := strings.TrimSpace(c.PostForm("group_name"))

	if name == "" || eqType == "" || groupName == "" {
		redirectMaster(c, "error", "Equipment name, type, and group are required.")
		return
	}

	var err error
	if id > 0 {
		_, err = config.DB.Exec(`UPDATE equipments SET name = ?, type = ?, group_name = ? WHERE id = ?`, name, eqType, groupName, id)
	} else {
		_, err = config.DB.Exec(`INSERT INTO equipments (name, type, group_name) VALUES (?, ?, ?)`, name, eqType, groupName)
	}

	if err != nil {
		redirectMaster(c, "error", "Failed to save equipment: "+err.Error())
		return
	}
	redirectMaster(c, "success", "Equipment saved successfully.")
}

func DeleteMasterEquipment(c *gin.Context) {
	id := parseFormID(c.Param("id"))
	if id <= 0 {
		redirectMaster(c, "error", "Invalid equipment ID.")
		return
	}

	var used int
	if err := config.DB.QueryRow(`SELECT COUNT(*) FROM trx_equipments WHERE equipment_id = ?`, id).Scan(&used); err != nil {
		redirectMaster(c, "error", "Failed to check equipment usage: "+err.Error())
		return
	}
	if used > 0 {
		redirectMaster(c, "error", "Equipment is already used in an assessment and cannot be deleted.")
		return
	}

	if _, err := config.DB.Exec(`DELETE FROM equipments WHERE id = ?`, id); err != nil {
		redirectMaster(c, "error", "Failed to delete equipment: "+err.Error())
		return
	}
	redirectMaster(c, "success", "Equipment deleted successfully.")
}

func SaveShellMaterial(c *gin.Context) {
	id := parseFormID(c.PostForm("id"))
	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		redirectMaster(c, "error", "Material name is required.")
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		redirectMaster(c, "error", "Failed to start transaction: "+err.Error())
		return
	}
	defer tx.Rollback()

	var materialID int64
	if id > 0 {
		if _, err = tx.Exec(`UPDATE shell_material SET name = ? WHERE id = ?`, name, id); err != nil {
			redirectMaster(c, "error", "Failed to update shell material: "+err.Error())
			return
		}
		materialID = int64(id)
	} else {
		res, err := tx.Exec(`INSERT INTO shell_material (name) VALUES (?)`, name)
		if err != nil {
			redirectMaster(c, "error", "Failed to create shell material: "+err.Error())
			return
		}
		materialID, _ = res.LastInsertId()
	}

	if _, err = tx.Exec(`
		INSERT INTO corrosion_resistant (id_shell_material, external, internal, co2_corr)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id_shell_material) DO UPDATE SET
			external = excluded.external,
			internal = excluded.internal,
			co2_corr = excluded.co2_corr
	`, materialID, c.PostForm("external"), c.PostForm("internal"), c.PostForm("co2_corr")); err != nil {
		redirectMaster(c, "error", "Failed to save corrosion data: "+err.Error())
		return
	}

	if _, err = tx.Exec(`
		INSERT INTO mic_resistant (id_shell_material, mic, amine_cracking, sulfide_cracking)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id_shell_material) DO UPDATE SET
			mic = excluded.mic,
			amine_cracking = excluded.amine_cracking,
			sulfide_cracking = excluded.sulfide_cracking
	`, materialID, c.PostForm("mic"), c.PostForm("amine_cracking"), c.PostForm("sulfide_cracking")); err != nil {
		redirectMaster(c, "error", "Failed to save cracking data: "+err.Error())
		return
	}

	if err = tx.Commit(); err != nil {
		redirectMaster(c, "error", "Failed to save material: "+err.Error())
		return
	}
	redirectMaster(c, "success", "Shell/head material saved successfully.")
}

func DeleteShellMaterial(c *gin.Context) {
	id := parseFormID(c.Param("id"))
	if id <= 0 {
		redirectMaster(c, "error", "Invalid material ID.")
		return
	}

	var used int
	if err := config.DB.QueryRow(`SELECT COUNT(*) FROM trx_equipments WHERE shell_material_id = ? OR head_material_id = ?`, id, id).Scan(&used); err != nil {
		redirectMaster(c, "error", "Failed to check material usage: "+err.Error())
		return
	}
	if used > 0 {
		redirectMaster(c, "error", "Material is already used in equipment data and cannot be deleted.")
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		redirectMaster(c, "error", "Failed to start transaction: "+err.Error())
		return
	}
	defer tx.Rollback()

	tx.Exec(`DELETE FROM corrosion_resistant WHERE id_shell_material = ?`, id)
	tx.Exec(`DELETE FROM mic_resistant WHERE id_shell_material = ?`, id)
	if _, err = tx.Exec(`DELETE FROM shell_material WHERE id = ?`, id); err != nil {
		redirectMaster(c, "error", "Failed to delete material: "+err.Error())
		return
	}
	if err = tx.Commit(); err != nil {
		redirectMaster(c, "error", "Failed to delete material: "+err.Error())
		return
	}
	redirectMaster(c, "success", "Shell/head material deleted successfully.")
}

func SaveSimpleMaterial(c *gin.Context) {
	id := parseFormID(c.PostForm("id"))
	kind := c.PostForm("kind")
	table, _ := simpleMaterialMeta(kind)
	if table == "" {
		redirectMaster(c, "error", "Invalid material type.")
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	cladStatus := strings.TrimSpace(c.PostForm("clad_status"))
	if name == "" || cladStatus == "" {
		redirectMaster(c, "error", "Material name and clad status are required.")
		return
	}

	var err error
	if id > 0 {
		_, err = config.DB.Exec(`UPDATE `+table+` SET name = ?, clad_status = ? WHERE id = ?`, name, cladStatus, id)
	} else {
		_, err = config.DB.Exec(`INSERT INTO `+table+` (name, clad_status) VALUES (?, ?)`, name, cladStatus)
	}
	if err != nil {
		redirectMaster(c, "error", "Failed to save material: "+err.Error())
		return
	}
	redirectMaster(c, "success", "Material saved successfully.")
}

func DeleteSimpleMaterial(c *gin.Context) {
	id := parseFormID(c.Param("id"))
	table, column := simpleMaterialMeta(c.Param("kind"))
	if id <= 0 || table == "" {
		redirectMaster(c, "error", "Invalid material ID.")
		return
	}

	var used int
	if err := config.DB.QueryRow(`SELECT COUNT(*) FROM trx_equipments WHERE `+column+` = ?`, id).Scan(&used); err != nil {
		redirectMaster(c, "error", "Failed to check material usage: "+err.Error())
		return
	}
	if used > 0 {
		redirectMaster(c, "error", "Material is already used in equipment data and cannot be deleted.")
		return
	}

	if _, err := config.DB.Exec(`DELETE FROM `+table+` WHERE id = ?`, id); err != nil {
		redirectMaster(c, "error", "Failed to delete material: "+err.Error())
		return
	}
	redirectMaster(c, "success", "Material deleted successfully.")
}

func getMasterEquipments(db *sql.DB) ([]MasterEquipmentRow, error) {
	rows, err := db.Query(`
		SELECT e.id, e.name, e.type, e.group_name, COUNT(t.id)
		FROM equipments e
		LEFT JOIN trx_equipments t ON t.equipment_id = e.id
		GROUP BY e.id, e.name, e.type, e.group_name
		ORDER BY e.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MasterEquipmentRow
	for rows.Next() {
		var item MasterEquipmentRow
		if err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.GroupName, &item.UsageCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func getMasterShellMaterials(db *sql.DB) ([]MasterShellMaterialRow, error) {
	rows, err := db.Query(`
		SELECT
			sm.id,
			sm.name,
			COALESCE(cr.external, ''),
			COALESCE(cr.internal, ''),
			COALESCE(cr.co2_corr, ''),
			COALESCE(mr.mic, ''),
			COALESCE(mr.amine_cracking, ''),
			COALESCE(mr.sulfide_cracking, ''),
			COUNT(t.id)
		FROM shell_material sm
		LEFT JOIN corrosion_resistant cr ON cr.id_shell_material = sm.id
		LEFT JOIN mic_resistant mr ON mr.id_shell_material = sm.id
		LEFT JOIN trx_equipments t ON t.shell_material_id = sm.id OR t.head_material_id = sm.id
		GROUP BY sm.id, sm.name, cr.external, cr.internal, cr.co2_corr, mr.mic, mr.amine_cracking, mr.sulfide_cracking
		ORDER BY sm.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MasterShellMaterialRow
	for rows.Next() {
		var item MasterShellMaterialRow
		if err := rows.Scan(&item.ID, &item.Name, &item.External, &item.Internal, &item.CO2Corr, &item.MIC, &item.AmineCracking, &item.SulfideCracking, &item.UsageCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func getSimpleMaterials(db *sql.DB, table string, usageColumn string) ([]MasterSimpleMaterialRow, error) {
	rows, err := db.Query(`
		SELECT m.id, m.name, m.clad_status, COUNT(t.id)
		FROM ` + table + ` m
		LEFT JOIN trx_equipments t ON t.` + usageColumn + ` = m.id
		GROUP BY m.id, m.name, m.clad_status
		ORDER BY m.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MasterSimpleMaterialRow
	for rows.Next() {
		var item MasterSimpleMaterialRow
		if err := rows.Scan(&item.ID, &item.Name, &item.CladStatus, &item.UsageCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func simpleMaterialMeta(kind string) (string, string) {
	switch kind {
	case "neck":
		return "neck_material", "neck_material_id"
	case "nozzle":
		return "nozzle_material", "nozzle_material_id"
	default:
		return "", ""
	}
}

func parseFormID(value string) int {
	id, _ := strconv.Atoi(value)
	return id
}

func redirectMaster(c *gin.Context, status string, message string) {
	values := url.Values{}
	values.Set("status", status)
	values.Set("message", message)
	c.Redirect(http.StatusSeeOther, "/master-data?"+values.Encode())
}
