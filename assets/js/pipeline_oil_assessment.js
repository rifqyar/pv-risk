$(function () {
  const $form = $("#pipelineOilForm");
  let pipelineStepper = null;
  let currentResult = null;
  let resultSignature = "";
  let resultIsStale = false;
  let isProgrammaticUpdate = false;
  let isRealtimeCalculating = false;
  let pendingRealtimeCalculation = false;
  let isSavingAssessment = false;
  let realtimeTimer = null;

  localizeDecimalInputs();

  const stepperEl = document.querySelector("#wizard-pipeline-assessment");
  if (stepperEl && typeof Stepper !== "undefined") {
    pipelineStepper = new Stepper(stepperEl, { linear: false });
  }

  syncApplicableCodeAndMaterialStress(false);
  updateConsequenceFields();
  updatePreviousConditionDamageLevelVisibility();
  applySavedInspectionPlan();
  updateReviewSummary();
  setTimeout(() => {
    loadSavedResult();
    updateSaveState();
  }, 0);

  $(".pipeline-step-next").on("click", function () {
    const $activeStep = $(".bs-stepper-content .content.active");
    if (!validateCurrentStep()) return;
    updateReviewSummary();
    if (pipelineStepper) pipelineStepper.next();
    if ($activeStep.attr("id") === "pipeline-step-3") {
      runRealtimePipelineCalculation();
    } else if ($activeStep.attr("id") === "pipeline-step-5") {
      runRealtimePipelineCalculation();
    }
  });

  $(".pipeline-step-prev").on("click", function () {
    if (pipelineStepper) pipelineStepper.previous();
  });

  $(".step-trigger").on("click", function () {
    const target = $(this).closest(".step").data("target");
    setTimeout(() => {
      refreshDamageMechanismScreeningIfActive(target);
      refreshReviewCalculationIfActive();
    }, 0);
  });

  $form.on("input change", "input, select, textarea", function () {
    if (isProgrammaticUpdate) return;
    clearFieldError($(this));
    if (isManualTextField(this.name)) {
      updateSaveState();
      return;
    }
    if (this.name === "service") {
      syncApplicableCodeAndMaterialStress(true);
    } else if (this.name === "material_specification") {
      applySelectedPipelineMaterial();
      syncApplicableCodeAndMaterialStress(false);
    } else if (this.name === "applicable_code" || this.name === "smys_psi") {
      syncApplicableCodeAndMaterialStress(false);
    }
    updateConsequenceFields();
    updatePreviousConditionDamageLevelVisibility();
    updateReviewSummary();
    scheduleRealtimePipelineCalculationIfActive();
    markResultStaleIfNeeded();
  });

  $form.on("blur", "input[inputmode='decimal']", function () {
    formatDecimalInput($(this));
  });

  $("#addPipelinePoint").on("click", function () {
    $("#pipelinePointsTable tbody").append(`
      <tr>
        <td><input class="form-control" name="inspection_point"></td>
        <td><input class="form-control" name="location_class"></td>
        <td><select class="form-select" name="installation_type"><option value="">-</option><option value="Underground">Underground</option><option value="Above Ground">Above Ground</option></select></td>
        <td><input type="text" inputmode="decimal" class="form-control" name="point_nominal_thickness_mm"></td>
        <td><input type="text" inputmode="decimal" class="form-control" name="point_actual_thickness_mm"></td>
        <td><input type="month" class="form-control" name="measured_year" value="${new Date().getFullYear()}-${String(new Date().getMonth() + 1).padStart(2, "0")}"></td>
        <td><button type="button" class="btn btn-sm btn-icon btn-outline-danger remove-point"><i class="mdi mdi-trash-can-outline"></i></button></td>
      </tr>
    `);
    runRealtimePipelineCalculationIfActive();
    markResultStaleIfNeeded();
  });

  $(document).on("click", ".remove-point", function () {
    $(this).closest("tr").remove();
    runRealtimePipelineCalculationIfActive();
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
        updateAutoCalcDisplay(riskInputOf(payload), payload);
        updateSaveState();
        if (showSuccess) {
          Swal.fire("Calculation ready", "Review the risk result before saving this assessment.", "success");
        } else {
          Swal.close();
        }
      })
      .catch(() => Swal.fire("Error", "Failed to calculate risk preview.", "error"));
  }

  async function saveReviewedAssessment() {
    if (isSavingAssessment) return;
    if (!validatePipelineForm(true)) {
      Swal.fire("Complete the required fields before saving.", "A few required inputs are missing or invalid.", "warning");
      return;
    }
    const id = $form.data("assessment-id");
    const payload = collectPayload();
    isSavingAssessment = true;
    updateSaveState();
    Swal.fire({ title: "Saving...", allowOutsideClick: false, showConfirmButton: false });
    Swal.showLoading();

    try {
      const preview = await backendPreview(payload);
      currentResult = preview.result;
      resultSignature = calculationSignature(payload);
      resultIsStale = false;
      renderCalculationResult(currentResult);

      const assessmentID = id || await createDraftAssessment(payload);
      await persistCalculation(assessmentID, payload);
    } catch (err) {
      Swal.fire("Check the form", humanizeServerError(err.message), "error");
      isSavingAssessment = false;
      updateSaveState();
    }
  }

  async function backendPreview(payload) {
    const response = await fetch("/assessment-pipeline/preview", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const res = await response.json();
    if (res.status !== "success" || !res.result) {
      throw new Error(res.message || "Backend preview failed.");
    }
    return res;
  }

  async function createDraftAssessment(payload) {
    const response = await fetch("/assessment-pipeline/submit", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const res = await response.json();
    if (res.status !== "success" || !res.id) {
      throw new Error(res.message || "Failed to create pipeline draft.");
    }
    return res.id;
  }

  async function persistCalculation(id, payload) {
    const response = await fetch(`/assessment-pipeline/calculate/${id}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const res = await response.json();
    if (res.status !== "success") {
      throw new Error(res.message || "Failed to save calculation result.");
    }
    Swal.fire("Success", res.message, "success").then(() => {
      window.location.href = `/assessment-pipeline/view/${id}`;
    });
  }

  function collectPayload(includeGoverningDamageMechanism = true) {
    const data = {};
    $form.serializeArray().forEach((item) => {
      if (item.name.startsWith("point_") || ["inspection_point", "location_class", "installation_type", "measured_year"].includes(item.name)) return;
      if (item.name === "right_of_way") {
        data[item.name] = String(item.value || "").trim();
        return;
      }
      data[item.name] = numericOrString(item.value);
    });
    if (numberValue(data.nominal_wall_thickness_mm) > 0 && numberValue(data.actual_wall_thickness_mm) <= 0) {
      data.actual_wall_thickness_mm = data.nominal_wall_thickness_mm;
    }
    if (numberValue(data.internal_design_pressure_psi) <= 0) {
      data.internal_design_pressure_psi = data.operating_pressure_psi;
    }
    data.applicable_code = normalizeApplicableCode(data.applicable_code || codeForService(data.service));
    data.material_stress_psi = derivedMaterialStress(data);

    data.risk_input = {
      damage_mechanism: includeGoverningDamageMechanism ? governingDamageMechanismCode() : "internal_corrosion",
      inspection_effectivity: $("[name='inspection_effectivity']").val() || "Representative",
      inspection_effectivity_by_damage_mechanism: collectInspectionEffectivityByDM(),
      inspection_plan_by_damage_mechanism: collectInspectionPlanByDM(),
      release_fluid: $("[name='service']").val(),
      generic_failure_frequency: numberValue($("[name='generic_failure_frequency']").val()),
      management_system_score: numberValue($("[name='management_system_score']").val()),
      base_tpd_rate: numberValue($("[name='base_tpd_rate']").val()),
      base_external_corr_rate: numberValue($("[name='base_external_corr_rate']").val()),
      base_internal_corr_rate: numberValue($("[name='base_internal_corr_rate']").val()),
      depth_of_cover: $("[name='depth_of_cover']").val(),
      patrol_frequency: $("[name='patrol_frequency']").val(),
      row_condition: $("[name='row_condition']").val(),
      soil_resistivity: $("[name='soil_resistivity']").val(),
      coating_condition: $("[name='coating_condition']").val(),
      cp_status: $("[name='cp_status']").val(),
      cp_potential_mv: numberValue($("input[name='cp_potential_mv']").val()),
      coating_damage_level: selectedDamageLevel("coating_condition", "coating_damage_level"),
      one_call_system: $("[name='one_call_system']").val(),
      ph_level: $("[name='ph_level']").val(),
      fluid_corrosivity_mpy: $("[name='fluid_corrosivity_mpy']").val(),
      inhibitor_effectiveness: $("[name='inhibitor_effectiveness']").val(),
      biocide_treatment: $("[name='biocide_treatment']").val(),
      corrosion_monitoring_result: $("[name='corrosion_monitoring_result']").val(),
      h2s_ppm: $("[name='h2s_ppm']").val(),
      pwht_status: $("[name='pwht_status']").val(),
      weld_joint_type: $("[name='weld_joint_type']").val(),
      flow_velocity_condition: $("[name='flow_velocity_condition']").val(),
      solid_content: $("[name='solid_content']").val(),
      prev_ext_corrosion: $("[name='prev_ext_corrosion']").val(),
      prev_int_thinning: $("[name='prev_int_thinning']").val(),
      prev_int_cracking: $("[name='prev_int_cracking']").val(),
      prev_loc_int_corrosion: $("[name='prev_loc_int_corrosion']").val(),
      insulation_condition: $("[name='insulation_condition']").val(),
      ext_coating_condition: $("[name='ext_coating_condition']").val(),
      env_ext_cracking: $("[name='env_ext_cracking']").val(),
      co2_content: numberValue($("input[name='co2_content']").val()),
      h2s_content: numberValue($("input[name='h2s_content']").val()),
      h2o_content: numberValue($("input[name='h2o_content']").val()),
      n2_content: numberValue($("input[name='n2_content']").val()),
      co_content: numberValue($("input[name='co_content']").val()),
      chloride_content: parseInt($("input[name='chloride_content']").val(), 10) || 0,
      conf_ext_corrosion: $("[name='conf_ext_corrosion']").val(),
      conf_int_thinning: $("[name='conf_int_thinning']").val(),
      conf_int_cracking: $("[name='conf_int_cracking']").val(),
      conf_loc_int_corrosion: $("[name='conf_loc_int_corrosion']").val(),
      insulation_damage_level: selectedDamageLevel("insulation_condition", "insulation_damage_level"),
      ext_coating_damage_level: selectedDamageLevel("ext_coating_condition", "ext_coating_damage_level"),
      fluida: $("[name='fluida']").val(),
      phase: $("[name='phase']").val(),
      pressure_cycle_count: numberValue($("input[name='pressure_cycle_count']").val()),
      pressure_range_pct: numberValue($("input[name='pressure_range_pct']").val()),
      building_count_inside_pir: parseInt($("input[name='building_count_inside_pir']").val(), 10) || 0,
      class_location: $("[name='class_location']").val(),
      flow_rate: numberValue($("input[name='flow_rate']").val()),
      detection_time_hours: numberValue($("input[name='detection_time_hours']").val()),
      segment_length_between_valves_m: numberValue($("input[name='segment_length_between_valves_m']").val()),
      environmental_sensitivity: $("[name='environmental_sensitivity']").val(),
      nearby_sensitive_receptor: $("input[name='nearby_sensitive_receptor']").is(":checked"),
      isolation_valve_available: $("input[name='isolation_valve_available']").is(":checked"),
      consequence_basis: "Pipeline MVP index-based CoF",
      probability_basis: "PoF category is calculated from pipeline probability inputs and the governing damage mechanism.",
      engineering_notes: $("textarea[name='engineering_notes']").val(),
      requires_confirmation: false,
      confirmation_todo_reason: "",
    };
    [
      "damage_mechanism", "inspection_effectivity", "generic_failure_frequency", "management_system_score",
      "base_tpd_rate", "base_external_corr_rate", "base_internal_corr_rate", "depth_of_cover",
      "patrol_frequency", "row_condition", "soil_resistivity", "coating_condition", "cp_status",
      "cp_potential_mv", "coating_damage_level", "one_call_system", "ph_level", "fluid_corrosivity_mpy",
      "inhibitor_effectiveness", "biocide_treatment", "corrosion_monitoring_result", "h2s_ppm",
      "pwht_status", "weld_joint_type", "flow_velocity_condition", "solid_content",
      "prev_ext_corrosion", "prev_int_thinning", "prev_int_cracking", "prev_loc_int_corrosion",
      "insulation_condition", "ext_coating_condition", "env_ext_cracking",
      "co2_content", "h2s_content", "h2o_content", "n2_content", "co_content", "chloride_content",
      "conf_ext_corrosion", "conf_int_thinning", "conf_int_cracking", "conf_loc_int_corrosion",
      "insulation_damage_level", "ext_coating_damage_level", "fluida", "phase",
      "pressure_cycle_count", "pressure_range_pct",
      "building_count_inside_pir", "class_location",
      "flow_rate", "detection_time_hours", "segment_length_between_valves_m", "environmental_sensitivity",
      "nearby_sensitive_receptor", "isolation_valve_available", "engineering_notes"
    ].forEach((key) => delete data[key]);
    Object.keys(data).filter((key) => key.startsWith("inspection_effectivity_") || key.startsWith("inspection_nonintrusive_") || key.startsWith("inspection_intrusive_")).forEach((key) => delete data[key]);

    data.inspection_points = [];
    $("#pipelinePointsTable tbody tr").each(function () {
      const pointName = String($(this).find("[name='inspection_point']").val() || "").trim();
      const nominal = numberValue($(this).find("[name='point_nominal_thickness_mm']").val());
      const actual = numberValue($(this).find("[name='point_actual_thickness_mm']").val());
      const measuredYear = String($(this).find("[name='measured_year']").val() || "");
      if (!pointName || nominal <= 0 || actual <= 0 || !measuredYear) return;
      data.inspection_points.push({
        inspection_point: pointName,
        location_class: $(this).find("[name='location_class']").val(),
        installation_type: $(this).find("[name='installation_type']").val(),
        nominal_thickness_mm: nominal,
        actual_thickness_mm: actual,
        measured_year: measuredYear,
      });
    });
    data.risk_input.co2_partial_pressure_psig = calculateCO2PartialPressureJS(data.risk_input, data);
    data.risk_input.h2s_partial_pressure_psig = calculateH2SPartialPressureJS(data.risk_input, data);
    data.risk_input.wall_thickness_ratio = calculateWallThicknessRatioJS(data);
    data.risk_input.smys_utilization_pct = calculateSMYSUtilizationPctJS(data);
    return data;
  }

  function riskInputOf(payload) {
    return (payload && (payload.risk_input || payload.RiskInput)) || {};
  }

  function selectedDamageLevel(conditionName, levelName) {
    return $(`[name='${conditionName}']`).val() === "Damaged" ? $(`[name='${levelName}']`).val() : "";
  }

  function updatePreviousConditionDamageLevelVisibility() {
    toggleDamageLevelField("coating_condition", "coating_damage_level");
    toggleDamageLevelField("insulation_condition", "insulation_damage_level");
    toggleDamageLevelField("ext_coating_condition", "ext_coating_damage_level");
  }

  function toggleDamageLevelField(conditionName, levelName) {
    const shouldShow = $(`[name='${conditionName}']`).val() === "Damaged";
    const $field = $(`[name='${levelName}']`);
    const $wrap = $field.closest("[data-damage-level-field]");
    $wrap.toggleClass("d-none", !shouldShow);
    $field.prop("disabled", !shouldShow);
  }

  function updateAutoCalcDisplay(input, riskData) {
    var test = calculateH2SPartialPressureJS(input || {}, riskData || {});
    var pCO2 = numberValue(input.co2_partial_pressure_psig) || calculateCO2PartialPressureJS(input || {}, riskData || {});
    var pH2S = numberValue(input.h2s_partial_pressure_psig) || calculateH2SPartialPressureJS(input || {}, riskData || {});
    var wtr = numberValue(input.wall_thickness_ratio) || calculateWallThicknessRatioJS(riskData || {});
    var smysPct = numberValue(input.smys_utilization_pct) || calculateSMYSUtilizationPctJS(riskData || {});
    $("#autoPCO2").text(pCO2 > 0 ? fmtRate(pCO2) : "-");
    $("#autoPH2S").text(pH2S > 0 ? fmtRate(pH2S) : "-");
    $("#autoWTR").text(wtr > 0 ? fmtRate(wtr) : "-");
    $("#autoSMYS").text(smysPct > 0 ? fmtRate(smysPct) + "%" : "-");
  }

  function renderCalculationResult(result) {
    $(".pipeline-result-section").removeClass("d-none");
    $("#pipelineResultEmptyNotice").addClass("d-none");
    $("#pipelineResultStaleNotice").toggleClass("d-none", !resultIsStale);
    $("#pipelineRiskCode").text(result.final_risk_code || "-");
    $("#pipelineRiskLevel")
      .attr("class", `badge rounded-pill px-3 py-2 ${riskBadgeClass(result.final_risk_level)}`)
      .text(normalizeRiskLevel(result.final_risk_level));
    $("#pipelineRiskExplanation").text(`Risk drivers: Governing DM ${result.governing_damage_mechanism || "-"}, PoF ${result.pof || "-"}, and CoF ${result.cof || "-"}. Final matrix cell ${result.final_risk_code || "-"} maps to ${normalizeRiskLevel(result.final_risk_level)}.`);
    $("#pipelineRiskMatrix").html(buildRiskMatrix(result.final_risk_code));
    $("#pipelineThicknessResult").html(buildThicknessResult(result.point_results || []));
    $("#pipelineRealtimeThicknessResult").html(buildRealtimeThicknessResult(result.point_results || []));
    renderDamageMechanismResults(result.damage_mechanism_results || []);
    renderInspectionPlanResults(result.inspection_plan_results || []);
    $("#pipelinePofBreakdown").html(listItems([
      ["Third-Party Damage Result", damageMechanismSummary(result, "third_party_mechanical_damage", result.third_party_damage_factor)],
      ["External Corrosion Result", damageMechanismSummary(result, "external_corrosion", result.external_corrosion_factor)],
      ["Internal Corrosion Result", damageMechanismSummary(result, "internal_corrosion", result.internal_corrosion_factor)],
      ["Governing Damage Mechanism", governingDamageMechanismSummary(result)],
      ["Main Failure Driver", result.governing_damage_mechanism],
      ["GFF Basis", selectedOptionLabel("generic_failure_frequency") || fmtEngineering(result.generic_failure_frequency)],
      ["Management System Basis", selectedOptionLabel("management_system_score") || "-"],
      ["Final PoF", fmtEngineering(result.pof_value)],
      ["PoF Category", result.pof],
    ]));
    $("#pipelineCofBreakdown").html(buildCofBreakdown(result));
    $("#pipelineRecommendationText").html(`<strong>Source:</strong> ${escapeHtml(result.recommendation_source || "Realtime browser preview; backend recalculates on save.")}<br><strong>Confidence:</strong> ${escapeHtml(result.recommendation_confidence || "Low")}<br><strong>Advisory:</strong> ${escapeHtml(result.recommendation || "-")}`);
    $("#pipelineFormulaTrace").html(buildFormulaTrace(result.formula_trace || []));
  }

  function buildCofBreakdown(result) {
    const isGas = isGasService($("[name='service']").val());
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

  function damageMechanismSummary(result, code, fallbackScore) {
    const mechanism = (result.damage_mechanism_results || []).find((item) => item.code === code);
    if (mechanism) {
      return `${mechanism.severity || severityFromScoreLabel(mechanism.score)} - ${mechanism.label || codeLabel(code)}`;
    }
    return `${severityFromScoreLabel(fallbackScore)} - ${codeLabel(code)}`;
  }

  function governingDamageMechanismSummary(result) {
    const label = result.governing_damage_mechanism || "-";
    const mechanism = (result.damage_mechanism_results || []).find((item) => item.label === label);
    if (mechanism) return `${mechanism.severity || severityFromScoreLabel(mechanism.score)} - ${label}`;
    return `${severityFromScoreLabel(result.governing_damage_factor)} - ${label}`;
  }

  function severityFromScoreLabel(score) {
    const value = numberValue(score);
    if (value <= 0) return "NOT";
    if (value < 1.5) return "Low";
    if (value < 3) return "Moderate";
    return "High";
  }

  function codeLabel(code) {
    return String(code || "-")
      .replaceAll("_", " ")
      .replace(/\b\w/g, (char) => char.toUpperCase());
  }

  function selectedOptionLabel(name) {
    const $field = $(`[name='${name}']`);
    if (!$field.length || !$field.is("select")) return "";
    return $field.find("option:selected").text();
  }

  function buildRecommendationGroups(result) {
    const groups = result.recommendation_groups || {};
    const immediate = groups.immediate_actions || [];
    const inspection = groups.inspection_monitoring || [];
    const longTerm = groups.long_term_mitigation || [];

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
    if (!points.length) return '<tr><td colspan="11" class="text-center text-muted py-4">No thickness appraisal result returned.</td></tr>';
    return points.map((point) => `
      <tr>
        <td>${escapeHtml(point.inspection_point || "-")}</td>
        <td>${escapeHtml(fmt(point.nominal_thickness_mm || point.actual_thickness_mm))}</td>
        <td>${escapeHtml(fmt(point.required_thickness_mm))}</td>
        <td>${escapeHtml(fmt(point.minimum_thickness_mm || point.required_thickness_mm))}</td>
        <td>${escapeHtml(fmt(point.actual_thickness_mm))}</td>
        <td>${escapeHtml(fmt(point.remaining_thickness_mm))}</td>
        <td>${escapeHtml(fmtRate(point.corrosion_rate_mm_year))}</td>
        <td>${escapeHtml(fmt(nonNegative(point.remaining_life_years)))}</td>
        <td>${escapeHtml(fmt(point.hoop_stress_psi))}</td>
        <td>${escapeHtml(fmt(point.maop_psi))}</td>
        <td>${statusBadges([overallTechnicalStatus(point)])}${technicalStatusCauses(point)}</td>
      </tr>
    `).join("");
  }

  function buildRealtimeThicknessResult(points) {
    updateThicknessSummary(points);
    if (!points.length) return '<tr><td colspan="11" class="text-center text-muted py-3">Enter thickness readings to see realtime calculation.</td></tr>';
    return points.map((point) => `
      <tr>
        <td>${escapeHtml(point.inspection_point || "-")}</td>
        <td>${escapeHtml(fmt(point.nominal_thickness_mm))}</td>
        <td>${escapeHtml(fmt(point.required_thickness_mm))}</td>
        <td>${escapeHtml(fmt(point.minimum_thickness_mm))}</td>
        <td>${escapeHtml(fmt(point.actual_thickness_mm))}</td>
        <td>${escapeHtml(fmt(point.remaining_thickness_mm))}</td>
        <td>${escapeHtml(fmtRate(point.corrosion_rate_mm_year))} mm/year</td>
        <td>${escapeHtml(fmt(nonNegative(point.remaining_life_years)))} years</td>
        <td>${escapeHtml(fmt(point.hoop_stress_psi))} psi</td>
        <td>${escapeHtml(fmt(point.maop_psi))} psi</td>
        <td>${statusBadges([overallTechnicalStatus(point)])}${technicalStatusCauses(point)}</td>
      </tr>
    `).join("");
  }

  function statusBadges(values) {
    const statuses = (values || []).filter(Boolean);
    if (!statuses.length) return "-";
    return statuses.map((status) => {
      const text = String(status || "").trim();
      const key = text.toLowerCase();
      let badgeClass = "conditional";
      let icon = "mdi-alert-circle-outline";
      if (key.includes("not acceptable")) {
        badgeClass = "not-acceptable";
        icon = "mdi-close-circle-outline";
      } else if (key.includes("acceptable")) {
        badgeClass = "acceptable";
        icon = "mdi-check-circle-outline";
      }
      return `<span class="pipeline-acceptance-badge ${badgeClass} me-1 mb-1"><i class="mdi ${icon}"></i>${escapeHtml(text)}</span>`;
    }).join("");
  }

  function overallTechnicalStatus(point) {
    const statuses = [point.thickness_status, point.hoop_stress_status, point.maop_status]
      .filter(Boolean)
      .map((status) => String(status).trim().toUpperCase());
    if (!statuses.length) return "";
    return statuses.some((status) => status === "NOT ACCEPTABLE")
      ? "ENGINEERING EVALUATION REQUIRED"
      : "ACCEPTABLE FOR CONTINUED SERVICE";
  }

  function technicalStatusCauses(point) {
    const causes = [];
    const notes = [];
    if (String(point.thickness_status || "").toUpperCase() === "NOT ACCEPTABLE") causes.push("Actual thickness below required/minimum thickness");
    if (String(point.hoop_stress_status || "").toUpperCase() === "NOT ACCEPTABLE") causes.push("Hoop stress exceeds allowable stress basis");
    if (String(point.maop_status || "").toUpperCase() === "NOT ACCEPTABLE") causes.push("MAOP is below design pressure");
    if (numberValue(point.actual_thickness_mm) > numberValue(point.nominal_thickness_mm)) {
      notes.push("Actual thickness exceeds nominal (no corrosion loss detected)");
    }
    let html = "";
    if (causes.length) {
      html += `<div class="small text-danger mt-1">${causes.map(escapeHtml).join("<br>")}</div>`;
    }
    if (notes.length) {
      html += `<div class="small text-info mt-1"><i class="mdi mdi-information-outline me-1"></i>${notes.map(escapeHtml).join("<br>")}</div>`;
    }
    return html;
  }

  function nonNegative(value) {
    return Math.max(numberValue(value), 0);
  }

  function renderDamageMechanismResults(results) {
    results.forEach((item) => {
      const badgeClass = severityBadgeClass(item.severity);
      $(`.pipeline-dm-badge[data-dm-code='${item.code}']`)
        .attr("class", `badge rounded-pill px-3 pipeline-dm-badge ${badgeClass}`)
        .text(item.severity || "NOT");
      $(`.pipeline-dm-effectivity-badge[data-dm-code='${item.code}']`)
        .attr("class", `badge pipeline-dm-effectivity-badge ${badgeClass}`)
        .text(item.severity === "NOT" ? "NOT" : `${item.severity || "NOT"}`);
    });
  }

  function renderInspectionPlanResults(results) {
    results.forEach((item) => {
      $(`.pipeline-inspection-period[data-dm-code='${item.code}'][data-scope='nonintrusive']`).text(`${item.non_intrusive_interval_months || "-"} months`);
      $(`.pipeline-inspection-period[data-dm-code='${item.code}'][data-scope='intrusive']`).text(`${item.intrusive_interval_months || "-"} months`);
      $(`.pipeline-inspection-effectivity[data-dm-code='${item.code}'][data-scope='nonintrusive']`)
        .toggleClass("d-none", item.severity === "NOT")
        .text(item.severity === "NOT" ? "" : `Effectivity: ${item.non_intrusive_effectivity || "-"}`);
      $(`.pipeline-inspection-effectivity[data-dm-code='${item.code}'][data-scope='intrusive']`)
        .toggleClass("d-none", item.severity === "NOT")
        .text(item.severity === "NOT" ? "" : `Effectivity: ${item.intrusive_effectivity || "-"}`);
    });
  }

  function updateThicknessSummary(points) {
    if (!points.length) {
      $("#pipelineSummaryRemainingLife").text("-");
      $("#pipelineSummaryRemainingPoint").text("Point: -");
      $("#pipelineSummaryMinActual").text("-");
      $("#pipelineSummaryHoopStress").text("-");
      $("#pipelineSummaryLowestMAOP").text("-");
      return;
    }
    const valid = points.filter((point) => Number.isFinite(Number(point.remaining_life_years)));
    const governing = valid.sort((a, b) => nonNegative(a.remaining_life_years) - nonNegative(b.remaining_life_years))[0];
    const minActual = Math.min(...points.map((point) => numberValue(point.actual_thickness_mm)).filter((value) => value > 0));
    const highestHoopStress = Math.max(...points.map((point) => numberValue(point.hoop_stress_psi)).filter((value) => value > 0));
    const lowestMAOP = Math.min(...points.map((point) => numberValue(point.maop_psi)).filter((value) => value > 0));
    $("#pipelineSummaryRemainingLife").text(governing ? `${fmt(nonNegative(governing.remaining_life_years))} years` : "-");
    $("#pipelineSummaryRemainingPoint").text(governing ? `Point: ${governing.inspection_point || "-"}` : "Point: -");
    $("#pipelineSummaryMinActual").text(Number.isFinite(minActual) ? `${fmt(minActual)} mm` : "-");
    $("#pipelineSummaryHoopStress").text(Number.isFinite(highestHoopStress) ? `${fmt(highestHoopStress)} psi` : "-");
    $("#pipelineSummaryLowestMAOP").text(Number.isFinite(lowestMAOP) ? `${fmt(lowestMAOP)} psi` : "-");
  }

  async function runRealtimePipelineCalculation() {
    if (isRealtimeCalculating) {
      pendingRealtimeCalculation = true;
      return;
    }
    isRealtimeCalculating = true;
    try {
      const payload = collectPayload();
      updateAutoCalcDisplay(riskInputOf(payload), payload);
      if (!hasMinimumRealtimeInputs(payload)) {
        return;
      }
      const preview = await backendPreview(payload);
      const result = preview.result;
      currentResult = result;
      resultSignature = calculationSignature(payload);
      resultIsStale = false;
      renderCalculationResult(result);
      updateAutoCalcDisplay(riskInputOf(payload), payload);
      updateSaveState();
    } catch (err) {
      updateSaveState();
    } finally {
      isRealtimeCalculating = false;
      if (pendingRealtimeCalculation) {
        pendingRealtimeCalculation = false;
        scheduleRealtimePipelineCalculation();
      }
    }
  }

  function hasMinimumRealtimeInputs(payload) {
    return (payload.inspection_points || []).length > 0 &&
      numberValue(payload.outside_diameter_in) > 0 &&
      numberValue(payload.internal_design_pressure_psi) > 0 &&
      numberValue(payload.smys_psi) > 0;
  }

  function refreshReviewCalculationIfActive() {
    if ($("#pipeline-step-6").hasClass("active")) {
      updateReviewSummary();
      runRealtimePipelineCalculation();
    }
  }

  function refreshDamageMechanismScreeningIfActive(target) {
    if (target === "#pipeline-step-4" || $("#pipeline-step-4").hasClass("active")) {
      runRealtimePipelineCalculation();
    }
  }

  function scheduleRealtimePipelineCalculation() {
    clearTimeout(realtimeTimer);
    realtimeTimer = setTimeout(runRealtimePipelineCalculation, 80);
  }

  function scheduleRealtimePipelineCalculationIfActive() {
    const stepID = activePipelineStepID();
    if (["pipeline-step-1","pipeline-step-2","pipeline-step-3", "pipeline-step-4", "pipeline-step-5", "pipeline-step-6"].includes(stepID)) {
      scheduleRealtimePipelineCalculation();
    }
  }

  function runRealtimePipelineCalculationIfActive() {
    const stepID = activePipelineStepID();
    if (["pipeline-step-1","pipeline-step-2","pipeline-step-3", "pipeline-step-4", "pipeline-step-5", "pipeline-step-6"].includes(stepID)) {
      runRealtimePipelineCalculation();
    }
  }

  function activePipelineStepID() {
    return $(".bs-stepper-content .content.active").attr("id") || "";
  }

  window.pipelineOilRecalculate = scheduleRealtimePipelineCalculationIfActive;

  // Confirmed engineering decision: Pipeline DM modifier values are fixed at 1.0.
  const dmFactors = {
    depth:       { "<1m": 1, "1-2m": 1, ">2m": 1 },
    patrol:      { rare: 1, monthly: 1, weekly_daily: 1 },
    row:         { poor: 1, fair: 1, good: 1 },
    soil:        { "<1000": 1, "1000-5000": 1, ">5000": 1 },
    coating:     { Good: 1, Damaged: 1, "Not Inspectable": 1, "Not Applicable": 1 },
    cp:          { failed: 1, borderline: 1, normal: 1 },
    environment: { low: 1, medium: 1.5, high: 2.5 },
    corrosivityMPY: { "<2 mpy": 1, "2-5 mpy": 1, "5-10 mpy": 1, ">10 mpy": 1 },
    ph:          { "≤4.5": 1, "4.5-6.5": 1, "6.5-8.5": 1, ">8.5": 1 },
    chloride:    { 1: 1, 2: 1, 3: 1, 4: 1, 5: 1 },
    inhibitor:   { "High (>90%)": 1, "Medium (60-90%)": 1, "Low (<60%)": 1, None: 1 },
    coatingDamage: { Small: 1, Medium: 1, Large: 1, Severe: 1 },
    insulationDamage: { Small: 1, Medium: 1, Large: 1, Severe: 1 },
    prevFinding:  { "No Finding": 0, Finding: 1, "Not Inspectable": 0.5 },
    confidence:   { high: 1, average: 1, low: 1 },
    weld:         { Seamless: 1, SAW: 1, ERW: 1, Other: 1 },
    pwht:         { Yes: 1, No: 1, Unknown: 1 },
    oneCall:      { "Active and Effective": 1, Limited: 1, None: 1 },
    h2sPpm:       { "<50 ppm": 1, "50-1000 ppm": 1, ">1000 ppm": 1 },
    flowVelocity: { "Low (<3 m/s)": 1, "Moderate (3-10 m/s)": 1, "High (10-20 m/s)": 1, "Very High (>20 m/s)": 1 },
    solidContent: { None: 1, Trace: 1, Moderate: 1, Heavy: 1 },
    sccStress:    { Low: 30, Moderate: 50, High: 1e9 },
    erosionVel:   { Low: 3, Moderate: 10, High: 20, "Very High": 1e9 },
    fatigueCycle:  { Low: 100, Moderate: 10000, High: 1000000 },
    wallThicknessRatio: { Acceptable: 1, "Conditionally Acceptable": 0.8, "Not Acceptable": 0 },
    extCracking:  { None: 1, H2S: 1, Chloride: 1, Hydrogen: 1, Marine: 1 },
    biocide:       { Yes: true, No: false, "Not Required": true },
    co2Severity:   { Low: 5, Moderate: 20, High: 1e9 },
    h2sSeverity:   { Not: 0.05, Low: 0.5, Moderate: 15, High: 1e9 },
    classLocation: { class_1: 1, class_2: 1, class_3: 1, class_4: 1 },
  };

  function pCO2SeverityJS(pCO2) {
    if (pCO2 <= 0) return "NOT";
    if (pCO2 <= 5) return "Low";
    return "Moderate";
  }

  function pH2SSeverityJS(pH2S) {
    if (pH2S < 0.05) return "NOT";
    if (pH2S < 0.5) return "Low";
    if (pH2S <= 15) return "Moderate";
    return "High";
  }

  const pipelineDMRuleMetadata = {
    external_corrosion: { source_standard: "API 571 / AMPP SP0169", confidence_level: "Medium", rule_status: "PARTIALLY_VERIFIED" },
    coating_degradation: { source_standard: "API 571 / AMPP SP0169", confidence_level: "Low", rule_status: "TODO_ENGINEERING_CONFIRMATION" },
    third_party_mechanical_damage: { source_standard: "API 570 / pipeline integrity management practice", confidence_level: "Low", rule_status: "TODO_ENGINEERING_CONFIRMATION" },
    internal_corrosion: { source_standard: "API 581 / API 571", confidence_level: "Medium", rule_status: "PARTIALLY_VERIFIED" },
    localized_corrosion: { source_standard: "API 571", confidence_level: "Low", rule_status: "TODO_ENGINEERING_CONFIRMATION" },
    erosion: { source_standard: "API 571 / DNV-RP-O501 concept", confidence_level: "Low", rule_status: "TODO_ENGINEERING_CONFIRMATION" },
    erosion_corrosion: { source_standard: "API 571", confidence_level: "Low", rule_status: "TODO_ENGINEERING_CONFIRMATION" },
    cracking: { source_standard: "NACE MR0175 / ISO 15156 / API 571", confidence_level: "Medium", rule_status: "PARTIALLY_VERIFIED" },
    scc: { source_standard: "API 571 / NACE MR0175 / ISO 15156", confidence_level: "Low", rule_status: "TODO_ENGINEERING_CONFIRMATION" },
    fatigue: { source_standard: "API 571", confidence_level: "Low", rule_status: "TODO_ENGINEERING_CONFIRMATION" },
    chemical_damage: { source_standard: "Engineering review stub", confidence_level: "Low", rule_status: "TODO_ENGINEERING_CONFIRMATION" },
  };

  function annotatePipelineDMResult(item) {
    const metadata = pipelineDMRuleMetadata[item.code] || {
      source_standard: "Engineering review required",
      confidence_level: "Low",
      rule_status: "TODO_ENGINEERING_CONFIRMATION",
    };
    const effectivityByDM = collectInspectionEffectivityByDM();
    return {
      ...item,
      inspection_effectivity: effectivityByDM[item.code] || "Medium",
      source: "Pipeline DM screening v2",
      source_standard: metadata.source_standard,
      confidence_level: metadata.confidence_level,
      rule_status: metadata.rule_status,
    };
  }

  function baseSeverityScoreJS(sev) {
    if (sev === "Low") return 1.0;
    if (sev === "Moderate") return 2.0;
    if (sev === "High") return 3.0;
    return 0;
  }

  function escalateByFindingJS(score, prevFinding, confidence) {
    if (prevFinding === "Finding") {
      let weight = dmFactors.confidence[String(confidence || "").toLowerCase()] || 1.0;
      return score + 1.0 * weight;
    }
    // "Not Inspectable" — no escalation applied, matching Go backend
    return score;
  }

  function scoreInternalCorrosionJS(risk, input) {
    const pCO2 = numberValue(input.co2_partial_pressure_psig) || calculateCO2PartialPressureJS(risk, input);
    const h2o = numberValue(risk.h2o_content);
    const corrosivity = risk.fluid_corrosivity_mpy;
    const hasCO2 = pCO2 > 0;
    const hasWater = h2o > 0;
    const hasCorrosivity = corrosivity && corrosivity !== "<2 mpy";
    const gatePassed = hasCO2 || hasWater || hasCorrosivity;
    if (!gatePassed) return { code: "internal_corrosion", label: "Internal Corrosion", category: "Internal Thinning", score: 0, severity: "NOT" };
    let score = 0;
    if (pCO2 > 0) {
      const sev = pCO2SeverityJS(pCO2);
      score = baseSeverityScoreJS(sev);
    } else {
      score = factor(dmFactors.corrosivityMPY, corrosivity) || 1.0;
    }
    if (h2o > 5) score += 0;
    const phMod = factor(dmFactors.ph, risk.ph_level) - 1; // currently 0
    score += phMod;
    const inhibitorMod = factor(dmFactors.inhibitor, risk.inhibitor_effectiveness) - 1;
    score += inhibitorMod;
    const biocideNoWater = (risk.biocide_treatment === "No" && h2o > 0);
    if (biocideNoWater) score += 0;
    score = escalateByFindingJS(score, risk.prev_int_thinning, risk.conf_int_thinning);
    const sev = severityFromScoreJS(score);
    return { code: "internal_corrosion", label: "Internal Corrosion", category: "Internal Thinning", score, severity: sev };
  }

  function scoreExternalCorrosionJS(risk, input) {
    const coating = risk.coating_condition;
    const cpStat = risk.cp_status;
    const cpPot = numberValue(risk.cp_potential_mv);
    const soil = risk.soil_resistivity;
    const baseRate = numberValue(risk.base_external_corr_rate);
    const coatingConcern = coating === "Damaged";
    const cpConcern = cpStat === "failed" || cpStat === "borderline";
    const soilConcern = soil === "<1000";
    const gatePassed = cpConcern || coatingConcern || soilConcern;
    if (!gatePassed) return { code: "external_corrosion", label: "External Corrosion", category: "External Damage", score: 0, severity: "NOT" };
    let soilFactor = factor(dmFactors.soil, soil);
    let coatingFactor = factor(dmFactors.coating, coating);
    let cpFactor = factor(dmFactors.cp, cpStat);
    if (cpPot !== 0 && cpPot > -850 && cpFactor < factor(dmFactors.cp, "borderline")) cpFactor = factor(dmFactors.cp, "borderline");
    let score = baseRate * soilFactor * coatingFactor * cpFactor;
    if (baseRate > 0) score = Math.max(score, 1.0);
    score = escalateByFindingJS(score, risk.prev_ext_corrosion, risk.conf_ext_corrosion);
    return { code: "external_corrosion", label: "External Corrosion", category: "External Damage", score, severity: severityFromScoreJS(score) };
  }

  function scoreLocalizedCorrosionJS(risk, input, internalResult) {
    const internalScore = internalResult ? internalResult.score : 0;
    const chloride = numberValue(risk.chloride_content);
    const phLevel = risk.ph_level;
    const prevLoc = risk.prev_loc_int_corrosion;
    const gatePassed = internalResult.severity !== "NOT" || chloride >= 3 || phLevel === "≤4.5" || prevLoc === "Finding";
    if (!gatePassed) return { code: "localized_corrosion", label: "Localized Corrosion", category: "Internal Thinning", score: 0, severity: "NOT" };
    let score = internalScore;
    if (chloride >= 3) score += factor(dmFactors.chloride, chloride) - 1;
    if (phLevel === "≤4.5" || phLevel === "4.5-6.5") score += factor(dmFactors.ph, phLevel) - 1;
    score = escalateByFindingJS(score, prevLoc, risk.conf_loc_int_corrosion);
    return { code: "localized_corrosion", label: "Localized Corrosion", category: "Internal Thinning", score, severity: severityFromScoreJS(score) };
  }

  function scoreErosionJS(risk) {
    const velocity = risk.flow_velocity_condition;
    const solids = risk.solid_content;
    const corrosivity = risk.fluid_corrosivity_mpy;
    const gatePassed = (velocity && velocity !== "" && velocity !== "Low (<3 m/s)") || (solids && solids !== "" && solids !== "None");
    if (!gatePassed) return { code: "erosion", label: "Erosion", category: "Internal Thinning", score: 0, severity: "NOT" };
    let score = 1.0;
    if (solids && solids !== "" && solids !== "None") score += factor(dmFactors.solidContent, solids) - 1;
    if (corrosivity && corrosivity !== "" && corrosivity !== "<2 mpy") score += factor(dmFactors.corrosivityMPY, corrosivity) - 1;
    return { code: "erosion", label: "Erosion", category: "Internal Thinning", score, severity: severityFromScoreJS(score) };
  }

  function scoreErosionCorrosionJS(risk, input, erosionResult) {
    if (erosionResult.severity === "NOT") return { code: "erosion_corrosion", label: "Erosion-Corrosion", category: "Internal Thinning", score: 0, severity: "NOT" };
    let score = erosionResult.score;
    const corrosivity = risk.fluid_corrosivity_mpy;
    score += factor(dmFactors.corrosivityMPY, corrosivity) - 1;
    return { code: "erosion_corrosion", label: "Erosion-Corrosion", category: "Internal Thinning", score, severity: severityFromScoreJS(score) };
  }

  function scoreCrackingJS(risk, input) {
    const pH2S = numberValue(input.h2s_partial_pressure_psig) || calculateH2SPartialPressureJS(risk, input);
    const h2sContent = numberValue(risk.h2s_content);
    const prevCracking = risk.prev_int_cracking || "No Finding";
    const gatePassed = pH2S >= 0.05 || prevCracking === "Finding" || h2sContent > 0;
    if (!gatePassed) return { code: "cracking", label: "Cracking", category: "Internal Cracking", score: 0, severity: "NOT" };
    let score = baseSeverityScoreJS(pH2SSeverityJS(pH2S));
    score += factor(dmFactors.pwht, risk.pwht_status) - 1;
    score += factor(dmFactors.weld, risk.weld_joint_type) - 1;
    score = escalateByFindingJS(score, prevCracking, risk.conf_int_cracking);
    return { code: "cracking", label: "Cracking", category: "Internal Cracking", score, severity: severityFromScoreJS(score) };
  }

  function scoreSCCJS(risk, input) {
    const smysPct = numberValue(input.smys_utilization_pct) || calculateSMYSUtilizationPctJS(input);
    const coating = risk.coating_condition;
    const cpStatus = risk.cp_status;
    const h2sContent = numberValue(risk.h2s_content);
    const stressConcern = smysPct >= 30;
    const coatingConcern = coating === "Damaged";
    const cpConcern = cpStatus === "failed" || cpStatus === "borderline";
    const h2sPresent = h2sContent > 0;
    const gatePassed = stressConcern && (coatingConcern || cpConcern || h2sPresent);
    if (!gatePassed) return { code: "scc", label: "SCC", category: "Internal Cracking", score: 0, severity: "NOT" };
    let score;
    if (smysPct >= 72) score = 3.0;
    else if (smysPct >= 50) score = 2.0;
    else score = 1.0;
    if (coatingConcern) score += factor(dmFactors.coating, coating) - 1;
    if (h2sPresent) score += 0;
    return { code: "scc", label: "SCC", category: "Internal Cracking", score, severity: severityFromScoreJS(score) };
  }

  function scoreFatigueJS(risk) {
    const cycles = numberValue(risk.pressure_cycle_count);
    const prevCracking = risk.prev_int_cracking || "No Finding";
    const gatePassed = cycles > 0 || prevCracking === "Finding";
    if (!gatePassed) return { code: "fatigue", label: "Fatigue", category: "Internal Cracking", score: 0, severity: "NOT" };
    let score = 1.0;
    if (numberValue(risk.pressure_range_pct) > 0) score += 0;
    score += factor(dmFactors.weld, risk.weld_joint_type) - 1;
    score = escalateByFindingJS(score, prevCracking, risk.conf_int_cracking);
    return { code: "fatigue", label: "Fatigue", category: "Internal Cracking", score, severity: severityFromScoreJS(score) };
  }

  function scoreCoatingDegradationJS(risk, input) {
    const coating = risk.coating_condition;
    const cpStatus = risk.cp_status;
    const cpPot = numberValue(risk.cp_potential_mv);
    const insulationCond = risk.insulation_condition;
    const coatingConcern = coating === "Damaged";
    const cpConcern = cpStatus === "failed" || cpStatus === "borderline";
    const insulationConcern = insulationCond === "Damaged";
    const gatePassed = coatingConcern || cpConcern || insulationConcern;
    if (!gatePassed) return { code: "coating_degradation", label: "Coating Degradation", category: "External Damage", score: 0, severity: "NOT" };
    let score = 1.0;
    const coatingDamageLevel = risk.coating_damage_level;
    if (coatingConcern && coatingDamageLevel) score += factor(dmFactors.coatingDamage, coatingDamageLevel) - 1;
    if (risk.soil_resistivity === "<1000") score += factor(dmFactors.soil, risk.soil_resistivity) - 1;
    if (cpPot !== 0 && cpPot > -850) score += 0;
    const opTempC = (5.0 / 9.0) * (numberValue(input.design_temperature_f) - 32);
    if (insulationConcern && opTempC >= 0 && opTempC <= 175) score += 0;
    score = escalateByFindingJS(score, risk.prev_ext_corrosion, risk.conf_ext_corrosion);
    return { code: "coating_degradation", label: "Coating Degradation", category: "External Damage", score, severity: severityFromScoreJS(score) };
  }

  function scoreThirdPartyDamageJS(risk) {
    const baseRate = numberValue(risk.base_tpd_rate);
    const depthFactor = factor(dmFactors.depth, risk.depth_of_cover);
    const patrolFactor = factor(dmFactors.patrol, risk.patrol_frequency);
    const rowFactor = factor(dmFactors.row, risk.row_condition);
    const oneCallFactor = factor(dmFactors.oneCall, risk.one_call_system);
    let score = baseRate;
    if (baseRate <= 0) score = 1.0;
    score = escalateByFindingJS(score, null, null);
    return { code: "third_party_mechanical_damage", label: "Third-Party / Mechanical Damage", category: "External Damage", score, severity: severityFromScoreJS(score) };
  }

  function scoreChemicalDamageJS() {
    return { code: "chemical_damage", label: "Chemical Damage", category: "Internal Thinning", score: 0, severity: "NOT" };
  }

  function severityFromScoreJS(score) {
    if (score <= 0) return "NOT";
    if (score < 1.5) return "Low";
    if (score < 3.0) return "Moderate";
    return "High";
  }

  function calculateCO2PartialPressureJS(risk, input) {
    const co2 = numberValue(risk.co2_content);
    const opPressure = numberValue(input.operating_pressure_psi);
    if (co2 <= 0 || opPressure <= 0) return 0;
    return co2 * opPressure;
  }

  function calculateH2SPartialPressureJS(risk, input) {
    const h2s = numberValue(risk.h2s_content);
    const opPressure = numberValue(input.operating_pressure_psi);
    if (h2s <= 0 || opPressure <= 0) return 0;
    return (h2s * opPressure) / 1000000;
  }

  function calculateWallThicknessRatioJS(input) {
    const points = input.inspection_points || [];
    if (!points.length) return 1.0;
    let minActual = points[0].actual_thickness_mm;
    let minRequired = numberValue(points[0].required_thickness_mm);
    const reqIn = requiredThicknessIn(input);
    if (minRequired <= 0) minRequired = reqIn * 25.4;
    for (const pt of points) {
      const actual = numberValue(pt.actual_thickness_mm);
      if (actual < minActual) minActual = actual;
      let reqMM = numberValue(pt.required_thickness_mm);
      if (reqMM <= 0) reqMM = reqIn * 25.4;
      if (reqMM < minRequired) minRequired = reqMM;
    }
    if (minRequired <= 0) return 1.0;
    return minActual / minRequired;
  }

  function calculateSMYSUtilizationPctJS(input) {
    const smys = numberValue(input.smys_psi);
    const od = numberValue(input.outside_diameter_in);
    const points = input.inspection_points || [];
    if (smys <= 0 || od <= 0 || !points.length) return 0;
    let minActualIn = numberValue(points[0].actual_thickness_mm) / 25.4;
    for (const pt of points) {
      const actualIn = numberValue(pt.actual_thickness_mm) / 25.4;
      if (actualIn < minActualIn) minActualIn = actualIn;
    }
    if (minActualIn <= 0) return 0;
    return (numberValue(input.operating_pressure_psi) * od) / (2 * minActualIn * smys) * 100;
  }

  function parseMonthYearToFloatJS(val) {
    if (!val) return 0;
    const raw = String(val).trim();
    if (!raw) return 0;
    if (raw.includes("/")) {
      const parts = raw.split("/");
      const month = parseFloat(parts[0]);
      const year = parseFloat(parts[1]);
      if (isNaN(year)) return 0;
      if (isNaN(month)) return year;
      return year + (month - 1) / 12;
    }
    if (raw.includes("-")) {
      const parts = raw.split("-");
      const first = parseFloat(parts[0]);
      const second = parseFloat(parts[1]);
      if (isNaN(first)) return 0;
      if (isNaN(second)) return first;
      if (second > 100) return second + (first - 1) / 12;
      return first + (second - 1) / 12;
    }
    const num = parseFloat(raw);
    return isNaN(num) ? 0 : num;
  }

  function calculatePipelineRealtime(input) {
    const points = input.inspection_points || [];
    if (!points.length || numberValue(input.outside_diameter_in) <= 0 || numberValue(input.internal_design_pressure_psi) <= 0 || numberValue(input.smys_psi) <= 0) {
      return null;
    }
    const risk = riskInputOf(input);
    const requiredIn = requiredThicknessIn(input);
    const pointResults = points.map((point, index) => {
      const actualIn = numberValue(point.actual_thickness_mm) / 25.4;
      const cr = corrosionRate(point, input);
      const appraisalRequiredIn = index > 0 ? roundToPlaces(requiredIn, 3) : requiredIn;
      const requiredMM = numberValue(point.required_thickness_mm) > 0 ? numberValue(point.required_thickness_mm) : appraisalRequiredIn * 25.4;
      const rl = cr > 0
        ? capRemainingLife(Math.max((numberValue(point.actual_thickness_mm) - requiredMM) / cr, 0))
        : (numberValue(point.actual_thickness_mm) >= requiredMM ? 20 : 0);
      const hs = actualIn > 0 ? (numberValue(input.internal_design_pressure_psi) * numberValue(input.outside_diameter_in)) / (2 * actualIn) : 0;
      const maop = maopPsi(input, actualIn);
      const allowableStress = allowableStressPsi(input);
      return {
        inspection_point: point.inspection_point,
        location_class: point.location_class,
        installation_type: point.installation_type,
        measured_year: point.measured_year,
        nominal_thickness_mm: point.nominal_thickness_mm,
        required_thickness_mm: requiredIn * 25.4,
        minimum_thickness_mm: requiredMM,
        actual_thickness_mm: point.actual_thickness_mm,
        remaining_thickness_mm: Math.max(numberValue(point.actual_thickness_mm) - requiredMM, 0),
        corrosion_rate_mm_year: cr,
        remaining_life_years: rl,
        hoop_stress_psi: hs,
        maop_psi: maop,
        thickness_status: actualIn >= appraisalRequiredIn ? "ACCEPTABLE" : "NOT ACCEPTABLE",
        hoop_stress_status: hs <= allowableStress ? "ACCEPTABLE" : "NOT ACCEPTABLE",
        maop_status: maop > numberValue(input.internal_design_pressure_psi) ? "ACCEPTABLE" : "NOT ACCEPTABLE",
      };
    });

    // Auto-calculate partial pressures and SMYS utilization
    input.co2_partial_pressure_psig = calculateCO2PartialPressureJS(risk, input);
    input.h2s_partial_pressure_psig = calculateH2SPartialPressureJS(risk, input);
    input.wall_thickness_ratio = calculateWallThicknessRatioJS(input);
    input.smys_utilization_pct = calculateSMYSUtilizationPctJS(input);

    console.log(input)
    // Gate-Modifier-Escalation scoring
    const internalResult = scoreInternalCorrosionJS(risk, input);
    const externalResult = scoreExternalCorrosionJS(risk, input);
    const localizedResult = scoreLocalizedCorrosionJS(risk, input, internalResult);
    const erosionResult = scoreErosionJS(risk);
    const erosionCorrosionResult = scoreErosionCorrosionJS(risk, input, erosionResult);
    const crackingResult = scoreCrackingJS(risk, input);
    const sccResult = scoreSCCJS(risk, input);
    const fatigueResult = scoreFatigueJS(risk);
    const coatingResult = scoreCoatingDegradationJS(risk, input);
    const tpdResult = scoreThirdPartyDamageJS(risk);
    const chemicalResult = scoreChemicalDamageJS();

    const dmResults = [externalResult, coatingResult, tpdResult, internalResult, localizedResult, erosionResult, erosionCorrosionResult, crackingResult, sccResult, fatigueResult, chemicalResult].map(annotatePipelineDMResult);

    const governingDM = dmResults.reduce((best, dm) => dm.score > best.score ? dm : best, dmResults[0]);
    const governingScore = governingDM.score || 0;

    const fms = Math.pow(10, (-0.02 * ((numberValue(risk.management_system_score) / 1000) * 100)) + 1);
    const pofValue = numberValue(risk.generic_failure_frequency) * governingScore * fms;
    const pof = pofCategory(pofValue);
    const isGas = isGasService(input.service);
    const cofData = isGas ? gasCoF(input, risk) : liquidCoF(input, risk);
    const finalRiskCode = `${pof}${cofData.cof}`;
    const finalRiskLevel = matrixLevel(pof, cofData.cof);
    const inspectionPlanResults = calculateRealtimeInspectionPlan(dmResults);
    const groups = buildRealtimeRecommendationGroups(governingDM.label, finalRiskLevel, isGas);

    // Map DM scores to legacy DF fields (0 → 1.0 for backward compat)
    const dfTPD = tpdResult.score || 1.0;
    const dfExternal = externalResult.score || 1.0;
    const dfInternal = internalResult.score || 1.0;

    return {
      final_risk_code: finalRiskCode,
      final_risk_level: `${finalRiskLevel} Risk`,
      pof,
      cof: cofData.cof,
      pof_value: pofValue,
      cof_value: cofData.value,
      risk_value: Number(pof) * cofNumeric(cofData.cof),
      generic_failure_frequency: risk.generic_failure_frequency,
      management_system_factor: fms,
      third_party_damage_factor: dfTPD,
      external_corrosion_factor: dfExternal,
      internal_corrosion_factor: dfInternal,
      governing_damage_factor: governingScore,
      governing_damage_mechanism: governingDM.label,
      damage_mechanism_results: dmResults,
      inspection_plan_results: inspectionPlanResults,
      point_results: pointResults,
      pir_feet: cofData.pir || 0,
      spill_volume: cofData.spill || 0,
      adjusted_spill_volume: cofData.adjustedSpill || 0,
      recommendation_groups: groups,
      recommendation_source: "Realtime browser preview; backend recalculates on save.",
      recommendation_rule_name: "pipeline-js-preview-v2",
      recommendation_confidence: "Low",
      recommendation: [...groups.immediate_actions, ...groups.inspection_monitoring, ...groups.long_term_mitigation].join(" "),
    };
  }

  function calculateRealtimeInspectionPlan(dmResults) {
    const plans = collectInspectionPlanByDM();
    return dmResults.map((item) => {
      const plan = plans[item.code] || {};
      const nonMethod = plan.non_intrusive_method || defaultNonIntrusiveMethod(item.code);
      const intMethod = plan.intrusive_method || defaultIntrusiveMethod(item.code);
      const nonEff = item.severity === "NOT" ? "" : methodEffectivity(nonMethod);
      const intEff = item.severity === "NOT" ? "" : methodEffectivity(intMethod);
      item.inspection_effectivity = nonEff;
      return {
        code: item.code,
        label: item.label,
        severity: item.severity,
        non_intrusive_method: nonMethod,
        non_intrusive_effectivity: nonEff,
        non_intrusive_interval_months: inspectionIntervalMonths(item.severity, nonEff, false),
        intrusive_method: intMethod,
        intrusive_effectivity: intEff,
        intrusive_interval_months: inspectionIntervalMonths(item.severity, intEff, true),
        source: "Realtime browser preview; backend recalculates on save.",
      };
    });
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
    $("#savePipelineDraft").prop("disabled", isSavingAssessment);
  }

  function isManualTextField(name) {
    return [
      "recommendation_immediate_actions",
      "recommendation_inspection_monitoring",
      "recommendation_long_term_mitigation",
      "engineering_notes",
      "assessment_by",
    ].includes(name);
  }

  function calculationSignature(payload) {
    const clone = JSON.parse(JSON.stringify(payload));
    delete clone.assessment_by;
    delete clone.recommendation_immediate_actions;
    delete clone.recommendation_inspection_monitoring;
    delete clone.recommendation_long_term_mitigation;
    if (clone.risk_input) delete clone.risk_input.engineering_notes;
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
      ["generic_failure_frequency", "Generic Failure Frequency (GFF) must be a valid number."],
      ["management_system_score", "Management score must be a valid number."],
      ["base_tpd_rate", "Base third-party damage rate must be a valid number."],
      ["base_external_corr_rate", "Base external corrosion rate must be a valid number."],
      ["base_internal_corr_rate", "Base internal corrosion rate must be a valid number."],
      ["co2_content", "CO2 content must be zero or a valid number.", true],
      ["h2s_content", "H2S content (ppm) must be zero or a valid number.", true],
      ["h2o_content", "H2O content (lb/mmscf) must be zero or a valid number.", true],
      ["n2_content", "N2 content must be zero or a valid number.", true],
      ["co_content", "CO content must be zero or a valid number.", true],
      ["pressure_cycle_count", "Pressure cycle count must be zero or a valid number.", true],
      ["pressure_range_pct", "Pressure range must be zero or a valid number.", true],
    ];
    if (!isGasService($("[name='service']").val())) {
      numericFields.push(["flow_rate", "Flow rate must be a valid number."]);
      numericFields.push(["detection_time_hours", "Leak detection time must be a valid number."]);
      numericFields.push(["segment_length_between_valves_m", "Segment length between isolation valves must be a valid number."]);
    }

    numericFields.forEach(([name, message, allowZero]) => {
      const $field = $(`[name='${name}']`);
      if (!$field.length) return;
      const value = parseLocalizedNumber($field.val());
      if (!Number.isFinite(value) || (allowZero && value < 0) || (!allowZero && value <= 0)) {
        setFieldError($field, message);
        valid = false;
      }
    });

    const rowWidthRaw = String($("[name='right_of_way']").val() || "").trim();
    const rowWidth = parseLocalizedNumber(rowWidthRaw);
    if (rowWidthRaw && (rowWidthRaw.includes("-") || !Number.isFinite(rowWidth) || rowWidth <= 0)) {
      setFieldError($("[name='right_of_way']"), "Right of Way width must be a single positive number.");
      valid = false;
    }

    const chloride = parseInt($("input[name='chloride_content']").val(), 10);
    if (!Number.isFinite(chloride) || chloride < 0 || chloride > 5) {
      setFieldError($("input[name='chloride_content']"), "Chloride content must be between 0 and 5.");
      valid = false;
    }

    let validPointCount = 0;
    $("#pipelinePointsTable tbody tr").each(function () {
      const pointName = String($(this).find("[name='inspection_point']").val() || "").trim();
      const nominal = numberValue($(this).find("[name='point_nominal_thickness_mm']").val());
      const actual = numberValue($(this).find("[name='point_actual_thickness_mm']").val());
      const measuredValue = String($(this).find("[name='measured_year']").val() || "");
      if (!pointName || nominal <= 0 || actual <= 0 || !measuredValue) return;
      validPointCount += 1;
      const $measured = $(this).find("[name='measured_year']");
      const measured = parseMonthYearToFloatJS($measured.val());
      const used = parseMonthYearToFloatJS($("[name='year_used']").val() || $("[name='year_built']").val());
      if (measured <= used) {
        setFieldError($measured, "Measured year must be after year used.");
        valid = false;
      }
    });
    if (!validPointCount) {
      setFieldError($("#pipelinePointsTable tbody tr:first [name='inspection_point']"), "At least one complete inspection point row is required.");
      valid = false;
    }

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
    const isGas = isGasService($("[name='service']").val());
    $(".pipeline-gas-fields").toggleClass("d-none", !isGas);
    $(".pipeline-liquid-fields").toggleClass("d-none", isGas);
    $("[name='flow_rate']").closest(".col-md-4").toggleClass("d-none", isGas);
  }

  function updateReviewSummary() {
    const $summary = $("#pipelineReviewSummary");
    if (!$summary.length) return;
    const fluidType = $("[name='service']").val() || "-";
    const isGas = isGasService(fluidType);
    const items = [
      ["Tag Number", $("[name='report_no']").val()],
      ["Pipeline Name", $("[name='line_identification']").val()],
      ["Fluid Type", fluidType],
      ["Damage Mechanism", "All configured mechanisms screened"],
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
      const metadata = [item.source_standard, item.confidence_level, item.rule_status].filter(Boolean).join(" | ");
      return `<tr><td>${escapeHtml(item.formula_name || "-")}</td><td>${escapeHtml(item.excel_ref || "-")}</td><td class="text-wrap" style="min-width:320px;"><div>${escapeHtml(item.expression || "-")}</div><div class="text-muted small mt-1">${inputs}</div>${metadata ? `<div class="text-primary small mt-1">${escapeHtml(metadata)}</div>` : ""}</td><td>${escapeHtml(formatTraceValue(item.output))}</td></tr>`;
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
    const numeric = parseLocalizedNumber(value);
    return Number.isFinite(numeric) && value.trim() !== "" ? numeric : value;
  }

  function numberValue(value) {
    const numeric = parseLocalizedNumber(value);
    return Number.isFinite(numeric) ? numeric : 0;
  }

  function valueWithUnit(name, unit) {
    const value = $(`[name='${name}']`).val();
    return value ? `${fmt(value)} ${unit}` : "-";
  }

  function labelFor($select) {
    return $select.find("option:selected").text() || "-";
  }

  function codeForService(service) {
    const family = serviceFormulaFamily(service);
    if (family === "gas") return "ASME B31.8";
    if (family === "liquid") return "ASME B31.3";
    return "ASME B31.4";
  }

  function serviceFormulaFamily(service) {
    const key = String(service || "").trim().toLowerCase();
    if (["gas", "natural gas", "dry gas", "wet gas"].includes(key)) return "gas";
    if (["liquid", "piping", "produce water", "produced water", "liquid hydrocarbon", "chemical"].includes(key)) return "liquid";
    return "oil";
  }

  function isGasService(service) {
    return serviceFormulaFamily(service) === "gas";
  }

  function normalizeApplicableCode(code) {
    const text = String(code || "").toUpperCase();
    if (text.includes("B31.3")) return "ASME B31.3";
    if (text.includes("B31.8")) return "ASME B31.8";
    if (text.includes("B31.4")) return "ASME B31.4";
    return codeForService($("[name='service']").val());
  }

  function derivedMaterialStress(input) {
    const smys = numberValue(input.smys_psi);
    if (smys <= 0) return 0;
    const code = normalizeApplicableCode(input.applicable_code);
    if (code.includes("B31.3")) {
      const masterStress = selectedMaterialAllowableStress();
      return masterStress > 0 ? masterStress : 0;
    }
    return smys;
  }

  function selectedMaterialAllowableStress() {
    const selected = $("#materialSpecSelect option:selected");
    return numberValue(selected.data("allowable-stress"));
  }

  function applySelectedPipelineMaterial() {
    const selected = $("#materialSpecSelect option:selected");
    const smys = numberValue(selected.data("smys"));
    if (smys > 0) {
      $("[name='smys_psi']").val(fmt(smys));
    }
  }

  function syncApplicableCodeAndMaterialStress(forceServiceCode) {
    const $code = $("[name='applicable_code']");
    if (!$code.length) return;
    const serviceCode = codeForService($("[name='service']").val());
    const nextCode = forceServiceCode ? serviceCode : normalizeApplicableCode($code.val() || serviceCode);
    isProgrammaticUpdate = true;
    try {
      $code.val(nextCode);
      const materialStress = derivedMaterialStress({
        applicable_code: nextCode,
        smys_psi: $("[name='smys_psi']").val(),
      });
      $("[name='material_stress_psi']").val(materialStress > 0 ? fmt(materialStress) : "");
    } finally {
      isProgrammaticUpdate = false;
    }
    const source = nextCode.includes("B31.3")
      ? (selectedMaterialAllowableStress() > 0 ? "Allowable stress S is sourced from Pipeline Material Master." : "ENGINEERING_REVIEW_REQUIRED: select a Pipeline Material Master allowable stress.")
      : "Derived from SMYS for B31.4/B31.8 pipeline formulas.";
    $("#pipelineMaterialStressSource").text(source);
  }

  function collectInspectionEffectivityByDM() {
    const values = {};
    $(".pipeline-inspection-method[data-scope='nonintrusive']").each(function () {
      values[$(this).data("dm-code")] = methodEffectivity($(this).val());
    });
    return values;
  }

  function applySavedInspectionPlan() {
    const raw = $("#pipelineSavedInspectionPlan").text().trim();
    if (!raw || raw === "null") return;
    try {
      const savedPlan = JSON.parse(raw);
      Object.entries(savedPlan || {}).forEach(([code, plan]) => {
        if (!plan) return;
        $(`.pipeline-inspection-method[data-dm-code='${code}'][data-scope='nonintrusive']`).val(plan.non_intrusive_method || "None");
        $(`.pipeline-inspection-method[data-dm-code='${code}'][data-scope='intrusive']`).val(plan.intrusive_method || "None");
      });
    } catch (err) {
      // Keep default inspection methods when older records do not contain plan JSON.
    }
  }

  function collectInspectionPlanByDM() {
    const values = {};
    $(".pipeline-inspection-method").each(function () {
      const code = $(this).data("dm-code");
      const scope = $(this).data("scope");
      if (!values[code]) values[code] = {};
      if (scope === "intrusive") values[code].intrusive_method = $(this).val() || "None";
      if (scope === "nonintrusive") values[code].non_intrusive_method = $(this).val() || "None";
    });
    return values;
  }

  function governingDamageMechanismCode() {
    const risk = collectPayloadShallowRiskInput();
    const input = collectPayload(false);
    input.co2_partial_pressure_psig = calculateCO2PartialPressureJS(risk, input);
    input.h2s_partial_pressure_psig = calculateH2SPartialPressureJS(risk, input);
    input.wall_thickness_ratio = calculateWallThicknessRatioJS(input);
    input.smys_utilization_pct = calculateSMYSUtilizationPctJS(input);
    const internalResult = scoreInternalCorrosionJS(risk, input);
    const externalResult = scoreExternalCorrosionJS(risk, input);
    const localizedResult = scoreLocalizedCorrosionJS(risk, input, internalResult);
    const erosionResult = scoreErosionJS(risk);
    const erosionCorrosionResult = scoreErosionCorrosionJS(risk, input, erosionResult);
    const crackingResult = scoreCrackingJS(risk, input);
    const sccResult = scoreSCCJS(risk, input);
    const fatigueResult = scoreFatigueJS(risk);
    const coatingResult = scoreCoatingDegradationJS(risk, input);
    const tpdResult = scoreThirdPartyDamageJS(risk);
    const chemicalResult = scoreChemicalDamageJS();
    const dmResults = [externalResult, coatingResult, tpdResult, internalResult, localizedResult, erosionResult, erosionCorrosionResult, crackingResult, sccResult, fatigueResult, chemicalResult];
    const highest = dmResults.reduce((best, dm) => dm.score > best.score ? dm : best, dmResults[0]);
    return highest && highest.score > 0 ? highest.code : "internal_corrosion";
  }

  function collectPayloadShallowRiskInput() {
    return {
      depth_of_cover: $("[name='depth_of_cover']").val(),
      patrol_frequency: $("[name='patrol_frequency']").val(),
      row_condition: $("[name='row_condition']").val(),
      soil_resistivity: $("[name='soil_resistivity']").val(),
      coating_condition: $("[name='coating_condition']").val(),
      cp_status: $("[name='cp_status']").val(),
      cp_potential_mv: numberValue($("input[name='cp_potential_mv']").val()),
      coating_damage_level: selectedDamageLevel("coating_condition", "coating_damage_level"),
      one_call_system: $("[name='one_call_system']").val(),
      ph_level: $("[name='ph_level']").val(),
      fluid_corrosivity_mpy: $("[name='fluid_corrosivity_mpy']").val(),
      inhibitor_effectiveness: $("[name='inhibitor_effectiveness']").val(),
      biocide_treatment: $("[name='biocide_treatment']").val(),
      corrosion_monitoring_result: $("[name='corrosion_monitoring_result']").val(),
      h2s_ppm: $("[name='h2s_ppm']").val(),
      pwht_status: $("[name='pwht_status']").val(),
      weld_joint_type: $("[name='weld_joint_type']").val(),
      flow_velocity_condition: $("[name='flow_velocity_condition']").val(),
      solid_content: $("[name='solid_content']").val(),
      prev_ext_corrosion: $("[name='prev_ext_corrosion']").val(),
      prev_int_thinning: $("[name='prev_int_thinning']").val(),
      prev_int_cracking: $("[name='prev_int_cracking']").val(),
      prev_loc_int_corrosion: $("[name='prev_loc_int_corrosion']").val(),
      insulation_condition: $("[name='insulation_condition']").val(),
      ext_coating_condition: $("[name='ext_coating_condition']").val(),
      env_ext_cracking: $("[name='env_ext_cracking']").val(),
      conf_ext_corrosion: $("[name='conf_ext_corrosion']").val(),
      conf_int_thinning: $("[name='conf_int_thinning']").val(),
      conf_int_cracking: $("[name='conf_int_cracking']").val(),
      conf_loc_int_corrosion: $("[name='conf_loc_int_corrosion']").val(),
      co2_content: numberValue($("input[name='co2_content']").val()),
      h2s_content: numberValue($("input[name='h2s_content']").val()),
      h2o_content: numberValue($("input[name='h2o_content']").val()),
      chloride_content: parseInt($("input[name='chloride_content']").val(), 10) || 0,
      flow_rate: numberValue($("input[name='flow_rate']").val()),
      base_tpd_rate: numberValue($("[name='base_tpd_rate']").val()),
      base_external_corr_rate: numberValue($("[name='base_external_corr_rate']").val()),
      base_internal_corr_rate: numberValue($("[name='base_internal_corr_rate']").val()),
      management_system_score: numberValue($("[name='management_system_score']").val()),
      generic_failure_frequency: numberValue($("[name='generic_failure_frequency']").val()),
      operating_pressure_psi: numberValue($("input[name='operating_pressure_psi']").val()),
      coating_damage_level: $("[name='coating_damage_level']").val(),
      insulation_damage_level: selectedDamageLevel("insulation_condition", "insulation_damage_level"),
      ext_coating_damage_level: selectedDamageLevel("ext_coating_condition", "ext_coating_damage_level"),
      pressure_cycle_count: numberValue($("input[name='pressure_cycle_count']").val()),
      pressure_range_pct: numberValue($("input[name='pressure_range_pct']").val()),
    };
  }

  function requiredThicknessIn(input) {
    const p = numberValue(input.internal_design_pressure_psi);
    const d = numberValue(input.outside_diameter_in);
    const c = numberValue(input.allowance_in);
    const code = normalizeApplicableCode(input.applicable_code);
    if (code.includes("B31.3")) {
      const s = derivedMaterialStress(input);
      const e = numberValue(input.quality_factor) || 1;
      const w = numberValue(input.weld_joint_strength_factor) || 1;
      const y = numberValue(input.design_factor);
      return (p * d) / (2 * ((s * e * w) + (p * y))) + c;
    }
    if (code.includes("B31.8")) {
      const f = numberValue(input.design_factor) || 0.72;
      const e = numberValue(input.quality_factor) || 1;
      const tempFactor = numberValue(input.temperature_derating_factor) || 1;
      const smys = numberValue(input.smys_psi);
      return ((p * d) / (2 * f * e * tempFactor * smys)) + c;
    }
    const f = numberValue(input.design_factor) || 0.72;
    const e = numberValue(input.quality_factor) || 1;
    const smys = numberValue(input.smys_psi);
    return ((p * d) / (2 * f * e * smys)) + c;
  }

  function corrosionRate(point, input) {
    const nominal = numberValue(point.nominal_thickness_mm);
    const actual = numberValue(point.actual_thickness_mm);
    const yMeasured = parseMonthYearToFloatJS(String(point.measured_year || ""));
    const yUsed = parseMonthYearToFloatJS(String(input.year_used || input.year_built || ""));
    const diff = yMeasured - yUsed;
    if (diff <= 0 || nominal <= 0) return 0;
    return (nominal - actual) / diff;
  }

  function capRemainingLife(value) {
    return Math.min(numberValue(value), 20);
  }

  function roundToPlaces(value, places) {
    const factorValue = Math.pow(10, places);
    return Math.round(numberValue(value) * factorValue) / factorValue;
  }

  function allowableStressPsi(input) {
    const code = normalizeApplicableCode(input.applicable_code);
    if (code.includes("B31.3")) {
      return derivedMaterialStress(input) * (numberValue(input.quality_factor) || 1) * (numberValue(input.weld_joint_strength_factor) || 1);
    }
    return numberValue(input.smys_psi) * (numberValue(input.design_factor) || 0.72) * (numberValue(input.quality_factor) || 1) * (numberValue(input.temperature_derating_factor) || 1);
  }

  function maopPsi(input, actualIn) {
    if (actualIn <= 0 || numberValue(input.outside_diameter_in) <= 0) return 0;
    const code = normalizeApplicableCode(input.applicable_code);
    if (code.includes("B31.3")) {
      const s = derivedMaterialStress(input);
      const e = numberValue(input.quality_factor) || 1;
      const w = numberValue(input.weld_joint_strength_factor) || 1;
      const y = numberValue(input.design_factor);
      return (2 * s * e * w * actualIn) / (numberValue(input.outside_diameter_in) - (2 * y * actualIn));
    }
    if (code.includes("B31.8")) {
      return (2 * actualIn * (numberValue(input.smys_psi) * (numberValue(input.design_factor) || 0.72) * (numberValue(input.quality_factor) || 1) * (numberValue(input.temperature_derating_factor) || 1))) / numberValue(input.outside_diameter_in);
    }
    return (2 * actualIn * (numberValue(input.smys_psi) * (numberValue(input.design_factor) || 0.72) * (numberValue(input.quality_factor) || 1))) / numberValue(input.outside_diameter_in);
  }

  function factor(map, key) {
    return map[String(key || "").trim()] || map[String(key || "").trim().toLowerCase()] || 1;
  }

  function avg(...values) {
    const valid = values.filter((value) => Number.isFinite(Number(value)) && Number(value) > 0);
    return valid.length ? valid.reduce((sum, value) => sum + Number(value), 0) / valid.length : 0;
  }

  function severity(score) {
    if (score <= 0) return "NOT";
    if (score < 1.5) return "Low";
    if (score < 3) return "Moderate";
    return "High";
  }

  function severityBadgeClass(value) {
    if (value === "High") return "bg-label-danger";
    if (value === "Moderate") return "bg-label-warning";
    if (value === "Low") return "bg-label-success";
    return "bg-label-secondary";
  }

  function methodEffectivity(method) {
    const $option = $(".pipeline-inspection-method option").filter(function () {
      return $(this).val() === method;
    }).first();
    const configured = $option.data("effectivity");
    if (configured) return configured;
    const text = String(method || "").toLowerCase();
    if (!text || text === "none") return "None";
    if (text.includes("vie") || text.includes("direct") || text.includes("mpt") || text.includes("dpt")) return "High";
    if (text.includes("ut") || text.includes("ultrasonic") || text.includes("cp") || text.includes("coating")) return "Medium";
    return "Low";
  }

  function inspectionIntervalMonths(severityValue, effectivity, intrusive) {
    const base = { High: 12, Moderate: 24, Low: 48, NOT: 60 }[severityValue] || 36;
    const multiplier = { High: 1.25, Medium: 1, Low: 0.75, None: 0.5 }[effectivity] || 1;
    const months = Math.round(base * multiplier * (intrusive ? 2 : 1));
    return Math.min(Math.max(months, 6), 120);
  }

  function defaultNonIntrusiveMethod(code) {
    const $select = $(`.pipeline-inspection-method[data-dm-code='${code}'][data-scope='nonintrusive']`);
    return $select.val() || $select.find("option:first").val() || "None";
  }

  function defaultIntrusiveMethod(code) {
    const $select = $(`.pipeline-inspection-method[data-dm-code='${code}'][data-scope='intrusive']`);
    return $select.val() || $select.find("option:first").val() || "None";
  }

  function pofCategory(value) {
    if (value >= 0.01) return "5";
    if (value >= 0.001) return "4";
    if (value >= 0.0001) return "3";
    if (value >= 0.00001) return "2";
    return "1";
  }

  function cofNumeric(category) {
    return ({ A: 1, B: 2, C: 3, D: 4, E: 5 }[category] || 1);
  }

  function gasCoF(input, risk) {
    const pir = 0.69 * numberValue(input.outside_diameter_in) * Math.sqrt(numberValue(input.operating_pressure_psi));
    let cof = "A";
    const buildings = numberValue(risk.building_count_inside_pir);
    if (buildings > 20) cof = "E";
    else if (buildings >= 6) cof = "D";
    else if (buildings >= 1) cof = "B";
    if ((risk.class_location === "village" || risk.class_location === "class_3") && cofNumeric(cof) < 3) cof = "C";
    if ((risk.class_location === "urban_dense" || risk.class_location === "class_4") && cofNumeric(cof) < 4) cof = "D";
    return { cof, value: cofNumeric(cof), pir };
  }

  function liquidCoF(input, risk) {
    const diameterM = numberValue(input.outside_diameter_in) * 0.0254;
    const pipelineVolumeM3 = Math.PI * Math.pow(diameterM / 2, 2) * numberValue(risk.segment_length_between_valves_m);
    const spill = numberValue(risk.flow_rate) * numberValue(risk.detection_time_hours) + pipelineVolumeM3 * 6.28981077;
    let adjustedSpill = spill * factor({ low: 1, medium: 1.5, high: 2.5 }, risk.environmental_sensitivity);
    if (risk.nearby_sensitive_receptor) adjustedSpill *= 1.25;
    if (!risk.isolation_valve_available) adjustedSpill *= 1.25;
    let cof = "A";
    if (adjustedSpill > 1000) cof = "E";
    else if (adjustedSpill > 300) cof = "D";
    else if (adjustedSpill > 100) cof = "C";
    else if (adjustedSpill > 25) cof = "B";
    return { cof, value: adjustedSpill, spill, adjustedSpill };
  }

  function buildRealtimeRecommendationGroups(driver, level, isGas) {
    const groups = { immediate_actions: [], inspection_monitoring: ["Keep the formula trace with the assessment record."], long_term_mitigation: ["Update the assessment after mitigation or inspection results are available."] };
    if (driver === "Third-Party / Mechanical Damage") groups.immediate_actions.push("Improve route markers and warning signs.", "Strengthen excavation permit control.");
    if (driver === "External Corrosion") groups.immediate_actions.push("Verify cathodic protection performance.", "Prioritize coating defect checks.");
    if (driver === "Internal Corrosion") groups.immediate_actions.push("Review inhibitor condition, fluid corrosivity, and water handling.");
    if (level === "Critical") groups.immediate_actions.push("Escalate to engineering review before continued operation.");
    if (level === "High") groups.immediate_actions.push("Assign mitigation owner and target date.");
    groups.long_term_mitigation.unshift(isGas ? "Review class location, public awareness, emergency response, and populated-area protection." : "Improve leak detection, isolation time, spill containment, and drainage/river protection.");
    return groups;
  }

  function selectedDamageMechanismLabel() {
    const $selected = $("input[name='damage_mechanism']:checked");
    if (!$selected.length) return "-";
    return $(`label[for='${$selected.attr("id")}']`).text() || $selected.val();
  }

  function fmt(value) {
    const numeric = parseLocalizedNumber(value);
    if (!Number.isFinite(numeric)) return value || "-";
    return numeric.toLocaleString("id-ID", { useGrouping: false, minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }

  function fmtRate(value) {
    const numeric = parseLocalizedNumber(value);
    if (!Number.isFinite(numeric)) return value || "-";
    return numeric.toLocaleString("id-ID", { useGrouping: false, minimumFractionDigits: 4, maximumFractionDigits: 4 });
  }

  function fmtEngineering(value) {
    const numeric = parseLocalizedNumber(value);
    if (!Number.isFinite(numeric)) return value || "-";
    const abs = Math.abs(numeric);
    const digits = abs > 0 && abs < 1 ? 8 : 4;
    return numeric.toLocaleString("id-ID", { useGrouping: false, minimumFractionDigits: 0, maximumFractionDigits: digits });
  }

  function parseLocalizedNumber(value) {
    if (typeof value === "number") return value;
    const raw = String(value ?? "").trim();
    if (!raw) return NaN;
    if (raw.includes(",")) return Number(raw.replace(/\./g, "").replace(",", "."));
    return Number(raw);
  }

  function localizeDecimalInputs() {
    $form.find("input[inputmode='decimal']").each(function () {
      formatDecimalInput($(this));
    });
  }

  function formatDecimalInput($field) {
    const raw = String($field.val() ?? "").trim();
    if (!raw) return;
    const numeric = parseLocalizedNumber(raw);
    if (Number.isFinite(numeric)) {
      $field.val(fmt(numeric));
    }
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
