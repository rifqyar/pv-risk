# Pipeline Report Template

Scope: Pipeline Oil and Gas RBI only.

## Report Sections

1. Executive Summary
   - Tag number
   - Pipeline name
   - Location
   - Fluid type
   - Assessment status
   - Final risk code
   - Final risk level
   - Governing damage mechanism
   - Inspection result

2. Pipeline Information
   - Owner/user
   - Contractor
   - place/date issued
   - year built/year used
   - pipe size
   - material specification
   - installation type
   - applicable code

3. Design Data
   - outside diameter
   - nominal wall thickness
   - actual wall thickness
   - design pressure
   - operating pressure
   - design temperature
   - SMYS
   - material stress
   - design factor
   - quality factor
   - weld joint strength factor

4. Process and Condition
   - coating type
   - corrosion control
   - CP status
   - CP potential
   - soil resistivity
   - pH
   - CO2 content
   - H2S content
   - H2O content
   - corrosion monitoring
   - previous findings

5. Damage Mechanism Results
   - category
   - damage mechanism
   - score
   - severity
   - triggered by
   - formula
   - applicable standards
   - confidence level
   - rule status

6. PoF Analysis
   - generic failure frequency
   - management system score
   - management system factor
   - governing damage mechanism score
   - PoF value
   - PoF category

7. CoF Analysis
   - gas PIR or liquid spill basis
   - CoF value
   - CoF category
   - environmental/sensitive receptor factors where applicable

8. Risk Matrix
   - PoF category
   - CoF category
   - final risk code
   - final risk level

9. Recommendations
   - recommendation text
   - immediate actions
   - inspection/monitoring actions
   - long-term mitigation actions
   - recommendation source
   - rule name
   - confidence level

10. Formula Trace
    - formula name
    - reference
    - expression
    - input values
    - output value
    - source standard
    - confidence level
    - rule status

11. Standards References
    - API 570
    - API RP 571
    - API RP 581 public methodology
    - ASME B31.3
    - ASME B31.4
    - ASME B31.8
    - NACE MR0175 / ISO 15156
    - AMPP SP0169
    - Pipeline DM screening v2

12. Audit Information
    - version number
    - created by
    - created at
    - updated by
    - updated at
    - audit action
    - audit actor
    - audit timestamp
    - affected fields

## Formula Trace Example

| Formula | Inputs | Expression | Output | Source | Confidence | Status |
|---|---|---|---|---|---|---|
| `pipeline_pof` | `GFF=0.00003`, `governing_dm_score=1`, `FMS=1` | `PoF = GFF * governing DM score * FMS` | `0.00003` | API 581 public methodology adapted to pipeline DM score | Medium | PARTIALLY_VERIFIED |
| `h2s_partial_pressure` | `H2S mole%=0.01`, `operating_pressure=650 psig` | `pH2S = (H2S mole% / 100) * operating pressure` | `0.065 psig` | NACE MR0175 / ISO 15156 concept | Medium | PARTIALLY_VERIFIED |
| `wall_thickness_ratio` | `minimum actual`, `minimum required` | `min(actual thickness) / min(required thickness)` | `1.42` | API 570 inspection practice | Medium | PARTIALLY_VERIFIED |

## Damage Mechanism Explainability Format

Example:

Damage Mechanism: Internal Corrosion

- Score: `2.0`
- Severity: `Moderate`
- Triggered by:
  - `H2O Content = 15%`
  - `pCO2 = 22 psi`
  - `Corrosion Monitoring = Poor`
- Applicable standards:
  - `API 581`
  - `API 571`
  - `Pipeline DM screening v2`
- Confidence level: `Medium`
- Rule status: `PARTIALLY_VERIFIED`
- Engineering notes: display trigger reasons and any `TODO_ENGINEERING_CONFIRMATION` markers.

## Standards Reference Format

Use concise source attribution:

`Source Standard / Confidence Level / Rule Status`

Examples:

- `API 570 inspection practice / Medium / PARTIALLY_VERIFIED`
- `ASME B31.4 / High / VERIFIED`
- `Pipeline MVP matrix pending licensed API 581 matrix confirmation / Low / TODO_ENGINEERING_CONFIRMATION`

Do not claim API/RBI origin for a recommendation unless the rule has a verified source. System-generated recommendations must remain identified as system advisory rules.
