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
		InspectionPoints: []PipelineOilInspectionPoint{
			{InspectionPoint: "IP-82", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 7.98, MeasuredYear: 2025},
			{InspectionPoint: "IP-8 A", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 6.12, MeasuredYear: 2025},
			{InspectionPoint: "IP-8 C", NominalThicknessMM: 8.18, RequiredThicknessMM: 4.34, ActualThicknessMM: 6.21, MeasuredYear: 2025},
		},
	}
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
	if result.RequiredThicknessStatus != "ACCEPTABLE" {
		t.Fatalf("expected aggregate thickness status to match workbook sample")
	}
	if result.PoF != "TODO_ENGINEERING_CONFIRMATION" {
		t.Fatalf("PoF must remain a TODO because workbook has no RBI formula")
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
			in.Service = "Gas"
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
