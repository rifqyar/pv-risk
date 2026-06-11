$(function () {
  const $form = $("#pipelineOilForm");
  let pipelineStepper = null;
  let currentResult = null;
  let resultSignature = "";
  let resultIsStale = false;

  const stepperEl = document.querySelector("#wizard-pipeline-assessment");
  if (stepperEl && typeof Stepper !== "undefined") {
    pipelineStepper = new Stepper(stepperEl, { linear: false });
  }

  updateConsequenceFields();
  updateReviewSummary();
  loadSavedResult();
  updateSaveState();

  $(".pipeline-step-next").on("click", function () {
    const $activeStep = $(".bs-stepper-content .content.active");
    if (!validateCurrentStep()) return;
    updateReviewSummary();
    if (pipelineStepper) pipelineStepper.next();
    if ($activeStep.attr("id") === "pipeline-step-3") {
      previewPipelineRisk(false);
    }
  });

  $(".pipeline-step-prev").on("click", function () {
    if (pipelineStepper) pipelineStepper.previous();
  });

  $form.on("input change", "input, select, textarea", function () {
    clearFieldError($(this));
    updateConsequenceFields();
    updateReviewSummary();
    markResultStaleIfNeeded();
  });

  $("#addPipelinePoint").on("click", function () {
    $("#pipelinePointsTable tbody").append(`
      <tr>
        <td><input class="form-control" name="inspection_point"></td>
        <td><input class="form-control" name="location_class"></td>
        <td><input class="form-control" name="installation_type"></td>
        <td><input type="number" step="any" class="form-control" name="point_nominal_thickness_mm"></td>
        <td><input type="number" step="any" class="form-control" name="point_actual_thickness_mm"></td>
        <td><input type="number" class="form-control" name="measured_year" value="${new Date().getFullYear()}"></td>
        <td><button type="button" class="btn btn-sm btn-icon btn-outline-danger remove-point"><i class="mdi mdi-trash-can-outline"></i></button></td>
      </tr>
    `);
    markResultStaleIfNeeded();
  });

  $(document).on("click", ".remove-point", function () {
    $(this).closest("tr").remove();
    markResultStaleIfNeeded();
  });

  $("#calculatePipelineOil").on("click", function () {
    previewPipelineRisk(true);
  });

  $("#savePipelineDraft").on("click", function () {
    saveReviewedAssessment();
  });

  function loadSavedResult() {
    const raw = $("#pipelineSavedResult").text().trim();
    if (!raw || raw === "null") return;
    try {
      currentResult = JSON.parse(raw);
      resultSignature = calculationSignature(collectPayload());
      resultIsStale = false;
      renderCalculationResult(currentResult);
    } catch (err) {
      currentResult = null;
    }
  }

  function previewPipelineRisk(showSuccess) {
    if (!validatePipelineForm(true)) {
      Swal.fire("Complete the required fields before calculating risk.", "A few required inputs are missing or invalid.", "warning");
      return;
    }

    const payload = collectPayload();
    Swal.fire({ title: "Calculating...", allowOutsideClick: false, showConfirmButton: false });
    Swal.showLoading();

    fetch("/assessment-pipeline/preview", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    })
      .then((r) => r.json())
      .then((res) => {
        if (res.status !== "success") {
          Swal.fire("Check the form", humanizeServerError(res.message), "error");
          return;
        }
        currentResult = res.result;
        resultSignature = calculationSignature(payload);
        resultIsStale = false;
        renderCalculationResult(currentResult);
        updateSaveState();
        if (showSuccess) {
          Swal.fire("Calculation ready", "Review the risk result before saving this assessment.", "success");
        } else {
          Swal.close();
        }
      })
      .catch(() => Swal.fire("Error", "Failed to calculate risk preview.", "error"));
  }

  function saveReviewedAssessment() {
    if (!validatePipelineForm(false)) return;
    if (!currentResult) {
      Swal.fire("Please calculate the risk before saving this assessment.", "The assessment can be saved after the result is available for review.", "warning");
      return;
    }
    if (resultIsStale || resultSignature !== calculationSignature(collectPayload())) {
      resultIsStale = true;
      updateSaveState();
      Swal.fire("Please calculate the risk before saving this assessment.", "Some input values changed. Please recalculate the risk result before saving.", "warning");
      return;
    }
    if (!currentResult.final_risk_code || !currentResult.final_risk_level) {
      Swal.fire("Please calculate the risk result before saving.", "Final risk code and risk level are required.", "warning");
      return;
    }

    const id = $form.data("assessment-id");
    const payload = collectPayload();
    Swal.fire({ title: "Saving...", allowOutsideClick: false, showConfirmButton: false });
    Swal.showLoading();

    if (id) {
      persistCalculation(id, payload);
      return;
    }

    fetch("/assessment-pipeline/submit", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    })
      .then((r) => r.json())
      .then((res) => {
        if (res.status !== "success" || !res.id) {
          Swal.fire("Check the form", humanizeServerError(res.message), "error");
          return;
        }
        persistCalculation(res.id, payload);
      })
      .catch(() => Swal.fire("Error", "Failed to save assessment.", "error"));
  }

  function persistCalculation(id, payload) {
    fetch(`/assessment-pipeline/calculate/${id}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    })
      .then((r) => r.json())
      .then((res) => {
        if (res.status !== "success") {
          Swal.fire("Check the form", humanizeServerError(res.message), "error");
          return;
        }
        Swal.fire("Success", res.message, "success").then(() => {
          window.location.href = `/assessment-pipeline/view/${id}`;
        });
      })
      .catch(() => Swal.fire("Error", "Failed to save calculation result.", "error"));
  }

  function collectPayload() {
    const data = {};
    $form.serializeArray().forEach((item) => {
      if (item.name.startsWith("point_") || ["inspection_point", "location_class", "installation_type", "measured_year"].includes(item.name)) return;
      data[item.name] = numericOrString(item.value);
    });
    if (numberValue(data.nominal_wall_thickness_mm) > 0 && (numberValue(data.actual_wall_thickness_mm) <= 0 || numberValue(data.actual_wall_thickness_mm) > numberValue(data.nominal_wall_thickness_mm))) {
      data.actual_wall_thickness_mm = data.nominal_wall_thickness_mm;
    }
    if (numberValue(data.internal_design_pressure_psi) <= 0) {
      data.internal_design_pressure_psi = data.operating_pressure_psi;
    }

    data.RiskInput = {
      damage_mechanism: $("input[name='damage_mechanism']").val() || "Pipeline MVP Index Risk",
      inspection_effectivity: $("input[name='inspection_effectivity']").val() || "Representative",
      release_fluid: $("[name='service']").val(),
      generic_failure_frequency: numberValue($("input[name='generic_failure_frequency']").val()),
      management_system_score: numberValue($("input[name='management_system_score']").val()),
      base_tpd_rate: numberValue($("input[name='base_tpd_rate']").val()),
      base_external_corr_rate: numberValue($("input[name='base_external_corr_rate']").val()),
      base_internal_corr_rate: numberValue($("input[name='base_internal_corr_rate']").val()),
      depth_of_cover: $("[name='depth_of_cover']").val(),
      patrol_frequency: $("[name='patrol_frequency']").val(),
      row_condition: $("[name='row_condition']").val(),
      soil_resistivity: $("[name='soil_resistivity']").val(),
      coating_condition: $("[name='coating_condition']").val(),
      cp_status: $("[name='cp_status']").val(),
      cp_potential_mv: numberValue($("input[name='cp_potential_mv']").val()),
      fluid_corrosivity: $("[name='fluid_corrosivity']").val(),
      water_content: $("[name='water_content']").val(),
      co2_h2s_presence: $("[name='co2_h2s_presence']").val(),
      mic_risk: $("[name='mic_risk']").val(),
      wall_thickness_condition: $("[name='wall_thickness_condition']").val(),
      building_count_inside_pir: parseInt($("input[name='building_count_inside_pir']").val(), 10) || 0,
      class_location: $("[name='class_location']").val(),
      emergency_response: $("[name='emergency_response']").val(),
      flow_rate: numberValue($("input[name='flow_rate']").val()),
      detection_time_hours: numberValue($("input[name='detection_time_hours']").val()),
      segment_length_between_valves_m: numberValue($("input[name='segment_length_between_valves_m']").val()),
      environmental_sensitivity: $("[name='environmental_sensitivity']").val(),
      nearby_sensitive_receptor: $("input[name='nearby_sensitive_receptor']").is(":checked"),
      isolation_valve_available: $("input[name='isolation_valve_available']").is(":checked"),
      consequence_basis: "Pipeline MVP index-based CoF",
      probability_basis: "PoF = GFF x max(DF_TPD, DF_EXTERNAL, DF_INTERNAL) x FMS",
      engineering_notes: $("textarea[name='engineering_notes']").val(),
      requires_confirmation: false,
      confirmation_todo_reason: "",
    };
    [
      "damage_mechanism", "inspection_effectivity", "generic_failure_frequency", "management_system_score",
      "base_tpd_rate", "base_external_corr_rate", "base_internal_corr_rate", "depth_of_cover",
      "patrol_frequency", "row_condition", "soil_resistivity", "coating_condition", "cp_status",
      "cp_potential_mv", "fluid_corrosivity", "water_content", "co2_h2s_presence", "mic_risk",
      "wall_thickness_condition", "building_count_inside_pir", "class_location", "emergency_response",
      "flow_rate", "detection_time_hours", "segment_length_between_valves_m", "environmental_sensitivity",
      "nearby_sensitive_receptor", "isolation_valve_available", "engineering_notes"
    ].forEach((key) => delete data[key]);

    data.inspection_points = [];
    $("#pipelinePointsTable tbody tr").each(function () {
      const nominal = numberValue($(this).find("[name='point_nominal_thickness_mm']").val());
      const actual = numberValue($(this).find("[name='point_actual_thickness_mm']").val());
      data.inspection_points.push({
        inspection_point: $(this).find("[name='inspection_point']").val(),
        location_class: $(this).find("[name='location_class']").val(),
        installation_type: $(this).find("[name='installation_type']").val(),
        nominal_thickness_mm: nominal,
        actual_thickness_mm: actual > 0 && actual <= nominal ? actual : nominal,
        measured_year: parseInt($(this).find("[name='measured_year']").val(), 10) || 0,
      });
    });
    return data;
  }

  function renderCalculationResult(result) {
    $(".pipeline-result-section").removeClass("d-none");
    $("#pipelineResultEmptyNotice").addClass("d-none");
    $("#pipelineResultStaleNotice").toggleClass("d-none", !resultIsStale);
    $("#pipelineRiskCode").text(result.final_risk_code || "-");
    $("#pipelineRiskLevel")
      .attr("class", `badge rounded-pill px-3 py-2 ${riskBadgeClass(result.final_risk_level)}`)
      .text(normalizeRiskLevel(result.final_risk_level));
    $("#pipelineRiskExplanation").text(`This pipeline is classified as ${normalizeRiskLevel(result.final_risk_level)} because the probability of failure is ${result.pof || "-"} and the consequence category is ${result.cof || "-"}.`);
    $("#pipelineRiskMatrix").html(buildRiskMatrix(result.final_risk_code));
    $("#pipelineThicknessResult").html(buildThicknessResult(result.point_results || []));
    $("#pipelinePofBreakdown").html(listItems([
      ["Third-Party Damage Factor", fmt(result.third_party_damage_factor)],
      ["External Corrosion Factor", fmt(result.external_corrosion_factor)],
      ["Internal Corrosion Factor", fmt(result.internal_corrosion_factor)],
      ["Governing DF", fmt(result.governing_damage_factor)],
      ["Main Failure Driver", result.governing_damage_mechanism],
      ["GFF", fmt(result.generic_failure_frequency)],
      ["FMS", fmt(result.management_system_factor)],
      ["Final PoF", fmt(result.pof_value)],
      ["PoF Category", result.pof],
    ]));
    $("#pipelineCofBreakdown").html(buildCofBreakdown(result));
    $("#pipelineRecommendationGroups").html(buildRecommendationGroups(result));
    $("#pipelineRecommendationText").html(`<strong>Generated recommendation:</strong> ${escapeHtml(result.recommendation || "-")}`);
    $("#pipelineFormulaTrace").html(buildFormulaTrace(result.formula_trace || []));
  }

  function buildCofBreakdown(result) {
    const isGas = String($("[name='service']").val()).toLowerCase() === "gas";
    if (isGas) {
      return listItems([
        ["Potential Impact Radius", `${fmt(result.pir_feet)} ft`],
        ["Buildings Inside Gas Impact Radius", $("input[name='building_count_inside_pir']").val() || 0],
        ["Class Location", labelFor($("[name='class_location']"))],
        ["CoF Category", result.cof],
      ]);
    }
    return listItems([
      ["Estimated Spill Volume", `${fmt(result.spill_volume)} bbl`],
      ["Adjusted Spill Volume", `${fmt(result.adjusted_spill_volume)} bbl`],
      ["Detection Time", valueWithUnit("detection_time_hours", "hr")],
      ["Environmental Sensitivity", labelFor($("[name='environmental_sensitivity']"))],
      ["CoF Category", result.cof],
    ]);
  }

  function buildRecommendationGroups(result) {
    const driver = result.governing_damage_mechanism || "";
    const isGas = String($("[name='service']").val()).toLowerCase() === "gas";
    const immediate = [];
    const inspection = ["Keep the formula trace with the assessment record."];
    const longTerm = ["Update the assessment after mitigation or inspection results are available."];
    if (driver === "Third-Party Damage") {
      immediate.push("Improve route markers and strengthen excavation permit control.");
      inspection.push("Increase ROW patrol frequency.");
    } else if (driver === "External Corrosion") {
      immediate.push("Verify cathodic protection performance and coating condition.");
      inspection.push("Plan CIPS/DCVG survey and soil monitoring.");
    } else if (driver === "Internal Corrosion") {
      immediate.push("Review inhibitor condition, fluid corrosivity, and water handling.");
      inspection.push("Schedule pigging, fluid sampling, and wall thickness inspection.");
    }
    if (["Critical Risk", "High Risk"].includes(result.final_risk_level)) immediate.push("Assign mitigation owner and target date.");
    longTerm.unshift(isGas ? "Review class location, public awareness, emergency response, and populated-area protection." : "Improve leak detection, isolation time, spill containment, and drainage/river protection.");

    return `
      <h6 class="fw-bold text-danger">Immediate Actions</h6>${bulletList(immediate)}
      <h6 class="fw-bold text-warning mt-3">Inspection / Monitoring</h6>${bulletList(inspection)}
      <h6 class="fw-bold text-success mt-3">Long-Term Mitigation</h6>${bulletList(longTerm)}
    `;
  }

  function buildRiskMatrix(activeCode) {
    const rows = ["5", "4", "3", "2", "1"];
    const cols = ["A", "B", "C", "D", "E"];
    const body = rows.map((row, index) => {
      const pofLabel = index === 0 ? '<td rowspan="5" class="rm-label" style="writing-mode: vertical-lr; transform: rotate(180deg);">PoF</td>' : "";
      const cells = cols.map((col) => {
        const code = `${row}${col}`;
        return `<td class="${matrixClass(row, col)} ${code === activeCode ? "active-cell" : ""}">${code}<br><small>${matrixLevel(row, col)}</small></td>`;
      }).join("");
      return `<tr>${pofLabel}<td class="rm-label">${row}</td>${cells}</tr>`;
    }).join("");
    return `<table class="pipeline-risk-matrix">${body}<tr><td class="border-0"></td><td class="border-0"></td>${cols.map((c) => `<td class="rm-label">${c}</td>`).join("")}</tr><tr><td colspan="7" class="border-0 text-muted small">Consequence of Failure (CoF)</td></tr></table>`;
  }

  function matrixLevel(row, col) {
    const score = Number(row) * ({ A: 1, B: 2, C: 3, D: 4, E: 5 }[col] || 1);
    if (score >= 20) return "Critical";
    if (score >= 12) return "High";
    if (score >= 6) return "Medium";
    return "Low";
  }

  function matrixClass(row, col) {
    return `risk-${matrixLevel(row, col).toLowerCase()}`;
  }

  function buildThicknessResult(points) {
    if (!points.length) return '<tr><td colspan="7" class="text-center text-muted py-4">No thickness appraisal result returned.</td></tr>';
    return points.map((point) => `
      <tr>
        <td>${escapeHtml(point.inspection_point || "-")}</td>
        <td>${escapeHtml(fmt(point.actual_thickness_mm))}</td>
        <td>${escapeHtml(fmt(point.corrosion_rate_mm_year))}</td>
        <td>${escapeHtml(fmt(point.remaining_life_years))}</td>
        <td>${escapeHtml(fmt(point.hoop_stress_psi))}</td>
        <td>${escapeHtml(fmt(point.maop_psi))}</td>
        <td>${escapeHtml([point.thickness_status, point.maop_status].filter(Boolean).join(" / ") || "-")}</td>
      </tr>
    `).join("");
  }

  function markResultStaleIfNeeded() {
    if (!currentResult) {
      updateSaveState();
      return;
    }
    resultIsStale = resultSignature !== calculationSignature(collectPayload());
    $("#pipelineResultStaleNotice").toggleClass("d-none", !resultIsStale);
    updateSaveState();
  }

  function updateSaveState() {
    const canSave = !!currentResult && !resultIsStale;
    $("#savePipelineDraft").prop("disabled", !canSave);
  }

  function calculationSignature(payload) {
    const clone = JSON.parse(JSON.stringify(payload));
    delete clone.assessment_by;
    if (clone.RiskInput) delete clone.RiskInput.engineering_notes;
    return JSON.stringify(clone);
  }

  function validateCurrentStep() {
    const $activeStep = $(".bs-stepper-content .content.active");
    return validateFields($activeStep.find("[required]:visible"));
  }

  function validatePipelineForm(calculate) {
    let valid = validateFields($form.find("[required]"), false);
    if (!calculate) {
      if (!valid) focusFirstInvalid();
      return valid;
    }

    const numericFields = [
      ["outside_diameter_in", "Outside diameter must be a valid number."],
      ["nominal_wall_thickness_mm", "Wall thickness must be a valid number."],
      ["actual_wall_thickness_mm", "Actual wall thickness must be a valid number."],
      ["pipe_length_m", "Segment length must be a valid number."],
      ["internal_design_pressure_psi", "Internal design pressure must be a valid number."],
      ["operating_pressure_psi", "Operating pressure must be a valid number."],
      ["smys_psi", "SMYS must be a valid number."],
      ["design_factor", "Design factor must be a valid number."],
      ["quality_factor", "Quality factor must be a valid number."],
      ["weld_joint_strength_factor", "Weld joint strength factor must be a valid number."],
      ["material_stress_psi", "Material stress must be a valid number."],
      ["generic_failure_frequency", "Generic Failure Frequency (GFF) must be a valid number."],
      ["management_system_score", "Management score must be a valid number."],
      ["base_tpd_rate", "Base third-party damage rate must be a valid number."],
      ["base_external_corr_rate", "Base external corrosion rate must be a valid number."],
      ["base_internal_corr_rate", "Base internal corrosion rate must be a valid number."],
    ];
    if (String($("[name='service']").val()).toLowerCase() !== "gas") {
      numericFields.push(["flow_rate", "Flow rate must be a valid number."]);
      numericFields.push(["detection_time_hours", "Leak detection time must be a valid number."]);
      numericFields.push(["segment_length_between_valves_m", "Segment length between isolation valves must be a valid number."]);
    }

    numericFields.forEach(([name, message]) => {
      const $field = $(`[name='${name}']`);
      if (!$field.length) return;
      if (numberValue($field.val()) <= 0) {
        setFieldError($field, message);
        valid = false;
      }
    });

    if (!valid) focusFirstInvalid();
    return valid;
  }

  function validateFields($fields, showAlert = true) {
    let valid = true;
    $fields.each(function () {
      const $field = $(this);
      if (String($field.val()).trim() === "") {
        setFieldError($field, $field.data("required-msg") || "Please complete this field.");
        valid = false;
      }
    });
    if (!valid) {
      focusFirstInvalid();
      if (showAlert) Swal.fire("Please complete the highlighted fields", "The form needs a few details before continuing.", "warning");
    }
    return valid;
  }

  function updateConsequenceFields() {
    const isGas = String($("[name='service']").val()).toLowerCase() === "gas";
    $(".pipeline-gas-fields").toggleClass("d-none", !isGas);
    $(".pipeline-liquid-fields").toggleClass("d-none", isGas);
    $("[name='flow_rate']").closest(".col-md-4").toggleClass("d-none", isGas);
  }

  function updateReviewSummary() {
    const $summary = $("#pipelineReviewSummary");
    if (!$summary.length) return;
    const fluidType = $("[name='service']").val() || "-";
    const isGas = String(fluidType).toLowerCase() === "gas";
    const items = [
      ["Tag Number", $("[name='report_no']").val()],
      ["Pipeline Name", $("[name='line_identification']").val()],
      ["Fluid Type", fluidType],
      ["Location / Area", $("[name='location']").val()],
      ["Class Location", labelFor($("[name='class_location']"))],
      ["Outside Diameter", valueWithUnit("outside_diameter_in", "in")],
      ["Wall Thickness", valueWithUnit("nominal_wall_thickness_mm", "mm")],
      ["Segment Length", valueWithUnit("pipe_length_m", "m")],
      ["Operating Pressure", valueWithUnit("operating_pressure_psi", "psi")],
      ["Main Consequence Basis", isGas ? `${$("input[name='building_count_inside_pir']").val() || 0} buildings inside PIR` : `${$("input[name='detection_time_hours']").val() || 0} hr leak detection`],
    ];
    $summary.html(items.map(([label, value]) => `
      <div class="col-md-6">
        <div class="border rounded p-3 h-100">
          <div class="text-muted small">${escapeHtml(label)}</div>
          <div class="fw-semibold">${escapeHtml(value || "-")}</div>
        </div>
      </div>
    `).join(""));
  }

  function setFieldError($field, message) {
    $field.addClass("is-invalid");
    const $group = $field.closest(".input-group");
    const $target = $group.length ? $group : $field;
    if (!$target.next(".invalid-feedback").length) {
      $target.after(`<div class="invalid-feedback d-block">${escapeHtml(message)}</div>`);
    } else {
      $target.next(".invalid-feedback").text(message);
    }
  }

  function clearFieldError($field) {
    $field.removeClass("is-invalid");
    const $group = $field.closest(".input-group");
    const $target = $group.length ? $group : $field;
    $target.next(".invalid-feedback").remove();
  }

  function focusFirstInvalid() {
    const $first = $(".is-invalid").first();
    if (!$first.length) return;
    const $step = $first.closest(".content");
    if ($step.length && pipelineStepper) {
      const stepIndex = $(".bs-stepper-content .content").index($step);
      if (stepIndex >= 0) pipelineStepper.to(stepIndex + 1);
    }
    setTimeout(() => $first.trigger("focus"), 150);
  }

  function listItems(items) {
    return items.map(([label, value]) => `<li class="list-group-item d-flex justify-content-between gap-3 px-0"><span>${escapeHtml(label)}</span><strong class="text-end">${escapeHtml(value || "-")}</strong></li>`).join("");
  }

  function bulletList(items) {
    const safeItems = items.length ? items : ["Review the generated recommendation and assign follow-up actions."];
    return `<ul class="mb-0 small">${safeItems.map((item) => `<li>${escapeHtml(item)}</li>`).join("")}</ul>`;
  }

  function buildFormulaTrace(trace) {
    if (!trace.length) return '<tr><td colspan="4" class="text-center text-muted py-4">No calculation details returned.</td></tr>';
    return trace.map((item) => {
      const inputs = item.inputs ? Object.entries(item.inputs).map(([key, value]) => `${key}: ${formatTraceValue(value)}`).join("<br>") : "-";
      return `<tr><td>${escapeHtml(item.formula_name || "-")}</td><td>${escapeHtml(item.excel_ref || "-")}</td><td class="text-wrap" style="min-width:320px;"><div>${escapeHtml(item.expression || "-")}</div><div class="text-muted small mt-1">${inputs}</div></td><td>${escapeHtml(formatTraceValue(item.output))}</td></tr>`;
    }).join("");
  }

  function humanizeServerError(message) {
    if (!message) return "Please review the highlighted inputs and try again.";
    return String(message)
      .replace("pipeline oil validation failed:", "Please review these fields:")
      .replaceAll("report_no required", "Please enter the pipeline tag number.")
      .replaceAll("line_identification required", "Please enter the pipeline name.")
      .replaceAll("assessment_by required", "Please enter the assessor name.")
      .replaceAll("outside_diameter_in must be greater than zero", "Outside diameter must be a valid number.")
      .replaceAll("nominal_wall_thickness_mm must be greater than zero", "Wall thickness must be a valid number.")
      .replaceAll("operating_pressure_psi cannot be negative", "Operating pressure cannot be negative.")
      .replaceAll("RiskInput.flow_rate must be greater than zero", "Flow rate must be a valid number.")
      .replaceAll("RiskInput.detection_time_hours must be greater than zero", "Leak detection time must be a valid number.")
      .replaceAll("RiskInput.segment_length_between_valves_m must be greater than zero", "Segment length between isolation valves must be a valid number.");
  }

  function numericOrString(value) {
    if (value === "") return "";
    const numeric = Number(value);
    return Number.isFinite(numeric) && value.trim() !== "" ? numeric : value;
  }

  function numberValue(value) {
    const numeric = Number(value);
    return Number.isFinite(numeric) ? numeric : 0;
  }

  function valueWithUnit(name, unit) {
    const value = $(`[name='${name}']`).val();
    return value ? `${value} ${unit}` : "-";
  }

  function labelFor($select) {
    return $select.find("option:selected").text() || "-";
  }

  function fmt(value) {
    const numeric = Number(value);
    if (!Number.isFinite(numeric)) return value || "-";
    if (Math.abs(numeric) > 0 && Math.abs(numeric) < 0.001) return numeric.toExponential(3);
    return numeric.toLocaleString(undefined, { maximumFractionDigits: 4 });
  }

  function formatTraceValue(value) {
    if (value === null || typeof value === "undefined") return "-";
    if (typeof value === "object") return JSON.stringify(value);
    return String(value);
  }

  function normalizeRiskLevel(level) {
    return String(level || "-").toUpperCase();
  }

  function riskBadgeClass(level) {
    if (level === "Critical Risk") return "bg-label-dark";
    if (level === "High Risk") return "bg-label-danger";
    if (level === "Medium Risk") return "bg-label-warning";
    return "bg-label-success";
  }

  function escapeHtml(value) {
    return String(value ?? "").replace(/[&<>"']/g, (char) => ({
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#039;",
    }[char]));
  }
});

