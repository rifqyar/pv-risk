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
		"ActiveMenu":             "assessment", // <--- INI KUNCINYA
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
	// 3. CEK APAKAH MASTER EQUIPMENT VALID
	// ==========================================
	masterEqID := payload.Equipment.MasterEquipmentID
	var checkMaster int
	err = tx.QueryRow("SELECT id FROM equipments WHERE id = ?", masterEqID).Scan(&checkMaster)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"status": "error", "message": "Selected Equipment not found in Master Data!"})
		return
	}

	// ==========================================
	// 4. UPSERT TABEL TRX_EQUIPMENTS (Simpan/Update Spek Teknis)
	// ==========================================
	var trxEqID int
	err = tx.QueryRow("SELECT id FROM trx_equipments WHERE equipment_id = ?", masterEqID).Scan(&trxEqID)

	if err == sql.ErrNoRows {
		// BELUM PERNAH DIISI SPEKNYA: Lakukan INSERT
		res, err := tx.Exec(`
			INSERT INTO trx_equipments 
			(equipment_id, tag_number, year_built, shell_material_id, design_pressure, design_pressure_tube, design_temp, design_temp_tube, diameter, diameter_tube, volume) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			masterEqID, payload.Equipment.TagNumber, payload.Equipment.YearBuilt,
			payload.Equipment.ShellMaterialID, payload.Equipment.DesignPressure, payload.Equipment.DesignPressureTube,
			payload.Equipment.DesignTemp, payload.Equipment.DesignTempTube, payload.Equipment.Diameter, payload.Equipment.DiameterTube, payload.Equipment.Volume,
		)

		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": "Failed to save equipment specs", "detail": err})
			return
		}
		lastID, _ := res.LastInsertId()
		trxEqID = int(lastID)

	} else if err == nil {
		// SUDAH PERNAH DIISI: Lakukan UPDATE (Biar spesifikasinya selalu yang paling baru)
		_, err = tx.Exec(`
			UPDATE trx_equipments SET 
			year_built=?, shell_material_id=?, design_pressure=?, design_temp=?, design_pressure_tube=?, design_temp_tube=?, diameter=?, diameter_tube=?, volume=? 
			WHERE id=?`,
			payload.Equipment.YearBuilt, payload.Equipment.ShellMaterialID,
			payload.Equipment.DesignPressure, payload.Equipment.DesignTemp,
			payload.Equipment.DesignPressureTube, payload.Equipment.DesignTempTube,
			payload.Equipment.Diameter, payload.Equipment.DiameterTube,
			payload.Equipment.Volume, trxEqID,
		)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": "Failed to update equipment specs"})
			return
		}
	}

	// ==========================================
	// 5. INSERT DETAIL 1: Assessments
	// ==========================================
	res, err := tx.Exec(`
		INSERT INTO assessments 
		(equipment_id, assessment_date, prev_inspection_date, act_inspection_date, operating_pressure, operating_temp, operating_pressure_tube, operating_temp_tube, phase, h2s_content, co2_content, h2o_content, chloride_index, ph_index, 
		impact_production, insulation_condition, insulation_damage_level, coating_condition, coating_damage_level, corrective_description, corrective_action, corrective_date) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		trxEqID, payload.Assessment.AssessmentDate, payload.Assessment.PrevInspectionDate, payload.Assessment.ActInspectionDate,
		payload.Assessment.OperatingPressure, payload.Assessment.OperatingTemp, payload.Assessment.OperatingPressureTube, payload.Assessment.OperatingTempTube,
		payload.Environment.Phase, payload.Environment.H2sContent, payload.Environment.Co2Content, payload.Environment.H2oContent,
		payload.Environment.ChlorideIndex, payload.Environment.PhIndex,
		payload.Environment.ImpactProduction, payload.Environment.InsulationCondition, payload.Environment.InsulationDamageLevel,
		payload.Environment.CoatingCondition, payload.Environment.CoatingDamageLevel, payload.Environment.CorrectiveDescription,
		payload.Environment.CorrectiveAction, payload.Environment.CorrectiveDate,
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
		(assessment_id, atmospheric, cui, ext_cracking, co2, mic, ssc, amine_scc, hic, ciscc, galvanic, lof_score) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		assessmentID, payload.DamageMechanisms.Atmospheric, payload.DamageMechanisms.Cui, payload.DamageMechanisms.ExtCracking,
		payload.DamageMechanisms.Co2, payload.DamageMechanisms.Mic, payload.DamageMechanisms.Ssc, payload.DamageMechanisms.AmineScc,
		payload.DamageMechanisms.Hic, payload.DamageMechanisms.Ciscc, payload.DamageMechanisms.Galvanic, payload.DamageMechanisms.LofScore,
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
		(assessment_id, lof_category, cof_financial, cof_safety, cof_category, risk_level, risk_index, insp_internal_thinning, insp_external_corrosion, insp_cracking, governing_component, max_interval_years, next_inspection_year, recommended_method) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		assessmentID, payload.Results.LofCategory, payload.Results.CofFinancial, payload.Results.CofSafety, payload.Results.CofCategory, payload.Results.RiskLevel,
		payload.Results.RiskIndex, payload.Results.InspInternalThinning, payload.Results.InspExternalCorrosion, payload.Results.InspCracking,
		payload.Results.GoverningComponent, payload.Results.MaxIntervalYears,
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
			"equipment_id":  trxEqID,
			"assessment_id": assessmentID,
		},
	})
}
