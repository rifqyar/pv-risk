package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

const PipelineOilFormulaVersion = "pipeline-oil-rbi581-v1"

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

type PipelineOilInput struct {
	ID                        int                           `json:"id" form:"id"`
	ReportNo                  string                        `json:"report_no" form:"report_no"`
	PlaceIssued               string                        `json:"place_issued" form:"place_issued"`
	DateIssued                string                        `json:"date_issued" form:"date_issued"`
	OwnerUser                 string                        `json:"owner_user" form:"owner_user"`
	Contractor                string                        `json:"contractor" form:"contractor"`
	Location                  string                        `json:"location" form:"location"`
	LineIdentification        string                        `json:"line_identification" form:"line_identification"`
	YearBuilt                 int                           `json:"year_built" form:"year_built"`
	YearUsed                  int                           `json:"year_used" form:"year_used"`
	Service                   string                        `json:"service" form:"service"`
	PipeSize                  string                        `json:"pipe_size" form:"pipe_size"`
	PipeLengthM               float64                       `json:"pipe_length_m" form:"pipe_length_m"`
	MaterialSpecification     string                        `json:"material_specification" form:"material_specification"`
	FlangeMaterialSpec        string                        `json:"flange_material_spec" form:"flange_material_spec"`
	SMYSPsi                   float64                       `json:"smys_psi" form:"smys_psi"`
	InternalDesignPressurePsi float64                       `json:"internal_design_pressure_psi" form:"internal_design_pressure_psi"`
	DesignTemperatureF        float64                       `json:"design_temperature_f" form:"design_temperature_f"`
	TestPressurePsi           float64                       `json:"test_pressure_psi" form:"test_pressure_psi"`
	MethodOfJoining           string                        `json:"method_of_joining" form:"method_of_joining"`
	JointEfficiency           float64                       `json:"joint_efficiency" form:"joint_efficiency"`
	CoatingType               string                        `json:"coating_type" form:"coating_type"`
	CorrosionControl          string                        `json:"corrosion_control" form:"corrosion_control"`
	AllowanceIn               float64                       `json:"allowance_in" form:"allowance_in"`
	RightOfWay                string                        `json:"right_of_way" form:"right_of_way"`
	SafetyDevice              string                        `json:"safety_device" form:"safety_device"`
	AreaClassification        string                        `json:"area_classification" form:"area_classification"`
	InspectionPeriod          string                        `json:"inspection_period" form:"inspection_period"`
	InspectionResult          string                        `json:"inspection_result" form:"inspection_result"`
	ApplicableCode            string                        `json:"applicable_code" form:"applicable_code"`
	OutsideDiameterIn         float64                       `json:"outside_diameter_in" form:"outside_diameter_in"`
	OperatingPressurePsi      float64                       `json:"operating_pressure_psi" form:"operating_pressure_psi"`
	RadiographicPercent       float64                       `json:"radiographic_percent" form:"radiographic_percent"`
	NominalWallThicknessMM    float64                       `json:"nominal_wall_thickness_mm" form:"nominal_wall_thickness_mm"`
	ActualWallThicknessMM     float64                       `json:"actual_wall_thickness_mm" form:"actual_wall_thickness_mm"`
	TypeOfInstallation        string                        `json:"type_of_installation" form:"type_of_installation"`
	QualityFactor             float64                       `json:"quality_factor" form:"quality_factor"`
	WeldJointStrengthFactor   float64                       `json:"weld_joint_strength_factor" form:"weld_joint_strength_factor"`
	DesignFactor              float64                       `json:"design_factor" form:"design_factor"`
	MaterialStressPsi         float64                       `json:"material_stress_psi" form:"material_stress_psi"`
	PreviousSKPP              string                        `json:"previous_skpp" form:"previous_skpp"`
	ExpirationDate            string                        `json:"expiration_date" form:"expiration_date"`
	CorrosionRateMPY          *float64                      `json:"corrosion_rate_mpy" form:"corrosion_rate_mpy"`
	TemperatureDeratingFactor float64                       `json:"temperature_derating_factor" form:"temperature_derating_factor"`
	AssessmentBy              string                        `json:"assessment_by" form:"assessment_by"`
	InspectionPoints          []PipelineOilInspectionPoint  `json:"inspection_points"`
	RBI                       PipelineOilRBIStructuralInput `json:"rbi"`
}

type PipelineOilRBIStructuralInput struct {
	DamageMechanism             string  `json:"damage_mechanism" form:"damage_mechanism"`
	InspectionEffectivity       string  `json:"inspection_effectivity" form:"inspection_effectivity"`
	ReleaseFluid                string  `json:"release_fluid" form:"release_fluid"`
	GenericFailureFrequency     float64 `json:"generic_failure_frequency" form:"generic_failure_frequency"`
	ManagementSystemScore       float64 `json:"management_system_score" form:"management_system_score"`
	DamageFactor                float64 `json:"damage_factor" form:"damage_factor"`
	BaseTPDRate                 float64 `json:"base_tpd_rate" form:"base_tpd_rate"`
	BaseExternalCorrRate        float64 `json:"base_external_corr_rate" form:"base_external_corr_rate"`
	BaseInternalCorrRate        float64 `json:"base_internal_corr_rate" form:"base_internal_corr_rate"`
	DepthOfCover                string  `json:"depth_of_cover" form:"depth_of_cover"`
	PatrolFrequency             string  `json:"patrol_frequency" form:"patrol_frequency"`
	ROWCondition                string  `json:"row_condition" form:"row_condition"`
	SoilResistivity             string  `json:"soil_resistivity" form:"soil_resistivity"`
	CoatingCondition            string  `json:"coating_condition" form:"coating_condition"`
	CPStatus                    string  `json:"cp_status" form:"cp_status"`
	CPPotentialMV               float64 `json:"cp_potential_mv" form:"cp_potential_mv"`
	FluidCorrosivity            string  `json:"fluid_corrosivity" form:"fluid_corrosivity"`
	WaterContent                string  `json:"water_content" form:"water_content"`
	CO2H2SPresence              string  `json:"co2_h2s_presence" form:"co2_h2s_presence"`
	MICRisk                     string  `json:"mic_risk" form:"mic_risk"`
	WallThicknessCondition      string  `json:"wall_thickness_condition" form:"wall_thickness_condition"`
	BuildingCountInsidePIR      int     `json:"building_count_inside_pir" form:"building_count_inside_pir"`
	ClassLocation               string  `json:"class_location" form:"class_location"`
	EmergencyResponse           string  `json:"emergency_response" form:"emergency_response"`
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
}

type PipelineOilInspectionPoint struct {
	InspectionPoint     string  `json:"inspection_point" form:"inspection_point"`
	LocationClass       string  `json:"location_class" form:"location_class"`
	InstallationType    string  `json:"installation_type" form:"installation_type"`
	NominalThicknessMM  float64 `json:"nominal_thickness_mm" form:"nominal_thickness_mm"`
	RequiredThicknessMM float64 `json:"required_thickness_mm" form:"required_thickness_mm"`
	ActualThicknessMM   float64 `json:"actual_thickness_mm" form:"actual_thickness_mm"`
	MeasuredYear        int     `json:"measured_year" form:"measured_year"`
}

type PipelineOilResult struct {
	FormulaVersion              string                    `json:"formula_version"`
	CalculatedAt                time.Time                 `json:"calculated_at"`
	DesignTemperatureC          float64                   `json:"design_temperature_c"`
	PipeLengthFt                float64                   `json:"pipe_length_ft"`
	OutsideDiameterMM           float64                   `json:"outside_diameter_mm"`
	AllowanceMM                 float64                   `json:"allowance_mm"`
	NominalWallThicknessIn      float64                   `json:"nominal_wall_thickness_in"`
	DesignPressureKgCM2         float64                   `json:"design_pressure_kg_cm2"`
	OperatingPressureKgCM2      float64                   `json:"operating_pressure_kg_cm2"`
	SMYSKgCM2                   float64                   `json:"smys_kg_cm2"`
	MaterialStressKgCM2         float64                   `json:"material_stress_kg_cm2"`
	RequiredThicknessIn         float64                   `json:"required_thickness_in"`
	RequiredThicknessMM         float64                   `json:"required_thickness_mm"`
	SummaryRequiredThicknessIn  float64                   `json:"summary_required_thickness_in"`
	SummaryRequiredThicknessMM  float64                   `json:"summary_required_thickness_mm"`
	MinimumActualThicknessMM    float64                   `json:"minimum_actual_thickness_mm"`
	HighestCorrosionRateMMYear  float64                   `json:"highest_corrosion_rate_mm_year"`
	RemainingLifeYears          float64                   `json:"remaining_life_years"`
	HighestHoopStressPsi        float64                   `json:"highest_hoop_stress_psi"`
	HighestHoopStressKgCM2      float64                   `json:"highest_hoop_stress_kg_cm2"`
	HighestHoopStressPercentSMY float64                   `json:"highest_hoop_stress_percent_smys"`
	LowestMAOPPsi               float64                   `json:"lowest_maop_psi"`
	LowestMAOPKgCM2             float64                   `json:"lowest_maop_kg_cm2"`
	RequiredThicknessStatus     string                    `json:"required_thickness_status"`
	HoopStressStatus            string                    `json:"hoop_stress_status"`
	MAOPStatus                  string                    `json:"maop_status"`
	GenericFailureFrequency     float64                   `json:"generic_failure_frequency"`
	ManagementSystemScore       float64                   `json:"management_system_score"`
	ManagementSystemFactor      float64                   `json:"management_system_factor"`
	DamageFactor                float64                   `json:"damage_factor"`
	ThirdPartyDamageFactor      float64                   `json:"third_party_damage_factor"`
	ExternalCorrosionFactor     float64                   `json:"external_corrosion_factor"`
	InternalCorrosionFactor     float64                   `json:"internal_corrosion_factor"`
	GoverningDamageFactor       float64                   `json:"governing_damage_factor"`
	GoverningDamageMechanism    string                    `json:"governing_damage_mechanism"`
	PoFValue                    float64                   `json:"pof_value"`
	CoFValue                    float64                   `json:"cof_value"`
	PIRFeet                     float64                   `json:"pir_feet"`
	SpillVolume                 float64                   `json:"spill_volume"`
	AdjustedSpillVolume         float64                   `json:"adjusted_spill_volume"`
	RiskValue                   float64                   `json:"risk_value"`
	PoF                         string                    `json:"pof"`
	CoF                         string                    `json:"cof"`
	FinalRiskCode               string                    `json:"final_risk_code"`
	FinalRiskLevel              string                    `json:"final_risk_level"`
	RiskRanking                 string                    `json:"risk_ranking"`
	InspectionEffectiveness     string                    `json:"inspection_effectiveness"`
	Recommendation              string                    `json:"recommendation"`
	PointResults                []PipelineOilPointResult  `json:"point_results"`
	FormulaTrace                []PipelineOilFormulaTrace `json:"formula_trace"`
	TODOEngineeringConfirmation []string                  `json:"todo_engineering_confirmation"`
}

type PipelineOilPointResult struct {
	InspectionPoint       string  `json:"inspection_point"`
	RequiredThicknessIn   float64 `json:"required_thickness_in"`
	RequiredThicknessMM   float64 `json:"required_thickness_mm"`
	AppraisalThicknessIn  float64 `json:"appraisal_thickness_in"`
	AppraisalThicknessMM  float64 `json:"appraisal_thickness_mm"`
	ActualThicknessIn     float64 `json:"actual_thickness_in"`
	ActualThicknessMM     float64 `json:"actual_thickness_mm"`
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
		WHERE id=? AND status=?`,
		PipelineOilStatusCalculated, input.ReportNo, input.LineIdentification, input.OwnerUser,
		input.Location, input.Service, input.AssessmentBy, PipelineOilFormulaVersion, inputJSON,
		resultJSON, traceJSON, snapshotJSON, input.AssessmentBy, id, PipelineOilStatusDraft,
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
		InspectionEffectiveness: input.RBI.InspectionEffectivity,
		TODOEngineeringConfirmation: []string{
			"RBI 581 probability of failure formula is not present in workbook.",
			"RBI 581 consequence of failure formula is not present in workbook.",
			"Risk ranking matrix/formula is not present in workbook.",
			"Workbook contains #REF! formulas in sheet '6 Verification' and downstream '2 Data' ranges for additional inspection rows.",
			"API 581 exact GFF tables, damage-factor tables, CoF category thresholds, and risk matrix thresholds require licensed engineering data.",
		},
	}

	applyPipelineIndexRisk(input, result)

	tReqIn := requiredThicknessIn(input.InternalDesignPressurePsi, input.OutsideDiameterIn, input.DesignFactor, input.QualityFactor, input.SMYSPsi, input.AllowanceIn)
	result.RequiredThicknessIn = tReqIn
	result.RequiredThicknessMM = tReqIn * 25.4
	result.FormulaTrace = append(result.FormulaTrace,
		trace("pipe_length_ft", "Input!G17", "((D17*100)/2.54)/12", map[string]interface{}{"pipe_length_m": input.PipeLengthM}, result.PipeLengthFt, ""),
		trace("design_temperature_c", "Input!G23", "(5/9)*(D23-32)", map[string]interface{}{"design_temperature_f": input.DesignTemperatureF}, result.DesignTemperatureC, ""),
		trace("outside_diameter_mm", "Input!G36", "D36*25.4", map[string]interface{}{"outside_diameter_in": input.OutsideDiameterIn}, result.OutsideDiameterMM, ""),
		trace("allowance_mm", "Input!G29", "D29*25.4", map[string]interface{}{"allowance_in": input.AllowanceIn}, result.AllowanceMM, ""),
		trace("required_thickness", "7 Appraisal!J63", "((P*D)/(2*F*E*SMYS))+c", map[string]interface{}{"P": input.InternalDesignPressurePsi, "D": input.OutsideDiameterIn, "F": input.DesignFactor, "E": input.QualityFactor, "SMYS": input.SMYSPsi, "c": input.AllowanceIn}, result.RequiredThicknessIn, ""),
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
		maop := maopPsi(actualIn, input.SMYSPsi, input.DesignFactor, input.QualityFactor, input.OutsideDiameterIn)

		pr := PipelineOilPointResult{
			InspectionPoint:       point.InspectionPoint,
			RequiredThicknessIn:   result.RequiredThicknessIn,
			RequiredThicknessMM:   result.RequiredThicknessMM,
			AppraisalThicknessIn:  appraisalRequiredIn,
			AppraisalThicknessMM:  appraisalRequiredMM,
			ActualThicknessIn:     actualIn,
			ActualThicknessMM:     point.ActualThicknessMM,
			CorrosionRateMMYear:   cr,
			RemainingLifeYears:    rl,
			HoopStressPsi:         hs,
			MAOPPsi:               maop,
			ThicknessStatus:       acceptable(actualIn > appraisalRequiredIn),
			HoopStressStatus:      acceptable(hs < input.SMYSPsi),
			MAOPStatus:            acceptable(maop > input.InternalDesignPressurePsi),
			SourceInspectionPoint: point.InspectionPoint,
		}
		result.PointResults = append(result.PointResults, pr)
		result.FormulaTrace = append(result.FormulaTrace,
			trace("corrosion_rate", "2 Data!O30:O32", "(nominal_thickness_mm-actual_thickness_mm)/(measured_year-year_used)", map[string]interface{}{"point": point.InspectionPoint, "nominal_thickness_mm": point.NominalThicknessMM, "actual_thickness_mm": point.ActualThicknessMM, "measured_year": point.MeasuredYear, "year_used": input.YearUsed}, cr, ""),
			trace("remaining_life", "2 Data!S30:S32 / 7 Appraisal!R159:R160", "(actual_thickness_mm-required_thickness_mm)/corrosion_rate_mm_year", map[string]interface{}{"point": point.InspectionPoint, "actual_thickness_mm": point.ActualThicknessMM, "required_thickness_mm": remainingLifeBasisMM, "corrosion_rate_mm_year": cr}, rl, "Workbook remaining life references '2 Data' required thickness where provided."),
			trace("hoop_stress", "7 Appraisal!O90:O92", "(P*D)/(2*actual_thickness_in)", map[string]interface{}{"point": point.InspectionPoint, "P": input.InternalDesignPressurePsi, "D": input.OutsideDiameterIn, "actual_thickness_in": actualIn}, hs, ""),
			trace("maop", "7 Appraisal!O122:O124", "(2*actual_thickness_in*SMYS*F*E)/D", map[string]interface{}{"point": point.InspectionPoint, "actual_thickness_in": actualIn, "SMYS": input.SMYSPsi, "F": input.DesignFactor, "E": input.QualityFactor, "D": input.OutsideDiameterIn}, maop, ""),
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
	result.Recommendation = "CAPABLE TO CONTINUE SERVICE AND SAFE TO BE OPERATED IN PRESSURE NOT GREATER THAN INTERNAL DESIGN PRESSURE (P) OR MAOP, WHICHEVER SMALLER."
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
	dfTPD := calculateThirdPartyDamageFactor(input.RBI.BaseTPDRate, input.RBI.DepthOfCover, input.RBI.PatrolFrequency, input.RBI.ROWCondition)
	dfExternal := calculateExternalCorrosionFactor(input.RBI.BaseExternalCorrRate, input.RBI.SoilResistivity, input.RBI.CoatingCondition, input.RBI.CPStatus, input.RBI.CPPotentialMV)
	dfInternal := calculateInternalCorrosionFactor(input.RBI.BaseInternalCorrRate, input.RBI.FluidCorrosivity, input.RBI.WaterContent, input.RBI.CO2H2SPresence, input.RBI.MICRisk, input.RBI.WallThicknessCondition)
	pofValue, fms, governingDF, governingDriver := calculatePipelinePoF(input.RBI.GenericFailureFrequency, input.RBI.ManagementSystemScore, dfTPD, dfExternal, dfInternal)

	result.GenericFailureFrequency = input.RBI.GenericFailureFrequency
	result.ManagementSystemScore = input.RBI.ManagementSystemScore
	result.ManagementSystemFactor = fms
	result.ThirdPartyDamageFactor = dfTPD
	result.ExternalCorrosionFactor = dfExternal
	result.InternalCorrosionFactor = dfInternal
	result.GoverningDamageFactor = governingDF
	result.GoverningDamageMechanism = governingDriver
	result.DamageFactor = governingDF
	result.PoFValue = pofValue
	result.PoF = pofCategory(pofValue)

	if isGasService(input.Service) {
		cofCategory, pir := calculateGasCoF(input.OutsideDiameterIn, input.OperatingPressurePsi, input.RBI.BuildingCountInsidePIR, input.RBI.ClassLocation)
		result.PIRFeet = pir
		result.CoF = cofCategory
		result.CoFValue = cofNumeric(cofCategory)
	} else {
		cofCategory, spillVolume, adjustedSpillVolume := calculateLiquidCoF(input.RBI.FlowRate, input.RBI.DetectionTimeHours, input.OutsideDiameterIn, input.RBI.SegmentLengthBetweenValvesM, input.RBI.EnvironmentalSensitivity, input.RBI.NearbySensitiveReceptor, input.RBI.IsolationValveAvailable)
		result.SpillVolume = spillVolume
		result.AdjustedSpillVolume = adjustedSpillVolume
		result.CoF = cofCategory
		result.CoFValue = adjustedSpillVolume
	}

	result.RiskValue = pofNumeric(result.PoF) * cofNumeric(result.CoF)
	result.FinalRiskCode, result.FinalRiskLevel = calculatePipelineRiskRanking(result.PoF, result.CoF)
	result.RiskRanking = result.FinalRiskCode + " - " + result.FinalRiskLevel
	result.Recommendation = generatePipelineRecommendation(result, input.Service)
	result.TODOEngineeringConfirmation = nil
	result.FormulaTrace = append(result.FormulaTrace,
		trace("pipeline_third_party_damage_df", "Pipeline MVP", "DF_TPD = Base_TPD_Rate / (Depth_Factor * Patrol_Factor * ROW_Factor)", map[string]interface{}{"base_tpd_rate": input.RBI.BaseTPDRate, "depth_of_cover": input.RBI.DepthOfCover, "patrol_frequency": input.RBI.PatrolFrequency, "row_condition": input.RBI.ROWCondition}, dfTPD, ""),
		trace("pipeline_external_corrosion_df", "Pipeline MVP", "DF_EXTERNAL = Base_Corr_Rate * Soil_Factor * Coating_Factor * CP_Factor", map[string]interface{}{"base_external_corr_rate": input.RBI.BaseExternalCorrRate, "soil_resistivity": input.RBI.SoilResistivity, "coating_condition": input.RBI.CoatingCondition, "cp_status": input.RBI.CPStatus, "cp_potential_mv": input.RBI.CPPotentialMV}, dfExternal, "CP potential around -850 mV or more negative is generally protective; non-compliant CP increases risk through CP status."),
		trace("pipeline_internal_corrosion_df", "Pipeline MVP", "DF_INTERNAL = Base_Internal_Corr_Rate * Fluid * Water * CO2/H2S * MIC * Wall", map[string]interface{}{"base_internal_corr_rate": input.RBI.BaseInternalCorrRate, "fluid_corrosivity": input.RBI.FluidCorrosivity, "water_content": input.RBI.WaterContent, "co2_h2s_presence": input.RBI.CO2H2SPresence, "mic_risk": input.RBI.MICRisk, "wall_thickness_condition": input.RBI.WallThicknessCondition}, dfInternal, ""),
		trace("pipeline_pof", "Pipeline MVP", "PoF = GFF * max(DF_TPD, DF_EXTERNAL, DF_INTERNAL) * FMS", map[string]interface{}{"gff": input.RBI.GenericFailureFrequency, "governing_df": governingDF, "fms": fms}, pofValue, ""),
		trace("pipeline_risk_ranking", "Pipeline MVP", "Risk = PoF Category x CoF Category", map[string]interface{}{"pof_category": result.PoF, "cof_category": result.CoF}, result.RiskRanking, ""),
	)
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

func generatePipelineRecommendation(result *PipelineOilResult, service string) string {
	recommendations := []string{}
	switch result.GoverningDamageMechanism {
	case "Third-Party Damage":
		recommendations = append(recommendations, "Improve ROW markers, increase patrol frequency, strengthen excavation permit control, add warning signs, and consider deeper cover where applicable.")
	case "External Corrosion":
		recommendations = append(recommendations, "Perform coating inspection/repair, CIPS/DCVG survey, CP verification, and soil/corrosion monitoring.")
	case "Internal Corrosion":
		recommendations = append(recommendations, "Review fluid analysis, inhibitor performance, pigging schedule, wall thickness inspection, and CO2/H2S/MIC monitoring.")
	}
	if isGasService(service) && cofNumeric(result.CoF) >= 4 {
		recommendations = append(recommendations, "Review class location, emergency response, public awareness, isolation valve spacing, and populated-area protection.")
	}
	if !isGasService(service) && cofNumeric(result.CoF) >= 4 {
		recommendations = append(recommendations, "Improve leak detection, shorten isolation time, prepare spill containment, protect drainage/river receptors, and maintain emergency spill response readiness.")
	}
	return strings.Join(recommendations, " ")
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
	return strings.EqualFold(service, "Gas")
}

func applyAPI581PublicMethodology(input PipelineOilInput, result *PipelineOilResult) {
	result.GenericFailureFrequency = input.RBI.GenericFailureFrequency
	result.ManagementSystemScore = input.RBI.ManagementSystemScore
	result.DamageFactor = input.RBI.DamageFactor

	if input.RBI.ManagementSystemScore > 0 {
		pscore := (input.RBI.ManagementSystemScore / 1000) * 100
		result.ManagementSystemFactor = math.Pow(10, (-0.02*pscore)+1)
		result.FormulaTrace = append(result.FormulaTrace,
			trace("api581_management_system_factor", "API 581 public methodology", "FMS = 10^((-0.02*(score/1000*100))+1)", map[string]interface{}{"management_system_score": input.RBI.ManagementSystemScore, "pscore": pscore}, result.ManagementSystemFactor, "Uses public API 581 methodology summary; scoring source remains engineer supplied."),
		)
	}

	if input.RBI.GenericFailureFrequency > 0 && result.ManagementSystemFactor > 0 && input.RBI.DamageFactor > 0 {
		result.PoFValue = input.RBI.GenericFailureFrequency * result.ManagementSystemFactor * input.RBI.DamageFactor
		result.PoF = formatEngineeringValue(result.PoFValue)
		result.FormulaTrace = append(result.FormulaTrace,
			trace("api581_probability_of_failure", "API 581 public methodology", "PoF(t) = GFF * FMS * Df(t)", map[string]interface{}{"gff": input.RBI.GenericFailureFrequency, "fms": result.ManagementSystemFactor, "damage_factor": input.RBI.DamageFactor}, result.PoFValue, "GFF and Df are engineer supplied because API 581 lookup tables are not included in the workbook."),
		)
	}

	switch {
	case input.RBI.ConsequenceFinancial > 0:
		result.CoFValue = input.RBI.ConsequenceFinancial
		result.CoF = formatEngineeringValue(result.CoFValue)
		result.FormulaTrace = append(result.FormulaTrace,
			trace("api581_consequence_financial", "API 581 public methodology", "CoF = financial consequence input", map[string]interface{}{"consequence_financial": input.RBI.ConsequenceFinancial}, result.CoFValue, "Financial consequence is engineer supplied; API 581 detailed Level 1/2 consequence model is not in workbook."),
		)
	case input.RBI.ConsequenceArea > 0:
		result.CoFValue = input.RBI.ConsequenceArea
		result.CoF = formatEngineeringValue(result.CoFValue)
		result.FormulaTrace = append(result.FormulaTrace,
			trace("api581_consequence_area", "API 581 public methodology", "CoF = affected area consequence input", map[string]interface{}{"consequence_area": input.RBI.ConsequenceArea}, result.CoFValue, "Affected area consequence is engineer supplied; API 581 detailed Level 1/2 consequence model is not in workbook."),
		)
	}

	if result.PoFValue > 0 && result.CoFValue > 0 {
		result.RiskValue = result.PoFValue * result.CoFValue
		result.FormulaTrace = append(result.FormulaTrace,
			trace("api581_risk", "API 581 public methodology", "Risk = PoF * CoF", map[string]interface{}{"pof": result.PoFValue, "cof": result.CoFValue}, result.RiskValue, ""),
		)
	}

	if input.RBI.PoFCategory != "" {
		result.PoF = input.RBI.PoFCategory
	}
	if input.RBI.CoFCategory != "" {
		result.CoF = input.RBI.CoFCategory
	}
	if input.RBI.RiskRanking != "" {
		result.RiskRanking = input.RBI.RiskRanking
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
	if input.Service != "" && !strings.EqualFold(input.Service, "Oil") && !strings.EqualFold(input.Service, "Liquid") && !strings.EqualFold(input.Service, "Gas") {
		errs = append(errs, PipelineOilValidationError{Field: "service", Message: "must be Oil, Liquid, or Gas"})
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
		{"design_factor", input.DesignFactor},
		{"material_stress_psi", input.MaterialStressPsi},
	}
	for _, item := range checkPositive {
		if item.value <= 0 {
			errs = append(errs, PipelineOilValidationError{Field: item.field, Message: "must be greater than zero"})
		}
	}
	if input.YearBuilt <= 0 || input.YearUsed <= 0 {
		errs = append(errs, PipelineOilValidationError{Field: "year_built/year_used", Message: "valid year built and used are required"})
	}
	if input.YearUsed < input.YearBuilt {
		errs = append(errs, PipelineOilValidationError{Field: "year_used", Message: "cannot be before year built"})
	}
	if input.OperatingPressurePsi < 0 {
		errs = append(errs, PipelineOilValidationError{Field: "operating_pressure_psi", Message: "cannot be negative"})
	}
	if input.InternalDesignPressurePsi > 20000 {
		errs = append(errs, PipelineOilValidationError{Field: "internal_design_pressure_psi", Message: "extreme pressure requires engineering confirmation"})
	}
	if input.RBI.ManagementSystemScore < 0 || input.RBI.ManagementSystemScore > 1000 {
		errs = append(errs, PipelineOilValidationError{Field: "rbi.management_system_score", Message: "must be between 0 and 1000"})
	}
	if input.RBI.GenericFailureFrequency <= 0 {
		errs = append(errs, PipelineOilValidationError{Field: "rbi.generic_failure_frequency", Message: "must be greater than zero"})
	}
	if input.RBI.BaseTPDRate <= 0 || input.RBI.BaseExternalCorrRate <= 0 || input.RBI.BaseInternalCorrRate <= 0 {
		errs = append(errs, PipelineOilValidationError{Field: "rbi.base_rates", Message: "must be greater than zero"})
	}
	validateOption := func(field, value string, allowed map[string]float64) {
		if _, ok := allowed[strings.TrimSpace(value)]; !ok {
			if _, ok := allowed[strings.ToLower(strings.TrimSpace(value))]; !ok {
				errs = append(errs, PipelineOilValidationError{Field: field, Message: "invalid option"})
			}
		}
	}
	validateOption("rbi.depth_of_cover", input.RBI.DepthOfCover, pipelineDepthFactors)
	validateOption("rbi.patrol_frequency", input.RBI.PatrolFrequency, pipelinePatrolFactors)
	validateOption("rbi.row_condition", input.RBI.ROWCondition, pipelineROWFactors)
	validateOption("rbi.soil_resistivity", input.RBI.SoilResistivity, pipelineSoilFactors)
	validateOption("rbi.coating_condition", input.RBI.CoatingCondition, pipelineConditionFactors)
	validateOption("rbi.cp_status", input.RBI.CPStatus, pipelineCPFactors)
	validateOption("rbi.fluid_corrosivity", input.RBI.FluidCorrosivity, pipelineInternalFactors)
	validateOption("rbi.water_content", input.RBI.WaterContent, pipelineInternalFactors)
	validateOption("rbi.co2_h2s_presence", input.RBI.CO2H2SPresence, pipelineInternalFactors)
	validateOption("rbi.mic_risk", input.RBI.MICRisk, pipelineInternalFactors)
	validateOption("rbi.wall_thickness_condition", input.RBI.WallThicknessCondition, pipelineInternalFactors)
	if isGasService(input.Service) {
		if input.RBI.BuildingCountInsidePIR < 0 {
			errs = append(errs, PipelineOilValidationError{Field: "rbi.building_count_inside_pir", Message: "cannot be negative"})
		}
	} else {
		if input.RBI.FlowRate <= 0 {
			errs = append(errs, PipelineOilValidationError{Field: "rbi.flow_rate", Message: "must be greater than zero"})
		}
		if input.RBI.DetectionTimeHours <= 0 {
			errs = append(errs, PipelineOilValidationError{Field: "rbi.detection_time_hours", Message: "must be greater than zero"})
		}
		if input.RBI.SegmentLengthBetweenValvesM <= 0 {
			errs = append(errs, PipelineOilValidationError{Field: "rbi.segment_length_between_valves_m", Message: "must be greater than zero"})
		}
		validateOption("rbi.environmental_sensitivity", input.RBI.EnvironmentalSensitivity, pipelineEnvironmentalMultipliers)
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
		if point.MeasuredYear <= input.YearUsed {
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
	if input.ApplicableCode == "" {
		input.ApplicableCode = "ASME B31.4"
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
	if input.RBI.GenericFailureFrequency == 0 {
		input.RBI.GenericFailureFrequency = defaultPipelineGFF
	}
	if input.RBI.ManagementSystemScore == 0 {
		input.RBI.ManagementSystemScore = defaultPipelineManagementScore
	}
	if input.RBI.BaseTPDRate == 0 {
		input.RBI.BaseTPDRate = defaultPipelineBaseTPDRate
	}
	if input.RBI.BaseExternalCorrRate == 0 {
		input.RBI.BaseExternalCorrRate = defaultPipelineBaseCorrRate
	}
	if input.RBI.BaseInternalCorrRate == 0 {
		input.RBI.BaseInternalCorrRate = defaultPipelineBaseCorrRate
	}
	if input.RBI.DepthOfCover == "" {
		input.RBI.DepthOfCover = "1-2m"
	}
	if input.RBI.PatrolFrequency == "" {
		input.RBI.PatrolFrequency = "monthly"
	}
	if input.RBI.ROWCondition == "" {
		input.RBI.ROWCondition = "fair"
	}
	if input.RBI.SoilResistivity == "" {
		input.RBI.SoilResistivity = "1000-5000"
	}
	if input.RBI.CoatingCondition == "" {
		input.RBI.CoatingCondition = "fair"
	}
	if input.RBI.CPStatus == "" {
		input.RBI.CPStatus = "normal"
	}
	if input.RBI.FluidCorrosivity == "" {
		input.RBI.FluidCorrosivity = "medium"
	}
	if input.RBI.WaterContent == "" {
		input.RBI.WaterContent = "medium"
	}
	if input.RBI.CO2H2SPresence == "" {
		input.RBI.CO2H2SPresence = "present"
	}
	if input.RBI.MICRisk == "" {
		input.RBI.MICRisk = "low"
	}
	if input.RBI.WallThicknessCondition == "" {
		input.RBI.WallThicknessCondition = "warning"
	}
	if input.RBI.ClassLocation == "" {
		input.RBI.ClassLocation = "remote"
	}
	if input.RBI.EmergencyResponse == "" {
		input.RBI.EmergencyResponse = "available"
	}
	if input.RBI.FlowRate == 0 {
		input.RBI.FlowRate = 100
	}
	if input.RBI.DetectionTimeHours == 0 {
		input.RBI.DetectionTimeHours = 1
	}
	if input.RBI.SegmentLengthBetweenValvesM == 0 {
		input.RBI.SegmentLengthBetweenValvesM = input.PipeLengthM
	}
	if input.RBI.EnvironmentalSensitivity == "" {
		input.RBI.EnvironmentalSensitivity = "medium"
	}
	if input.YearUsed == 0 {
		input.YearUsed = input.YearBuilt
	}
	for i := range input.InspectionPoints {
		if input.InspectionPoints[i].NominalThicknessMM == 0 {
			input.InspectionPoints[i].NominalThicknessMM = input.NominalWallThicknessMM
		}
		if input.InspectionPoints[i].MeasuredYear == 0 {
			input.InspectionPoints[i].MeasuredYear = time.Now().Year()
		}
	}
}

func requiredThicknessIn(p, d, f, e, s, c float64) float64 {
	return ((p * d) / (2 * f * e * s)) + c
}

func hoopStressPsi(p, d, actualThicknessIn float64) float64 {
	return (p * d) / (2 * actualThicknessIn)
}

func maopPsi(actualThicknessIn, s, f, e, d float64) float64 {
	return (2 * actualThicknessIn * s * f * e) / d
}

func psiToKgCM2(psi float64) float64 {
	return psi / 14.223
}

func roundToPlaces(value float64, places int) float64 {
	factor := math.Pow10(places)
	return math.Round(value*factor) / factor
}

func corrosionRateMMYear(nominal, actual float64, yearUsed, measuredYear int) float64 {
	return (nominal - actual) / float64(measuredYear-yearUsed)
}

func remainingLifeYears(actual, required, corrosionRate float64) float64 {
	if corrosionRate <= 0 {
		return 0
	}
	return (actual - required) / corrosionRate
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
