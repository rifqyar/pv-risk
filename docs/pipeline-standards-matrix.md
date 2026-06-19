# Pipeline Standards Matrix

Scope: Pipeline Oil and Gas RBI only. No Pressure Vessel scope.

Status legend:

- VERIFIED: Direct code/workbook/conversion trace.
- PARTIALLY_VERIFIED: Standard concept is aligned, detailed licensed tables or factor calibration not embedded.
- TODO_ENGINEERING_CONFIRMATION: Current implementation is retained pending engineering confirmation.

## Formula Matrix

| Field / Output | Formula or Rule | Damage Mechanism | Threshold | Source | Status |
|---|---|---|---|---|---|
| Pipe length ft | `((m * 100) / 2.54) / 12` | N/A | N/A | Engineering unit conversion | VERIFIED |
| Temperature C | `(5 / 9) * (F - 32)` | N/A | N/A | Engineering unit conversion | VERIFIED |
| OD mm | `OD in * 25.4` | N/A | N/A | Engineering unit conversion | VERIFIED |
| Allowance mm | `allowance in * 25.4` | N/A | N/A | Engineering unit conversion | VERIFIED |
| Required thickness | ASME branch formula by applicable code | Mechanical appraisal | Per selected ASME code | ASME B31.3 / B31.4 / B31.8 | VERIFIED |
| Corrosion rate | `(nominal - actual) / elapsed years` | Internal / external thinning support | Elapsed years must be positive | API 570 inspection practice / workbook formula | PARTIALLY_VERIFIED |
| Remaining life | `(actual - required) / corrosion rate` | Internal / external thinning support | Capped display life | API 570 inspection practice / workbook formula | PARTIALLY_VERIFIED |
| Hoop stress | `(P * D) / (2 * t)` | Mechanical appraisal | Compared through status checks | ASME B31 pressure design concept | VERIFIED |
| MAOP | ASME branch formula by applicable code | Mechanical appraisal | Must be acceptable versus operation | ASME B31.3 / B31.4 / B31.8 | VERIFIED |
| SMYS utilization | `(P * OD) / (2 * actual t * SMYS) * 100` | SCC / stress support | 30, 50, 72 percent screening bands | ASME B31 stress utilization concept | PARTIALLY_VERIFIED |
| Wall thickness ratio | `min(actual) / min(required)` | Internal / localized corrosion support | 1.0 acceptance reference | API 570 inspection practice | PARTIALLY_VERIFIED |
| CO2 partial pressure | `(CO2 mole percent / 100) * operating pressure` | Internal corrosion | Current pCO2 bands in code | API 581 / API 571 concept | PARTIALLY_VERIFIED |
| H2S partial pressure | `(H2S mole percent / 100) * operating pressure` | Cracking | 0.05 psia sour screening concept | NACE MR0175 / ISO 15156 | PARTIALLY_VERIFIED |
| Management system factor | `10^((-0.02 * ((score / 1000) * 100)) + 1)` | PoF support | Management score 0 to 1000 | API 581 public methodology | PARTIALLY_VERIFIED |
| Pipeline PoF | `GFF * governing DM score * FMS` | All governing mechanisms | PoF category lookup in code | API 581 concept adapted to pipeline DM score | PARTIALLY_VERIFIED |
| Risk ranking | `PoF category * CoF category` | All governing mechanisms | MVP matrix in code | Pipeline MVP matrix pending API 581 matrix confirmation | TODO_ENGINEERING_CONFIRMATION |
| Gas PIR | Existing PIR formula in implementation | Consequence | Building count categories in code | Pipeline consequence screening practice | TODO_ENGINEERING_CONFIRMATION |
| Liquid spill volume | Flow and detection/isolation inputs | Consequence | Environmental adjustment in code | Pipeline consequence screening practice | TODO_ENGINEERING_CONFIRMATION |
| Inspection intervals | Severity and effectivity multiplier | All mechanisms | 6 to 120 months clamp | API 570 inspection planning concept / engineer-defined table | TODO_ENGINEERING_CONFIRMATION |

## Damage Mechanism Matrix

| Damage Mechanism | Gate / Driver Basis | Modifier / Escalation Basis | Source | Status |
|---|---|---|---|---|
| External Corrosion | CP concern, damaged coating, low soil resistivity | Base external rate, soil, coating, CP, prior finding | API 571 / AMPP SP0169 | PARTIALLY_VERIFIED |
| Coating Degradation | Damaged coating, CP concern, damaged insulation | Coating damage, soil, CP potential, CUI temperature, prior finding | API 571 / AMPP SP0169 | TODO_ENGINEERING_CONFIRMATION |
| Third-Party / Mechanical Damage | Base TPD rate | Depth, patrol, ROW, one-call placeholders | API 570 / pipeline integrity management practice | TODO_ENGINEERING_CONFIRMATION |
| Internal Corrosion | CO2, water, or corrosive fluid condition | pCO2 severity, pH, inhibitor, biocide, prior thinning | API 581 / API 571 | PARTIALLY_VERIFIED |
| Localized Corrosion | Internal corrosion, chlorides, low pH, previous localized corrosion | Chloride, pH, prior localized finding | API 571 | TODO_ENGINEERING_CONFIRMATION |
| Erosion | Velocity or solids concern | Solids and corrosivity modifiers | API 571 / DNV-RP-O501 concept | TODO_ENGINEERING_CONFIRMATION |
| Erosion-Corrosion | Erosion present | Corrosivity modifier | API 571 | TODO_ENGINEERING_CONFIRMATION |
| Cracking | H2S partial pressure, H2S content, or prior cracking | pH2S severity, PWHT, weld type, prior cracking | NACE MR0175 / ISO 15156 / API 571 | PARTIALLY_VERIFIED |
| SCC | Stress utilization plus coating/CP/H2S concern | Stress band, coating, H2S placeholder | API 571 / NACE MR0175 / ISO 15156 | TODO_ENGINEERING_CONFIRMATION |
| Fatigue | Pressure cycles or prior cracking | Pressure range placeholder, weld type, prior cracking | API 571 | TODO_ENGINEERING_CONFIRMATION |
| Chemical Damage | Placeholder only | No active modifier | Engineering review stub | TODO_ENGINEERING_CONFIRMATION |

## Threshold Classification

| Threshold | Current Use | Source | Status |
|---|---|---|---|
| H2S partial pressure 0.05 psia | Sour cracking gate | NACE MR0175 / ISO 15156 concept | PARTIALLY_VERIFIED |
| pH2S bands 0.05, 0.5, 15 psia | Cracking severity bands | Internal calibration pending | TODO_ENGINEERING_CONFIRMATION |
| CO2 pCO2 bands 5 and 20 psig | Internal corrosion severity bands | API 581 / API 571 concept, exact bands pending | TODO_ENGINEERING_CONFIRMATION |
| CP potential greater than -850 mV | External/coating concern marker | AMPP SP0169 concept | PARTIALLY_VERIFIED |
| SMYS utilization 30, 50, 72 percent | SCC stress screening | ASME B31 stress concept, SCC bands pending | TODO_ENGINEERING_CONFIRMATION |
| Severity score less than 1.5 / less than 3.0 / 3.0 or greater | NOT/Low/Moderate/High mapping | Pipeline MVP calibration | TODO_ENGINEERING_CONFIRMATION |
| PoF category thresholds | PoF category mapping | API 581 concept adapted to MVP | TODO_ENGINEERING_CONFIRMATION |
| CoF category thresholds | CoF category mapping | Pipeline MVP consequence screening | TODO_ENGINEERING_CONFIRMATION |
| Inspection interval 6 to 120 months | Clamp for generated inspection plan | Engineer-defined application limit | TODO_ENGINEERING_CONFIRMATION |
