$(function () {
  const $form = $("#pipelineOilForm");
  let pipelineStepper = null;
  let currentResult = null;
  let resultSignature = "";
  let resultIsStale = false;

  localizeDecimalInputs();

  const stepperEl = document.querySelector("#wizard-pipeline-assessment");
  if (stepperEl && typeof Stepper !== "undefined") {
    pipelineStepper = new Stepper(stepperEl, { linear: false });
  }

  syncApplicableCodeAndMaterialStress(false);
  updateConsequenceFields();
  applySavedInspectionPlan();
  updateReviewSummary();
  loadSavedResult();
  runRealtimePipelineCalculation();
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
    if (this.name === "service") {
      syncApplicableCodeAndMaterialStress(true);
    } else if (this.name === "applicable_code" || this.name === "smys_psi") {
      syncApplicableCodeAndMaterialStress(false);
    }
    updateConsequenceFields();
    updateReviewSummary();
    runRealtimePipelineCalculation();
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
        <td><input class="form-control" name="installation_type"></td>
        <td><input type="text" inputmode="decimal" class="form-control" name="point_nominal_thickness_mm"></td>
        <td><input type="text" inputmode="decimal" class="form-control" name="point_actual_thickness_mm"></td>
        <td><input type="number" class="form-control" name="measured_year" value="${new Date().getFullYear()}"></td>
        <td><button type="button" class="btn btn-sm btn-icon btn-outline-danger remove-point"><i class="mdi mdi-trash-can-outline"></i></button></td>
      </tr>
    `);
    runRealtimePipelineCalculation();
    markResultStaleIfNeeded();
  });

  $(document).on("click", ".remove-point", function () {
    $(this).closest("tr").remove();
    runRealtimePipelineCalculation();
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
    data.applicable_code = normalizeApplicableCode(data.applicable_code || codeForService(data.service));
    data.material_stress_psi = derivedMaterialStress(data);

    data.RiskInput = {
      damage_mechanism: governingDamageMechanismCode(),
      inspection_effectivity: $("[name='inspection_effectivity']").val() || "Representative",
      inspection_effectivity_by_damage_mechanism: collectInspectionEffectivityByDM(),
      inspection_plan_by_damage_mechanism: collectInspectionPlanByDM(),
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
    Object.keys(data).filter((key) => key.startsWith("inspection_effectivity_") || key.startsWith("inspection_nonintrusive_") || key.startsWith("inspection_intrusive_")).forEach((key) => delete data[key]);

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
    $("#pipelineRealtimeThicknessResult").html(buildRealtimeThicknessResult(result.point_results || []));
    renderDamageMechanismResults(result.damage_mechanism_results || []);
    renderInspectionPlanResults(result.inspection_plan_results || []);
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
    $("#pipelineRecommendationText").html(`<strong>Source:</strong> ${escapeHtml(result.recommendation_source || "TODO_ENGINEERING_CONFIRMATION")}<br><strong>Advisory:</strong> ${escapeHtml(result.recommendation || "-")}`);
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
        <td>${escapeHtml(fmt(point.corrosion_rate_mm_year))}</td>
        <td>${escapeHtml(fmt(point.remaining_life_years))}</td>
        <td>${escapeHtml(fmt(point.hoop_stress_psi))}</td>
        <td>${escapeHtml(fmt(point.maop_psi))}</td>
        <td>${escapeHtml([point.thickness_status, point.maop_status].filter(Boolean).join(" / ") || "-")}</td>
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
        <td>${escapeHtml(fmt(point.corrosion_rate_mm_year))} mm/year</td>
        <td>${escapeHtml(fmt(point.remaining_life_years))} years</td>
        <td>${escapeHtml(fmt(point.hoop_stress_psi))} psi</td>
        <td>${escapeHtml(fmt(point.maop_psi))} psi</td>
        <td>${escapeHtml([point.thickness_status, point.hoop_stress_status, point.maop_status].filter(Boolean).join(" / ") || "-")}</td>
      </tr>
    `).join("");
  }

  function renderDamageMechanismResults(results) {
    results.forEach((item) => {
      const badgeClass = severityBadgeClass(item.severity);
      $(`.pipeline-dm-badge[data-dm-code='${item.code}']`)
        .attr("class", `badge rounded-pill px-3 pipeline-dm-badge ${badgeClass}`)
        .text(item.severity || "NOT");
      $(`.pipeline-dm-effectivity-badge[data-dm-code='${item.code}']`)
        .attr("class", `badge pipeline-dm-effectivity-badge ${badgeClass}`)
        .text(`${item.severity || "NOT"} / ${item.inspection_effectivity || "Medium"}`);
    });
  }

  function renderInspectionPlanResults(results) {
    results.forEach((item) => {
      $(`.pipeline-inspection-period[data-dm-code='${item.code}'][data-scope='nonintrusive']`).text(`${item.non_intrusive_interval_months || "-"} months`);
      $(`.pipeline-inspection-period[data-dm-code='${item.code}'][data-scope='intrusive']`).text(`${item.intrusive_interval_months || "-"} months`);
      $(`.pipeline-inspection-effectivity[data-dm-code='${item.code}'][data-scope='nonintrusive']`).text(`Effectivity: ${item.non_intrusive_effectivity || "-"}`);
      $(`.pipeline-inspection-effectivity[data-dm-code='${item.code}'][data-scope='intrusive']`).text(`Effectivity: ${item.intrusive_effectivity || "-"}`);
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
    const governing = valid.sort((a, b) => Number(a.remaining_life_years) - Number(b.remaining_life_years))[0];
    const minActual = Math.min(...points.map((point) => numberValue(point.actual_thickness_mm)).filter((value) => value > 0));
    const highestHoopStress = Math.max(...points.map((point) => numberValue(point.hoop_stress_psi)).filter((value) => value > 0));
    const lowestMAOP = Math.min(...points.map((point) => numberValue(point.maop_psi)).filter((value) => value > 0));
    $("#pipelineSummaryRemainingLife").text(governing ? `${fmt(governing.remaining_life_years)} years` : "-");
    $("#pipelineSummaryRemainingPoint").text(governing ? `Point: ${governing.inspection_point || "-"}` : "Point: -");
    $("#pipelineSummaryMinActual").text(Number.isFinite(minActual) ? `${fmt(minActual)} mm` : "-");
    $("#pipelineSummaryHoopStress").text(Number.isFinite(highestHoopStress) ? `${fmt(highestHoopStress)} psi` : "-");
    $("#pipelineSummaryLowestMAOP").text(Number.isFinite(lowestMAOP) ? `${fmt(lowestMAOP)} psi` : "-");
  }

  function runRealtimePipelineCalculation() {
    const payload = collectPayload();
    const result = calculatePipelineRealtime(payload);
    if (!result) return;
    currentResult = result;
    resultSignature = calculationSignature(payload);
    resultIsStale = false;
    renderCalculationResult(result);
    updateSaveState();
  }

  function calculatePipelineRealtime(input) {
    const points = input.inspection_points || [];
    if (!points.length || numberValue(input.outside_diameter_in) <= 0 || numberValue(input.internal_design_pressure_psi) <= 0 || numberValue(input.smys_psi) <= 0) {
      return null;
    }
    const risk = input.RiskInput || {};
    const requiredIn = requiredThicknessIn(input);
    const pointResults = points.map((point) => {
      const actualIn = numberValue(point.actual_thickness_mm) / 25.4;
      const cr = corrosionRate(point, input);
      const requiredMM = numberValue(point.required_thickness_mm) > 0 ? numberValue(point.required_thickness_mm) : requiredIn * 25.4;
      const rl = cr > 0 ? capRemainingLife(Math.max((numberValue(point.actual_thickness_mm) - requiredMM) / cr, 0)) : 20;
      const hs = actualIn > 0 ? (numberValue(input.internal_design_pressure_psi) * numberValue(input.outside_diameter_in)) / (2 * actualIn) : 0;
      const maop = maopPsi(input, actualIn);
      return {
        inspection_point: point.inspection_point,
        nominal_thickness_mm: point.nominal_thickness_mm,
        required_thickness_mm: requiredMM,
        minimum_thickness_mm: requiredMM,
        actual_thickness_mm: point.actual_thickness_mm,
        remaining_thickness_mm: Math.max(numberValue(point.actual_thickness_mm) - requiredMM, 0),
        corrosion_rate_mm_year: cr,
        remaining_life_years: rl,
        hoop_stress_psi: hs,
        maop_psi: maop,
        thickness_status: actualIn > requiredIn ? "ACCEPTABLE" : "NOT ACCEPTABLE",
        hoop_stress_status: hs < numberValue(input.smys_psi) ? "ACCEPTABLE" : "NOT ACCEPTABLE",
        maop_status: maop > numberValue(input.internal_design_pressure_psi) ? "ACCEPTABLE" : "NOT ACCEPTABLE",
      };
    });
    const dfTPD = numberValue(risk.base_tpd_rate) / (factor({ "<1m": 0.8, "1-2m": 1.2, ">2m": 1.8 }, risk.depth_of_cover) * factor({ rare: 0.8, monthly: 1.2, weekly_daily: 1.8 }, risk.patrol_frequency) * factor({ poor: 0.8, fair: 1.2, good: 1.8 }, risk.row_condition));
    let cpFactor = factor({ failed: 3, borderline: 1.8, normal: 1 }, risk.cp_status);
    if (numberValue(risk.cp_potential_mv) > -850 && cpFactor < 1.8) cpFactor = 1.8;
    const dfExternal = numberValue(risk.base_external_corr_rate) * factor({ "<1000": 3, "1000-5000": 1.8, ">5000": 1 }, risk.soil_resistivity) * factor({ poor: 3, fair: 1.8, good: 1 }, risk.coating_condition) * cpFactor;
    const internalMap = { low: 1, none: 1, healthy: 1, medium: 1.9, present: 1.9, warning: 1.9, high: 3.5, critical: 3.5 };
    const dfInternal = numberValue(risk.base_internal_corr_rate) * factor(internalMap, risk.fluid_corrosivity) * factor(internalMap, risk.water_content) * factor(internalMap, risk.co2_h2s_presence) * factor(internalMap, risk.mic_risk) * factor(internalMap, risk.wall_thickness_condition);
    const governing = [
      { label: "Third-Party Damage", code: "third_party_mechanical_damage", value: dfTPD },
      { label: "External Corrosion", code: "external_corrosion", value: dfExternal },
      { label: "Internal Corrosion", code: "internal_corrosion", value: dfInternal },
    ].sort((a, b) => b.value - a.value)[0];
    const fms = Math.pow(10, (-0.02 * ((numberValue(risk.management_system_score) / 1000) * 100)) + 1);
    const pofValue = numberValue(risk.generic_failure_frequency) * governing.value * fms;
    const pof = pofCategory(pofValue);
    const isGas = isGasService(input.service);
    const cofData = isGas ? gasCoF(input, risk) : liquidCoF(input, risk);
    const finalRiskCode = `${pof}${cofData.cof}`;
    const finalRiskLevel = matrixLevel(pof, cofData.cof);
    const dmResults = calculateRealtimeDamageMechanisms(risk, dfTPD, dfExternal, dfInternal, pointResults);
    const inspectionPlanResults = calculateRealtimeInspectionPlan(dmResults);
    const groups = buildRealtimeRecommendationGroups(governing.label, finalRiskLevel, isGas);
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
      governing_damage_factor: governing.value,
      governing_damage_mechanism: governing.label,
      damage_mechanism_results: dmResults,
      inspection_plan_results: inspectionPlanResults,
      point_results: pointResults,
      pir_feet: cofData.pir || 0,
      spill_volume: cofData.spill || 0,
      adjusted_spill_volume: cofData.adjustedSpill || 0,
      recommendation_groups: groups,
      recommendation_source: "Realtime browser preview; backend recalculates on save.",
      recommendation_rule_name: "pipeline-js-preview-v1 TODO_ENGINEERING_CONFIRMATION",
      recommendation: [...groups.immediate_actions, ...groups.inspection_monitoring, ...groups.long_term_mitigation].join(" "),
    };
  }

  function calculateRealtimeDamageMechanisms(risk, dfTPD, dfExternal, dfInternal, points) {
    const remainingLives = points.map((p) => numberValue(p.remaining_life_years)).filter((v) => v > 0);
    const rl = remainingLives.length ? Math.min(...remainingLives) : 20;
    const rlScore = rl < 2 ? 3.5 : rl < 5 ? 2.2 : rl < 10 ? 1.2 : 0.8;
    const flowScore = numberValue(risk.flow_rate) >= 1000 ? 3.5 : numberValue(risk.flow_rate) >= 500 ? 2.2 : numberValue(risk.flow_rate) >= 100 ? 1.2 : 0;
    const internalMap = { low: 1, none: 1, healthy: 1, medium: 1.9, present: 1.9, warning: 1.9, high: 3.5, critical: 3.5 };
    const effectivity = collectInspectionEffectivityByDM();
    const definitions = [
      ["external_corrosion", "External Corrosion", "External Damage", dfExternal],
      ["coating_cui_degradation", "Coating / CUI Degradation", "External Damage", avg(factor({ poor: 3, fair: 1.8, good: 1 }, risk.coating_condition), factor({ "<1000": 3, "1000-5000": 1.8, ">5000": 1 }, risk.soil_resistivity), factor({ failed: 3, borderline: 1.8, normal: 1 }, risk.cp_status))],
      ["third_party_mechanical_damage", "Third-Party / Mechanical Damage", "External Damage", dfTPD],
      ["internal_corrosion", "Internal Corrosion", "Internal Thinning", dfInternal],
      ["localized_corrosion_pitting", "Localized Corrosion / Pitting", "Internal Thinning", avg(dfInternal, rlScore, factor(internalMap, risk.wall_thickness_condition))],
      ["erosion_corrosion", "Erosion / Erosion-Corrosion", "Internal Thinning", avg(flowScore, factor(internalMap, risk.fluid_corrosivity), factor(internalMap, risk.wall_thickness_condition))],
      ["cracking_scc_fatigue", "Cracking / SCC / Fatigue", "Internal Cracking", avg(factor(internalMap, risk.co2_h2s_presence), factor(internalMap, risk.mic_risk), factor({ failed: 3.5, borderline: 1.9, normal: 1 }, risk.cp_status))],
      ["other_engineering_review", "Other / Engineering Review", "Internal Cracking", 0],
    ];
    return definitions.map(([code, label, category, score]) => ({ code, label, category, score, severity: severity(score), inspection_effectivity: effectivity[code] || "Medium" }));
  }

  function calculateRealtimeInspectionPlan(dmResults) {
    const plans = collectInspectionPlanByDM();
    return dmResults.map((item) => {
      const plan = plans[item.code] || {};
      const nonMethod = plan.non_intrusive_method || defaultNonIntrusiveMethod(item.code);
      const intMethod = plan.intrusive_method || defaultIntrusiveMethod(item.code);
      const nonEff = methodEffectivity(nonMethod);
      const intEff = methodEffectivity(intMethod);
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
      ["generic_failure_frequency", "Generic Failure Frequency (GFF) must be a valid number."],
      ["management_system_score", "Management score must be a valid number."],
      ["base_tpd_rate", "Base third-party damage rate must be a valid number."],
      ["base_external_corr_rate", "Base external corrosion rate must be a valid number."],
      ["base_internal_corr_rate", "Base internal corrosion rate must be a valid number."],
    ];
    if (!isGasService($("[name='service']").val())) {
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
    if (["gas", "natural gas", "dwr gas", "wet gas"].includes(key)) return "gas";
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
    if (code.includes("B31.3")) return Math.min((smys * 2) / 3, 20000);
    return smys;
  }

  function syncApplicableCodeAndMaterialStress(forceServiceCode) {
    const $code = $("[name='applicable_code']");
    if (!$code.length) return;
    const serviceCode = codeForService($("[name='service']").val());
    const nextCode = forceServiceCode ? serviceCode : normalizeApplicableCode($code.val() || serviceCode);
    $code.val(nextCode);
    const materialStress = derivedMaterialStress({
      applicable_code: nextCode,
      smys_psi: $("[name='smys_psi']").val(),
    });
    console.log("Derived material stress:", materialStress);
    $("[name='material_stress_psi']").val(materialStress > 0 ? fmt(materialStress) : "");
    const source = nextCode.includes("B31.3")
      ? "Derived as min(2/3 x SMYS, 20,000 psi) for B31.3 allowable stress."
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
    const results = calculateRealtimeDamageMechanisms(collectPayloadShallowRiskInput(), 0, 0, 0, []);
    const highest = results.sort((a, b) => (b.score || 0) - (a.score || 0))[0];
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
      fluid_corrosivity: $("[name='fluid_corrosivity']").val(),
      water_content: $("[name='water_content']").val(),
      co2_h2s_presence: $("[name='co2_h2s_presence']").val(),
      mic_risk: $("[name='mic_risk']").val(),
      wall_thickness_condition: $("[name='wall_thickness_condition']").val(),
      flow_rate: numberValue($("input[name='flow_rate']").val()),
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
    const years = numberValue(point.measured_year) - numberValue(input.year_used);
    if (years <= 0) return 0;
    return Math.max((numberValue(point.nominal_thickness_mm) - numberValue(point.actual_thickness_mm)) / years, 0);
  }

  function capRemainingLife(value) {
    return Math.min(numberValue(value), 20);
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
    if (["external_corrosion", "coating_cui_degradation"].includes(code)) return "Visual + CP / Coating Survey";
    if (code === "third_party_mechanical_damage") return "ROW Patrol + Visual Survey";
    if (code === "cracking_scc_fatigue") return "Shear Wave Ultrasonic Testing";
    return "Wall Thickness measurement by UT";
  }

  function defaultIntrusiveMethod(code) {
    if (code === "cracking_scc_fatigue") return "Wet Fluorescent MPT / DPT";
    if (code === "third_party_mechanical_damage") return "Direct Examination";
    return "VIE + Wall Thickness measurement by UT";
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
    if (risk.class_location === "village" && cofNumeric(cof) < 3) cof = "C";
    if (risk.class_location === "urban_dense" && cofNumeric(cof) < 4) cof = "D";
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
    if (driver === "Third-Party Damage") groups.immediate_actions.push("Improve route markers and warning signs.", "Strengthen excavation permit control.");
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
    if (Math.abs(numeric) > 0 && Math.abs(numeric) < 0.001) return numeric.toExponential(3).replace(".", ",");
    return numeric.toLocaleString("id-ID", { useGrouping: false, maximumFractionDigits: 2 });
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
