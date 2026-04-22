package models

// Struct ini persis ngikutin format JSON Payload dari JavaScript vessel_assessment.js
type AssessmentPayload struct {
	Equipment        EquipmentPayload       `json:"equipment"`
	Assessment       AssessmentGeneral      `json:"assessment"`
	ThicknessData    ThicknessDataPayload   `json:"thickness_data"`
	Environment      EnvironmentPayload     `json:"environment"`
	DamageMechanisms DamageMechanismPayload `json:"damage_mechanisms"`
	Results          ResultPayload          `json:"results"`
	CladdingData     CladdingPayload        `json:"cladding_data"`
}

type EquipmentPayload struct {
	MasterEquipmentID int    `json:"master_equipment_id" binding:"required"`
	TagNumber         string `json:"tag_number" binding:"required"`
	YearBuilt         int    `json:"year_built" binding:"required"`
	FirstUse          int    `json:"first_use" binding:"required"`
	Location          string `json:"location"`
	ShellMaterialID   int    `json:"shell_material_id"`
	HeadMaterialID    int    `json:"head_material_id"`
	TypeHead          int    `json:"type_head"`
	NeckMaterialID    int    `json:"neck_material_id"`
	NozzleMaterialID  int    `json:"nozzle_material_id"`

	DesignPressure     float64 `json:"design_pressure"`
	DesignPressureTube float64 `json:"design_pressure_tube"`
	DesignTemp         float64 `json:"design_temp"`
	DesignTempTube     float64 `json:"design_temp_tube"`
	Diameter           float64 `json:"diameter"`
	DiameterTube       float64 `json:"diameter_tube"`
	Volume             float64 `json:"volume"`
	Length             float64 `json:"length"`
	Nozzle             float64 `json:"nozzle"`

	// Tipe Data Text / String
	DiameterType       string  `json:"diameter_type"`
	DiameterUnit       string  `json:"diameter_unit"`
	DiameterTubeType   *string `json:"diameter_tube_type"` // Pointer karena bisa null dari JS
	DiameterTubeUnit   *string `json:"diameter_tube_unit"`
	LengthUnit         string  `json:"length_unit"`
	VolumeUnit         string  `json:"volume_unit"`
	TempDesignUnit     string  `json:"temp_design_unit"`
	TempDesignTubeUnit *string `json:"temp_design_tube_unit"`
	Pwht               string  `json:"pwht"`
	Certificate        string  `json:"certificate"`
	DataReference      string  `json:"data_reference"`
	NozzleUnit         string  `json:"nozzle_unit"`

	// FIX: Dipindahkan ke sini dari Environment agar cocok dengan JS
	PhaseType          string `json:"phase_type"`
	InternalLining     string `json:"internal_lining"`
	Insulation         string `json:"insulation"`
	SpecialService     string `json:"special_service"`
	Protection         string `json:"protection"`
	CathodicProtection string `json:"cathodic_protection"`

	// --- NEW FIELDS STEP 1 ---
	SerialNumber          string  `json:"serial_number"`
	EquipLife             int     `json:"equip_life"`
	PartType              string  `json:"part_type"`
	ConstructionCode      string  `json:"construction_code"`
	JointEfficiency       float64 `json:"joint_efficiency"`
	JointEfficiencyHead   float64 `json:"joint_efficiency_head"`
	JointType             string  `json:"joint_type"`
	Radiographic          string  `json:"radiographic"`
	ConstructionType      string  `json:"construction_type"`
	Mawp                  float64 `json:"mawp"`
	HydroTest             float64 `json:"hydro_test"`
	CrownRadius           float64 `json:"crown_radius"`
	KnuckleRadius         float64 `json:"knuckle_radius"`
	InternalPartsMaterial string  `json:"internal_parts_material"`
	ShellContaminant      string  `json:"shell_contaminant"`
	MaxBrinell            string  `json:"max_brinell"`
	AllowableStress       float64 `json:"allowable_stress"`

	// Thickness Data Baseline (Step 1)
	InspectionInterval  int     `json:"inspection_interval"`
	PrevInspection      string  `json:"prev_inspection"`
	ActInspection       string  `json:"act_inspection"`
	CorrosionAllowance  float64 `json:"corrosion_allowance"`
	ShellCladBaseMetal  float64 `json:"shell_clad_base_metal"`
	HeadCladBaseMetal   float64 `json:"head_clad_base_metal"`
	NozzleCladBaseMetal float64 `json:"nozzle_clad_base_metal"`
	ShellWallThickness  float64 `json:"shell_wall_thickness"`
	HeadWallThickness   float64 `json:"head_wall_thickness"`
	NozzleWallThick     float64 `json:"nozzle_wall_thick"`
	ShellThickCladded   float64 `json:"shell_thick_cladded"`
	HeadThickCladded    float64 `json:"head_thick_cladded"`
	NozzleThickCladded  float64 `json:"nozzle_thick_cladded"`
	PrevThickShell      float64 `json:"prev_thick_shell"`
	PrevThickHead       float64 `json:"prev_thick_head"`
	NozzlePreviousThick float64 `json:"nozzle_previous_thick"`
	ActThickShell       float64 `json:"act_thick_shell"`
	ActThickHead        float64 `json:"act_thick_head"`
	NozzleActualThick   float64 `json:"nozzle_actual_thick"`
}

type AssessmentGeneral struct {
	AssessmentDate     string `json:"assessment_date"`
	PrevInspectionDate string `json:"prev_inspection_date"`
	ActInspectionDate  string `json:"act_inspection_date"`

	OperatingPressure     float64 `json:"operating_pressure"`
	OperatingPressureTube float64 `json:"operating_pressure_tube"`
	OperatingTemp         float64 `json:"operating_temp"`
	OperatingTempTube     float64 `json:"operating_temp_tube"`

	TempOpUnit     string  `json:"temp_op_unit"`
	TempOpTubeUnit *string `json:"temp_op_tube_unit"`
}

type EnvironmentPayload struct {
	Phase         string  `json:"phase"`
	H2SContent    float64 `json:"h2s_content"`
	CO2Content    float64 `json:"co2_content"`
	H2OContent    float64 `json:"h2o_content"`
	ChlorideIndex int     `json:"chloride_index"`
	PHIndex       int     `json:"ph_index"`

	// FIX: Data Step 3 dipindahkan ke sini agar cocok dengan JS
	ImpactProduction      string  `json:"impact_production"`
	InsulationCondition   string  `json:"insulation_condition"`
	InsulationDamageLevel string  `json:"insulation_damage_level"`
	CoatingCondition      string  `json:"coating_condition"`
	CoatingDamageLevel    string  `json:"coating_damage_level"`
	CorrectiveDescription string  `json:"corrective_description"`
	CorrectiveAction      string  `json:"corrective_action"`
	CorrectiveDate        *string `json:"corrective_date"` // Pointer karena bisa null

	// --- NEW FIELDS STEP 3 ---
	ContaminantAmine     string `json:"contaminant_amine"`
	FlowVelocity         string `json:"flow_velocity"`
	PreventiveCorrosion  string `json:"preventive_corrosion"`
	InhibitorEffectivity string `json:"inhibitor_effectivity"`
	EnvExtCracking       string `json:"env_ext_cracking"`
	Vibration            string `json:"vibration"`

	// --- TAMBAHAN FULL STEP 3 ---
	ImpactForProduction string  `json:"impact_for_production"`
	CompNitrogen        float64 `json:"comp_nitrogen"`
	CompMethane         float64 `json:"comp_methane"`
	CompEthane          float64 `json:"comp_ethane"`
	CompPropane         float64 `json:"comp_propane"`
	CompButane          float64 `json:"comp_butane"`
	CompSolvent         float64 `json:"comp_solvent"`
	CompAir             float64 `json:"comp_air"`
	H2SPpm              int     `json:"h2s_ppm"`

	Fluida                  string `json:"fluida"`
	Pollutant               string `json:"pollutant"`
	CpCondition             string `json:"cp_condition"`
	CorrosionMonitoring     string `json:"corrosion_monitoring"`
	BiocideTreatment        string `json:"biocide_treatment"`
	ReleaseFluidContainment string `json:"release_fluid_containment"`
	CleanUpTime             string `json:"clean_up_time"`
	HeatTraced              int    `json:"heat_traced"`
	SteamOut                int    `json:"steam_out"`

	PrevExtCorrosion    string `json:"prev_ext_corrosion"`
	ConfExtCorrosion    string `json:"conf_ext_corrosion"`
	PrevIntCracking     string `json:"prev_int_cracking"`
	ConfIntCracking     string `json:"conf_int_cracking"`
	PrevIntThinning     string `json:"prev_int_thinning"`
	ConfIntThinning     string `json:"conf_int_thinning"`
	PrevLocIntCorrosion string `json:"prev_loc_int_corrosion"`
	ConfLocIntCorrosion string `json:"conf_loc_int_corrosion"`
}

type ThicknessDataPayload struct {
	Shell  ComponentThickness `json:"shell"`
	Head   ComponentThickness `json:"head"`
	Nozzle ComponentThickness `json:"nozzle"`
}

type ComponentThickness struct {
	PrevThick     float64 `json:"prev_thick"`
	ActThick      float64 `json:"act_thick"`
	TReq          float64 `json:"t_req"`
	CorrosionRate float64 `json:"corrosion_rate"`
	RemainingLife float64 `json:"remaining_life"`
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
	LofCategory        int               `json:"lof_category"`
	CofFinancial       string            `json:"cof_financial"`
	CofSafety          string            `json:"cof_safety"`
	CofCategory        string            `json:"cof_category"`
	RiskLevel          string            `json:"risk_level"`
	RiskIndex          int               `json:"risk_index"`
	InspectionData     map[string]string `json:"inspection_data"`
	GoverningComponent string            `json:"governing_component"`
	MaxIntervalYears   float64           `json:"max_interval_years"`
	NextInspectionYear int               `json:"next_inspection_year"`
	RecommendedMethod  string            `json:"recommended_method"`
}

type CladdingPayload struct {
	Shell  CladDetail `json:"shell"`
	Head   CladDetail `json:"head"`
	Nozzle CladDetail `json:"nozzle"`
}

type CladDetail struct {
	BaseMetal float64 `json:"base_metal"`
	Cladding  float64 `json:"cladding"`
	TotalInit float64 `json:"total_init"`
	ActNow    float64 `json:"act_now"`
	ActCons   float64 `json:"act_cons"`
	Status    string  `json:"status"`
}
