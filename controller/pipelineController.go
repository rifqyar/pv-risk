package controller

import (
	"database/sql"
	"errors"
	"net/http"
	"pv-risk/config"
	"pv-risk/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NewPipelineController struct {
}

func PipelineController() *NewPipelineController {
	return &NewPipelineController{}
}

func (ctrl *NewPipelineController) ShowForm(c *gin.Context) {
	c.HTML(http.StatusOK, "pipeline_assessment_form.html", gin.H{
		"ActiveMenu": "pipeline-oil-form",
		"Mode":       "create",
		"Input":      pipelineOilDefaultInput(),
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
		"ActiveMenu": "pipeline-oil-detail",
		"Assessment": assessment,
	})
}

func (ctrl *NewPipelineController) EditAssessment(c *gin.Context) {
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
	c.HTML(http.StatusOK, "pipeline_assessment_form.html", gin.H{
		"ActiveMenu":   "pipeline-oil-form",
		"Mode":         "edit",
		"Assessment":   assessment,
		"Input":        assessment.Input,
		"AssessmentID": id,
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

func (ctrl *NewPipelineController) ShowGasComingSoon(c *gin.Context) {
	input := pipelineGasDefaultInput()
	c.HTML(http.StatusOK, "pipeline_assessment_form.html", gin.H{
		"ActiveMenu": "pipeline-oil-form",
		"Mode":       "create",
		"Input":      input,
	})
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
		YearUsed:                  2017,
		Service:                   "Gas",
		PipeSize:                  "323,85 mm (OD) x 10,31 mm (T) x 9966 m (L)",
		PipeLengthM:               9966,
		MaterialSpecification:     "API 5L X52",
		FlangeMaterialSpec:        "-",
		SMYSPsi:                   52000,
		InternalDesignPressurePsi: 1350,
		DesignTemperatureF:        200,
		MethodOfJoining:           "Welding",
		JointEfficiency:           1,
		CoatingType:               "3 LPE & Painting",
		CorrosionControl:          "SACP",
		AllowanceIn:               0,
		RightOfWay:                "6 - 9",
		SafetyDevice:              "3 Unit Ball Valve, 4 Unit Gate Valve, 2 Unit Check Valve, & 1 Unit PSV",
		AreaClassification:        "2",
		InspectionPeriod:          "March 2025",
		InspectionResult:          "ACCEPTABLE",
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
		MaterialStressPsi:         20000,
		PreviousSKPP:              "MIT-2021.1557-03072-PL",
		TemperatureDeratingFactor: 1,
		AssessmentBy:              "Engineer",
		InspectionPoints: []models.PipelineOilInspectionPoint{
			{InspectionPoint: "IP-2", InstallationType: "Underground", NominalThicknessMM: 10.31, RequiredThicknessMM: 7.01, ActualThicknessMM: 9.22, MeasuredYear: 2024},
			{InspectionPoint: "IP-11", InstallationType: "Underground", NominalThicknessMM: 10.31, RequiredThicknessMM: 7.01, ActualThicknessMM: 9.58, MeasuredYear: 2024},
		},
		RBI: models.PipelineOilRBIStructuralInput{
			DamageMechanism:             "Pipeline MVP Index RBI",
			InspectionEffectivity:       "Representative",
			ReleaseFluid:                "Gas",
			GenericFailureFrequency:     0.00003,
			ManagementSystemScore:       500,
			BaseTPDRate:                 1,
			BaseExternalCorrRate:        1,
			BaseInternalCorrRate:        1,
			DepthOfCover:                "1-2m",
			PatrolFrequency:             "monthly",
			ROWCondition:                "fair",
			SoilResistivity:             "1000-5000",
			CoatingCondition:            "fair",
			CPStatus:                    "normal",
			CPPotentialMV:               -900,
			FluidCorrosivity:            "medium",
			WaterContent:                "medium",
			CO2H2SPresence:              "present",
			MICRisk:                     "low",
			WallThicknessCondition:      "warning",
			BuildingCountInsidePIR:      3,
			ClassLocation:               "village",
			EmergencyResponse:           "available",
			FlowRate:                    100,
			DetectionTimeHours:          1,
			SegmentLengthBetweenValvesM: 9966,
			EnvironmentalSensitivity:    "medium",
			IsolationValveAvailable:     true,
			ConsequenceBasis:            "Pipeline MVP index-based CoF",
			ProbabilityBasis:            "PoF = GFF x max(DF_TPD, DF_EXTERNAL, DF_INTERNAL) x FMS",
			EngineeringNotes:            "Gas pipeline defaults from workbook 1. Kalkulasi NPS 12 Main Gas Trunkline from SB#1 to WB PPF.xlsx.",
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
		YearUsed:                  2017,
		Service:                   "Oil",
		PipeSize:                  "219,08 mm (OD) x 8,18 mm (T) x 9966 m (L)",
		PipeLengthM:               9966,
		MaterialSpecification:     "API 5L Gr.B",
		FlangeMaterialSpec:        "-",
		SMYSPsi:                   35000,
		InternalDesignPressurePsi: 1000,
		DesignTemperatureF:        200,
		MethodOfJoining:           "Welding",
		JointEfficiency:           1,
		CoatingType:               "3LPE & Painting",
		CorrosionControl:          "-",
		AllowanceIn:               0,
		RightOfWay:                "6 - 9",
		SafetyDevice:              "1 Unit Ball Valve & 3 Unit Check Valve",
		AreaClassification:        "-",
		InspectionPeriod:          "March 2025",
		InspectionResult:          "ACCEPTABLE",
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
		MaterialStressPsi:         20000,
		PreviousSKPP:              "MIT-2021.1557-03072-PL",
		TemperatureDeratingFactor: 1,
		AssessmentBy:              "Engineer",
		InspectionPoints: []models.PipelineOilInspectionPoint{
			{InspectionPoint: "IP-82", InstallationType: "Above Ground", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 7.98, MeasuredYear: 2025},
			{InspectionPoint: "IP-8 A", InstallationType: "Above Ground", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 6.12, MeasuredYear: 2025},
			{InspectionPoint: "IP-8 C", InstallationType: "Above Ground", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 6.21, MeasuredYear: 2025},
		},
		RBI: models.PipelineOilRBIStructuralInput{
			DamageMechanism:             "Pipeline MVP Index RBI",
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
			CoatingCondition:            "fair",
			CPStatus:                    "normal",
			CPPotentialMV:               -900,
			FluidCorrosivity:            "medium",
			WaterContent:                "medium",
			CO2H2SPresence:              "present",
			MICRisk:                     "low",
			WallThicknessCondition:      "warning",
			BuildingCountInsidePIR:      0,
			ClassLocation:               "remote",
			EmergencyResponse:           "available",
			FlowRate:                    100,
			DetectionTimeHours:          1,
			SegmentLengthBetweenValvesM: 9966,
			EnvironmentalSensitivity:    "medium",
			IsolationValveAvailable:     true,
			ConsequenceBasis:            "Pipeline MVP index-based CoF",
			ProbabilityBasis:            "PoF = GFF x max(DF_TPD, DF_EXTERNAL, DF_INTERNAL) x FMS",
			EngineeringNotes:            "MVP simplified pipeline RBI calculation.",
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
