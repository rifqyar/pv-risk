package main

import (
	"encoding/json"
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
		"toJSON": func(v interface{}) template.JS {
			b, err := json.Marshal(v)
			if err != nil {
				return template.JS("null")
			}
			return template.JS(b)
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
