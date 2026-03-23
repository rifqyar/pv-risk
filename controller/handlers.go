package controller

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"pv-risk/config"
	"pv-risk/models"
	"time"

	"github.com/gin-gonic/gin"
)

func ShowForm(c *gin.Context) {
	// Fetch equipment list from database
	var db = config.DB
	equipments, err := models.GetEquipments(db)
	headType, err := models.GetHeadTypes(db)
	ShellType, err := models.GetShellMaterial(db)
	NeckMaterial, err := models.GetNeckMaterial(db)
	NozzleMaterial, err := models.GetNozzleMaterial(db)
	DamageMechanical, err := models.GetDamageMechanical(db)
	FluidaData, err := models.GetFluida(db)
	PhaseData, err := models.GetPhases(db)
	PollutionData, err := models.GetPollutions(db)
	VelocityData, err := models.GetVelocities(db)
	PHCategoryData, err := models.GetPHCategories(db)
	H2SContentData, err := models.GetH2SContents(db)
	CISCCMatrixData, err := models.GetCISCCMatrix(db)
	CISCCJson, _ := json.Marshal(CISCCMatrixData)
	CO2CorrosionData, err := models.GetCO2CorrosionPreventives(db)
	InhibitorInjectionData, err := models.GetInhibitorInjectionReliability(db)
	ReleaseProductData, err := models.GetReleaseProducts(db)
	CleanupTimeData, err := models.GetCleanupTimes(db)
	ChlorideContentData, err := models.GetChlorideContents(db)

	currentYear := time.Now().Year()
	totalYears := 70

	var years []int

	for i := 0; i < totalYears; i++ {
		years = append(years, currentYear-i)
	}

	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching equipments: %v", err)
		return
	}
	c.HTML(http.StatusOK, "assessment_form.html", gin.H{
		"Equipments":             equipments,
		"HeadTypes":              headType,
		"Years":                  years,
		"ShellType":              ShellType,
		"NeckMaterial":           NeckMaterial,
		"NozzleMaterial":         NozzleMaterial,
		"DamageMechanical":       DamageMechanical,
		"FluidaData":             FluidaData,
		"PhaseData":              PhaseData,
		"PollutionData":          PollutionData,
		"VelocityData":           VelocityData,
		"PHCategoryData":         PHCategoryData,
		"H2SContentData":         H2SContentData,
		"CISCCMasterJSON":        template.JS(CISCCJson),
		"CO2CorrosionData":       CO2CorrosionData,
		"InhibitorInjectionData": InhibitorInjectionData,
		"ReleaseProductData":     ReleaseProductData,
		"CleanupTimeData":        CleanupTimeData,
		"ChlorideContentData":    ChlorideContentData,
	})
}

func SubmitAssessment(c *gin.Context) {
	var payload models.AssessmentPayload

	// ==========================================
	// 1. TANGKAP DAN VALIDASI JSON PAYLOAD
	// ==========================================
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid JSON Payload: " + err.Error(),
		})
		return
	}

	// ==========================================
	// 2. MULAI DATABASE TRANSACTION
	// ==========================================
	tx, err := config.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to start database transaction",
		})
		return
	}
	// PENTING: Jika fungsi return sebelum tx.Commit(), semua query akan dibatalkan (Rollback) otomatis!
	defer tx.Rollback()

	// ==========================================
	// 3. VALIDASI DUPLIKAT EQUIPMENT (Sesuai Excel VBA)
	// ==========================================
	var existingID int
	err = tx.QueryRow("SELECT id FROM equipments WHERE tag_number = ?", payload.Equipment.TagNumber).Scan(&existingID)

	if err == nil {
		// Datanya KETEMU! Berarti Tag Number duplikat (Tolak)
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": "Tag Number / Serial Number '" + payload.Equipment.TagNumber + "' already exists in database!",
		})
		return
	} else if err != sql.ErrNoRows {
		// Error koneksi/sintaks saat ngecek DB
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Database error while checking equipment tag number"})
		return
	}

	// ==========================================
	// 4. INSERT HEADER: Equipments
	// ==========================================
	res, err := tx.Exec(`
		INSERT INTO equipments 
		(tag_number, description, equipment_type_id, year_built, shell_material_id, design_pressure, design_temp, diameter, volume) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		payload.Equipment.TagNumber, payload.Equipment.Description, payload.Equipment.EquipmentTypeID,
		payload.Equipment.YearBuilt, payload.Equipment.ShellMaterialID, payload.Equipment.DesignPressure,
		payload.Equipment.DesignTemp, payload.Equipment.Diameter, payload.Equipment.Volume,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to insert equipment data"})
		return
	}

	// Ambil ID equipment yang baru saja terbuat
	lastEqID, _ := res.LastInsertId()
	equipmentID := int(lastEqID)

	// ==========================================
	// 5. INSERT DETAIL 1: Assessments
	// ==========================================
	res, err = tx.Exec(`
		INSERT INTO assessments 
		(equipment_id, assessment_date, prev_inspection_date, act_inspection_date, operating_pressure, operating_temp, phase, h2s_content, co2_content, h2o_content, chloride_index, ph_index) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		equipmentID, payload.Assessment.AssessmentDate, payload.Assessment.PrevInspectionDate, payload.Assessment.ActInspectionDate,
		payload.Assessment.OperatingPressure, payload.Assessment.OperatingTemp, payload.Environment.Phase,
		payload.Environment.H2sContent, payload.Environment.Co2Content, payload.Environment.H2oContent,
		payload.Environment.ChlorideIndex, payload.Environment.PhIndex,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to insert assessment detail"})
		return
	}

	lastAssID, _ := res.LastInsertId()
	assessmentID := int(lastAssID)

	// ==========================================
	// 6. INSERT DETAIL 2: Thickness Data (Shell & Head)
	// ==========================================
	// 6a. Insert Shell Thickness
	_, err = tx.Exec(`
		INSERT INTO assessment_thicknesses 
		(assessment_id, component_type, prev_thick, act_thick, t_req, corrosion_rate, remaining_life) 
		VALUES (?, 'shell', ?, ?, ?, ?, ?)`,
		assessmentID, payload.ThicknessData.Shell.PrevThick, payload.ThicknessData.Shell.ActThick,
		payload.ThicknessData.Shell.TReq, payload.ThicknessData.Shell.CorrosionRate, payload.ThicknessData.Shell.RemainingLife,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to insert shell thickness data"})
		return
	}

	// 6b. Insert Head Thickness
	_, err = tx.Exec(`
		INSERT INTO assessment_thicknesses 
		(assessment_id, component_type, prev_thick, act_thick, t_req, corrosion_rate, remaining_life) 
		VALUES (?, 'head', ?, ?, ?, ?, ?)`,
		assessmentID, payload.ThicknessData.Head.PrevThick, payload.ThicknessData.Head.ActThick,
		payload.ThicknessData.Head.TReq, payload.ThicknessData.Head.CorrosionRate, payload.ThicknessData.Head.RemainingLife,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to insert head thickness data"})
		return
	}

	// ==========================================
	// 7. INSERT DETAIL 3: Damage Mechanisms
	// ==========================================
	_, err = tx.Exec(`
		INSERT INTO assessment_damage_mechanisms 
		(assessment_id, atmospheric, cui, ext_cracking, co2, mic, ssc, amine_scc, hic, ciscc, total_damage_factor) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		assessmentID, payload.DamageMechanisms.Atmospheric, payload.DamageMechanisms.Cui, payload.DamageMechanisms.ExtCracking,
		payload.DamageMechanisms.Co2, payload.DamageMechanisms.Mic, payload.DamageMechanisms.Ssc, payload.DamageMechanisms.AmineScc,
		payload.DamageMechanisms.Hic, payload.DamageMechanisms.Ciscc, payload.DamageMechanisms.TotalDamageFactor,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to insert damage mechanism data"})
		return
	}

	// ==========================================
	// 8. INSERT DETAIL 4: Final Results & Strategy
	// ==========================================
	_, err = tx.Exec(`
		INSERT INTO assessment_results 
		(assessment_id, lof_category, cof_category, risk_level, risk_index, governing_component, max_interval_years, next_inspection_year, recommended_method) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		assessmentID, payload.Results.LofCategory, payload.Results.CofCategory, payload.Results.RiskLevel,
		payload.Results.RiskIndex, payload.Results.GoverningComponent, payload.Results.MaxIntervalYears,
		payload.Results.NextInspectionYear, payload.Results.RecommendedMethod,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to insert assessment results"})
		return
	}

	// ==========================================
	// 9. SELESAI & COMMIT TRANSACTION
	// ==========================================
	// Jika kode sampai sini tanpa error, kita simpan permanen ke database
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to commit transaction"})
		return
	}

	// Kirim respon sukses ke Frontend (JavaScript lu bakal nangkep ini dan ngeluarin Alert Success)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Assessment successfully saved!",
		"data": gin.H{
			"equipment_id":  equipmentID,
			"assessment_id": assessmentID,
		},
	})
}
