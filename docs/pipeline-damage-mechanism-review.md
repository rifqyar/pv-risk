# Pipeline Damage Mechanism Model — v2 Review

## Summary of Changes

The Pipeline Damage Mechanism (DM) model has been redesigned from a qualitative multiplicative factor model to a **Gate-Modifier-Escalation** scoring model with traceable numeric inputs.

### Removed Fields (6)
| Old Field | Reason |
|-----------|--------|
| `water_content` | Replaced by numeric `h2o_content` (mol%) |
| `fluid_corrosivity` | Replaced by `fluid_corrosivity_mpy` (NACE RP0775 categories) |
| `co2_h2s_presence` | Replaced by numeric `co2_content`, `h2s_content` with calculated partial pressures |
| `mic_risk` | Replaced by `biocide_treatment` and `corrosion_monitoring_result` |
| `wall_thickness_condition` | Replaced by calculated `wall_thickness_ratio` from inspection data |
| `emergency_response` | Replaced by `one_call_system` for TPD screening |

### Added Fields (28)
| Section | Fields |
|---------|--------|
| **A — Component Composition** | `co2_content`, `h2s_content`, `h2o_content`, `n2_content`, `co_content`, `co2_partial_pressure_psig` (auto-calc), `h2s_partial_pressure_psig` (auto-calc), `h2s_ppm` |
| **B — Corrosion Indicators** | `ph_level`, `chloride_content`, `fluid_corrosivity_mpy`, `inhibitor_effectiveness`, `biocide_treatment`, `corrosion_monitoring_result`, `wall_thickness_ratio` (auto-calc) |
| **C — Operating Condition** | `fluida`, `phase`, `pwht_status`, `weld_joint_type`, `pressure_cycle_count`, `pressure_range_pct`, `smys_utilization_pct` (auto-calc), `flow_velocity_condition`, `solid_content` |
| **D — Previous Equipment Condition** | `prev_ext_corrosion`, `conf_ext_corrosion`, `prev_int_thinning`, `conf_int_thinning`, `prev_int_cracking`, `conf_int_cracking`, `prev_loc_int_corrosion`, `conf_loc_int_corrosion`, `insulation_condition`, `insulation_damage_level`, `ext_coating_condition`, `ext_coating_damage_level` |
| **E — Cracking Indicators** | `env_ext_cracking` |
| **G — Consequence** | `one_call_system` |

### Mechanism List Changes
| Old Code | New Code | Notes |
|----------|----------|-------|
| `coating_cui_degradation` | `coating_degradation` | Removed CUI (pipeline scope) |
| `localized_corrosion_pitting` | `localized_corrosion` | Shortened |
| `cracking_damage` | `cracking` | Generalized |
| `cracking_scc_fatigue` | `scc` | Split from fatigue |
| `other_engineering_review` | *(removed)* | Stub removed |
| *(new)* | `erosion` | New |
| *(new)* | `fatigue` | New |
| `chemical_damage` | `chemical_damage` | Retained as stub |

### Scoring Model: Gate-Modifier-Escalation

Each mechanism scoring function follows this pattern:

1. **Gate checks** — Boolean conditions that determine if the mechanism is applicable. If all gates fail, severity = NOT.
2. **Base severity** — Determined from sourced thresholds (API 581, NACE MR0175, API 570, etc.).
3. **Modifiers** — Additive adjustments from operating conditions (all currently `0.0` with `TODO_ENGINEERING_CONFIRMATION`).
4. **Previous finding escalation** — If a previous finding exists, severity is escalated.

The **governing DM score** replaces the old multiplicative DF model for PoF calculation:

```
PoF = GFF × governing_DM_score × FMS
```

Where `governing_DM_score` is the maximum score across all screened mechanisms.

## Sourced Thresholds

| Threshold | Source | Values |
|-----------|--------|--------|
| CO2 partial pressure | API 581 Section 6 | Low ≤ 5 psig, Moderate ≤ 20 psig, High > 20 psig |
| H2S partial pressure | NACE MR0175 | Not ≤ 0.05 psig, Low ≤ 0.5 psig, Moderate ≤ 15 psig, High > 15 psig |
| CP potential | NACE SP0169 | -850 mV criterion |
| MIC temperature range | PV implementation | 10–93°C |

## TODO_ENGINEERING_CONFIRMATION Items

All modifier values are set to `1.0` (neutral) / `0.0` (no adjustment) pending engineering review. These include:
- All `pipeline*Factors` and `pipeline*Modifiers` maps
- Modifier magnitudes in scoring functions
- Previous finding escalation magnitudes
- Base severity screening thresholds for mechanisms without sourced data

The `PipelineOilFormulaVersion` is set to `"pipeline-oil-risk-v2"` to distinguish from the prior multiplicative model.

## Backward Compatibility

- Deprecated fields (`water_content`, `fluid_corrosivity`, `co2_h2s_presence`, `mic_risk`, `wall_thickness_condition`, `emergency_response`) are preserved in the `PipelineOilRiskInput` struct with `omitempty` JSON tags
- Old factor maps (`pipelineConditionFactors`, `pipelineInternalFactors`) are retained for `NormalizePipelineDamageMechanism` mapping
- Old DF calculation functions (`calculateThirdPartyDamageFactor`, `calculateExternalCorrosionFactor`, `calculateInternalCorrosionFactor`) are still called in `applyPipelineIndexRisk` for the `ThirdPartyDamageFactor`, `ExternalCorrosionFactor`, `InternalCorrosionFactor` result fields, but they use neutral values from the new maps
- The `CalculatePipelineOil` result struct remains unchanged; new DM scoring populates `DamageMechanismResults` with `Score` and `TriggerInputs` fields

## Files Modified

| File | Changes |
|------|---------|
| `models/pipeline_damage_mechanism.go` | Rewritten mechanism list, `PipelineTriggerInput` struct, legacy normalization |
| `models/pipeline_oil.go` | New fields, factor maps, scoring functions, defaults, validation, PoF order fix |
| `models/pipeline_oil_test.go` | Updated assertions for neutral-factor model, fixed mechanism codes |
| `controller/pipelineController.go` | Updated defaults to new field names/values |
| `templates/pipeline_assessment_form.html` | New Sections A-E fields, replaced deprecated dropdowns, updated coating_condition options, class_location options |
| `assets/js/pipeline_oil_assessment.js` | Updated `collectPayload()`, `collectPayloadShallowRiskInput()`, DM codes, factor maps, recommendation driver label |
| `templates/pipeline_assessment_detail.html` | Updated Section III to show new fields, added Score column to DM table |