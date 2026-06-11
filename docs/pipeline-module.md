# Pipeline Risk Assessment Module

## Overview

This module adds an isolated Pipeline Risk Assessment area beside the existing Pressure Vessel assessment module. It is an MVP simplified/index-based RiskInput module for gas, oil, and liquid pipeline segments.

The storage pattern remains the existing `pipeline_oil_assessments` table with full input/result JSON snapshots, so old records stay reproducible and no disruptive schema migration is required for the new factor fields.

## Main Files

- `controller/pipelineController.go`: route handlers, default sample data, draft/calculate flow.
- `models/pipeline_oil.go`: DTOs, repository, service orchestration, validation, calculation functions, formula trace.
- `templates/pipeline_assessment_form.html`: pipeline form with general data, PoF inputs, CoF inputs, and calculation actions.
- `templates/pipeline_assessment_detail.html`: result summary, DF drivers, PoF/CoF/risk ranking, recommendation, and formula trace.
- `assets/js/pipeline_oil_assessment.js`: form serialization into the existing JSON API payload.
- `models/pipeline_oil_test.go`: calculation and validation tests.

## Routes

- `GET /assessment-pipeline/list`
- `GET /assessment-pipeline/form`
- `GET /assessment-pipeline/gas`
- `POST /assessment-pipeline/submit`
- `GET /assessment-pipeline/view/:id`
- `GET /assessment-pipeline/edit/:id`
- `POST /assessment-pipeline/update/:id`
- `POST /assessment-pipeline/calculate/:id`
- `DELETE /assessment-pipeline/delete/:id`

## Calculation

Pipeline PoF uses:

`PoF = GFF * DF * FMS`

The governing damage factor is:

`DF = max(DF_TPD, DF_EXTERNAL_CORROSION, DF_INTERNAL_CORROSION)`

Implemented helper functions:

- `calculateThirdPartyDamageFactor()`
- `calculateExternalCorrosionFactor()`
- `calculateInternalCorrosionFactor()`
- `calculatePipelinePoF()`
- `calculateGasCoF()`
- `calculateLiquidCoF()`
- `calculatePipelineRiskRanking()`
- `generatePipelineRecommendation()`

Gas CoF uses PIR:

`PIR = 0.69 * outside_diameter_in * sqrt(operating_pressure_psi)`

Oil/liquid CoF uses spill volume:

`Spill Volume = flow_rate * detection_time + internal_pipeline_volume_between_valves`

Environmental sensitivity, nearby receptor, and isolation valve availability adjust liquid consequence before category assignment.

## Result Fields

The detail page shows:

- DF_TPD, DF_External_Corrosion, DF_Internal_Corrosion
- Governing DF and driver
- GFF, FMS, final PoF, and PoF category
- PIR for gas
- Spill volume and adjusted spill volume for oil/liquid
- CoF category
- Final risk code and level
- Recommendation based on the dominant damage mechanism and high CoF drivers
- Formula trace for auditability

## Validation

- Required draft fields: report number, line identification, service, assessment by.
- Service must be `Oil`, `Liquid`, or `Gas`.
- Calculation requires positive diameter, pressure, wall thickness, design factors, base rates, and GFF.
- Gas consequence validates building count.
- Oil/liquid consequence validates flow rate, detection time, valve segment length, and environmental sensitivity.
- Inspection point validation remains for the existing mechanical/thickness appraisal outputs.

## Manual QA Checklist

- Open Pipeline list and create a draft from sample values.
- Calculate an oil/liquid record and confirm spill volume, CoF, risk code, and recommendation.
- Open `/assessment-pipeline/gas`, calculate a gas record, and confirm PIR, CoF, risk code, and recommendation.
- Confirm formula trace includes the pipeline MVP formulas.
- Confirm Pressure Vessel list/form/detail still opens.

