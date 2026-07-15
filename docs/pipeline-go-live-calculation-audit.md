# Pipeline Go-Live Calculation Audit

## Scope

This audit covers the Pipeline Oil/Gas assessment flow only. Pressure Vessel calculation, templates, routes, schema, and workflow were not modified.

## Root Causes

### Stuck Medium / Inspection Method Flow

The method dropdown was already included in the general form change handler, but the realtime damage-mechanism badge used `damage_mechanism_results[].inspection_effectivity` while the realtime inspection plan calculated method effectivity separately. That allowed the displayed DM/effectivity badge to remain stale or fall back to `Medium` even after the selected method changed.

Backend calculation had the same split behavior for legacy/sparse payloads: damage-mechanism results preferred `InspectionEffectivityByDM`, while inspection-plan results independently resolved the selected method. If the per-DM effectivity map was absent, saved detail/PDF could diverge from the selected method.

### Required WT = 0 in PDF

The backend calculates authoritative required wall thickness in `PipelineOilResult.PointResults[].RequiredThicknessMM`. The PDF design inspection-point table rendered `Input.InspectionPoints[].RequiredThicknessMM` instead. The form payload does not populate point-level `required_thickness_mm`, so the raw saved input field could remain zero and the PDF displayed `0 mm`.

## Remediation

### Inspection Method / Effectivity

The backend now resolves damage-mechanism effectivity through one helper:

- per-DM `InspectionEffectivityByDM`, when provided
- selected non-intrusive inspection method from `InspectionPlanByDM`
- global inspection effectivity fallback
- `Medium` only as a final compatibility fallback

The frontend now annotates each realtime damage-mechanism result with method-derived effectivity from the Pipeline Inspection Method Master option metadata. Realtime DM badges, plan effectivity, and next inspection intervals are recalculated from the selected method without recursive event triggering.

Inspection effectiveness is not used to alter Pipeline DM severity scores. Inspection results may update factual PoF inputs such as measured condition, corrosion rate, defect dimensions, inspection confidence, or detection confidence, but the inspection-effectiveness selection itself does not increase or decrease Pipeline DM severity.

Confirmed configuration:

- `inspection_effectiveness_changes_dm_severity = false`

### Required Wall Thickness / PDF

`PipelineOilPointResult` now carries:

- inspection point
- location class
- installation type
- measured year
- authoritative required thickness

The PDF inspection-point table now renders `Assessment.Result.PointResults` and shows `N/A` when required thickness is unavailable instead of showing misleading `0 mm`.

### Empty Inspection Point Rows

Existing `validPipelineInspectionPoints` behavior skips fully empty rows before calculation. Tests now cover that empty rows do not become result rows or governing thickness rows.

## Pressure Vessel Pattern Reused

Pressure Vessel was used as a workflow reference only:

- dropdown change triggers recalculation
- selected method resolves effectivity
- effectivity updates adjacent UI text
- calculation output is rendered from calculated result data rather than raw form defaults

No Pressure Vessel damage formulas were copied into Pipeline.

## Pipeline Formulas Retained

The existing Pipeline formulas were retained with the confirmed shared threshold and material-source decisions:

- ASME B31.3 required thickness using material stress `S`
- ASME B31.4/B31.8 required thickness using SMYS/design factors
- corrosion rate from nominal, actual, year used, and measured year
- remaining life from actual, required thickness, and corrosion rate
- Pipeline DM gate/modifier/escalation framework
- Pipeline PoF = GFF x governing DM score x FMS
- Pipeline CoF gas/liquid screening
- Pipeline DM modifier default fixed at `1.0`; custom Pipeline DM modifiers disabled
- Pipeline risk and CoF categorization reuse the approved pressure-vessel risk matrix/classification helper
- Pipeline B31.3 allowable stress is sourced only from Pipeline Material Master data

## Tests Added

- selected inspection method changes DM effectivity
- selected inspection method changes inspection interval
- required WT is calculated into authoritative point results when raw input required WT is zero
- empty inspection point rows are skipped
- point metadata required by PDF is preserved in result rows
- inspection effectiveness changes do not change Pipeline DM severity
- Pipeline DM modifier validation requires all active modifier values to remain `1.0`
- custom Pipeline DM modifier configuration is disabled for production calculations
- Pipeline and pressure-vessel modules use the same risk category boundaries for the same matrix score
- Pipeline and pressure-vessel modules use the same CoF category thresholds for the same CoF metric
- Pipeline calculations use the pipeline-specific material-stress dataset metadata
- Pipeline calculations do not fall back to pressure-vessel material-stress data
- unsupported Pipeline B31.3 material allowable-stress mappings return engineering review required
- shared threshold boundary inclusivity is tested
- the five resolved engineering-confirmation phrases are absent from this go-live audit

## Go-Live Limitations

Status: READY.

The affected UI/backend/PDF mapping defects are remediated. The five previously blocked engineering decisions in this audit have been resolved as confirmed implementation requirements.

## Confirmed Sources

- Shared CoF and risk thresholds: `models/risk_classification.go`, using the approved pressure-vessel matrix/classification configuration (`pressure-vessel-risk-matrix-v1`, version `pressure-vessel-approved-v1`, effective `2025-07-01`).
- Pipeline material allowable stress: `seeder/pipeline_material.go`, dataset source `pipeline_specific_dataset`, version `pipeline-b31-material-stress-v1`.
- Pressure-vessel material data remains separate and is not used as a Pipeline fallback.
