package controller

import (
	"encoding/json"
	"html/template"
	"net/http"
	"pv-risk/config"
	"pv-risk/models"
	"strconv"
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
	yearBuilt, _ := strconv.Atoi(c.PostForm("year_built"))
	tActual, _ := strconv.ParseFloat(c.PostForm("thickness_actual"), 64)
	tMin, _ := strconv.ParseFloat(c.PostForm("thickness_min"), 64)
	cr, _ := strconv.ParseFloat(c.PostForm("corrosion_rate"), 64)
	pressure, _ := strconv.ParseFloat(c.PostForm("operating_pressure"), 64)
	InvtVolume, _ := strconv.ParseFloat(c.PostForm("inventory_volume"), 64)
	DmgMechanism := c.PostForm("damage_mechanism")
	InspQuality := c.PostForm("inspection_quality")

	pv := models.Assessment{
		TagNumber:         c.PostForm("tag_number"),
		YearBuilt:         yearBuilt,
		ThicknessActual:   tActual,
		ThicknessMin:      tMin,
		CorrosionRate:     cr,
		OperatingPressure: pressure,
		FluidType:         c.PostForm("fluid_type"),
		IsCritical:        c.PostForm("is_critical") == "on",
		DamageMechanism:   DmgMechanism,
		InspectionQuality: InspQuality,
		InventoryVolume:   InvtVolume,
	}

	pv = CalculateRisk(pv)

	stmt, _ := config.DB.Prepare(`
	INSERT INTO assessments
	(tag_number, year_built, thickness_actual,
	thickness_min, corrosion_rate,
	operating_pressure, fluid_type, is_critical,
	remaining_life, lof, cof, risk_index, risk_level)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
	)

	stmt.Exec(
		pv.TagNumber,
		pv.YearBuilt,
		pv.ThicknessActual,
		pv.ThicknessMin,
		pv.CorrosionRate,
		pv.OperatingPressure,
		pv.FluidType,
		pv.IsCritical,
		pv.RemainingLife,
		pv.LoF,
		pv.CoF,
		pv.RiskIndex,
		pv.RiskLevel,
	)

	c.HTML(http.StatusOK, "result.html", pv)
}
