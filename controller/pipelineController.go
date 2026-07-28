package controller

import (
	"database/sql"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"pv-risk/config"
	"pv-risk/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type NewPipelineController struct {
}

func PipelineController() *NewPipelineController {
	return &NewPipelineController{}
}

func (ctrl *NewPipelineController) ShowForm(c *gin.Context) {
	material, methods, err := pipelineMasterDataForForm()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching pipeline master data: %v", err)
		return
	}

	c.HTML(http.StatusOK, "pipeline_assessment_form.html", gin.H{
		"ActiveMenu":                    "pipeline-oil-form",
		"Mode":                          "create",
		"Material":                      material,
		"PipelineInspectionMethods":     methods,
		"PipelineNonIntrusiveMethods":   models.PipelineInspectionMethodsByScope(methods, "nonintrusive"),
		"PipelineIntrusiveMethods":      models.PipelineInspectionMethodsByScope(methods, "intrusive"),
		"Input":                         pipelineOilDefaultInput(),
		"PipelineDamageMechanismGroups": models.PipelineDamageMechanismGroups(),
		"PipelineDamageMechanismSource": models.PipelineDamageMechanismSource,
	})
}

func (ctrl *NewPipelineController) SubmitAssessment(c *gin.Context) {
	var input models.PipelineOilInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid JSON Payload: " + err.Error()})
		return
	}

	service := models.NewPipelineOilService(config.DB)
	id, err := service.CreateDraftAssessment(input)
	if err != nil {
		c.JSON(statusFromPipelineError(err), gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Pipeline draft saved.", "id": id})
}

func (ctrl *NewPipelineController) ShowListAssessment(c *gin.Context) {
	service := models.NewPipelineOilService(config.DB)
	assessments, err := service.ListAssessments()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.HTML(http.StatusOK, "pipeline_assessment_list.html", gin.H{
		"ActiveMenu":  "pipeline-oil-list",
		"Assessments": assessments,
	})
}

func (ctrl *NewPipelineController) ViewAssessmentDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid assessment id")
		return
	}
	service := models.NewPipelineOilService(config.DB)
	assessment, err := service.GetAssessmentDetail(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.String(http.StatusNotFound, "Pipeline assessment not found")
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.HTML(http.StatusOK, "pipeline_assessment_detail.html", gin.H{
		"ActiveMenu":          "pipeline-oil-detail",
		"Assessment":          assessment,
		"Versions":            mustPipelineVersions(service, id),
		"AuditEvents":         mustPipelineAuditEvents(service, id),
		"StandardsReferences": models.PipelineStandardsReferences(),
	})
}

func (ctrl *NewPipelineController) CompareAssessment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid assessment id")
		return
	}
	service := models.NewPipelineOilService(config.DB)
	assessment, err := service.GetAssessmentDetail(id)
	if err != nil {
		c.String(http.StatusNotFound, "Pipeline assessment not found")
		return
	}
	comparison, err := service.GetAssessmentComparison(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.HTML(http.StatusOK, "pipeline_assessment_compare.html", gin.H{
		"ActiveMenu":  "pipeline-oil-compare",
		"Assessment":  assessment,
		"Comparison":  comparison,
		"AuditEvents": mustPipelineAuditEvents(service, id),
	})
}

func (ctrl *NewPipelineController) EditAssessment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid assessment id")
		return
	}
	material, methods, err := pipelineMasterDataForForm()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching pipeline master data: %v", err)
		return
	}
	service := models.NewPipelineOilService(config.DB)
	assessment, err := service.GetAssessmentDetail(id)
	if err != nil {
		c.String(http.StatusNotFound, "Pipeline assessment not found")
		return
	}
	input := assessment.Input
	if !models.IsValidPipelineDamageMechanism(input.RiskInput.DamageMechanism) {
		input.RiskInput.DamageMechanism = "internal_corrosion"
	} else {
		input.RiskInput.DamageMechanism = models.NormalizePipelineDamageMechanism(input.RiskInput.DamageMechanism)
	}
	c.HTML(http.StatusOK, "pipeline_assessment_form.html", gin.H{
		"ActiveMenu":                    "pipeline-oil-form",
		"Mode":                          "edit",
		"Material":                      material,
		"PipelineInspectionMethods":     methods,
		"PipelineNonIntrusiveMethods":   models.PipelineInspectionMethodsByScope(methods, "nonintrusive"),
		"PipelineIntrusiveMethods":      models.PipelineInspectionMethodsByScope(methods, "intrusive"),
		"Assessment":                    assessment,
		"Input":                         input,
		"AssessmentID":                  id,
		"PipelineDamageMechanismGroups": models.PipelineDamageMechanismGroups(),
		"PipelineDamageMechanismSource": models.PipelineDamageMechanismSource,
	})
}

func (ctrl *NewPipelineController) DeleteAssessment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid assessment id"})
		return
	}
	service := models.NewPipelineOilService(config.DB)
	if err = service.ArchiveAssessment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Pipeline assessment archived."})
}

func (ctrl *NewPipelineController) UpdateAssessment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid assessment id"})
		return
	}
	var input models.PipelineOilInput
	if err = c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid JSON Payload: " + err.Error()})
		return
	}
	service := models.NewPipelineOilService(config.DB)
	if err = service.UpdateDraftAssessment(id, input); err != nil {
		c.JSON(statusFromPipelineError(err), gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Pipeline draft updated.", "id": id})
}

func (ctrl *NewPipelineController) CalculateAssessment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid assessment id"})
		return
	}
	var input models.PipelineOilInput
	if err = c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid JSON Payload: " + err.Error()})
		return
	}
	service := models.NewPipelineOilService(config.DB)
	result, err := service.CalculateAssessment(id, input)
	if err != nil {
		c.JSON(statusFromPipelineError(err), gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Pipeline assessment saved.", "id": id, "result": result})
}

func (ctrl *NewPipelineController) PreviewAssessment(c *gin.Context) {
	var input models.PipelineOilInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid JSON Payload: " + err.Error()})
		return
	}
	service := models.NewPipelineOilService(config.DB)
	result, err := service.PreviewAssessment(input)
	if err != nil {
		c.JSON(statusFromPipelineError(err), gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Pipeline calculation ready for review.", "result": result})
}

func (ctrl *NewPipelineController) ApproveAssessment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid assessment id"})
		return
	}
	var payload struct {
		Actor string `json:"actor"`
		Note  string `json:"note"`
	}
	_ = c.ShouldBindJSON(&payload)
	service := models.NewPipelineOilService(config.DB)
	assessment, err := service.GetAssessmentDetail(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Pipeline assessment not found"})
		return
	}
	if strings.TrimSpace(payload.Actor) == "" {
		payload.Actor = assessment.Input.AssessmentBy
	}
	if err = service.RecordApproval(id, payload.Actor, payload.Note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Pipeline approval recorded."})
}

func (ctrl *NewPipelineController) AuditPDFExport(c *gin.Context) {
	ctrl.recordExport(c, "PDF")
}

func (ctrl *NewPipelineController) ExportExcel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid assessment id")
		return
	}
	service := models.NewPipelineOilService(config.DB)
	assessment, err := service.GetAssessmentDetail(id)
	if err != nil {
		c.String(http.StatusNotFound, "Pipeline assessment not found")
		return
	}
	_ = service.RecordExport(id, assessment.Input.AssessmentBy, "EXCEL")
	filename := fmt.Sprintf("Pipeline_Assessment_%s.xls", safeExportName(assessment.Input.ReportNo))
	c.Header("Content-Type", "application/vnd.ms-excel; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.String(http.StatusOK, buildPipelineExcelHTML(assessment, mustPipelineAuditEvents(service, id)))
}

func (ctrl *NewPipelineController) recordExport(c *gin.Context, format string) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid assessment id"})
		return
	}
	service := models.NewPipelineOilService(config.DB)
	assessment, err := service.GetAssessmentDetail(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Pipeline assessment not found"})
		return
	}
	if err = service.RecordExport(id, assessment.Input.AssessmentBy, format); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Export audit recorded."})
}

func (ctrl *NewPipelineController) ShowGasComingSoon(c *gin.Context) {
	material, methods, err := pipelineMasterDataForForm()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching pipeline master data: %v", err)
		return
	}
	input := pipelineGasDefaultInput()
	c.HTML(http.StatusOK, "pipeline_assessment_form.html", gin.H{
		"ActiveMenu":                    "pipeline-oil-form",
		"Mode":                          "create",
		"Material":                      material,
		"PipelineInspectionMethods":     methods,
		"PipelineNonIntrusiveMethods":   models.PipelineInspectionMethodsByScope(methods, "nonintrusive"),
		"PipelineIntrusiveMethods":      models.PipelineInspectionMethodsByScope(methods, "intrusive"),
		"Input":                         input,
		"PipelineDamageMechanismGroups": models.PipelineDamageMechanismGroups(),
		"PipelineDamageMechanismSource": models.PipelineDamageMechanismSource,
	})
}

func (ctrl *NewPipelineController) ShowPipelineMasterData(c *gin.Context) {
	materials, err := models.GetPipelineMaterials(config.DB, false)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching pipeline materials: %v", err)
		return
	}
	methods, err := models.GetPipelineInspectionMethods(config.DB, false)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching pipeline inspection methods: %v", err)
		return
	}
	c.HTML(http.StatusOK, "pipeline_master_data.html", gin.H{
		"ActiveMenu":        "pipeline-master-data",
		"Materials":         materials,
		"InspectionMethods": methods,
		"QueryStatus":       c.Query("status"),
		"QueryMessage":      c.Query("message"),
	})
}

func (ctrl *NewPipelineController) SavePipelineMaterial(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	smys, _ := strconv.ParseFloat(strings.ReplaceAll(c.PostForm("smys"), ",", "."), 64)
	allowableStress, _ := strconv.ParseFloat(strings.ReplaceAll(c.PostForm("allowable_stress"), ",", "."), 64)
	err := models.SavePipelineMaterial(config.DB, models.PipelineMaterial{
		ID:                    id,
		Name:                  strings.TrimSpace(c.PostForm("name")),
		MaterialSpecification: strings.TrimSpace(c.PostForm("material_specification")),
		Grade:                 strings.TrimSpace(c.PostForm("grade")),
		SMYS:                  smys,
		AllowableStress:       allowableStress,
		Notes:                 strings.TrimSpace(c.PostForm("notes")),
	})
	redirectPipelineMasterData(c, err)
}

func (ctrl *NewPipelineController) DeactivatePipelineMaterial(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err == nil {
		err = models.DeactivatePipelineMaterial(config.DB, id)
	}
	redirectPipelineMasterData(c, err)
}

func (ctrl *NewPipelineController) ActivatePipelineMaterial(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err == nil {
		err = models.ActivatePipelineMaterial(config.DB, id)
	}
	redirectPipelineMasterData(c, err)
}

func (ctrl *NewPipelineController) SavePipelineInspectionMethod(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	err := models.SavePipelineInspectionMethod(config.DB, models.PipelineInspectionMethod{
		ID:          id,
		Name:        strings.TrimSpace(c.PostForm("name")),
		Scope:       c.PostForm("scope"),
		Effectivity: c.PostForm("effectivity"),
		Notes:       strings.TrimSpace(c.PostForm("notes")),
	})
	redirectPipelineMasterData(c, err)
}

func (ctrl *NewPipelineController) DeactivatePipelineInspectionMethod(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err == nil {
		err = models.DeactivatePipelineInspectionMethod(config.DB, id)
	}
	redirectPipelineMasterData(c, err)
}

func (ctrl *NewPipelineController) ActivatePipelineInspectionMethod(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err == nil {
		err = models.ActivatePipelineInspectionMethod(config.DB, id)
	}
	redirectPipelineMasterData(c, err)
}

func pipelineMasterDataForForm() ([]models.PipelineMaterial, []models.PipelineInspectionMethod, error) {
	materials, err := models.GetMaterial(config.DB)
	if err != nil {
		return nil, nil, err
	}
	methods, err := models.GetPipelineInspectionMethods(config.DB, true)
	if err != nil {
		return nil, nil, err
	}
	return materials, methods, nil
}

func redirectPipelineMasterData(c *gin.Context, err error) {
	values := url.Values{}
	if err != nil {
		values.Set("status", "error")
		values.Set("message", err.Error())
		c.Redirect(http.StatusSeeOther, "/assessment-pipeline/master-data?"+values.Encode())
		return
	}
	values.Set("status", "success")
	c.Redirect(http.StatusSeeOther, "/assessment-pipeline/master-data?"+values.Encode())
}

func pipelineGasDefaultInput() models.PipelineOilInput {
	return models.PipelineOilInput{
		ReportNo:                  "PR 2057/PL-MIT/120/2025",
		PlaceIssued:               "Jakarta",
		DateIssued:                "July 2025",
		OwnerUser:                 "PetroChina International Jabung Ltd",
		Contractor:                "PT Meindo Elang Indah",
		Location:                  "Betara, Tanjung Jabung Barat, Jambi",
		LineIdentification:        "NPS 12 Main Gas Trunkline from Pig Launcher SB#1 Station to Pig Receiver WB PPF",
		YearBuilt:                 2017,
		YearUsed:                  "2017",
		Service:                   "Natural gas",
		PipeSize:                  "323,85 mm (OD) x 10,31 mm (T) x 9966 m (L)",
		PipeLengthM:               9966,
		MaterialSpecification:     "API 5L X 52",
		FlangeMaterialSpec:        "-",
		SMYSPsi:                   52000,
		InternalDesignPressurePsi: 1350,
		DesignTemperatureF:        200,
		MethodOfJoining:           "Welding",
		JointEfficiency:           1,
		CoatingType:               "3 LPE & Painting",
		CorrosionControl:          "SACP",
		AllowanceIn:               0,
		RightOfWay:                "10",
		SafetyDevice:              "3 Unit Ball Valve, 4 Unit Gate Valve, 2 Unit Check Valve, & 1 Unit PSV",
		AreaClassification:        "2",
		InspectionPeriod:          "March 2025",
		ApplicableCode:            "ASME B31.8",
		OutsideDiameterIn:         12.75,
		OperatingPressurePsi:      650,
		RadiographicPercent:       100,
		NominalWallThicknessMM:    11.13,
		ActualWallThicknessMM:     9.22,
		TypeOfInstallation:        "Underground",
		QualityFactor:             1,
		WeldJointStrengthFactor:   1,
		DesignFactor:              0.6,
		MaterialStressPsi:         52000,
		PreviousSKPP:              "MIT-2021.1557-03072-PL",
		TemperatureDeratingFactor: 1,
		AssessmentBy:              "Engineer",
		InspectionPoints: []models.PipelineOilInspectionPoint{
			{InspectionPoint: "IP-2", InstallationType: "Underground", NominalThicknessMM: 10.31, RequiredThicknessMM: 7.01, ActualThicknessMM: 9.22, MeasuredYear: "2024"},
			{InspectionPoint: "IP-11", InstallationType: "Underground", NominalThicknessMM: 10.31, RequiredThicknessMM: 7.01, ActualThicknessMM: 9.58, MeasuredYear: "2024"},
		},
		RiskInput: models.PipelineOilRiskInput{
			DamageMechanism:             "internal_corrosion",
			InspectionEffectivity:       "Representative",
			ReleaseFluid:                "Natural gas",
			GenericFailureFrequency:     0.00003,
			ManagementSystemScore:       500,
			BaseTPDRate:                 1,
			BaseExternalCorrRate:        1,
			BaseInternalCorrRate:        1,
			DepthOfCover:                "1-2m",
			PatrolFrequency:             "monthly",
			ROWCondition:                "fair",
			SoilResistivity:             "1000-5000",
			CoatingCondition:            "Good",
			CPStatus:                    "normal",
			CPPotentialMV:               -900,
			PHLevel:                     "6.5-8.5",
			FluidCorrosivityMPY:         "2-5 mpy",
			InhibitorEffectiveness:      "None",
			BiocideTreatment:            "Not Required",
			CorrosionMonitoringResult:   "Not Applicable",
			H2SPpm:                      "<50 ppm",
			PWHTStatus:                  "Unknown",
			WeldJointType:               "Seamless",
			FlowVelocityCondition:       "Moderate (3-10 m/s)",
			SolidContent:                "None",
			EnvExtCracking:              "None",
			OneCallSystem:               "None",
			PrevExtCorrosion:            "No Finding",
			PrevIntThinning:             "No Finding",
			PrevIntCracking:             "No Finding",
			PrevLocIntCorrosion:         "No Finding",
			InsulationCondition:         "Not Applicable",
			ExtCoatingCondition:         "Good",
			BuildingCountInsidePIR:      3,
			ClassLocation:               "village",
			FlowRate:                    100,
			DetectionTimeHours:          1,
			SegmentLengthBetweenValvesM: 9966,
			EnvironmentalSensitivity:    "medium",
			IsolationValveAvailable:     true,
			ConsequenceBasis:            "Pipeline MVP index-based CoF",
			ProbabilityBasis:            "PoF = GFF x governing DM score x FMS",
			EngineeringNotes:            "Gas pipeline defaults from workbook 1.",
			RequiresConfirmation:        false,
		},
	}
}

func pipelineOilDefaultInput() models.PipelineOilInput {
	return models.PipelineOilInput{
		ReportNo:                  "PR 2057/PL-MIT/13/2025",
		PlaceIssued:               "Jakarta",
		DateIssued:                "July 2025",
		OwnerUser:                 "PetroChina International Jabung Ltd",
		Contractor:                "PT Meindo Elang Indah",
		Location:                  "Tanjung Jabung Barat, Jambi",
		LineIdentification:        "NPS 8 Main Oil Trunkline from KP-01 (Jembatan Kembar) to Pig Receiver (202-VR-100) WB PPF",
		YearBuilt:                 2017,
		YearUsed:                  "2017",
		Service:                   "Oil",
		PipeSize:                  "219,08 mm (OD) x 8,18 mm (T) x 9966 m (L)",
		PipeLengthM:               9966,
		MaterialSpecification:     "API 5L Gr B",
		FlangeMaterialSpec:        "-",
		SMYSPsi:                   35000,
		InternalDesignPressurePsi: 1000,
		DesignTemperatureF:        200,
		MethodOfJoining:           "Welding",
		JointEfficiency:           1,
		CoatingType:               "3LPE & Painting",
		CorrosionControl:          "-",
		AllowanceIn:               0,
		RightOfWay:                "10",
		SafetyDevice:              "1 Unit Ball Valve & 3 Unit Check Valve",
		AreaClassification:        "-",
		InspectionPeriod:          "March 2025",
		ApplicableCode:            "ASME B31.4",
		OutsideDiameterIn:         8.625,
		OperatingPressurePsi:      17.65,
		RadiographicPercent:       100,
		NominalWallThicknessMM:    8.18,
		ActualWallThicknessMM:     6.12,
		TypeOfInstallation:        "Aboveground & Underground",
		QualityFactor:             1,
		WeldJointStrengthFactor:   1,
		DesignFactor:              0.72,
		MaterialStressPsi:         35000,
		PreviousSKPP:              "MIT-2021.1557-03072-PL",
		TemperatureDeratingFactor: 1,
		AssessmentBy:              "Engineer",
		InspectionPoints: []models.PipelineOilInspectionPoint{
			{InspectionPoint: "IP-82", InstallationType: "Above Ground", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 7.98, MeasuredYear: "2025"},
			{InspectionPoint: "IP-8 A", InstallationType: "Above Ground", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 6.12, MeasuredYear: "2025"},
			{InspectionPoint: "IP-8 C", InstallationType: "Above Ground", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 6.21, MeasuredYear: "2025"},
		},
		RiskInput: models.PipelineOilRiskInput{
			DamageMechanism:             "internal_corrosion",
			InspectionEffectivity:       "Representative",
			ReleaseFluid:                "Oil",
			GenericFailureFrequency:     0.00003,
			ManagementSystemScore:       500,
			BaseTPDRate:                 1,
			BaseExternalCorrRate:        1,
			BaseInternalCorrRate:        1,
			DepthOfCover:                "1-2m",
			PatrolFrequency:             "monthly",
			ROWCondition:                "fair",
			SoilResistivity:             "1000-5000",
			CoatingCondition:            "Good",
			CPStatus:                    "normal",
			CPPotentialMV:               -900,
			PHLevel:                     "6.5-8.5",
			FluidCorrosivityMPY:         "2-5 mpy",
			InhibitorEffectiveness:      "None",
			BiocideTreatment:            "Not Required",
			CorrosionMonitoringResult:   "Not Applicable",
			H2SPpm:                      "<50 ppm",
			PWHTStatus:                  "Unknown",
			WeldJointType:               "Seamless",
			FlowVelocityCondition:       "Moderate (3-10 m/s)",
			SolidContent:                "None",
			EnvExtCracking:              "None",
			OneCallSystem:               "None",
			PrevExtCorrosion:            "No Finding",
			PrevIntThinning:             "No Finding",
			PrevIntCracking:             "No Finding",
			PrevLocIntCorrosion:         "No Finding",
			InsulationCondition:         "Not Applicable",
			ExtCoatingCondition:         "Good",
			BuildingCountInsidePIR:      0,
			ClassLocation:               "remote",
			FlowRate:                    100,
			DetectionTimeHours:          1,
			SegmentLengthBetweenValvesM: 9966,
			EnvironmentalSensitivity:    "medium",
			IsolationValveAvailable:     true,
			ConsequenceBasis:            "Pipeline MVP index-based CoF",
			ProbabilityBasis:            "PoF = GFF x governing DM score x FMS",
			EngineeringNotes:            "MVP simplified pipeline RiskInput calculation.",
			RequiresConfirmation:        false,
		},
	}
}

func statusFromPipelineError(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.Is(err, models.ErrPipelineFinalized) {
		return http.StatusConflict
	}
	if errors.Is(err, models.ErrPipelineValidation) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func mustPipelineVersions(service *models.PipelineOilService, id int) []models.PipelineOilAssessmentVersion {
	versions, err := service.GetAssessmentVersions(id)
	if err != nil {
		return nil
	}
	return versions
}

func mustPipelineAuditEvents(service *models.PipelineOilService, id int) []models.PipelineOilAuditEvent {
	events, err := service.GetAssessmentAuditEvents(id)
	if err != nil {
		return nil
	}
	return events
}

func safeExportName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "pipeline-assessment"
	}
	replacer := strings.NewReplacer("\\", "_", "/", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_", " ", "_")
	return replacer.Replace(value)
}

func buildPipelineExcelHTML(assessment *models.PipelineOilAssessment, auditEvents []models.PipelineOilAuditEvent) string {
	var b strings.Builder
	b.WriteString("<html><head><meta charset=\"utf-8\"></head><body>")
	b.WriteString("<h2>Pipeline Risk Assessment Report</h2>")
	b.WriteString("<table border=\"1\">")
	writeExcelRow(&b, "Tag Number", assessment.Input.ReportNo)
	writeExcelRow(&b, "Pipeline Name", assessment.Input.LineIdentification)
	writeExcelRow(&b, "Location", assessment.Input.Location)
	writeExcelRow(&b, "Fluid Type", assessment.Input.Service)
	writeExcelRow(&b, "Applicable Code", assessment.Input.ApplicableCode)
	writeExcelRow(&b, "Assessment Status", string(assessment.Status))
	writeExcelRow(&b, "Formula Version", assessment.FormulaVersion)
	writeExcelRow(&b, "Created At", assessment.CreatedAt)
	writeExcelRow(&b, "Updated At", assessment.UpdatedAt)
	b.WriteString("</table>")

	b.WriteString("<h3>Process and Condition</h3><table border=\"1\">")
	writeExcelRow(&b, "CO2 Content", fmt.Sprintf("%.8g mol%%", assessment.Input.RiskInput.CO2Content))
	writeExcelRow(&b, "H2S Content", fmt.Sprintf("%.8g ppm", assessment.Input.RiskInput.H2SContent))
	writeExcelRow(&b, "H2O Content", fmt.Sprintf("%.8g lb/mmscf", assessment.Input.RiskInput.H2OContent))
	writeExcelRow(&b, "pCO2", fmt.Sprintf("%.8g psig", assessment.Input.RiskInput.CO2PartialPressurePSIG))
	writeExcelRow(&b, "pH2S", fmt.Sprintf("%.8g psig", assessment.Input.RiskInput.H2SPartialPressurePSIG))
	b.WriteString("</table>")

	if assessment.Result != nil {
		b.WriteString("<h3>Risk Result</h3><table border=\"1\">")
		writeExcelRow(&b, "PoF", assessment.Result.PoF)
		writeExcelRow(&b, "PoF Value", fmt.Sprintf("%.8g", assessment.Result.PoFValue))
		writeExcelRow(&b, "CoF", assessment.Result.CoF)
		writeExcelRow(&b, "CoF Value", fmt.Sprintf("%.8g", assessment.Result.CoFValue))
		writeExcelRow(&b, "Risk Code", assessment.Result.FinalRiskCode)
		writeExcelRow(&b, "Risk Level", assessment.Result.FinalRiskLevel)
		writeExcelRow(&b, "Governing DM", assessment.Result.GoverningDamageMechanism)
		writeExcelMultilineRow(&b, "Immediate Actions", strings.Join(assessment.Result.RecommendationGroups.ImmediateActions, "\n"))
		writeExcelMultilineRow(&b, "Inspection / Monitoring", strings.Join(assessment.Result.RecommendationGroups.InspectionMonitor, "\n"))
		writeExcelMultilineRow(&b, "Long-Term Mitigation", strings.Join(assessment.Result.RecommendationGroups.LongTermMitigation, "\n"))
		writeExcelRow(&b, "Recommendation Source", assessment.Result.RecommendationSource)
		writeExcelRow(&b, "Recommendation Rule", assessment.Result.RecommendationRuleName)
		writeExcelRow(&b, "Recommendation Confidence", assessment.Result.RecommendationConfidence)
		writeExcelMultilineRow(&b, "Engineering Notes", assessment.Input.RiskInput.EngineeringNotes)
		b.WriteString("</table>")

		b.WriteString("<h3>Damage Mechanism Results</h3><table border=\"1\"><tr><th>Category</th><th>Mechanism</th><th>Severity</th><th>Score</th><th>Source</th><th>Confidence</th><th>Status</th></tr>")
		for _, dm := range assessment.Result.DamageMechanismResults {
			b.WriteString("<tr><td>" + html.EscapeString(dm.Category) + "</td><td>" + html.EscapeString(dm.Label) + "</td><td>" + html.EscapeString(dm.Severity) + "</td><td>" + fmt.Sprintf("%.2f", dm.Score) + "</td><td>" + html.EscapeString(dm.SourceStandard) + "</td><td>" + html.EscapeString(dm.ConfidenceLevel) + "</td><td>" + html.EscapeString(dm.RuleStatus) + "</td></tr>")
		}
		b.WriteString("</table>")

	}

	b.WriteString("<h3>Audit Information</h3><table border=\"1\"><tr><th>Time</th><th>Action</th><th>User</th><th>Note</th></tr>")
	for _, event := range auditEvents {
		b.WriteString("<tr><td>" + html.EscapeString(event.CreatedAt) + "</td><td>" + html.EscapeString(event.Action) + "</td><td>" + html.EscapeString(event.Actor) + "</td><td>" + html.EscapeString(event.Note) + "</td></tr>")
	}
	b.WriteString("</table>")
	b.WriteString("<p>Exported at " + html.EscapeString(time.Now().Format(time.RFC3339)) + "</p>")
	b.WriteString("</body></html>")
	return b.String()
}

func writeExcelRow(b *strings.Builder, label, value string) {
	b.WriteString("<tr><th>")
	b.WriteString(html.EscapeString(label))
	b.WriteString("</th><td>")
	b.WriteString(html.EscapeString(value))
	b.WriteString("</td></tr>")
}

func writeExcelMultilineRow(b *strings.Builder, label, value string) {
	b.WriteString("<tr><th>")
	b.WriteString(html.EscapeString(label))
	b.WriteString("</th><td style=\"white-space: pre-wrap;\">")
	b.WriteString(html.EscapeString(value))
	b.WriteString("</td></tr>")
}
