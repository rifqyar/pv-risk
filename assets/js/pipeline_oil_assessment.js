$(function () {
  const $form = $("#pipelineOilForm");
  let pipelineStepper = null;
  const stepperEl = document.querySelector("#wizard-pipeline-assessment");
  if (stepperEl && typeof Stepper !== "undefined") {
    pipelineStepper = new Stepper(stepperEl, { linear: false });
  }

  $(".pipeline-step-next").on("click", function () {
    if (pipelineStepper) pipelineStepper.next();
  });

  $(".pipeline-step-prev").on("click", function () {
    if (pipelineStepper) pipelineStepper.previous();
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
  });

  $(document).on("click", ".remove-point", function () {
    $(this).closest("tr").remove();
  });

  $("#savePipelineDraft").on("click", function () {
    savePipeline(false);
  });

  $("#calculatePipelineOil").on("click", function () {
    savePipeline(true);
  });

  function savePipeline(calculate) {
    const id = $form.data("assessment-id");
    const mode = $form.data("mode");
    const payload = collectPayload();
    let url = "/assessment-pipeline/submit";
    if (calculate && id) url = `/assessment-pipeline/calculate/${id}`;
    else if (mode === "edit" && id) url = `/assessment-pipeline/update/${id}`;

    Swal.fire({ title: calculate ? "Calculating..." : "Saving...", allowOutsideClick: false, showConfirmButton: false });
    Swal.showLoading();

    fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    })
      .then((r) => r.json())
      .then((res) => {
        if (res.status !== "success") {
          Swal.fire("Failed", res.message || "Pipeline Oil assessment failed.", "error");
          return;
        }
        if (calculate && !id && res.id) {
          calculateExisting(res.id, payload);
          return;
        }
        const target = calculate ? `/assessment-pipeline/view/${res.id}` : `/assessment-pipeline/edit/${res.id}`;
        Swal.fire("Success", res.message, "success").then(() => {
          window.location.href = target;
        });
      })
      .catch(() => Swal.fire("Error", "Failed to connect to server.", "error"));
  }

  function calculateExisting(id, payload) {
    fetch(`/assessment-pipeline/calculate/${id}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    })
      .then((r) => r.json())
      .then((res) => {
        if (res.status !== "success") {
          Swal.fire("Draft Saved, Calculation Failed", res.message || "Pipeline Oil calculation failed.", "error").then(() => {
            window.location.href = `/assessment-pipeline/edit/${id}`;
          });
          return;
        }
        Swal.fire("Success", res.message, "success").then(() => {
          window.location.href = `/assessment-pipeline/view/${id}`;
        });
      })
      .catch(() => Swal.fire("Error", "Draft saved, but calculation request failed.", "error"));
  }

  function collectPayload() {
    const data = {};
    $form.serializeArray().forEach((item) => {
      if (item.name.startsWith("point_") || ["inspection_point", "location_class", "installation_type", "measured_year"].includes(item.name)) return;
      data[item.name] = numericOrString(item.value);
    });

    data.rbi = {
      damage_mechanism: $("input[name='damage_mechanism']").val(),
      inspection_effectivity: $("input[name='inspection_effectivity']").val(),
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
    delete data.damage_mechanism;
    delete data.inspection_effectivity;
    delete data.generic_failure_frequency;
    delete data.management_system_score;
    delete data.base_tpd_rate;
    delete data.base_external_corr_rate;
    delete data.base_internal_corr_rate;
    delete data.depth_of_cover;
    delete data.patrol_frequency;
    delete data.row_condition;
    delete data.soil_resistivity;
    delete data.coating_condition;
    delete data.cp_status;
    delete data.cp_potential_mv;
    delete data.fluid_corrosivity;
    delete data.water_content;
    delete data.co2_h2s_presence;
    delete data.mic_risk;
    delete data.wall_thickness_condition;
    delete data.building_count_inside_pir;
    delete data.class_location;
    delete data.emergency_response;
    delete data.flow_rate;
    delete data.detection_time_hours;
    delete data.segment_length_between_valves_m;
    delete data.environmental_sensitivity;
    delete data.nearby_sensitive_receptor;
    delete data.isolation_valve_available;
    delete data.engineering_notes;

    data.inspection_points = [];
    $("#pipelinePointsTable tbody tr").each(function () {
      data.inspection_points.push({
        inspection_point: $(this).find("[name='inspection_point']").val(),
        location_class: $(this).find("[name='location_class']").val(),
        installation_type: $(this).find("[name='installation_type']").val(),
        nominal_thickness_mm: numberValue($(this).find("[name='point_nominal_thickness_mm']").val()),
        actual_thickness_mm: numberValue($(this).find("[name='point_actual_thickness_mm']").val()),
        measured_year: parseInt($(this).find("[name='measured_year']").val(), 10) || 0,
      });
    });
    return data;
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
});
