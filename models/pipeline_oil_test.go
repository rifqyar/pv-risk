package models

import (
	"math"
	"testing"
)

func samplePipelineOilInput() PipelineOilInput {
	return PipelineOilInput{
		ReportNo:                  "PR 2057/PL-MIT/13/2025",
		LineIdentification:        "NPS 8 Main Oil Trunkline",
		YearBuilt:                 2017,
		YearUsed:                  2017,
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
			{InspectionPoint: "IP-82", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 7.98, MeasuredYear: 2025},
			{InspectionPoint: "IP-8 A", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 6.12, MeasuredYear: 2025},
			{InspectionPoint: "IP-8 C", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 6.21, MeasuredYear: 2025},
		},
	}
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
		{InspectionPoint: "IP-2", NominalThicknessMM: 10.31, RequiredThicknessMM: 7.01, ActualThicknessMM: 9.22, MeasuredYear: 2024},
		{InspectionPoint: "IP-11", NominalThicknessMM: 10.31, RequiredThicknessMM: 7.01, ActualThicknessMM: 9.58, MeasuredYear: 2024},
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
	in.YearUsed = 2018
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
		{InspectionPoint: "IP-2", NominalThicknessMM: 8.56, RequiredThicknessMM: 4.19, ActualThicknessMM: 8.19, MeasuredYear: 2025},
		{InspectionPoint: "IP-15", NominalThicknessMM: 8.56, RequiredThicknessMM: 4.19, ActualThicknessMM: 8.4, MeasuredYear: 2025},
		{InspectionPoint: "IP-22", NominalThicknessMM: 8.56, RequiredThicknessMM: 4.19, ActualThicknessMM: 8.44, MeasuredYear: 2025},
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
	assertClose(t, result.MaterialStressKgCM2, 1406.1730999085987, 1e-9)
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
	assertClose(t, result.ThirdPartyDamageFactor, 0.5787037037037037, 1e-12)
	assertClose(t, result.ExternalCorrosionFactor, 3.24, 1e-12)
	assertClose(t, result.InternalCorrosionFactor, 13.0321, 1e-12)
	assertClose(t, result.GoverningDamageFactor, 13.0321, 1e-12)
	assertClose(t, result.PoFValue, 0.000390963, 1e-12)
	assertClose(t, result.CoFValue, 4617.82163797161, 1e-9)
	assertClose(t, result.RiskValue, 15, 1e-9)
	if result.PoF != "3" || result.CoF != "E" || result.FinalRiskCode != "3E" || result.FinalRiskLevel != "High Risk" {
		t.Fatalf("unexpected pipeline risk ranking: pof=%s cof=%s code=%s level=%s", result.PoF, result.CoF, result.FinalRiskCode, result.FinalRiskLevel)
	}
	if result.RequiredThicknessStatus != "ACCEPTABLE" {
		t.Fatalf("expected aggregate thickness status to match workbook sample")
	}
	if result.GoverningDamageMechanism != "Internal Corrosion" {
		t.Fatalf("expected internal corrosion to govern")
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
	assertClose(t, result.MaterialStressKgCM2, 1406.1730999085987, 1e-9)
	assertClose(t, result.SummaryRequiredThicknessMM, 7.006370192307692, 1e-9)
	assertClose(t, result.SummaryRequiredThicknessIn, 0.27584134615384615, 1e-12)
	assertClose(t, result.PointResults[0].CorrosionRateMMYear, 0.15571428571428569, 1e-12)
	assertClose(t, result.PointResults[1].CorrosionRateMMYear, 0.10428571428571434, 1e-12)
	assertClose(t, result.PointResults[0].RemainingLifeYears, 14.192660550458722, 1e-9)
	assertClose(t, result.PointResults[1].RemainingLifeYears, 24.643835616438345, 1e-9)
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
	assertClose(t, result.PointResults[0].RemainingLifeYears, 75.67567567567545, 1e-9)
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
	if result.CoF != "E" || result.FinalRiskCode != "3E" || result.FinalRiskLevel != "High Risk" {
		t.Fatalf("unexpected gas risk ranking: cof=%s code=%s level=%s", result.CoF, result.FinalRiskCode, result.FinalRiskLevel)
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
			in.InspectionPoints[0].MeasuredYear = 2017
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

func assertClose(t *testing.T, actual, expected, tolerance float64) {
	t.Helper()
	if math.Abs(actual-expected) > tolerance {
		t.Fatalf("expected %.12f got %.12f", expected, actual)
	}
}

