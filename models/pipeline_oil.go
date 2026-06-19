package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const PipelineOilFormulaVersion = "pipeline-oil-risk-v2"

var (
	ErrPipelineValidation = errors.New("pipeline oil validation failed")
	ErrPipelineFinalized  = errors.New("pipeline oil assessment is finalized")
)

type PipelineOilAssessmentStatus string

const (
	PipelineOilStatusDraft      PipelineOilAssessmentStatus = "DRAFT"
	PipelineOilStatusCalculated PipelineOilAssessmentStatus = "CALCULATED"
	PipelineOilStatusArchived   PipelineOilAssessmentStatus = "ARCHIVED"
)

const (
	defaultPipelineGFF             = 0.00003
	defaultPipelineBaseTPDRate     = 1.0
	defaultPipelineBaseCorrRate    = 1.0
	defaultPipelineManagementScore = 500.0
	cubicMetersToBarrels           = 6.28981077
	maxPipelineRemainingLifeYears  = 20.0
)

var (
	// TODO_ENGINEERING_CONFIRMATION: All factor values below are neutral placeholders.
	// Engineering must confirm each value per API 581 pipeline damage factor tables.
	// Ranges and categories reference established standards (NACE, API, ASME) but
	// the numeric multiplier for each level requires engineering sign-off.
	pipelineDepthFactors = map[string]float64{
		"<1m":  1.0, // TODO_ENGINEERING_CONFIRMATION; range is standard practice
		"1-2m": 1.0, // TODO_ENGINEERING_CONFIRMATION
		">2m":  1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelinePatrolFactors = map[string]float64{
		"rare":         1.0, // TODO_ENGINEERING_CONFIRMATION
		"monthly":      1.0, // TODO_ENGINEERING_CONFIRMATION
		"weekly_daily": 1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineROWFactors = map[string]float64{
		"poor": 1.0, // TODO_ENGINEERING_CONFIRMATION
		"fair": 1.0, // TODO_ENGINEERING_CONFIRMATION
		"good": 1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineSoilFactors = map[string]float64{
		"<1000":     1.0, // TODO_ENGINEERING_CONFIRMATION; ranges aligned with NACE soil classification
		"1000-5000": 1.0, // TODO_ENGINEERING_CONFIRMATION
		">5000":     1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineCoatingConditionFactors = map[string]float64{
		"Good":            1.0, // TODO_ENGINEERING_CONFIRMATION; categories from PV Section D
		"Damaged":         1.0, // TODO_ENGINEERING_CONFIRMATION
		"Not Inspectable": 1.0, // TODO_ENGINEERING_CONFIRMATION
		"Not Applicable":  1.0,
	}
	pipelineCPFactors = map[string]float64{
		"failed":     1.0, // TODO_ENGINEERING_CONFIRMATION; -850mV gate is sourced from NACE SP0169
		"borderline": 1.0, // TODO_ENGINEERING_CONFIRMATION
		"normal":     1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineEnvironmentalMultipliers = map[string]float64{
		"low":    1.0,
		"medium": 1.5,
		"high":   2.5,
	}
	pipelineClassLocationFactors = map[string]float64{
		"class_1": 1.0,
		"class_2": 1.0,
		"class_3": 1.0,
		"class_4": 1.0,
	}

	// pCO2 partial pressure thresholds for sweet corrosion severity (psig)
	pipelineCO2PartialPressureSeverity = map[string]float64{
		"Low":      5.0,
		"Moderate": 20.0,
		"High":     1e9,
	}

	// Sourced: NACE MR0175 / PV SSC logic
	// pH2S partial pressure thresholds for sour service severity (psig)
	pipelineH2SPartialPressureSeverity = map[string]float64{
		"Not":      0.05,
		"Low":      0.5,
		"Moderate": 15.0,
		"High":     1e9,
	}

	// TODO_ENGINEERING_CONFIRMATION: All values below are neutral placeholders.
	pipelineFluidCorrosivityMPYFactors = map[string]float64{
		"<2 mpy":   1.0, // NACE RP0775 categories
		"2-5 mpy":  1.0, // TODO_ENGINEERING_CONFIRMATION
		"5-10 mpy": 1.0, // TODO_ENGINEERING_CONFIRMATION
		">10 mpy":  1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelinePHSeverity = map[string]float64{
		"â‰¤4.5":  1.0, // TODO_ENGINEERING_CONFIRMATION
		"4.5-6.5": 1.0, // TODO_ENGINEERING_CONFIRMATION
		"6.5-8.5": 1.0, // TODO_ENGINEERING_CONFIRMATION
		">8.5":    1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineChlorideSeverity = map[int]float64{
		1: 1.0, 2: 1.0, 3: 1.0, 4: 1.0, 5: 1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineInhibitorModifiers = map[string]float64{
		"High (>90%)":     1.0, // TODO_ENGINEERING_CONFIRMATION; percentage ranges from PV
		"Medium (60-90%)": 1.0, // TODO_ENGINEERING_CONFIRMATION
		"Low (<60%)":      1.0, // TODO_ENGINEERING_CONFIRMATION
		"None":            1.0,
	}
	pipelineCoatingDamageModifiers = map[string]float64{
		"Small":  1.0, // TODO_ENGINEERING_CONFIRMATION; categories from PV Section D
		"Medium": 1.0, // TODO_ENGINEERING_CONFIRMATION
		"Large":  1.0, // TODO_ENGINEERING_CONFIRMATION
		"Severe": 1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineInsulationDamageModifiers = map[string]float64{
		"Small":  1.0, // TODO_ENGINEERING_CONFIRMATION
		"Medium": 1.0, // TODO_ENGINEERING_CONFIRMATION
		"Large":  1.0, // TODO_ENGINEERING_CONFIRMATION
		"Severe": 1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelinePreviousFindingSeverity = map[string]float64{
		"none":            0.0,
		"finding":         1.0, // TODO_ENGINEERING_CONFIRMATION â€” escalation magnitude
		"not_inspectable": 0.5, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineConfidenceWeight = map[string]float64{
		"high":    1.0, // TODO_ENGINEERING_CONFIRMATION
		"average": 1.0, // TODO_ENGINEERING_CONFIRMATION
		"low":     1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineWeldCrackingModifiers = map[string]float64{
		"Seamless": 1.0, // TODO_ENGINEERING_CONFIRMATION; direction from PV SSC logic
		"SAW":      1.0, // TODO_ENGINEERING_CONFIRMATION
		"ERW":      1.0, // TODO_ENGINEERING_CONFIRMATION
		"Other":    1.0,
	}
	pipelinePWHTModifiers = map[string]float64{
		"Yes":     1.0, // TODO_ENGINEERING_CONFIRMATION; direction from PV SSC logic
		"No":      1.0, // TODO_ENGINEERING_CONFIRMATION
		"Unknown": 1.0,
	}
	pipelineOneCallModifiers = map[string]float64{
		"Active and Effective": 1.0, // TODO_ENGINEERING_CONFIRMATION
		"Limited":              1.0, // TODO_ENGINEERING_CONFIRMATION
		"None":                 1.0,
	}
	pipelineH2SPpmSeverity = map[string]float64{
		"<50 ppm":     1.0, // TODO_ENGINEERING_CONFIRMATION; ranges from NACE MR0175
		"50-1000 ppm": 1.0, // TODO_ENGINEERING_CONFIRMATION
		">1000 ppm":   1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineFlowVelocityModifiers = map[string]float64{
		"Low (<3 m/s)":        1.0, // TODO_ENGINEERING_CONFIRMATION; thresholds reference DNV-RP-O501 concept
		"Moderate (3-10 m/s)": 1.0, // TODO_ENGINEERING_CONFIRMATION
		"High (10-20 m/s)":    1.0, // TODO_ENGINEERING_CONFIRMATION
		"Very High (>20 m/s)": 1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineSolidContentModifiers = map[string]float64{
		"None":     1.0, // TODO_ENGINEERING_CONFIRMATION
		"Trace":    1.0, // TODO_ENGINEERING_CONFIRMATION
		"Moderate": 1.0, // TODO_ENGINEERING_CONFIRMATION
		"Heavy":    1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineSCCStressThresholds = map[string]float64{
		"Low":      30.0, // TODO_ENGINEERING_CONFIRMATION; % SMYS thresholds
		"Moderate": 50.0, // TODO_ENGINEERING_CONFIRMATION
		"High":     1e9,  // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineErosionVelocityThresholds = map[string]float64{
		"Low":       3.0,  // TODO_ENGINEERING_CONFIRMATION; m/s thresholds
		"Moderate":  10.0, // TODO_ENGINEERING_CONFIRMATION
		"High":      20.0, // TODO_ENGINEERING_CONFIRMATION
		"Very High": 1e9,  // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineFatigueCycleThresholds = map[string]int{
		"Low":      100,     // TODO_ENGINEERING_CONFIRMATION; cycles/year
		"Moderate": 10000,   // TODO_ENGINEERING_CONFIRMATION
		"High":     1000000, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineWallThicknessRatioThresholds = map[string]float64{
		"Acceptable":               1.0, // ratio >= 1.0; API 579 concept
		"Conditionally Acceptable": 0.8, // TODO_ENGINEERING_CONFIRMATION
		"Not Acceptable":           0.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineExtCrackingOptions = map[string]float64{
		"None":     1.0, // TODO_ENGINEERING_CONFIRMATION; categories from PV Section C
		"H2S":      1.0, // TODO_ENGINEERING_CONFIRMATION
		"Chloride": 1.0, // TODO_ENGINEERING_CONFIRMATION
		"Hydrogen": 1.0, // TODO_ENGINEERING_CONFIRMATION
		"Marine":   1.0, // TODO_ENGINEERING_CONFIRMATION
	}
	pipelineBiocideTreatmentValues = map[string]bool{
		"Yes":          true,
		"No":           false,
		"Not Required": true, // treated as no MIC concern
	}
	pipelinePrevFindingOptions         = []string{"No Finding", "Finding", "Not Inspectable"}
	pipelineConfidenceOptions          = []string{"", "High", "Average", "Low"}
	pipelinePHLevelOptions             = []string{"â‰¤4.5", "4.5-6.5", "6.5-8.5", ">8.5"}
	pipelineCorrosivityMPYOptions      = []string{"<2 mpy", "2-5 mpy", "5-10 mpy", ">10 mpy"}
	pipelineInhibitorOptions           = []string{"High (>90%)", "Medium (60-90%)", "Low (<60%)", "None"}
	pipelineBiocideOptions             = []string{"Yes", "No", "Not Required"}
	pipelineCorrosionMonitoringOptions = []string{"Satisfactory", "Unsatisfactory", "Not Applicable"}
	pipelinePWHTOptions                = []string{"Yes", "No", "Unknown"}
	pipelineWeldJointOptions           = []string{"Seamless", "SAW", "ERW", "Other"}
	pipelineCoatingConditionOptions    = []string{"Good", "Damaged", "Not Inspectable", "Not Applicable"}
	pipelineInsulationConditionOptions = []string{"Good", "Damaged", "Not Inspectable", "Not Applicable"}
	pipelineDamageLevelOptions         = []string{"Small", "Medium", "Large", "Severe"}
	pipelineExtCrackingEnvOptions      = []string{"None", "H2S", "Chloride", "Hydrogen", "Marine"}
	pipelineOneCallOptions             = []string{"Active and Effective", "Limited", "None"}
	pipelineFlowVelocityOptions        = []string{"Low (<3 m/s)", "Moderate (3-10 m/s)", "High (10-20 m/s)", "Very High (>20 m/s)"}
	pipelineSolidContentOptions        = []string{"None", "Trace", "Moderate", "Heavy"}
	pipelineH2SPpmOptions              = []string{"<50 ppm", "50-1000 ppm", ">1000 ppm"}
	pipelineClassLocationOptions       = []string{"class_1", "class_2", "class_3", "class_4"}
)

type FlexibleYear string

func (fy *FlexibleYear) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*fy = FlexibleYear(s)
		return nil
	}
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*fy = FlexibleYear(strconv.Itoa(i))
		return nil
	}
	return fmt.Errorf("FlexibleYear: cannot unmarshal %s", string(data))
}

func (fy FlexibleYear) Float() float64 {
	return parseMonthYearToFloat(string(fy))
}

type PipelineOilInput struct {
	ID                        int                          `json:"id" form:"id"`
	ReportNo                  string                       `json:"report_no" form:"report_no"`
	PlaceIssued               string                       `json:"place_issued" form:"place_issued"`
	DateIssued                string                       `json:"date_issued" form:"date_issued"`
	OwnerUser                 string                       `json:"owner_user" form:"owner_user"`
	Contractor                string                       `json:"contractor" form:"contractor"`
	Location                  string                       `json:"location" form:"location"`
	LineIdentification        string                       `json:"line_identification" form:"line_identification"`
	YearBuilt                 int                          `json:"year_built" form:"year_built"`
	YearUsed                  FlexibleYear                 `json:"year_used" form:"year_used"`
	Service                   string                       `json:"service" form:"service"`
	PipeSize                  string                       `json:"pipe_size" form:"pipe_size"`
	PipeLengthM               float64                      `json:"pipe_length_m" form:"pipe_length_m"`
	MaterialSpecification     string                       `json:"material_specification" form:"material_specification"`
	FlangeMaterialSpec        string                       `json:"flange_material_spec" form:"flange_material_spec"`
	SMYSPsi                   float64                      `json:"smys_psi" form:"smys_psi"`
	InternalDesignPressurePsi float64                      `json:"internal_design_pressure_psi" form:"internal_design_pressure_psi"`
	DesignTemperatureF        float64                      `json:"design_temperature_f" form:"design_temperature_f"`
	OperatingTemperatureF     float64                      `json:"operating_temperature_f" form:"operating_temperature_f"`
	TestPressurePsi           float64                      `json:"test_pressure_psi" form:"test_pressure_psi"`
	MethodOfJoining           string                       `json:"method_of_joining" form:"method_of_joining"`
	JointEfficiency           float64                      `json:"joint_efficiency" form:"joint_efficiency"`
	CoatingType               string                       `json:"coating_type" form:"coating_type"`
	CorrosionControl          string                       `json:"corrosion_control" form:"corrosion_control"`
	AllowanceIn               float64                      `json:"allowance_in" form:"allowance_in"`
	RightOfWay                string                       `json:"right_of_way" form:"right_of_way"`
	SafetyDevice              string                       `json:"safety_device" form:"safety_device"`
	AreaClassification        string                       `json:"area_classification" form:"area_classification"`
	InspectionPeriod          string                       `json:"inspection_period" form:"inspection_period"`
	ApplicableCode            string                       `json:"applicable_code" form:"applicable_code"`
	OutsideDiameterIn         float64                      `json:"outside_diameter_in" form:"outside_diameter_in"`
	OperatingPressurePsi      float64                      `json:"operating_pressure_psi" form:"operating_pressure_psi"`
	RadiographicPercent       float64                      `json:"radiographic_percent" form:"radiographic_percent"`
	NominalWallThicknessMM    float64                      `json:"nominal_wall_thickness_mm" form:"nominal_wall_thickness_mm"`
	ActualWallThicknessMM     float64                      `json:"actual_wall_thickness_mm" form:"actual_wall_thickness_mm"`
	TypeOfInstallation        string                       `json:"type_of_installation" form:"type_of_installation"`
	QualityFactor             float64                      `json:"quality_factor" form:"quality_factor"`
	WeldJointStrengthFactor   float64                      `json:"weld_joint_strength_factor" form:"weld_joint_strength_factor"`
	DesignFactor              float64                      `json:"design_factor" form:"design_factor"`
	MaterialStressPsi         float64                      `json:"material_stress_psi" form:"material_stress_psi"`
	PreviousSKPP              string                       `json:"previous_skpp" form:"previous_skpp"`
	ExpirationDate            string                       `json:"expiration_date" form:"expiration_date"`
	CorrosionRateMPY          *float64                     `json:"corrosion_rate_mpy" form:"corrosion_rate_mpy"`
	TemperatureDeratingFactor float64                      `json:"temperature_derating_factor" form:"temperature_derating_factor"`
	ManualRecommendation      string                       `json:"manual_recommendation" form:"manual_recommendation"`
	AssessmentBy              string                       `json:"assessment_by" form:"assessment_by"`
	InspectionPoints          []PipelineOilInspectionPoint `json:"inspection_points"`
	RiskInput                 PipelineOilRiskInput         `json:"risk_input"`
}

type PipelineOilRiskInput struct {
	DamageMechanism           string                                 `json:"damage_mechanism" form:"damage_mechanism"`
	InspectionEffectivity     string                                 `json:"inspection_effectivity" form:"inspection_effectivity"`
	InspectionEffectivityByDM map[string]string                      `json:"inspection_effectivity_by_damage_mechanism" form:"inspection_effectivity_by_damage_mechanism"`
	InspectionPlanByDM        map[string]PipelineInspectionPlanInput `json:"inspection_plan_by_damage_mechanism" form:"inspection_plan_by_damage_mechanism"`
	ReleaseFluid              string                                 `json:"release_fluid" form:"release_fluid"`
	GenericFailureFrequency   float64                                `json:"generic_failure_frequency" form:"generic_failure_frequency"`
	ManagementSystemScore     float64                                `json:"management_system_score" form:"management_system_score"`
	DamageFactor              float64                                `json:"damage_factor" form:"damage_factor"`
	BaseTPDRate               float64                                `json:"base_tpd_rate" form:"base_tpd_rate"`
	BaseExternalCorrRate      float64                                `json:"base_external_corr_rate" form:"base_external_corr_rate"`
	BaseInternalCorrRate      float64                                `json:"base_internal_corr_rate" form:"base_internal_corr_rate"`
	// Inspection History
	DepthOfCover       string  `json:"depth_of_cover" form:"depth_of_cover"`
	PatrolFrequency    string  `json:"patrol_frequency" form:"patrol_frequency"`
	ROWCondition       string  `json:"row_condition" form:"row_condition"`
	SoilResistivity    string  `json:"soil_resistivity" form:"soil_resistivity"`
	CoatingCondition   string  `json:"coating_condition" form:"coating_condition"`
	CoatingDamageLevel string  `json:"coating_damage_level" form:"coating_damage_level"`
	CPStatus           string  `json:"cp_status" form:"cp_status"`
	CPPotentialMV      float64 `json:"cp_potential_mv" form:"cp_potential_mv"`
	OneCallSystem      string  `json:"one_call_system" form:"one_call_system"`
	// Component Composition (Section A)
	CO2Content             float64 `json:"co2_content" form:"co2_content"`
	H2SContent             float64 `json:"h2s_content" form:"h2s_content"`
	H2OContent             float64 `json:"h2o_content" form:"h2o_content"`
	N2Content              float64 `json:"n2_content" form:"n2_content"`
	COContent              float64 `json:"co_content" form:"co_content"`
	CO2PartialPressurePSIG float64 `json:"co2_partial_pressure_psig" form:"co2_partial_pressure_psig"`
	H2SPartialPressurePSIG float64 `json:"h2s_partial_pressure_psig" form:"h2s_partial_pressure_psig"`
	H2SPpm                 string  `json:"h2s_ppm" form:"h2s_ppm"`
	// Corrosion Indicators (Section B)
	PHLevel                   string  `json:"ph_level" form:"ph_level"`
	ChlorideContent           int     `json:"chloride_content" form:"chloride_content"`
	FluidCorrosivityMPY       string  `json:"fluid_corrosivity_mpy" form:"fluid_corrosivity_mpy"`
	InhibitorEffectiveness    string  `json:"inhibitor_effectiveness" form:"inhibitor_effectiveness"`
	BiocideTreatment          string  `json:"biocide_treatment" form:"biocide_treatment"`
	CorrosionMonitoringResult string  `json:"corrosion_monitoring_result" form:"corrosion_monitoring_result"`
	WallThicknessRatio        float64 `json:"wall_thickness_ratio" form:"wall_thickness_ratio"`
	// Operating Condition (Section C)
	Fluida                string  `json:"fluida" form:"fluida"`
	Phase                 string  `json:"phase" form:"phase"`
	PWHTStatus            string  `json:"pwht_status" form:"pwht_status"`
	WeldJointType         string  `json:"weld_joint_type" form:"weld_joint_type"`
	PressureCycleCount    float64 `json:"pressure_cycle_count" form:"pressure_cycle_count"`
	PressureRangePct      float64 `json:"pressure_range_pct" form:"pressure_range_pct"`
	SMYSUtilizationPct    float64 `json:"smys_utilization_pct" form:"smys_utilization_pct"`
	FlowVelocityCondition string  `json:"flow_velocity_condition" form:"flow_velocity_condition"`
	SolidContent          string  `json:"solid_content" form:"solid_content"`
	// Previous Equipment Condition (Section D)
	PrevExtCorrosion      string `json:"prev_ext_corrosion" form:"prev_ext_corrosion"`
	ConfExtCorrosion      string `json:"conf_ext_corrosion" form:"conf_ext_corrosion"`
	PrevIntThinning       string `json:"prev_int_thinning" form:"prev_int_thinning"`
	ConfIntThinning       string `json:"conf_int_thinning" form:"conf_int_thinning"`
	PrevIntCracking       string `json:"prev_int_cracking" form:"prev_int_cracking"`
	ConfIntCracking       string `json:"conf_int_cracking" form:"conf_int_cracking"`
	PrevLocIntCorrosion   string `json:"prev_loc_int_corrosion" form:"prev_loc_int_corrosion"`
	ConfLocIntCorrosion   string `json:"conf_loc_int_corrosion" form:"conf_loc_int_corrosion"`
	InsulationCondition   string `json:"insulation_condition" form:"insulation_condition"`
	InsulationDamageLevel string `json:"insulation_damage_level" form:"insulation_damage_level"`
	ExtCoatingCondition   string `json:"ext_coating_condition" form:"ext_coating_condition"`
	ExtCoatingDamageLevel string `json:"ext_coating_damage_level" form:"ext_coating_damage_level"`
	// Cracking Indicators (Section E)
	EnvExtCracking string `json:"env_ext_cracking" form:"env_ext_cracking"`
	// Consequence Factors (Section G â€” retained with updates)
	BuildingCountInsidePIR      int     `json:"building_count_inside_pir" form:"building_count_inside_pir"`
	ClassLocation               string  `json:"class_location" form:"class_location"`
	FlowRate                    float64 `json:"flow_rate" form:"flow_rate"`
	DetectionTimeHours          float64 `json:"detection_time_hours" form:"detection_time_hours"`
	SegmentLengthBetweenValvesM float64 `json:"segment_length_between_valves_m" form:"segment_length_between_valves_m"`
	EnvironmentalSensitivity    string  `json:"environmental_sensitivity" form:"environmental_sensitivity"`
	NearbySensitiveReceptor     bool    `json:"nearby_sensitive_receptor" form:"nearby_sensitive_receptor"`
	IsolationValveAvailable     bool    `json:"isolation_valve_available" form:"isolation_valve_available"`
	ConsequenceArea             float64 `json:"consequence_area" form:"consequence_area"`
	ConsequenceFinancial        float64 `json:"consequence_financial" form:"consequence_financial"`
	PoFCategory                 string  `json:"pof_category" form:"pof_category"`
	CoFCategory                 string  `json:"cof_category" form:"cof_category"`
	RiskRanking                 string  `json:"risk_ranking" form:"risk_ranking"`
	ConsequenceBasis            string  `json:"consequence_basis" form:"consequence_basis"`
	ProbabilityBasis            string  `json:"probability_basis" form:"probability_basis"`
	EngineeringNotes            string  `json:"engineering_notes" form:"engineering_notes"`
	RequiresConfirmation        bool    `json:"requires_confirmation" form:"requires_confirmation"`
	ConfirmationTODOReason      string  `json:"confirmation_todo_reason" form:"confirmation_todo_reason"`
	// Deprecated fields preserved for backward compatibility with saved assessments
	// Do not use in new scoring logic.
	WaterContent           string `json:"water_content,omitempty" form:"water_content"`
	FluidCorrosivity       string `json:"fluid_corrosivity,omitempty" form:"fluid_corrosivity"`
	CO2H2SPresence         string `json:"co2_h2s_presence,omitempty" form:"co2_h2s_presence"`
	MICRisk                string `json:"mic_risk,omitempty" form:"mic_risk"`
	WallThicknessCondition string `json:"wall_thickness_condition,omitempty" form:"wall_thickness_condition"`
	EmergencyResponse      string `json:"emergency_response,omitempty" form:"emergency_response"`
}

type PipelineOilInspectionPoint struct {
	InspectionPoint     string       `json:"inspection_point" form:"inspection_point"`
	LocationClass       string       `json:"location_class" form:"location_class"`
	InstallationType    string       `json:"installation_type" form:"installation_type"`
	NominalThicknessMM  float64      `json:"nominal_thickness_mm" form:"nominal_thickness_mm"`
	RequiredThicknessMM float64      `json:"required_thickness_mm" form:"required_thickness_mm"`
	ActualThicknessMM   float64      `json:"actual_thickness_mm" form:"actual_thickness_mm"`
	MeasuredYear        FlexibleYear `json:"measured_year" form:"measured_year"`
}

type PipelineOilResult struct {
	FormulaVersion              string                          `json:"formula_version"`
	CalculatedAt                time.Time                       `json:"calculated_at"`
	DesignTemperatureC          float64                         `json:"design_temperature_c"`
	OperatingTemperatureC       float64                         `json:"operating_temperature_c"`
	PipeLengthFt                float64                         `json:"pipe_length_ft"`
	OutsideDiameterMM           float64                         `json:"outside_diameter_mm"`
	AllowanceMM                 float64                         `json:"allowance_mm"`
	NominalWallThicknessIn      float64                         `json:"nominal_wall_thickness_in"`
	DesignPressureKgCM2         float64                         `json:"design_pressure_kg_cm2"`
	OperatingPressureKgCM2      float64                         `json:"operating_pressure_kg_cm2"`
	SMYSKgCM2                   float64                         `json:"smys_kg_cm2"`
	MaterialStressKgCM2         float64                         `json:"material_stress_kg_cm2"`
	RequiredThicknessIn         float64                         `json:"required_thickness_in"`
	RequiredThicknessMM         float64                         `json:"required_thickness_mm"`
	SummaryRequiredThicknessIn  float64                         `json:"summary_required_thickness_in"`
	SummaryRequiredThicknessMM  float64                         `json:"summary_required_thickness_mm"`
	MinimumActualThicknessMM    float64                         `json:"minimum_actual_thickness_mm"`
	HighestCorrosionRateMMYear  float64                         `json:"highest_corrosion_rate_mm_year"`
	RemainingLifeYears          float64                         `json:"remaining_life_years"`
	HighestHoopStressPsi        float64                         `json:"highest_hoop_stress_psi"`
	HighestHoopStressKgCM2      float64                         `json:"highest_hoop_stress_kg_cm2"`
	HighestHoopStressPercentSMY float64                         `json:"highest_hoop_stress_percent_smys"`
	LowestMAOPPsi               float64                         `json:"lowest_maop_psi"`
	LowestMAOPKgCM2             float64                         `json:"lowest_maop_kg_cm2"`
	RequiredThicknessStatus     string                          `json:"required_thickness_status"`
	HoopStressStatus            string                          `json:"hoop_stress_status"`
	MAOPStatus                  string                          `json:"maop_status"`
	GenericFailureFrequency     float64                         `json:"generic_failure_frequency"`
	ManagementSystemScore       float64                         `json:"management_system_score"`
	ManagementSystemFactor      float64                         `json:"management_system_factor"`
	DamageFactor                float64                         `json:"damage_factor"`
	SelectedDamageMechanism     string                          `json:"selected_damage_mechanism"`
	DamageMechanismResults      []PipelineDamageMechanismResult `json:"damage_mechanism_results"`
	InspectionPlanResults       []PipelineInspectionPlanResult  `json:"inspection_plan_results"`
	ThirdPartyDamageFactor      float64                         `json:"third_party_damage_factor"`
	ExternalCorrosionFactor     float64                         `json:"external_corrosion_factor"`
	InternalCorrosionFactor     float64                         `json:"internal_corrosion_factor"`
	GoverningDamageFactor       float64                         `json:"governing_damage_factor"`
	GoverningDamageMechanism    string                          `json:"governing_damage_mechanism"`
	PoFValue                    float64                         `json:"pof_value"`
	CoFValue                    float64                         `json:"cof_value"`
	PIRFeet                     float64                         `json:"pir_feet"`
	SpillVolume                 float64                         `json:"spill_volume"`
	AdjustedSpillVolume         float64                         `json:"adjusted_spill_volume"`
	RiskValue                   float64                         `json:"risk_value"`
	PoF                         string                          `json:"pof"`
	CoF                         string                          `json:"cof"`
	FinalRiskCode               string                          `json:"final_risk_code"`
	FinalRiskLevel              string                          `json:"final_risk_level"`
	RiskRanking                 string                          `json:"risk_ranking"`
	InspectionEffectiveness     string                          `json:"inspection_effectiveness"`
	InspectionResult            string                          `json:"inspection_result"`
	Recommendation              string                          `json:"recommendation"`
	RecommendationSource        string                          `json:"recommendation_source"`
	RecommendationRuleName      string                          `json:"recommendation_rule_name"`
	RecommendationGroups        PipelineAdvisoryGroups          `json:"recommendation_groups"`
	PointResults                []PipelineOilPointResult        `json:"point_results"`
	FormulaTrace                []PipelineOilFormulaTrace       `json:"formula_trace"`
	TODOEngineeringConfirmation []string                        `json:"todo_engineering_confirmation"`
}

type PipelineAdvisoryGroups struct {
	ImmediateActions   []string `json:"immediate_actions"`
	InspectionMonitor  []string `json:"inspection_monitoring"`
	LongTermMitigation []string `json:"long_term_mitigation"`
}

type PipelineDamageMechanismResult struct {
	Code                  string                 `json:"code"`
	Label                 string                 `json:"label"`
	Category              string                 `json:"category"`
	Severity              string                 `json:"severity"`
	Score                 float64                `json:"score"`
	InspectionEffectivity string                 `json:"inspection_effectivity"`
	Source                string                 `json:"source"`
	Formula               string                 `json:"formula"`
	TriggerInputs         []PipelineTriggerInput `json:"trigger_inputs"`
}

type PipelineInspectionPlanInput struct {
	NonIntrusiveMethod string `json:"non_intrusive_method"`
}

type PipelineInspectionPlanResult struct {
	Code                       string `json:"code"`
	Label                      string `json:"label"`
	Severity                   string `json:"severity"`
	NonIntrusiveMethod         string `json:"non_intrusive_method"`
	NonIntrusiveEffectivity    string `json:"non_intrusive_effectivity"`
	NonIntrusiveIntervalMonths int    `json:"non_intrusive_interval_months"`
	Source                     string `json:"source"`
}

type PipelineOilPointResult struct {
	InspectionPoint       string  `json:"inspection_point"`
	NominalThicknessMM    float64 `json:"nominal_thickness_mm"`
	RequiredThicknessIn   float64 `json:"required_thickness_in"`
	RequiredThicknessMM   float64 `json:"required_thickness_mm"`
	MinimumThicknessMM    float64 `json:"minimum_thickness_mm"`
	AppraisalThicknessIn  float64 `json:"appraisal_thickness_in"`
	AppraisalThicknessMM  float64 `json:"appraisal_thickness_mm"`
	ActualThicknessIn     float64 `json:"actual_thickness_in"`
	ActualThicknessMM     float64 `json:"actual_thickness_mm"`
	RemainingThicknessMM  float64 `json:"remaining_thickness_mm"`
	CorrosionRateMMYear   float64 `json:"corrosion_rate_mm_year"`
	RemainingLifeYears    float64 `json:"remaining_life_years"`
	HoopStressPsi         float64 `json:"hoop_stress_psi"`
	MAOPPsi               float64 `json:"maop_psi"`
	ThicknessStatus       string  `json:"thickness_status"`
	HoopStressStatus      string  `json:"hoop_stress_status"`
	MAOPStatus            string  `json:"maop_status"`
	SourceInspectionPoint string  `json:"source_inspection_point"`
}

type PipelineOilFormulaTrace struct {
	FormulaName string                 `json:"formula_name"`
	ExcelRef    string                 `json:"excel_ref"`
	Expression  string                 `json:"expression"`
	Inputs      map[string]interface{} `json:"inputs"`
	Output      interface{}            `json:"output"`
	Note        string                 `json:"note"`
}

type PipelineOilAssessment struct {
	ID             int                         `json:"id"`
	Status         PipelineOilAssessmentStatus `json:"status"`
	Input          PipelineOilInput            `json:"input"`
	Result         *PipelineOilResult          `json:"result,omitempty"`
	FormulaVersion string                      `json:"formula_version"`
	CreatedAt      string                      `json:"created_at"`
	UpdatedAt      string                      `json:"updated_at"`
	CreatedBy      string                      `json:"created_by"`
	UpdatedBy      string                      `json:"updated_by"`
}

type PipelineOilValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type PipelineOilRepository struct {
	db *sql.DB
}

type PipelineOilService struct {
	repo *PipelineOilRepository
}

func NewPipelineOilRepository(db *sql.DB) *PipelineOilRepository {
	return &PipelineOilRepository{db: db}
}

func NewPipelineOilService(db *sql.DB) *PipelineOilService {
	return &PipelineOilService{repo: NewPipelineOilRepository(db)}
}

func (s *PipelineOilService) CreateDraftAssessment(input PipelineOilInput) (int, error) {
	applyPipelineOilDefaults(&input)
	if errs := ValidatePipelineOilDraft(input); len(errs) > 0 {
		return 0, fmt.Errorf("%w: %s", ErrPipelineValidation, formatPipelineValidationErrors(errs))
	}
	return s.repo.CreateAssessment(input, PipelineOilStatusDraft, nil)
}

func (s *PipelineOilService) UpdateDraftAssessment(id int, input PipelineOilInput) error {
	current, err := s.repo.GetAssessment(id)
	if err != nil {
		return err
	}
	if current.Status != PipelineOilStatusDraft {
		return ErrPipelineFinalized
	}
	applyPipelineOilDefaults(&input)
	if errs := ValidatePipelineOilDraft(input); len(errs) > 0 {
		return fmt.Errorf("%w: %s", ErrPipelineValidation, formatPipelineValidationErrors(errs))
	}
	return s.repo.UpdateAssessment(id, input, PipelineOilStatusDraft, nil)
}

func (s *PipelineOilService) CalculateAssessment(id int, input PipelineOilInput) (*PipelineOilResult, error) {
	applyPipelineOilDefaults(&input)
	result, errs := CalculatePipelineOil(input)
	if len(errs) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrPipelineValidation, formatPipelineValidationErrors(errs))
	}
	return result, s.repo.SaveCalculationResult(id, input, result)
}

func (s *PipelineOilService) PreviewAssessment(input PipelineOilInput) (*PipelineOilResult, error) {
	applyPipelineOilDefaults(&input)
	result, errs := CalculatePipelineOil(input)
	if len(errs) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrPipelineValidation, formatPipelineValidationErrors(errs))
	}
	return result, nil
}

func (s *PipelineOilService) GetAssessmentDetail(id int) (*PipelineOilAssessment, error) {
	return s.repo.GetAssessment(id)
}

func (s *PipelineOilService) ListAssessments() ([]PipelineOilAssessment, error) {
	return s.repo.ListAssessments()
}

func (s *PipelineOilService) ArchiveAssessment(id int) error {
	return s.repo.ArchiveAssessment(id)
}

func (r *PipelineOilRepository) CreateAssessment(input PipelineOilInput, status PipelineOilAssessmentStatus, result *PipelineOilResult) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	inputJSON, resultJSON, traceJSON, snapshotJSON, err := pipelineOilJSONPayloads(input, result)
	if err != nil {
		return 0, err
	}

	res, err := tx.Exec(`
		INSERT INTO pipeline_oil_assessments (
			status, report_no, line_identification, owner_user, location, service,
			assessment_by, formula_version, input_json, result_json, formula_trace_json,
			snapshot_json, created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		status, input.ReportNo, input.LineIdentification, input.OwnerUser, input.Location, input.Service,
		input.AssessmentBy, PipelineOilFormulaVersion, inputJSON, resultJSON, traceJSON, snapshotJSON,
		input.AssessmentBy, input.AssessmentBy,
	)
	if err != nil {
		return 0, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return int(id64), nil
}

func (r *PipelineOilRepository) UpdateAssessment(id int, input PipelineOilInput, status PipelineOilAssessmentStatus, result *PipelineOilResult) error {
	inputJSON, resultJSON, traceJSON, snapshotJSON, err := pipelineOilJSONPayloads(input, result)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
		UPDATE pipeline_oil_assessments
		SET status=?, report_no=?, line_identification=?, owner_user=?, location=?, service=?,
			assessment_by=?, formula_version=?, input_json=?, result_json=?, formula_trace_json=?,
			snapshot_json=?, updated_by=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=? AND status <> ?`,
		status, input.ReportNo, input.LineIdentification, input.OwnerUser, input.Location, input.Service,
		input.AssessmentBy, PipelineOilFormulaVersion, inputJSON, resultJSON, traceJSON, snapshotJSON,
		input.AssessmentBy, id, PipelineOilStatusArchived,
	)
	return err
}

func (r *PipelineOilRepository) SaveCalculationResult(id int, input PipelineOilInput, result *PipelineOilResult) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	inputJSON, resultJSON, traceJSON, snapshotJSON, err := pipelineOilJSONPayloads(input, result)
	if err != nil {
		return err
	}
	res, err := tx.Exec(`
		UPDATE pipeline_oil_assessments
		SET status=?, report_no=?, line_identification=?, owner_user=?, location=?, service=?,
			assessment_by=?, formula_version=?, input_json=?, result_json=?, formula_trace_json=?,
			snapshot_json=?, updated_by=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=? AND status <> ?`,
		PipelineOilStatusCalculated, input.ReportNo, input.LineIdentification, input.OwnerUser,
		input.Location, input.Service, input.AssessmentBy, PipelineOilFormulaVersion, inputJSON,
		resultJSON, traceJSON, snapshotJSON, input.AssessmentBy, id, PipelineOilStatusArchived,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrPipelineFinalized
	}
	return tx.Commit()
}

func (r *PipelineOilRepository) GetAssessment(id int) (*PipelineOilAssessment, error) {
	var a PipelineOilAssessment
	var inputJSON, resultJSON sql.NullString
	query := `
		SELECT id, status, formula_version, input_json, result_json, created_at, updated_at,
			COALESCE(created_by, ''), COALESCE(updated_by, '')
		FROM pipeline_oil_assessments
		WHERE id=?`
	err := r.db.QueryRow(query, id).Scan(
		&a.ID, &a.Status, &a.FormulaVersion, &inputJSON, &resultJSON,
		&a.CreatedAt, &a.UpdatedAt, &a.CreatedBy, &a.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	if inputJSON.Valid && inputJSON.String != "" {
		if err = json.Unmarshal([]byte(inputJSON.String), &a.Input); err != nil {
			return nil, err
		}
	}
	if resultJSON.Valid && resultJSON.String != "" {
		var result PipelineOilResult
		if err = json.Unmarshal([]byte(resultJSON.String), &result); err != nil {
			return nil, err
		}
		a.Result = &result
	}
	return &a, nil
}

func (r *PipelineOilRepository) ListAssessments() ([]PipelineOilAssessment, error) {
	rows, err := r.db.Query(`
		SELECT id, status, formula_version, input_json, result_json, created_at, updated_at,
			COALESCE(created_by, ''), COALESCE(updated_by, '')
		FROM pipeline_oil_assessments
		WHERE status <> ?
		ORDER BY id DESC`, PipelineOilStatusArchived)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PipelineOilAssessment
	for rows.Next() {
		var a PipelineOilAssessment
		var inputJSON, resultJSON sql.NullString
		if err = rows.Scan(&a.ID, &a.Status, &a.FormulaVersion, &inputJSON, &resultJSON, &a.CreatedAt, &a.UpdatedAt, &a.CreatedBy, &a.UpdatedBy); err != nil {
			return nil, err
		}
		if inputJSON.Valid && inputJSON.String != "" {
			if err = json.Unmarshal([]byte(inputJSON.String), &a.Input); err != nil {
				return nil, err
			}
		}
		if resultJSON.Valid && resultJSON.String != "" {
			var result PipelineOilResult
			if err = json.Unmarshal([]byte(resultJSON.String), &result); err != nil {
				return nil, err
			}
			a.Result = &result
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

func (r *PipelineOilRepository) ArchiveAssessment(id int) error {
	_, err := r.db.Exec(`UPDATE pipeline_oil_assessments SET status=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, PipelineOilStatusArchived, id)
	return err
}

func CalculatePipelineOil(input PipelineOilInput) (*PipelineOilResult, []PipelineOilValidationError) {
	applyPipelineOilDefaults(&input)
	errs := ValidatePipelineOilCalculation(input)
	if len(errs) > 0 {
		return nil, errs
	}

	result := &PipelineOilResult{
		FormulaVersion:          PipelineOilFormulaVersion,
		CalculatedAt:            time.Now(),
		DesignTemperatureC:      (5.0 / 9.0) * (input.DesignTemperatureF - 32),
		OperatingTemperatureC:   (5.0 / 9.0) * (input.OperatingTemperatureF - 32),
		PipeLengthFt:            ((input.PipeLengthM * 100) / 2.54) / 12,
		OutsideDiameterMM:       input.OutsideDiameterIn * 25.4,
		AllowanceMM:             input.AllowanceIn * 25.4,
		NominalWallThicknessIn:  input.NominalWallThicknessMM / 25.4,
		DesignPressureKgCM2:     psiToKgCM2(input.InternalDesignPressurePsi),
		OperatingPressureKgCM2:  psiToKgCM2(input.OperatingPressurePsi),
		SMYSKgCM2:               psiToKgCM2(input.SMYSPsi),
		MaterialStressKgCM2:     psiToKgCM2(input.MaterialStressPsi),
		PoF:                     "",
		CoF:                     "",
		RiskRanking:             "",
		InspectionEffectiveness: input.RiskInput.InspectionEffectivity,
		TODOEngineeringConfirmation: []string{
			"Probability of failure formula is not present in workbook.",
			"Consequence of failure formula is not present in workbook.",
			"Risk ranking matrix/formula is not present in workbook.",
			"Workbook contains #REF! formulas in sheet '6 Verification' and downstream '2 Data' ranges for additional inspection rows.",
			"API 581 exact GFF tables, damage-factor tables, CoF category thresholds, and risk matrix thresholds require licensed engineering data.",
		},
	}

	applyPipelineIndexRisk(input, result)

	tReqIn := requiredThicknessInForInput(input)
	result.RequiredThicknessIn = tReqIn
	result.RequiredThicknessMM = tReqIn * 25.4
	requiredThicknessExpression := "((P*D)/(2*F*E*SMYS))+c"
	requiredThicknessInputs := map[string]interface{}{"P": input.InternalDesignPressurePsi, "D": input.OutsideDiameterIn, "F": input.DesignFactor, "E": input.QualityFactor, "SMYS": input.SMYSPsi, "c": input.AllowanceIn}
	if isASMECode(input.ApplicableCode, "B31.3") {
		requiredThicknessExpression = "(P*D)/(2*(S*E*W+P*Y))+c"
		requiredThicknessInputs = map[string]interface{}{"P": input.InternalDesignPressurePsi, "D": input.OutsideDiameterIn, "S": input.MaterialStressPsi, "E": input.QualityFactor, "W": input.WeldJointStrengthFactor, "Y": input.DesignFactor, "c": input.AllowanceIn}
	} else if isASMECode(input.ApplicableCode, "B31.8") {
		requiredThicknessExpression = "((P*D)/(2*F*E*T*SMYS))+c"
		requiredThicknessInputs = map[string]interface{}{"P": input.InternalDesignPressurePsi, "D": input.OutsideDiameterIn, "F": input.DesignFactor, "E": input.QualityFactor, "T": input.TemperatureDeratingFactor, "SMYS": input.SMYSPsi, "c": input.AllowanceIn}
	}
	result.FormulaTrace = append(result.FormulaTrace,
		trace("pipe_length_ft", "Input!G17", "((D17*100)/2.54)/12", map[string]interface{}{"pipe_length_m": input.PipeLengthM}, result.PipeLengthFt, ""),
		trace("design_temperature_c", "Input!G23", "(5/9)*(D23-32)", map[string]interface{}{"design_temperature_f": input.DesignTemperatureF}, result.DesignTemperatureC, ""),
		trace("outside_diameter_mm", "Input!G36", "D36*25.4", map[string]interface{}{"outside_diameter_in": input.OutsideDiameterIn}, result.OutsideDiameterMM, ""),
		trace("allowance_mm", "Input!G29", "D29*25.4", map[string]interface{}{"allowance_in": input.AllowanceIn}, result.AllowanceMM, ""),
		trace("required_thickness", "7 Appraisal!J63", requiredThicknessExpression, requiredThicknessInputs, result.RequiredThicknessIn, ""),
		trace("design_pressure_kg_cm2", "7 Appraisal!N25", "design_pressure_psi/14.223", map[string]interface{}{"design_pressure_psi": input.InternalDesignPressurePsi}, result.DesignPressureKgCM2, ""),
		trace("operating_pressure_kg_cm2", "7 Appraisal!N26", "operating_pressure_psi/14.223", map[string]interface{}{"operating_pressure_psi": input.OperatingPressurePsi}, result.OperatingPressureKgCM2, "Workbook cached value is #VALUE!, formula chain indicates psi to kg/cm2 conversion."),
		trace("smys_kg_cm2", "7 Appraisal!N37", "smys_psi/14.223", map[string]interface{}{"smys_psi": input.SMYSPsi}, result.SMYSKgCM2, ""),
		trace("material_stress_kg_cm2", "7 Appraisal!N38", "material_stress_psi/14.223", map[string]interface{}{"material_stress_psi": input.MaterialStressPsi}, result.MaterialStressKgCM2, ""),
	)

	result.MinimumActualThicknessMM = math.MaxFloat64
	result.LowestMAOPPsi = math.MaxFloat64
	result.SummaryRequiredThicknessIn = math.MaxFloat64
	for i, point := range input.InspectionPoints {
		actualIn := point.ActualThicknessMM / 25.4
		appraisalRequiredIn := tReqIn
		if i > 0 {
			appraisalRequiredIn = roundToPlaces(tReqIn, 3)
		}
		appraisalRequiredMM := appraisalRequiredIn * 25.4
		cr := corrosionRateMMYear(point.NominalThicknessMM, point.ActualThicknessMM, input.YearUsed, point.MeasuredYear)
		remainingLifeBasisMM := point.RequiredThicknessMM
		if remainingLifeBasisMM <= 0 {
			remainingLifeBasisMM = appraisalRequiredMM
		}
		rl := remainingLifeYears(point.ActualThicknessMM, remainingLifeBasisMM, cr)
		hs := hoopStressPsi(input.InternalDesignPressurePsi, input.OutsideDiameterIn, actualIn)
		maop := maopPsiForInput(input, actualIn)

		allowableStress := input.SMYSPsi * input.DesignFactor * input.QualityFactor * input.TemperatureDeratingFactor
		if isASMECode(input.ApplicableCode, "B31.3") {
			allowableStress = input.MaterialStressPsi * input.QualityFactor * input.WeldJointStrengthFactor
		}

		pr := PipelineOilPointResult{
			InspectionPoint:       point.InspectionPoint,
			NominalThicknessMM:    point.NominalThicknessMM,
			RequiredThicknessIn:   result.RequiredThicknessIn,
			RequiredThicknessMM:   result.RequiredThicknessMM,
			MinimumThicknessMM:    appraisalRequiredMM,
			AppraisalThicknessIn:  appraisalRequiredIn,
			AppraisalThicknessMM:  appraisalRequiredMM,
			ActualThicknessIn:     actualIn,
			ActualThicknessMM:     point.ActualThicknessMM,
			RemainingThicknessMM:  math.Max(point.ActualThicknessMM-appraisalRequiredMM, 0),
			CorrosionRateMMYear:   cr,
			RemainingLifeYears:    rl,
			HoopStressPsi:         hs,
			MAOPPsi:               maop,
			ThicknessStatus:       acceptable(actualIn > appraisalRequiredIn),
			HoopStressStatus:      acceptable(hs <= allowableStress),
			MAOPStatus:            acceptable(maop > input.InternalDesignPressurePsi),
			SourceInspectionPoint: point.InspectionPoint,
		}
		result.PointResults = append(result.PointResults, pr)
		result.FormulaTrace = append(result.FormulaTrace,
			trace("corrosion_rate", "2 Data!O30:O32", "(nominal_thickness_mm-actual_thickness_mm)/(measured_year-year_used)", map[string]interface{}{"point": point.InspectionPoint, "nominal_thickness_mm": point.NominalThicknessMM, "actual_thickness_mm": point.ActualThicknessMM, "measured_year": string(point.MeasuredYear), "year_used": string(input.YearUsed)}, cr, ""),
			trace("remaining_life", "2 Data!S30:S32 / 7 Appraisal!R159:R160", "(actual_thickness_mm-required_thickness_mm)/corrosion_rate_mm_year", map[string]interface{}{"point": point.InspectionPoint, "actual_thickness_mm": point.ActualThicknessMM, "required_thickness_mm": remainingLifeBasisMM, "corrosion_rate_mm_year": cr}, rl, "Workbook remaining life references '2 Data' required thickness where provided."),
			trace("hoop_stress", "7 Appraisal!O90:O92", "(P*D)/(2*actual_thickness_in)", map[string]interface{}{"point": point.InspectionPoint, "P": input.InternalDesignPressurePsi, "D": input.OutsideDiameterIn, "actual_thickness_in": actualIn}, hs, ""),
			trace("maop", "7 Appraisal!O122:O124", maopExpression(input), maopInputs(input, point.InspectionPoint, actualIn), maop, ""),
		)
		result.SummaryRequiredThicknessIn = math.Min(result.SummaryRequiredThicknessIn, appraisalRequiredIn)
		result.MinimumActualThicknessMM = math.Min(result.MinimumActualThicknessMM, point.ActualThicknessMM)
		result.HighestCorrosionRateMMYear = math.Max(result.HighestCorrosionRateMMYear, cr)
		result.RemainingLifeYears = minPositive(result.RemainingLifeYears, rl)
		result.HighestHoopStressPsi = math.Max(result.HighestHoopStressPsi, hs)
		result.LowestMAOPPsi = math.Min(result.LowestMAOPPsi, maop)
	}
	if result.MinimumActualThicknessMM == math.MaxFloat64 {
		result.MinimumActualThicknessMM = 0
	}
	if result.LowestMAOPPsi == math.MaxFloat64 {
		result.LowestMAOPPsi = 0
	}
	if result.SummaryRequiredThicknessIn == math.MaxFloat64 {
		result.SummaryRequiredThicknessIn = 0
	}
	result.SummaryRequiredThicknessMM = result.SummaryRequiredThicknessIn * 25.4
	result.HighestHoopStressKgCM2 = psiToKgCM2(result.HighestHoopStressPsi)
	result.LowestMAOPKgCM2 = psiToKgCM2(result.LowestMAOPPsi)
	if input.SMYSPsi > 0 {
		result.HighestHoopStressPercentSMY = (result.HighestHoopStressPsi / input.SMYSPsi) * 100
	}
	result.RequiredThicknessStatus = aggregateStatus(result.PointResults, func(p PipelineOilPointResult) string { return p.ThicknessStatus })
	result.HoopStressStatus = aggregateStatus(result.PointResults, func(p PipelineOilPointResult) string { return p.HoopStressStatus })
	result.MAOPStatus = aggregateStatus(result.PointResults, func(p PipelineOilPointResult) string { return p.MAOPStatus })
	result.InspectionResult = aggregateInspectionResult(result.RequiredThicknessStatus, result.HoopStressStatus, result.MAOPStatus)
	result.FormulaTrace = append(result.FormulaTrace,
		trace("highest_hoop_stress", "7 Appraisal!H106", "MAX(O90:Q104)", map[string]interface{}{"point_results": len(result.PointResults)}, result.HighestHoopStressPsi, ""),
		trace("highest_hoop_stress_kg_cm2", "7 Appraisal!L106", "H106/14.223", map[string]interface{}{"highest_hoop_stress_psi": result.HighestHoopStressPsi}, result.HighestHoopStressKgCM2, ""),
		trace("highest_hoop_stress_percent_smys", "7 Appraisal!H107", "(H106/K37)*100", map[string]interface{}{"highest_hoop_stress_psi": result.HighestHoopStressPsi, "smys_psi": input.SMYSPsi}, result.HighestHoopStressPercentSMY, ""),
		trace("lowest_maop", "7 Appraisal!H138", "MIN(O122:O136)", map[string]interface{}{"point_results": len(result.PointResults)}, result.LowestMAOPPsi, ""),
		trace("lowest_maop_kg_cm2", "7 Appraisal!L138", "H138/14.223", map[string]interface{}{"lowest_maop_psi": result.LowestMAOPPsi}, result.LowestMAOPKgCM2, ""),
		trace("summary_required_thickness", "7 Appraisal!O195/R195", "MIN(K159:M160), O195/25.4", map[string]interface{}{"summary_required_thickness_in": result.SummaryRequiredThicknessIn}, result.SummaryRequiredThicknessMM, ""),
		trace("summary_remaining_life", "7 Appraisal!O201", "N165", map[string]interface{}{"remaining_life_years": result.RemainingLifeYears}, result.RemainingLifeYears, ""),
	)
	return result, nil
}

func applyPipelineIndexRisk(input PipelineOilInput, result *PipelineOilResult) {
	result.DamageMechanismResults = calculatePipelineDamageMechanismResults(input)
	result.InspectionPlanResults = calculatePipelineInspectionPlanResults(input, result.DamageMechanismResults)
	pofValue, fms, governingDF, governingDriver := calculatePipelinePoF(input.RiskInput.GenericFailureFrequency, input.RiskInput.ManagementSystemScore, result.DamageMechanismResults)

	var dfTPD, dfExternal, dfInternal float64
	for _, dm := range result.DamageMechanismResults {
		score := dm.Score
		if score == 0 {
			score = 1.0
		}
		switch dm.Code {
		case "third_party_mechanical_damage":
			dfTPD = score
		case "external_corrosion":
			dfExternal = score
		case "internal_corrosion":
			dfInternal = score
		}
	}

	result.GenericFailureFrequency = input.RiskInput.GenericFailureFrequency
	result.ManagementSystemScore = input.RiskInput.ManagementSystemScore
	result.ManagementSystemFactor = fms
	result.SelectedDamageMechanism = PipelineDamageMechanismLabel(input.RiskInput.DamageMechanism)
	result.ThirdPartyDamageFactor = dfTPD
	result.ExternalCorrosionFactor = dfExternal
	result.InternalCorrosionFactor = dfInternal
	result.GoverningDamageFactor = governingDF
	result.GoverningDamageMechanism = governingDriver
	result.DamageFactor = governingDF
	result.PoFValue = pofValue
	result.PoF = pofCategory(pofValue)

	if isGasService(input.Service) {
		cofCategory, pir := calculateGasCoF(input.OutsideDiameterIn, input.OperatingPressurePsi, input.RiskInput.BuildingCountInsidePIR, input.RiskInput.ClassLocation)
		result.PIRFeet = pir
		result.CoF = cofCategory
		result.CoFValue = cofNumeric(cofCategory)
	} else {
		cofCategory, spillVolume, adjustedSpillVolume := calculateLiquidCoF(input.RiskInput.FlowRate, input.RiskInput.DetectionTimeHours, input.OutsideDiameterIn, input.RiskInput.SegmentLengthBetweenValvesM, input.RiskInput.EnvironmentalSensitivity, input.RiskInput.NearbySensitiveReceptor, input.RiskInput.IsolationValveAvailable)
		result.SpillVolume = spillVolume
		result.AdjustedSpillVolume = adjustedSpillVolume
		result.CoF = cofCategory
		result.CoFValue = adjustedSpillVolume
	}

	result.RiskValue = pofNumeric(result.PoF) * cofNumeric(result.CoF)
	result.FinalRiskCode, result.FinalRiskLevel = calculatePipelineRiskRanking(result.PoF, result.CoF)
	result.RiskRanking = result.FinalRiskCode + " - " + result.FinalRiskLevel
	advisory := generatePipelineEngineeringAdvisory(result, input)
	result.Recommendation = advisory.Recommendation
	result.RecommendationGroups = advisory.Groups
	result.RecommendationSource = advisory.Source
	result.RecommendationRuleName = advisory.RuleName
	result.FormulaTrace = append(result.FormulaTrace,
		trace("pipeline_damage_mechanism_screening", PipelineDamageMechanismSource, "Each mechanism screened from Pipeline-specific factor inputs.", map[string]interface{}{"mechanism_count": len(result.DamageMechanismResults)}, result.DamageMechanismResults, ""),
		trace("pipeline_inspection_scope_interval_method", PipelineDamageMechanismSource, "Inspection method and interval generated per damage mechanism from severity and effectivity.", map[string]interface{}{"mechanism_count": len(result.InspectionPlanResults)}, result.InspectionPlanResults, ""),
		trace("pipeline_pof", "Pipeline PoF", "PoF = GFF Ã— governing DM score Ã— FMS", map[string]interface{}{"gff": input.RiskInput.GenericFailureFrequency, "governing_dm_score": governingDF, "fms": fms}, pofValue, ""),
		trace("pipeline_risk_ranking", "Pipeline MVP", "Risk = PoF Category Ã— CoF Category", map[string]interface{}{"pof_category": result.PoF, "cof_category": result.CoF}, result.RiskRanking, ""),
		trace("pipeline_damage_mechanism_metadata", PipelineDamageMechanismSource, "Selected pipeline damage mechanism stored as classification metadata.", map[string]interface{}{"selected_damage_mechanism": PipelineDamageMechanismLabel(input.RiskInput.DamageMechanism)}, PipelineDamageMechanismLabel(input.RiskInput.DamageMechanism), ""),
		trace("pipeline_engineering_advisory", result.RecommendationSource, result.RecommendationRuleName, map[string]interface{}{"risk_level": result.FinalRiskLevel, "cof": result.CoF, "governing_driver": result.GoverningDamageMechanism, "selected_damage_mechanism": PipelineDamageMechanismLabel(input.RiskInput.DamageMechanism)}, result.Recommendation, "System-generated advisory."),
	)
}

func calculatePipelineInspectionPlanResults(input PipelineOilInput, mechanisms []PipelineDamageMechanismResult) []PipelineInspectionPlanResult {
	var results []PipelineInspectionPlanResult
	for _, mechanism := range mechanisms {
		plan := input.RiskInput.InspectionPlanByDM[mechanism.Code]
		if strings.TrimSpace(plan.NonIntrusiveMethod) == "" {
			plan.NonIntrusiveMethod = defaultPipelineNonIntrusiveMethod(mechanism.Code)
		}
		nonEff := pipelineMethodEffectivity(plan.NonIntrusiveMethod)
		results = append(results, PipelineInspectionPlanResult{
			Code:                       mechanism.Code,
			Label:                      mechanism.Label,
			Severity:                   mechanism.Severity,
			NonIntrusiveMethod:         plan.NonIntrusiveMethod,
			NonIntrusiveEffectivity:    nonEff,
			NonIntrusiveIntervalMonths: pipelineInspectionIntervalMonths(mechanism.Severity, nonEff, false),
			Source:                     PipelineDamageMechanismSource,
		})
	}
	return results
}

func defaultPipelineNonIntrusiveMethod(code string) string {
	switch code {
	case "external_corrosion", "coating_degradation":
		return "Visual + CP / Coating Survey"
	case "third_party_mechanical_damage":
		return "ROW Patrol + Visual Survey"
	case "cracking_scc_fatigue":
		return "Shear Wave Ultrasonic Testing"
	default:
		return "Wall Thickness measurement by UT"
	}
}

func pipelineMethodEffectivity(method string) string {
	method = strings.ToLower(strings.TrimSpace(method))
	switch {
	case method == "" || method == "none":
		return "None"
	case strings.Contains(method, "vie") || strings.Contains(method, "direct") || strings.Contains(method, "mpt") || strings.Contains(method, "dpt"):
		return "High"
	case strings.Contains(method, "ut") || strings.Contains(method, "ultrasonic") || strings.Contains(method, "cp") || strings.Contains(method, "coating"):
		return "Medium"
	default:
		return "Low"
	}
}

func pipelineInspectionIntervalMonths(severity, effectivity string, intrusive bool) int {
	base := map[string]int{"High": 12, "Moderate": 24, "Low": 48, "NOT": 60}[severity]
	if base == 0 {
		base = 36
	}
	multiplier := map[string]float64{"High": 1.25, "Medium": 1.0, "Low": 0.75, "None": 0.5}[effectivity]
	if multiplier == 0 {
		multiplier = 1
	}
	if intrusive {
		multiplier *= 2
	}
	months := int(math.Round(float64(base) * multiplier))
	if months < 6 {
		return 6
	}
	if months > 120 {
		return 120
	}
	return months
}

func calculatePipelineDamageMechanismResults(input PipelineOilInput) []PipelineDamageMechanismResult {
	results := []PipelineDamageMechanismResult{}
	for _, option := range PipelineDamageMechanismOptions() {
		var dmResult pipelineDMScore
		switch option.Code {
		case "external_corrosion":
			dmResult = scoreExternalCorrosion(input)
		case "coating_degradation":
			dmResult = scoreCoatingDegradation(input)
		case "third_party_mechanical_damage":
			dmResult = scoreThirdPartyDamage(input)
		case "internal_corrosion":
			dmResult = scoreInternalCorrosion(input)
		case "localized_corrosion":
			internalScore := scoreInternalCorrosion(input)
			dmResult = scoreLocalizedCorrosion(input, internalScore)
		case "erosion":
			dmResult = scoreErosion(input)
		case "erosion_corrosion":
			dmResult = scoreErosionCorrosion(input)
		case "cracking":
			dmResult = scoreCracking(input)
		case "scc":
			dmResult = scoreSCC(input)
		case "fatigue":
			dmResult = scoreFatigue(input)
		case "chemical_damage":
			dmResult = scoreChemicalDamage(input)
		default:
			dmResult = newDMScore()
			dmResult.severity = "NOT"
			dmResult.formula = "Mechanism not implemented"
		}
		effectivity := input.RiskInput.InspectionEffectivity
		if byDM := input.RiskInput.InspectionEffectivityByDM; byDM != nil {
			if selected := strings.TrimSpace(byDM[option.Code]); selected != "" {
				effectivity = selected
			}
		}
		if effectivity == "" {
			effectivity = "Representative"
		}
		results = append(results, PipelineDamageMechanismResult{
			Code:                  option.Code,
			Label:                 option.Label,
			Category:              option.Category,
			Severity:              dmResult.severity,
			Score:                 dmResult.score,
			InspectionEffectivity: effectivity,
			Source:                PipelineDamageMechanismSource,
			Formula:               dmResult.formula,
			TriggerInputs:         dmResult.triggerInputs,
		})
	}
	return results
}

func severityFromScore(score float64) string {
	switch {
	case score <= 0:
		return "NOT"
	case score < 1.5:
		return "Low"
	case score < 3.0:
		return "Moderate"
	default:
		return "High"
	}
}

func statusFromSeverity(severity string) string {
	switch severity {
	case "NOT", "Low":
		return "ACCEPTABLE"
	case "Moderate":
		return "CONDITIONALLY ACCEPTABLE"
	default:
		return "NOT ACCEPTABLE"
	}
}

type pipelineDMScore struct {
	score         float64
	severity      string
	formula       string
	triggerInputs []PipelineTriggerInput
}

func newDMScore() pipelineDMScore {
	return pipelineDMScore{score: 0, severity: "NOT", formula: "", triggerInputs: nil}
}

func (s *pipelineDMScore) addTrigger(field, value, reason string) {
	s.triggerInputs = append(s.triggerInputs, PipelineTriggerInput{
		Field:  field,
		Value:  value,
		Reason: reason,
	})
}

func (s *pipelineDMScore) addGate(field, value, reason string, passed bool) {
	status := "PASS"
	if !passed {
		status = "FAIL"
	}
	s.triggerInputs = append(s.triggerInputs, PipelineTriggerInput{
		Field:  field,
		Value:  value + " [" + status + "]",
		Reason: reason,
	})
}

func (s *pipelineDMScore) addModifier(field, value, reason string, modifier float64) {
	s.score += modifier
	s.triggerInputs = append(s.triggerInputs, PipelineTriggerInput{
		Field:  field,
		Value:  value,
		Reason: reason,
	})
}

func (s *pipelineDMScore) escalateByFinding(prevFinding, confidence, mechanismName string) {
	if prevFinding == "Finding" {
		weight := pipelineConfidenceWeight[confidence]
		if weight <= 0 {
			weight = 1.0
		}
		escalation := 1.0 * weight
		s.score += escalation
		s.triggerInputs = append(s.triggerInputs, PipelineTriggerInput{
			Field:  mechanismName + "_previous_finding",
			Value:  prevFinding + " (confidence: " + confidence + ")",
			Reason: "Previous confirmed finding for " + mechanismName + "; +1.0 severity",
		})
	} else if prevFinding == "Not Inspectable" {
		s.triggerInputs = append(s.triggerInputs, PipelineTriggerInput{
			Field:  mechanismName + "_previous_finding",
			Value:  prevFinding,
			Reason: "Not inspectable â€” no escalation applied",
		})
	}
}

func baseSeverityScore(severity string) float64 {
	switch severity {
	case "Low":
		return 1.0
	case "Moderate":
		return 2.0
	case "High":
		return 3.0
	default:
		return 0.0
	}
}

// --- Individual Mechanism Scoring Functions (Gate-Modifier-Escalation) ---

func scoreInternalCorrosion(input PipelineOilInput) pipelineDMScore {
	s := newDMScore()
	pCO2 := input.RiskInput.CO2PartialPressurePSIG
	h2o := input.RiskInput.H2OContent
	corrosivity := input.RiskInput.FluidCorrosivityMPY

	hasCO2 := pCO2 > 0
	hasWater := h2o > 0
	hasCorrosivity := corrosivity != "" && corrosivity != "<2 mpy"
	gatePassed := hasCO2 || hasWater || hasCorrosivity

	s.addGate("co2_partial_pressure_psig", fmt.Sprintf("%.2f", pCO2),
		"pCO2 screening gate (API 581 Section 6)", hasCO2)
	s.addGate("h2o_content", fmt.Sprintf("%.2f", h2o),
		"Water presence gate for electrochemical corrosion", hasWater)

	if !gatePassed {
		s.severity = "NOT"
		s.formula = "Gate: No corrosion driver present (pCO2=0, H2O=0, corrosivity low)"
		return s
	}

	if pCO2 > 0 {
		baseSev := pCO2Severity(pCO2)
		s.score = baseSeverityScore(baseSev)
		s.formula = fmt.Sprintf("Base: pCO2=%.2f psig â†’ %s (API 581)", pCO2, baseSev)
		s.addTrigger("co2_partial_pressure_psig", fmt.Sprintf("%.2f", pCO2),
			fmt.Sprintf("pCO2 %.2f psig â†’ %s per API 581 Section 6", pCO2, baseSev))
	} else {
		corrosivityScore, ok := pipelineFluidCorrosivityMPYFactors[corrosivity]
		if !ok || corrosivityScore == 0 {
			corrosivityScore = 1.0
		}
		s.score = corrosivityScore // TODO_ENGINEERING_CONFIRMATION: currently 1.0
		s.formula = "Base: fluid_corrosivity_mpy (pCO2 not available, using mpy fallback) TODO_ENGINEERING_CONFIRMATION"
		s.addTrigger("fluid_corrosivity_mpy", corrosivity,
			"Corrosivity fallback (pCO2=0) TODO_ENGINEERING_CONFIRMATION")
	}

	// Additive modifiers (all 0.0 placeholder until engineering confirmation)
	if h2o > 5.0 {
		modifier := 0.0 // TODO_ENGINEERING_CONFIRMATION: water content modifier magnitude
		s.addModifier("h2o_content", fmt.Sprintf("%.2f", h2o),
			"Water content >5 mole% TODO_ENGINEERING_CONFIRMATION", modifier)
	}

	if input.RiskInput.PHLevel != "" {
		modifier := pipelinePHSeverity[input.RiskInput.PHLevel] - 1.0 // currently 0.0
		s.addModifier("ph_level", input.RiskInput.PHLevel,
			"pH level modifier TODO_ENGINEERING_CONFIRMATION", modifier)
	}

	if input.RiskInput.InhibitorEffectiveness != "" {
		modifier := pipelineInhibitorModifiers[input.RiskInput.InhibitorEffectiveness] - 1.0 // currently 0.0
		s.addModifier("inhibitor_effectiveness", input.RiskInput.InhibitorEffectiveness,
			"Inhibitor effectiveness modifier TODO_ENGINEERING_CONFIRMATION", modifier)
	}

	if input.RiskInput.BiocideTreatment == "No" && h2o > 0 {
		// MIC gate: biocide=No AND water present is relevant
		s.addModifier("biocide_treatment", input.RiskInput.BiocideTreatment,
			"No biocide treatment with water present TODO_ENGINEERING_CONFIRMATION", 0.0)
	}

	s.escalateByFinding(input.RiskInput.PrevIntThinning, input.RiskInput.ConfIntThinning, "internal_thinning")

	s.severity = severityFromScore(s.score)
	s.triggerInputs = append(s.triggerInputs, PipelineTriggerInput{
		Field:  "internal_corrosion_result",
		Value:  s.severity,
		Reason: fmt.Sprintf("Score=%.2f â†’ %s (modifiers pending engineering confirmation)", s.score, s.severity),
	})
	return s
}

func scoreExternalCorrosion(input PipelineOilInput) pipelineDMScore {
	s := newDMScore()
	cpStatus := input.RiskInput.CPStatus
	coatingCond := input.RiskInput.CoatingCondition
	soilResist := input.RiskInput.SoilResistivity
	cpPotential := input.RiskInput.CPPotentialMV

	gatePassed := cpStatus != "normal" || coatingCond == "Damaged" || soilResist == "<1000"
	s.addGate("cp_status", cpStatus, "CP status gate for external corrosion", cpStatus != "normal")
	s.addGate("coating_condition", coatingCond, "Coating condition gate for external corrosion", coatingCond == "Damaged")
	s.addGate("soil_resistivity", soilResist, "Soil resistivity gate for external corrosion", soilResist == "<1000")

	if !gatePassed {
		s.severity = "NOT"
		s.formula = "Gate: No external corrosion driver present (CP=normal, coating=Good, soil>5000)"
		return s
	}

	// Base severity from CP/coating/soil combination
	// TODO_ENGINEERING_CONFIRMATION: all factor values are 1.0 placeholders
	cpFactor, ok := pipelineCPFactors[cpStatus]
	if !ok {
		cpFactor = 1.0
	}
	// CP potential override sourced from NACE SP0169: -850mV threshold
	if cpPotential != 0 && cpPotential > -850 && cpFactor < pipelineCPFactors["borderline"] {
		cpFactor = pipelineCPFactors["borderline"]
		s.addTrigger("cp_potential_mv", fmt.Sprintf("%.0f", cpPotential),
			"CP potential >-850mV overrides CP status to borderline (NACE SP0169)")
	}
	coatingFactor, ok := pipelineCoatingConditionFactors[coatingCond]
	if !ok {
		coatingFactor = 1.0
	}
	soilFactor, ok := pipelineSoilFactors[soilResist]
	if !ok {
		soilFactor = 1.0
	}
	// TODO_ENGINEERING_CONFIRMATION: multiplicative model replaced by additive framework
	// Currently all factors are 1.0 placeholders so score = base rate * 1.0 * 1.0 * 1.0
	s.score = input.RiskInput.BaseExternalCorrRate * soilFactor * coatingFactor * cpFactor
	// Since all factors are 1.0, score equals base rate; severity is Low for non-zero base
	if input.RiskInput.BaseExternalCorrRate > 0 {
		s.score = math.Max(s.score, 1.0) // minimum Non-NOT score when gate passes
	}

	s.formula = fmt.Sprintf("Base: BaseExternalCorrRate * Soil(%.2f) * Coating(%.2f) * CP(%.2f) TODO_ENGINEERING_CONFIRMATION", soilFactor, coatingFactor, cpFactor)
	s.addTrigger("base_external_corr_rate", fmt.Sprintf("%.4f", input.RiskInput.BaseExternalCorrRate), "Base external corrosion rate")
	s.addTrigger("soil_resistivity", soilResist, fmt.Sprintf("Soil factor=%.2f TODO_ENGINEERING_CONFIRMATION", soilFactor))
	s.addTrigger("coating_condition", coatingCond, fmt.Sprintf("Coating factor=%.2f TODO_ENGINEERING_CONFIRMATION", coatingFactor))
	s.addTrigger("cp_status", cpStatus, fmt.Sprintf("CP factor=%.2f TODO_ENGINEERING_CONFIRMATION", cpFactor))

	s.escalateByFinding(input.RiskInput.PrevExtCorrosion, input.RiskInput.ConfExtCorrosion, "external_corrosion")

	s.severity = severityFromScore(s.score)
	s.triggerInputs = append(s.triggerInputs, PipelineTriggerInput{
		Field:  "external_corrosion_result",
		Value:  s.severity,
		Reason: fmt.Sprintf("Score=%.2f â†’ %s (factors pending engineering confirmation)", s.score, s.severity),
	})
	return s
}

func scoreLocalizedCorrosion(input PipelineOilInput, internalScore pipelineDMScore) pipelineDMScore {
	s := newDMScore()
	chloride := input.RiskInput.ChlorideContent
	phLevel := input.RiskInput.PHLevel

	icSev := internalScore.severity
	gatePassed := icSev != "NOT" || chloride >= 3 || phLevel == "â‰¤4.5" || input.RiskInput.PrevLocIntCorrosion == "Finding"
	s.addGate("internal_corrosion_severity", icSev, "Internal Corrosion gate for localized corrosion", icSev != "NOT")
	s.addGate("chloride_content", fmt.Sprintf("%d", chloride), "Chloride gate for localized corrosion (>=3)", chloride >= 3)
	s.addGate("ph_level", phLevel, "pH gate for localized corrosion (acidic)", phLevel == "â‰¤4.5")

	if !gatePassed {
		s.severity = "NOT"
		s.formula = "Gate: No localized corrosion driver present"
		return s
	}

	s.score = internalScore.score
	s.formula = fmt.Sprintf("Base: inherit from Internal Corrosion (%.2f) + modifiers", internalScore.score)
	s.addTrigger("internal_corrosion_score", fmt.Sprintf("%.2f", internalScore.score), "Inherited from Internal Corrosion")

	if chloride >= 3 {
		modifier := pipelineChlorideSeverity[chloride] - 1.0 // currently 0.0
		s.addModifier("chloride_content", fmt.Sprintf("%d", chloride),
			"Chloride modifier TODO_ENGINEERING_CONFIRMATION", modifier)
	}
	if phLevel == "â‰¤4.5" || phLevel == "4.5-6.5" {
		modifier := pipelinePHSeverity[phLevel] - 1.0 // currently 0.0
		s.addModifier("ph_level", phLevel,
			"pH modifier for localized corrosion TODO_ENGINEERING_CONFIRMATION", modifier)
	}

	s.escalateByFinding(input.RiskInput.PrevLocIntCorrosion, input.RiskInput.ConfLocIntCorrosion, "localized_corrosion")

	s.severity = severityFromScore(s.score)
	return s
}

func scoreErosion(input PipelineOilInput) pipelineDMScore {
	s := newDMScore()
	velocity := input.RiskInput.FlowVelocityCondition
	solids := input.RiskInput.SolidContent
	corrosivity := input.RiskInput.FluidCorrosivityMPY

	gatePassed := (velocity != "" && velocity != "Low (<3 m/s)") || (solids != "" && solids != "None")
	s.addGate("flow_velocity_condition", velocity, "Velocity gate for erosion", velocity != "" && velocity != "Low (<3 m/s)")
	s.addGate("solid_content", solids, "Solid content gate for erosion", solids != "" && solids != "None")

	if !gatePassed {
		s.severity = "NOT"
		s.formula = "Gate: No erosion driver present (velocity low, no solids)"
		return s
	}

	// TODO_ENGINEERING_CONFIRMATION: base severity from velocity
	s.score = 1.0
	s.formula = "Base: velocity+solids screening TODO_ENGINEERING_CONFIRMATION"
	s.addTrigger("flow_velocity_condition", velocity, "Velocity condition for erosion TODO_ENGINEERING_CONFIRMATION")
	s.addTrigger("solid_content", solids, "Solid content for erosion TODO_ENGINEERING_CONFIRMATION")

	if solids != "" && solids != "None" {
		modifier := pipelineSolidContentModifiers[solids] - 1.0 // currently 0.0
		s.addModifier("solid_content", solids, "Solid content modifier TODO_ENGINEERING_CONFIRMATION", modifier)
	}
	if corrosivity != "" && corrosivity != "<2 mpy" {
		modifier := pipelineFluidCorrosivityMPYFactors[corrosivity] - 1.0 // currently 0.0
		s.addModifier("fluid_corrosivity_mpy", corrosivity, "Corrosivity modifier for erosion-corrosion TODO_ENGINEERING_CONFIRMATION", modifier)
	}

	s.severity = severityFromScore(s.score)
	return s
}

func scoreErosionCorrosion(input PipelineOilInput) pipelineDMScore {
	s := scoreErosion(input)
	if s.severity == "NOT" {
		return s
	}
	erosionSev := s.severity
	corrosivity := input.RiskInput.FluidCorrosivityMPY
	s.addModifier("fluid_corrosivity_mpy", corrosivity,
		"Corrosivity contribution for erosion-corrosion synergy TODO_ENGINEERING_CONFIRMATION",
		pipelineFluidCorrosivityMPYFactors[corrosivity]-1.0)
	s.formula = fmt.Sprintf("Erosion-Corrosion: inherit erosion (%s) + corrosivity synergy", erosionSev)
	s.severity = severityFromScore(s.score)
	return s
}

func scoreCracking(input PipelineOilInput) pipelineDMScore {
	s := newDMScore()
	pH2S := input.RiskInput.H2SPartialPressurePSIG
	h2sContent := input.RiskInput.H2SContent
	prevCracking := input.RiskInput.PrevIntCracking

	// Gate: pH2S >= 0.05 OR previous cracking finding OR H2S present
	gatePassed := pH2S >= 0.05 || prevCracking == "Finding" || h2sContent > 0
	s.addGate("h2s_partial_pressure_psig", fmt.Sprintf("%.4f", pH2S),
		"pH2S screening gate (NACE MR0175: 0.05 psig threshold)", pH2S >= 0.05)
	s.addGate("h2s_content", fmt.Sprintf("%.2f", h2sContent),
		"H2S presence gate for cracking", h2sContent > 0)
	s.addGate("prev_int_cracking", prevCracking,
		"Previous cracking finding gate", prevCracking == "Finding")

	if !gatePassed {
		s.severity = "NOT"
		s.formula = "Gate: No cracking driver present (pH2S<0.05, no H2S, no previous cracking)"
		return s
	}

	// Base severity from pH2S (SOURCED: NACE MR0175)
	baseSev := pH2SSeverity(pH2S)
	s.score = baseSeverityScore(baseSev)
	s.formula = fmt.Sprintf("Base: pH2S=%.4f psig â†’ %s (NACE MR0175)", pH2S, baseSev)
	s.addTrigger("h2s_partial_pressure_psig", fmt.Sprintf("%.4f", pH2S),
		fmt.Sprintf("pH2S %.4f psig â†’ %s per NACE MR0175", pH2S, baseSev))

	// Additive modifiers (all 0.0 placeholder)
	pwht := input.RiskInput.PWHTStatus
	if pwht != "" {
		modifier := pipelinePWHTModifiers[pwht] - 1.0 // currently 0.0
		s.addModifier("pwht_status", pwht, "PWHT modifier for cracking TODO_ENGINEERING_CONFIRMATION", modifier)
	}
	weldType := input.RiskInput.WeldJointType
	if weldType != "" {
		modifier := pipelineWeldCrackingModifiers[weldType] - 1.0 // currently 0.0
		s.addModifier("weld_joint_type", weldType, "Weld type modifier for cracking TODO_ENGINEERING_CONFIRMATION", modifier)
	}

	s.escalateByFinding(prevCracking, input.RiskInput.ConfIntCracking, "internal_cracking")

	s.severity = severityFromScore(s.score)
	s.triggerInputs = append(s.triggerInputs, PipelineTriggerInput{
		Field:  "cracking_result",
		Value:  s.severity,
		Reason: fmt.Sprintf("Score=%.2f â†’ %s (sour service cracking per NACE MR0175)", s.score, s.severity),
	})
	return s
}

func scoreSCC(input PipelineOilInput) pipelineDMScore {
	s := newDMScore()
	smysPct := input.RiskInput.SMYSUtilizationPct
	coatingCond := input.RiskInput.CoatingCondition
	cpStatus := input.RiskInput.CPStatus
	h2sContent := input.RiskInput.H2SContent

	// Gate: stress > 30% SMYS AND (coating concern OR CP concern OR H2S present)
	coatingConcern := coatingCond == "Damaged"
	cpConcern := cpStatus == "Failed" || cpStatus == "Borderline"
	h2sPresent := h2sContent > 0
	stressConcern := smysPct >= 30.0

	gatePassed := stressConcern && (coatingConcern || cpConcern || h2sPresent)
	s.addGate("smys_utilization_pct", fmt.Sprintf("%.1f", smysPct),
		fmt.Sprintf("SCC stress gate: %.1f%% SMYS >= 30%% threshold TODO_ENGINEERING_CONFIRMATION", smysPct), stressConcern)
	s.addGate("coating_condition", coatingCond, "Coating condition for SCC", coatingConcern)
	s.addGate("cp_status", cpStatus, "CP concern for SCC", cpConcern)
	s.addGate("h2s_content", fmt.Sprintf("%.2f", h2sContent), "H2S presence for SCC", h2sPresent)

	if !gatePassed {
		s.severity = "NOT"
		s.formula = "Gate: SCC conditions not met (stress<30% SMYS or no environmental driver)"
		return s
	}

	// TODO_ENGINEERING_CONFIRMATION: SCC base severity from stress thresholds
	// 30-50% SMYS = Low, 50-72% = Moderate, >72% = High
	if smysPct >= 50 && smysPct < 72 {
		s.score = 2.0
	} else if smysPct >= 72 {
		s.score = 3.0
	} else {
		s.score = 1.0
	}
	s.formula = fmt.Sprintf("Base: SMYS utilization=%.1f%% â†’ SCC severity TODO_ENGINEERING_CONFIRMATION", smysPct)
	s.addTrigger("smys_utilization_pct", fmt.Sprintf("%.1f%%", smysPct), "SCC stress-based severity TODO_ENGINEERING_CONFIRMATION")

	if coatingConcern {
		s.addModifier("coating_condition", coatingCond, "Coating disbondment for near-neutral pH SCC TODO_ENGINEERING_CONFIRMATION",
			pipelineCoatingConditionFactors[coatingCond]-1.0)
	}
	if h2sPresent {
		s.addModifier("h2s_content", fmt.Sprintf("%.2f", h2sContent), "H2S for sour SCC TODO_ENGINEERING_CONFIRMATION", 0.0)
	}

	s.severity = severityFromScore(s.score)
	return s
}

func scoreFatigue(input PipelineOilInput) pipelineDMScore {
	s := newDMScore()
	cycles := input.RiskInput.PressureCycleCount
	prevCracking := input.RiskInput.PrevIntCracking

	// Gate: pressure cycling exists OR previous fatigue/cracking finding
	gatePassed := cycles > 0 || prevCracking == "Finding"
	s.addGate("pressure_cycle_count", fmt.Sprintf("%.0f", cycles),
		"Fatigue cycling gate", cycles > 0)
	s.addGate("prev_int_cracking", prevCracking,
		"Previous cracking finding for fatigue", prevCracking == "Finding")

	if !gatePassed {
		s.severity = "NOT"
		s.formula = "Gate: No fatigue conditions (no cycling, no previous cracking)"
		return s
	}

	// TODO_ENGINEERING_CONFIRMATION: fatigue severity from cycle count and stress range
	s.score = 1.0
	s.formula = "Base: pressure cycling screening TODO_ENGINEERING_CONFIRMATION"
	s.addTrigger("pressure_cycle_count", fmt.Sprintf("%.0f", cycles),
		"Pressure cycle count for fatigue TODO_ENGINEERING_CONFIRMATION")

	if input.RiskInput.PressureRangePct > 0 {
		s.addModifier("pressure_range_pct", fmt.Sprintf("%.1f%%", input.RiskInput.PressureRangePct),
			"Stress range for fatigue TODO_ENGINEERING_CONFIRMATION", 0.0)
	}
	weldType := input.RiskInput.WeldJointType
	if weldType != "" {
		modifier := pipelineWeldCrackingModifiers[weldType] - 1.0
		s.addModifier("weld_joint_type", weldType, "Weld susceptibility for fatigue TODO_ENGINEERING_CONFIRMATION", modifier)
	}

	s.escalateByFinding(prevCracking, input.RiskInput.ConfIntCracking, "internal_cracking")

	s.severity = severityFromScore(s.score)
	return s
}

func scoreCoatingDegradation(input PipelineOilInput) pipelineDMScore {
	s := newDMScore()
	coatingCond := input.RiskInput.CoatingCondition
	cpStatus := input.RiskInput.CPStatus
	cpPotential := input.RiskInput.CPPotentialMV
	insulationCond := input.RiskInput.InsulationCondition

	gatePassed := coatingCond == "Damaged" || cpStatus != "normal" || insulationCond == "Damaged"
	s.addGate("coating_condition", coatingCond, "Coating condition gate", coatingCond == "Damaged")
	s.addGate("cp_status", cpStatus, "CP status gate", cpStatus != "normal")
	s.addGate("insulation_condition", insulationCond, "Insulation condition gate", insulationCond == "Damaged")

	if !gatePassed {
		s.severity = "NOT"
		s.formula = "Gate: No coating degradation driver present"
		return s
	}

	// TODO_ENGINEERING_CONFIRMATION: base severity from coating+CP combination
	s.score = 1.0
	s.formula = "Base: coating+CP screening TODO_ENGINEERING_CONFIRMATION"
	s.addTrigger("coating_condition", coatingCond, "Coating condition for degradation")

	if coatingCond == "Damaged" && input.RiskInput.CoatingDamageLevel != "" {
		modifier := pipelineCoatingDamageModifiers[input.RiskInput.CoatingDamageLevel] - 1.0
		s.addModifier("coating_damage_level", input.RiskInput.CoatingDamageLevel,
			"Coating damage level modifier TODO_ENGINEERING_CONFIRMATION", modifier)
	}
	if input.RiskInput.SoilResistivity == "<1000" {
		s.addModifier("soil_resistivity", input.RiskInput.SoilResistivity,
			"High corrosivity soil modifier TODO_ENGINEERING_CONFIRMATION", pipelineSoilFactors[input.RiskInput.SoilResistivity]-1.0)
	}
	// CP potential override sourced from NACE SP0169
	if cpPotential != 0 && cpPotential > -850 {
		s.addModifier("cp_potential_mv", fmt.Sprintf("%.0f", cpPotential),
			"CP potential >-850mV increases coating degradation risk (NACE SP0169)", 0.0) // TODO_ENGINEERING_CONFIRMATION
	}
	// CUI check: insulation damaged AND temperature in CUI range
	opTempC := (5.0 / 9.0) * (input.DesignTemperatureF - 32)
	if insulationCond == "Damaged" && opTempC >= 0 && opTempC <= 175 {
		s.addModifier("insulation_condition", insulationCond,
			fmt.Sprintf("CUI: insulation damaged + operating temp %.0fÂ°C in 0-175Â°C range", opTempC), 0.0) // TODO_ENGINEERING_CONFIRMATION
	}

	s.escalateByFinding(input.RiskInput.PrevExtCorrosion, input.RiskInput.ConfExtCorrosion, "external_corrosion")

	s.severity = severityFromScore(s.score)
	return s
}

func scoreThirdPartyDamage(input PipelineOilInput) pipelineDMScore {
	s := newDMScore()

	// TPD is always screened for buried pipelines
	s.addGate("pipeline_type", input.TypeOfInstallation, "TPD screening for buried/above-ground pipeline", true)

	baseRate := input.RiskInput.BaseTPDRate
	// TODO_ENGINEERING_CONFIRMATION: all factors are 1.0 placeholders
	depthFactor, ok := pipelineDepthFactors[input.RiskInput.DepthOfCover]
	if !ok {
		depthFactor = 1.0
	}
	patrolFactor, ok := pipelinePatrolFactors[input.RiskInput.PatrolFrequency]
	if !ok {
		patrolFactor = 1.0
	}
	rowFactor, ok := pipelineROWFactors[input.RiskInput.ROWCondition]
	if !ok {
		rowFactor = 1.0
	}
	oneCallFactor, ok := pipelineOneCallModifiers[input.RiskInput.OneCallSystem]
	if !ok {
		oneCallFactor = 1.0
	}

	// Framework: base rate with additive mitigation/escalation (not multiplicative/division)
	// Since all factors are 1.0, score = base rate for now
	s.score = baseRate
	s.formula = fmt.Sprintf("Base: TPD base rate=%.4f + depth(%.2f) + patrol(%.2f) + ROW(%.2f) + one-call(%.2f) TODO_ENGINEERING_CONFIRMATION",
		baseRate, depthFactor, patrolFactor, rowFactor, oneCallFactor)
	s.addTrigger("base_tpd_rate", fmt.Sprintf("%.4f", baseRate), "Base TPD rate")
	s.addTrigger("depth_of_cover", input.RiskInput.DepthOfCover, fmt.Sprintf("Depth factor=%.2f TODO_ENGINEERING_CONFIRMATION", depthFactor))
	s.addTrigger("patrol_frequency", input.RiskInput.PatrolFrequency, fmt.Sprintf("Patrol factor=%.2f TODO_ENGINEERING_CONFIRMATION", patrolFactor))
	s.addTrigger("row_condition", input.RiskInput.ROWCondition, fmt.Sprintf("ROW factor=%.2f TODO_ENGINEERING_CONFIRMATION", rowFactor))
	s.addTrigger("one_call_system", input.RiskInput.OneCallSystem, fmt.Sprintf("One-call factor=%.2f TODO_ENGINEERING_CONFIRMATION", oneCallFactor))

	if baseRate <= 0 {
		s.score = 1.0
	}
	s.severity = severityFromScore(s.score)
	return s
}

func scoreChemicalDamage(input PipelineOilInput) pipelineDMScore {
	s := newDMScore()
	s.severity = "NOT"
	s.score = 0
	s.formula = "Chemical damage requires engineering review; no screening formula implemented"
	s.addTrigger("chemical_damage", "N/A", "Score=0 placeholder; mechanism requires engineering confirmation")
	return s
}

func averagePositive(values ...float64) float64 {
	total, count := 0.0, 0.0
	for _, value := range values {
		if value > 0 {
			total += value
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / count
}

func severityInputScore(value, low, moderate, high float64) float64 {
	switch {
	case value >= high:
		return 3.5
	case value >= moderate:
		return 2.2
	case value >= low:
		return 1.2
	default:
		return 0
	}
}

func severityTextScore(value string, scores map[string]float64) float64 {
	if score, ok := scores[strings.ToLower(strings.TrimSpace(value))]; ok {
		return score
	}
	return 1
}

func remainingLifeSeverityScore(input PipelineOilInput) float64 {
	lowest := math.MaxFloat64
	for _, point := range input.InspectionPoints {
		cr := corrosionRateMMYear(point.NominalThicknessMM, point.ActualThicknessMM, input.YearUsed, point.MeasuredYear)
		required := point.RequiredThicknessMM
		if required <= 0 {
			required = requiredThicknessInForInput(input) * 25.4
		}
		rl := remainingLifeYears(point.ActualThicknessMM, required, cr)
		if rl > 0 && rl < lowest {
			lowest = rl
		}
	}
	if lowest == math.MaxFloat64 {
		return 1
	}
	switch {
	case lowest < 2:
		return 3.5
	case lowest < 5:
		return 2.2
	case lowest < 10:
		return 1.2
	default:
		return 0.8
	}
}

func calculatePipelinePoF(gff, managementSystemScore float64, dmResults []PipelineDamageMechanismResult) (float64, float64, float64, string) {
	fms := managementSystemFactor(managementSystemScore)
	var governingScore float64
	var governingDriver string
	for _, dm := range dmResults {
		if dm.Score > governingScore {
			governingScore = dm.Score
			governingDriver = dm.Label
		}
	}
	// PoF formula uses governing DM score

	pofValue := gff * governingScore * fms
	return pofValue, fms, governingScore, governingDriver
}

func calculateGasCoF(outsideDiameterIn, operatingPressurePsi float64, buildingCount int, classLocation string) (string, float64) {
	pir := 0.69 * outsideDiameterIn * math.Sqrt(operatingPressurePsi)
	category := "A"
	switch {
	case buildingCount > 20:
		category = "E"
	case buildingCount >= 6:
		category = "D"
	case buildingCount >= 1:
		category = "B"
	default:
		category = "A"
	}
	if strings.EqualFold(classLocation, "class_3") || strings.EqualFold(classLocation, "village") {
		if cofNumeric(category) < 3 {
			category = "C"
		}
	}
	if strings.EqualFold(classLocation, "class_4") || strings.EqualFold(classLocation, "urban_dense") {
		if cofNumeric(category) < 4 {
			category = "D"
		}
	}
	return category, pir
}

func calculateLiquidCoF(flowRate, detectionTimeHours, outsideDiameterIn, segmentLengthBetweenValvesM float64, environmentalSensitivity string, nearbySensitiveReceptor, isolationValveAvailable bool) (string, float64, float64) {
	diameterM := outsideDiameterIn * 0.0254
	pipelineVolumeM3 := math.Pi * math.Pow(diameterM/2, 2) * segmentLengthBetweenValvesM
	spillVolume := flowRate*detectionTimeHours + pipelineVolumeM3*cubicMetersToBarrels
	adjustedSpillVolume := spillVolume * lookupPipelineFactor(pipelineEnvironmentalMultipliers, environmentalSensitivity)
	if nearbySensitiveReceptor {
		adjustedSpillVolume *= 1.25
	}
	if !isolationValveAvailable {
		adjustedSpillVolume *= 1.25
	}
	category := "A"
	switch {
	case adjustedSpillVolume > 1000:
		category = "E"
	case adjustedSpillVolume > 300:
		category = "D"
	case adjustedSpillVolume > 100:
		category = "C"
	case adjustedSpillVolume > 25:
		category = "B"
	}
	return category, spillVolume, adjustedSpillVolume
}

func calculatePipelineRiskRanking(pofCategory, cofCategory string) (string, string) {
	score := pofNumeric(pofCategory) * cofNumeric(cofCategory)
	level := "Low Risk"
	switch {
	case score >= 20:
		level = "Critical Risk"
	case score >= 12:
		level = "High Risk"
	case score >= 6:
		level = "Medium Risk"
	}
	return pofCategory + cofCategory, level
}

type pipelineEngineeringAdvisory struct {
	Groups         PipelineAdvisoryGroups
	Recommendation string
	Source         string
	RuleName       string
}

func generatePipelineEngineeringAdvisory(result *PipelineOilResult, input PipelineOilInput) pipelineEngineeringAdvisory {
	groups := PipelineAdvisoryGroups{
		InspectionMonitor:  []string{"Keep the formula trace with the assessment record."},
		LongTermMitigation: []string{"Update the assessment after mitigation or inspection results are available."},
	}
	switch result.GoverningDamageMechanism {
	case "Third-Party / Mechanical Damage":
		groups.ImmediateActions = append(groups.ImmediateActions, "Improve route markers and warning signs.", "Strengthen excavation permit control.")
		groups.InspectionMonitor = append(groups.InspectionMonitor, "Increase ROW patrol frequency.")
	case "External Corrosion":
		groups.ImmediateActions = append(groups.ImmediateActions, "Verify cathodic protection performance.", "Prioritize coating defect checks.")
		groups.InspectionMonitor = append(groups.InspectionMonitor, "Plan CIPS/DCVG survey and soil monitoring.")
	case "Internal Corrosion":
		groups.ImmediateActions = append(groups.ImmediateActions, "Review inhibitor condition, fluid corrosivity, and water handling.")
		groups.InspectionMonitor = append(groups.InspectionMonitor, "Schedule pigging, fluid sampling, and wall thickness inspection.")
	}
	if result.FinalRiskLevel == "Critical Risk" {
		groups.ImmediateActions = append(groups.ImmediateActions, "Escalate to engineering review before continued operation.")
	}
	if result.FinalRiskLevel == "High Risk" {
		groups.ImmediateActions = append(groups.ImmediateActions, "Assign mitigation owner and target date.")
	}
	if isGasService(input.Service) {
		groups.LongTermMitigation = append([]string{"Review class location, public awareness, emergency response, and populated-area protection."}, groups.LongTermMitigation...)
	} else {
		groups.LongTermMitigation = append([]string{"Improve leak detection, isolation time, spill containment, and drainage/river protection."}, groups.LongTermMitigation...)
	}
	all := append([]string{}, groups.ImmediateActions...)
	all = append(all, groups.InspectionMonitor...)
	all = append(all, groups.LongTermMitigation...)

	recStr := strings.Join(all, " ")
	srcStr := "System advisory rule based on risk category, CoF factors, and governing pipeline DM score driver."

	if input.ManualRecommendation != "" {
		recStr = input.ManualRecommendation
		srcStr = "User overridden recommendation."
	}

	return pipelineEngineeringAdvisory{
		Groups:         groups,
		Recommendation: recStr,
		Source:         srcStr,
		RuleName:       "pipeline-system-advisory-v2",
	}
}

func aggregateInspectionResult(statuses ...string) string {
	for _, status := range statuses {
		if strings.EqualFold(status, "NOT ACCEPTABLE") {
			return "NOT ACCEPTABLE"
		}
	}
	for _, status := range statuses {
		if strings.TrimSpace(status) != "" && !strings.EqualFold(status, "ACCEPTABLE") {
			return "REVIEW REQUIRED"
		}
	}
	return "ACCEPTABLE"
}

func lookupPipelineFactor(factors map[string]float64, key string) float64 {
	if value, ok := factors[strings.ToLower(strings.TrimSpace(key))]; ok {
		return value
	}
	if value, ok := factors[strings.TrimSpace(key)]; ok {
		return value
	}
	return 1
}

func managementSystemFactor(score float64) float64 {
	pscore := (score / 1000) * 100
	return math.Pow(10, (-0.02*pscore)+1)
}

func pofCategory(pofValue float64) string {
	switch {
	case pofValue >= 0.01:
		return "5"
	case pofValue >= 0.001:
		return "4"
	case pofValue >= 0.0001:
		return "3"
	case pofValue >= 0.00001:
		return "2"
	default:
		return "1"
	}
}

func pofNumeric(category string) float64 {
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

func cofNumeric(category string) float64 {
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

func isGasService(service string) bool {
	return pipelineServiceFormulaFamily(service) == "gas"
}

func pipelineServiceFormulaFamily(service string) string {
	switch strings.ToLower(strings.TrimSpace(service)) {
	case "gas", "natural gas", "dwr gas", "wet gas":
		return "gas"
	case "liquid", "piping", "produce water", "produced water", "liquid hydrocarbon", "chemical":
		return "liquid"
	default:
		return "oil"
	}
}

func isValidPipelineService(service string) bool {
	switch strings.ToLower(strings.TrimSpace(service)) {
	case "gas", "natural gas", "dwr gas", "wet gas", "oil", "liquid", "piping", "produce water", "produced water", "liquid hydrocarbon", "chemical":
		return true
	default:
		return false
	}
}

func applyAPI581PublicMethodology(input PipelineOilInput, result *PipelineOilResult) {
	result.GenericFailureFrequency = input.RiskInput.GenericFailureFrequency
	result.ManagementSystemScore = input.RiskInput.ManagementSystemScore
	result.DamageFactor = input.RiskInput.DamageFactor

	if input.RiskInput.ManagementSystemScore > 0 {
		pscore := (input.RiskInput.ManagementSystemScore / 1000) * 100
		result.ManagementSystemFactor = math.Pow(10, (-0.02*pscore)+1)
		result.FormulaTrace = append(result.FormulaTrace,
			trace("api581_management_system_factor", "API 581 public methodology", "FMS = 10^((-0.02*(score/1000*100))+1)", map[string]interface{}{"management_system_score": input.RiskInput.ManagementSystemScore, "pscore": pscore}, result.ManagementSystemFactor, "Uses public API 581 methodology summary; scoring source remains engineer supplied."),
		)
	}

	if input.RiskInput.GenericFailureFrequency > 0 && result.ManagementSystemFactor > 0 && input.RiskInput.DamageFactor > 0 {
		result.PoFValue = input.RiskInput.GenericFailureFrequency * result.ManagementSystemFactor * input.RiskInput.DamageFactor
		result.PoF = formatEngineeringValue(result.PoFValue)
		result.FormulaTrace = append(result.FormulaTrace,
			trace("api581_probability_of_failure", "API 581 public methodology", "PoF(t) = GFF * FMS * Df(t)", map[string]interface{}{"gff": input.RiskInput.GenericFailureFrequency, "fms": result.ManagementSystemFactor, "damage_factor": input.RiskInput.DamageFactor}, result.PoFValue, "GFF and Df are engineer supplied because API 581 lookup tables are not included in the workbook."),
		)
	}

	switch {
	case input.RiskInput.ConsequenceFinancial > 0:
		result.CoFValue = input.RiskInput.ConsequenceFinancial
		result.CoF = formatEngineeringValue(result.CoFValue)
		result.FormulaTrace = append(result.FormulaTrace,
			trace("api581_consequence_financial", "API 581 public methodology", "CoF = financial consequence input", map[string]interface{}{"consequence_financial": input.RiskInput.ConsequenceFinancial}, result.CoFValue, "Financial consequence is engineer supplied; API 581 detailed Level 1/2 consequence model is not in workbook."),
		)
	case input.RiskInput.ConsequenceArea > 0:
		result.CoFValue = input.RiskInput.ConsequenceArea
		result.CoF = formatEngineeringValue(result.CoFValue)
		result.FormulaTrace = append(result.FormulaTrace,
			trace("api581_consequence_area", "API 581 public methodology", "CoF = affected area consequence input", map[string]interface{}{"consequence_area": input.RiskInput.ConsequenceArea}, result.CoFValue, "Affected area consequence is engineer supplied; API 581 detailed Level 1/2 consequence model is not in workbook."),
		)
	}

	if result.PoFValue > 0 && result.CoFValue > 0 {
		result.RiskValue = result.PoFValue * result.CoFValue
		result.FormulaTrace = append(result.FormulaTrace,
			trace("api581_risk", "API 581 public methodology", "Risk = PoF * CoF", map[string]interface{}{"pof": result.PoFValue, "cof": result.CoFValue}, result.RiskValue, ""),
		)
	}

	if input.RiskInput.PoFCategory != "" {
		result.PoF = input.RiskInput.PoFCategory
	}
	if input.RiskInput.CoFCategory != "" {
		result.CoF = input.RiskInput.CoFCategory
	}
	if input.RiskInput.RiskRanking != "" {
		result.RiskRanking = input.RiskInput.RiskRanking
	} else if result.RiskValue > 0 {
		result.RiskRanking = formatEngineeringValue(result.RiskValue)
	}
}

func ValidatePipelineOilDraft(input PipelineOilInput) []PipelineOilValidationError {
	var errs []PipelineOilValidationError
	required := []struct {
		field string
		value string
	}{
		{"report_no", input.ReportNo},
		{"line_identification", input.LineIdentification},
		{"service", input.Service},
		{"assessment_by", input.AssessmentBy},
	}
	for _, item := range required {
		if strings.TrimSpace(item.value) == "" {
			errs = append(errs, PipelineOilValidationError{Field: item.field, Message: "required"})
		}
	}
	if input.Service != "" && !isValidPipelineService(input.Service) {
		errs = append(errs, PipelineOilValidationError{Field: "service", Message: "must be one of Natural gas, Dwr gas, Wet gas, Oil, Produce water, Liquid hydrocarbon, or Chemical"})
	}
	if input.RiskInput.DamageMechanism != "" && !IsValidPipelineDamageMechanism(input.RiskInput.DamageMechanism) {
		errs = append(errs, PipelineOilValidationError{Field: "RiskInput.damage_mechanism", Message: "invalid pipeline damage mechanism"})
	}
	return errs
}

func ValidatePipelineOilCalculation(input PipelineOilInput) []PipelineOilValidationError {
	errs := ValidatePipelineOilDraft(input)
	checkPositive := []struct {
		field string
		value float64
	}{
		{"pipe_length_m", input.PipeLengthM},
		{"smys_psi", input.SMYSPsi},
		{"internal_design_pressure_psi", input.InternalDesignPressurePsi},
		{"outside_diameter_in", input.OutsideDiameterIn},
		{"nominal_wall_thickness_mm", input.NominalWallThicknessMM},
		{"actual_wall_thickness_mm", input.ActualWallThicknessMM},
		{"quality_factor", input.QualityFactor},
		{"weld_joint_strength_factor", input.WeldJointStrengthFactor},
		{"design_factor", input.DesignFactor},
	}
	for _, item := range checkPositive {
		if item.value <= 0 {
			errs = append(errs, PipelineOilValidationError{Field: item.field, Message: "must be greater than zero"})
		}
	}
	if input.YearBuilt <= 0 || input.YearUsed.Float() <= 0 {
		errs = append(errs, PipelineOilValidationError{Field: "year_built/year_used", Message: "valid year built and used are required"})
	}
	if input.YearUsed.Float() < float64(input.YearBuilt) {
		errs = append(errs, PipelineOilValidationError{Field: "year_used", Message: "cannot be before year built"})
	}
	if input.OperatingPressurePsi < 0 {
		errs = append(errs, PipelineOilValidationError{Field: "operating_pressure_psi", Message: "cannot be negative"})
	}
	if input.InternalDesignPressurePsi > 20000 {
		errs = append(errs, PipelineOilValidationError{Field: "internal_design_pressure_psi", Message: "extreme pressure requires engineering confirmation"})
	}
	if input.RiskInput.ManagementSystemScore < 0 || input.RiskInput.ManagementSystemScore > 1000 {
		errs = append(errs, PipelineOilValidationError{Field: "RiskInput.management_system_score", Message: "must be between 0 and 1000"})
	}
	if input.RiskInput.GenericFailureFrequency <= 0 {
		errs = append(errs, PipelineOilValidationError{Field: "RiskInput.generic_failure_frequency", Message: "must be greater than zero"})
	}
	if input.RiskInput.BaseTPDRate <= 0 || input.RiskInput.BaseExternalCorrRate <= 0 || input.RiskInput.BaseInternalCorrRate <= 0 {
		errs = append(errs, PipelineOilValidationError{Field: "RiskInput.base_rates", Message: "must be greater than zero"})
	}
	validateOption := func(field, value string, allowed map[string]float64) {
		if _, ok := allowed[strings.TrimSpace(value)]; !ok {
			if _, ok := allowed[strings.ToLower(strings.TrimSpace(value))]; !ok {
				errs = append(errs, PipelineOilValidationError{Field: field, Message: "invalid option"})
			}
		}
	}
	validateOption("RiskInput.depth_of_cover", input.RiskInput.DepthOfCover, pipelineDepthFactors)
	validateOption("RiskInput.patrol_frequency", input.RiskInput.PatrolFrequency, pipelinePatrolFactors)
	validateOption("RiskInput.row_condition", input.RiskInput.ROWCondition, pipelineROWFactors)
	validateOption("RiskInput.soil_resistivity", input.RiskInput.SoilResistivity, pipelineSoilFactors)
	validateOption("RiskInput.coating_condition", input.RiskInput.CoatingCondition, pipelineCoatingConditionFactors)
	validateOption("RiskInput.cp_status", input.RiskInput.CPStatus, pipelineCPFactors)
	// Removed old validations: fluid_corrosivity, water_content, co2_h2s_presence, mic_risk, wall_thickness_condition, emergency_response
	// New validations for v2 fields
	validateOption("RiskInput.ph_level", input.RiskInput.PHLevel, pipelinePHSeverity)
	validateOption("RiskInput.fluid_corrosivity_mpy", input.RiskInput.FluidCorrosivityMPY, pipelineFluidCorrosivityMPYFactors)
	validateOption("RiskInput.inhibitor_effectiveness", input.RiskInput.InhibitorEffectiveness, pipelineInhibitorModifiers)
	validateOption("RiskInput.h2s_ppm", input.RiskInput.H2SPpm, pipelineH2SPpmSeverity)
	validateOption("RiskInput.pwht_status", input.RiskInput.PWHTStatus, pipelinePWHTModifiers)
	validateOption("RiskInput.weld_joint_type", input.RiskInput.WeldJointType, pipelineWeldCrackingModifiers)
	validateOption("RiskInput.flow_velocity_condition", input.RiskInput.FlowVelocityCondition, pipelineFlowVelocityModifiers)
	validateOption("RiskInput.solid_content", input.RiskInput.SolidContent, pipelineSolidContentModifiers)
	validateOption("RiskInput.one_call_system", input.RiskInput.OneCallSystem, pipelineOneCallModifiers)
	validateOption("RiskInput.ext_coating_condition", input.RiskInput.ExtCoatingCondition, pipelineCoatingConditionFactors)
	validateOption("RiskInput.insulation_condition", input.RiskInput.InsulationCondition, map[string]float64{"Not Applicable": 1, "Good": 1, "Damaged": 1, "Not Inspectable": 1})
	if input.RiskInput.ChlorideContent < 0 || input.RiskInput.ChlorideContent > 5 {
		errs = append(errs, PipelineOilValidationError{Field: "RiskInput.chloride_content", Message: "must be between 0 and 5"})
	}
	if input.RiskInput.PressureCycleCount < 0 {
		errs = append(errs, PipelineOilValidationError{Field: "RiskInput.pressure_cycle_count", Message: "cannot be negative"})
	}
	if input.RiskInput.PressureRangePct < 0 {
		errs = append(errs, PipelineOilValidationError{Field: "RiskInput.pressure_range_pct", Message: "cannot be negative"})
	}
	validatePrevFinding := func(field, value string) {
		if value != "" && value != "No Finding" && value != "Finding" && value != "Not Inspectable" {
			errs = append(errs, PipelineOilValidationError{Field: field, Message: "must be No Finding, Finding, or Not Inspectable"})
		}
	}
	validatePrevFinding("RiskInput.prev_ext_corrosion", input.RiskInput.PrevExtCorrosion)
	validatePrevFinding("RiskInput.prev_int_thinning", input.RiskInput.PrevIntThinning)
	validatePrevFinding("RiskInput.prev_int_cracking", input.RiskInput.PrevIntCracking)
	validatePrevFinding("RiskInput.prev_loc_int_corrosion", input.RiskInput.PrevLocIntCorrosion)
	validateConfidence := func(field, value string) {
		if value != "" && value != "High" && value != "Average" && value != "Low" {
			errs = append(errs, PipelineOilValidationError{Field: field, Message: "must be High, Average, or Low"})
		}
	}
	validateConfidence("RiskInput.conf_ext_corrosion", input.RiskInput.ConfExtCorrosion)
	validateConfidence("RiskInput.conf_int_thinning", input.RiskInput.ConfIntThinning)
	validateConfidence("RiskInput.conf_int_cracking", input.RiskInput.ConfIntCracking)
	validateConfidence("RiskInput.conf_loc_int_corrosion", input.RiskInput.ConfLocIntCorrosion)
	validFluida := map[string]float64{"": 1, "Oil": 1, "Gas": 1, "Water": 1, "Multi-phase": 1}
	validPhase := map[string]float64{"": 1, "Single-phase": 1, "Two-phase": 1, "Multi-phase": 1}
	validateOption("RiskInput.fluida", input.RiskInput.Fluida, validFluida)
	validateOption("RiskInput.phase", input.RiskInput.Phase, validPhase)
	validateOption("RiskInput.insulation_damage_level", input.RiskInput.InsulationDamageLevel, map[string]float64{"Small": 1, "Medium": 1, "Large": 1, "Severe": 1})
	validateOption("RiskInput.ext_coating_damage_level", input.RiskInput.ExtCoatingDamageLevel, map[string]float64{"Small": 1, "Medium": 1, "Large": 1, "Severe": 1})
	validateOption("RiskInput.env_ext_cracking", input.RiskInput.EnvExtCracking, map[string]float64{"None": 1, "H2S": 1, "Chloride": 1, "Hydrogen": 1, "Marine": 1})
	if isGasService(input.Service) {
		if input.RiskInput.BuildingCountInsidePIR < 0 {
			errs = append(errs, PipelineOilValidationError{Field: "RiskInput.building_count_inside_pir", Message: "cannot be negative"})
		}
	} else {
		if input.RiskInput.FlowRate <= 0 {
			errs = append(errs, PipelineOilValidationError{Field: "RiskInput.flow_rate", Message: "must be greater than zero"})
		}
		if input.RiskInput.DetectionTimeHours <= 0 {
			errs = append(errs, PipelineOilValidationError{Field: "RiskInput.detection_time_hours", Message: "must be greater than zero"})
		}
		if input.RiskInput.SegmentLengthBetweenValvesM <= 0 {
			errs = append(errs, PipelineOilValidationError{Field: "RiskInput.segment_length_between_valves_m", Message: "must be greater than zero"})
		}
		validateOption("RiskInput.environmental_sensitivity", input.RiskInput.EnvironmentalSensitivity, pipelineEnvironmentalMultipliers)
	}
	if len(input.InspectionPoints) == 0 {
		errs = append(errs, PipelineOilValidationError{Field: "inspection_points", Message: "at least one inspection point is required"})
	}
	for i, point := range input.InspectionPoints {
		prefix := fmt.Sprintf("inspection_points[%d]", i)
		if strings.TrimSpace(point.InspectionPoint) == "" {
			errs = append(errs, PipelineOilValidationError{Field: prefix + ".inspection_point", Message: "required"})
		}
		if point.NominalThicknessMM <= 0 {
			errs = append(errs, PipelineOilValidationError{Field: prefix + ".nominal_thickness_mm", Message: "must be greater than zero"})
		}
		if point.ActualThicknessMM <= 0 {
			errs = append(errs, PipelineOilValidationError{Field: prefix + ".actual_thickness_mm", Message: "must be greater than zero"})
		}
		if point.MeasuredYear.Float() <= input.YearUsed.Float() {
			errs = append(errs, PipelineOilValidationError{Field: prefix + ".measured_year", Message: "must be after year used to avoid divide-by-zero"})
		}
		if point.ActualThicknessMM > point.NominalThicknessMM {
			errs = append(errs, PipelineOilValidationError{Field: prefix + ".actual_thickness_mm", Message: "cannot exceed nominal thickness for workbook corrosion-rate formula"})
		}
	}
	return errs
}

func applyPipelineOilDefaults(input *PipelineOilInput) {
	if input.Service == "" {
		input.Service = "Oil"
	}
	input.ApplicableCode = normalizePipelineApplicableCode(input.ApplicableCode)
	if input.ApplicableCode == "" {
		input.ApplicableCode = pipelineApplicableCodeForService(input.Service)
	}
	if input.JointEfficiency == 0 {
		input.JointEfficiency = 1
	}
	if input.QualityFactor == 0 {
		input.QualityFactor = input.JointEfficiency
	}
	if input.WeldJointStrengthFactor == 0 {
		input.WeldJointStrengthFactor = 1
	}
	if input.DesignFactor == 0 {
		input.DesignFactor = 0.72
	}
	if input.TemperatureDeratingFactor == 0 {
		input.TemperatureDeratingFactor = 1
	}
	if input.RiskInput.GenericFailureFrequency == 0 {
		input.RiskInput.GenericFailureFrequency = defaultPipelineGFF
	}
	if input.RiskInput.ManagementSystemScore == 0 {
		input.RiskInput.ManagementSystemScore = defaultPipelineManagementScore
	}
	if input.RiskInput.BaseTPDRate == 0 {
		input.RiskInput.BaseTPDRate = defaultPipelineBaseTPDRate
	}
	if input.RiskInput.BaseExternalCorrRate == 0 {
		input.RiskInput.BaseExternalCorrRate = defaultPipelineBaseCorrRate
	}
	if input.RiskInput.BaseInternalCorrRate == 0 {
		input.RiskInput.BaseInternalCorrRate = defaultPipelineBaseCorrRate
	}
	if input.RiskInput.DepthOfCover == "" {
		input.RiskInput.DepthOfCover = "1-2m"
	}
	if input.RiskInput.PatrolFrequency == "" {
		input.RiskInput.PatrolFrequency = "monthly"
	}
	if input.RiskInput.ROWCondition == "" {
		input.RiskInput.ROWCondition = "fair"
	}
	if input.RiskInput.SoilResistivity == "" {
		input.RiskInput.SoilResistivity = "1000-5000"
	}
	if input.RiskInput.CoatingCondition == "" {
		input.RiskInput.CoatingCondition = "Good"
	}
	if input.RiskInput.CPStatus == "" {
		input.RiskInput.CPStatus = "normal"
	}
	if input.RiskInput.ClassLocation == "" {
		input.RiskInput.ClassLocation = "class_2"
	}
	if input.RiskInput.PHLevel == "" {
		input.RiskInput.PHLevel = "6.5-8.5"
	}
	if input.RiskInput.FluidCorrosivityMPY == "" {
		input.RiskInput.FluidCorrosivityMPY = "2-5 mpy"
	}
	if input.RiskInput.H2SPpm == "" {
		input.RiskInput.H2SPpm = "<50 ppm"
	}
	if input.RiskInput.BiocideTreatment == "" {
		input.RiskInput.BiocideTreatment = "Not Required"
	}
	if input.RiskInput.InhibitorEffectiveness == "" {
		input.RiskInput.InhibitorEffectiveness = "None"
	}
	if input.RiskInput.CorrosionMonitoringResult == "" {
		input.RiskInput.CorrosionMonitoringResult = "Not Applicable"
	}
	if input.RiskInput.PWHTStatus == "" {
		input.RiskInput.PWHTStatus = "Unknown"
	}
	if input.RiskInput.WeldJointType == "" {
		input.RiskInput.WeldJointType = "Seamless"
	}
	if input.RiskInput.OneCallSystem == "" {
		input.RiskInput.OneCallSystem = "None"
	}
	if input.RiskInput.FlowVelocityCondition == "" {
		input.RiskInput.FlowVelocityCondition = "Moderate (3-10 m/s)"
	}
	if input.RiskInput.SolidContent == "" {
		input.RiskInput.SolidContent = "None"
	}
	if input.RiskInput.EnvExtCracking == "" {
		input.RiskInput.EnvExtCracking = "None"
	}
	if input.RiskInput.PrevExtCorrosion == "" {
		input.RiskInput.PrevExtCorrosion = "No Finding"
	}
	if input.RiskInput.PrevIntThinning == "" {
		input.RiskInput.PrevIntThinning = "No Finding"
	}
	if input.RiskInput.PrevIntCracking == "" {
		input.RiskInput.PrevIntCracking = "No Finding"
	}
	if input.RiskInput.PrevLocIntCorrosion == "" {
		input.RiskInput.PrevLocIntCorrosion = "No Finding"
	}
	if input.RiskInput.InsulationCondition == "" {
		input.RiskInput.InsulationCondition = "Not Applicable"
	}
	if input.RiskInput.ExtCoatingCondition == "" {
		input.RiskInput.ExtCoatingCondition = "Good"
	}
	if input.RiskInput.ExtCoatingDamageLevel == "" {
		input.RiskInput.ExtCoatingDamageLevel = "Small"
	}
	if input.RiskInput.InsulationDamageLevel == "" {
		input.RiskInput.InsulationDamageLevel = "Small"
	}
	if input.RiskInput.PrevIntCracking == "" {
		input.RiskInput.PrevIntCracking = "No Finding"
	}
	if input.RiskInput.ConfExtCorrosion == "" {
		input.RiskInput.ConfExtCorrosion = ""
	}
	if input.RiskInput.ConfIntThinning == "" {
		input.RiskInput.ConfIntThinning = ""
	}
	if input.RiskInput.ConfIntCracking == "" {
		input.RiskInput.ConfIntCracking = ""
	}
	if input.RiskInput.ConfLocIntCorrosion == "" {
		input.RiskInput.ConfLocIntCorrosion = ""
	}
	// Auto-calculated fields
	if input.OperatingPressurePsi > 0 {
		input.RiskInput.CO2PartialPressurePSIG = calculateCO2PartialPressure(input.RiskInput.CO2Content, input.OperatingPressurePsi)
		input.RiskInput.H2SPartialPressurePSIG = calculateH2SPartialPressure(input.RiskInput.H2SContent, input.OperatingPressurePsi)
	}
	if len(input.InspectionPoints) > 0 {
		input.RiskInput.WallThicknessRatio = calculateWallThicknessRatio(*input)
	}
	if input.SMYSPsi > 0 && input.OperatingPressurePsi > 0 {
		input.RiskInput.SMYSUtilizationPct = calculateSMYSUtilizationPct(*input)
	}
	if input.RiskInput.FlowRate == 0 {
		input.RiskInput.FlowRate = 100
	}
	if input.RiskInput.DetectionTimeHours == 0 {
		input.RiskInput.DetectionTimeHours = 1
	}
	if input.RiskInput.SegmentLengthBetweenValvesM == 0 {
		input.RiskInput.SegmentLengthBetweenValvesM = input.PipeLengthM
	}
	if input.RiskInput.EnvironmentalSensitivity == "" {
		input.RiskInput.EnvironmentalSensitivity = "medium"
	}
	if input.RiskInput.DamageMechanism == "" {
		input.RiskInput.DamageMechanism = "internal_corrosion"
	} else {
		input.RiskInput.DamageMechanism = NormalizePipelineDamageMechanism(input.RiskInput.DamageMechanism)
	}
	if input.RiskInput.InspectionEffectivity == "" {
		input.RiskInput.InspectionEffectivity = "Representative"
	}
	if input.YearUsed.Float() == 0 {
		input.YearUsed = FlexibleYear(strconv.Itoa(input.YearBuilt))
	}
	input.MaterialStressPsi = derivePipelineMaterialStressPsi(*input)
	for i := range input.InspectionPoints {
		if input.InspectionPoints[i].NominalThicknessMM == 0 {
			input.InspectionPoints[i].NominalThicknessMM = input.NominalWallThicknessMM
		}
		if input.InspectionPoints[i].MeasuredYear.Float() == 0 {
			input.InspectionPoints[i].MeasuredYear = FlexibleYear(strconv.Itoa(time.Now().Year()))
		}
	}
}

func pipelineApplicableCodeForService(service string) string {
	switch pipelineServiceFormulaFamily(service) {
	case "gas":
		return "ASME B31.8"
	case "liquid":
		return "ASME B31.3"
	default:
		return "ASME B31.4"
	}
}

func normalizePipelineApplicableCode(code string) string {
	upper := strings.ToUpper(strings.TrimSpace(code))
	switch {
	case strings.Contains(upper, "B31.3"):
		return "ASME B31.3"
	case strings.Contains(upper, "B31.4"):
		return "ASME B31.4"
	case strings.Contains(upper, "B31.8"):
		return "ASME B31.8"
	default:
		return ""
	}
}

func derivePipelineMaterialStressPsi(input PipelineOilInput) float64 {
	if input.SMYSPsi <= 0 {
		return 0
	}
	if isASMECode(input.ApplicableCode, "B31.3") {
		return math.Min((input.SMYSPsi*2)/3, 20000)
	}
	return input.SMYSPsi
}

func requiredThicknessIn(p, d, f, e, s, c float64) float64 {
	return ((p * d) / (2 * f * e * s)) + c
}

func requiredThicknessInForInput(input PipelineOilInput) float64 {
	if isASMECode(input.ApplicableCode, "B31.3") {
		denominator := 2 * ((input.MaterialStressPsi * input.QualityFactor * input.WeldJointStrengthFactor) + (input.InternalDesignPressurePsi * input.DesignFactor))
		if denominator == 0 {
			return 0
		}
		return ((input.InternalDesignPressurePsi * input.OutsideDiameterIn) / denominator) + input.AllowanceIn
	}
	if isASMECode(input.ApplicableCode, "B31.8") {
		denominator := 2 * input.DesignFactor * input.QualityFactor * input.TemperatureDeratingFactor * input.SMYSPsi
		if denominator == 0 {
			return 0
		}
		return ((input.InternalDesignPressurePsi * input.OutsideDiameterIn) / denominator) + input.AllowanceIn
	}
	return requiredThicknessIn(input.InternalDesignPressurePsi, input.OutsideDiameterIn, input.DesignFactor, input.QualityFactor, input.SMYSPsi, input.AllowanceIn)
}

func hoopStressPsi(p, d, actualThicknessIn float64) float64 {
	return (p * d) / (2 * actualThicknessIn)
}

func maopPsi(actualThicknessIn, s, f, e, d float64) float64 {
	return (2 * actualThicknessIn * s * f * e) / d
}

func maopPsiForInput(input PipelineOilInput, actualThicknessIn float64) float64 {
	if isASMECode(input.ApplicableCode, "B31.3") {
		denominator := input.OutsideDiameterIn - (2 * input.DesignFactor * actualThicknessIn)
		if denominator == 0 {
			return 0
		}
		return (2 * input.MaterialStressPsi * input.QualityFactor * input.WeldJointStrengthFactor * actualThicknessIn) / denominator
	}
	if isASMECode(input.ApplicableCode, "B31.8") {
		return maopPsi(actualThicknessIn, input.SMYSPsi, input.DesignFactor, input.QualityFactor, input.OutsideDiameterIn) * input.TemperatureDeratingFactor
	}
	return maopPsi(actualThicknessIn, input.SMYSPsi, input.DesignFactor, input.QualityFactor, input.OutsideDiameterIn)
}

func maopExpression(input PipelineOilInput) string {
	if isASMECode(input.ApplicableCode, "B31.3") {
		return "(2*S*E*W*actual_thickness_in)/(D-2*Y*actual_thickness_in)"
	}
	if isASMECode(input.ApplicableCode, "B31.8") {
		return "(2*actual_thickness_in*SMYS*F*E*T)/D"
	}
	return "(2*actual_thickness_in*SMYS*F*E)/D"
}

func maopInputs(input PipelineOilInput, point string, actualThicknessIn float64) map[string]interface{} {
	if isASMECode(input.ApplicableCode, "B31.3") {
		return map[string]interface{}{"point": point, "actual_thickness_in": actualThicknessIn, "S": input.MaterialStressPsi, "E": input.QualityFactor, "W": input.WeldJointStrengthFactor, "D": input.OutsideDiameterIn, "Y": input.DesignFactor}
	}
	if isASMECode(input.ApplicableCode, "B31.8") {
		return map[string]interface{}{"point": point, "actual_thickness_in": actualThicknessIn, "SMYS": input.SMYSPsi, "F": input.DesignFactor, "E": input.QualityFactor, "T": input.TemperatureDeratingFactor, "D": input.OutsideDiameterIn}
	}
	return map[string]interface{}{"point": point, "actual_thickness_in": actualThicknessIn, "SMYS": input.SMYSPsi, "F": input.DesignFactor, "E": input.QualityFactor, "D": input.OutsideDiameterIn}
}

func psiToKgCM2(psi float64) float64 {
	return psi / 14.223
}

func roundToPlaces(value float64, places int) float64 {
	factor := math.Pow10(places)
	return math.Round(value*factor) / factor
}

func parseMonthYearToFloat(val string) float64 {
	raw := strings.TrimSpace(val)
	if raw == "" {
		return 0
	}
	if strings.Contains(raw, "/") {
		parts := strings.Split(raw, "/")
		if len(parts) == 2 {
			month, _ := strconv.ParseFloat(parts[0], 64)
			year, _ := strconv.ParseFloat(parts[1], 64)
			if year <= 0 {
				return 0
			}
			if month <= 0 {
				return year
			}
			return year + (month-1)/12.0
		}
	}
	if strings.Contains(raw, "-") {
		parts := strings.Split(raw, "-")
		if len(parts) == 2 {
			first, _ := strconv.ParseFloat(parts[0], 64)
			second, _ := strconv.ParseFloat(parts[1], 64)
			if first <= 0 {
				return 0
			}
			if second > 100 {
				return second + (first-1)/12.0
			}
			return first + (second-1)/12.0
		}
	}
	year, _ := strconv.ParseFloat(raw, 64)
	return year
}

func corrosionRateMMYear(nominal, actual float64, yearUsed, measuredYear FlexibleYear) float64 {
	yUsed := parseMonthYearToFloat(string(yearUsed))
	yMeasured := parseMonthYearToFloat(string(measuredYear))
	diff := yMeasured - yUsed
	if diff <= 0 {
		return 0
	}
	return (nominal - actual) / diff
}

func remainingLifeYears(actual, required, corrosionRate float64) float64 {
	if corrosionRate <= 0 {
		return 0
	}
	return math.Min((actual-required)/corrosionRate, maxPipelineRemainingLifeYears)
}

func acceptable(ok bool) string {
	if ok {
		return "ACCEPTABLE"
	}
	return "UNACCEPTABLE"
}

func aggregateStatus(points []PipelineOilPointResult, pick func(PipelineOilPointResult) string) string {
	if len(points) == 0 {
		return "REVIEW REQUIRED"
	}
	for _, point := range points {
		if pick(point) != "ACCEPTABLE" {
			return "UNACCEPTABLE"
		}
	}
	return "ACCEPTABLE"
}

func minPositive(current, candidate float64) float64 {
	if candidate <= 0 {
		return current
	}
	if current == 0 || candidate < current {
		return candidate
	}
	return current
}

func trace(name, excelRef, expression string, inputs map[string]interface{}, output interface{}, note string) PipelineOilFormulaTrace {
	return PipelineOilFormulaTrace{FormulaName: name, ExcelRef: excelRef, Expression: expression, Inputs: inputs, Output: output, Note: note}
}

func formatEngineeringValue(value float64) string {
	return fmt.Sprintf("%.6g", value)
}

func isASMECode(applicableCode, code string) bool {
	normalizedCode := strings.ReplaceAll(strings.ToUpper(applicableCode), " ", "")
	normalizedNeedle := strings.ReplaceAll(strings.ToUpper(code), " ", "")
	return strings.Contains(normalizedCode, normalizedNeedle)
}

// --- Partial Pressure Calculation Helpers (Sourced) ---

// Sourced: API 581 Section 6 / PV CO2 corrosion logic
// pCO2 = mole% CO2 ├ù operating pressure (psig) / 100
func calculateCO2PartialPressure(co2ContentMolePct, operatingPressurePSIG float64) float64 {
	return (co2ContentMolePct / 100.0) * operatingPressurePSIG
}

// Sourced: NACE MR0175 / PV SSC logic
// pH2S = mole% H2S ├ù operating pressure (psig) / 100
func calculateH2SPartialPressure(h2sContentMolePct, operatingPressurePSIG float64) float64 {
	return (h2sContentMolePct / 100.0) * operatingPressurePSIG
}

// Auto-calculated from inspection points: min(actual) / min(required)
func calculateWallThicknessRatio(input PipelineOilInput) float64 {
	if len(input.InspectionPoints) == 0 {
		return 1.0
	}
	minActual := input.InspectionPoints[0].ActualThicknessMM
	minRequired := input.InspectionPoints[0].RequiredThicknessMM
	if minRequired <= 0 {
		minRequired = requiredThicknessInForInput(input) * 25.4
	}
	for _, pt := range input.InspectionPoints {
		if pt.ActualThicknessMM < minActual {
			minActual = pt.ActualThicknessMM
		}
		reqMM := pt.RequiredThicknessMM
		if reqMM <= 0 {
			reqMM = requiredThicknessInForInput(input) * 25.4
		}
		if reqMM < minRequired {
			minRequired = reqMM
		}
	}
	if minRequired <= 0 {
		return 1.0
	}
	return minActual / minRequired
}

// Auto-calculated: (operating_pressure ├ù OD) / (2 ├ù min_actual_thickness ├ù SMYS) ├ù 100
func calculateSMYSUtilizationPct(input PipelineOilInput) float64 {
	if input.SMYSPsi <= 0 || input.OutsideDiameterIn <= 0 || len(input.InspectionPoints) == 0 {
		return 0
	}
	minActualIn := input.InspectionPoints[0].ActualThicknessMM / 25.4
	for _, pt := range input.InspectionPoints {
		actualIn := pt.ActualThicknessMM / 25.4
		if actualIn < minActualIn {
			minActualIn = actualIn
		}
	}
	if minActualIn <= 0 {
		return 0
	}
	return (input.OperatingPressurePsi * input.OutsideDiameterIn) / (2 * minActualIn * input.SMYSPsi) * 100
}

// Sourced: API 581 Section 6 ΓÇö pCO2 partial pressure thresholds
func pCO2Severity(pCO2 float64) string {
	if pCO2 <= 0 {
		return "NOT"
	}
	if pCO2 <= pipelineCO2PartialPressureSeverity["Low"] {
		return "Low"
	}
	if pCO2 <= pipelineCO2PartialPressureSeverity["Moderate"] {
		return "Moderate"
	}
	return "High"
}

// Sourced: NACE MR0175 — pH2S partial pressure thresholds
func pH2SSeverity(pH2S float64) string {
	if pH2S < pipelineH2SPartialPressureSeverity["Not"] {
		return "NOT"
	}
	if pH2S < pipelineH2SPartialPressureSeverity["Low"] {
		return "Low"
	}
	if pH2S <= pipelineH2SPartialPressureSeverity["Moderate"] {
		return "Moderate"
	}
	return "High"
}

func pipelineOilJSONPayloads(input PipelineOilInput, result *PipelineOilResult) (string, string, string, string, error) {
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return "", "", "", "", err
	}
	resultBytes := []byte{}
	traceBytes := []byte("[]")
	if result != nil {
		resultBytes, err = json.Marshal(result)
		if err != nil {
			return "", "", "", "", err
		}
		traceBytes, err = json.Marshal(result.FormulaTrace)
		if err != nil {
			return "", "", "", "", err
		}
	}
	snapshot := map[string]interface{}{
		"formula_version": PipelineOilFormulaVersion,
		"input":           input,
		"result":          result,
		"snapshot_at":     time.Now(),
	}
	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		return "", "", "", "", err
	}
	return string(inputBytes), string(resultBytes), string(traceBytes), string(snapshotBytes), nil
}

func formatPipelineValidationErrors(errs []PipelineOilValidationError) string {
	parts := make([]string, 0, len(errs))
	for _, err := range errs {
		parts = append(parts, err.Field+" "+err.Message)
	}
	return strings.Join(parts, "; ")
}

type PipelineMaterial struct {
	ID   int
	Name string
	SMYS int
}

func GetMaterial(db *sql.DB) ([]PipelineMaterial, error) {
	rows, err := db.Query("SELECT id, name, smys FROM pipeline_materials")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []PipelineMaterial

	for rows.Next() {
		var eq PipelineMaterial
		err := rows.Scan(&eq.ID, &eq.Name, &eq.SMYS)
		if err != nil {
			return nil, err
		}
		materials = append(materials, eq)
	}

	return materials, nil
}
