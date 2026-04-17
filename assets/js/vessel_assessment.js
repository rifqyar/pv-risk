/**
 * Vessel Assessment Form Controller
 * Handles Stepper navigation, real-time calculations, and UI updates.
 */
const INSPECTION_DB = {
  thinning_internal: [
    { method: "UT", effectiveness: 3, level: "High" },
    { method: "UT_advanced", effectiveness: 2, level: "Medium" },
    { method: "RT", effectiveness: 1, level: "Low" },
  ],

  thinning_external: [
    { method: "visual_ut", effectiveness: 3, level: "High" },
    { method: "ut", effectiveness: 2, level: "Medium" },
    { method: "rt", effectiveness: 1, level: "Low" },
  ],

  cracking_internal: [
    { method: "swut", effectiveness: 3, level: "High" },
    { method: "rt", effectiveness: 2, level: "Medium" },
    { method: "visual", effectiveness: 1, level: "Low" },
  ],

  cracking_external: [
    { method: "wfmp", effectiveness: 3, level: "High" },
    { method: "swut", effectiveness: 2, level: "Medium" },
    { method: "rt", effectiveness: 1, level: "Low" },
  ],
};

$(function () {
  // ==========================================
  // 1. INITIALIZATION & CONFIG
  // ==========================================

  $("input[name='cathodic_protection']:checked").trigger("change");

  // Parse master data from Golang template (injected in main HTML)
  if (window.VESSEL_APP && window.VESSEL_APP.cisccMasterJSON) {
    localStorage.setItem(
      "master_ciscc",
      JSON.stringify(window.VESSEL_APP.cisccMasterJSON),
    );
  }

  // Initialize Bootstrap Stepper
  const stepperEl = document.querySelector("#wizard-vessel-assessment");
  const stepper = new Stepper(stepperEl, { linear: false });

  // Cache frequent DOM elements
  const elements = {
    eqType: $("#select2_equipment"),
    outPressStep3: $("#step3_op_pressure"),
    outTempStep3: $("#step3_op_temperature"),
    headType: $("select[name='type_head']"),
  };

  // Initialize default states
  initDefaultStates();

  // ==========================================
  // 2. EVENT BINDINGS
  // ==========================================

  // Stepper Controls
  $(".btn-next").on("click", () => stepper.next());
  $(".btn-prev").on("click", () => stepper.previous());

  // Global Calculation Trigger
  $(document).on("input change", "input, select", function () {
    calculateAndStore();
    runStep2Calculations();
  });

  if ($(".select2").length) {
    $(".select2").on("select2:select select2:unselect", function () {
      calculateAndStore();
    });

    // Equipment Type Selection
    elements.eqType.on(
      "select2:select select2:unselect",
      handleEquipmentTypeChange,
    );
  }

  // Head Type Selection
  elements.headType.on("change", handleHeadTypeChange);

  // Insulation & Coating Toggles
  $("#insulation_condition").on("change", function () {
    toggleDamageLevel("#insulation_level_wrapper", $(this).val());
  });

  $("#ext_coating_condition").on("change", function () {
    toggleDamageLevel("#ext_coating_level_wrapper", $(this).val());
  });

  // Step 3 Triggers (Syncing values)
  $(
    "input[name*='operating_press'], input[name*='operating_temp'], select[name='suhu_opr'], select[name='suhu_opr_top'], select[name='suhu_opr_bottom']",
  ).on("keyup change", syncOperatingConditions);

  $("input[name='he_side']").on("change", syncOperatingConditions);
  elements.eqType.on("change", syncOperatingConditions);

  $("select[name='description'], input[name='tag_number']").on(
    "keyup change",
    generateEquipmentName,
  );

  $("#select2_ph_contents, #select2_h2s_contents").on("change", function () {
    calculateEnvironmentalSeverity();
    calculateDamageMechanisms();
  });

  // Step 4 Triggers (Damage Mechanisms)
  $("#step-3 input, #step-3 select, #step-3 textarea").on(
    "change keyup",
    calculateDamageMechanisms,
  );

  $("input[name='he_side']").on("change", function () {
    calculateDamageMechanisms;
  });

  // Step 5 Trigger (Risk Assessments)
  $("#lof_category, #cof_financial, #cof_safety").on(
    "change",
    calculateCriticalityMatrix,
  );

  // Step 5 Trigger (Dropdown Effectivity Update)
  $("select[id^='insp_'], input[type='checkbox'][id^='opt_']").on(
    "change",
    function () {
      // --- 1. Logika Warna Teks Effectivity (Hanya jalan kalau yang diganti itu dropdown) ---
      if ($(this).is("select")) {
        let val = $(this).val();
        let effectivity = $(this).find("option:selected").data("eff") || "None";
        let $effSpan = $(this).closest("td").find("span[id^='eff_']");

        if (val === "None" || val === "" || effectivity === "None") {
          $effSpan
            .text("Effectivity: None")
            .removeClass("text-danger text-warning text-success fw-medium")
            .addClass("text-secondary");
        } else {
          let colorClass = "text-danger";
          if (effectivity === "Medium") colorClass = "text-warning";
          if (effectivity === "Low") colorClass = "text-dark";

          $effSpan
            .text("Effectivity: " + effectivity)
            .removeClass(
              "text-secondary text-danger text-warning text-dark text-success",
            )
            .addClass(colorClass + " fw-medium");
        }
      }

      // --- 2. Panggil fungsi hitung bulan interval ---
      calculateInspectionPeriods();

      // --- 3. Trigger kalkulasi ulang DM supaya Final Risk-nya otomatis update ---
      if (typeof calculateDamageMechanisms === "function") {
        calculateDamageMechanisms();
      }
    },
  );

  // Panggil sekali pas awal render biar statusnya sinkron
  $("select[id^='insp_']").trigger("change");

  $("#step6-calculation, #recalc-step6, #btn-next-5").on(
    "click",
    calculateInspectionStrategy,
  );

  // Trigger insulation condition change untuk mengatur dropdown level dan strategi inspeksi terkait
  $("select[name='insulation']").on("change", function () {
    let insulationStatus = $(this).val();
    let $insulationCond = $("#insulation_condition");

    if (insulationStatus === "No" || insulationStatus === "Not Required") {
      // Otomatis ubah ke Not Applicable dan kunci dropdown-nya
      $insulationCond.val("Not Applicable").prop("disabled", true);
    } else {
      // Buka kembali kuncinya jika milih "Yes"
      $insulationCond.prop("disabled", false);

      // Kembalikan ke pilihan default (misal "Good") jika sebelumnya di-disable
      if ($insulationCond.val() === "Not Applicable") {
        $insulationCond.val("Good");
      }
    }
  });

  // Event Listener pas Radio Button Cathodic Protection di klik
  $("input[name='cathodic_protection']").on("change", function () {
    let cpStatus = $(this).val(); // Ngambil value: Available / Unavailable / Not Required
    let $cpConditionSelect = $("#step3_cp_condition"); // Nembak Dropdown Step 3

    if (cpStatus === "Not Required" || cpStatus === "Unavailable") {
      // Kalo nggak butuh atau nggak ada -> Ubah jadi N/A dan Kunci!
      $cpConditionSelect.val("Not Applicable").prop("disabled", true);
    } else if (cpStatus === "Available") {
      // Kalo Available -> Buka kuncinya
      $cpConditionSelect.prop("disabled", false);

      // Reset pilihan kalau sebelumnya nyangkut di Not Applicable
      if ($cpConditionSelect.val() === "Not Applicable") {
        $cpConditionSelect.val("");
      }
    }
  });

  // Event Listener untuk Preventive of Corrosion
  $("select[name='preventive_corrosion']").on("change", function () {
    // KUNCI PENTING: Ambil TEKS dari option yang dipilih (bukan value ID-nya), lalu ubah ke huruf kecil biar kebal typo
    let selectedText = $(this)
      .find("option:selected")
      .text()
      .trim()
      .toLowerCase();

    // Target dropdown Inhibitor Effectivity
    let $inhibitorDropdown = $("select[name='inhibitor_effectivity']");

    // Cek apakah teksnya mengandung "not required", "none", atau user belum milih (kosong)
    if (
      selectedText.includes("not required") ||
      selectedText.includes("none") ||
      selectedText === "-- select option --" ||
      $(this).val() === ""
    ) {
      // Kunci dan reset form inhibitor
      $inhibitorDropdown.val("").prop("disabled", true);
    } else {
      // Buka kuncinya kalau butuh prevention (misal ada injeksi dll)
      $inhibitorDropdown.prop("disabled", false);
    }
  });

  // Panggil sekali saat halaman di-load biar otomatis nyesuain kalau lagi mode Edit Data
  $("select[name='preventive_corrosion']").trigger("change");

  // Trigger untuk Step 6 (Save Assessment)
  $("#btn_save_assessment").on("click", function (e) {
    e.preventDefault();
    calculateInspectionStrategy();
    let payload = validateAndCollectPayload();

    if (!payload) return;

    let $btn = $(this);
    let originalText = $btn.html();
    $btn.html(
      '<span class="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span> Saving...',
    );
    $btn.prop("disabled", true);

    console.log("SENDING PAYLOAD:", JSON.stringify(payload, null, 2));

    $.ajax({
      url: "/submit",
      type: "POST",
      contentType: "application/json",
      data: JSON.stringify(payload),
      success: function (response) {
        Swal.fire({
          title: "Success",
          text: "Assessment saved successfully!",
          icon: "success",
          customClass: {
            confirmButton: "btn btn-success waves-effect waves-light",
          },
        }).then(() => {
          window.location.href = "/dashboard"; // Redirect ke halaman list
        });
      },
      error: function (xhr, status, error) {
        console.error("Error saving data:", xhr.responseText);

        let errorMessage =
          "Terjadi kesalahan saat menyimpan data. Silakan hubungi tim IT.";
        if (xhr.responseJSON && xhr.responseJSON.message) {
          errorMessage = xhr.responseJSON.message;
        }

        Swal.fire({
          title: "Gagal Menyimpan Assessment",
          text: errorMessage,
          icon: "error",
          showDenyButton: false,
          showCancelButton: false,
          customClass: {
            confirmButton: "btn btn-danger waves-effect waves-light",
          },
        });
      },
      complete: function () {
        $btn.html(originalText);
        $btn.prop("disabled", false);
      },
    });
  });

  // ==========================================
  // EVENT: Saat User Memilih Master Equipment (AUTOFILL)
  // ==========================================
  $("#select2_equipment").on("change", function () {
    const eqId = $(this).val();
    const eqType = $(this).find(":selected").data("type"); // EQT1 (Piping), EQT3 (Vessel/Exchanger)
    const tagNumber = $(this).find(":selected").text().split(" (")[0]; // Ambil Tag No aja

    if (!eqId) return;

    // 1. Tampilkan Loading (Biar user tau aplikasi lagi kerja)
    Swal.fire({
      title: "Fetching History...",
      text: "Mengambil data teknis terakhir dari database",
      allowOutsideClick: false,
      didOpen: () => {
        Swal.showLoading();
      },
    });

    // 2. Tembak API Autofill yang baru kita buat
    fetch(`/api/equipment-autofill/${eqId}`)
      .then((response) => response.json())
      .then((res) => {
        Swal.close();

        // Update UI Dasar dulu
        $("#step3_equipment").val($(this).find(":selected").text());

        if (res.status === "success" && res.data.tag_number != "") {
          const d = res.data;

          // --- A. PENGISIAN DATA TEKNIS (STEP 1) ---
          $("input[name='tag_number']").val(d.tag_number);
          $("select[name='first_use']").val(d.year_built).trigger("change");
          $("#shell_material").val(d.shell_material_id).trigger("change");
          $("input[name='location']").val(d.location);

          // Design Data
          $("input[name='design_press']").val(d.design_pressure);
          $("input[name='design_temp']").val(d.design_temp);
          $("select[name='suhu_design']").val(d.temp_design_unit);

          if (eqType === "EQT3") {
            $("input[name='design_press_tube']").val(d.design_pressure_tube);
            $("input[name='design_temp_tube']").val(d.design_temp_tube);
            $("select[name='suhu_design_tube']").val(d.temp_design_tube_unit);
            $("input[name='diameter_shell']").val(d.diameter);
            $("input[name='diameter_tube']").val(d.diameter_tube);
            $("select[name='diameter_type_shell']").val(d.diameter_type);
            $("select[name='diameter_type_tube']").val(d.diameter_tube_type);
          } else {
            $("input[name='diameter']").val(d.diameter);
            $("select[name='diameter_type']").val(d.diameter_type);
            $("select[name='satuan_diameter']").val(d.diameter_unit);
          }

          // Geometry & Units
          $("input[name='length']").val(d.length);
          $("select[name='satuan_pjg']").val(d.length_unit);
          $("input[name='total_volume']").val(d.volume);
          $("select[name='volume_type']").val(d.volume_unit);
          $("input[name='nozzle']").val(d.nozzle);
          $("input[name='satuan_nozzle']").val(d.nozzle_unit);

          // Specs & Services
          $("select[name='pwht']").val(d.pwht);
          $("select[name='phase_type']").val(d.phase_type);
          $("select[name='internal_lining']").val(d.internal_lining);
          $("select[name='insulation']").val(d.insulation);
          $("input[name='cathodic_protection']").val(d.cathodic_protection);

          // --- B. PENGISIAN CHECKBOX ARRAYS (SPECIAL TRICK) ---
          const setCheckboxes = (className, valueString) => {
            $(`.${className}`).prop("checked", false); // Reset dulu
            if (valueString && valueString !== "-") {
              const vals = valueString.split(", ");
              vals.forEach((v) => {
                $(`.${className}[value="${v}"]`).prop("checked", true);
              });
            }
          };

          setCheckboxes("cert-checkbox", d.certificate);
          setCheckboxes("ref-checkbox", d.data_reference);
          setCheckboxes("special-checkbox", d.special_service);
          setCheckboxes("prot-checkbox", d.protection);

          // --- C. DATA OPERATING (STEP 3) ---
          if (eqType === "EQT1") {
            $("#step3_op_pressure").val(d.operating_pressure);
            $("#step3_op_temperature").val(d.operating_temp);
            $("select[name='suhu_opr']").val(d.temp_op_unit);
            $("input[name='operating_press']").val(d.operating_pressure);
            $("input[name='operating_temp']").val(d.operating_temp);
          } else {
            $("input[name='operating_press_top']").val(d.operating_pressure);
            $("input[name='operating_temp_top']").val(d.operating_temp);
            $("select[name='suhu_opr_top']").val(d.temp_op_unit);
            $("input[name='operating_press_bottom']").val(
              d.operating_pressure_tube,
            );
            $("input[name='operating_temp_bottom']").val(d.operating_temp_tube);
            $("select[name='suhu_opr_bottom']").val(d.temp_op_tube_unit);
          }

          // Fluid Comp
          $("select[name='phase']").val(d.phase);
          $("input[name='comp_h2s']").val(d.h2s_content);
          $("input[name='comp_co2']").val(d.co2_content);
          $("#select2_chloride_contents")
            .val(d.chloride_index)
            .trigger("change");
          $("#select2_ph_contents").val(d.ph_index).trigger("change");

          // Notifikasi Toast Sukses
          const Toast = Swal.mixin({
            toast: true,
            position: "top-end",
            showConfirmButton: false,
            timer: 3000,
          });
          Toast.fire({
            icon: "success",
            title: "Riwayat data ditemukan & otomatis terisi!",
          });

          initDefaultStates();
          runStep2Calculations();
        } else {
          // Jika data kosong, beri notifikasi ringan saja
          console.log("No previous assessment for this equipment.");
        }
      })
      .catch((err) => {
        Swal.close();
        console.error("Autofill Error:", err);
      });
  });

  // ==========================================
  // 3. CORE FUNCTIONS
  // ==========================================

  function initDefaultStates() {
    calculateAndStore();
    generateEquipmentName();
    calculateEnvironmentalSeverity(); // Step 3
    syncOperatingConditions(); //Sync Function 3-4
    calculateDamageMechanisms(); //Step 4
    calculateCriticalityMatrix(); //Step 5
    calculateInspectionStrategy(); //Step 6

    // Trigger default UI toggles
    $("#insulation_condition").trigger("change");
    $("#ext_coating_condition").trigger("change");
  }

  function handleEquipmentTypeChange() {
    const eqType = elements.eqType.find("option:selected").data("type");

    if (eqType === "EQT1") {
      $("#inside_diameter, #operating_press, #design_eq1").fadeIn();
      $(
        "#diameter_shell_tube, #operating_press_top_bottom, #design_eq3",
      ).fadeOut();
    } else if (eqType === "EQT2") {
      $("#inside_diameter, #operating_press_top_bottom, #design_eq3").fadeIn();
      $("#diameter_shell_tube, #operating_press, #design_eq1").fadeOut();
    } else if (eqType === "EQT3") {
      $(
        "#diameter_shell_tube, #operating_press_top_bottom, #design_eq3",
      ).fadeIn();
      $("#inside_diameter, #operating_press, #design_eq1").fadeOut();
    }
  }

  function handleHeadTypeChange() {
    if ($(this).val() === "5") {
      $("#div_crown_knuckle_radius").fadeIn();
    } else {
      $("#div_crown_knuckle_radius").fadeOut();
      $("input[name='crown_radius'], input[name='knuckle_radius']").val("");
    }
  }

  function toggleDamageLevel(wrapperSelector, conditionValue) {
    if (conditionValue === "Damaged") {
      $(wrapperSelector).removeClass("d-none");
    } else {
      $(wrapperSelector).addClass("d-none");
      $(wrapperSelector).find("select").val("Small");
    }
  }

  function calculateAndStore() {
    // Retrieve Form Values
    const press = parseFloat($("input[name='design_press']").val()) || 0;
    const diameter = parseFloat($("input[name='diameter']").val()) || 0;
    const satuan_d = $("select[name='satuan_diameter']").val();
    const stressS = parseFloat($("input[name='allowable_stress']").val()) || 0;
    const jointE = parseFloat($("input[name='joint_efficiency']").val()) || 1.0;

    // 1. Volume Calculation
    const volVal = parseFloat($("input[name='total_volume']").val()) || 0;
    const volType = $("select[name='volume_type']").val();
    const vol_m3 = volType === "ft" ? volVal * 0.0283168 : volVal;

    // 2. Required Thickness (t-min) Calculation (ASME)
    const d_inch = satuan_d === "mm" ? diameter / 25.4 : diameter;
    const R = d_inch / 2;
    let t_min_mm = 0;

    if (press > 0 && R > 0 && stressS > 0) {
      const t_min_inch = (press * R) / (stressS * jointE - 0.6 * press);
      t_min_mm = t_min_inch * 25.4;
      $("input[name='required_thickness']").val(t_min_mm.toFixed(3));
    } else {
      $("input[name='required_thickness']").val("");
    }

    // 3. Corrosion Rate (CR) Calculation
    const prev_t = parseFloat($("input[name='prev_thick_shell']").val()) || 0;
    const act_t = parseFloat($("input[name='act_thick_shell']").val()) || 0;
    const interval =
      parseFloat($("input[name='inspection_interval']").val()) || 0;
    let cr = 0;

    if (prev_t > 0 && act_t > 0 && interval > 0 && prev_t > act_t) {
      cr = (prev_t - act_t) / interval;
      $("input[name='corrosion_rate']").val(cr.toFixed(3));
    } else {
      $("input[name='corrosion_rate']").val("");
    }

    // 4. Save to LocalStorage
    const assessmentData = {
      volume_m3: vol_m3 > 0 ? vol_m3.toFixed(3) : "-",
      required_thickness_mm: t_min_mm > 0 ? t_min_mm.toFixed(3) : "-",
      corrosion_rate_mmyr: cr > 0 ? cr.toFixed(3) : "-",
      equipment: $("#select2_equipment option:selected").text().trim(),
      tagNumber: $("input[name='tag_number']").val(),
    };

    localStorage.setItem(
      "step1_assessmentData",
      JSON.stringify(assessmentData),
    );

    // Update Offcanvas UI
    $("#store_volume").text(
      assessmentData.volume_m3 !== "-" ? `${assessmentData.volume_m3} m³` : "-",
    );
    $("#store_tmin").text(
      assessmentData.required_thickness_mm !== "-"
        ? `${assessmentData.required_thickness_mm} mm`
        : "-",
    );
    $("#store_cr").text(
      assessmentData.corrosion_rate_mmyr !== "-"
        ? `${assessmentData.corrosion_rate_mmyr} mm/yr`
        : "-",
    );
  }

  function syncOperatingConditions() {
    const currentEqType =
      elements.eqType.find("option:selected").data("type") ||
      elements.eqType.val();

    let pressVal = "";
    let tempVal = "";
    let tempUnit = "c";

    if (currentEqType === "EQT3" || currentEqType === "EQT2") {
      $("#he_side_selector").fadeIn();
      const selectedSide = $("input[name='he_side']:checked").val();

      if (selectedSide === "shell") {
        pressVal = $("input[name='operating_press_top']").val();
        tempVal = $("input[name='operating_temp_top']").val();
        tempUnit =
          $("select[name='suhu_opr_top']").val() ||
          $("select[name='suhu_opr']").val();
      } else {
        pressVal = $("input[name='operating_press_bottom']").val();
        tempVal = $("input[name='operating_temp_bottom']").val();
        tempUnit =
          $("select[name='suhu_opr_bottom']").val() ||
          $("select[name='suhu_opr']").val();
      }
    } else {
      $("#he_side_selector").fadeOut();
      pressVal = $("input[name='operating_press']").val();
      tempVal = $("input[name='operating_temp']").val();
      tempUnit = $("select[name='suhu_opr']").val();
    }

    const parsedPress = parseFloat(pressVal);
    const parsedTemp = parseFloat(tempVal);

    // Update Step 3 Inputs
    elements.outPressStep3.val(!isNaN(parsedPress) ? parsedPress : "");

    if (!isNaN(parsedTemp)) {
      let finalTempCelcius =
        tempUnit.toLowerCase() === "f"
          ? ((parsedTemp - 32) * 5) / 9
          : parsedTemp;
      elements.outTempStep3.val(finalTempCelcius.toFixed(2));
    } else {
      elements.outTempStep3.val("");
    }

    calculateEnvironmentalSeverity();
  }

  function generateEquipmentName() {
    const desc =
      $("select[name='description'] option:selected").html() ||
      "Unknown Equipment";
    const tagNo = $("input[name='tag_number']").val() || "Unknown Tag";
    $("#step3_equipment").val(`${desc} (${tagNo})`);
  }

  function calculateEnvironmentalSeverity() {
    const phIndex = parseInt($("#select2_ph_contents").val());
    const h2sIndex = parseInt($("#select2_h2s_contents").val());
    const $severityInput = $("input[name=environmental_severity]");

    if (!phIndex || !h2sIndex || isNaN(phIndex) || isNaN(h2sIndex)) {
      $severityInput
        .val("")
        .removeClass("text-danger text-warning text-success text-muted");
      return;
    }

    const cisccData = localStorage.getItem("master_ciscc");
    let cisccMaster = cisccData ? JSON.parse(JSON.parse(cisccData)) : [];
    console.log("Loaded CISCC Master Data:", cisccMaster);

    let severity = "Unknown";
    const matrixMatch = cisccMaster.find(
      (m) => m.ph_index === phIndex && m.h2s_index === h2sIndex,
    );

    if (matrixMatch) severity = matrixMatch.susceptibility;

    $severityInput
      .val(severity)
      .removeClass("text-danger text-warning text-success text-muted");

    if (severity === "High") $severityInput.addClass("text-danger");
    else if (severity === "Moderate") $severityInput.addClass("text-warning");
    else if (severity === "Low") $severityInput.addClass("text-success");
    else $severityInput.addClass("text-muted");
  }

  function calculateDamageMechanisms() {
    // ==========================================
    // 1. CEK TIPE ALAT & BAGIAN YANG DI-ASES (SHELL / TUBE)
    // ==========================================
    const eqType = $("#select2_equipment").val();
    const assessmentSide = $("input[name='he_side']:checked").val() || "shell";

    // ==========================================
    // 2. AMBIL INPUT DENGAN LOGIKA DINAMIS
    // ==========================================
    let opTemp =
      parseFloat($("#step3_op_temperature").val()) ||
      parseFloat($("input[name='operating_temp']").val()) ||
      0;
    let opPress =
      parseFloat($("#step3_op_pressure").val()) ||
      parseFloat($("input[name='operating_press']").val()) ||
      0;

    if (eqType === "EQT3" && assessmentSide === "head") {
      opTemp = parseFloat($("input[name='op_temp_tube']").val()) || 0;
      opPress = parseFloat($("input[name='op_press_tube']").val()) || 0;
    }

    const vibration = $("select[name='vibration']").val();
    const $selVelocity = $("select[name='velocity'] option:selected");
    const velMic = parseFloat($selVelocity.data("mic")) || 0;
    const prevLevel =
      $("select[name='preventive_corrosion'] option:selected").data("level") ||
      "NONE";

    const phaseValue = $("select[name='phase']").val() || "";
    const molH2S = parseFloat($("input[name='comp_h2s']").val()) || 0;
    const molCO2 = parseFloat($("input[name='comp_co2']").val()) || 0;
    const h2oContent = parseFloat($("input[name='comp_h2o']").val()) || 0;

    const h2sContent = $("#select2_h2s_contents option:selected").text();
    const chlorideLevel = parseInt($("#select2_chloride_contents").val()) || 0;
    const envExtCracking = $("select[name='env_ext_cracking']").val();

    // Input PWHT, Joint Type, Hardness untuk Cracking Logic Baru
    let pwhtStatus = $("input[name='pwht']:checked").val() || "No";
    let jointType = $("select[name='joint_type']").val() || "Welded"; // Welded, As-Welded, atau Seamless
    let hardnessCategory = $("select[name='max_brinell']").val() || "A"; // Value nya langsung "A", "B", atau "C"

    const isAmineChecked =
      $("input[name='contaminant_amine_cracking']:checked").length > 0;
    const steamOut = $("input[name='steam_out']:checked").val() === "1";
    const heatTraced = $("input[name='heat_traced']:checked").val() === "1";

    const insulationStatus = $("select[name='insulation']").val() || "No";
    const insulationCond = $("#insulation_condition").val();
    const insulationLevel = $("#insulation_damage_level").val() || "Small";
    const coatingCond = $("#ext_coating_condition").val();

    // ==========================================
    // 2.5. AMBIL DNA MATERIAL DARI DATABASE (LAPIS 1 FILTER)
    // ==========================================
    let $selectedShell = $("select[name='shell_material'] option:selected");
    let shellExternalRes = $selectedShell.data("external") || "NonRes";

    // Tarik data material (Asumsi lu nyimpen data ini di atribut option atau narik dr variable JS)
    // Kalau belum disimpen di HTML, wajib disimpen dulu bro! Contoh: <option data-co2="CS" data-sulfide="NonRes"...>
    let mat = {
      name: $selectedShell.text(),
      co2_corr: $selectedShell.data("co2corr") || "CS",
      internal: $selectedShell.data("internal") || "",
      external: shellExternalRes,
      mic: $selectedShell.data("mic") || "MICNR",
      amine_cracking: $selectedShell.data("amine-cracking") || "NonRes",
      sulfide_cracking: $selectedShell.data("sulfide-cracking") || "NonRes",
    };

    // ==========================================
    // 3. DAMAGE MECHANISM (API 571 STYLE - UPDATED)
    // ==========================================
    let res = {
      atmospheric: "Not",
      cui: "Not",
      ext_cracking: "Not",
      ssc: "Not",
      amine_scc: "Not",
      hic: "Not",
      ciscc: "Not",
      co2: "Not",
      mic: "Not",
      galvanic: "Not",
    };

    // A. LOGIKA EXTERNAL DAMAGE (Tidak Berubah)
    let isExposedToOutside = !(eqType === "EQT3" && assessmentSide === "head");
    if (isExposedToOutside) {
      if (shellExternalRes === "NonRes") {
        if (coatingCond === "Damaged") res.atmospheric = "High";
        else if (coatingCond === "Good") res.atmospheric = "Low";
        else res.atmospheric = "Medium";

        // ==========================================
        // LOGIKA CUI (Corrosion Under Insulation)
        // ==========================================
        if (
          insulationStatus === "No" ||
          insulationStatus === "Not Required" ||
          insulationCond === "Not Applicable"
        ) {
          res.cui = "Not"; // Mutlak Not karena alat telanjang
        } else {
          // Alat punya insulasi (Yes)
          if (insulationCond === "Damaged") {
            // Cek seberapa parah rusaknya (Bisa lu sesuaikan sama Excel client kalau ada beda)
            if (insulationLevel === "Small") {
              res.cui = "Low";
            } else if (insulationLevel === "Medium") {
              res.cui = "Medium";
            } else if (
              insulationLevel === "Large" ||
              insulationLevel === "Severe"
            ) {
              res.cui = "High";
            } else {
              res.cui = "High"; // Fallback aman
            }
          } else {
            // Kondisi insulasi "Good"
            res.cui = "Low";
          }
        }
      }
      let ext = (envExtCracking || "").toUpperCase();
      if (ext === "HIGH") res.ext_cracking = "High";
      else if (ext === "MEDIUM") res.ext_cracking = "Medium";
      else if (ext === "LOW") res.ext_cracking = "Low";

      if (vibration === "Observed") {
        if (insulationCond === "Damaged" || coatingCond === "Damaged")
          res.ext_cracking = "High";
        else res.ext_cracking = "Medium";
      }
    }

    // B. LOGIKA CRACKING (INTERNAL) MENGGUNAKAN ATURAN BARU
    // 1. Tentukan Severity Awal H2S
    let envSeveritySSC = "None";
    if (h2oContent > 0 && molH2S > 0) {
      let pH2S = molH2S * opPress;
      if (molH2S >= 0.098 && opPress >= 65) envSeveritySSC = "High";
      else if (molH2S >= 0.098 || pH2S >= 0.05) envSeveritySSC = "Low"; // Anggap Low/Med sesuai kodingan lama lu yg diringkas
    }

    // 2. Tentukan Severity Awal Amine
    let envSeverityAmine = "None";
    if (isAmineChecked) {
      if (opTemp > 82 || steamOut || heatTraced) envSeverityAmine = "High";
      else if (opTemp >= 60) envSeverityAmine = "Medium";
      else envSeverityAmine = "Low";
    }

    // 3. Panggil Fungsi Pembantu (Sesuai Matriks Client)
    let crackingResults = calculateAllCracking(
      mat,
      envSeveritySSC,
      envSeverityAmine,
      jointType,
      pwhtStatus,
      hardnessCategory,
    );

    console.log("Cracking Results:", crackingResults);
    res.ssc = crackingResults.ssc;
    res.amine_scc = crackingResults.amine;

    // C. HIC / SOHIC (Hanya menyerang Carbon Steel)
    if (mat.co2_corr === "SS" || mat.name.includes("Stainless")) {
      res.hic = "Not"; // Stainless Steel kebal
    } else {
      if (h2oContent > 0 && h2sContent.includes("PPM")) {
        if (h2sContent.includes("> 10000")) res.hic = "High";
        else if (h2sContent.includes("1000 <")) res.hic = "Medium";
        else res.hic = "Low";
      }
    }

    // D. CISCC (Hanya menyerang Stainless Steel)
    if (mat.co2_corr === "SS" || mat.name.includes("Stainless")) {
      if (chlorideLevel > 1 && opTemp > 60) {
        if (chlorideLevel >= 4) res.ciscc = "High";
        else res.ciscc = "Medium";
      } else if (chlorideLevel > 0) res.ciscc = "Low";
    } else {
      res.ciscc = "Not"; // Carbon Steel kebal CISCC
    }

    // E. CO2 CORROSION (Tergantung DNA Material)
    if (mat.co2_corr === "SS") {
      res.co2 = "Not"; // SS Kebal
    } else {
      if (h2oContent > 0 && molCO2 > 0) {
        let pCO2 = molCO2 * opPress;
        if (pCO2 > 20) res.co2 = "High";
        else if (pCO2 > 5) res.co2 = "Medium";
        else res.co2 = "Low";

        // mitigation
        let prevUpper = prevLevel.toUpperCase();
        if (prevUpper === "HIGH") res.co2 = "Low";
        else if (prevUpper === "MEDIUM") {
          if (res.co2 === "High") res.co2 = "Medium";
          else if (res.co2 === "Medium") res.co2 = "Low";
        }
      }
    }

    // F. MIC & GALVANIC (Sama seperti sebelumnya)
    if (mat.mic === "MICNR") {
      // Bisa tambahin logic kalau kebal
    }
    if (["Ph2", "Ph3", "Ph4"].includes(phaseValue)) {
      if (velMic >= 35) res.mic = "High";
      else if (velMic >= 30) res.mic = "Medium";
      else res.mic = "Low";
    }

    if (h2oContent > 5) res.galvanic = "Medium";
    else if (h2oContent > 0) res.galvanic = "Low";

    // ==========================================
    // SISA KODINGAN DF MAPPING & UPDATE UI (Tidak Berubah)
    // ==========================================
    function mapToDF(level, mech) {
      const table = {
        co2: { Low: 2, Medium: 10, High: 50, Not: 1 },
        mic: { Low: 3, Medium: 15, High: 40, Not: 1 },
        galvanic: { Low: 2, Medium: 5, High: 15, Not: 1 },
        ssc: { Low: 5, Medium: 20, High: 80, Not: 1 },
        amine_scc: { Low: 3, Medium: 10, High: 30, Not: 1 },
        hic: { Low: 5, Medium: 25, High: 70, Not: 1 },
        ciscc: { Low: 3, Medium: 15, High: 50, Not: 1 },
        atmospheric: { Low: 2, Medium: 10, High: 30, Not: 1 },
        cui: { Low: 5, Medium: 20, High: 50, Not: 1 },
        ext_cracking: { Low: 5, Medium: 15, High: 40, Not: 1 },
      };
      return table[mech]?.[level] || 1;
    }

    // (Lanjutkan dengan kode DF_list push, localstorage, dst. seperti aslinya...)
    let DF_list = [];
    Object.keys(res).forEach((key) => {
      let df = mapToDF(res[key], key);
      let df_adj = df * 1; // getInspectionFactor disederhanakan utk contoh, pastikan fungsi aslinya dipanggil

      DF_list.push({ mech: key, df: df, df_adj: df_adj });
    });

    let DF_after_inspection = Math.max(...DF_list.map((d) => d.df_adj));

    let minRL =
      assessmentSide === "shell"
        ? parseFloat($("#sum_rlst_shell").text()) || 20
        : parseFloat($("#sum_rlst_head").text()) || 20;

    function getDFfromRL(rl) {
      if (rl < 2) return 5;
      if (rl < 5) return 3;
      if (rl < 10) return 2;
      if (rl < 20) return 1.5;
      return 1;
    }

    let RL_factor = getDFfromRL(minRL);
    let DF_final = DF_after_inspection * RL_factor;

    const gff = 3.06e-5;
    const FMS = 1.0;
    const PoF = gff * FMS * DF_final;

    function mapPoFToLoF(p) {
      if (p < 1e-5) return 1;
      if (p < 1e-4) return 2;
      if (p < 1e-3) return 3;
      if (p < 1e-2) return 4;
      return 5;
    }

    const lofCategory = mapPoFToLoF(PoF);

    // ==========================================
    // 7. UPDATE UI
    // ==========================================
    // (Asumsi lu punya fungsi updateBadgeState, kalau belum, buat fungsinya spt di contoh sebelumnya)
    console.log(res);
    updateBadgeState("#dm_atmospheric", res.atmospheric);
    updateBadgeState("#dm_cui", res.cui);
    updateBadgeState("#dm_ext_cracking", res.ext_cracking);
    updateBadgeState("#dm_ssc", res.ssc);
    updateBadgeState("#dm_amine_scc", res.amine_scc);
    updateBadgeState("#dm_hic", res.hic);
    updateBadgeState("#dm_ciscc", res.ciscc);
    updateBadgeState("#dm_co2", res.co2);
    updateBadgeState("#dm_mic", res.mic);
    updateBadgeState("#dm_galvanic", res.galvanic);

    $("#ui_lof_score").val(DF_final.toFixed(2));
    $("#lof_category").val(lofCategory).trigger("change");

    if (typeof syncStep5Data === "function") {
      syncStep5Data();
    }
  }

  // ==========================================
  // FUNGSI PEMBANTU (Jangan Lupa Disertakan)
  // Ganti parameter hardnessValue jadi hardnessCategory (karena isinya sekarang huruf A/B/C)
  // ==========================================
  function calculateAllCracking(
    material,
    envSeveritySSC,
    envSeverityAmine,
    jointType,
    pwhtStatus,
    hardnessCategory,
  ) {
    let resultSSC = "Not";
    let resultAmine = "Not";

    // --- LOGIKA SSC ---
    if (material.sulfide_cracking === "Res") {
      resultSSC = "Not";
    } else {
      // Cek apakah ada tegangan sisa (Residual Stress)
      let isAsWelded = true;

      // Kalau dia "Seamless" ATAU di-"PWHT", tegangan sisa dianggap hilang/aman
      if (jointType === "Seamless" || pwhtStatus === "Yes") {
        isAsWelded = false;
      }

      let category = "";

      // Penentuan Kategori A-F langsung pake value dari dropdown lu ("A", "B", atau "C")
      if (hardnessCategory === "A") {
        // < 200
        category = isAsWelded ? "A" : "D";
      } else if (hardnessCategory === "B") {
        // 200 - 237
        category = isAsWelded ? "B" : "E";
      } else if (hardnessCategory === "C") {
        // > 237
        category = isAsWelded ? "C" : "F";
      } else {
        category = isAsWelded ? "A" : "D"; // Fallback aman
      }

      if (envSeveritySSC !== "None" && envSeveritySSC !== "") {
        switch (category) {
          case "D": // Seamless/PWHT + < 200
            resultSSC = "Not";
            break;
          case "A": // As-Welded + < 200
          case "E": // Seamless/PWHT + 200-237
            if (envSeveritySSC === "High") resultSSC = "High";
            else if (envSeveritySSC === "Moderate") resultSSC = "Moderate";
            else resultSSC = "Low";
            break;
          case "B": // As-Welded + 200-237
          case "F": // Seamless/PWHT + > 237
            if (envSeveritySSC === "High" || envSeveritySSC === "Moderate")
              resultSSC = "High";
            else resultSSC = "Moderate";
            break;
          case "C": // As-Welded + > 237 (Paling parah)
            resultSSC = "High";
            break;
        }
      }
    }

    // --- LOGIKA AMINE ---
    if (material.amine_cracking === "Res") {
      resultAmine = "Not";
    } else {
      if (pwhtStatus === "Yes") {
        resultAmine = "Not";
      } else {
        if (envSeverityAmine === "High") resultAmine = "High";
        else if (
          envSeverityAmine === "Medium" ||
          envSeverityAmine === "Moderate"
        )
          resultAmine = "Moderate";
        else if (envSeverityAmine === "Low") resultAmine = "Low";
      }
    }

    return { ssc: resultSSC, amine: resultAmine };
  }

  function updateBadgeState(selector, value) {
    let $el = $(selector);
    $el
      .text(value)
      .removeClass(
        "bg-label-danger bg-label-warning bg-label-success bg-label-secondary bg-label-info",
      );

    if (value === "HIGH" || value === "High") $el.addClass("bg-label-danger");
    else if (value === "MODERATE" || value === "Medium" || value === "Med")
      $el.addClass("bg-label-warning");
    else if (value === "LOW" || value === "Low")
      $el.addClass("bg-label-success");
    else $el.addClass("bg-label-secondary");

    $el.attr("data-value", value);
    $el.addClass("dm-badge");
  }

  function runStep2Calculations() {
    const INCH_TO_MM = 25.4;

    // ==========================================
    // 1. AMBIL PARAMETER DASAR (DARI STEP 1) & KONVERSI KE MM
    // ==========================================
    // Desain
    let P = parseFloat($("input[name='design_press']").val()) || 0;
    let S = parseFloat($("input[name='allowable_stress']").val()) || 0;
    let E = parseFloat($("input[name='joint_efficiency']").val()) || 1.0;

    // Ambil Satuan dari Form
    let unitDia = $("select[name='satuan_diameter']").val() || "inch";
    let unitNozzle = $("select[name='satuan_nozzle']").val() || "inch";

    // DIAMETER (Ubah ke mm)
    let D_raw = parseFloat($("input[name='diameter']").val()) || 0;
    let D = unitDia === "inch" ? D_raw * INCH_TO_MM : D_raw;

    // Shell & Tube (Di HTML labelnya tertulis statis "inch", jadi paksa ke mm)
    let D_Shell_raw = parseFloat($("input[name='diameter_shell']").val()) || 0;
    let D_Shell = D_Shell_raw * INCH_TO_MM;

    let D_Tube_raw = parseFloat($("input[name='diameter_tube']").val()) || 0;
    let D_Tube = D_Tube_raw * INCH_TO_MM;

    let R = D / 2; // Radius dalam mm

    // Ambil Corrosion Allowance (CA) dari Step 1 dan JADIKAN MM
    let CA_in = parseFloat($("input[name='corrosion_allowance']").val()) || 0;
    let CA_mm = CA_in * INCH_TO_MM; // 25.4 (Konstanta dari atas)

    // Parameter Nozzle (Ubah ke mm)
    let D_Nozzle_raw = parseFloat($("input[name='nozzle']").val()) || 0;
    let D_Nozzle =
      unitNozzle === "inch" ? D_Nozzle_raw * INCH_TO_MM : D_Nozzle_raw;
    let R_Nozzle = D_Nozzle / 2;

    // Input fallback nozzle (kalau pakai input manual inch, ubah ke mm)
    let min_req_nozzle_input =
      (parseFloat($("input[name='min_req_thk_nozzle_inch']").val()) || 0) *
      INCH_TO_MM;

    // Equipment Type
    let EQType = $("#select2_equipment option:selected").data("type");

    // Parameter Khusus Torispherical Head (Ubah ke mm sesuai satuan diameter)
    let L_crown_raw = parseFloat($("input[name='crown_radius']").val()) || 0;
    let L_crown = unitDia === "inch" ? L_crown_raw * INCH_TO_MM : L_crown_raw;

    let r_knuckle_raw =
      parseFloat($("input[name='knuckle_radius']").val()) || 0;
    let r_knuckle =
      unitDia === "inch" ? r_knuckle_raw * INCH_TO_MM : r_knuckle_raw;

    // Tipe Dimensi & Head
    let dimType = $("input[name='diameter_type']:checked").val();
    let dimShelltype = $("input[name='diameter_type_shell']:checked").val();
    let dimTubetype = $("input[name='diameter_type_tube']:checked").val();
    let headType = $("select[name='type_head']").val();

    let yearBuilt = parseInt($("select[name='first_use']").val()) || 0;
    let yearPrev = extractYear($("input[name='prev_inspection']").val());
    let yearAct = extractYear($("input[name='act_inspection']").val());
    if (yearAct === 0) {
      yearAct = new Date().getFullYear();
    }

    // Ketebalan SHELL (Tetap, karena dari UI sudah mm)
    let t_act_shell = parseFloat($("input[name='act_thick_shell']").val()) || 0;
    let t_prev_shell =
      parseFloat($("input[name='prev_thick_shell']").val()) || 0;
    let t_init_shell =
      parseFloat($("input[name='shell_wall_thickness']").val()) || 0;

    // Ketebalan HEAD (Tetap, karena dari UI sudah mm)
    let t_act_head = parseFloat($("input[name='act_thick_head']").val()) || 0;
    let t_prev_head = parseFloat($("input[name='prev_thick_head']").val()) || 0;
    let t_init_head =
      parseFloat($("input[name='head_wall_thickness']").val()) || 0;

    // Ketebalan NOZZLE (Tambahan konversi kalau inputnya ambil dari ID inch)
    let t_act_nozzle =
      parseFloat($("input[name='nozzle_actual_thick']").val()) ||
      parseFloat($("input[name='act_thk_nozzle_inch']").val()) * INCH_TO_MM ||
      0;
    let t_prev_nozzle =
      parseFloat($("input[name='nozzle_previous_thick']").val()) || 0;
    let t_init_nozzle =
      parseFloat($("input[name='nozzle_wall_thick']").val()) ||
      parseFloat($("input[name='nom_thk_nozzle_inch']").val()) * INCH_TO_MM ||
      0;

    // Variabel Penampung Hasil
    let results = {
      shell: { treq: 0, mawp: 0, cr_st: 0, cr_lt: 0, rl_st: 0, rl_lt: 0 },
      head: { treq: 0, mawp: 0, cr_st: 0, cr_lt: 0, rl_st: 0, rl_lt: 0 },
      nozzle: { treq: 0, mawp: 0, cr_st: 0, cr_lt: 0, rl_st: 0, rl_lt: 0 },
    };

    // ==========================================
    // 2. RUMUS REQUIRED THICKNESS (treq) & MAWP
    // ==========================================

    if (EQType === "EQT1") {
      P_shell =
        parseFloat($("#design_eq1 input[name='design_press']").val()) || 0;
    } else {
      // EQT2 & EQT3 menggunakan div design_eq3
      P_shell =
        parseFloat($("#design_eq3 input[name='design_press']").val()) || 0;
      P_tube =
        parseFloat($("#design_eq3 input[name='design_press_tube']").val()) || 0;
    }

    let R_shell = D_Shell / 2;
    let R_tube = D_Tube / 2;
    let K = 1.0;

    // Pisahkan logika EQT3 (Heat Exchanger) dengan EQT2/EQT1
    if (EQType === "EQT3") {
      // --- EQT3: HEAT EXCHANGER (Punya Shell & Tube terpisah) ---
      // 1. Hitung Shell EQT3
      if (P_shell > 0 && S > 0 && R_shell > 0) {
        if (dimShelltype === "inside") {
          results.shell.treq = (P_shell * R_shell) / (S * E - 0.6 * P_shell);
          if (t_act_shell > 0)
            results.shell.mawp =
              (S * E * t_act_shell) / (R_shell + 0.6 * t_act_shell);
        } else if (dimShelltype === "outside") {
          results.shell.treq = (P_shell * R_shell) / (S * E + 0.4 * P_shell);
          if (t_act_shell > 0)
            results.shell.mawp =
              (S * E * t_act_shell) / (R_shell - 0.4 * t_act_shell);
        }
      }

      // 2. Hitung Head EQT3 (Menggunakan parameter Tube)
      if (P_tube > 0 && S > 0 && D_Tube > 0) {
        if (headType === "3") {
          if (dimTubetype === "inside") {
            results.head.treq = (P_tube * D_Tube) / (2 * S * E - 0.2 * P_tube);
            if (t_act_head > 0)
              results.head.mawp =
                (2 * S * E * t_act_head) / (D_Tube + 0.2 * t_act_head);
          } else if (dimTubetype === "outside") {
            results.head.treq =
              (P_tube * D_Tube * K) / (2 * S * E + 2 * P_tube * (K - 0.1));
            if (t_act_head > 0)
              results.head.mawp =
                (2 * S * E * t_act_head) /
                (K * D_Tube - 2 * t_act_head * (K - 0.1));
          }
        } else if (headType === "4") {
          if (dimTubetype === "inside") {
            results.head.treq = (P_tube * R_tube) / (2 * S * E - 0.2 * P_tube);
            if (t_act_head > 0)
              results.head.mawp =
                (2 * S * E * t_act_head) / (R_tube + 0.2 * t_act_head);
          } else if (dimTubetype === "outside") {
            results.head.treq = (P_tube * R_tube) / (2 * S * E + 0.8 * P_tube);
            if (t_act_head > 0)
              results.head.mawp =
                (2 * S * E * t_act_head) / (R_tube - 0.8 * t_act_head);
          }
        } else if (headType === "5") {
          if (L_crown > 0 && r_knuckle > 0) {
            let M = 0.25 * (3 + Math.sqrt(L_crown / r_knuckle));
            if (dimTubetype === "inside") {
              results.head.treq =
                (P_tube * L_crown * M) / (2 * S * E - 0.2 * P_tube);
              if (t_act_head > 0)
                results.head.mawp =
                  (2 * S * E * t_act_head) / (M * L_crown + 0.2 * t_act_head);
            } else if (dimTubetype === "outside") {
              results.head.treq =
                (P_tube * L_crown * M) / (2 * S * E + P_tube * (M - 0.2));
              if (t_act_head > 0)
                results.head.mawp =
                  (2 * S * E * t_act_head) /
                  (M * L_crown - t_act_head * (M - 0.2));
            }
          }
        }
      }
    } else {
      let P = P_shell;

      if (P > 0 && S > 0 && R > 0) {
        // 1. Hitung Shell
        if (dimType === "inside") {
          results.shell.treq = (P * R) / (S * E - 0.6 * P);
          if (t_act_shell > 0)
            results.shell.mawp =
              (S * E * t_act_shell) / (R + 0.6 * t_act_shell);
        } else if (dimType === "outside") {
          results.shell.treq = (P * R) / (S * E + 0.4 * P);
          if (t_act_shell > 0)
            results.shell.mawp =
              (S * E * t_act_shell) / (R - 0.4 * t_act_shell);
        }

        // 2. Hitung Head
        if (headType === "3") {
          if (dimType === "inside") {
            results.head.treq = (P * D) / (2 * S * E - 0.2 * P);
            if (t_act_head > 0)
              results.head.mawp =
                (2 * S * E * t_act_head) / (D + 0.2 * t_act_head);
          } else if (dimType === "outside") {
            results.head.treq = (P * D * K) / (2 * S * E + 2 * P * (K - 0.1));
            if (t_act_head > 0)
              results.head.mawp =
                (2 * S * E * t_act_head) / (K * D - 2 * t_act_head * (K - 0.1));
          }
        } else if (headType === "4") {
          if (dimType === "inside") {
            results.head.treq = (P * R) / (2 * S * E - 0.2 * P);
            if (t_act_head > 0)
              results.head.mawp =
                (2 * S * E * t_act_head) / (R + 0.2 * t_act_head);
          } else if (dimType === "outside") {
            results.head.treq = (P * R) / (2 * S * E + 0.8 * P);
            if (t_act_head > 0)
              results.head.mawp =
                (2 * S * E * t_act_head) / (R - 0.8 * t_act_head);
          }
        } else if (headType === "5") {
          if (L_crown > 0 && r_knuckle > 0) {
            let M = 0.25 * (3 + Math.sqrt(L_crown / r_knuckle));
            if (dimType === "inside") {
              results.head.treq = (P * L_crown * M) / (2 * S * E - 0.2 * P);
              if (t_act_head > 0)
                results.head.mawp =
                  (2 * S * E * t_act_head) / (M * L_crown + 0.2 * t_act_head);
            } else if (dimType === "outside") {
              results.head.treq =
                (P * L_crown * M) / (2 * S * E + P * (M - 0.2));
              if (t_act_head > 0)
                results.head.mawp =
                  (2 * S * E * t_act_head) /
                  (M * L_crown - t_act_head * (M - 0.2));
            }
          }
        }
      }
    }

    // --- NOZZLE REQUIRED THICKNESS (Berlaku untuk semua) ---
    if (P_shell > 0 && S > 0 && R_Nozzle > 0) {
      results.nozzle.treq = (P_shell * R_Nozzle) / (S * E - 0.6 * P_shell);
      if (t_act_nozzle > 0) {
        results.nozzle.mawp =
          (S * E * t_act_nozzle) / (R_Nozzle + 0.6 * t_act_nozzle);
      }
    } else if (min_req_nozzle_input > 0) {
      results.nozzle.treq = min_req_nozzle_input;
    }

    // RUMUS: Minimum Thickness = Req Thickness + CA
    let min_thick_shell =
      results.shell.treq > 0 ? results.shell.treq + CA_mm : 0;
    let min_thick_head = results.head.treq > 0 ? results.head.treq + CA_mm : 0;
    let min_thick_nozzle =
      results.nozzle.treq > 0 ? results.nozzle.treq + CA_mm : 0;

    // RUMUS: Remaining Thickness = Actual Thickness - Minimum Thickness
    let rem_thick_shell = t_act_shell > 0 ? t_act_shell - min_thick_shell : 0;
    let rem_thick_head = t_act_head > 0 ? t_act_head - min_thick_head : 0;
    let rem_thick_nozzle =
      t_act_nozzle > 0 ? t_act_nozzle - min_thick_nozzle : 0;

    // ==========================================
    // 3. RUMUS CORROSION RATE (CR)
    // ==========================================
    let interval_st = yearAct - yearPrev;
    let interval_lt = yearAct - yearBuilt;

    // Shell CR
    if (interval_st > 0 && t_prev_shell > 0)
      results.shell.cr_st = (t_prev_shell - t_act_shell) / interval_st;
    if (interval_lt > 0 && t_init_shell > 0)
      results.shell.cr_lt = (t_init_shell - t_act_shell) / interval_lt;

    // Head CR
    if (interval_st > 0 && t_prev_head > 0)
      results.head.cr_st = (t_prev_head - t_act_head) / interval_st;
    if (interval_lt > 0 && t_init_head > 0)
      results.head.cr_lt = (t_init_head - t_act_head) / interval_lt;

    // Nozzle CR
    if (interval_st > 0 && t_prev_nozzle > 0)
      results.nozzle.cr_st = (t_prev_nozzle - t_act_nozzle) / interval_st;
    if (interval_lt > 0 && t_init_nozzle > 0)
      results.nozzle.cr_lt = (t_init_nozzle - t_act_nozzle) / interval_lt;

    // Handle CR negatif
    results.shell.cr_st = results.shell.cr_st > 0 ? results.shell.cr_st : 0.001;
    results.shell.cr_lt = results.shell.cr_lt > 0 ? results.shell.cr_lt : 0.001;
    results.head.cr_st = results.head.cr_st > 0 ? results.head.cr_st : 0.001;
    results.head.cr_lt = results.head.cr_lt > 0 ? results.head.cr_lt : 0.001;
    results.nozzle.cr_st =
      results.nozzle.cr_st > 0 ? results.nozzle.cr_st : 0.001;
    results.nozzle.cr_lt =
      results.nozzle.cr_lt > 0 ? results.nozzle.cr_lt : 0.001;

    // ==========================================
    // 4. RUMUS REMAINING LIFE (RL)
    // ==========================================
    if (results.shell.treq > 0 && t_act_shell > results.shell.treq) {
      results.shell.rl_st =
        (t_act_shell - results.shell.treq) / results.shell.cr_st;
      results.shell.rl_lt =
        (t_act_shell - results.shell.treq) / results.shell.cr_lt;
    }
    if (results.head.treq > 0 && t_act_head > results.head.treq) {
      results.head.rl_st =
        (t_act_head - results.head.treq) / results.head.cr_st;
      results.head.rl_lt =
        (t_act_head - results.head.treq) / results.head.cr_lt;
    }
    if (results.nozzle.treq > 0 && t_act_nozzle > results.nozzle.treq) {
      results.nozzle.rl_st =
        (t_act_nozzle - results.nozzle.treq) / results.nozzle.cr_st;
      results.nozzle.rl_lt =
        (t_act_nozzle - results.nozzle.treq) / results.nozzle.cr_lt;
    }

    // ==========================================
    // 5. UPDATE TAMPILAN HTML (UI)
    // ==========================================
    let inch_conv = 25.4;

    if (results.shell.treq > 0 || t_act_shell > 0) {
      $("input[name='req_thick_shell_mm']").val(results.shell.treq.toFixed(3));
      $("input[name='req_thick_shell_in']").val(
        (results.shell.treq / inch_conv).toFixed(4),
      );
      $("input[name='mawp_shell']").val(results.shell.mawp.toFixed(2));
      $("input[name='cr_st_shell_mm']").val(results.shell.cr_st.toFixed(4));
      $("input[name='cr_st_shell_in']").val(
        (results.shell.cr_st / inch_conv).toFixed(5),
      );
      $("input[name='cr_lt_shell_mm']").val(results.shell.cr_lt.toFixed(4));
      $("input[name='cr_lt_shell_in']").val(
        (results.shell.cr_lt / inch_conv).toFixed(5),
      );
      $("input[name='rem_life_st_shell']").val(
        results.shell.rl_st > 20 ? "> 20" : results.shell.rl_st.toFixed(1),
      );
      $("input[name='rem_life_lt_shell']").val(
        results.shell.rl_lt > 20 ? "> 20" : results.shell.rl_lt.toFixed(1),
      );
    }

    if (results.head.treq > 0 || t_act_head > 0) {
      $("input[name='req_thick_head_mm']").val(results.head.treq.toFixed(3));
      $("input[name='req_thick_head_in']").val(
        (results.head.treq / inch_conv).toFixed(4),
      );
      $("input[name='mawp_head']").val(results.head.mawp.toFixed(2));
      $("input[name='cr_st_head_mm']").val(results.head.cr_st.toFixed(4));
      $("input[name='cr_st_head_in']").val(
        (results.head.cr_st / inch_conv).toFixed(5),
      );
      $("input[name='cr_lt_head_mm']").val(results.head.cr_lt.toFixed(4));
      $("input[name='cr_lt_head_in']").val(
        (results.head.cr_lt / inch_conv).toFixed(5),
      );
      $("input[name='rem_life_st_head']").val(
        results.head.rl_st > 20 ? "> 20" : results.head.rl_st.toFixed(1),
      );
      $("input[name='rem_life_lt_head']").val(
        results.head.rl_lt > 20 ? "> 20" : results.head.rl_lt.toFixed(1),
      );
    }

    if (results.nozzle.treq > 0 || t_act_nozzle > 0) {
      $("input[name='req_thick_nozzle_mm']").val(
        results.nozzle.treq.toFixed(3),
      );
      $("input[name='req_thick_nozzle_in']").val(
        (results.nozzle.treq / inch_conv).toFixed(4),
      );
      $("input[name='mawp_nozzle']").val(results.nozzle.mawp.toFixed(2));
      $("input[name='cr_st_nozzle_mm']").val(results.nozzle.cr_st.toFixed(4));
      $("input[name='cr_st_nozzle_in']").val(
        (results.nozzle.cr_st / inch_conv).toFixed(5),
      );
      $("input[name='cr_lt_nozzle_mm']").val(results.nozzle.cr_lt.toFixed(4));
      $("input[name='cr_lt_nozzle_in']").val(
        (results.nozzle.cr_lt / inch_conv).toFixed(5),
      );
      $("input[name='rem_life_st_nozzle']").val(
        results.nozzle.rl_st > 20 ? "> 20" : results.nozzle.rl_st.toFixed(1),
      );
      $("input[name='rem_life_lt_nozzle']").val(
        results.nozzle.rl_lt > 20 ? "> 20" : results.nozzle.rl_lt.toFixed(1),
      );
    }

    // UPDATE SUMMARY GOVERNING
    let all_rl = [
      results.shell.rl_st,
      results.shell.rl_lt,
      results.head.rl_st,
      results.head.rl_lt,
      results.nozzle.rl_st,
      results.nozzle.rl_lt,
    ];
    let min_rl = Math.min(...all_rl.filter((v) => v > 0));

    let governing_comp = "-";
    if (min_rl === results.shell.rl_st || min_rl === results.shell.rl_lt)
      governing_comp = "Shell";
    else if (min_rl === results.head.rl_st || min_rl === results.head.rl_lt)
      governing_comp = "Head";
    else if (min_rl === results.nozzle.rl_st || min_rl === results.nozzle.rl_lt)
      governing_comp = "Nozzle";

    if (min_rl && min_rl !== Infinity) {
      $("#summary_rem_life").html(
        (min_rl > 20 ? "> 20" : min_rl.toFixed(1)) +
          ' <small class="fs-6 fw-normal text-muted">year(s)</small>',
      );
      $("#summary_rem_life_source").text("Component: " + governing_comp);
      let next_insp_interval = Math.min(min_rl / 2, 10);
      $("#summary_next_insp").text(Math.floor(yearAct + next_insp_interval));
      $("#summary_calc_date").text(new Date().toISOString().split("T")[0]);
    }

    // ==========================================
    // UPDATE TABEL 1: SUMMARY MATRIX
    // ==========================================
    $("#sum_nom_shell").text(t_init_shell || "-");
    $("#sum_req_shell").text(
      results.shell.treq > 0 ? results.shell.treq.toFixed(2) : "-",
    );
    $("#sum_min_shell").text(
      min_thick_shell > 0 ? min_thick_shell.toFixed(2) : "-",
    );
    $("#sum_act_shell").text(t_act_shell || "-");
    $("#sum_remaining_shell").text(
      t_act_shell > 0 ? rem_thick_shell.toFixed(2) : "-",
    );
    $("#sum_crst_shell").text(
      results.shell.cr_st > 0 ? results.shell.cr_st.toFixed(3) : "-",
    );
    $("#sum_crlt_shell").text(
      results.shell.cr_lt > 0 ? results.shell.cr_lt.toFixed(3) : "-",
    );
    $("#sum_rlst_shell").text(
      results.shell.rl_st > 20
        ? "> 20"
        : results.shell.rl_st > 0
          ? results.shell.rl_st.toFixed(1)
          : "-",
    );
    $("#sum_rllt_shell").text(
      results.shell.rl_lt > 20
        ? "> 20"
        : results.shell.rl_lt > 0
          ? results.shell.rl_lt.toFixed(1)
          : "-",
    );

    $("#sum_nom_head").text(t_init_head || "-");
    $("#sum_req_head").text(
      results.head.treq > 0 ? results.head.treq.toFixed(2) : "-",
    );
    $("#sum_min_head").text(
      min_thick_head > 0 ? min_thick_head.toFixed(2) : "-",
    );
    $("#sum_act_head").text(t_act_head || "-");
    $("#sum_remaining_head").text(
      t_act_head > 0 ? rem_thick_head.toFixed(2) : "-",
    );

    $("#sum_crst_head").text(
      results.head.cr_st > 0 ? results.head.cr_st.toFixed(3) : "-",
    );
    $("#sum_crlt_head").text(
      results.head.cr_lt > 0 ? results.head.cr_lt.toFixed(3) : "-",
    );
    $("#sum_rlst_head").text(
      results.head.rl_st > 20
        ? "> 20"
        : results.head.rl_st > 0
          ? results.head.rl_st.toFixed(1)
          : "-",
    );
    $("#sum_rllt_head").text(
      results.head.rl_lt > 20
        ? "> 20"
        : results.head.rl_lt > 0
          ? results.head.rl_lt.toFixed(1)
          : "-",
    );

    $("#sum_nom_nozzle").text(t_init_nozzle || "-");
    $("#sum_req_nozzle").text(
      results.nozzle.treq > 0 ? results.nozzle.treq.toFixed(2) : "-",
    );
    $("#sum_min_nozzle").text(
      min_thick_nozzle > 0 ? min_thick_nozzle.toFixed(2) : "-",
    );
    $("#sum_act_nozzle").text(t_act_nozzle || "-");
    $("#sum_remaining_nozzle").text(
      t_act_nozzle > 0 ? rem_thick_nozzle.toFixed(2) : "-",
    );

    $("#sum_crst_nozzle").text(
      results.nozzle.cr_st > 0 ? results.nozzle.cr_st.toFixed(3) : "-",
    );
    $("#sum_crlt_nozzle").text(
      results.nozzle.cr_lt > 0 ? results.nozzle.cr_lt.toFixed(3) : "-",
    );
    $("#sum_rlst_nozzle").text(
      results.nozzle.rl_st > 20
        ? "> 20"
        : results.nozzle.rl_st > 0
          ? results.nozzle.rl_st.toFixed(1)
          : "-",
    );
    $("#sum_rllt_nozzle").text(
      results.nozzle.rl_lt > 20
        ? "> 20"
        : results.nozzle.rl_lt > 0
          ? results.nozzle.rl_lt.toFixed(1)
          : "-",
    );

    // ==========================================
    // UPDATE TABEL 2: CLADDING CHECKING
    // ==========================================
    let clad_shell =
      parseFloat($("input[name='shell_thick_cladded']").val()) || 0;
    let base_shell =
      parseFloat($("input[name='shell_clad_base_metal']").val()) || 0;

    let clad_head =
      parseFloat($("input[name='head_thick_cladded']").val()) || 0;
    let base_head =
      parseFloat($("input[name='head_clad_base_metal']").val()) || 0;

    let clad_nozzle =
      parseFloat($("input[name='nozzle_thick_cladded']").val()) || 0;
    let base_nozzle =
      parseFloat($("input[name='nozzle_clad_base_metal']").val()) || 0;

    function checkCladding(base, clad, actual, compPrefix) {
      if (clad > 0 && base > 0 && actual > 0) {
        let total_initial = base + clad;
        let clad_consumable = actual > base ? actual - base : 0;

        $(`#clad_base_${compPrefix}`).text(base.toFixed(2));
        $(`#clad_thick_${compPrefix}`).text(clad.toFixed(2));
        $(`#clad_total_${compPrefix}`).text(total_initial.toFixed(2));
        $(`#clad_act_${compPrefix}`).text(actual.toFixed(2));
        $(`#clad_cons_${compPrefix}`).text(clad_consumable.toFixed(2));

        if (actual > base) {
          $(`#clad_stat_${compPrefix}`)
            .attr("class", "badge bg-label-success w-100")
            .text("Cladding Still Exists");
        } else {
          $(`#clad_stat_${compPrefix}`)
            .attr("class", "badge bg-label-danger w-100")
            .text("Cladding Does Not Exist");
        }
      } else {
        $(`#clad_stat_${compPrefix}`)
          .attr("class", "badge bg-label-secondary w-100")
          .text("No Data");
      }
    }

    checkCladding(base_shell, clad_shell, t_act_shell, "shell");
    checkCladding(base_head, clad_head, t_act_head, "head");
    checkCladding(base_nozzle, clad_nozzle, t_act_nozzle, "nozzle");

    localStorage.setItem("resultStep2", JSON.stringify(results));
    localStorage.setItem("min_rl", min_rl);
    localStorage.setItem("governing", governing_comp);
  }

  // ==========================================
  // FUNGSI HITUNG BULAN INSPEKSI (API 581 LOGIC)
  // ==========================================
  function calculateInspectionPeriods() {
    // 1. Ambil Remaining Life (RL) dari Step 2 yang disimpen di localStorage
    // Jika data RL belum ada, kita asumsikan 20 tahun (240 bulan)
    let rl_years = parseFloat(localStorage.getItem("min_rl")) || 20;
    let rl_months = rl_years * 12;

    // Daftar semua kode Damage Mechanism di HTML kita
    let mechanisms = [
      "ext_corr",
      "ext_crack",
      "int_thin",
      "int_crack",
      "loc_corr",
    ];

    mechanisms.forEach((mech) => {
      // --- A. LOGIKA INTRUSIVE (Maks 120 bulan) ---
      let effInt =
        $(`#insp_${mech}_intrusive option:selected`).data("eff") || "None";
      let $periodInt = $(`#period_${mech}_intrusive`);

      if (effInt === "None" || effInt === "") {
        $periodInt.text("-");
      } else {
        // Base API: Half Life (RL / 2)
        let calcInt = rl_months / 2;

        // Penalti kalau metode inspeksinya kurang efektif
        if (effInt === "Medium") calcInt *= 0.8;
        if (effInt === "Low") calcInt *= 0.5;

        // Cap Maksimal 120 bulan (10 tahun)
        let finalInt = Math.min(120, Math.round(calcInt));
        $periodInt.text(finalInt);
      }

      // --- B. LOGIKA NON-INTRUSIVE (Maks 60 bulan) ---
      let effNonInt =
        $(`#insp_${mech}_nonintrusive option:selected`).data("eff") || "None";
      let $periodNonInt = $(`#period_${mech}_nonintrusive`);
      let isOptimized = $(`#opt_${mech}_nonintrusive`).is(":checked");

      if (effNonInt === "None" || effNonInt === "") {
        $periodNonInt.text("-");
      } else {
        // Base API: RL / 3 (Lebih cepet karena ga buka alat)
        let calcNonInt = rl_months / 3;

        if (effNonInt === "High") calcNonInt = rl_months / 2.5; // High confidence
        if (effNonInt === "Medium") calcNonInt = rl_months / 3;
        if (effNonInt === "Low") calcNonInt = rl_months / 4;

        // Jika Optimized dicentang, extend waktunya 25%
        if (isOptimized) {
          calcNonInt *= 1.25;
        }

        // Cap Maksimal 60 bulan (Atau 72 bulan kalau optimized, bebas lu tentuin)
        let maxBulan = isOptimized ? 72 : 60;
        let finalNonInt = Math.min(maxBulan, Math.round(calcNonInt));

        $periodNonInt.text(finalNonInt);
      }
    });
  }

  function calculateCriticalityMatrix() {
    let lof_cat = $("#lof_category").val();

    let cof_fin = $("#cof_financial").val() || "";
    let cof_saf = $("#cof_safety").val() || "";

    // Tentukan Final CoF (Ambil huruf yang paling tinggi / paling parah)
    let cof_final = "";
    if (cof_fin && cof_saf) {
      cof_final = cof_saf > cof_fin ? cof_saf : cof_fin; // 'E' > 'A' di JavaScript string comparison
    } else {
      cof_final = cof_saf || cof_fin;
    }

    // Tembak hasil Final CoF ke input text
    $("#cof_category").val(cof_final);

    // Jika LoF dan Final CoF sudah terisi, jalankan Matrix
    if (lof_cat && cof_final) {
      updateRiskMatrix(lof_cat, cof_final);
    } else {
      // Reset kalau data belum lengkap
      resetRiskMatrix();
    }
  }

  function updateRiskMatrix(lof, cof) {
    // 1. Redupkan semua sel matriks (Reset)
    $("#risk_matrix_table td")
      .removeClass("border border-3 border-dark fw-bolder fs-5 shadow-lg")
      .css("opacity", "0.2");
    $("#risk_matrix_table td.bg-label-dark").css("opacity", "1"); // Kembalikan header A,B,C dan 1,2,3

    // 2. Cari target sel (Contoh: #cell-3-C)
    let targetCellId = `#cell-${lof}-${cof}`;
    let $targetCell = $(targetCellId);

    // 3. Nyalakan sel target!
    $targetCell
      .addClass("border border-3 border-dark fw-bolder fs-5 shadow-lg")
      .css("opacity", "1");

    // 4. Tentukan Risk Level berdasarkan warna sel yang menyala
    let riskLevel = "";
    let badgeClass = "";

    // Kita hitung skor indeks sederhana untuk database (A=1, B=2, dst)
    let cofIndex = cof.charCodeAt(0) - 64;
    let riskIndex = parseInt(lof) * cofIndex;

    if ($targetCell.hasClass("bg-dark")) {
      riskLevel = "EXTREME RISK";
      badgeClass = "bg-dark text-white";
    } else if ($targetCell.hasClass("bg-danger")) {
      riskLevel = "HIGH RISK";
      badgeClass = "bg-danger text-white";
    } else if ($targetCell.hasClass("bg-warning")) {
      riskLevel = "MEDIUM RISK";
      badgeClass = "bg-warning text-dark";
    } else {
      riskLevel = "LOW RISK";
      badgeClass = "bg-success text-white";
    }

    // 5. Update UI Badge Final & Hidden Inputs
    $("#final_risk_label").html(
      `<span class="badge ${badgeClass} rounded-pill px-5 py-3 fs-5 shadow-sm">${riskLevel}</span>`,
    );
    $("#risk_level").val(riskLevel);
    $("#risk_index").val(riskIndex);
  }

  function resetRiskMatrix() {
    $("#risk_matrix_table td")
      .removeClass("border border-3 border-dark fw-bolder fs-5 shadow-lg")
      .css("opacity", "1");
    $("#final_risk_label").html(
      `<span class="badge bg-secondary rounded-pill px-5 py-3 fs-5 shadow-sm">Awaiting Input...</span>`,
    );
    $("#risk_level").val("");
    $("#risk_index").val("");
  }

  function syncStep5Data() {
    // ==========================================
    // 1. KUMPULKAN SEVERITY DARI STEP 4
    // ==========================================
    let dms = {
      atmospheric: $("#dm_atmospheric").text().trim().toUpperCase(),
      ext_cracking: $("#dm_ext_cracking").text().trim().toUpperCase(),
      amine_scc: $("#dm_amine_scc").text().trim().toUpperCase(),
      hic: $("#dm_hic").text().trim().toUpperCase(),
      ciscc: $("#dm_ciscc").text().trim().toUpperCase(),
      co2: $("#dm_co2").text().trim().toUpperCase(),
      mic: $("#dm_mic").text().trim().toUpperCase(),
      ssc: $("#dm_ssc").text().trim().toUpperCase(),
    };

    // ==========================================
    // 2. AMBIL REMAINING LIFE (RL) DARI STEP 2
    // ==========================================
    function parseRL(selector) {
      let text = $(selector)
        .text()
        .replace(/[^0-9.-]/g, "");
      let val = parseFloat(text);
      return isNaN(val) ? 999 : val;
    }

    let rl_shell = parseRL("#sum_rlst_shell");
    let rl_head = parseRL("#sum_rlst_head");
    let rl_nozzle = parseRL("#sum_rlst_nozzle");
    let min_rl = Math.min(rl_shell, rl_head, rl_nozzle);

    // ==========================================
    // 3. KALKULASI DAMAGE FACTOR (DF) - API 581
    // ==========================================

    // FUNGSI BARU: Ambil Diskon Resiko dari Tabel Step 5!
    // API 581 Logic: Inspeksi "High" ngasih diskon DF sampai 80% (multiplier 0.2)
    function getEffDiscount(mechCode) {
      let effInt =
        $(`#insp_${mechCode}_intrusive option:selected`).data("eff") || "None";
      let effNon =
        $(`#insp_${mechCode}_nonintrusive option:selected`).data("eff") ||
        "None";

      let levels = [effInt, effNon];
      if (levels.includes("High")) return 0.2; // Resiko didiskon 80%
      if (levels.includes("Medium")) return 0.5; // Resiko didiskon 50%
      if (levels.includes("Low")) return 0.8; // Resiko didiskon 20%
      return 1.0; // None (Resiko mentok / tidak didiskon)
    }

    // a. Thinning DF (Base)
    let df_thinning = 1.0;
    if (min_rl <= 0)
      df_thinning = 2000.0; // Kritis / Past Failure
    else if (min_rl <= 2) df_thinning = 500.0;
    else if (min_rl <= 5) df_thinning = 100.0;
    else if (min_rl <= 10) df_thinning = 20.0;
    else if (min_rl <= 20) df_thinning = 5.0;
    else df_thinning = 1.0; // Aman (> 20 tahun)

    // APLIKASIKAN DISKON EFFECTIVENESS KE THINNING
    df_thinning = df_thinning * getEffDiscount("int_thin");

    // b. Cracking & External DF (Base)
    function getCrackingDF(severity) {
      if (severity === "HIGH") return 100.0;
      if (severity === "MEDIUM") return 20.0;
      if (severity === "LOW") return 5.0;
      return 1.0; // Not Susceptible
    }

    // APLIKASIKAN DISKON EFFECTIVENESS KE MASING-MASING DM
    let df_ext_crack =
      getCrackingDF(dms.ext_cracking) * getEffDiscount("ext_crack");

    // Untuk Internal Cracking, ambil resiko cracking paling parah di dalam alat
    let df_int_crack =
      Math.max(
        getCrackingDF(dms.amine_scc),
        getCrackingDF(dms.hic),
        getCrackingDF(dms.ciscc),
        getCrackingDF(dms.ssc),
      ) * getEffDiscount("int_crack");

    let df_ext_corr =
      getCrackingDF(dms.atmospheric) * getEffDiscount("ext_corr");
    let df_loc_corr =
      Math.max(getCrackingDF(dms.co2), getCrackingDF(dms.mic)) *
      getEffDiscount("loc_corr");

    let max_df_cracking = Math.max(
      df_ext_crack,
      df_int_crack,
      df_ext_corr,
      df_loc_corr,
    );

    // TOTAL DF (Skenario terburuk)
    let total_df = Math.max(df_thinning, max_df_cracking);

    // ==========================================
    // 4. KALKULASI PROBABILITY OF FAILURE (PoF)
    // ==========================================
    const GFF = 3.06e-5;
    const FMS = 1.0;
    const PoF = GFF * FMS * total_df;

    // ==========================================
    // 5. MAPPING PoF KE LIKELIHOOD CATEGORY (1-5)
    // ==========================================
    let autoLofCat = "1";
    if (PoF > 1e-2) autoLofCat = "5";
    else if (PoF > 1e-3) autoLofCat = "4";
    else if (PoF > 1e-4) autoLofCat = "3";
    else if (PoF > 1e-5) autoLofCat = "2";
    else autoLofCat = "1";

    // ==========================================
    // 6. UPDATE UI & TRIGGER MATRIX
    // ==========================================
    let displayPoF = PoF.toExponential(2).toUpperCase();
    $("#ui_lof_score").val(`PoF: ${displayPoF} (DF: ${total_df.toFixed(2)})`);

    if ($("#lof_category").val() !== autoLofCat) {
      $("#lof_category").val(autoLofCat).trigger("change");
    }
  }

  function calculateInspectionStrategy() {
    // ==========================================
    // 1. AMBIL DATA & BASE INTERVAL
    // ==========================================
    let minRL = parseFloat(localStorage.getItem("min_rl")) || 20;
    let governing = localStorage.getItem("governing") || "-";
    let riskLevel = $("#risk_level").val() || "LOW RISK";

    let baseInterval = Math.min(minRL / 2, 10);

    // ==========================================
    // 2. RISK & MECH FACTOR (FACTOR BREAKDOWN)
    // ==========================================
    let riskFactorMap = {
      "LOW RISK": 1.0,
      "MEDIUM RISK": 0.75,
      "HIGH RISK": 0.5,
      "EXTREME RISK": 0.25,
    };
    let riskFactor = riskFactorMap[riskLevel] || 1.0;

    function getWorstDM() {
      let levels = [];
      $(".dm-badge").each(function () {
        let val = $(this).text().trim();
        if (val && val !== "Not") levels.push(val);
      });
      if (levels.includes("High")) return "High";
      if (levels.includes("Medium")) return "Medium";
      if (levels.includes("Low")) return "Low";
      return "Not";
    }

    let worstDM = getWorstDM();
    let mechFactorMap = { High: 0.5, Medium: 0.75, Low: 1.0, Not: 1.1 };
    let mechFactor = mechFactorMap[worstDM] || 1.0;

    // ==========================================
    // 3. FINAL MASTER INTERVAL
    // ==========================================
    let finalInterval = baseInterval * riskFactor * mechFactor;
    finalInterval = Math.max(1, Math.min(finalInterval, 10)); // Clamp 1-10 years

    let currentYear = new Date().getFullYear();
    let nextInspection = Math.floor(currentYear + finalInterval);

    // ==========================================
    // 4. UPDATE UI - KARTU ATAS & FACTOR BREAKDOWN
    // ==========================================
    $("#insp_interval").text(finalInterval.toFixed(1) + " Years");
    $("#insp_next_year").text(nextInspection);
    $("#insp_governing").text(governing);

    // List Factor Breakdown
    $("#ui_factor_risklevel").text(riskLevel || "Not Set");
    $("#ui_factor_risk").text("x " + riskFactor);

    // Set Badge Warna untuk Worst DM
    let badgeColor = "bg-secondary";
    if (worstDM === "High") badgeColor = "bg-danger";
    if (worstDM === "Medium") badgeColor = "bg-warning text-dark";
    $("#ui_factor_worst_dm")
      .text(worstDM)
      .attr("class", "badge rounded-pill " + badgeColor);

    $("#ui_factor_mech").text("x " + mechFactor);

    // ==========================================
    // 5. KOMPILASI RECOMMENDED NDT METHOD (TOP CARD)
    // ==========================================
    const mechs = [
      { id: "ext_corr", label: "External Corrosion" },
      { id: "ext_crack", label: "External Surface Cracking" },
      { id: "int_thin", label: "Internal Thinning" },
      { id: "int_crack", label: "Internal Cracking" },
      { id: "loc_corr", label: "Localised Internal Corrosion" },
    ];

    function getShortMethod(val) {
      if (!val || val === "None") return null;
      let txt = val.toLowerCase();
      if (txt.includes("ut") || txt.includes("ultrasonic")) return "UT";
      if (txt.includes("rt") || txt.includes("radiographic")) return "RT";
      if (txt.includes("visual") || txt.includes("vie")) return "VT";
      if (txt.includes("mpt") || txt.includes("dpt") || txt.includes("wfmt"))
        return "MT/PT";
      return val;
    }

    // 1. Ambil semua metode yang dipilih dari tabel di Step 5
    let allMethods = [];
    
    // Daftar ID Damage Mechanism di tabel Step 5 lu
    const dmKeys = ['ext_corr', 'ext_crack', 'int_thin', 'int_crack', 'loc_corr'];

    dmKeys.forEach((key) => {
      // Ambil TEKS yang tampil di layar (bukan value-nya)
      let nonIntrusiveText = $(`#insp_${key}_nonintrusive option:selected`).text().trim();
      let intrusiveText = $(`#insp_${key}_intrusive option:selected`).text().trim();

      // Hanya masukkan jika bukan "None"
      if (nonIntrusiveText !== "None" && nonIntrusiveText !== "") {
        allMethods.push(nonIntrusiveText);
      }
      if (intrusiveText !== "None" && intrusiveText !== "") {
        allMethods.push(intrusiveText);
      }
    });

    // 2. Hilangkan duplikasi (biar kalau ada metode sama di DM berbeda nggak dobel)
    let uniqueMethods = [...new Set(allMethods)];

    // 3. Buat tampilan UI yang "Manusiawi" (Tanpa Singkatan membingungkan)
    let methodHtmlArray = uniqueMethods.map((method) => {
      return `<div class="d-flex align-items-center mb-2">
                <i class="mdi mdi-check-circle text-primary me-2"></i>
                <span class="fw-bold text-dark">${method}</span>
              </div>`;
    });

    // 4. Render ke UI Step 6
    let finalMethodHtml = methodHtmlArray.length > 0
        ? `<div class="bg-light p-3 rounded-3 border-start border-primary border-4">
             ${methodHtmlArray.join('')}
           </div>`
        : '<span class="text-danger fw-bold">No Inspection Planned</span>';

    $("#insp_method").html(finalMethodHtml);

    // 5. SIMPAN KE PAYLOAD (Teks Lengkap untuk Database & PDF)
    // Kita gabung pake tanda " + " biar di laporan PDF nanti bacanya enak
    localStorage.setItem('recommended_methods', uniqueMethods.join(" + "));
    // payload.results.recommended_method = uniqueMethods.join(" + ");

    // ==========================================
    // INITIALIZE BOOTSTRAP TOOLTIP
    // ==========================================
    // Hancurkan tooltip lama (kalau ada) biar gak error pas recalculate, lalu pasang yang baru
    $('[data-bs-toggle="tooltip"]').tooltip("dispose");
    $('[data-bs-toggle="tooltip"]').tooltip();

    // ==========================================
    // 6. RENDER TIMELINE SERIES (PROGRESS BAR) BAWAH
    // ==========================================
    let tbodyHtml = "";

    mechs.forEach((m) => {
      let ni_eff =
        $(`#insp_${m.id}_nonintrusive option:selected`).data("eff") || "None";
      let ni_period = parseInt($(`#period_${m.id}_nonintrusive`).text()) || 0;

      let in_eff =
        $(`#insp_${m.id}_intrusive option:selected`).data("eff") || "None";
      let in_period = parseInt($(`#period_${m.id}_intrusive`).text()) || 0;

      let plans = [];
      if (ni_eff !== "None" && ni_period > 0)
        plans.push({ type: "(On-Stream)", eff: ni_eff, period: ni_period });
      if (in_eff !== "None" && in_period > 0)
        plans.push({ type: "(Off-Stream)", eff: in_eff, period: in_period });

      if (plans.length === 0) {
        tbodyHtml += `
            <tr>
               <td class="fw-bold text-muted">${m.label}</td>
               <td class="text-center text-muted">-</td>
               <td class="text-center text-muted">-</td>
               <td class="px-3">
                  <div class="progress shadow-none" style="height: 22px; background-color: #f0f2f5; border-radius: 6px;">
                     <div class="progress-bar bg-transparent" style="width: 100%;"></div>
                  </div>
               </td>
            </tr>`;
      } else {
        plans.forEach((plan, index) => {
          let dmLabel =
            index === 0
              ? `<td class="fw-bold text-dark" rowspan="${plans.length}">${m.label}</td>`
              : "";

          let badgeClass = "bg-secondary";
          let barClass = "bg-secondary";
          if (plan.eff === "High") {
            badgeClass = "bg-danger";
            barClass = "bg-danger";
          }
          if (plan.eff === "Medium") {
            badgeClass = "bg-warning text-dark";
            barClass = "bg-warning";
          }

          let widthPercent = Math.min((plan.period / 120) * 100, 100);

          tbodyHtml += `
                <tr>
                   ${dmLabel}
                   <td class="text-center">
                      <span class="badge ${badgeClass} mb-1 d-block">${plan.eff}</span>
                      <small class="text-dark" style="font-size: 0.65rem;">${plan.type}</small>
                   </td>
                   <td class="text-center fw-bold fs-6">${plan.period} <span class="fw-normal text-muted" style="font-size: 0.7rem;">mo</span></td>
                   <td class="align-middle px-3">
                      <div class="progress shadow-none" style="height: 22px; background-color: #f0f2f5; border-radius: 6px;">
                         <div class="progress-bar progress-bar-striped progress-bar-animated ${barClass}" role="progressbar" style="width: ${widthPercent}%"></div>
                      </div>
                   </td>
                </tr>`;
        });
      }
    });

    $("#final_inspection_timeline tbody").html(tbodyHtml);
  }

  function getInspectionFactor(mech, method) {
    const table = {
      co2: { UT: 0.2, RT: 0.5, VT: 0.8 },
      mic: { UT: 0.2, RT: 0.5, VT: 0.8 },
      atmospheric: { "VT+UT": 0.2, UT: 0.5, RT: 0.8 },
      cui: { "VT+UT": 0.3, UT: 0.6, RT: 0.9 },

      ssc: { WFMT: 0.2, UT: 0.5, RT: 0.8 },
      hic: { UT: 0.3, RT: 0.6 },
      ciscc: { WFMT: 0.2, UT: 0.5 },
    };

    return table[mech]?.[method] || 1.0;
  }

  // ==========================================
  // 8. DATA COLLECTION & PAYLOAD GENERATION
  // ==========================================

  function validateAndCollectPayload() {
    // --- A. VALIDASI INPUTAN WAJIB (Mencegah data sampah masuk DB) ---
    let errors = [];

    // Validasi Step 1 (Header Data)
    let tagNumber = $("input[name='tag_number']").val();
    let eqId = $("#select2_equipment").val();
    let eqType = $("#select2_equipment option:selected").data("type");
    let shellMaterial = $("select[name='shell_material']").val();

    if (!tagNumber) errors.push("Step 1: Tag Number is required.");
    if (!eqId) errors.push("Step 1: Equipment must be selected.");
    if (!shellMaterial) errors.push("Step 1: Shell Material must be selected.");

    // Validasi Wajib (Mandatory)
    if (!tagNumber) {
      errors.push("Tag Number / Serial Number is strictly required!");
      // Otomatis pindah ke Step 1 biar user lihat kesalahannya
      stepper.to(1);
    }

    // Validasi Step 2 (Thickness Data)
    let actThickShell = parseFloat($("input[name='act_thick_shell']").val());
    if (isNaN(actThickShell) || actThickShell <= 0) {
      errors.push(
        "Step 2: Actual Shell Thickness is required and must be > 0.",
      );
    }

    // Validasi Step 5 (Risk)
    let finalRiskLevel = $("#risk_level").val();
    if (!finalRiskLevel || finalRiskLevel === "Pending Evaluation") {
      errors.push(
        "Step 5: Risk Matrix must be evaluated. Please select LoF and CoF.",
      );
    }

    if (errors.length > 0) {
      Swal.fire({
        title: "Validation Error",
        icon: "error",
        html: errors.join("<br>"),
        showDenyButton: false,
        showCancelButton: false,
        customClass: {
          confirmButton: "btn btn-danger waves-effect waves-light",
        },
      });
      return null;
    }

    // --- B. BENTUK JSON PAYLOAD ---
    // 1. Kumpulin data Checkbox Arrays menjadi String (separator koma)
    const certArray = Array.from(
      document.querySelectorAll(".cert-checkbox:checked"),
    ).map((el) => el.value);
    const refArray = Array.from(
      document.querySelectorAll(".ref-checkbox:checked"),
    ).map((el) => el.value);
    const specialArray = Array.from(
      document.querySelectorAll(".special-checkbox:checked"),
    ).map((el) => el.value);
    const protArray = Array.from(
      document.querySelectorAll(".prot-checkbox:checked"),
    ).map((el) => el.value);

    let payload = {
      // 1. HEADER: Equipment Data
      equipment: {
        tag_number: tagNumber,
        master_equipment_id: parseInt(eqId),
        location: $("input[name='location']").val() || "", // FIX BUG TYPO DI SINI
        year_built: parseInt($("select[name='first_use']").val()) || 0,
        shell_material_id: parseInt(shellMaterial) || 0,
        design_pressure: parseFloat($("input[name='design_press']").val()) || 0,
        design_pressure_tube:
          parseFloat($("input[name='design_press_tube']").val()) || 0,
        design_temp: parseFloat($("input[name='design_temp']").val()) || 0,
        design_temp_tube:
          parseFloat($("input[name='design_temp_tube']").val()) || 0,
        diameter:
          (eqType == "EQT3"
            ? parseFloat($("input[name='diameter_shell']").val())
            : parseFloat($("input[name='diameter']").val())) || 0,
        diameter_tube:
          (eqType == "EQT3"
            ? parseFloat($("input[name='diameter_tube']").val())
            : parseFloat($("input[name='diameter']").val())) || 0,
        volume: parseFloat($("input[name='total_volume']").val()) || 0,
        diameter_type:
          (eqType == "EQT3"
            ? $("select[name='diameter_type_shell']").val()
            : $("select[name='diameter_type']").val()) || "inside",
        diameter_unit:
          (eqType == "EQT3"
            ? "inch"
            : $("select[name='satuan_diameter']").val()) || "inch",
        diameter_tube_type:
          eqType == "EQT3"
            ? $("select[name='diameter_type_tube']").val()
            : null,
        diameter_tube_unit: eqType == "EQT3" ? "inch" : null,
        length: parseFloat($("input[name='length']").val()) || 0,
        length_unit: $("select[name='satuan_pjg']").val() || "ft",
        nozzle: parseFloat($("input[name='nozzle']").val()) || 0,
        nozzle_unit: $("input[name='satuan_nozzle']").val() || "inch",
        volume_unit: $("select[name='volume_type']").val() || "m",
        temp_design_unit: $("select[name='suhu_design']").val() || "c",
        temp_design_tube_unit:
          eqType == "EQT3" ? $("select[name='suhu_design_tube']").val() : null,
        pwht: $("select[name='pwht']").val() || "No",
        certificate: certArray.join(", ") || "-",
        data_reference: refArray.join(", ") || "-",
        phase_type: $("select[name='phase_type']").val() || "multi phase",
        internal_lining: $("select[name='internal_lining']").val() || "None",
        insulation: $("select[name='insulation']").val() || "No",
        special_service: specialArray.join(", ") || "-",
        protection: protArray.join(", ") || "-",
        cathodic_protection:
          $("input[name='cathodic_protection']:checked").val() || "No",
      },

      // 2. DETAIL: Assessment General Info
      assessment: {
        assessment_date: new Date().toISOString().split("T")[0],
        prev_inspection_date: $("input[name='prev_inspection']").val() || null,
        act_inspection_date: $("input[name='act_inspection']").val() || null,
        operating_pressure:
          (eqType == "EQT1"
            ? parseFloat($("input[name='operating_press']").val())
            : parseFloat($("input[name='operating_press_top']").val())) || 0,
        operating_temp:
          (eqType == "EQT1"
            ? parseFloat($("input[name='operating_temp']").val())
            : parseFloat($("input[name='operating_temp_top']").val())) || 0,
        operating_pressure_tube:
          (eqType == "EQT1"
            ? 0
            : parseFloat($("input[name='operating_press_bottom']").val())) || 0,
        operating_temp_tube:
          (eqType == "EQT1"
            ? 0
            : parseFloat($("input[name='operating_temp_bottom']").val())) || 0,

        temp_op_unit:
          (eqType == "EQT3"
            ? $("select[name='suhu_opr_top']").val()
            : $("select[name='suhu_opr']").val()) || "c",
        temp_op_tube_unit:
          eqType == "EQT3" ? $("select[name='suhu_opr_bottom']").val() : null,
      },

      // 3. DETAIL: Thickness & Corrosion Rate Data
      thickness_data: {
        shell: {
          prev_thick:
            parseFloat($("input[name='prev_thick_shell']").val()) || 0,
          act_thick: actThickShell,
          t_req: parseFloat($("input[name='req_thick_shell_mm']").val()) || 0,
          corrosion_rate:
            parseFloat($("input[name='cr_st_shell_mm']").val()) || 0,
          remaining_life:
            parseRemainingLife($("input[name='rem_life_st_shell']").val()) || 0,
        },
        head: {
          prev_thick: parseFloat($("input[name='prev_thick_head']").val()) || 0,
          act_thick: parseFloat($("input[name='act_thick_head']").val()) || 0,
          t_req: parseFloat($("input[name='req_thick_head_mm']").val()) || 0,
          corrosion_rate:
            parseFloat($("input[name='cr_st_head_mm']").val()) || 0,
          remaining_life:
            parseRemainingLife($("input[name='rem_life_st_head']").val()) || 0,
        },
        // Tambahan object nozzle biar komplit ngikutin struct Go-nya
        nozzle: {
          prev_thick:
            parseFloat($("input[name='nozzle_previous_thick']").val()) || 0,
          act_thick:
            parseFloat($("input[name='nozzle_actual_thick']").val()) || 0,
          t_req: parseFloat($("input[name='req_thick_nozzle_mm']").val()) || 0,
          corrosion_rate:
            parseFloat($("input[name='cr_st_nozzle_mm']").val()) || 0,
          remaining_life:
            parseRemainingLife($("input[name='rem_life_st_nozzle']").val()) ||
            0,
        },
      },

      // 4. DETAIL: Operating Environment (Dari Step 3)
      environment: {
        phase: $("select[name='phase']").val() || "",
        h2s_content: parseFloat($("input[name='comp_h2s']").val()) || 0,
        co2_content: parseFloat($("input[name='comp_co2']").val()) || 0,
        h2o_content: parseFloat($("input[name='comp_h2o']").val()) || 0,
        chloride_index: parseInt($("#select2_chloride_contents").val()) || 0,
        ph_index: parseInt($("#select2_ph_contents").val()) || 0,

        impact_production:
          $("select[name='impact_for_production']").val() || "",
        insulation_condition: $("#insulation_condition").val() || "",
        insulation_damage_level: $("#insulation_damage_level").val() || "",
        coating_condition: $("#ext_coating_condition").val() || "",
        coating_damage_level: $("#ext_coating_damage_level").val() || "",
        corrective_description:
          $("textarea[name='corrective_description']").val() || "",
        corrective_action:
          $("textarea[name='corrective_action_taken']").val() || "",
        corrective_date: $("input[name='corrective_date']").val() || null,
      },

      // 5. DETAIL: Damage Mechanisms (Dari Step 4)
      damage_mechanisms: {
        atmospheric: $("#dm_atmospheric").data("value") || "Not",
        cui: $("#dm_cui").data("value") || "Not",
        ext_cracking: $("#dm_ext_cracking").data("value") || "Not",
        co2: $("#dm_co2").data("value") || "Not",
        mic: $("#dm_mic").data("value") || "Not",
        ssc: $("#dm_ssc").data("value") || "Not",
        amine_scc: $("#dm_amine_scc").data("value") || "Not",
        hic: $("#dm_hic").data("value") || "Not",
        ciscc: $("#dm_ciscc").data("value") || "Not",
        galvanic: $("#dm_galvanic").text().trim() || "Not",
        lof_score: $("#ui_lof_score").val() || "",
      },

      // 6. DETAIL: Final Risk & Strategy (Dari Step 5 & 6)
      results: {
        lof_category: parseInt($("#lof_category").val()) || 0,
        cof_financial: $("#cof_financial").val() || "",
        cof_safety: $("#cof_safety").val() || "",
        cof_category: $("#cof_category").val() || "",
        risk_level: finalRiskLevel,
        risk_index: parseInt($("#risk_index").val()) || 0,

        // FIX TERPENTING: Form Step 5 yang kompleks ditampung di 1 objek ini
        inspection_data: {
          ext_corr_nonintrusive:
            $("#insp_ext_corr_nonintrusive").val() || "None",
          ext_corr_intrusive: $("#insp_ext_corr_intrusive").val() || "None",
          ext_crack_nonintrusive:
            $("#insp_ext_crack_nonintrusive").val() || "None",
          ext_crack_intrusive: $("#insp_ext_crack_intrusive").val() || "None",
          int_thin_nonintrusive:
            $("#insp_int_thin_nonintrusive").val() || "None",
          int_thin_intrusive: $("#insp_int_thin_intrusive").val() || "None",
          int_crack_nonintrusive:
            $("#insp_int_crack_nonintrusive").val() || "None",
          int_crack_intrusive: $("#insp_int_crack_intrusive").val() || "None",
          loc_corr_nonintrusive:
            $("#insp_loc_corr_nonintrusive").val() || "None",
          loc_corr_intrusive: $("#insp_loc_corr_intrusive").val() || "None",
        },

        governing_component: localStorage.getItem("governing") || "-",
        max_interval_years: parseFloat($("#insp_interval").text()) || 0,
        next_inspection_year: parseInt($("#insp_next_year").text()) || 0,

        // .text() aman dipakai untuk ambil isi tooltip (buang HTML <span> nya)
        recommended_method: localStorage.getItem('recommended_methods') //$("#insp_method").text().trim() || "",
      },

      cladding_data: {
        shell: {
          base_metal: parseFloat($("#clad_base_shell").text()) || 0,
          cladding: parseFloat($("#clad_thick_shell").text()) || 0,
          total_init: parseFloat($("#clad_total_shell").text()) || 0,
          act_now: parseFloat($("#clad_act_shell").text()) || 0,
          act_cons: parseFloat($("#clad_cons_shell").text()) || 0,
          status: $("#clad_stat_shell").text().trim() || "No Data",
        },
        head: {
          base_metal: parseFloat($("#clad_base_head").text()) || 0,
          cladding: parseFloat($("#clad_thick_head").text()) || 0,
          total_init: parseFloat($("#clad_total_head").text()) || 0,
          act_now: parseFloat($("#clad_act_head").text()) || 0,
          act_cons: parseFloat($("#clad_cons_head").text()) || 0,
          status: $("#clad_stat_head").text().trim() || "No Data",
        },
        nozzle: {
          base_metal: parseFloat($("#clad_base_nozzle").text()) || 0,
          cladding: parseFloat($("#clad_thick_nozzle").text()) || 0,
          total_init: parseFloat($("#clad_total_nozzle").text()) || 0,
          act_now: parseFloat($("#clad_act_nozzle").text()) || 0,
          act_cons: parseFloat($("#clad_cons_nozzle").text()) || 0,
          status: $("#clad_stat_nozzle").text().trim() || "No Data",
        },
      },
    };

    return payload;
  }

  // Fungsi buat ngekstrak angka dari string (misal: "> 20" jadi 20)
  function parseRemainingLife(val) {
    if (!val) return 0;

    // Pastikan jadi string dulu, lalu hapus semua karakter selain angka dan titik (desimal)
    let cleanString = String(val).replace(/[^0-9.]/g, "");

    // Ubah balik jadi Float (angka desimal)
    let num = parseFloat(cleanString);

    return isNaN(num) ? 0 : num;
  }
});

// Utility Functions (Outside ready scope if needed globally, but safer inside if not exported)
function extractYear(val) {
  if (!val) return 0;
  return parseInt(val.split("-")[0]) || 0;
}
