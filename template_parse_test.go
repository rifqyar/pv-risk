package main

import (
	"bytes"
	"html/template"
	"pv-risk/models"
	"strings"
	"testing"
)

func TestEmbeddedTemplatesParse(t *testing.T) {
	_, err := template.New("").Funcs(appTemplateFuncs()).ParseFS(templateFS,
		"templates/*.html",
		"templates/layouts/*.html",
		"templates/partials/*.html",
	)
	if err != nil {
		t.Fatalf("embedded templates failed to parse: %v", err)
	}
}

func TestPipelinePDFEngineeringNotesConditionalRendering(t *testing.T) {
	tmpl, err := template.New("").Funcs(appTemplateFuncs()).ParseFS(templateFS,
		"templates/*.html",
		"templates/layouts/*.html",
		"templates/partials/*.html",
	)
	if err != nil {
		t.Fatalf("embedded templates failed to parse: %v", err)
	}

	data := func(notes string) map[string]interface{} {
		return map[string]interface{}{
			"Assessment": &models.PipelineOilAssessment{
				Input: models.PipelineOilInput{
					ReportNo: "PIPE-1",
					RiskInput: models.PipelineOilRiskInput{
						EngineeringNotes: notes,
					},
				},
				Result: &models.PipelineOilResult{
					RecommendationSource:     "source",
					RecommendationRuleName:   "advisory",
					RecommendationConfidence: "Low",
					RecommendationGroups: models.PipelineAdvisoryGroups{
						ImmediateActions: []string{"Act now"},
					},
				},
			},
			"AuditEvents":         []models.PipelineOilAuditEvent{},
			"StandardsReferences": []string{},
		}
	}

	var withNotes bytes.Buffer
	if err = tmpl.ExecuteTemplate(&withNotes, "pipeline_detail_content", data("First line\nSecond line")); err != nil {
		t.Fatalf("execute detail with notes: %v", err)
	}
	if !strings.Contains(withNotes.String(), "Engineering Notes") ||
		!strings.Contains(withNotes.String(), "pipeline-pdf-preline") ||
		!strings.Contains(withNotes.String(), "First line") ||
		!strings.Contains(withNotes.String(), "Second line") {
		t.Fatalf("expected PDF Engineering Notes section with multiline content")
	}

	var emptyNotes bytes.Buffer
	if err = tmpl.ExecuteTemplate(&emptyNotes, "pipeline_detail_content", data("")); err != nil {
		t.Fatalf("execute detail without notes: %v", err)
	}
	if strings.Contains(emptyNotes.String(), "<div class=\"pipeline-pdf-section\">Engineering Notes</div>") {
		t.Fatalf("did not expect empty PDF Engineering Notes section")
	}
}
