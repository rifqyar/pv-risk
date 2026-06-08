package main

import (
	"html/template"
	"testing"
)

func TestEmbeddedTemplatesParse(t *testing.T) {
	_, err := template.New("").Funcs(template.FuncMap{
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		"mul": func(a, b int) int {
			return a * b
		},
	}).ParseFS(templateFS,
		"templates/*.html",
		"templates/layouts/*.html",
		"templates/partials/*.html",
	)
	if err != nil {
		t.Fatalf("embedded templates failed to parse: %v", err)
	}
}
