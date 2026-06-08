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
	DamageMechanism        string `json:"damage_mechanism" form:"damage_mechanism"`
	InspectionEffectivity  string `json:"inspection_effectivity" form:"inspection_effectivity"`
	ReleaseFluid           string `json:"release_fluid" form:"release_fluid"`
	ConsequenceBasis       string `json:"consequence_basis" form:"consequence_basis"`
	ProbabilityBasis       string `json:"probability_basis" form:"probability_basis"`
	EngineeringNotes       string `json:"engineering_notes" form:"engineering_notes"`
	RequiresConfirmation   bool   `json:"requires_confirmation" form:"requires_confirmation"`
	ConfirmationTODOReason string `json:"confirmation_todo_reason" form:"confirmation_todo_reason"`
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
	PoF                         string                    `json:"pof"`
	CoF                         string                    `json:"cof"`
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
		},
	}

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
	if input.Service != "" && !strings.EqualFold(input.Service, "Oil") && !strings.EqualFold(input.Service, "Gas") {
		errs = append(errs, PipelineOilValidationError{Field: "service", Message: "must be Oil or Gas"})
	}
	return errs
}

func ValidatePipelineOilCalculation(input PipelineOilInput) []PipelineOilValidationError {
	errs := ValidatePipelineOilDraft(input)
	if !strings.EqualFold(input.Service, "Oil") {
		errs = append(errs, PipelineOilValidationError{Field: "service", Message: "TODO_ENGINEERING_CONFIRMATION: Gas formulas are not implemented yet"})
	}
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
	if input.YearUsed == 0 {
		input.YearUsed = input.YearBuilt
	}
	if input.RBI.ConfirmationTODOReason == "" {
		input.RBI.RequiresConfirmation = true
		input.RBI.ConfirmationTODOReason = "TODO_ENGINEERING_CONFIRMATION: RBI PoF/CoF/risk ranking formulas are not present in Excel."
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
