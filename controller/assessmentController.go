package controller

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"pv-risk/config"
	"pv-risk/models"
	"strings"
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

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid JSON Payload: " + err.Error()})
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to start database transaction"})
		return
	}
	defer tx.Rollback()

	masterEqID := payload.Equipment.MasterEquipmentID
	var checkMaster int
	err = tx.QueryRow("SELECT id FROM equipments WHERE id = ?", masterEqID).Scan(&checkMaster)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"status": "error", "message": "Selected Equipment not found!"})
		return
	}

	var trxEqID int
	err = tx.QueryRow("SELECT id FROM trx_equipments WHERE equipment_id = ?", masterEqID).Scan(&trxEqID)

	if err == sql.ErrNoRows {
		res, err := tx.Exec(`
			INSERT INTO trx_equipments (
				equipment_id, tag_number, location, year_built, shell_material_id, 
				design_pressure, design_pressure_tube, design_temp, design_temp_tube, diameter, diameter_tube, volume, 
				diameter_type, diameter_unit, diameter_tube_type, diameter_tube_unit, length, length_unit, volume_unit, 
				temp_design_unit, temp_design_tube_unit, pwht, certificate, data_reference, nozzle, nozzle_unit, phase_type, 
				internal_lining, insulation, special_service, protection, cathodic_protection, head_material_id, type_head, 
				neck_material_id, nozzle_material_id, first_use, 
				serial_number, equip_life, part_type, construction_code, joint_efficiency, joint_efficiency_head, joint_type, 
				radiographic, construction_type, mawp, hydro_test, crown_radius, knuckle_radius, internal_parts_material, 
				shell_contaminant, max_brinell, allowable_stress, inspection_interval, prev_inspection, act_inspection, 
				corrosion_allowance, shell_clad_base_metal, head_clad_base_metal, nozzle_clad_base_metal, shell_wall_thickness, 
				head_wall_thickness, nozzle_wall_thick, shell_thick_cladded, head_thick_cladded, nozzle_thick_cladded, 
				prev_thick_shell, prev_thick_head, nozzle_previous_thick, act_thick_shell, act_thick_head, nozzle_actual_thick
			) VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
			)`,
			masterEqID, payload.Equipment.TagNumber, payload.Equipment.Location, payload.Equipment.YearBuilt, payload.Equipment.ShellMaterialID,
			payload.Equipment.DesignPressure, payload.Equipment.DesignPressureTube, payload.Equipment.DesignTemp, payload.Equipment.DesignTempTube,
			payload.Equipment.Diameter, payload.Equipment.DiameterTube, payload.Equipment.Volume, payload.Equipment.DiameterType, payload.Equipment.DiameterUnit,
			payload.Equipment.DiameterTubeType, payload.Equipment.DiameterTubeUnit, payload.Equipment.Length, payload.Equipment.LengthUnit, payload.Equipment.VolumeUnit,
			payload.Equipment.TempDesignUnit, payload.Equipment.TempDesignTubeUnit, payload.Equipment.Pwht, payload.Equipment.Certificate, payload.Equipment.DataReference,
			payload.Equipment.Nozzle, payload.Equipment.NozzleUnit, payload.Equipment.PhaseType, payload.Equipment.InternalLining, payload.Equipment.Insulation,
			payload.Equipment.SpecialService, payload.Equipment.Protection, payload.Equipment.CathodicProtection, payload.Equipment.HeadMaterialID, payload.Equipment.TypeHead,
			payload.Equipment.NeckMaterialID, payload.Equipment.NozzleMaterialID, payload.Equipment.FirstUse,
			// New Variables
			payload.Equipment.SerialNumber, payload.Equipment.EquipLife, payload.Equipment.PartType, payload.Equipment.ConstructionCode, payload.Equipment.JointEfficiency, payload.Equipment.JointEfficiencyHead, payload.Equipment.JointType,
			payload.Equipment.Radiographic, payload.Equipment.ConstructionType, payload.Equipment.Mawp, payload.Equipment.HydroTest, payload.Equipment.CrownRadius, payload.Equipment.KnuckleRadius, payload.Equipment.InternalPartsMaterial,
			payload.Equipment.ShellContaminant, payload.Equipment.MaxBrinell, payload.Equipment.AllowableStress, payload.Equipment.InspectionInterval, payload.Equipment.PrevInspection, payload.Equipment.ActInspection,
			payload.Equipment.CorrosionAllowance, payload.Equipment.ShellCladBaseMetal, payload.Equipment.HeadCladBaseMetal, payload.Equipment.NozzleCladBaseMetal, payload.Equipment.ShellWallThickness,
			payload.Equipment.HeadWallThickness, payload.Equipment.NozzleWallThick, payload.Equipment.ShellThickCladded, payload.Equipment.HeadThickCladded, payload.Equipment.NozzleThickCladded,
			payload.Equipment.PrevThickShell, payload.Equipment.PrevThickHead, payload.Equipment.NozzlePreviousThick, payload.Equipment.ActThickShell, payload.Equipment.ActThickHead, payload.Equipment.NozzleActualThick,
		)

		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed: trx_equipments.tag_number") {
				c.JSON(http.StatusConflict, gin.H{"status": "error", "message": "Tag Number ini sudah pernah didaftarkan."})
				return
			}
			log.Printf("[ERROR INSERT EQUIPMENT]: %v", err)
			c.JSON(500, gin.H{"status": "error", "message": "Kendala saat menyimpan data."})
			return
		}
		lastID, _ := res.LastInsertId()
		trxEqID = int(lastID)

	} else {
		_, err = tx.Exec(`
			UPDATE trx_equipments SET 
				location=?, year_built=?, shell_material_id=?, design_pressure=?, design_temp=?, design_pressure_tube=?, design_temp_tube=?, diameter=?, diameter_tube=?, volume=?,
				diameter_type=?, diameter_unit=?, diameter_tube_type=?, diameter_tube_unit=?, length=?, length_unit=?, volume_unit=?, temp_design_unit=?, temp_design_tube_unit=?,
				pwht=?, certificate=?, data_reference=?, nozzle=?, nozzle_unit=?, phase_type=?, internal_lining=?, insulation=?, special_service=?, protection=?, cathodic_protection=?, head_material_id=?, type_head=?, neck_material_id=?, nozzle_material_id=?, first_use=?,
				serial_number=?, equip_life=?, part_type=?, construction_code=?, joint_efficiency=?, joint_efficiency_head=?, joint_type=?, radiographic=?, construction_type=?, mawp=?, hydro_test=?, crown_radius=?, knuckle_radius=?, internal_parts_material=?,
				shell_contaminant=?, max_brinell=?, allowable_stress=?, inspection_interval=?, prev_inspection=?, act_inspection=?, corrosion_allowance=?, shell_clad_base_metal=?, head_clad_base_metal=?, nozzle_clad_base_metal=?, shell_wall_thickness=?,
				head_wall_thickness=?, nozzle_wall_thick=?, shell_thick_cladded=?, head_thick_cladded=?, nozzle_thick_cladded=?, prev_thick_shell=?, prev_thick_head=?, nozzle_previous_thick=?, act_thick_shell=?, act_thick_head=?, nozzle_actual_thick=?
			WHERE id=?`,
			payload.Equipment.Location, payload.Equipment.YearBuilt, payload.Equipment.ShellMaterialID, payload.Equipment.DesignPressure, payload.Equipment.DesignTemp, payload.Equipment.DesignPressureTube, payload.Equipment.DesignTempTube, payload.Equipment.Diameter, payload.Equipment.DiameterTube, payload.Equipment.Volume,
			payload.Equipment.DiameterType, payload.Equipment.DiameterUnit, payload.Equipment.DiameterTubeType, payload.Equipment.DiameterTubeUnit, payload.Equipment.Length, payload.Equipment.LengthUnit, payload.Equipment.VolumeUnit, payload.Equipment.TempDesignUnit, payload.Equipment.TempDesignTubeUnit,
			payload.Equipment.Pwht, payload.Equipment.Certificate, payload.Equipment.DataReference, payload.Equipment.Nozzle, payload.Equipment.NozzleUnit, payload.Equipment.PhaseType, payload.Equipment.InternalLining, payload.Equipment.Insulation, payload.Equipment.SpecialService, payload.Equipment.Protection, payload.Equipment.CathodicProtection, payload.Equipment.HeadMaterialID, payload.Equipment.TypeHead, payload.Equipment.NeckMaterialID, payload.Equipment.NozzleMaterialID, payload.Equipment.FirstUse,
			// New Variables
			payload.Equipment.SerialNumber, payload.Equipment.EquipLife, payload.Equipment.PartType, payload.Equipment.ConstructionCode, payload.Equipment.JointEfficiency, payload.Equipment.JointEfficiencyHead, payload.Equipment.JointType, payload.Equipment.Radiographic, payload.Equipment.ConstructionType, payload.Equipment.Mawp, payload.Equipment.HydroTest, payload.Equipment.CrownRadius, payload.Equipment.KnuckleRadius, payload.Equipment.InternalPartsMaterial,
			payload.Equipment.ShellContaminant, payload.Equipment.MaxBrinell, payload.Equipment.AllowableStress, payload.Equipment.InspectionInterval, payload.Equipment.PrevInspection, payload.Equipment.ActInspection, payload.Equipment.CorrosionAllowance, payload.Equipment.ShellCladBaseMetal, payload.Equipment.HeadCladBaseMetal, payload.Equipment.NozzleCladBaseMetal, payload.Equipment.ShellWallThickness,
			payload.Equipment.HeadWallThickness, payload.Equipment.NozzleWallThick, payload.Equipment.ShellThickCladded, payload.Equipment.HeadThickCladded, payload.Equipment.NozzleThickCladded, payload.Equipment.PrevThickShell, payload.Equipment.PrevThickHead, payload.Equipment.NozzlePreviousThick, payload.Equipment.ActThickShell, payload.Equipment.ActThickHead, payload.Equipment.NozzleActualThick,
			trxEqID,
		)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": "Failed to update specs: " + err.Error()})
			return
		}
	}

	var prevInsp, actInsp, corrDate interface{}
	if payload.Assessment.PrevInspectionDate != "" {
		prevInsp = payload.Assessment.PrevInspectionDate
	}
	if payload.Assessment.ActInspectionDate != "" {
		actInsp = payload.Assessment.ActInspectionDate
	}
	if payload.Environment.CorrectiveDate != nil && *payload.Environment.CorrectiveDate != "" {
		corrDate = *payload.Environment.CorrectiveDate
	}

	res, err := tx.Exec(`
		INSERT INTO assessments (
			equipment_id, assessment_date, prev_inspection_date, act_inspection_date, operating_pressure, operating_temp, operating_pressure_tube, operating_temp_tube, 
			temp_op_unit, temp_op_tube_unit, phase, h2s_content, co2_content, h2o_content, chloride_index, ph_index, 
			impact_production, insulation_condition, insulation_damage_level, coating_condition, coating_damage_level, corrective_description, corrective_action, corrective_date,
			contaminant_amine, flow_velocity, preventive_corrosion, inhibitor_effectivity, env_ext_cracking, vibration,
			impact_for_production, comp_nitrogen, comp_methane, comp_ethane, comp_propane, comp_butane, comp_solvent, comp_air, h2s_ppm,
			fluida, pollutant, cp_condition, corrosion_monitoring, biocide_treatment, release_fluid_containment, clean_up_time, heat_traced, steam_out,
			prev_ext_corrosion, conf_ext_corrosion, prev_int_cracking, conf_int_cracking, prev_int_thinning, conf_int_thinning, prev_loc_int_corrosion, conf_loc_int_corrosion
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		trxEqID, payload.Assessment.AssessmentDate, prevInsp, actInsp, payload.Assessment.OperatingPressure, payload.Assessment.OperatingTemp, payload.Assessment.OperatingPressureTube, payload.Assessment.OperatingTempTube,
		payload.Assessment.TempOpUnit, payload.Assessment.TempOpTubeUnit, payload.Environment.Phase, payload.Environment.H2SContent, payload.Environment.CO2Content, payload.Environment.H2OContent, payload.Environment.ChlorideIndex, payload.Environment.PHIndex,
		payload.Environment.ImpactProduction, payload.Environment.InsulationCondition, payload.Environment.InsulationDamageLevel, payload.Environment.CoatingCondition, payload.Environment.CoatingDamageLevel, payload.Environment.CorrectiveDescription, payload.Environment.CorrectiveAction, corrDate,
		payload.Environment.ContaminantAmine, payload.Environment.FlowVelocity, payload.Environment.PreventiveCorrosion, payload.Environment.InhibitorEffectivity, payload.Environment.EnvExtCracking, payload.Environment.Vibration,
		payload.Environment.ImpactForProduction, payload.Environment.CompNitrogen, payload.Environment.CompMethane, payload.Environment.CompEthane, payload.Environment.CompPropane, payload.Environment.CompButane, payload.Environment.CompSolvent, payload.Environment.CompAir, payload.Environment.H2SPpm,
		payload.Environment.Fluida, payload.Environment.Pollutant, payload.Environment.CpCondition, payload.Environment.CorrosionMonitoring, payload.Environment.BiocideTreatment, payload.Environment.ReleaseFluidContainment, payload.Environment.CleanUpTime, payload.Environment.HeatTraced, payload.Environment.SteamOut,
		payload.Environment.PrevExtCorrosion, payload.Environment.ConfExtCorrosion, payload.Environment.PrevIntCracking, payload.Environment.ConfIntCracking, payload.Environment.PrevIntThinning, payload.Environment.ConfIntThinning, payload.Environment.PrevLocIntCorrosion, payload.Environment.ConfLocIntCorrosion,
	)

	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to insert assessment: " + err.Error()})
		return
	}
	lastAssID, _ := res.LastInsertId()
	assessmentID := int(lastAssID)

	components := []struct {
		name string
		data models.ComponentThickness
	}{
		{"shell", payload.ThicknessData.Shell},
		{"head", payload.ThicknessData.Head},
		{"nozzle", payload.ThicknessData.Nozzle},
	}

	for _, comp := range components {
		_, err = tx.Exec(`INSERT INTO assessment_thicknesses (assessment_id, component_type, prev_thick, act_thick, t_req, corrosion_rate, remaining_life) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			assessmentID, comp.name, comp.data.PrevThick, comp.data.ActThick, comp.data.TReq, comp.data.CorrosionRate, comp.data.RemainingLife)
	}

	_, err = tx.Exec(`INSERT INTO assessment_damage_mechanisms (assessment_id, atmospheric, cui, ext_cracking, co2, mic, ssc, amine_scc, hic, ciscc, galvanic, lof_score) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		assessmentID, payload.DamageMechanisms.Atmospheric, payload.DamageMechanisms.Cui, payload.DamageMechanisms.ExtCracking, payload.DamageMechanisms.Co2, payload.DamageMechanisms.Mic, payload.DamageMechanisms.Ssc, payload.DamageMechanisms.AmineScc, payload.DamageMechanisms.Hic, payload.DamageMechanisms.Ciscc, payload.DamageMechanisms.Galvanic, payload.DamageMechanisms.LofScore)

	inspectionBytes, _ := json.Marshal(payload.Results.InspectionData)
	cladBytes, _ := json.Marshal(payload.CladdingData)

	_, err = tx.Exec(`INSERT INTO assessment_results (assessment_id, lof_category, cof_financial, cof_safety, cof_category, risk_level, risk_index, inspection_plan_json, governing_component, max_interval_years, next_inspection_year, recommended_method, cladding_json) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		assessmentID, payload.Results.LofCategory, payload.Results.CofFinancial, payload.Results.CofSafety, payload.Results.CofCategory, payload.Results.RiskLevel, payload.Results.RiskIndex, string(inspectionBytes), payload.Results.GoverningComponent, payload.Results.MaxIntervalYears, payload.Results.NextInspectionYear, payload.Results.RecommendedMethod, string(cladBytes))

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Assessment successfully saved!"})
}

// Struct Autofill Data
type AutofillData struct {
	TagNumber          string  `json:"tag_number"`
	Location           string  `json:"location"`
	YearBuilt          int     `json:"year_built"`
	ShellMaterialID    int     `json:"shell_material_id"`
	HeadMaterialID     int     `json:"head_material_id"`
	TypeHead           int     `json:"type_head"`
	NeckMaterialID     int     `json:"neck_material_id"`
	NozzleMaterialID   int     `json:"nozzle_material_id"`
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
	FirstUse           int     `json:"first_use"`

	SerialNumber          string  `json:"serial_number"`
	EquipLife             int     `json:"equip_life"`
	PartType              string  `json:"part_type"`
	ConstructionCode      string  `json:"construction_code"`
	JointEfficiency       float64 `json:"joint_efficiency"`
	JointEfficiencyHead   float64 `json:"joint_efficiency_head"`
	JointType             string  `json:"joint_type"`
	Radiographic          string  `json:"radiographic"`
	ConstructionType      string  `json:"construction_type"`
	Mawp                  float64 `json:"mawp"`
	HydroTest             float64 `json:"hydro_test"`
	CrownRadius           float64 `json:"crown_radius"`
	KnuckleRadius         float64 `json:"knuckle_radius"`
	InternalPartsMaterial string  `json:"internal_parts_material"`
	ShellContaminant      string  `json:"shell_contaminant"`
	MaxBrinell            string  `json:"max_brinell"`
	AllowableStress       float64 `json:"allowable_stress"`

	InspectionInterval  int     `json:"inspection_interval"`
	PrevInspection      string  `json:"prev_inspection"`
	ActInspection       string  `json:"act_inspection"`
	CorrosionAllowance  float64 `json:"corrosion_allowance"`
	ShellCladBaseMetal  float64 `json:"shell_clad_base_metal"`
	HeadCladBaseMetal   float64 `json:"head_clad_base_metal"`
	NozzleCladBaseMetal float64 `json:"nozzle_clad_base_metal"`
	ShellWallThickness  float64 `json:"shell_wall_thickness"`
	HeadWallThickness   float64 `json:"head_wall_thickness"`
	NozzleWallThick     float64 `json:"nozzle_wall_thick"`
	ShellThickCladded   float64 `json:"shell_thick_cladded"`
	HeadThickCladded    float64 `json:"head_thick_cladded"`
	NozzleThickCladded  float64 `json:"nozzle_thick_cladded"`
	PrevThickShell      float64 `json:"prev_thick_shell"`
	PrevThickHead       float64 `json:"prev_thick_head"`
	NozzlePreviousThick float64 `json:"nozzle_previous_thick"`
	ActThickShell       float64 `json:"act_thick_shell"`
	ActThickHead        float64 `json:"act_thick_head"`
	NozzleActualThick   float64 `json:"nozzle_actual_thick"`

	OperatingPressure float64 `json:"operating_pressure"`
	OperatingTemp     float64 `json:"operating_temp"`
	TempOpUnit        string  `json:"temp_op_unit"`
	Phase             string  `json:"phase"`
	H2sContent        float64 `json:"h2s_content"`
	H2oContent        float64 `json:"h2o_content"`
	Co2Content        float64 `json:"co2_content"`
	ChlorideIndex     int     `json:"chloride_index"`
	PhIndex           int     `json:"ph_index"`

	// Environment Step 3
	ContaminantAmine        string  `json:"contaminant_amine"`
	FlowVelocity            string  `json:"flow_velocity"`
	PreventiveCorrosion     string  `json:"preventive_corrosion"`
	InhibitorEffectivity    string  `json:"inhibitor_effectivity"`
	EnvExtCracking          string  `json:"env_ext_cracking"`
	Vibration               string  `json:"vibration"`
	ImpactForProduction     string  `json:"impact_for_production"`
	CompNitrogen            float64 `json:"comp_nitrogen"`
	CompMethane             float64 `json:"comp_methane"`
	CompEthane              float64 `json:"comp_ethane"`
	CompPropane             float64 `json:"comp_propane"`
	CompButane              float64 `json:"comp_butane"`
	CompSolvent             float64 `json:"comp_solvent"`
	CompAir                 float64 `json:"comp_air"`
	H2SPpm                  int     `json:"h2s_ppm"`
	Fluida                  string  `json:"fluida"`
	Pollutant               string  `json:"pollutant"`
	CpCondition             string  `json:"cp_condition"`
	CorrosionMonitoring     string  `json:"corrosion_monitoring"`
	BiocideTreatment        string  `json:"biocide_treatment"`
	ReleaseFluidContainment string  `json:"release_fluid_containment"`
	CleanUpTime             string  `json:"clean_up_time"`
	HeatTraced              int     `json:"heat_traced"`
	SteamOut                int     `json:"steam_out"`
	PrevExtCorrosion        string  `json:"prev_ext_corrosion"`
	ConfExtCorrosion        string  `json:"conf_ext_corrosion"`
	PrevIntCracking         string  `json:"prev_int_cracking"`
	ConfIntCracking         string  `json:"conf_int_cracking"`
	PrevIntThinning         string  `json:"prev_int_thinning"`
	ConfIntThinning         string  `json:"conf_int_thinning"`
	PrevLocIntCorrosion     string  `json:"prev_loc_int_corrosion"`
	ConfLocIntCorrosion     string  `json:"conf_loc_int_corrosion"`
}

func GetEquipmentAutofill(c *gin.Context) {
	eqID := c.Param("id")
	db := config.DB
	var d AutofillData

	query := `
		SELECT 
			COALESCE(t.tag_number, ''), COALESCE(t.location, ''), COALESCE(t.year_built, 0), COALESCE(t.shell_material_id, 0), COALESCE(t.head_material_id, 0), COALESCE(t.type_head, 0), COALESCE(t.neck_material_id, 0), COALESCE(t.nozzle_material_id, 0),
			COALESCE(t.design_pressure, 0), COALESCE(t.design_temp, 0), COALESCE(t.design_pressure_tube, 0), COALESCE(t.design_temp_tube, 0), COALESCE(t.diameter, 0), COALESCE(t.diameter_tube, 0), COALESCE(t.volume, 0),
			COALESCE(t.diameter_type, 'inside'), COALESCE(t.diameter_unit, 'inch'), COALESCE(t.diameter_tube_type, 'inside'), COALESCE(t.diameter_tube_unit, 'inch'), COALESCE(CAST(NULLIF(t.length, '-') AS REAL), 0), COALESCE(t.length_unit, 'ft'), COALESCE(t.volume_unit, 'm3'), COALESCE(t.temp_design_unit, 'C'), COALESCE(t.temp_design_tube_unit, 'C'),
			COALESCE(t.pwht, 'No'), COALESCE(t.certificate, '-'), COALESCE(t.data_reference, '-'), COALESCE(CAST(NULLIF(t.nozzle, '-') AS REAL), 0), COALESCE(t.nozzle_unit, 'inch'), COALESCE(t.phase_type, 'multi phase'), COALESCE(t.internal_lining, 'None'), COALESCE(t.insulation, 'No'), COALESCE(t.special_service, '-'), COALESCE(t.protection, '-'), COALESCE(t.cathodic_protection, 'No'), COALESCE(t.first_use, 0),
			COALESCE(t.serial_number, ''), COALESCE(t.equip_life, 0), COALESCE(t.part_type, ''), COALESCE(t.construction_code, ''), COALESCE(t.joint_efficiency, 0), COALESCE(t.joint_efficiency_head, 0), COALESCE(t.joint_type, ''), COALESCE(t.radiographic, ''), COALESCE(t.construction_type, ''), COALESCE(t.mawp, 0), COALESCE(t.hydro_test, 0), COALESCE(t.crown_radius, 0), COALESCE(t.knuckle_radius, 0), COALESCE(t.internal_parts_material, ''),
			COALESCE(t.shell_contaminant, ''), COALESCE(t.max_brinell, ''), COALESCE(t.allowable_stress, 0), COALESCE(t.inspection_interval, 0), COALESCE(t.prev_inspection, ''), COALESCE(t.act_inspection, ''), COALESCE(t.corrosion_allowance, 0), COALESCE(t.shell_clad_base_metal, 0), COALESCE(t.head_clad_base_metal, 0), COALESCE(t.nozzle_clad_base_metal, 0), COALESCE(t.shell_wall_thickness, 0),
			COALESCE(t.head_wall_thickness, 0), COALESCE(t.nozzle_wall_thick, 0), COALESCE(t.shell_thick_cladded, 0), COALESCE(t.head_thick_cladded, 0), COALESCE(t.nozzle_thick_cladded, 0), COALESCE(t.prev_thick_shell, 0), COALESCE(t.prev_thick_head, 0), COALESCE(t.nozzle_previous_thick, 0), COALESCE(t.act_thick_shell, 0), COALESCE(t.act_thick_head, 0), COALESCE(t.nozzle_actual_thick, 0),
			
			COALESCE(a.operating_pressure, 0), COALESCE(a.operating_temp, 0), COALESCE(a.temp_op_unit, 'C'), COALESCE(a.phase, ''), COALESCE(a.h2s_content, 0), COALESCE(a.h2o_content, 0), COALESCE(a.co2_content, 0), COALESCE(a.chloride_index, 0), COALESCE(a.ph_index, 0),
			COALESCE(a.contaminant_amine, ''), COALESCE(a.flow_velocity, ''), COALESCE(a.preventive_corrosion, ''), COALESCE(a.inhibitor_effectivity, ''), COALESCE(a.env_ext_cracking, ''), COALESCE(a.vibration, ''),
			COALESCE(a.impact_for_production, ''), COALESCE(a.comp_nitrogen, 0), COALESCE(a.comp_methane, 0), COALESCE(a.comp_ethane, 0), COALESCE(a.comp_propane, 0), COALESCE(a.comp_butane, 0), COALESCE(a.comp_solvent, 0), COALESCE(a.comp_air, 0),
			COALESCE(a.fluida, ''), COALESCE(a.pollutant, ''), COALESCE(a.cp_condition, ''), COALESCE(a.corrosion_monitoring, ''), COALESCE(a.biocide_treatment, ''), COALESCE(a.release_fluid_containment, ''), COALESCE(a.clean_up_time, ''), COALESCE(a.heat_traced, 0), COALESCE(a.steam_out, 0),
			COALESCE(a.prev_ext_corrosion, ''), COALESCE(a.conf_ext_corrosion, ''), COALESCE(a.prev_int_cracking, ''), COALESCE(a.conf_int_cracking, ''), COALESCE(a.prev_int_thinning, ''), COALESCE(a.conf_int_thinning, ''), COALESCE(a.prev_loc_int_corrosion, ''), COALESCE(a.conf_loc_int_corrosion, '')
		FROM equipments e
		LEFT JOIN trx_equipments t ON e.id = t.equipment_id
		LEFT JOIN assessments a ON t.id = a.equipment_id
		WHERE e.id = ?
		ORDER BY a.id DESC LIMIT 1
	`

	err := db.QueryRow(query, eqID).Scan(
		&d.TagNumber, &d.Location, &d.YearBuilt, &d.ShellMaterialID, &d.HeadMaterialID, &d.TypeHead, &d.NeckMaterialID, &d.NozzleMaterialID,
		&d.DesignPressure, &d.DesignTemp, &d.DesignPressureTube, &d.DesignTempTube, &d.Diameter, &d.DiameterTube, &d.Volume,
		&d.DiameterType, &d.DiameterUnit, &d.DiameterTubeType, &d.DiameterTubeUnit, &d.Length, &d.LengthUnit, &d.VolumeUnit, &d.TempDesignUnit, &d.TempDesignTubeUnit,
		&d.Pwht, &d.Certificate, &d.DataReference, &d.Nozzle, &d.NozzleUnit, &d.PhaseType, &d.InternalLining, &d.Insulation, &d.SpecialService, &d.Protection, &d.CathodicProtection, &d.FirstUse,
		&d.SerialNumber, &d.EquipLife, &d.PartType, &d.ConstructionCode, &d.JointEfficiency, &d.JointEfficiencyHead, &d.JointType, &d.Radiographic, &d.ConstructionType, &d.Mawp, &d.HydroTest, &d.CrownRadius, &d.KnuckleRadius, &d.InternalPartsMaterial,
		&d.ShellContaminant, &d.MaxBrinell, &d.AllowableStress, &d.InspectionInterval, &d.PrevInspection, &d.ActInspection, &d.CorrosionAllowance, &d.ShellCladBaseMetal, &d.HeadCladBaseMetal, &d.NozzleCladBaseMetal, &d.ShellWallThickness,
		&d.HeadWallThickness, &d.NozzleWallThick, &d.ShellThickCladded, &d.HeadThickCladded, &d.NozzleThickCladded, &d.PrevThickShell, &d.PrevThickHead, &d.NozzlePreviousThick, &d.ActThickShell, &d.ActThickHead, &d.NozzleActualThick,

		&d.OperatingPressure, &d.OperatingTemp, &d.TempOpUnit, &d.Phase, &d.H2sContent, &d.H2oContent, &d.Co2Content, &d.ChlorideIndex, &d.PhIndex,
		&d.ContaminantAmine, &d.FlowVelocity, &d.PreventiveCorrosion, &d.InhibitorEffectivity, &d.EnvExtCracking, &d.Vibration,
		&d.ImpactForProduction, &d.CompNitrogen, &d.CompMethane, &d.CompEthane, &d.CompPropane, &d.CompButane, &d.CompSolvent, &d.CompAir,
		&d.Fluida, &d.Pollutant, &d.CpCondition, &d.CorrosionMonitoring, &d.BiocideTreatment, &d.ReleaseFluidContainment, &d.CleanUpTime, &d.HeatTraced, &d.SteamOut,
		&d.PrevExtCorrosion, &d.ConfExtCorrosion, &d.PrevIntCracking, &d.ConfIntCracking, &d.PrevIntThinning, &d.ConfIntThinning, &d.PrevLocIntCorrosion, &d.ConfLocIntCorrosion,
	)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "empty", "message": "No previous data found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": d})
}
