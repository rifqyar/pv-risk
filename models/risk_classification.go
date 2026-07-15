package models

import (
	"math"
	"strings"
)

const (
	ApprovedRiskThresholdSource    = "pressure-vessel-risk-matrix-v1"
	ApprovedRiskThresholdVersion   = "pressure-vessel-approved-v1"
	ApprovedRiskThresholdEffective = "2025-07-01"
)

func ApprovedPoFCategory(value float64) string {
	switch {
	case value >= 0.01:
		return "5"
	case value >= 0.001:
		return "4"
	case value >= 0.0001:
		return "3"
	case value >= 0.00001:
		return "2"
	default:
		return "1"
	}
}

func ApprovedPoFNumeric(category string) float64 {
	switch strings.TrimSpace(category) {
	case "5":
		return 5
	case "4":
		return 4
	case "3":
		return 3
	case "2":
		return 2
	default:
		return 1
	}
}

func ApprovedCoFCategoryFromIndex(value float64) string {
	switch {
	case value >= 5:
		return "E"
	case value >= 4:
		return "D"
	case value >= 3:
		return "C"
	case value >= 2:
		return "B"
	default:
		return "A"
	}
}

func ApprovedCoFNumeric(category string) float64 {
	switch strings.ToUpper(strings.TrimSpace(category)) {
	case "E":
		return 5
	case "D":
		return 4
	case "C":
		return 3
	case "B":
		return 2
	default:
		return 1
	}
}

func ApprovedRiskLevelFromMatrixScore(score int) string {
	switch {
	case score <= 5:
		return "Low"
	case score <= 10:
		return "Medium"
	case score <= 15:
		return "High"
	default:
		return "Extreme"
	}
}

func ApprovedPipelineRiskLevelFromMatrixScore(score int) string {
	switch ApprovedRiskLevelFromMatrixScore(score) {
	case "Extreme":
		return "Critical Risk"
	case "High":
		return "High Risk"
	case "Medium":
		return "Medium Risk"
	default:
		return "Low Risk"
	}
}

func ApprovedRiskLevelFromCategories(pofCategory, cofCategory string) string {
	score := int(math.Ceil(ApprovedPoFNumeric(pofCategory)) * math.Ceil(ApprovedCoFNumeric(cofCategory)))
	return ApprovedPipelineRiskLevelFromMatrixScore(score)
}
