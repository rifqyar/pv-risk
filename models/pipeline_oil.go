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

const PipelineOilFormulaVersion = "pipeline-oil-risk-v1"

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
	pipelineDepthFactors = map[string]float64{
		"<1m":  0.8,
		"1-2m": 1.2,
		">2m":  1.8,
	}
	pipelinePatrolFactors = map[string]float64{
		"rare":         0.8,
		"monthly":      1.2,
		"weekly_daily": 1.8,
	}
	pipelineROWFactors = map[string]float64{
		"poor": 0.8,
		"fair": 1.2,
		"good": 1.8,
	}
	pipelineSoilFactors = map[string]float64{
		"<1000":     3.0,
		"1000-5000": 1.8,
		">5000":     1.0,
	}
	pipelineConditionFactors = map[string]float64{
		"poor": 3.0,
		"fair": 1.8,
		"good": 1.0,
	}
	pipelineCPFactors = map[string]float64{
		"failed":     3.0,
		"borderline": 1.8,
		"normal":     1.0,
	}
	pipelineInternalFactors = map[string]float64{
		"low":      1.0,
		"none":     1.0,
		"healthy":  1.0,
		"medium":   1.9,
		"present":  1.9,
		"warning":  1.9,
		"high":     3.5,
		"critical": 3.5,
	}
	pipelineEnvironmentalMultipliers = map[string]float64{
		"low":    1.0,
		"medium": 1.5,
		"high":   2.5,
	}
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
	DamageMechanism             string                                 `json:"damage_mechanism" form:"damage_mechanism"`
	InspectionEffectivity       string                                 `json:"inspection_effectivity" form:"inspection_effectivity"`
	InspectionEffectivityByDM   map[string]string                      `json:"inspection_effectivity_by_damage_mechanism" form:"inspection_effectivity_by_damage_mechanism"`
	InspectionPlanByDM          map[string]PipelineInspectionPlanInput `json:"inspection_plan_by_damage_mechanism" form:"inspection_plan_by_damage_mechanism"`
	ReleaseFluid                string                                 `json:"release_fluid" form:"release_fluid"`
	GenericFailureFrequency     float64                                `json:"generic_failure_frequency" form:"generic_failure_frequency"`
	ManagementSystemScore       float64                                `json:"management_system_score" form:"management_system_score"`
	DamageFactor                float64                                `json:"damage_factor" form:"damage_factor"`
	BaseTPDRate                 float64                                `json:"base_tpd_rate" form:"base_tpd_rate"`
	BaseExternalCorrRate        float64                                `json:"base_external_corr_rate" form:"base_external_corr_rate"`
	BaseInternalCorrRate        float64                                `json:"base_internal_corr_rate" form:"base_internal_corr_rate"`
	DepthOfCover                string                                 `json:"depth_of_cover" form:"depth_of_cover"`
	PatrolFrequency             string                                 `json:"patrol_frequency" form:"patrol_frequency"`
	ROWCondition                string                                 `json:"row_condition" form:"row_condition"`
	SoilResistivity             string                                 `json:"soil_resistivity" form:"soil_resistivity"`
	CoatingCondition            string                                 `json:"coating_condition" form:"coating_condition"`
	CPStatus                    string                                 `json:"cp_status" form:"cp_status"`
	CPPotentialMV               float64                                `json:"cp_potential_mv" form:"cp_potential_mv"`
	FluidCorrosivity            string                                 `json:"fluid_corrosivity" form:"fluid_corrosivity"`
	WaterContent                string                                 `json:"water_content" form:"water_content"`
	CO2H2SPresence              string                                 `json:"co2_h2s_presence" form:"co2_h2s_presence"`
	MICRisk                     string                                 `json:"mic_risk" form:"mic_risk"`
	WallThicknessCondition      string                                 `json:"wall_thickness_condition" form:"wall_thickness_condition"`
	BuildingCountInsidePIR      int                                    `json:"building_count_inside_pir" form:"building_count_inside_pir"`
	ClassLocation               string                                 `json:"class_location" form:"class_location"`
	EmergencyResponse           string                                 `json:"emergency_response" form:"emergency_response"`
	FlowRate                    float64                                `json:"flow_rate" form:"flow_rate"`
	DetectionTimeHours          float64                                `json:"detection_time_hours" form:"detection_time_hours"`
	SegmentLengthBetweenValvesM float64                                `json:"segment_length_between_valves_m" form:"segment_length_between_valves_m"`
	EnvironmentalSensitivity    string                                 `json:"environmental_sensitivity" form:"environmental_sensitivity"`
	NearbySensitiveReceptor     bool                                   `json:"nearby_sensitive_receptor" form:"nearby_sensitive_receptor"`
	IsolationValveAvailable     bool                                   `json:"isolation_valve_available" form:"isolation_valve_available"`
	ConsequenceArea             float64                                `json:"consequence_area" form:"consequence_area"`
	ConsequenceFinancial        float64                                `json:"consequence_financial" form:"consequence_financial"`
	PoFCategory                 string                                 `json:"pof_category" form:"pof_category"`
	CoFCategory                 string                                 `json:"cof_category" form:"cof_category"`
	RiskRanking                 string                                 `json:"risk_ranking" form:"risk_ranking"`
	ConsequenceBasis            string                                 `json:"consequence_basis" form:"consequence_basis"`
	ProbabilityBasis            string                                 `json:"probability_basis" form:"probability_basis"`
	EngineeringNotes            string                                 `json:"engineering_notes" form:"engineering_notes"`
	RequiresConfirmation        bool                                   `json:"requires_confirmation" form:"requires_confirmation"`
	ConfirmationTODOReason      string                                 `json:"confirmation_todo_reason" form:"confirmation_todo_reason"`
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
	Code                  string  `json:"code"`
	Label                 string  `json:"label"`
	Category              string  `json:"category"`
	Severity              string  `json:"severity"`
	Score                 float64 `json:"score"`
	InspectionEffectivity string  `json:"inspection_effectivity"`
	Source                string  `json:"source"`
	Formula               string  `json:"formula"`
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
		PoF:                     "TODO_ENGINEERING_CONFIRMATION",
		CoF:                     "TODO_ENGINEERING_CONFIRMATION",
		RiskRanking:             "TODO_ENGINEERING_CONFIRMATION",
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
	dfTPD := calculateThirdPartyDamageFactor(input.RiskInput.BaseTPDRate, input.RiskInput.DepthOfCover, input.RiskInput.PatrolFrequency, input.RiskInput.ROWCondition)
	dfExternal := calculateExternalCorrosionFactor(input.RiskInput.BaseExternalCorrRate, input.RiskInput.SoilResistivity, input.RiskInput.CoatingCondition, input.RiskInput.CPStatus, input.RiskInput.CPPotentialMV)
	dfInternal := calculateInternalCorrosionFactor(input.RiskInput.BaseInternalCorrRate, input.RiskInput.FluidCorrosivity, input.RiskInput.WaterContent, input.RiskInput.CO2H2SPresence, input.RiskInput.MICRisk, input.RiskInput.WallThicknessCondition)
	pofValue, fms, governingDF, governingDriver := calculatePipelinePoF(input.RiskInput.GenericFailureFrequency, input.RiskInput.ManagementSystemScore, dfTPD, dfExternal, dfInternal)

	result.GenericFailureFrequency = input.RiskInput.GenericFailureFrequency
	result.ManagementSystemScore = input.RiskInput.ManagementSystemScore
	result.ManagementSystemFactor = fms
	result.SelectedDamageMechanism = PipelineDamageMechanismLabel(input.RiskInput.DamageMechanism)
	result.DamageMechanismResults = calculatePipelineDamageMechanismResults(input, dfTPD, dfExternal, dfInternal)
	result.InspectionPlanResults = calculatePipelineInspectionPlanResults(input, result.DamageMechanismResults)
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
		trace("pipeline_third_party_damage_df", "Pipeline MVP", "DF_TPD = Base_TPD_Rate / (Depth_Factor * Patrol_Factor * ROW_Factor)", map[string]interface{}{"base_tpd_rate": input.RiskInput.BaseTPDRate, "depth_of_cover": input.RiskInput.DepthOfCover, "patrol_frequency": input.RiskInput.PatrolFrequency, "row_condition": input.RiskInput.ROWCondition}, dfTPD, ""),
		trace("pipeline_external_corrosion_df", "Pipeline MVP", "DF_EXTERNAL = Base_Corr_Rate * Soil_Factor * Coating_Factor * CP_Factor", map[string]interface{}{"base_external_corr_rate": input.RiskInput.BaseExternalCorrRate, "soil_resistivity": input.RiskInput.SoilResistivity, "coating_condition": input.RiskInput.CoatingCondition, "cp_status": input.RiskInput.CPStatus, "cp_potential_mv": input.RiskInput.CPPotentialMV}, dfExternal, "CP potential around -850 mV or more negative is generally protective; non-compliant CP increases risk through CP status."),
		trace("pipeline_internal_corrosion_df", "Pipeline MVP", "DF_INTERNAL = Base_Internal_Corr_Rate * Fluid * Water * CO2/H2S * MIC * Wall", map[string]interface{}{"base_internal_corr_rate": input.RiskInput.BaseInternalCorrRate, "fluid_corrosivity": input.RiskInput.FluidCorrosivity, "water_content": input.RiskInput.WaterContent, "co2_h2s_presence": input.RiskInput.CO2H2SPresence, "mic_risk": input.RiskInput.MICRisk, "wall_thickness_condition": input.RiskInput.WallThicknessCondition}, dfInternal, ""),
		trace("pipeline_damage_mechanism_screening", "Pipeline system screening TODO_ENGINEERING_CONFIRMATION", "Each configured mechanism is screened from Pipeline-specific factor inputs and shown as NOT/Low/Moderate/High.", map[string]interface{}{"mechanism_count": len(result.DamageMechanismResults)}, result.DamageMechanismResults, "Screening supports UI prioritization; exact API 581/API 570 calculation linkage remains pending engineering confirmation."),
		trace("pipeline_inspection_scope_interval_method", "Pipeline system inspection planning TODO_ENGINEERING_CONFIRMATION", "Inspection method and interval are generated per damage mechanism from severity and selected method effectivity.", map[string]interface{}{"mechanism_count": len(result.InspectionPlanResults)}, result.InspectionPlanResults, "Intervals are planning aids pending engineering confirmation."),
		trace("pipeline_pof", "Pipeline MVP", "PoF = GFF * max(DF_TPD, DF_EXTERNAL, DF_INTERNAL) * FMS", map[string]interface{}{"gff": input.RiskInput.GenericFailureFrequency, "governing_df": governingDF, "fms": fms}, pofValue, ""),
		trace("pipeline_risk_ranking", "Pipeline MVP", "Risk = PoF Category x CoF Category", map[string]interface{}{"pof_category": result.PoF, "cof_category": result.CoF}, result.RiskRanking, ""),
		trace("pipeline_damage_mechanism_metadata", "TODO_ENGINEERING_CONFIRMATION", "Selected pipeline damage mechanism is stored as classification metadata only.", map[string]interface{}{"selected_damage_mechanism": PipelineDamageMechanismLabel(input.RiskInput.DamageMechanism), "configured_source": PipelineDamageMechanismSource}, PipelineDamageMechanismLabel(input.RiskInput.DamageMechanism), "Calculation impact is not linked until engineering confirms the mechanism-to-factor rules."),
		trace("pipeline_engineering_advisory", result.RecommendationSource, result.RecommendationRuleName, map[string]interface{}{"risk_level": result.FinalRiskLevel, "cof": result.CoF, "governing_driver": result.GoverningDamageMechanism, "selected_damage_mechanism": PipelineDamageMechanismLabel(input.RiskInput.DamageMechanism)}, result.Recommendation, "System-generated advisory; not an official RBI/API recommendation."),
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
			Source:                     "Pipeline system inspection planning TODO_ENGINEERING_CONFIRMATION",
		})
	}
	return results
}

func defaultPipelineNonIntrusiveMethod(code string) string {
	switch code {
	case "external_corrosion", "coating_cui_degradation":
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

func calculatePipelineDamageMechanismResults(input PipelineOilInput, dfTPD, dfExternal, dfInternal float64) []PipelineDamageMechanismResult {
	rlRisk := remainingLifeSeverityScore(input)
	flowRisk := severityInputScore(input.RiskInput.FlowRate, 100, 500, 1000)
	results := []PipelineDamageMechanismResult{}
	for _, option := range PipelineDamageMechanismOptions() {
		score, formula := 0.0, "TODO_ENGINEERING_CONFIRMATION"
		switch option.Code {
		case "external_corrosion":
			score = dfExternal
			formula = "DF_EXTERNAL = Base_Corr_Rate * Soil_Factor * Coating_Factor * CP_Factor"
		case "coating_cui_degradation":
			score = averagePositive(
				lookupPipelineFactor(pipelineConditionFactors, input.RiskInput.CoatingCondition),
				lookupPipelineFactor(pipelineSoilFactors, input.RiskInput.SoilResistivity),
				lookupPipelineFactor(pipelineCPFactors, input.RiskInput.CPStatus),
			)
			formula = "Average(Coating_Factor, Soil_Factor, CP_Factor)"
		case "third_party_mechanical_damage":
			score = dfTPD
			formula = "DF_TPD = Base_TPD_Rate / (Depth_Factor * Patrol_Factor * ROW_Factor)"
		case "internal_corrosion":
			score = dfInternal
			formula = "DF_INTERNAL = Base_Internal_Corr_Rate * Fluid * Water * CO2/H2S * MIC * Wall"
		case "localized_corrosion_pitting":
			score = averagePositive(dfInternal, rlRisk, lookupPipelineFactor(pipelineInternalFactors, input.RiskInput.WallThicknessCondition))
			formula = "Average(DF_INTERNAL, Remaining_Life_Severity, Wall_Thickness_Factor)"
		case "erosion_corrosion":
			score = averagePositive(flowRisk, lookupPipelineFactor(pipelineInternalFactors, input.RiskInput.FluidCorrosivity), lookupPipelineFactor(pipelineInternalFactors, input.RiskInput.WallThicknessCondition))
			formula = "Average(Flow_Rate_Severity, Fluid_Corrosivity_Factor, Wall_Thickness_Factor)"
		case "cracking_scc_fatigue":
			score = averagePositive(
				lookupPipelineFactor(pipelineInternalFactors, input.RiskInput.CO2H2SPresence),
				lookupPipelineFactor(pipelineInternalFactors, input.RiskInput.MICRisk),
				severityTextScore(input.RiskInput.CPStatus, map[string]float64{"failed": 3.5, "borderline": 1.9, "normal": 1.0}),
			)
			formula = "Average(CO2/H2S_Factor, MIC_Factor, CP_Stress_Screening_Factor)"
		default:
			score = 0
			formula = "Other mechanisms require engineering review."
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
			Severity:              pipelineSeverity(score),
			Score:                 score,
			InspectionEffectivity: effectivity,
			Source:                "Pipeline system screening TODO_ENGINEERING_CONFIRMATION",
			Formula:               formula,
		})
	}
	return results
}

func pipelineSeverity(score float64) string {
	switch {
	case score <= 0:
		return "NOT"
	case score < 1.5:
		return "Low"
	case score < 3:
		return "Moderate"
	default:
		return "High"
	}
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

func calculateThirdPartyDamageFactor(baseTPDRate float64, depthOfCover, patrolFrequency, rowCondition string) float64 {
	return baseTPDRate / (lookupPipelineFactor(pipelineDepthFactors, depthOfCover) * lookupPipelineFactor(pipelinePatrolFactors, patrolFrequency) * lookupPipelineFactor(pipelineROWFactors, rowCondition))
}

func calculateExternalCorrosionFactor(baseCorrRate float64, soilResistivity, coatingCondition, cpStatus string, cpPotentialMV float64) float64 {
	cpFactor := lookupPipelineFactor(pipelineCPFactors, cpStatus)
	if cpPotentialMV != 0 && cpPotentialMV > -850 && cpFactor < pipelineCPFactors["borderline"] {
		cpFactor = pipelineCPFactors["borderline"]
	}
	return baseCorrRate * lookupPipelineFactor(pipelineSoilFactors, soilResistivity) * lookupPipelineFactor(pipelineConditionFactors, coatingCondition) * cpFactor
}

func calculateInternalCorrosionFactor(baseInternalCorrRate float64, fluidCorrosivity, waterContent, co2H2SPresence, micRisk, wallThicknessCondition string) float64 {
	return baseInternalCorrRate *
		lookupPipelineFactor(pipelineInternalFactors, fluidCorrosivity) *
		lookupPipelineFactor(pipelineInternalFactors, waterContent) *
		lookupPipelineFactor(pipelineInternalFactors, co2H2SPresence) *
		lookupPipelineFactor(pipelineInternalFactors, micRisk) *
		lookupPipelineFactor(pipelineInternalFactors, wallThicknessCondition)
}

func calculatePipelinePoF(gff, managementSystemScore, dfTPD, dfExternal, dfInternal float64) (float64, float64, float64, string) {
	fms := managementSystemFactor(managementSystemScore)
	governingDF, governingDriver := dfTPD, "Third-Party Damage"
	if dfExternal > governingDF {
		governingDF, governingDriver = dfExternal, "External Corrosion"
	}
	if dfInternal > governingDF {
		governingDF, governingDriver = dfInternal, "Internal Corrosion"
	}
	return gff * governingDF * fms, fms, governingDF, governingDriver
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
	if strings.EqualFold(classLocation, "village") && cofNumeric(category) < 3 {
		category = "C"
	}
	if strings.EqualFold(classLocation, "urban_dense") && cofNumeric(category) < 4 {
		category = "D"
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
	case "Third-Party Damage":
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
	srcStr := "System advisory rule based on risk category, CoF factors, and governing pipeline damage-factor driver."

	if input.ManualRecommendation != "" {
		recStr = input.ManualRecommendation
		srcStr = "User overridden recommendation."
	}

	return pipelineEngineeringAdvisory{
		Groups:         groups,
		Recommendation: recStr,
		Source:         srcStr,
		RuleName:       "pipeline-system-advisory-v1 TODO_ENGINEERING_CONFIRMATION",
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
	validateOption("RiskInput.coating_condition", input.RiskInput.CoatingCondition, pipelineConditionFactors)
	validateOption("RiskInput.cp_status", input.RiskInput.CPStatus, pipelineCPFactors)
	validateOption("RiskInput.fluid_corrosivity", input.RiskInput.FluidCorrosivity, pipelineInternalFactors)
	validateOption("RiskInput.water_content", input.RiskInput.WaterContent, pipelineInternalFactors)
	validateOption("RiskInput.co2_h2s_presence", input.RiskInput.CO2H2SPresence, pipelineInternalFactors)
	validateOption("RiskInput.mic_risk", input.RiskInput.MICRisk, pipelineInternalFactors)
	validateOption("RiskInput.wall_thickness_condition", input.RiskInput.WallThicknessCondition, pipelineInternalFactors)
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
		input.RiskInput.CoatingCondition = "fair"
	}
	if input.RiskInput.CPStatus == "" {
		input.RiskInput.CPStatus = "normal"
	}
	if input.RiskInput.FluidCorrosivity == "" {
		input.RiskInput.FluidCorrosivity = "medium"
	}
	if input.RiskInput.WaterContent == "" {
		input.RiskInput.WaterContent = "medium"
	}
	if input.RiskInput.CO2H2SPresence == "" {
		input.RiskInput.CO2H2SPresence = "present"
	}
	if input.RiskInput.MICRisk == "" {
		input.RiskInput.MICRisk = "low"
	}
	if input.RiskInput.WallThicknessCondition == "" {
		input.RiskInput.WallThicknessCondition = "warning"
	}
	if input.RiskInput.ClassLocation == "" {
		input.RiskInput.ClassLocation = "remote"
	}
	if input.RiskInput.EmergencyResponse == "" {
		input.RiskInput.EmergencyResponse = "available"
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
	if val == "" {
		return 0
	}
	parts := strings.Split(val, "-")
	if len(parts) >= 1 {
		year, _ := strconv.ParseFloat(parts[0], 64)
		if len(parts) == 2 {
			month, _ := strconv.ParseFloat(parts[1], 64)
			return year + (month-1)/12.0
		}
		return year
	}
	return 0
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
		return "TODO_ENGINEERING_CONFIRMATION"
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
