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

// Mega-Struct untuk Laporan Komplit
type AssessmentFullDetail struct {
	AssessmentID   int
	TagNumber      string
	Description    string
	AssessmentDate string
	YearBuilt      int

	// Parameter Desain & Operasi
	DesignPressure    float64
	DesignTemp        float64
	OperatingPressure float64
	OperatingTemp     float64
	Diameter          float64
	Volume            float64
	Phase             string

	// Fluida & Lingkungan
	H2sContent    float64
	Co2Content    float64
	ChlorideIndex int
	PhIndex       int

	// Damage Mechanisms
	Atmospheric string
	Cui         string
	ExtCracking string
	Ssc         string
	Hic         string
	Co2         string
	LofScore    string // (Nilai PoF dan DF)

	// Risk & Strategy
	LofCategory        int
	CofCategory        string
	RiskLevel          string
	RiskIndex          int
	GoverningComponent string
	MaxIntervalYears   float64
	NextInspectionYear int
	RecommendedMethod  string
}

func ViewAssessmentDetail(c *gin.Context) {
	id := c.Param("id")
	db := config.DB
	var d AssessmentFullDetail

	// Super Query JOIN 4 Tabel
	query := `
		SELECT 
			a.id, COALESCE(t.tag_number, '-'), COALESCE(e.name, '-') as description, COALESCE(a.assessment_date, '-'), COALESCE(t.year_built, 0),
			COALESCE(t.design_pressure, 0), COALESCE(t.design_temp, 0), COALESCE(a.operating_pressure, 0), COALESCE(a.operating_temp, 0),
			COALESCE(t.diameter, 0), COALESCE(t.volume, 0), COALESCE(a.phase, '-'),
			COALESCE(a.h2s_content, 0), COALESCE(a.co2_content, 0), COALESCE(a.chloride_index, 0), COALESCE(a.ph_index, 0),
			COALESCE(dm.atmospheric, 'Not'), COALESCE(dm.cui, 'Not'), COALESCE(dm.ext_cracking, 'Not'), COALESCE(dm.ssc, 'Not'), 
			COALESCE(dm.hic, 'Not'), COALESCE(dm.co2, 'Not'), COALESCE(dm.lof_score, '-'),
			COALESCE(r.lof_category, 0), COALESCE(r.cof_category, '-'), COALESCE(r.risk_level, 'Pending'), COALESCE(r.risk_index, 0),
			COALESCE(r.governing_component, '-'), COALESCE(r.max_interval_years, 0), COALESCE(r.next_inspection_year, 0), COALESCE(r.recommended_method, '-')
		FROM assessments a
		JOIN trx_equipments t ON a.equipment_id = t.id
		JOIN equipments e ON t.equipment_id = e.id
		LEFT JOIN assessment_damage_mechanisms dm ON dm.assessment_id = a.id
		LEFT JOIN assessment_results r ON r.assessment_id = a.id
		WHERE a.id = ?
	`

	err := db.QueryRow(query, id).Scan(
		&d.AssessmentID, &d.TagNumber, &d.Description, &d.AssessmentDate, &d.YearBuilt,
		&d.DesignPressure, &d.DesignTemp, &d.OperatingPressure, &d.OperatingTemp,
		&d.Diameter, &d.Volume, &d.Phase,
		&d.H2sContent, &d.Co2Content, &d.ChlorideIndex, &d.PhIndex,
		&d.Atmospheric, &d.Cui, &d.ExtCracking, &d.Ssc, &d.Hic, &d.Co2, &d.LofScore,
		&d.LofCategory, &d.CofCategory, &d.RiskLevel, &d.RiskIndex,
		&d.GoverningComponent, &d.MaxIntervalYears, &d.NextInspectionYear, &d.RecommendedMethod,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found or incomplete data"})
		return
	}

	c.HTML(http.StatusOK, "detail_assessment.html", gin.H{
		"Detail":     d, // Pakai variabel "Report"
		"ActiveMenu": "detail-assessment",
	})
}
