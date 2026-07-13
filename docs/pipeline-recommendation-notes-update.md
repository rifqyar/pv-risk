# Pipeline Recommendation and Engineering Notes Update

## Manual Recommendation Fields

Pipeline Step 6 now stores recommendation content in three editable multiline fields:

- `recommendation_immediate_actions`
- `recommendation_inspection_monitoring`
- `recommendation_long_term_mitigation`

These fields are part of `PipelineOilInput`, are saved in the existing JSON payload flow, and are used to populate `PipelineOilResult.RecommendationGroups` during calculation. Recalculation refreshes risk and metadata, but it does not replace text typed into these fields.

## Metadata Behavior

The Step 6 metadata panel remains visible beside the manual recommendation text:

- Source
- Confidence
- Advisory

Recommendation text is engineer-entered when the new fields are populated. Source describes the basis of the recommendation content, Confidence remains the existing advisory confidence value, and Advisory displays the advisory rule/classification name.

## Legacy Manual Recommendation

The old `manual_recommendation` field has been removed from the active Pipeline form and frontend workflow. The Go struct still accepts it so old saved JSON can deserialize.

If an old assessment has `manual_recommendation` and the three new recommendation fields are empty, the legacy value is exposed as a single legacy recommendation note for display/export. The old value is not copied into all three new fields and no database column or JSON data is destructively removed.

## Engineering Notes PDF Fix

Engineering Notes were already captured in `RiskInput.engineering_notes`, but the export surfaces were inconsistent. The Pipeline PDF template now renders an `Engineering Notes` section only when the value is non-empty, preserving multiline formatting and using normal Go template escaping.

## Files Modified

- `models/pipeline_oil.go`
- `models/pipeline_oil_test.go`
- `controller/pipelineController.go`
- `controller/pipeline_export_test.go`
- `templates/pipeline_assessment_form.html`
- `templates/pipeline_assessment_detail.html`
- `assets/js/pipeline_oil_assessment.js`
- `template_parse_test.go`

Pressure Vessel files were not modified.
