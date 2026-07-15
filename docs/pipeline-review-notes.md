# Pipeline Review Notes

## Screenshot Review

Two Pipeline Oil & Gas screens were reviewed:

- Step 1 showed `Inspection Result = ACCEPTABLE` inside Pipeline Design Data.
- Step 4/detail showed recommendation text beside CoF breakdown without a visible source.

## Inspection Result

`Inspection Result` was removed from Step 1 because it is not a true pipeline input. The previous value came from workbook/report-style output and was being seeded as a default input, which could imply the assessment was already accepted before calculation.

It is now a calculated result field only. The Pipeline result derives it from mechanical appraisal statuses:

- `ACCEPTABLE` when thickness, hoop stress, and MAOP statuses are acceptable.
- `NOT ACCEPTABLE` when any calculated status is not acceptable.
- `REVIEW REQUIRED` for unexpected non-empty status values.

## Field Classification

INPUT:
`report_no`, `place_issued`, `date_issued`, `owner_user`, `contractor`, `location`, `line_identification`, `year_built`, `year_used`, `service`, `pipe_size`, `pipe_length_m`, `material_specification`, `flange_material_spec`, `smys_psi`, `internal_design_pressure_psi`, `design_temperature_f`, `test_pressure_psi`, `method_of_joining`, `joint_efficiency`, `coating_type`, `corrosion_control`, `allowance_in`, `right_of_way`, `safety_device`, `area_classification`, `inspection_period`, `applicable_code`, `outside_diameter_in`, `operating_pressure_psi`, `radiographic_percent`, `nominal_wall_thickness_mm`, `actual_wall_thickness_mm`, `type_of_installation`, `quality_factor`, `weld_joint_strength_factor`, `design_factor`, `material_stress_psi`, `previous_skpp`, `expiration_date`, `temperature_derating_factor`, `assessment_by`, and inspection point readings.

INPUT metadata:
`RiskInput.damage_mechanism` and `RiskInput.inspection_effectivity`.

CALCULATED_OUTPUT:
`inspection_result`, required thickness, corrosion rate, remaining life, hoop stress, MAOP, PoF, CoF, risk code, risk level, governing damage-factor driver, formula trace, and advisory source metadata.

REPORT_ONLY:
PDF/display layout labels, prepared/validated-by display fields, and historical report identifiers.

CONFIG_REFERENCE:
Pipeline damage mechanism option list and factor option maps for depth, patrol, ROW, soil, coating, CP, internal corrosion, and environmental sensitivity.

Engineering note:
Pipeline risk and CoF category thresholds now reuse the approved pressure-vessel threshold source. Exact API 581 GFF and detailed damage-factor lookup tables remain outside the workbook implementation unless supplied as approved licensed data.

## Recommendation Source

Recommendation text is code-based, not Excel-based and not verified as a direct RBI 571, RBI 581, API 570, ASME B31.4, or ASME B31.8 recommendation.

It is now labeled:

`Source: System advisory rule based on risk category, CoF factors, and governing pipeline damage-factor driver.`

Rule name:

`pipeline-system-advisory-v1 TODO_ENGINEERING_CONFIRMATION`

The UI title was changed to `Engineering Advisory / Suggested Actions` to avoid presenting it as an official RBI/API requirement.

## Pipeline Damage Mechanism

Pipeline damage mechanisms are implemented in `models/pipeline_damage_mechanism.go` as Pipeline-specific classification metadata. The form is intentionally shaped like the Pressure Vessel damage mechanism screen by grouping mechanisms into:

- External Damage
- Internal Thinning
- Internal Cracking

The Pipeline options were reduced to avoid an overly long dropdown and to stay closer to the Pressure Vessel grouping style:

- External Corrosion
- Coating / CUI Degradation
- Third-Party / Mechanical Damage
- Internal Corrosion
- Localized Corrosion / Pitting
- Erosion / Erosion-Corrosion
- Cracking / SCC / Fatigue
- Other / Engineering Review

Pipeline now calculates every configured mechanism. Users no longer choose which damage mechanism to calculate. The backend stores `damage_mechanism_results` in the existing Pipeline result JSON, including severity, score, inspection effectivity, source, and formula note.

Pipeline also stores inspection scope, interval, and method per damage mechanism in JSON:

- Non intrusive method, effectivity, and interval.
- Intrusive method, effectivity, and interval.

The severities follow the same display language users expect from Pressure Vessel screening: `NOT`, `Low`, `Moderate`, and `High`. The formulas are Pipeline-specific system screening formulas and are marked `TODO_ENGINEERING_CONFIRMATION` where engineering source confirmation is still needed.

No Pressure Vessel damage mechanism formulas were copied.

## Pipeline vs Pressure Vessel

Pressure Vessel has its own production damage mechanism flow and tables. Pipeline now has a separate configuration and validation path. Pipeline damage mechanism currently classifies the assessment and appears in formula trace metadata; it does not yet alter PoF until engineering confirms the mapping.

## API Alignment Notes

API 570 supports inspection planning concepts for in-service piping systems, including deterioration mechanisms, but this implementation does not claim an API 570 formula.

RBI 571 and RBI 581 references are limited to alignment notes and TODO confirmation. The current Pipeline risk model remains the existing MVP/index-based model.

## Files Changed

- `models/pipeline_oil.go`
- `models/pipeline_damage_mechanism.go`
- `models/pipeline_oil_test.go`
- `controller/pipelineController.go`
- `templates/pipeline_assessment_form.html`
- `templates/pipeline_assessment_detail.html`
- `assets/js/pipeline_oil_assessment.js`
- `docs/pipeline-review-notes.md`

## Routes / Pages Affected

- `GET /assessment-pipeline/form`
- `GET /assessment-pipeline/gas`
- `GET /assessment-pipeline/edit/:id`
- `GET /assessment-pipeline/view/:id`
- `POST /assessment-pipeline/preview`
- `POST /assessment-pipeline/calculate/:id`

## DB Changes

No DB migration was required. Pipeline records are stored as JSON in `pipeline_oil_assessments`, so the new damage mechanism result list, per-mechanism inspection effectivity map, inspection scope/interval/method plan, and advisory source fields are persisted in existing JSON columns.

## Tests Added

- Validation test confirming `Inspection Result` is not required as input.
- Calculation test confirming `Inspection Result` is produced as output.
- Damage mechanism selection/legacy normalization validation test.
- Test confirming all Pipeline damage mechanisms are calculated.
- Test confirming per-mechanism inspection effectivity is persisted in result data.
- Recommendation source/rule metadata test.

## Prepared Answers

Why was Inspection Result in Step 1?
It came from a report/output-style sample, not from a confirmed user input requirement. It has been removed from Step 1.

Where does Recommendation come from?
It comes from a Pipeline system advisory rule in code, based on risk level, CoF, and the governing damage-factor driver. It is not presented as an official RBI/API recommendation.

Why was Damage Mechanism missing?
The new Pipeline module initially used only simplified damage-factor drivers. A separate Pipeline damage mechanism metadata selector has now been added.

Is this directly from RBI/API or system advisory?
The recommendation is system advisory. The damage mechanism list is aligned to common API 570-style inspection concepts, but calculation linkage remains `TODO_ENGINEERING_CONFIRMATION`.
