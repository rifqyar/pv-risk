package controller

import (
	"net/http"
	"pv-risk/config"

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
			COALESCE(r.next_inspection_year, 0) as next_inspection_year
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
	YearBuilt         int
	DesignPressure    float64
	DesignTemp        float64
	OperatingPressure float64
	OperatingTemp     float64
	Diameter          float64
	Volume            float64
	Phase             string

	// --- KOLOM BARU: GEOMETRY, UNIT, & DOCS ---
	DiameterType   string
	DiameterUnit   string
	Length         float64
	LengthUnit     string
	VolumeUnit     string
	TempDesignUnit string
	TempOpUnit     string
	Pwht           string
	Certificate    string
	DataReference  string
	Nozzle         float64
	NozzleUnit     string

	// --- KOLOM BARU: ENVIRONMENT & PROTECTION ---
	PhaseType          string
	InternalLining     string
	Insulation         string
	SpecialService     string
	Protection         string
	CathodicProtection string

	// --- FLUIDA ---
	H2sContent    float64
	Co2Content    float64
	ChlorideIndex int
	PhIndex       int

	// --- DAMAGE MECHANISM (Termasuk CISCC) ---
	Atmospheric string
	Cui         string
	ExtCracking string
	Ssc         string
	Hic         string
	Co2         string
	Ciscc       string
	LofScore    string

	// --- RESULTS & STRATEGY ---
	LofCategory        int
	CofCategory        string
	RiskLevel          string
	RiskIndex          int
	GoverningComponent string
	MaxIntervalYears   float64
	NextInspectionYear int
	RecommendedMethod  string
}

// 2. Update Fungsi View-nya
func ViewAssessmentDetail(c *gin.Context) {
	id := c.Param("id")
	db := config.DB
	var d AssessmentFullDetail

	// Query JOIN 5 Tabel
	query := `
		SELECT 
			a.id, COALESCE(t.tag_number, '-'), COALESCE(e.name, '-') as description, COALESCE(a.assessment_date, '-'), COALESCE(t.year_built, 0),
			COALESCE(t.design_pressure, 0), COALESCE(t.design_temp, 0), COALESCE(a.operating_pressure, 0), COALESCE(a.operating_temp, 0),
			COALESCE(t.diameter, 0), COALESCE(t.volume, 0), COALESCE(a.phase, '-'),
			
			-- KOLOM BARU: GEOMETRY & DOCS (Dari tabel trx_equipments & assessments)
			COALESCE(t.diameter_type, 'Inside'), COALESCE(t.diameter_unit, 'inch'), COALESCE(t.length, 0), COALESCE(t.length_unit, 'ft'),
			COALESCE(t.volume_unit, 'm3'), COALESCE(t.temp_design_unit, 'C'), COALESCE(a.temp_op_unit, 'C'), 
			COALESCE(t.pwht, 'No'), COALESCE(t.certificate, '-'), COALESCE(t.data_reference, '-'), COALESCE(t.nozzle, 0), COALESCE(t.nozzle_unit, 'inch'),
			
			-- KOLOM BARU: ENVIRONMENT & PROTECTION (Dari tabel trx_equipments)
			COALESCE(t.phase_type, 'multi phase'), COALESCE(t.internal_lining, 'None'), COALESCE(t.insulation, 'No'), 
			COALESCE(t.special_service, '-'), COALESCE(t.protection, '-'), COALESCE(t.cathodic_protection, 'No'),

			-- FLUIDA
			COALESCE(a.h2s_content, 0), COALESCE(a.co2_content, 0), COALESCE(a.chloride_index, 0), COALESCE(a.ph_index, 0),
			
			-- DAMAGE MECHANISMS
			COALESCE(dm.atmospheric, 'Not'), COALESCE(dm.cui, 'Not'), COALESCE(dm.ext_cracking, 'Not'), COALESCE(dm.ssc, 'Not'), 
			COALESCE(dm.hic, 'Not'), COALESCE(dm.co2, 'Not'), COALESCE(dm.ciscc, 'Not'), COALESCE(dm.lof_score, '-'),
			
			-- RESULTS
			COALESCE(r.lof_category, 0), COALESCE(r.cof_category, '-'), COALESCE(r.risk_level, 'Pending'), COALESCE(r.risk_index, 0),
			COALESCE(r.governing_component, '-'), COALESCE(r.max_interval_years, 0), COALESCE(r.next_inspection_year, 0), COALESCE(r.recommended_method, '-')
		FROM assessments a
		JOIN trx_equipments t ON a.equipment_id = t.id
		JOIN equipments e ON t.equipment_id = e.id
		LEFT JOIN assessment_damage_mechanisms dm ON dm.assessment_id = a.id
		LEFT JOIN assessment_results r ON r.assessment_id = a.id
		WHERE a.id = ?
	`

	// Scanning harus urut persis sama SELECT di atas
	err := db.QueryRow(query, id).Scan(
		&d.AssessmentID, &d.TagNumber, &d.Description, &d.AssessmentDate, &d.YearBuilt,
		&d.DesignPressure, &d.DesignTemp, &d.OperatingPressure, &d.OperatingTemp,
		&d.Diameter, &d.Volume, &d.Phase,

		// SCAN KOLOM BARU GEOMETRY
		&d.DiameterType, &d.DiameterUnit, &d.Length, &d.LengthUnit,
		&d.VolumeUnit, &d.TempDesignUnit, &d.TempOpUnit,
		&d.Pwht, &d.Certificate, &d.DataReference, &d.Nozzle, &d.NozzleUnit,

		// SCAN KOLOM BARU ENVIRONMENT
		&d.PhaseType, &d.InternalLining, &d.Insulation,
		&d.SpecialService, &d.Protection, &d.CathodicProtection,

		// FLUIDA
		&d.H2sContent, &d.Co2Content, &d.ChlorideIndex, &d.PhIndex,

		// DM
		&d.Atmospheric, &d.Cui, &d.ExtCracking, &d.Ssc, &d.Hic, &d.Co2, &d.Ciscc, &d.LofScore,

		// RESULTS
		&d.LofCategory, &d.CofCategory, &d.RiskLevel, &d.RiskIndex,
		&d.GoverningComponent, &d.MaxIntervalYears, &d.NextInspectionYear, &d.RecommendedMethod,
	)

	if err != nil {
		// Ini bakal nampilin error aslinya langsung di layar browser lu gede-gede!
		c.String(http.StatusInternalServerError, "❌ ERROR GET DETAIL: "+err.Error())
		return
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found or incomplete data"})
		return
	}

	// Render file HTML-nya
	c.HTML(http.StatusOK, "detail_assessment.html", gin.H{
		"Detail":     d,
		"ActiveMenu": "detail-assessment",
	})
}
