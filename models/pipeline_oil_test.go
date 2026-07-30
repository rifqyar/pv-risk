package models

import (
	"database/sql"
	"encoding/json"
	"math"
	"os"
	"pv-risk/migrations"
	"pv-risk/seeder"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func samplePipelineOilInput() PipelineOilInput {
	return PipelineOilInput{
		ReportNo:                  "PR 2057/PL-MIT/13/2025",
		LineIdentification:        "NPS 8 Main Oil Trunkline",
		YearBuilt:                 2017,
		YearUsed:                  "2017",
		Service:                   "Oil",
		PipeLengthM:               9966,
		SMYSPsi:                   35000,
		InternalDesignPressurePsi: 1000,
		DesignTemperatureF:        200,
		JointEfficiency:           1,
		AllowanceIn:               0,
		ApplicableCode:            "ASME B31.4",
		OutsideDiameterIn:         8.625,
		OperatingPressurePsi:      17.65,
		NominalWallThicknessMM:    8.18,
		ActualWallThicknessMM:     6.12,
		QualityFactor:             1,
		WeldJointStrengthFactor:   1,
		DesignFactor:              0.72,
		MaterialStressPsi:         20000,
		AssessmentBy:              "Engineer",
		RiskInput: PipelineOilRiskInput{
			GenericFailureFrequency: 0.00003,
			ManagementSystemScore:   500,
			DamageFactor:            12,
			ConsequenceFinancial:    250000,
		},
		InspectionPoints: []PipelineOilInspectionPoint{
			{InspectionPoint: "IP-82", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 7.98, MeasuredYear: "2025"},
			{InspectionPoint: "IP-8 A", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 6.12, MeasuredYear: "2025"},
			{InspectionPoint: "IP-8 C", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 6.21, MeasuredYear: "2025"},
		},
	}
}

func TestPipelineOilInputJSONUsesCanonicalRiskInput(t *testing.T) {
	var input PipelineOilInput
	payload := []byte(`{"report_no":"P-001","risk_input":{"co2_content":20,"h2s_content":30,"h2o_content":10}}`)
	if err := json.Unmarshal(payload, &input); err != nil {
		t.Fatalf("unmarshal pipeline input: %v", err)
	}
	if input.RiskInput.CO2Content != 20 || input.RiskInput.H2SContent != 30 || input.RiskInput.H2OContent != 10 {
		t.Fatalf("risk input did not bind from canonical JSON: co2=%.2f h2s=%.2f h2o=%.2f", input.RiskInput.CO2Content, input.RiskInput.H2SContent, input.RiskInput.H2OContent)
	}
}

func TestPipelinePreviousConditionDamageLevelsOnlyApplyWhenDamaged(t *testing.T) {
	input := samplePipelineOilInput()
	input.RiskInput.InsulationCondition = "Good"
	input.RiskInput.InsulationDamageLevel = "Severe"
	input.RiskInput.ExtCoatingCondition = "Not Inspectable"
	input.RiskInput.ExtCoatingDamageLevel = "Large"

	applyPipelineOilDefaults(&input)

	if input.RiskInput.InsulationDamageLevel != "" || input.RiskInput.ExtCoatingDamageLevel != "" {
		t.Fatalf("non-damaged condition levels should be ignored, got insulation=%q ext_coating=%q", input.RiskInput.InsulationDamageLevel, input.RiskInput.ExtCoatingDamageLevel)
	}
	if errs := ValidatePipelineOilCalculation(input); len(errs) > 0 {
		t.Fatalf("hidden damage levels should not fail validation: %v", errs)
	}

	input.RiskInput.InsulationCondition = "Damaged"
	input.RiskInput.ExtCoatingCondition = "Damaged"
	applyPipelineOilDefaults(&input)
	if input.RiskInput.InsulationDamageLevel != "Small" || input.RiskInput.ExtCoatingDamageLevel != "Small" {
		t.Fatalf("damaged conditions should default blank levels to Small, got insulation=%q ext_coating=%q", input.RiskInput.InsulationDamageLevel, input.RiskInput.ExtCoatingDamageLevel)
	}
}

func newPipelineTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)
	migrations.PipelineOilAssessmentTables(db)
	migrations.PipelineOilReportingTables(db)
	return db
}

func samplePipelineGasWorkbookInput() PipelineOilInput {
	in := samplePipelineOilInput()
	in.ReportNo = "PR 2057/PL-MIT/120/2025"
	in.LineIdentification = "NPS 12 Main Gas Trunkline from Pig Launcher SB#1 Station to Pig Receiver WB PPF"
	in.Location = "Betara, Tanjung Jabung Barat, Jambi"
	in.Service = "Gas"
	in.PipeSize = "323,85 mm (OD) x 10,31 mm (T) x 9966 m (L)"
	in.MaterialSpecification = "API 5L X52"
	in.SMYSPsi = 52000
	in.InternalDesignPressurePsi = 1350
	in.CoatingType = "3 LPE & Painting"
	in.CorrosionControl = "SACP"
	in.SafetyDevice = "3 Unit Ball Valve, 4 Unit Gate Valve, 2 Unit Check Valve, & 1 Unit PSV"
	in.AreaClassification = "2"
	in.ApplicableCode = "ASME B31.8"
	in.OutsideDiameterIn = 12.75
	in.OperatingPressurePsi = 650
	in.NominalWallThicknessMM = 11.13
	in.ActualWallThicknessMM = 9.22
	in.TypeOfInstallation = "Underground"
	in.DesignFactor = 0.6
	in.InspectionPoints = []PipelineOilInspectionPoint{
		{InspectionPoint: "IP-2", NominalThicknessMM: 10.31, RequiredThicknessMM: 7.01, ActualThicknessMM: 9.22, MeasuredYear: "2024"},
		{InspectionPoint: "IP-11", NominalThicknessMM: 10.31, RequiredThicknessMM: 7.01, ActualThicknessMM: 9.58, MeasuredYear: "2024"},
	}
	in.RiskInput.ReleaseFluid = "Gas"
	in.RiskInput.BuildingCountInsidePIR = 3
	in.RiskInput.ClassLocation = "village"
	return in
}

func sampleRawGasPipingWorkbookInput() PipelineOilInput {
	in := samplePipelineOilInput()
	in.ReportNo = "PR 2057/PL-MIT/132/2025"
	in.LineIdentification = "NPS 4 Raw Gas Flowline from NEB #84 Wellhead to NEB #94 Manifold"
	in.OwnerUser = "PetroChina International Jabung Ltd"
	in.Contractor = "PetroChina International Jabung Ltd"
	in.Location = "Tanjung Jabung Barat, Jambi"
	in.YearUsed = "2018"
	in.Service = "Gas"
	in.PipeSize = "114,3 mm (OD) x 8,56 mm (T) x 280 m (L)"
	in.PipeLengthM = 280
	in.MaterialSpecification = "ASTM A312 TP 316L"
	in.SMYSPsi = 37700
	in.InternalDesignPressurePsi = 1520
	in.CoatingType = "Wrapping"
	in.CorrosionControl = "N/A"
	in.RightOfWay = "9"
	in.SafetyDevice = "1 Unit Gate Valve & 1 Unit Shut Down Valve"
	in.AreaClassification = "1 Div.2"
	in.ApplicableCode = "ASME B31.3"
	in.OutsideDiameterIn = 4.5
	in.OperatingPressurePsi = 600
	in.NominalWallThicknessMM = 8.56
	in.ActualWallThicknessMM = 8.19
	in.TypeOfInstallation = "Underground & Aboveground"
	in.WeldJointStrengthFactor = 1
	in.DesignFactor = 0.4
	in.MaterialStressPsi = 20000
	in.InspectionPoints = []PipelineOilInspectionPoint{
		{InspectionPoint: "IP-2", NominalThicknessMM: 8.56, RequiredThicknessMM: 4.19, ActualThicknessMM: 8.19, MeasuredYear: "2025"},
		{InspectionPoint: "IP-15", NominalThicknessMM: 8.56, RequiredThicknessMM: 4.19, ActualThicknessMM: 8.4, MeasuredYear: "2025"},
		{InspectionPoint: "IP-22", NominalThicknessMM: 8.56, RequiredThicknessMM: 4.19, ActualThicknessMM: 8.44, MeasuredYear: "2025"},
	}
	in.RiskInput.ReleaseFluid = "Gas"
	in.RiskInput.BuildingCountInsidePIR = 0
	in.RiskInput.ClassLocation = "remote"
	return in
}

func TestCalculatePipelineOilMatchesWorkbookSamples(t *testing.T) {
	result, errs := CalculatePipelineOil(samplePipelineOilInput())
	if len(errs) > 0 {
		t.Fatalf("unexpected validation errors: %+v", errs)
	}

	assertClose(t, result.PipeLengthFt, 32696.850393700788, 1e-9)
	assertClose(t, result.DesignTemperatureC, 93.333333333333343, 1e-9)
	assertClose(t, result.OutsideDiameterMM, 219.07499999999999, 1e-9)
	assertClose(t, result.RequiredThicknessIn, 0.17113095238095238, 1e-9)
	assertClose(t, result.RequiredThicknessMM, 4.34672619047619, 1e-9)
	assertClose(t, result.DesignPressureKgCM2, 70.30865499542993, 1e-9)
	assertClose(t, result.SMYSKgCM2, 2460.8029248400476, 1e-9)
	assertClose(t, result.MaterialStressKgCM2, 2460.8029248400476, 1e-9)
	assertClose(t, result.SummaryRequiredThicknessIn, 0.171, 1e-12)
	assertClose(t, result.SummaryRequiredThicknessMM, 4.3434, 1e-9)
	assertClose(t, result.PointResults[1].CorrosionRateMMYear, 0.2575, 1e-9)
	assertClose(t, result.PointResults[1].RemainingLifeYears, 6.912621359223301, 1e-9)
	assertClose(t, result.PointResults[1].HoopStressPsi, 17898.28431372549, 1e-9)
	assertClose(t, result.PointResults[1].MAOPPsi, 1407.9561793906196, 1e-9)
	assertClose(t, result.HighestHoopStressPsi, 17898.28431372549, 1e-9)
	assertClose(t, result.HighestHoopStressKgCM2, 1258.4042968238409, 1e-9)
	assertClose(t, result.HighestHoopStressPercentSMY, 51.13795518207282, 1e-9)
	assertClose(t, result.LowestMAOPPsi, 1407.9561793906196, 1e-9)
	assertClose(t, result.LowestMAOPKgCM2, 98.99150526545873, 1e-9)
	assertClose(t, result.ManagementSystemFactor, 1, 1e-12)
	assertClose(t, result.ThirdPartyDamageFactor, 1.0, 1e-12)
	assertClose(t, result.ExternalCorrosionFactor, 1.0, 1e-12)
	assertClose(t, result.InternalCorrosionFactor, 1.0, 1e-12)
	assertClose(t, result.GoverningDamageFactor, 1.0, 1e-12)
	assertClose(t, result.PoFValue, 0.00003, 1e-12)
	assertClose(t, result.CoFValue, 4617.8216379719998, 1e-9)
	assertClose(t, result.RiskValue, 10, 1e-9)
	if result.PoF != "2" || result.CoF != "E" || result.FinalRiskCode != "2E" || result.FinalRiskLevel != "Medium Risk" {
		t.Fatalf("unexpected pipeline risk ranking: pof=%s cof=%s code=%s level=%s", result.PoF, result.CoF, result.FinalRiskCode, result.FinalRiskLevel)
	}
	if result.RequiredThicknessStatus != "ACCEPTABLE" {
		t.Fatalf("expected aggregate thickness status to match workbook sample")
	}
	if result.GoverningDamageMechanism == "" {
		t.Fatalf("expected a governing damage mechanism, got empty")
	}
}

func TestCalculatePipelineGasMatchesAppraisalWorkbook(t *testing.T) {
	result, errs := CalculatePipelineOil(samplePipelineGasWorkbookInput())
	if len(errs) > 0 {
		t.Fatalf("unexpected validation errors: %+v", errs)
	}

	assertClose(t, result.PipeLengthFt, 32696.850393700788, 1e-9)
	assertClose(t, result.DesignTemperatureC, 93.333333333333343, 1e-9)
	assertClose(t, result.OutsideDiameterMM, 323.84999999999997, 1e-9)
	assertClose(t, result.RequiredThicknessIn, 0.27584134615384615, 1e-12)
	assertClose(t, result.RequiredThicknessMM, 7.006370192307692, 1e-9)
	assertClose(t, result.DesignPressureKgCM2, 94.916684243830417, 1e-9)
	assertClose(t, result.OperatingPressureKgCM2, 45.700625747029456, 1e-9)
	assertClose(t, result.SMYSKgCM2, 3656.0500597623563, 1e-9)
	assertClose(t, result.MaterialStressKgCM2, 3656.0500597623563, 1e-9)
	assertClose(t, result.SummaryRequiredThicknessMM, 7.006370192307692, 1e-9)
	assertClose(t, result.SummaryRequiredThicknessIn, 0.27584134615384615, 1e-12)
	assertClose(t, result.PointResults[0].CorrosionRateMMYear, 0.15571428571428569, 1e-12)
	assertClose(t, result.PointResults[1].CorrosionRateMMYear, 0.10428571428571434, 1e-12)
	assertClose(t, result.PointResults[0].RemainingLifeYears, 14.192660550458722, 1e-9)
	assertClose(t, result.PointResults[1].RemainingLifeYears, 20, 1e-9)
	assertClose(t, result.HighestHoopStressPsi, 23709.191973969628, 1e-9)
	assertClose(t, result.HighestHoopStressKgCM2, 1666.961398718247, 1e-9)
	assertClose(t, result.HighestHoopStressPercentSMY, 45.594599949941596, 1e-9)
	assertClose(t, result.LowestMAOPPsi, 1776.5261695229274, 1e-9)
	assertClose(t, result.LowestMAOPKgCM2, 124.90516554334017, 1e-9)
	if result.RequiredThicknessStatus != "ACCEPTABLE" || result.HoopStressStatus != "ACCEPTABLE" || result.MAOPStatus != "ACCEPTABLE" {
		t.Fatalf("expected gas appraisal statuses to match workbook, got thickness=%s hoop=%s maop=%s", result.RequiredThicknessStatus, result.HoopStressStatus, result.MAOPStatus)
	}
}

func TestCalculateRawGasPipingUsesASMEB313AppraisalFormula(t *testing.T) {
	result, errs := CalculatePipelineOil(sampleRawGasPipingWorkbookInput())
	if len(errs) > 0 {
		t.Fatalf("unexpected validation errors: %+v", errs)
	}

	assertClose(t, result.PipeLengthFt, 918.63517060367451, 1e-9)
	assertClose(t, result.OutsideDiameterMM, 114.3, 1e-12)
	assertClose(t, result.RequiredThicknessIn, 0.1659549689440994, 1e-12)
	assertClose(t, result.RequiredThicknessMM, 4.215256211180124, 1e-12)
	assertClose(t, result.DesignPressureKgCM2, 106.86915559305351, 1e-9)
	assertClose(t, result.OperatingPressureKgCM2, 42.185192997257957, 1e-9)
	assertClose(t, result.SMYSKgCM2, 2650.6362933277087, 1e-9)
	assertClose(t, result.PointResults[0].CorrosionRateMMYear, 0.052857142857143, 1e-12)
	assertClose(t, result.PointResults[0].RemainingLifeYears, 20, 1e-9)
	assertClose(t, result.HighestHoopStressPsi, 10606.593406593407, 1e-9)
	assertClose(t, result.HighestHoopStressKgCM2, 745.73531650097777, 1e-9)
	assertClose(t, result.HighestHoopStressPercentSMY, 28.134200017489142, 1e-9)
	assertClose(t, result.LowestMAOPPsi, 3040.427664550618, 1e-9)
	assertClose(t, result.LowestMAOPKgCM2, 213.76837970545017, 1e-9)
	if result.RequiredThicknessStatus != "ACCEPTABLE" || result.HoopStressStatus != "ACCEPTABLE" || result.MAOPStatus != "ACCEPTABLE" {
		t.Fatalf("expected raw gas appraisal statuses to be acceptable, got thickness=%s hoop=%s maop=%s", result.RequiredThicknessStatus, result.HoopStressStatus, result.MAOPStatus)
	}
}

func TestCalculatePipelineGasCoFUsesPIRAndBuildings(t *testing.T) {
	in := samplePipelineOilInput()
	in.Service = "Gas"
	in.RiskInput.BuildingCountInsidePIR = 25
	in.RiskInput.ClassLocation = "urban_dense"

	result, errs := CalculatePipelineOil(in)
	if len(errs) > 0 {
		t.Fatalf("unexpected validation errors: %+v", errs)
	}

	assertClose(t, result.PIRFeet, 25.002333817628404, 1e-12)
	if result.CoF != "E" {
		t.Fatalf("expected CoF=E, got %s", result.CoF)
	}
}

func TestPipelineOilValidation(t *testing.T) {
	tests := []struct {
		name  string
		input func() PipelineOilInput
	}{
		{"zero thickness", func() PipelineOilInput {
			in := samplePipelineOilInput()
			in.InspectionPoints[0].ActualThicknessMM = 0
			return in
		}},
		{"negative corrosion rate input", func() PipelineOilInput {
			in := samplePipelineOilInput()
			in.InspectionPoints[0].ActualThicknessMM = 9
			return in
		}},
		{"missing required data", func() PipelineOilInput {
			in := samplePipelineOilInput()
			in.ReportNo = ""
			return in
		}},
		{"invalid service unit", func() PipelineOilInput {
			in := samplePipelineOilInput()
			in.Service = "Steam"
			return in
		}},
		{"divide by zero date", func() PipelineOilInput {
			in := samplePipelineOilInput()
			in.InspectionPoints[0].MeasuredYear = "2017"
			return in
		}},
		{"extreme pressure", func() PipelineOilInput {
			in := samplePipelineOilInput()
			in.InternalDesignPressurePsi = 25000
			return in
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, errs := CalculatePipelineOil(tt.input()); len(errs) == 0 {
				t.Fatalf("expected validation errors")
			}
		})
	}
}

func TestPipelineInspectionResultIsCalculatedOutputNotDraftInput(t *testing.T) {
	in := samplePipelineOilInput()
	errs := ValidatePipelineOilDraft(in)
	if len(errs) > 0 {
		t.Fatalf("inspection result should not be required in draft input: %+v", errs)
	}

	result, calcErrs := CalculatePipelineOil(in)
	if len(calcErrs) > 0 {
		t.Fatalf("unexpected validation errors: %+v", calcErrs)
	}
	if result.InspectionResult != "ACCEPTABLE" {
		t.Fatalf("expected calculated inspection result, got %q", result.InspectionResult)
	}
}

func TestPipelineDamageMechanismSelectionValidation(t *testing.T) {
	in := samplePipelineOilInput()
	in.RiskInput.DamageMechanism = "localized_corrosion"
	if errs := ValidatePipelineOilDraft(in); len(errs) > 0 {
		t.Fatalf("expected selected pipeline damage mechanism to validate: %+v", errs)
	}
	in.RiskInput.DamageMechanism = "mechanical_damage"
	if errs := ValidatePipelineOilDraft(in); len(errs) > 0 {
		t.Fatalf("expected legacy pipeline damage mechanism to normalize: %+v", errs)
	}
	if NormalizePipelineDamageMechanism(in.RiskInput.DamageMechanism) != "third_party_mechanical_damage" {
		t.Fatalf("expected legacy mechanical damage to map into grouped pipeline mechanism")
	}

	in.RiskInput.DamageMechanism = "pressure-vessel-only"
	if errs := ValidatePipelineOilDraft(in); len(errs) == 0 {
		t.Fatalf("expected invalid pipeline damage mechanism to fail validation")
	}
}

func TestPipelineCalculatesAllDamageMechanisms(t *testing.T) {
	in := samplePipelineOilInput()
	in.RiskInput.InspectionEffectivityByDM = map[string]string{
		"external_corrosion": "High",
		"internal_corrosion": "Low",
	}
	in.RiskInput.InspectionPlanByDM = map[string]PipelineInspectionPlanInput{
		"internal_corrosion": {
			NonIntrusiveMethod: "Wall Thickness measurement by UT",
		},
	}
	result, errs := CalculatePipelineOil(in)
	if len(errs) > 0 {
		t.Fatalf("unexpected validation errors: %+v", errs)
	}
	if len(result.DamageMechanismResults) != len(PipelineDamageMechanismOptions()) {
		t.Fatalf("expected all pipeline mechanisms to be calculated, got %d", len(result.DamageMechanismResults))
	}
	foundInternal := false
	for _, item := range result.DamageMechanismResults {
		if item.Severity == "" {
			t.Fatalf("expected severity for %s", item.Code)
		}
		if item.Code == "internal_corrosion" {
			foundInternal = true
			if item.InspectionEffectivity != "Low" {
				t.Fatalf("expected per-mechanism inspection effectivity, got %q", item.InspectionEffectivity)
			}
		}
	}
	if !foundInternal {
		t.Fatalf("expected internal corrosion screening result")
	}
	if len(result.InspectionPlanResults) != len(PipelineDamageMechanismOptions()) {
		t.Fatalf("expected inspection plan for all mechanisms, got %d", len(result.InspectionPlanResults))
	}
	for _, item := range result.InspectionPlanResults {
		if item.Code == "internal_corrosion" {
			if item.NonIntrusiveMethod != "Wall Thickness measurement by UT" || item.NonIntrusiveIntervalMonths <= 0 {
				t.Fatalf("expected internal corrosion inspection plan to be stored and interval calculated: %+v", item)
			}
		}
	}
}

func TestPipelineDamageMechanismsCarryStandardsMetadata(t *testing.T) {
	result, errs := CalculatePipelineOil(samplePipelineOilInput())
	if len(errs) > 0 {
		t.Fatalf("unexpected validation errors: %+v", errs)
	}

	for _, item := range result.DamageMechanismResults {
		if item.SourceStandard == "" || item.ConfidenceLevel == "" || item.RuleStatus == "" {
			t.Fatalf("expected standards metadata for %s: %+v", item.Code, item)
		}
		if item.RuleStatus != PipelineRuleVerified &&
			item.RuleStatus != PipelineRulePartiallyVerified &&
			item.RuleStatus != PipelineRuleTODOEngineeringConfirmation {
			t.Fatalf("unexpected rule status for %s: %q", item.Code, item.RuleStatus)
		}
	}
}

func TestPipelineInspectionMethodSelectionUpdatesEffectivityAndInterval(t *testing.T) {
	none := samplePipelineOilInput()
	none.RiskInput.InspectionEffectivityByDM = nil
	none.RiskInput.InspectionPlanByDM = map[string]PipelineInspectionPlanInput{
		"internal_corrosion": {NonIntrusiveMethod: "None"},
	}
	noneResult, errs := CalculatePipelineOil(none)
	if len(errs) > 0 {
		t.Fatalf("unexpected validation errors for none method: %+v", errs)
	}

	high := samplePipelineOilInput()
	high.RiskInput.InspectionEffectivityByDM = nil
	high.RiskInput.InspectionPlanByDM = map[string]PipelineInspectionPlanInput{
		"internal_corrosion": {NonIntrusiveMethod: "VIE + Wall Thickness measurement by UT"},
	}
	highResult, errs := CalculatePipelineOil(high)
	if len(errs) > 0 {
		t.Fatalf("unexpected validation errors for high method: %+v", errs)
	}

	noneDM := pipelineDMResultByCode(t, noneResult, "internal_corrosion")
	highDM := pipelineDMResultByCode(t, highResult, "internal_corrosion")
	if noneDM.InspectionEffectivity != "None" {
		t.Fatalf("expected selected None method to set DM effectivity None, got %q", noneDM.InspectionEffectivity)
	}
	if highDM.InspectionEffectivity != "High" {
		t.Fatalf("expected selected high method to set DM effectivity High, got %q", highDM.InspectionEffectivity)
	}

	nonePlan := pipelinePlanResultByCode(t, noneResult, "internal_corrosion")
	highPlan := pipelinePlanResultByCode(t, highResult, "internal_corrosion")
	if nonePlan.NonIntrusiveEffectivity != "None" || highPlan.NonIntrusiveEffectivity != "High" {
		t.Fatalf("expected plan effectivity to follow selected methods, got none=%q high=%q", nonePlan.NonIntrusiveEffectivity, highPlan.NonIntrusiveEffectivity)
	}
	if nonePlan.NonIntrusiveIntervalMonths == highPlan.NonIntrusiveIntervalMonths {
		t.Fatalf("expected inspection interval to change when method effectivity changes: none=%d high=%d", nonePlan.NonIntrusiveIntervalMonths, highPlan.NonIntrusiveIntervalMonths)
	}
}

func TestPipelineNotSeveritySuppressesInspectionEffectivity(t *testing.T) {
	in := samplePipelineOilInput()
	in.RiskInput.InspectionEffectivityByDM = map[string]string{"chemical_damage": "High"}
	in.RiskInput.InspectionPlanByDM = map[string]PipelineInspectionPlanInput{
		"chemical_damage": {
			NonIntrusiveMethod: "VIE + Wall Thickness measurement by UT",
			IntrusiveMethod:    "Direct examination",
		},
	}

	result, errs := CalculatePipelineOil(in)
	if len(errs) > 0 {
		t.Fatalf("unexpected validation errors: %+v", errs)
	}

	dm := pipelineDMResultByCode(t, result, "chemical_damage")
	if dm.Severity != "NOT" {
		t.Fatalf("expected chemical damage severity NOT, got %q", dm.Severity)
	}
	if dm.InspectionEffectivity != "" {
		t.Fatalf("expected NOT severity DM effectivity to be hidden, got %q", dm.InspectionEffectivity)
	}

	plan := pipelinePlanResultByCode(t, result, "chemical_damage")
	if plan.NonIntrusiveEffectivity != "" || plan.IntrusiveEffectivity != "" {
		t.Fatalf("expected NOT severity plan effectivity to be hidden, got non=%q intrusive=%q", plan.NonIntrusiveEffectivity, plan.IntrusiveEffectivity)
	}
}

func TestPipelineInspectionEffectivityDoesNotChangeDMSeverity(t *testing.T) {
	low := samplePipelineOilInput()
	low.RiskInput.InspectionEffectivityByDM = map[string]string{"internal_corrosion": "Low"}
	high := samplePipelineOilInput()
	high.RiskInput.InspectionEffectivityByDM = map[string]string{"internal_corrosion": "High"}

	lowResult, errs := CalculatePipelineOil(low)
	if len(errs) > 0 {
		t.Fatalf("unexpected validation errors for low effectivity: %+v", errs)
	}
	highResult, errs := CalculatePipelineOil(high)
	if len(errs) > 0 {
		t.Fatalf("unexpected validation errors for high effectivity: %+v", errs)
	}

	for _, lowDM := range lowResult.DamageMechanismResults {
		highDM := pipelineDMResultByCode(t, highResult, lowDM.Code)
		if lowDM.Score != highDM.Score || lowDM.Severity != highDM.Severity {
			t.Fatalf("inspection effectivity changed %s severity/score: low=%s %.2f high=%s %.2f", lowDM.Code, lowDM.Severity, lowDM.Score, highDM.Severity, highDM.Score)
		}
	}
}

func TestPipelineDMModifiersAreFixedAtOneAndCustomModifiersDisabled(t *testing.T) {
	if pipelineDMModifierDefault != 1.0 {
		t.Fatalf("expected pipeline DM modifier default 1.0, got %.2f", pipelineDMModifierDefault)
	}
	if customPipelineDMModifiers {
		t.Fatalf("custom pipeline DM modifiers must be disabled")
	}
	if err := ValidatePipelineDMModifierConfiguration(); err != nil {
		t.Fatalf("expected all active pipeline DM modifier values to be 1.0: %v", err)
	}
	if got := adjustedPipelineDMScore(2.75); got != 2.75 {
		t.Fatalf("expected adjusted score to equal base score with 1.0 modifier, got %.2f", got)
	}
}

func TestPipelineAndPressureVesselUseSharedRiskThresholds(t *testing.T) {
	tests := []struct {
		score       int
		pofCategory string
		cofCategory string
	}{
		{1, "1", "A"},
		{5, "1", "E"},
		{6, "2", "C"},
		{10, "2", "E"},
		{12, "3", "D"},
		{15, "3", "E"},
		{16, "4", "D"},
		{25, "5", "E"},
	}
	for _, tt := range tests {
		pvLevel := ApprovedRiskLevelFromMatrixScore(tt.score)
		_, pipelineLevel := calculatePipelineRiskRanking(tt.pofCategory, tt.cofCategory)
		if pipelineLevel != pvLevel+" Risk" && !(pvLevel == "Extreme" && pipelineLevel == "Critical Risk") {
			t.Fatalf("shared risk level mismatch for score %d: pv=%s pipeline=%s", tt.score, pvLevel, pipelineLevel)
		}
	}
}

func TestPipelineAndPressureVesselUseSharedCoFThresholds(t *testing.T) {
	for _, value := range []float64{1, 2, 3, 4, 5} {
		cof := ApprovedCoFCategoryFromIndex(value)
		if got := ApprovedCoFNumeric(cof); got != value {
			t.Fatalf("expected CoF value %.0f to round-trip through shared category, got %s %.0f", value, cof, got)
		}
	}
}

func TestPipelineMaterialStressUsesPipelineSpecificDataset(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	migrations.PipelineMaterialTables(db)
	migrations.PipelineOilMasterDataTables(db)
	if err := seeder.PipelineMaterial(db); err != nil {
		t.Fatalf("seed pipeline material dataset: %v", err)
	}

	var source, version, code, edition string
	if err := db.QueryRow(`SELECT stress_source, stress_dataset_version, governing_code, code_edition FROM pipeline_materials WHERE name = 'API 5L Gr B'`).Scan(&source, &version, &code, &edition); err != nil {
		t.Fatalf("read seeded pipeline material metadata: %v", err)
	}
	if source != seeder.PipelineMaterialStressDatasetSource || version != seeder.PipelineMaterialStressDatasetVersion || code == "" || edition == "" {
		t.Fatalf("unexpected pipeline material dataset metadata: source=%q version=%q code=%q edition=%q", source, version, code, edition)
	}
}

func TestPipelineMaterialStressNeverFallsBackToPressureVesselData(t *testing.T) {
	in := sampleRawGasPipingWorkbookInput()
	in.MaterialStressPsi = 0
	if got := derivePipelineMaterialStressPsi(in); got != 0 {
		t.Fatalf("expected no B31.3 fallback material stress, got %.2f", got)
	}
	errs := ValidatePipelineOilCalculation(in)
	found := false
	for _, err := range errs {
		if err.Field == "material_stress_psi" && strings.Contains(err.Message, "ENGINEERING_REVIEW_REQUIRED") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected unsupported B31.3 material mapping to require engineering review, got %+v", errs)
	}
}

func TestSharedThresholdBoundaryValues(t *testing.T) {
	for _, tt := range []struct {
		value float64
		want  string
	}{
		{0.000009999, "1"},
		{0.00001, "2"},
		{0.0001, "3"},
		{0.001, "4"},
		{0.01, "5"},
	} {
		if got := ApprovedPoFCategory(tt.value); got != tt.want {
			t.Fatalf("PoF boundary %.8f got %s want %s", tt.value, got, tt.want)
		}
	}
	for _, tt := range []struct {
		score int
		want  string
	}{
		{5, "Low"},
		{6, "Medium"},
		{10, "Medium"},
		{11, "High"},
		{15, "High"},
		{16, "Extreme"},
	} {
		if got := ApprovedRiskLevelFromMatrixScore(tt.score); got != tt.want {
			t.Fatalf("risk boundary %d got %s want %s", tt.score, got, tt.want)
		}
	}
}

func TestResolvedEngineeringConfirmationItemsRemovedFromGoLiveAudit(t *testing.T) {
	body, err := os.ReadFile("../docs/pipeline-go-live-calculation-audit.md")
	if err != nil {
		t.Fatalf("read audit doc: %v", err)
	}
	text := string(body)
	for _, phrase := range []string{
		"whether inspection effectiveness should alter Pipeline DM severity scores",
		"Pipeline DM modifier magnitudes",
		"licensed API 581 tables/risk thresholds",
		"detailed CoF thresholds",
		"approved B31.3 material stress table values",
	} {
		if strings.Contains(strings.ToLower(text), strings.ToLower(phrase)) {
			t.Fatalf("resolved engineering confirmation phrase still present in audit doc: %s", phrase)
		}
	}
}

func TestPipelineRequiredThicknessPopulatesResultPointRowsAndSkipsEmptyRows(t *testing.T) {
	in := samplePipelineOilInput()
	in.InspectionPoints = append(in.InspectionPoints, PipelineOilInspectionPoint{})
	in.InspectionPoints[0].RequiredThicknessMM = 0
	in.InspectionPoints[0].LocationClass = "Class 2"
	in.InspectionPoints[0].InstallationType = "Above Ground"

	result, errs := CalculatePipelineOil(in)
	if len(errs) > 0 {
		t.Fatalf("unexpected validation errors: %+v", errs)
	}
	if len(result.PointResults) != 3 {
		t.Fatalf("expected empty inspection point row to be skipped, got %d point results", len(result.PointResults))
	}
	first := result.PointResults[0]
	if first.RequiredThicknessMM <= 0 {
		t.Fatalf("expected authoritative result point required WT to be calculated, got %.6f", first.RequiredThicknessMM)
	}
	if first.LocationClass != "Class 2" || first.InstallationType != "Above Ground" || first.MeasuredYear != "2025" {
		t.Fatalf("expected point metadata to be carried to result row, got %+v", first)
	}
}

func pipelineDMResultByCode(t *testing.T, result *PipelineOilResult, code string) PipelineDamageMechanismResult {
	t.Helper()
	for _, item := range result.DamageMechanismResults {
		if item.Code == code {
			return item
		}
	}
	t.Fatalf("damage mechanism %s not found", code)
	return PipelineDamageMechanismResult{}
}

func pipelinePlanResultByCode(t *testing.T, result *PipelineOilResult, code string) PipelineInspectionPlanResult {
	t.Helper()
	for _, item := range result.InspectionPlanResults {
		if item.Code == code {
			return item
		}
	}
	t.Fatalf("inspection plan %s not found", code)
	return PipelineInspectionPlanResult{}
}

func TestPipelineFormulaTraceCarriesStandardsMetadata(t *testing.T) {
	result, errs := CalculatePipelineOil(samplePipelineOilInput())
	if len(errs) > 0 {
		t.Fatalf("unexpected validation errors: %+v", errs)
	}
	if len(result.FormulaTrace) == 0 {
		t.Fatalf("expected formula trace entries")
	}

	foundRequiredThickness := false
	foundPipelinePoF := false
	for _, item := range result.FormulaTrace {
		if item.SourceStandard == "" || item.ConfidenceLevel == "" || item.RuleStatus == "" {
			t.Fatalf("expected formula trace metadata for %s: %+v", item.FormulaName, item)
		}
		if item.FormulaName == "required_thickness" {
			foundRequiredThickness = true
			if item.RuleStatus != PipelineRuleVerified {
				t.Fatalf("expected required thickness to be verified, got %q", item.RuleStatus)
			}
		}
		if item.FormulaName == "pipeline_pof" {
			foundPipelinePoF = true
			if item.RuleStatus != PipelineRulePartiallyVerified {
				t.Fatalf("expected pipeline PoF to be partially verified, got %q", item.RuleStatus)
			}
		}
	}
	if !foundRequiredThickness || !foundPipelinePoF {
		t.Fatalf("expected required_thickness and pipeline_pof trace entries")
	}
}

func TestPipelineAssessmentVersionsAndAuditTrail(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	migrations.PipelineOilAssessmentTables(db)
	migrations.PipelineOilReportingTables(db)

	service := NewPipelineOilService(db)
	input := samplePipelineOilInput()
	id, err := service.CreateDraftAssessment(input)
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}
	if _, err = service.CalculateAssessment(id, input); err != nil {
		t.Fatalf("calculate assessment: %v", err)
	}
	if err = service.RecordApproval(id, "Lead Engineer", "Reviewed for issue"); err != nil {
		t.Fatalf("record approval: %v", err)
	}
	if err = service.RecordExport(id, "Lead Engineer", "EXCEL"); err != nil {
		t.Fatalf("record export: %v", err)
	}

	versions, err := service.GetAssessmentVersions(id)
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected create and calculation versions, got %d", len(versions))
	}
	if versions[0].VersionNumber != 2 || versions[0].Result == nil {
		t.Fatalf("expected latest version to contain calculated result: %+v", versions[0])
	}

	events, err := service.GetAssessmentAuditEvents(id)
	if err != nil {
		t.Fatalf("list audit events: %v", err)
	}
	if len(events) < 4 {
		t.Fatalf("expected create, recalculated, approval, and export events, got %d", len(events))
	}
	seen := map[string]bool{}
	for _, event := range events {
		seen[event.Action] = true
	}
	for _, action := range []string{"CREATED", "RECALCULATED", "APPROVED", "EXPORTED_EXCEL"} {
		if !seen[action] {
			t.Fatalf("expected audit action %s in %+v", action, events)
		}
	}
}

func TestPipelineComparisonUsesLatestTwoVersions(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	migrations.PipelineOilAssessmentTables(db)
	migrations.PipelineOilReportingTables(db)

	service := NewPipelineOilService(db)
	input := samplePipelineOilInput()
	id, err := service.CreateDraftAssessment(input)
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}
	changed := input
	changed.OperatingPressurePsi = 25
	if _, err = service.CalculateAssessment(id, changed); err != nil {
		t.Fatalf("calculate changed assessment: %v", err)
	}

	comparison, err := service.GetAssessmentComparison(id)
	if err != nil {
		t.Fatalf("compare versions: %v", err)
	}
	if comparison.CurrentVersion == nil || comparison.PreviousVersion == nil {
		t.Fatalf("expected latest two versions in comparison")
	}
	foundPressureChange := false
	for _, change := range comparison.Changes {
		if change.Field == "input.operating_pressure_psi" {
			foundPressureChange = true
			break
		}
	}
	if !foundPressureChange {
		t.Fatalf("expected operating pressure change in comparison")
	}
}

func TestPipelineRecommendationCarriesSystemAdvisorySource(t *testing.T) {
	result, errs := CalculatePipelineOil(samplePipelineOilInput())
	if len(errs) > 0 {
		t.Fatalf("unexpected validation errors: %+v", errs)
	}
	if result.RecommendationSource != "System advisory rule based on risk category, CoF factors, and governing pipeline DM score driver." {
		t.Fatalf("unexpected recommendation source: %q", result.RecommendationSource)
	}
	if result.RecommendationRuleName != "pipeline-system-advisory-v2" {
		t.Fatalf("unexpected recommendation rule: %q", result.RecommendationRuleName)
	}
	if result.Recommendation == "" {
		t.Fatalf("expected non-empty recommendation summary")
	}
	if result.GoverningDamageMechanism == "" {
		t.Fatalf("expected a governing damage mechanism, got empty")
	}
}

func TestPipelineManualRecommendationSectionsAreSavedEditedAndNotOverwritten(t *testing.T) {
	db := newPipelineTestDB(t)
	service := NewPipelineOilService(db)
	input := samplePipelineOilInput()
	input.RecommendationImmediateActions = "Verify CP\nAssign owner"
	input.RecommendationInspectionMonitoring = "Run UT survey\nReview CIPS"
	input.RecommendationLongTermMitigation = "Install permanent sleeve"
	input.RiskInput.EngineeringNotes = "Line one\nLine two"

	id, err := service.CreateDraftAssessment(input)
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}
	saved, err := service.GetAssessmentDetail(id)
	if err != nil {
		t.Fatalf("get draft: %v", err)
	}
	if saved.Input.RecommendationImmediateActions != input.RecommendationImmediateActions {
		t.Fatalf("saved immediate actions mismatch: %q", saved.Input.RecommendationImmediateActions)
	}

	input.RecommendationImmediateActions = "Edited immediate\nKeep line break"
	if err = service.UpdateDraftAssessment(id, input); err != nil {
		t.Fatalf("edit draft: %v", err)
	}
	result, err := service.CalculateAssessment(id, input)
	if err != nil {
		t.Fatalf("calculate assessment: %v", err)
	}
	if got := strings.Join(result.RecommendationGroups.ImmediateActions, "\n"); got != input.RecommendationImmediateActions {
		t.Fatalf("manual immediate actions overwritten: %q", got)
	}
	if got := strings.Join(result.RecommendationGroups.InspectionMonitor, "\n"); got != input.RecommendationInspectionMonitoring {
		t.Fatalf("manual inspection monitoring overwritten: %q", got)
	}
	if got := strings.Join(result.RecommendationGroups.LongTermMitigation, "\n"); got != input.RecommendationLongTermMitigation {
		t.Fatalf("manual long-term mitigation overwritten: %q", got)
	}
	if result.RecommendationConfidence == "" || result.RecommendationSource == "" || result.RecommendationRuleName == "" {
		t.Fatalf("expected recommendation metadata to remain available: %+v", result)
	}
	if strings.Contains(result.Recommendation, "Edited immediate") || strings.Contains(result.Recommendation, "Run UT survey") || strings.Contains(result.Recommendation, "Install permanent sleeve") {
		t.Fatalf("expected advisory text to remain system-generated, got %q", result.Recommendation)
	}
	if !strings.Contains(result.Recommendation, "Keep the formula trace") {
		t.Fatalf("expected system-generated advisory text, got %q", result.Recommendation)
	}

	recalculated := input
	recalculated.OperatingPressurePsi = input.OperatingPressurePsi + 10
	result, err = service.CalculateAssessment(id, recalculated)
	if err != nil {
		t.Fatalf("recalculate assessment: %v", err)
	}
	if got := strings.Join(result.RecommendationGroups.ImmediateActions, "\n"); got != input.RecommendationImmediateActions {
		t.Fatalf("recalculation overwrote manual recommendation: %q", got)
	}
}

func TestPipelineManualRecommendationLegacyJSONStillDeserializes(t *testing.T) {
	raw := []byte(`{"report_no":"legacy","manual_recommendation":"Legacy note\nsecond line","risk_input":{"engineering_notes":"old notes"}}`)
	var input PipelineOilInput
	if err := json.Unmarshal(raw, &input); err != nil {
		t.Fatalf("legacy unmarshal failed: %v", err)
	}
	if input.ManualRecommendation != "Legacy note\nsecond line" {
		t.Fatalf("legacy manual recommendation lost: %q", input.ManualRecommendation)
	}
	result := &PipelineOilResult{
		Recommendation:       input.ManualRecommendation,
		RecommendationSource: "User overridden recommendation.",
	}
	applyPipelineRecommendationCompatibility(&input, result)
	if got := strings.Join(result.RecommendationGroups.ImmediateActions, "\n"); got != "Legacy manual recommendation: Legacy note\nsecond line" {
		t.Fatalf("legacy recommendation not exposed as note: %q", got)
	}
}

func TestPipelineManualRecommendationIsNotRequired(t *testing.T) {
	input := samplePipelineOilInput()
	input.ManualRecommendation = ""
	input.RecommendationImmediateActions = ""
	input.RecommendationInspectionMonitoring = ""
	input.RecommendationLongTermMitigation = ""
	if errs := ValidatePipelineOilDraft(input); len(errs) > 0 {
		t.Fatalf("manual recommendation should not be required: %+v", errs)
	}
}

func TestPipelineH2SPartialPressureUsesPPM(t *testing.T) {
	got := calculateH2SPartialPressure(1000, 650)
	want := 0.65
	if math.Abs(got-want) > 0.000001 {
		t.Fatalf("expected pH2S from ppm units %.6f, got %.6f", want, got)
	}
}

func TestConfirmedPipelineModifierMapsRemainNeutral(t *testing.T) {
	modifierMaps := []struct {
		name string
		m    map[string]float64
	}{
		{"pipelineDepthFactors", pipelineDepthFactors},
		{"pipelinePatrolFactors", pipelinePatrolFactors},
		{"pipelineROWFactors", pipelineROWFactors},
		{"pipelineSoilFactors", pipelineSoilFactors},
		{"pipelineCoatingConditionFactors", pipelineCoatingConditionFactors},
		{"pipelineCPFactors", pipelineCPFactors},
		{"pipelineClassLocationFactors", pipelineClassLocationFactors},
		{"pipelineFluidCorrosivityMPYFactors", pipelineFluidCorrosivityMPYFactors},
		{"pipelinePHSeverity", pipelinePHSeverity},
		{"pipelineInhibitorModifiers", pipelineInhibitorModifiers},
		{"pipelineCoatingDamageModifiers", pipelineCoatingDamageModifiers},
		{"pipelineInsulationDamageModifiers", pipelineInsulationDamageModifiers},
		{"pipelineConfidenceWeight", pipelineConfidenceWeight},
		{"pipelineWeldCrackingModifiers", pipelineWeldCrackingModifiers},
		{"pipelinePWHTModifiers", pipelinePWHTModifiers},
		{"pipelineOneCallModifiers", pipelineOneCallModifiers},
		{"pipelineH2SPpmSeverity", pipelineH2SPpmSeverity},
		{"pipelineFlowVelocityModifiers", pipelineFlowVelocityModifiers},
		{"pipelineSolidContentModifiers", pipelineSolidContentModifiers},
		{"pipelineExtCrackingOptions", pipelineExtCrackingOptions},
	}

	for _, pm := range modifierMaps {
		for k, v := range pm.m {
			if v != 1.0 {
				t.Fatalf("%s[%q] = %.4f, want fixed neutral 1.0", pm.name, k, v)
			}
		}
	}
}

func assertClose(t *testing.T, actual, expected, tolerance float64) {
	t.Helper()
	if math.Abs(actual-expected) > tolerance {
		t.Fatalf("expected %.12f got %.12f", expected, actual)
	}
}
