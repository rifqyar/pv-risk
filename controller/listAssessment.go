package controller

import (
	"encoding/json"
	"net/http"
	"pv-risk/config"
	"pv-risk/models"

	"github.com/gin-gonic/gin"
)

// Struct untuk nampung data baris di List Assessment
type AssessmentListItem struct {
	AssessmentID       int
	TagNumber          string
	Description        string
	AssessmentDate     string
	RiskLevel          string
	NextInspectionYear int
	AssessmentBy       string
}

func ShowListAssessment(c *gin.Context) {
	var listData []AssessmentListItem
	db := config.DB

	// Query JOIN untuk ngambil data dari 3 tabel sekaligus
	// Diurutin dari asesmen yang paling baru dibuat (ORDER BY a.id DESC)
	query := `
		SELECT 
			a.id,
			COALESCE(t.tag_number, 'Unknown') as tag_number,
			COALESCE(e.name, '-') as description,
			COALESCE(a.assessment_date, '-') as assessment_date,
			COALESCE(r.risk_level, 'Pending') as risk_level,
			COALESCE(r.next_inspection_year, 0) as next_inspection_year,
			COALESCE(a.assessment_by, 'Unknown') as assessment_by
		FROM assessments a
		JOIN trx_equipments t ON a.equipment_id = t.id
		JOIN equipments e ON t.equipment_id = e.id
		LEFT JOIN assessment_results r ON r.assessment_id = a.id
		ORDER BY a.id DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item AssessmentListItem
		err := rows.Scan(
			&item.AssessmentID,
			&item.TagNumber,
			&item.Description,
			&item.AssessmentDate,
			&item.RiskLevel,
			&item.NextInspectionYear,
			&item.AssessmentBy,
		)
		if err == nil {
			listData = append(listData, item)
		}
	}

	// Render ke master layout, lempar ActiveMenu "list-assessment"
	c.HTML(http.StatusOK, "list_assessment.html", gin.H{
		"Assessments": listData,
		"ActiveMenu":  "list-assessment",
	})
}

// 1. Update Struct Penampung Data Detail (Biar sinkron sama HTML PDF yang baru)
type AssessmentFullDetail struct {
	AssessmentID      int
	TagNumber         string
	Description       string
	AssessmentDate    string
	AssessmentBy      string
	YearBuilt         int
	Location          string
	DesignPressure    float64
	DesignTemp        float64
	OperatingPressure float64
	OperatingTemp     float64
	Diameter          float64
	Volume            float64
	Phase             string

	// --- GEOMETRY & ENVIRONMENT ---
	DiameterType       string
	DiameterUnit       string
	Length             float64
	LengthUnit         string
	VolumeUnit         string
	TempDesignUnit     string
	TempOpUnit         string
	Pwht               string
	Certificate        string
	DataReference      string
	Nozzle             float64
	NozzleUnit         string
	PhaseType          string
	InternalLining     string
	Insulation         string
	SpecialService     string
	Protection         string
	CathodicProtection string

	// --- FLUIDA ---
	H2sContent    float64
	Co2Content    float64
	ChlorideIndex string
	PhIndex       string

	// --- DAMAGE MECHANISM ---
	Atmospheric string
	Cui         string
	ExtCracking string
	Ssc         string
	Hic         string
	Co2         string
	Ciscc       string
	Galvanic    string
	LofScore    string
	Mic         string
	AmineScc    string

	// --- RESULTS & STRATEGY ---
	LofCategory        int
	CofCategory        string
	RiskLevel          string
	RiskIndex          int
	GoverningComponent string
	MaxIntervalYears   float64
	NextInspectionYear int
	RecommendedMethod  string

	// --- THICKNESS SUMMARY (UI) ---
	CorrosionRate float64
	RemainingLife float64

	// --- THICKNESS DETAIL (PDF SECTION 3) ---
	ShellActThick      float64
	ShellTReq          float64
	ShellCR            float64
	ShellRL            float64
	HeadActThick       float64
	HeadTReq           float64
	HeadCR             float64
	HeadRL             float64
	NozActThick        float64
	NozTReq            float64
	NozCR              float64
	NozRL              float64
	InspectionPlanJSON string
	InspectionPlan     map[string]string
	CladdingJSON       string
	CladdingData       models.CladdingPayload
}

func ViewAssessmentDetail(c *gin.Context) {
	id := c.Param("id")
	db := config.DB
	var d AssessmentFullDetail

	query := `
		SELECT 
			a.id, COALESCE(t.tag_number, '-'), COALESCE(e.name, '-') as description, COALESCE(a.assessment_date, '-'), COALESCE(a.assessment_by, '-'), COALESCE(t.year_built, 0), COALESCE(t.location, '-'),
			COALESCE(t.design_pressure, 0), COALESCE(t.design_temp, 0), COALESCE(a.operating_pressure, 0), COALESCE(a.operating_temp, 0),
			COALESCE(t.diameter, 0), COALESCE(t.volume, 0), COALESCE(p.name, '-') as phase,
			
			COALESCE(t.diameter_type, 'Inside'), COALESCE(t.diameter_unit, 'inch'), 
			COALESCE(CAST(NULLIF(t.length, '-') AS REAL), 0), COALESCE(t.length_unit, 'ft'),
			COALESCE(t.volume_unit, 'm3'), COALESCE(UPPER(t.temp_design_unit), 'C'), COALESCE(UPPER(a.temp_op_unit), 'C'), 
			COALESCE(t.pwht, 'No'), COALESCE(t.certificate, '-'), COALESCE(t.data_reference, '-'), 
			COALESCE(CAST(NULLIF(t.nozzle, '-') AS REAL), 0), COALESCE(t.nozzle_unit, 'inch'),
			
			COALESCE(t.phase_type, 'multi phase'), COALESCE(t.internal_lining, 'None'), COALESCE(t.insulation, 'No'), 
			COALESCE(t.special_service, '-'), COALESCE(t.protection, '-'), COALESCE(t.cathodic_protection, 'No'),

			COALESCE(a.h2s_content, 0), COALESCE(a.co2_content, 0), COALESCE(cc.description, 0) as chloride_index, COALESCE(pc.ph_range, 0) as ph_index,
			
			COALESCE(dm.atmospheric, 'Not'), COALESCE(dm.cui, 'Not'), COALESCE(dm.ext_cracking, 'Not'), COALESCE(dm.ssc, 'Not'), 
			COALESCE(dm.hic, 'Not'), COALESCE(dm.co2, 'Not'), COALESCE(dm.ciscc, 'Not'), COALESCE(dm.galvanic, 'Not'), COALESCE(dm.lof_score, '-'),
			COALESCE(dm.mic, 'Not'), COALESCE(dm.amine_scc, 'Not'),
			
			COALESCE(r.lof_category, 0), COALESCE(r.cof_category, '-'), COALESCE(r.risk_level, 'Pending'), COALESCE(r.risk_index, 0),
			COALESCE(r.governing_component, '-'), COALESCE(r.max_interval_years, 0), COALESCE(r.next_inspection_year, 0), COALESCE(r.recommended_method, '-'),

			COALESCE(th.cr, 0), COALESCE(th.rl, 0),
			COALESCE(th.shell_act, 0), COALESCE(th.shell_treq, 0), COALESCE(th.shell_cr, 0), COALESCE(th.shell_rl, 0),
			COALESCE(th.head_act, 0), COALESCE(th.head_treq, 0), COALESCE(th.head_cr, 0), COALESCE(th.head_rl, 0),
			COALESCE(th.noz_act, 0), COALESCE(th.noz_treq, 0), COALESCE(th.noz_cr, 0), COALESCE(th.noz_rl, 0),

			-- TAMBAHAN BARU UNTUK 10 PLAN JSON
			COALESCE(r.inspection_plan_json, '{}'),
			COALESCE(r.cladding_json, '{}')

		FROM assessments a
		JOIN trx_equipments t ON a.equipment_id = t.id
		JOIN equipments e ON t.equipment_id = e.id
		LEFT JOIN assessment_damage_mechanisms dm ON dm.assessment_id = a.id
		LEFT JOIN assessment_results r ON r.assessment_id = a.id
		LEFT JOIN phase p ON a.phase = p.code
		LEFT JOIN chloride_content cc ON a.chloride_index = cc.level
		LEFT JOIN ph_category pc ON a.ph_index = pc.ph_index
		LEFT JOIN (
			SELECT assessment_id, 
				MAX(corrosion_rate) as cr, MIN(remaining_life) as rl,
				MAX(CASE WHEN component_type = 'shell' THEN act_thick END) as shell_act,
				MAX(CASE WHEN component_type = 'shell' THEN t_req END) as shell_treq,
				MAX(CASE WHEN component_type = 'shell' THEN corrosion_rate END) as shell_cr,
				MAX(CASE WHEN component_type = 'shell' THEN remaining_life END) as shell_rl,
				MAX(CASE WHEN component_type = 'head' THEN act_thick END) as head_act,
				MAX(CASE WHEN component_type = 'head' THEN t_req END) as head_treq,
				MAX(CASE WHEN component_type = 'head' THEN corrosion_rate END) as head_cr,
				MAX(CASE WHEN component_type = 'head' THEN remaining_life END) as head_rl,
				MAX(CASE WHEN component_type = 'nozzle' THEN act_thick END) as noz_act,
				MAX(CASE WHEN component_type = 'nozzle' THEN t_req END) as noz_treq,
				MAX(CASE WHEN component_type = 'nozzle' THEN corrosion_rate END) as noz_cr,
				MAX(CASE WHEN component_type = 'nozzle' THEN remaining_life END) as noz_rl
			FROM assessment_thicknesses
			GROUP BY assessment_id
		) th ON th.assessment_id = a.id
		WHERE a.id = ?
	`

	err := db.QueryRow(query, id).Scan(
		&d.AssessmentID, &d.TagNumber, &d.Description, &d.AssessmentDate, &d.AssessmentBy, &d.YearBuilt, &d.Location,
		&d.DesignPressure, &d.DesignTemp, &d.OperatingPressure, &d.OperatingTemp,
		&d.Diameter, &d.Volume, &d.Phase,
		&d.DiameterType, &d.DiameterUnit, &d.Length, &d.LengthUnit,
		&d.VolumeUnit, &d.TempDesignUnit, &d.TempOpUnit,
		&d.Pwht, &d.Certificate, &d.DataReference, &d.Nozzle, &d.NozzleUnit,
		&d.PhaseType, &d.InternalLining, &d.Insulation,
		&d.SpecialService, &d.Protection, &d.CathodicProtection,
		&d.H2sContent, &d.Co2Content, &d.ChlorideIndex, &d.PhIndex,
		&d.Atmospheric, &d.Cui, &d.ExtCracking, &d.Ssc, &d.Hic, &d.Co2, &d.Ciscc, &d.Galvanic, &d.LofScore,
		&d.Mic, &d.AmineScc,
		&d.LofCategory, &d.CofCategory, &d.RiskLevel, &d.RiskIndex,
		&d.GoverningComponent, &d.MaxIntervalYears, &d.NextInspectionYear, &d.RecommendedMethod,
		&d.CorrosionRate, &d.RemainingLife,
		&d.ShellActThick, &d.ShellTReq, &d.ShellCR, &d.ShellRL,
		&d.HeadActThick, &d.HeadTReq, &d.HeadCR, &d.HeadRL,
		&d.NozActThick, &d.NozTReq, &d.NozCR, &d.NozRL,

		// SCAN JSON-NYA
		&d.InspectionPlanJSON,
		&d.CladdingJSON,
	)

	if err != nil {
		c.String(http.StatusInternalServerError, "❌ ERROR GET DETAIL: "+err.Error())
		return
	}

	// CONVERT JSON STRING JADI MAP BIAR BISA DILOOPING DI HTML
	if d.InspectionPlanJSON != "" {
		json.Unmarshal([]byte(d.InspectionPlanJSON), &d.InspectionPlan)
	}

	if d.CladdingJSON != "" && d.CladdingJSON != "{}" {
		json.Unmarshal([]byte(d.CladdingJSON), &d.CladdingData)
	}

	c.HTML(http.StatusOK, "detail_assessment.html", gin.H{
		"Detail":     d,
		"ActiveMenu": "detail-assessment",
	})
}
