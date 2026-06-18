package models

import "strings"

const PipelineDamageMechanismSource = "Pipeline metadata classification aligned to API 570 inspection concepts; calculation linkage TODO_ENGINEERING_CONFIRMATION."

type PipelineDamageMechanismOption struct {
	Code        string `json:"code"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

var pipelineDamageMechanismOptions = []PipelineDamageMechanismOption{
	{Code: "external_corrosion", Label: "External Corrosion", Description: "External metal loss from coating condition, soil, CP, or atmospheric exposure.", Category: "External Damage"},
	{Code: "coating_cui_degradation", Label: "Coating / CUI Degradation", Description: "Coating breakdown or corrosion under insulation where the pipeline exposure makes it applicable.", Category: "External Damage"},
	{Code: "third_party_mechanical_damage", Label: "Third-Party / Mechanical Damage", Description: "Excavation, encroachment, dents, gouges, deformation, or construction damage.", Category: "External Damage"},
	{Code: "internal_corrosion", Label: "Internal Corrosion", Description: "Internal metal loss from fluid corrosivity, water, CO2/H2S, MIC, or deposits.", Category: "Internal Thinning"},
	{Code: "localized_corrosion_pitting", Label: "Localized Corrosion / Pitting", Description: "Localized wall loss requiring engineering review for remaining strength impact.", Category: "Internal Thinning"},
	{Code: "erosion_corrosion", Label: "Erosion / Erosion-Corrosion", Description: "Velocity or solids assisted metal loss.", Category: "Internal Thinning"},
	{Code: "chemical_damage", Label: "Chemical Damage", Description: "Internal metal loss due to process chemicals.", Category: "Internal Thinning"},
	{Code: "cracking_damage", Label: "Cracking Damage", Description: "Damage from cracking mechanisms.", Category: "Internal Cracking"},
	{Code: "cracking_scc_fatigue", Label: "Cracking / SCC / Fatigue", Description: "Cracking, stress corrosion cracking, cyclic pressure, vibration, or thermal fatigue screening metadata.", Category: "Internal Cracking"},
	{Code: "other_engineering_review", Label: "Other / Engineering Review", Description: "Use when the mechanism requires engineering confirmation outside configured classifications.", Category: "Internal Cracking"},
}

type PipelineDamageMechanismGroup struct {
	Name    string
	Icon    string
	Style   string
	Options []PipelineDamageMechanismOption
}

func PipelineDamageMechanismOptions() []PipelineDamageMechanismOption {
	options := make([]PipelineDamageMechanismOption, len(pipelineDamageMechanismOptions))
	copy(options, pipelineDamageMechanismOptions)
	return options
}

func PipelineDamageMechanismGroups() []PipelineDamageMechanismGroup {
	groups := []PipelineDamageMechanismGroup{
		{Name: "External Damage", Icon: "mdi-weather-pouring", Style: "dark"},
		{Name: "Internal Thinning", Icon: "mdi-water-percent", Style: "warning"},
		{Name: "Internal Cracking", Icon: "mdi-lightning-bolt", Style: "danger"},
	}
	for i := range groups {
		for _, option := range pipelineDamageMechanismOptions {
			if option.Category == groups[i].Name {
				groups[i].Options = append(groups[i].Options, option)
			}
		}
	}
	return groups
}

func IsValidPipelineDamageMechanism(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	value = NormalizePipelineDamageMechanism(value)
	for _, option := range pipelineDamageMechanismOptions {
		if strings.EqualFold(value, option.Code) || strings.EqualFold(value, option.Label) {
			return true
		}
	}
	return false
}

func NormalizePipelineDamageMechanism(value string) string {
	value = strings.TrimSpace(value)
	legacy := map[string]string{
		"coating_degradation": "coating_cui_degradation",
		"cui":                 "coating_cui_degradation",
		"third_party_damage":  "third_party_mechanical_damage",
		"mechanical_damage":   "third_party_mechanical_damage",
		"cracking_scc":        "cracking_scc_fatigue",
		"fatigue":             "cracking_scc_fatigue",
	}
	if normalized, ok := legacy[strings.ToLower(value)]; ok {
		return normalized
	}
	for _, option := range pipelineDamageMechanismOptions {
		if strings.EqualFold(value, option.Code) || strings.EqualFold(value, option.Label) {
			return option.Code
		}
	}
	return value
}

func PipelineDamageMechanismLabel(value string) string {
	value = strings.TrimSpace(value)
	for _, option := range pipelineDamageMechanismOptions {
		if strings.EqualFold(value, option.Code) || strings.EqualFold(value, option.Label) {
			return option.Label
		}
	}
	return value
}
