# Pipeline Reporting Architecture

Scope: Pipeline Oil and Gas RBI only. Pressure Vessel RBI is locked and outside this architecture.

## Objective

Phase 3 adds reporting, explainability, and audit trail support without changing Pipeline scoring behavior. The calculation model remains Gate -> Modifier -> Escalation, and PoF remains `GFF * GoverningDMScore * FMS`.

## Snapshot Design

The current `pipeline_oil_assessments` table remains the latest working record. It still stores:

- `input_json`
- `result_json`
- `formula_trace_json`
- `snapshot_json`

Phase 3 adds append-only immutable snapshots in `pipeline_oil_assessment_versions`.

Each version stores:

- assessment id
- version number
- status
- formula version
- input JSON
- result JSON
- formula trace JSON
- snapshot JSON
- created by
- created at

The snapshot JSON includes:

- formula version
- all input values
- calculated result
- standards references
- recommendation source
- recommendation rule name
- recommendation confidence
- snapshot timestamp

This preserves historical decisions when future rules, thresholds, or UI fields change.

## Audit Trail Design

Audit events are stored in `pipeline_oil_audit_events`.

Tracked actions:

- `CREATED`
- `MODIFIED`
- `RECALCULATED`
- `APPROVED`
- `EXPORTED_PDF`
- `EXPORTED_EXCEL`
- `ARCHIVED`

Each event stores:

- assessment id
- optional version id
- action
- actor
- affected fields JSON
- old values JSON
- new values JSON
- note
- created at

PDF export remains client-side, but an audit endpoint records the export event after generation. Excel-compatible export is server-side and records the export before returning the file.

## Versioning Strategy

A new immutable version is created on:

- assessment creation
- draft update
- calculation/recalculation

Version numbers increment per assessment. The latest version is not edited after creation.

Approval and export actions are audit events only. They do not create new calculation versions and do not change risk results.

## Comparison Strategy

The comparison view compares the latest two immutable versions.

The comparison engine flattens input and result JSON into field paths, then reports:

- changed field
- old value
- new value

This supports changed inputs, changed scores, and changed risk category without duplicating calculation logic.

## Export Strategy

PDF:

- uses the existing client-side `html2pdf` workflow
- includes executive summary, pipeline information, design data, condition/process inputs, damage mechanism results, PoF/CoF, risk matrix, recommendations, formula trace, standards references, and audit information
- records `EXPORTED_PDF` through a backend audit endpoint after generation

Excel:

- uses a server-generated Excel-compatible `.xls` HTML workbook
- avoids adding a new binary spreadsheet dependency
- includes key report sections, damage mechanism results, formula trace, standards references, and audit information
- records `EXPORTED_EXCEL`

## Explainability Model

The implementation reuses existing structures:

- `PipelineOilFormulaTrace`
- `PipelineDamageMechanismResult`
- `PipelineTriggerInput`

Formula trace entries include:

- formula name
- reference
- expression
- input values
- output value
- source standard
- confidence level
- rule status

Damage mechanism explainability includes:

- score
- severity
- trigger inputs
- formula text
- source standard
- confidence level
- rule status

Recommendation transparency includes:

- recommendation text
- source
- rule name
- confidence level

## Remaining TODO Engineering Confirmation

The Phase 3 reporting layer preserves and displays existing `TODO_ENGINEERING_CONFIRMATION` markers. It does not resolve engineering calibration items.

Outstanding items remain those documented in:

- `docs/pipeline-engineering-basis.md`
- `docs/pipeline-standards-matrix.md`
