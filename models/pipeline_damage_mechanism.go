package models

import "strings"

const PipelineDamageMechanismSource = "Pipeline DM screening v2"
const PipelineRuleVerified = "VERIFIED"
const PipelineRulePartiallyVerified = "PARTIALLY_VERIFIED"
const PipelineRuleTODOEngineeringConfirmation = "TODO_ENGINEERING_CONFIRMATION"

type PipelineRuleMetadata struct {
	SourceStandard  string `json:"source_standard"`
	ConfidenceLevel string `json:"confidence_level"`
	RuleStatus      string `json:"rule_status"`
}

type PipelineDamageMechanismOption struct {
	Code        string `json:"code"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type PipelineTriggerInput struct {
	Field  string `json:"field"`
	Value  string `json:"value"`
	Reason string `json:"reason"`
}

var pipelineDamageMechanismOptions = []PipelineDamageMechanismOption{
	{Code: "external_corrosion", Label: "External Corrosion", Description: "External metal loss from coating condition, soil, CP, or atmospheric exposure.", Category: "External Damage"},
	{Code: "coating_degradation", Label: "Coating Degradation", Description: "Coating breakdown or corrosion under insulation where the pipeline exposure makes it applicable.", Category: "External Damage"},
	{Code: "third_party_mechanical_damage", Label: "Third-Party / Mechanical Damage", Description: "Excavation, encroachment, dents, gouges, deformation, or construction damage.", Category: "External Damage"},
	{Code: "internal_corrosion", Label: "Internal Corrosion", Description: "Internal metal loss driven by CO2 partial pressure, water content, pH, and corrosion monitoring.", Category: "Internal Thinning"},
	{Code: "localized_corrosion", Label: "Localized Corrosion", Description: "Localized wall loss from pitting or crevice corrosion, influenced by chlorides and pH.", Category: "Internal Thinning"},
	{Code: "erosion", Label: "Erosion", Description: "Mechanical metal loss from high-velocity fluid or solids without significant corrosion contribution.", Category: "Internal Thinning"},
	{Code: "erosion_corrosion", Label: "Erosion-Corrosion", Description: "Synergistic metal loss from velocity-assisted corrosion combining erosion and electrochemical attack.", Category: "Internal Thinning"},
	{Code: "cracking", Label: "Cracking", Description: "Sour service cracking (SSC, HIC) driven by H2S partial pressure per NACE MR0175.", Category: "Internal Cracking"},
	{Code: "scc", Label: "SCC", Description: "Stress corrosion cracking from hoop stress, coating disbondment, and environmental conditions.", Category: "Internal Cracking"},
	{Code: "fatigue", Label: "Fatigue", Description: "Cyclic pressure fatigue from pressure cycling at weld seams and stress concentrations.", Category: "Internal Cracking"},
	{Code: "chemical_damage", Label: "Chemical Damage", Description: "Internal metal loss due to process chemicals.", Category: "Internal Thinning"},
}

var pipelineDamageMechanismMetadata = map[string]PipelineRuleMetadata{
	"external_corrosion":            {SourceStandard: "API 571 / AMPP SP0169", ConfidenceLevel: "Medium", RuleStatus: PipelineRulePartiallyVerified},
	"coating_degradation":           {SourceStandard: "API 571 / AMPP SP0169", ConfidenceLevel: "Low", RuleStatus: PipelineRuleTODOEngineeringConfirmation},
	"third_party_mechanical_damage": {SourceStandard: "API 570 / pipeline integrity management practice", ConfidenceLevel: "Low", RuleStatus: PipelineRuleTODOEngineeringConfirmation},
	"internal_corrosion":            {SourceStandard: "API 581 / API 571", ConfidenceLevel: "Medium", RuleStatus: PipelineRulePartiallyVerified},
	"localized_corrosion":           {SourceStandard: "API 571", ConfidenceLevel: "Low", RuleStatus: PipelineRuleTODOEngineeringConfirmation},
	"erosion":                       {SourceStandard: "API 571 / DNV-RP-O501 concept", ConfidenceLevel: "Low", RuleStatus: PipelineRuleTODOEngineeringConfirmation},
	"erosion_corrosion":             {SourceStandard: "API 571", ConfidenceLevel: "Low", RuleStatus: PipelineRuleTODOEngineeringConfirmation},
	"cracking":                      {SourceStandard: "NACE MR0175 / ISO 15156 / API 571", ConfidenceLevel: "Medium", RuleStatus: PipelineRulePartiallyVerified},
	"scc":                           {SourceStandard: "API 571 / NACE MR0175 / ISO 15156", ConfidenceLevel: "Low", RuleStatus: PipelineRuleTODOEngineeringConfirmation},
	"fatigue":                       {SourceStandard: "API 571", ConfidenceLevel: "Low", RuleStatus: PipelineRuleTODOEngineeringConfirmation},
	"chemical_damage":               {SourceStandard: "Engineering review stub", ConfidenceLevel: "Low", RuleStatus: PipelineRuleTODOEngineeringConfirmation},
}

func PipelineDamageMechanismMetadata(code string) PipelineRuleMetadata {
	if metadata, ok := pipelineDamageMechanismMetadata[NormalizePipelineDamageMechanism(code)]; ok {
		return metadata
	}
	return PipelineRuleMetadata{
		SourceStandard:  "Engineering review required",
		ConfidenceLevel: "Low",
		RuleStatus:      PipelineRuleTODOEngineeringConfirmation,
	}
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
		"coating_degradation":         "coating_degradation",
		"coating_cui_degradation":     "coating_degradation",
		"cui":                         "coating_degradation",
		"third_party_damage":          "third_party_mechanical_damage",
		"mechanical_damage":           "third_party_mechanical_damage",
		"localized_corrosion_pitting": "localized_corrosion",
		"cracking_damage":             "cracking",
		"cracking_scc_fatigue":        "scc",
		"cracking_scc":                "scc",
		"fatigue":                     "fatigue",
		"other_engineering_review":    "chemical_damage",
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
