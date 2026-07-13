package controller

import (
	"strings"
	"testing"

	"pv-risk/models"
)

func TestPipelineExcelExportIncludesRecommendationMetadataAndEngineeringNotes(t *testing.T) {
	assessment := &models.PipelineOilAssessment{
		Input: models.PipelineOilInput{
			ReportNo:           "PIPE-1",
			LineIdentification: "Line A",
			Location:           "Field",
			Service:            "Oil",
			ApplicableCode:     "ASME B31.4",
			RiskInput: models.PipelineOilRiskInput{
				EngineeringNotes: "First note\nSecond note",
			},
		},
		Status:         models.PipelineOilStatusCalculated,
		FormulaVersion: models.PipelineOilFormulaVersion,
		Result: &models.PipelineOilResult{
			PoF:                      "1",
			CoF:                      "A",
			FinalRiskCode:            "1A",
			FinalRiskLevel:           "Low Risk",
			GoverningDamageMechanism: "Internal Corrosion",
			RecommendationSource:     "Engineer-entered recommendation content; metadata describes advisory basis.",
			RecommendationRuleName:   "pipeline-system-advisory-v2",
			RecommendationConfidence: "Low",
			RecommendationGroups: models.PipelineAdvisoryGroups{
				ImmediateActions:   []string{"Immediate one", "Immediate two"},
				InspectionMonitor:  []string{"Inspect one"},
				LongTermMitigation: []string{"Mitigate one"},
			},
		},
	}

	html := buildPipelineExcelHTML(assessment, nil)
	for _, want := range []string{
		"Immediate one\nImmediate two",
		"Inspect one",
		"Mitigate one",
		"Engineer-entered recommendation content",
		"pipeline-system-advisory-v2",
		"Low",
		"First note\nSecond note",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected export to contain %q in:\n%s", want, html)
		}
	}
}
