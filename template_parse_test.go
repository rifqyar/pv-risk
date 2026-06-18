package main

import (
	"html/template"
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
