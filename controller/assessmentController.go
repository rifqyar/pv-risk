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
		// BELUM PERNAH DIISI SPEKNYA: Lakukan INSERT (Lengkap dengan kolom baru)
		res, err := tx.Exec(`
			INSERT INTO trx_equipments 
			(equipment_id, tag_number, year_built, shell_material_id, design_pressure, design_pressure_tube, design_temp, design_temp_tube, diameter, diameter_tube, volume,
			diameter_type, diameter_unit, diameter_tube_type, diameter_tube_unit, length, length_unit, volume_unit, temp_design_unit, temp_design_tube_unit,
			pwht, certificate, data_reference, nozzle, nozzle_unit, phase_type, internal_lining, insulation, special_service, protection, cathodic_protection, location) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			masterEqID, payload.Equipment.TagNumber, payload.Equipment.YearBuilt, payload.Equipment.ShellMaterialID,
			payload.Equipment.DesignPressure, payload.Equipment.DesignPressureTube, payload.Equipment.DesignTemp, payload.Equipment.DesignTempTube,
			payload.Equipment.Diameter, payload.Equipment.DiameterTube, payload.Equipment.Volume,
			// KOLOM BARU
			payload.Equipment.DiameterType, payload.Equipment.DiameterUnit, payload.Equipment.DiameterTubeType, payload.Equipment.DiameterTubeUnit,
			payload.Equipment.Length, payload.Equipment.LengthUnit, payload.Equipment.VolumeUnit, payload.Equipment.TempDesignUnit, payload.Equipment.TempDesignTubeUnit,
			payload.Equipment.Pwht, payload.Equipment.Certificate, payload.Equipment.DataReference, payload.Equipment.Nozzle, payload.Equipment.NozzleUnit,
			payload.Equipment.PhaseType, payload.Equipment.InternalLining, payload.Equipment.Insulation, payload.Equipment.SpecialService, payload.Equipment.Protection, payload.Equipment.CathodicProtection,
			payload.Equipment.Location,
		)

		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": "Failed to save equipment specs", "detail": err.Error()})
			return
		}
		lastID, _ := res.LastInsertId()
		trxEqID = int(lastID)

	} else if err == nil {
		// SUDAH PERNAH DIISI: Lakukan UPDATE (Biar spesifikasinya selalu yang paling baru)
		_, err = tx.Exec(`
			UPDATE trx_equipments SET 
			year_built=?, shell_material_id=?, design_pressure=?, design_temp=?, design_pressure_tube=?, design_temp_tube=?, diameter=?, diameter_tube=?, volume=?,
			diameter_type=?, diameter_unit=?, diameter_tube_type=?, diameter_tube_unit=?, length=?, length_unit=?, volume_unit=?, temp_design_unit=?, temp_design_tube_unit=?,
			pwht=?, certificate=?, data_reference=?, nozzle=?, nozzle_unit=?, phase_type=?, internal_lining=?, insulation=?, special_service=?, protection=?, cathodic_protection=?, location=?
			WHERE id=?`,
			payload.Equipment.YearBuilt, payload.Equipment.ShellMaterialID,
			payload.Equipment.DesignPressure, payload.Equipment.DesignTemp,
			payload.Equipment.DesignPressureTube, payload.Equipment.DesignTempTube,
			payload.Equipment.Diameter, payload.Equipment.DiameterTube, payload.Equipment.Volume,
			// KOLOM BARU
			payload.Equipment.DiameterType, payload.Equipment.DiameterUnit, payload.Equipment.DiameterTubeType, payload.Equipment.DiameterTubeUnit,
			payload.Equipment.Length, payload.Equipment.LengthUnit, payload.Equipment.VolumeUnit, payload.Equipment.TempDesignUnit, payload.Equipment.TempDesignTubeUnit,
			payload.Equipment.Pwht, payload.Equipment.Certificate, payload.Equipment.DataReference, payload.Equipment.Nozzle, payload.Equipment.NozzleUnit,
			payload.Equipment.PhaseType, payload.Equipment.InternalLining, payload.Equipment.Insulation, payload.Equipment.SpecialService, payload.Equipment.Protection,
			payload.Equipment.CathodicProtection, payload.Equipment.Location,
			trxEqID,
		)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": "Failed to update equipment specs: " + err.Error()})
			return
		}
	}

	// ==========================================
	// 5. INSERT DETAIL 1: Assessments
	// ==========================================
	// Helper untuk merubah tanggal kosong ("") menjadi NULL di Database
	var prevInsp, actInsp interface{}
	if payload.Assessment.PrevInspectionDate != "" {
		prevInsp = payload.Assessment.PrevInspectionDate
	}
	if payload.Assessment.ActInspectionDate != "" {
		actInsp = payload.Assessment.ActInspectionDate
	}

	res, err := tx.Exec(`
		INSERT INTO assessments 
		(equipment_id, assessment_date, prev_inspection_date, act_inspection_date, operating_pressure, operating_temp, operating_pressure_tube, operating_temp_tube, temp_op_unit, temp_op_tube_unit, phase, h2s_content, co2_content, h2o_content, chloride_index, ph_index, 
		impact_production, insulation_condition, insulation_damage_level, coating_condition, coating_damage_level, corrective_description, corrective_action, corrective_date) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		trxEqID, payload.Assessment.AssessmentDate, prevInsp, actInsp,
		payload.Assessment.OperatingPressure, payload.Assessment.OperatingTemp, payload.Assessment.OperatingPressureTube, payload.Assessment.OperatingTempTube,
		payload.Assessment.TempOpUnit, payload.Assessment.TempOpTubeUnit,
		payload.Environment.Phase, payload.Environment.H2sContent, payload.Environment.Co2Content, payload.Environment.H2oContent,
		payload.Environment.ChlorideIndex, payload.Environment.PhIndex,
		payload.Environment.ImpactProduction, payload.Environment.InsulationCondition, payload.Environment.InsulationDamageLevel,
		payload.Environment.CoatingCondition, payload.Environment.CoatingDamageLevel, payload.Environment.CorrectiveDescription,
		payload.Environment.CorrectiveAction, payload.Environment.CorrectiveDate,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to insert assessment detail: " + err.Error()})
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

	// Kirim respon sukses ke Frontend
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Assessment successfully saved!",
		"data": gin.H{
			"equipment_id":  trxEqID,
			"assessment_id": assessmentID,
		},
	})
}

// Struct penampung data Autofill
type AutofillData struct {
	TagNumber          string  `json:"tag_number"`
	Location           string  `json:"location"`
	YearBuilt          int     `json:"year_built"`
	ShellMaterialID    int     `json:"shell_material_id"`
	DesignPressure     float64 `json:"design_pressure"`
	DesignTemp         float64 `json:"design_temp"`
	DesignPressureTube float64 `json:"design_pressure_tube"`
	DesignTempTube     float64 `json:"design_temp_tube"`
	Diameter           float64 `json:"diameter"`
	DiameterTube       float64 `json:"diameter_tube"`
	Volume             float64 `json:"volume"`
	DiameterType       string  `json:"diameter_type"`
	DiameterUnit       string  `json:"diameter_unit"`
	DiameterTubeType   string  `json:"diameter_tube_type"`
	DiameterTubeUnit   string  `json:"diameter_tube_unit"`
	Length             float64 `json:"length"`
	LengthUnit         string  `json:"length_unit"`
	VolumeUnit         string  `json:"volume_unit"`
	TempDesignUnit     string  `json:"temp_design_unit"`
	TempDesignTubeUnit string  `json:"temp_design_tube_unit"`
	Pwht               string  `json:"pwht"`
	Certificate        string  `json:"certificate"`
	DataReference      string  `json:"data_reference"`
	Nozzle             float64 `json:"nozzle"`
	NozzleUnit         string  `json:"nozzle_unit"`
	PhaseType          string  `json:"phase_type"`
	InternalLining     string  `json:"internal_lining"`
	Insulation         string  `json:"insulation"`
	SpecialService     string  `json:"special_service"`
	Protection         string  `json:"protection"`
	CathodicProtection string  `json:"cathodic_protection"`

	// Data Assessment Terakhir
	OperatingPressure float64 `json:"operating_pressure"`
	OperatingTemp     float64 `json:"operating_temp"`
	TempOpUnit        string  `json:"temp_op_unit"`
	Phase             string  `json:"phase"`
	H2sContent        float64 `json:"h2s_content"`
	Co2Content        float64 `json:"co2_content"`
	ChlorideIndex     int     `json:"chloride_index"`
	PhIndex           int     `json:"ph_index"`
}

func GetEquipmentAutofill(c *gin.Context) {
	eqID := c.Param("id")
	db := config.DB
	var d AutofillData

	// Query ngambil data trx_equipments dan assessments PALING TERAKHIR (ORDER BY a.id DESC)
	query := `
		SELECT 
			COALESCE(t.tag_number, ''), COALESCE(t.location, ''), COALESCE(t.year_built, 0), COALESCE(t.shell_material_id, 0),
			COALESCE(t.design_pressure, 0), COALESCE(t.design_temp, 0), COALESCE(t.design_pressure_tube, 0), COALESCE(t.design_temp_tube, 0),
			COALESCE(t.diameter, 0), COALESCE(t.diameter_tube, 0), COALESCE(t.volume, 0),
			COALESCE(t.diameter_type, 'Inside'), COALESCE(t.diameter_unit, 'inch'), COALESCE(t.diameter_tube_type, 'Inside'), COALESCE(t.diameter_tube_unit, 'inch'),
			COALESCE(CAST(NULLIF(t.length, '-') AS REAL), 0), COALESCE(t.length_unit, 'ft'), COALESCE(t.volume_unit, 'm3'), COALESCE(t.temp_design_unit, 'C'), COALESCE(t.temp_design_tube_unit, 'C'),
			COALESCE(t.pwht, 'No'), COALESCE(t.certificate, '-'), COALESCE(t.data_reference, '-'), COALESCE(CAST(NULLIF(t.nozzle, '-') AS REAL), 0), COALESCE(t.nozzle_unit, 'inch'),
			COALESCE(t.phase_type, 'multi phase'), COALESCE(t.internal_lining, 'None'), COALESCE(t.insulation, 'No'), COALESCE(t.special_service, '-'), COALESCE(t.protection, '-'), COALESCE(t.cathodic_protection, 'No'),

			COALESCE(a.operating_pressure, 0), COALESCE(a.operating_temp, 0), COALESCE(a.temp_op_unit, 'C'), COALESCE(a.phase, ''), 
			COALESCE(a.h2s_content, 0), COALESCE(a.co2_content, 0), COALESCE(a.chloride_index, 0), COALESCE(a.ph_index, 0)
		FROM equipments e
		LEFT JOIN trx_equipments t ON e.id = t.equipment_id
		LEFT JOIN assessments a ON t.id = a.equipment_id
		WHERE e.id = ?
		ORDER BY a.id DESC LIMIT 1
	`

	err := db.QueryRow(query, eqID).Scan(
		&d.TagNumber, &d.Location, &d.YearBuilt, &d.ShellMaterialID, &d.DesignPressure, &d.DesignTemp, &d.DesignPressureTube, &d.DesignTempTube,
		&d.Diameter, &d.DiameterTube, &d.Volume, &d.DiameterType, &d.DiameterUnit, &d.DiameterTubeType, &d.DiameterTubeUnit,
		&d.Length, &d.LengthUnit, &d.VolumeUnit, &d.TempDesignUnit, &d.TempDesignTubeUnit,
		&d.Pwht, &d.Certificate, &d.DataReference, &d.Nozzle, &d.NozzleUnit,
		&d.PhaseType, &d.InternalLining, &d.Insulation, &d.SpecialService, &d.Protection, &d.CathodicProtection,
		&d.OperatingPressure, &d.OperatingTemp, &d.TempOpUnit, &d.Phase, &d.H2sContent, &d.Co2Content, &d.ChlorideIndex, &d.PhIndex,
	)

	if err != nil {
		// Kalau data belum pernah ada, kita kirim status "empty" biar form tetep kosong
		c.JSON(http.StatusOK, gin.H{"status": "empty", "message": "No previous data found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": d})
}
