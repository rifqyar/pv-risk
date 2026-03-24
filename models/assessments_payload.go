package models

// Struct ini persis ngikutin format JSON Payload dari JavaScript
type AssessmentPayload struct {
	Equipment        EquipmentPayload       `json:"equipment"`
	Assessment       AssessmentGeneral      `json:"assessment"`
	ThicknessData    ThicknessDataPayload   `json:"thickness_data"`
	Environment      EnvironmentPayload     `json:"environment"`
	DamageMechanisms DamageMechanismPayload `json:"damage_mechanisms"`
	Results          ResultPayload          `json:"results"`
}

type EquipmentPayload struct {
	MasterEquipmentID  int      `json:"master_equipment_id" binding:"required"`
	TagNumber          string   `json:"tag_number" binding:"required"`
	YearBuilt          int      `json:"year_built" binding:"required"`
	ShellMaterialID    int      `json:"shell_material_id"`
	DesignPressure     float64  `json:"design_pressure"`
	DesignPressureTube *float64 `json:"design_pressure_tube"`
	DesignTemp         float64  `json:"design_temp"`
	DesignTempTube     *float64 `json:"design_temp_tube"`
	Diameter           float64  `json:"diameter"`
	DiameterTube       float64  `json:"diameter_tube"`
	Volume             float64  `json:"volume"`
}
type AssessmentGeneral struct {
	AssessmentDate        string  `json:"assessment_date"`
	PrevInspectionDate    string  `json:"prev_inspection_date"`
	ActInspectionDate     string  `json:"act_inspection_date"`
	OperatingPressure     float64 `json:"operating_pressure"`
	OperatingTemp         float64 `json:"operating_temp"`
	OperatingPressureTube float64 `json:"operating_pressure_tube"`
	OperatingTempTube     float64 `json:"operating_temp_tube"`
}

type ThicknessDataPayload struct {
	Shell ComponentThickness `json:"shell"`
	Head  ComponentThickness `json:"head"`
}

type ComponentThickness struct {
	PrevThick     float64 `json:"prev_thick"`
	ActThick      float64 `json:"act_thick"`
	TReq          float64 `json:"t_req"`
	CorrosionRate float64 `json:"corrosion_rate"`
	RemainingLife float64 `json:"remaining_life"`
}

type EnvironmentPayload struct {
	Phase                 string  `json:"phase"`
	H2sContent            float64 `json:"h2s_content"`
	Co2Content            float64 `json:"co2_content"`
	H2oContent            float64 `json:"h2o_content"`
	ChlorideIndex         int     `json:"chloride_index"`
	PhIndex               int     `json:"ph_index"`
	ImpactProduction      string  `json:"impact_production"`
	InsulationCondition   string  `json:"insulation_condition"`
	InsulationDamageLevel string  `json:"insulation_damage_level"`
	CoatingCondition      string  `json:"coating_condition"`
	CoatingDamageLevel    string  `json:"coating_damage_level"`
	CorrectiveDescription string  `json:"corrective_description"`
	CorrectiveAction      string  `json:"corrective_action"`
	CorrectiveDate        *string `json:"corrective_date"` // Pointer string buat ngakalin NULL kalau kosong
}

type DamageMechanismPayload struct {
	Atmospheric       string  `json:"atmospheric"`
	Cui               string  `json:"cui"`
	ExtCracking       string  `json:"ext_cracking"`
	Co2               string  `json:"co2"`
	Mic               string  `json:"mic"`
	Ssc               string  `json:"ssc"`
	AmineScc          string  `json:"amine_scc"`
	Hic               string  `json:"hic"`
	Ciscc             string  `json:"ciscc"`
	Galvanic          string  `json:"galvanic"`
	LofScore          string  `json:"lof_score"`
	TotalDamageFactor float64 `json:"total_damage_factor"`
}

type ResultPayload struct {
	LofCategory           int     `json:"lof_category"`
	CofFinancial          string  `json:"cof_financial"` // Tambahan
	CofSafety             string  `json:"cof_safety"`    // Tambahan
	CofCategory           string  `json:"cof_category"`
	RiskLevel             string  `json:"risk_level"`
	RiskIndex             int     `json:"risk_index"`
	InspInternalThinning  string  `json:"insp_internal_thinning"`  // Tambahan
	InspExternalCorrosion string  `json:"insp_external_corrosion"` // Tambahan
	InspCracking          string  `json:"insp_cracking"`           // Tambahan
	GoverningComponent    string  `json:"governing_component"`
	MaxIntervalYears      float64 `json:"max_interval_years"`
	NextInspectionYear    int     `json:"next_inspection_year"`
	RecommendedMethod     string  `json:"recommended_method"`
}
