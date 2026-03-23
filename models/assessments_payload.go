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
	TagNumber       string  `json:"tag_number" binding:"required"`
	Description     string  `json:"description"`
	EquipmentTypeID string  `json:"equipment_type_id" binding:"required"`
	YearBuilt       int     `json:"year_built" binding:"required"`
	ShellMaterialID int     `json:"shell_material_id"`
	DesignPressure  float64 `json:"design_pressure"`
	DesignTemp      float64 `json:"design_temp"`
	Diameter        float64 `json:"diameter"`
	Volume          float64 `json:"volume"`
}
type AssessmentGeneral struct {
	AssessmentDate     string  `json:"assessment_date"`
	PrevInspectionDate string  `json:"prev_inspection_date"`
	ActInspectionDate  string  `json:"act_inspection_date"`
	OperatingPressure  float64 `json:"operating_pressure"`
	OperatingTemp      float64 `json:"operating_temp"`
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
	Phase         string  `json:"phase"`
	H2sContent    float64 `json:"h2s_content"`
	Co2Content    float64 `json:"co2_content"`
	H2oContent    float64 `json:"h2o_content"`
	ChlorideIndex int     `json:"chloride_index"`
	PhIndex       int     `json:"ph_index"`
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
	TotalDamageFactor float64 `json:"total_damage_factor"`
}

type ResultPayload struct {
	LofCategory        int     `json:"lof_category"`
	CofCategory        string  `json:"cof_category"`
	RiskLevel          string  `json:"risk_level"`
	RiskIndex          int     `json:"risk_index"`
	GoverningComponent string  `json:"governing_component"`
	MaxIntervalYears   float64 `json:"max_interval_years"`
	NextInspectionYear int     `json:"next_inspection_year"`
	RecommendedMethod  string  `json:"recommended_method"`
}
