# Pipeline Oil and Gas RBI Engineering Basis

Stage: Phase 2, Stage 2 - Engineering Calibration and Standards Alignment

Scope: Pipeline Oil and Gas RBI only. Pressure Vessel RBI is not part of this basis.

## Calculation Lock

The active scoring model remains Gate -> Modifier -> Escalation. This document does not change calculation behavior. It records the engineering source basis, confidence level, and remaining confirmation items for the current implementation.

Rule status taxonomy:

- VERIFIED: Direct unit conversion, workbook aggregation, or code formula whose implementation is directly traceable.
- PARTIALLY_VERIFIED: Standard concept is traceable, but licensed tables, detailed factors, or full standard procedures are not embedded.
- TODO_ENGINEERING_CONFIRMATION: Current value or threshold is retained pending engineering confirmation.

## Standards Basis

- API 570: in-service inspection, rating, repair, and alteration code for piping systems. Used as the inspection and remaining life practice basis.
- API RP 571: damage mechanism identification and examination guidance. Used as the basis for mechanism taxonomy and screening traceability.
- API RP 581: RBI public methodology concepts for probability of failure, consequence of failure, risk drivers, RBI planning, and reassessment. Licensed detailed tables are not embedded.
- ASME B31.3, ASME B31.4, and ASME B31.8: required thickness, pressure design, hoop stress, and MAOP formula basis according to selected applicable code.
- NACE MR0175 / ISO 15156: sour service cracking threshold concept for H2S partial pressure and SSC/HIC screening.
- AMPP SP0169: external corrosion control basis for underground or submerged metallic piping systems.

Reference URLs:

- API 570: https://www.api.org/products-and-services/standards/important-standards-announcements/570
- API RP 571: https://www.api.org/products-and-services/standards/important-standards-announcements/571
- API RP 581 training summary: https://www.api.org/products-and-services/training/api-u/rasci/581
- AMPP SP0169: https://store.ampp.org/sp0169-2024
- ISO 15156 / NACE MR0175 overview: https://www.iso.org/standard/72485.html

## Current Formula Basis

Partial pressure:

- CO2 partial pressure: `pCO2 = (CO2 mole percent / 100) * operating pressure psig`
- H2S partial pressure: `pH2S = (H2S mole percent / 100) * operating pressure psig`

Mechanical appraisal:

- Required thickness uses ASME B31.3, B31.4, or B31.8 branch logic based on `applicable_code`.
- Corrosion rate: `(nominal thickness mm - actual thickness mm) / (measured year - year used)`
- Remaining life: `(actual thickness mm - required thickness mm) / corrosion rate mm/year`, capped at the configured maximum display life.
- Hoop stress: `(P * D) / (2 * actual thickness in)`
- MAOP uses the selected ASME B31 branch formula.
- SMYS utilization: `(P * OD) / (2 * actual thickness in * SMYS) * 100`
- Wall thickness ratio: `minimum actual thickness / minimum required thickness`

Risk model:

- Management system factor: `FMS = 10^((-0.02 * ((management score / 1000) * 100)) + 1)`
- Pipeline PoF: `PoF = generic failure frequency * governing damage mechanism score * FMS`
- Risk ranking: `PoF category * CoF category`
- Gas consequence uses Potential Impact Radius and building count screening.
- Liquid consequence uses spill volume and environmental adjustment screening.

## Damage Mechanism Basis

Damage mechanism rows now include:

- `source_standard`
- `confidence_level`
- `rule_status`

The current mechanism taxonomy is aligned to pipeline damage families:

- External Corrosion
- Coating Degradation
- Third-Party / Mechanical Damage
- Internal Corrosion
- Localized Corrosion
- Erosion
- Erosion-Corrosion
- Cracking
- SCC
- Fatigue
- Chemical Damage

API 571 is used as the main damage mechanism reference. API 581 concepts are used for RBI PoF traceability. NACE MR0175 / ISO 15156 is used for sour cracking traceability. AMPP SP0169 is used for external corrosion control traceability.

## Remaining TODO Engineering Confirmation

The following areas intentionally retain current behavior and are marked as TODO where source detail is not verified:

- Neutral placeholder factor maps for depth, patrol, ROW, soil, coating condition, CP, class location, pH, inhibitor, coating damage, insulation damage, confidence weighting, weld cracking, PWHT, one-call, H2S ppm, flow velocity, solids, external cracking environment, and fatigue cycle thresholds.
- Pipeline-specific inspection interval multipliers.
- Detailed API 581 damage-factor table values beyond the current approved Pipeline screening inputs.
- Liquid spill consequence factors and environmental adjustment.
- Gas PIR consequence category thresholds and building count category thresholds.
- Coating degradation, localized corrosion, erosion, erosion-corrosion, SCC, fatigue, chemical damage, and third-party damage modifiers.

## Pressure Vessel Boundary

No Pressure Vessel calculation, template, or JavaScript file is included in this Stage 2 basis. Pressure Vessel RBI remains locked.
