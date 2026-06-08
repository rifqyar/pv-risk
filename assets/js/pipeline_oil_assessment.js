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
      release_fluid: "Oil",
      consequence_basis: "TODO_ENGINEERING_CONFIRMATION",
      probability_basis: "TODO_ENGINEERING_CONFIRMATION",
      engineering_notes: $("textarea[name='engineering_notes']").val(),
      requires_confirmation: true,
      confirmation_todo_reason: "TODO_ENGINEERING_CONFIRMATION: RBI PoF/CoF/risk ranking formulas are not present in Excel.",
    };
    delete data.damage_mechanism;
    delete data.inspection_effectivity;
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
