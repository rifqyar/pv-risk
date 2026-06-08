# Pipeline Oil & Gas Module

## Overview

This module adds an isolated Pipeline Oil & Gas assessment area beside the existing Pressure Vessel assessment module. Pipeline Oil is implemented from `1. pipeline oil.xlsx`; Pipeline Gas is registered as coming soon only.

## Folder Structure

- `controller/pipelineController.go`: Pipeline route handlers.
- `models/pipeline_oil.go`: Pipeline Oil DTOs, repository, service orchestration, validation, and calculation engine.
- `migrations/pipeline_oil_migration.go`: Pipeline Oil SQLite table and indexes.
- `templates/pipeline_assessment_*.html`: list, form, detail, and gas placeholder pages.
- `assets/js/pipeline_oil_assessment.js`: Pipeline Oil form payload and actions.
- `models/pipeline_oil_test.go`: formula and validation tests.

## Database Tables

- `pipeline_oil_assessments`

Stored columns include status, report number, line identification, owner/user, location, service, assessment by, formula version, input JSON, result JSON, formula trace JSON, snapshot JSON, audit timestamps, created_by, and updated_by.

The table stores a full calculation snapshot so old records remain reproducible if formulas change later.

## Routes

- `GET /assessment-pipeline/list`
- `GET /assessment-pipeline/form`
- `POST /assessment-pipeline/submit`
- `GET /assessment-pipeline/view/:id`
- `GET /assessment-pipeline/edit/:id`
- `POST /assessment-pipeline/update/:id`
- `POST /assessment-pipeline/calculate/:id`
- `DELETE /assessment-pipeline/delete/:id`
- `GET /assessment-pipeline/gas`

## Service Flow

1. Create draft validates required identity fields and stores input as JSON.
2. Update draft prevents changes to calculated or archived assessments.
3. Calculate validates engineering inputs, runs pure calculation functions, stores input/result/trace/snapshot in one transaction, and changes status to `CALCULATED`.
4. Archive marks the record `ARCHIVED`.

## Calculation Flow

Implemented formulas from workbook:

- `Input!G17`: pipe length feet = `((pipe_length_m*100)/2.54)/12`
- `Input!G23`: design temperature Celsius = `(5/9)*(design_temperature_f-32)`
- `Input!G29`: allowance mm = `allowance_in*25.4`
- `Input!G36`: outside diameter mm = `outside_diameter_in*25.4`
- `Input!D39`: nominal wall thickness inch = `nominal_wall_thickness_mm/25.4`
- `7 Appraisal!J63`: required thickness inch = `((P*D)/(2*F*E*SMYS))+c`
- `7 Appraisal!N25/N37/N38`: pressure/stress conversions = `psi/14.223`
- `2 Data!O30:O32`: corrosion rate = `(nominal_thickness_mm-actual_thickness_mm)/(measured_year-year_used)`
- `2 Data!S30:S32`: remaining life = `(actual_thickness_mm-required_thickness_mm)/corrosion_rate_mm_year`
- `7 Appraisal!O90:O92`: hoop stress = `(P*D)/(2*actual_thickness_in)`
- `7 Appraisal!O122:O124`: MAOP = `(2*actual_thickness_in*SMYS*F*E)/D`
- `7 Appraisal!H106`: highest hoop stress = `MAX(O90:Q104)`
- `7 Appraisal!H107`: percent SMYS = `(H106/K37)*100`
- `7 Appraisal!H138`: lowest MAOP = `MIN(O122:O136)`
- `7 Appraisal!O195/R195`: summary required thickness = `MIN(K159:M160)` and inch conversion
- `7 Appraisal!O197/O199/O201`: conclusion summary for hoop stress, MAOP, and remaining life

## Formula Versioning

Current formula version: `pipeline-oil-rbi581-v1`.

## Validation Rules

- Required: report number, line identification, service, assessment by.
- Service must be `Oil`.
- Positive values required for pressure, diameter, thickness, factors, stress, and length.
- Inspection measured year must be after year used.
- Actual thickness cannot exceed nominal thickness for workbook corrosion-rate formula.
- Extreme design pressure over 20,000 psig requires engineering confirmation.

## Known Limitations

- RBI 581 PoF formula is not present in the workbook.
- RBI 581 CoF formula is not present in the workbook.
- Risk ranking matrix/formula is not present in the workbook.
- `6 Verification` contains `#REF!` formulas, and downstream `2 Data` ranges for additional points also contain `#REF!`.

## Pipeline Gas Status

Pipeline Gas has a disabled/coming-soon route and page only. No formulas are implemented.

## TODO_ENGINEERING_CONFIRMATION

- Confirm RBI 581 PoF calculation for Pipeline Oil.
- Confirm RBI 581 CoF calculation for Pipeline Oil.
- Confirm risk matrix/ranking rules.
- Confirm how to handle workbook `#REF!` inspection rows.
- Confirm whether remaining-life calculation should continue using workbook `2 Data` required thickness values or design-appraisal required thickness for all future records.

## Adding Pipeline Gas Later

1. Add a gas workbook/formula source.
2. Create gas input/result structs beside Pipeline Oil.
3. Add gas calculation tests from workbook samples.
4. Add gas migration/table or shared pipeline table only after confirming storage requirements.
5. Replace the coming-soon handler/template with list/form/detail flow.

## Manual QA Checklist

- Open Pipeline Oil list.
- Create a draft from default workbook sample values.
- Edit the draft and save.
- Calculate the draft.
- Confirm detail page shows formula version, result, trace, and TODOs.
- Confirm calculated records cannot be overwritten from edit flow.
- Archive a record.
- Confirm Pressure Vessel list/form/detail still opens.
