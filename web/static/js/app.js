document.querySelectorAll("[data-password-toggle]").forEach((button) => {
  button.addEventListener("click", () => {
    const input = document.getElementById(button.dataset.target);
    if (!input) return;

    const visible = input.type === "text";
    input.type = visible ? "password" : "text";
    button.setAttribute("aria-pressed", String(!visible));
    button.setAttribute("aria-label", visible ? "Tampilkan password" : "Sembunyikan password");
    const icon = button.querySelector("[data-password-icon]");
    icon?.classList.toggle("fa-eye", visible);
    icon?.classList.toggle("fa-eye-slash", !visible);
  });
});

document.querySelectorAll('select[name="employment_status"]').forEach((select) => {
  const statuses = [
    ["daily_worker", "Harian Lepas (HL)"],
    ["post_retirement_daily_worker", "Harian Lepas Pasca Pensiun (HLPP)"],
  ];
  statuses.forEach(([value, label]) => {
    if (![...select.options].some((option) => option.value === value)) select.add(new Option(label, value));
  });
});

const employmentStatusLabels = {
  trainee: "Trainee",
  probation: "Probation",
  contract: "Kontrak",
  permanent: "Tetap",
  daily_worker: "HL",
  post_retirement_daily_worker: "HLPP",
};

document.querySelectorAll("[data-filter-row] [data-edit-employee]").forEach((button) => {
  const details = document.querySelector(`[data-employee-details] [data-employee-id="${button.dataset.id}"]`);
  const badge = button.closest("tr")?.querySelector("td:nth-child(5) span");
  if (!badge || !details) return;
  const employmentStatus = details.dataset.employmentStatus ?? "";
  badge.textContent = (employmentStatusLabels[employmentStatus] ?? employmentStatus) || "Belum diisi";
});

const dashboardSidebar = document.getElementById("dashboardSidebar");
const sidebarBackdrop = document.getElementById("sidebarBackdrop");

function setSidebar(open) {
  if (!dashboardSidebar || !sidebarBackdrop) return;
  dashboardSidebar.classList.toggle("-translate-x-full", !open);
  sidebarBackdrop.classList.toggle("hidden", !open);
  document.body.classList.toggle("overflow-hidden", open && window.innerWidth < 1024);
}

document.getElementById("openSidebar")?.addEventListener("click", () => setSidebar(true));
document.getElementById("closeSidebar")?.addEventListener("click", () => setSidebar(false));
sidebarBackdrop?.addEventListener("click", () => setSidebar(false));

document.querySelectorAll("[data-sidebar-toggle]").forEach((button) => {
  button.addEventListener("click", () => {
    const parent = button.closest("[data-sidebar-parent]");
    const submenu = parent?.querySelector("[data-sidebar-submenu]");
    const chevron = button.querySelector("[data-sidebar-chevron]");
    const opening = submenu?.classList.contains("hidden") ?? false;
    submenu?.classList.toggle("hidden", !opening);
    chevron?.classList.toggle("rotate-180", opening);
    button.setAttribute("aria-expanded", String(opening));
  });
});

const toast = document.getElementById("dashboardToast");
let toastTimer;

function showDashboardToast(message) {
  if (!toast) return;
  toast.textContent = message;
  toast.classList.remove("hidden");
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => toast.classList.add("hidden"), 2600);
}

document.querySelectorAll("[data-demo-action]").forEach((button) => {
  button.addEventListener("click", () => showDashboardToast(button.dataset.demoAction));
});

const recordRows = [...document.querySelectorAll("[data-record]")];
const recordSearch = document.getElementById("recordSearch");
const recordCount = document.getElementById("recordCount");
const recordEmpty = document.getElementById("recordEmpty");
let activeRecordFilter = "Semua";

function filterRecords() {
  const query = recordSearch?.value.trim().toLowerCase() ?? "";
  let visible = 0;
  recordRows.forEach((row) => {
    const matchesFilter = activeRecordFilter === "Semua" || row.dataset.status === activeRecordFilter;
    const matchesSearch = (row.dataset.search ?? "").includes(query);
    const shown = matchesFilter && matchesSearch;
    row.classList.toggle("hidden", !shown);
    if (shown) visible += 1;
  });
  if (recordCount) recordCount.textContent = `${visible} catatan`;
  recordEmpty?.classList.toggle("hidden", visible !== 0);
}

recordSearch?.addEventListener("input", filterRecords);
document.querySelectorAll("[data-record-filter]").forEach((button) => {
  button.addEventListener("click", () => {
    activeRecordFilter = button.dataset.recordFilter;
    document.querySelectorAll("[data-record-filter]").forEach((item) => {
      const active = item === button;
      item.classList.toggle("bg-orange-50", active);
      item.classList.toggle("font-bold", active);
      item.classList.toggle("text-orange-700", active);
      item.classList.toggle("text-slate-500", !active);
    });
    filterRecords();
  });
});

const recordModal = document.getElementById("recordModal");

function setText(id, value) {
  const element = document.getElementById(id);
  if (element) element.textContent = value;
}

function openRecordModal(row) {
  if (!recordModal) return;
  const initials = (row.dataset.name ?? "").split(/\s+/).map((part) => part[0]).slice(0, 2).join("");
  setText("recordModalTitle", row.dataset.name);
  setText("recordModalInitials", initials);
  setText("recordModalAlias", row.dataset.alias);
  setText("recordModalStatus", row.dataset.status);
  setText("recordModalLocation", row.dataset.location);
  setText("recordModalDate", row.dataset.date);
  setText("recordModalNote", row.dataset.note);
  recordModal.classList.remove("hidden");
  recordModal.classList.add("flex");
  document.body.classList.add("overflow-hidden");
}

function closeRecordModal() {
  if (!recordModal) return;
  recordModal.classList.add("hidden");
  recordModal.classList.remove("flex");
  document.body.classList.remove("overflow-hidden");
}

recordRows.forEach((row) => {
  row.addEventListener("click", () => openRecordModal(row));
  row.addEventListener("keydown", (event) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      openRecordModal(row);
    }
  });
});
document.getElementById("closeRecordModal")?.addEventListener("click", closeRecordModal);
recordModal?.addEventListener("click", (event) => {
  if (event.target === recordModal) closeRecordModal();
});
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") {
    closeRecordModal();
    setSidebar(false);
  }
});

document.querySelectorAll("[data-open-dialog]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById(button.dataset.openDialog);
    if (!dialog) return;
    dialog.showModal();
    const content = dialog.firstElementChild;
    if (content) content.scrollTop = 0;
    document.documentElement.classList.add("overflow-hidden");
    document.body.classList.add("overflow-hidden");
  });
});
document.querySelectorAll("[data-close-dialog]").forEach((button) => {
  button.addEventListener("click", () => button.closest("dialog")?.close());
});
document.querySelectorAll("dialog").forEach((dialog) => {
  dialog.addEventListener("click", (event) => {
    if (event.target === dialog) dialog.close();
  });
  dialog.addEventListener("close", () => {
    if (!document.querySelector("dialog[open]")) {
      document.documentElement.classList.remove("overflow-hidden");
      document.body.classList.remove("overflow-hidden");
    }
  });
});

document.querySelectorAll("[data-edit-role]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editRoleDialog");
    const form = document.getElementById("editRoleForm");
    if (!dialog || !form) return;
    form.action = `/settings/roles/${button.dataset.id}/update`;
    form.elements.name.value = button.dataset.name;
    form.elements.guard_name.value = button.dataset.guard;
    form.elements.description.value = button.dataset.description;
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-permission]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editPermissionDialog");
    const form = document.getElementById("editPermissionForm");
    if (!dialog || !form) return;
    form.action = `/settings/permissions/${button.dataset.id}/update`;
    form.elements.name.value = button.dataset.name;
    form.elements.guard_name.value = button.dataset.guard;
    form.elements.description.value = button.dataset.description;
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-user]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editUserDialog");
    const form = document.getElementById("editUserForm");
    if (!dialog || !form) return;
    form.action = `/settings/users/${button.dataset.id}/update`;
    form.elements.name.value = button.dataset.name ?? "";
    form.elements.employee_number.value = button.dataset.employeeNumber ?? "";
    form.elements.username.value = button.dataset.username ?? "";
    form.elements.email.value = button.dataset.email ?? "";
    form.elements.phone.value = button.dataset.phone ?? "";
    form.elements.password.value = "";
    form.elements.status.value = button.dataset.status ?? "active";
    form.elements.store_id.value = button.dataset.storeId ?? "";
    form.elements.role_id.value = button.dataset.roleId ?? "";
    form.elements.employee_number.dispatchEvent(new Event("input", { bubbles: true }));
    dialog.showModal();
  });
});

const employeeDirectory = new Map([...document.querySelectorAll("[data-employee-directory] [data-employee-number]")].map((item) => [item.dataset.employeeNumber.trim().toUpperCase(), item.dataset.employeeName]));
document.querySelectorAll('form input[name="employee_number"]').forEach((input) => {
  const output = input.form?.querySelector("[data-employee-lookup-name]");
  const updateEmployeeName = () => {
    if (output) output.value = employeeDirectory.get(input.value.trim().toUpperCase()) ?? "";
  };
  input.addEventListener("input", updateEmployeeName);
  input.addEventListener("change", updateEmployeeName);
  updateEmployeeName();
});

document.querySelectorAll("[data-edit-store]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editStoreDialog");
    const form = document.getElementById("editStoreForm");
    if (!dialog || !form) return;
    form.action = `/master/stores/${button.dataset.id}/update`;
    if (form.elements.employment_status) form.elements.employment_status.value = button.dataset.employmentStatus ?? "trainee";
    form.elements.code.value = button.dataset.code ?? "";
    form.elements.name.value = button.dataset.name ?? "";
    form.elements.address.value = button.dataset.address ?? "";
    form.elements.phone.value = button.dataset.phone ?? "";
    form.elements.timezone.value = button.dataset.timezone ?? "Asia/Jakarta";
    form.elements.status.value = button.dataset.status ?? "active";
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-aspect]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editAspectDialog");
    const form = document.getElementById("editAspectForm");
    if (!dialog || !form) return;
    form.action = `/master/assessment-aspects/${button.dataset.id}/update`;
    if (form.elements.employment_status) form.elements.employment_status.value = button.dataset.employmentStatus ?? "trainee";
    form.elements.code.value = button.dataset.code ?? "";
    form.elements.name.value = button.dataset.name ?? "";
    form.elements.sort_order.value = button.dataset.sortOrder ?? "";
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-criterion]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editCriterionDialog");
    const form = document.getElementById("editCriterionForm");
    if (!dialog || !form) return;
    form.action = `/master/assessment-criteria/${button.dataset.id}/update`;
    const positionIDs = (document.querySelector(`[data-criterion-position-directory] [data-criterion-id="${button.dataset.id}"]`)?.dataset.positionIds ?? "").split(",");
    [...form.elements.position_ids.options].forEach((option) => { option.selected = positionIDs.includes(option.value); });
    form.elements.aspect_id.value = button.dataset.aspectId ?? "";
    if (form.elements.employment_status) form.elements.employment_status.value = button.dataset.employmentStatus ?? "trainee";
    form.elements.code.value = button.dataset.code ?? "";
    form.elements.description.value = button.dataset.description ?? "";
    form.elements.weight.value = button.dataset.weight ?? "1";
    form.elements.sort_order.value = button.dataset.sortOrder ?? "";
    dialog.showModal();
  });
});

function filterRuleCriteria(form, aspectId, selectedCriterion = "") {
  const criterion = form?.querySelector("[data-rule-criterion]");
	const rubric = form?.querySelector("[data-rule-rubric]");
	const departmentId = rubric?.selectedOptions[0]?.dataset.departmentId ?? "";
  if (!criterion) return;
	const available = Boolean(aspectId && departmentId);
  criterion.disabled = !available;
  const placeholder = criterion.options[0];
  if (placeholder) placeholder.textContent = available ? "Pilih subaspek" : "Pilih aspek terlebih dahulu";
  [...criterion.options].slice(1).forEach((option) => {
		const visible = option.dataset.aspectId === String(aspectId) && option.dataset.departmentId === departmentId;
    option.hidden = !visible;
    option.disabled = !visible;
  });
  criterion.value = selectedCriterion;
  if (!criterion.value) criterion.value = "";
}

document.querySelectorAll("[data-rule-form]").forEach((form) => {
  const aspect = form.querySelector("[data-rule-aspect]");
	const rubric = form.querySelector("[data-rule-rubric]");
  aspect?.addEventListener("change", () => filterRuleCriteria(form, aspect.value));
	rubric?.addEventListener("change", () => filterRuleCriteria(form, aspect?.value ?? ""));
  filterRuleCriteria(form, aspect?.value ?? "");
});

document.querySelectorAll("[data-edit-rule]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editRuleDialog");
    const form = document.getElementById("editRuleForm");
    if (!dialog || !form) return;
		form.action = `/master/assessment-score-rules/${button.dataset.rubricId}/${button.dataset.criterionId}/${button.dataset.score}/update`;
		form.elements.rubric_id.value = button.dataset.rubricId ?? "";
    const criterionID = button.dataset.criterionId ?? "";
    const criterionOption = [...form.elements.criterion_id.options].find((option) => option.value === criterionID);
    const aspect = form.querySelector("[data-rule-aspect]");
    if (aspect) aspect.value = criterionOption?.dataset.aspectId ?? "";
    filterRuleCriteria(form, aspect?.value ?? "", criterionID);
    form.elements.score.value = button.dataset.score ?? "";
    form.elements.finding_min.value = button.dataset.findingMin ?? "";
    form.elements.finding_max.value = button.dataset.findingMax ?? "";
    form.elements.description.value = button.dataset.description ?? "";
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-organization]").forEach((button) => {
  button.addEventListener("click", () => {
    const names = {
      "job-level": ["editJobLevelDialog", "editJobLevelForm", "job-levels"],
      position: ["editPositionDialog", "editPositionForm", "positions"],
      department: ["editDepartmentDialog", "editDepartmentForm", "departments"],
    };
    const config = names[button.dataset.kind];
    if (!config) return;
    const dialog = document.getElementById(config[0]);
    const form = document.getElementById(config[1]);
    if (!dialog || !form) return;
    form.action = `/master/${config[2]}/${button.dataset.id}/update`;
    if (form.elements.employment_status) form.elements.employment_status.value = button.dataset.employmentStatus ?? "trainee";
    form.elements.code.value = button.dataset.code ?? "";
    form.elements.name.value = button.dataset.name ?? "";
    if (form.elements.rank_order) form.elements.rank_order.value = button.dataset.rankOrder ?? "";
    if (form.elements.job_level_id) form.elements.job_level_id.value = button.dataset.jobLevelId ?? "";
    if (form.elements.department_id && button.dataset.kind === "position") {
      const mapping = document.querySelector(`[data-position-departments] [data-position-id="${button.dataset.id}"]`);
      form.elements.department_id.value = mapping?.dataset.departmentId ?? "";
    }
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-employee]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editEmployeeDialog");
    const form = document.getElementById("editEmployeeForm");
    if (!dialog || !form) return;
    form.action = `/master/employees/${button.dataset.id}/update`;
    form.elements.employee_number.value = button.dataset.number ?? "";
    form.elements.name.value = button.dataset.name ?? "";
    form.elements.email.value = button.dataset.email ?? "";
    form.elements.phone.value = button.dataset.phone ?? "";
    form.elements.status.value = button.dataset.status ?? "active";
    const details = document.querySelector(`[data-employee-details] [data-employee-id="${button.dataset.id}"]`);
    const values = { birth_place: "birthPlace", birth_date: "birthDate", gender: "gender", address: "address", joined_date: "joinedDate", employment_ended_at: "employmentEndedAt", employment_status: "employmentStatus", emergency_contact_name: "emergencyName", emergency_contact_phone: "emergencyPhone", emergency_contact_relationship: "emergencyRelationship" };
    Object.entries(values).forEach(([field, data]) => { form.elements[field].value = details?.dataset[data] ?? ""; });
    form.elements.department_id.value = details?.dataset.departmentId ?? "";
    form.elements.department_id.dispatchEvent(new Event("change", { bubbles: true }));
    form.elements.position_id.value = details?.dataset.positionId ?? "";
    form.elements.store_id.value = details?.dataset.storeId ?? "";
    form.elements.assignment_started_at.value = details?.dataset.assignmentStartedAt ?? "";
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-profile]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editProfileDialog");
    const form = document.getElementById("editProfileForm");
    if (!dialog || !form) return;
    form.action = `/master/employee-profiles/${button.dataset.id}/update`;
    const values = { birth_place: "birthPlace", birth_date: "birthDate", gender: "gender", address: "address", joined_date: "joinedDate", employment_ended_at: "employmentEndedAt", employment_status: "employmentStatus", photo_path: "photoPath", emergency_contact_name: "emergencyName", emergency_contact_phone: "emergencyPhone", emergency_contact_relationship: "emergencyRelationship" };
    Object.entries(values).forEach(([field, data]) => { form.elements[field].value = button.dataset[data] ?? ""; });
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-assignment]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editAssignmentDialog");
    const form = document.getElementById("editAssignmentForm");
    if (!dialog || !form) return;
    form.action = `/master/employee-assignments/${button.dataset.id}/update`;
    const values = { employee_id: "employeeId", position_id: "positionId", department_id: "departmentId", store_id: "storeId", started_at: "startedAt" };
    Object.entries(values).forEach(([field, data]) => { form.elements[field].value = button.dataset[data] ?? ""; });
    dialog.showModal();
  });
});

document.querySelectorAll("[data-assignment-fields]").forEach((fields) => {
  const position = fields.querySelector('[name="position_id"]');
  const department = fields.querySelector('[name="department_id"]');
  if (!position || !department) return;
  const syncDepartment = () => {
    department.value = position.selectedOptions[0]?.dataset.departmentId ?? "";
  };
  const filterPositions = () => {
    const departmentID = department.value;
    [...position.options].slice(1).forEach((option) => {
      const visible = !departmentID || option.dataset.departmentId === departmentID;
      option.hidden = !visible;
      option.disabled = !visible;
    });
    if (position.value && position.selectedOptions[0]?.dataset.departmentId !== departmentID) position.value = "";
  };
  position.addEventListener("change", syncDepartment);
  department.addEventListener("change", filterPositions);
  filterPositions();
});

document.querySelectorAll("[data-edit-reporting]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editReportingDialog");
    const form = document.getElementById("editReportingForm");
    if (!dialog || !form) return;
    form.action = `/master/employee-reporting-lines/${button.dataset.id}/update`;
    form.elements.employee_assignment_id.value = button.dataset.employeeAssignmentId ?? "";
    form.elements.supervisor_assignment_id.value = button.dataset.supervisorAssignmentId ?? "";
    form.elements.started_at.value = button.dataset.startedAt ?? "";
    form.elements.ended_at.value = button.dataset.endedAt ?? "";
    dialog.showModal();
  });
});

const assessmentForm = document.getElementById("assessmentForm");
if (assessmentForm) {
  const trainee = document.getElementById("assessmentTrainee");
  const assessmentDate = assessmentForm.elements.assessment_date;
  if (assessmentDate && !assessmentDate.value) assessmentDate.value = new Date().toISOString().slice(0, 10);

  const enhanceTraineeSelect = () => {
    if (!trainee || trainee.dataset.enhanced === "true") return;
    trainee.dataset.enhanced = "true";
    trainee.required = false;

    const options = [...trainee.options].filter((option) => option.value);
    const wrapper = document.createElement("div");
    wrapper.className = "relative mt-2";
    wrapper.innerHTML = `
      <div class="relative">
        <i class="fa-solid fa-magnifying-glass pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-sm text-slate-400"></i>
        <input type="text" data-trainee-search autocomplete="off" role="combobox" aria-autocomplete="list" aria-expanded="false" aria-controls="traineeAutocompleteList" required placeholder="Cari NIK atau nama trainee..." class="w-full rounded-xl border border-slate-200 bg-white py-3 pl-10 pr-10 text-sm font-normal outline-none transition focus:border-orange-400 focus:ring-4 focus:ring-orange-100">
        <button type="button" data-trainee-clear class="absolute right-2 top-1/2 hidden size-8 -translate-y-1/2 place-items-center rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600" aria-label="Hapus pilihan"><i class="fa-solid fa-xmark"></i></button>
      </div>
      <div id="traineeAutocompleteList" data-trainee-options role="listbox" class="absolute z-40 mt-2 hidden max-h-72 w-full overflow-y-auto rounded-xl border border-slate-200 bg-white p-1.5 shadow-xl shadow-slate-900/10"></div>`;
    trainee.before(wrapper);
    trainee.classList.add("hidden");

    const search = wrapper.querySelector("[data-trainee-search]");
    const list = wrapper.querySelector("[data-trainee-options]");
    const clear = wrapper.querySelector("[data-trainee-clear]");
    let activeIndex = -1;
    let visibleOptions = [];

    const closeList = () => {
      list.classList.add("hidden");
      search.setAttribute("aria-expanded", "false");
      activeIndex = -1;
    };

    const choose = (option) => {
      trainee.value = option.value;
      search.value = option.textContent.trim();
      search.setCustomValidity("");
      clear.classList.remove("hidden");
      clear.classList.add("grid");
      trainee.dispatchEvent(new Event("change", { bubbles: true }));
      closeList();
    };

    const render = () => {
      const query = search.value.trim().toLowerCase();
      visibleOptions = options.filter((option) => option.textContent.toLowerCase().includes(query));
      list.replaceChildren();
      visibleOptions.slice(0, 50).forEach((option, index) => {
        const button = document.createElement("button");
        button.type = "button";
        button.setAttribute("role", "option");
        button.className = "flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left text-xs text-slate-700 hover:bg-orange-50 hover:text-orange-700";
        button.innerHTML = '<span class="grid size-8 shrink-0 place-items-center rounded-lg bg-orange-100 font-bold text-orange-700"><i class="fa-solid fa-user-graduate"></i></span><span class="min-w-0"><strong class="block truncate"></strong><small class="mt-0.5 block truncate text-[10px] text-slate-400"></small></span>';
        button.querySelector("strong").textContent = option.textContent.trim();
        button.querySelector("small").textContent = `${option.dataset.position || "—"} · ${option.dataset.department || "—"} · ${option.dataset.store || "—"}`;
        button.addEventListener("mousedown", (event) => event.preventDefault());
        button.addEventListener("click", () => choose(option));
        list.append(button);
        if (index === activeIndex) {
          button.classList.add("bg-orange-50", "text-orange-700");
          button.scrollIntoView({ block: "nearest" });
        }
      });
      if (!visibleOptions.length) {
        const empty = document.createElement("p");
        empty.className = "px-3 py-5 text-center text-xs text-slate-400";
        empty.textContent = "Trainee tidak ditemukan.";
        list.append(empty);
      }
      list.classList.remove("hidden");
      search.setAttribute("aria-expanded", "true");
    };

    search.addEventListener("focus", render);
    search.addEventListener("input", () => {
      if (trainee.value) {
        trainee.value = "";
        trainee.dispatchEvent(new Event("change", { bubbles: true }));
      }
      clear.classList.add("hidden");
      clear.classList.remove("grid");
      search.setCustomValidity("Pilih trainee dari hasil pencarian.");
      activeIndex = -1;
      render();
    });
    search.addEventListener("keydown", (event) => {
      if (event.key === "ArrowDown" || event.key === "ArrowUp") {
        event.preventDefault();
        if (list.classList.contains("hidden")) render();
        const last = Math.min(visibleOptions.length, 50) - 1;
        activeIndex = event.key === "ArrowDown" ? Math.min(activeIndex + 1, last) : Math.max(activeIndex - 1, 0);
        render();
      } else if (event.key === "Enter" && activeIndex >= 0 && visibleOptions[activeIndex]) {
        event.preventDefault();
        choose(visibleOptions[activeIndex]);
      } else if (event.key === "Escape") closeList();
    });
    clear.addEventListener("click", () => {
      trainee.value = "";
      search.value = "";
      search.setCustomValidity("Pilih trainee dari hasil pencarian.");
      clear.classList.add("hidden");
      clear.classList.remove("grid");
      trainee.dispatchEvent(new Event("change", { bubbles: true }));
      search.focus();
      render();
    });
    document.addEventListener("click", (event) => {
      if (!wrapper.contains(event.target)) closeList();
    });

    const selected = trainee.selectedOptions[0];
    if (selected?.value) choose(selected);
  };

  enhanceTraineeSelect();

  const updateTraineeSummary = () => {
    const option = trainee?.selectedOptions[0];
    const values = { number: "number", position: "position", department: "department", store: "store" };
    Object.entries(values).forEach(([target, data]) => {
      const element = assessmentForm.querySelector(`[data-trainee-${target}]`);
      if (element) element.textContent = option?.dataset[data] || "—";
    });
  };
  trainee?.addEventListener("change", updateTraineeSummary);
  updateTraineeSummary();

  const updateAssessmentTotal = () => {
    const groups = [...new Set([...assessmentForm.querySelectorAll('input[type="radio"][name^="score_"]:enabled')].map((input) => input.name))];
    let total = 0;
    let totalWeight = 0;
    let filled = 0;
    groups.forEach((name) => {
      const selected = assessmentForm.querySelector(`input[type="radio"][name="${name}"]:checked`);
      if (!selected) return;
      const criterionWeight = Number(selected.closest('[data-rubric-row]').dataset.weight);
      total += Number(selected.value) * criterionWeight;
      totalWeight += criterionWeight;
      filled += 1;
			const criterionKey = name.replace("score_", "");
			const description = assessmentForm.querySelector(`[data-score-description="${criterionKey}"]`);
      if (description) description.textContent = selected.dataset.ruleDescription || "";
    });
    const average = filled === groups.length && totalWeight ? Math.round(total / totalWeight * 100) / 100 : 0;
    const category = [...assessmentForm.querySelectorAll("[data-result-category]")].find((item) => average >= Number(item.dataset.minimum) && average <= Number(item.dataset.maximum));
    const rating = average <= 0 ? "Belum lengkap" : category?.dataset.name ?? "Kategori belum diatur";
    const weight = assessmentForm.querySelector("[data-assessment-weight]");
    const averageElement = assessmentForm.querySelector("[data-assessment-average]");
    const ratingElement = assessmentForm.querySelector("[data-assessment-rating]");
    if (weight) weight.textContent = `${filled}/${groups.length}`;
    if (averageElement) averageElement.textContent = average.toFixed(2);
    if (ratingElement) ratingElement.textContent = rating;
  };
  assessmentForm.querySelectorAll('input[type="radio"][name^="score_"]').forEach((input) => input.addEventListener("change", updateAssessmentTotal));
  const updateRubric = () => {
    const department = trainee?.selectedOptions[0]?.dataset.departmentId;
    const month = Number(assessmentForm.elements.training_month.value);
    const rows = [...assessmentForm.querySelectorAll('[data-rubric-row]')];
    const available = rows.filter(row => row.dataset.employmentStatus === trainee?.selectedOptions[0]?.dataset.employmentStatus && row.dataset.departmentId === department && month >= Number(row.dataset.monthFrom));
    const exact = available.filter(row => !Number(row.dataset.monthTo) || month <= Number(row.dataset.monthTo));
    const latestStart = available.length ? Math.max(...available.map(row => Number(row.dataset.monthFrom))) : 0;
    const fallback = !exact.length && latestStart > 0;
    const matches = exact.length ? exact : available.filter(row => Number(row.dataset.monthFrom) === latestStart);
    const rubrics = new Set(matches.map(row => row.dataset.rubricId));
    const complete = matches.length > 0 && matches.every(row => row.querySelectorAll('input[type="radio"]').length === 5);
    const valid = rubrics.size === 1 && complete && month >= 1 && month <= 24;
    rows.forEach(row => {
      const visible = valid && matches.includes(row);
      row.hidden = !visible;
      row.querySelectorAll('input').forEach(input => {
        input.disabled = !visible;
        if (!visible && input.type === 'radio') input.checked = false;
      });
      if (visible) row.cells[0].textContent = String(matches.indexOf(row) + 1);
    });
    const message = assessmentForm.querySelector('[data-rubric-message]');
		if (message) message.textContent = !trainee?.value ? 'Pilih trainee dan bulan pelatihan untuk menampilkan rubrik.' : valid ? `Rubrik aktif: ${matches[0].dataset.rubricName}${fallback ? ' (menggunakan periode terakhir)' : ''}` : 'Rubrik jabatan ini belum tersedia atau belum lengkap. Hubungi pengelola master penilaian.';
    assessmentForm.querySelector('[data-submit-final]').disabled = !valid;
    updateAssessmentTotal();
  };
  trainee?.addEventListener('change', updateRubric);
  assessmentForm.elements.training_month.addEventListener('input', updateRubric);

  const placementStart = document.createElement("span");
  placementStart.innerHTML = 'Mulai jabatan: <strong data-trainee-started-at>—</strong>';
  const traineeSummary = assessmentForm.querySelector("[data-trainee-store]")?.closest("div");
  traineeSummary?.classList.replace("lg:grid-cols-4", "lg:grid-cols-5");
  traineeSummary?.append(placementStart);
  const traineeStarts = new Map([...assessmentForm.querySelectorAll("[data-trainee-start-record]")].map((item) => [item.dataset.employeeId, item.dataset.startedAt]));
  const updateTrainingMonth = (recalculate = true) => {
    const startedAt = traineeStarts.get(trainee?.value ?? "") ?? "";
    const startElement = assessmentForm.querySelector("[data-trainee-started-at]");
    if (startElement) startElement.textContent = startedAt ? new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "long", year: "numeric" }).format(new Date(`${startedAt}T00:00:00`)) : "—";
    if (!recalculate || !startedAt || !assessmentDate?.value) return;
    const start = new Date(`${startedAt}T00:00:00`);
    const assessed = new Date(`${assessmentDate.value}T00:00:00`);
    const calculated = (assessed.getFullYear() - start.getFullYear()) * 12 + assessed.getMonth() - start.getMonth() + 1;
    assessmentForm.elements.training_month.value = String(Math.min(24, Math.max(1, calculated)));
    assessmentForm.elements.training_month.dispatchEvent(new Event("input", { bubbles: true }));
  };
  if (assessmentDate && !assessmentDate.value) {
    const now = new Date();
    assessmentDate.value = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
  }
  trainee?.addEventListener("change", updateTrainingMonth);
  assessmentDate?.addEventListener("change", updateTrainingMonth);
  updateTrainingMonth(!assessmentForm.action.endsWith("/update"));
  updateRubric();

  assessmentForm.querySelector("[data-submit-final]")?.addEventListener("click", (event) => {
    const total = assessmentForm.querySelectorAll('input[type="radio"][name^="score_"]:enabled:checked').length;
    const expected = new Set([...assessmentForm.querySelectorAll('input[type="radio"][name^="score_"]:enabled')].map((input) => input.name)).size;
    if (!expected || total !== expected || !assessmentForm.elements.notes.value.trim()) {
      event.preventDefault();
      showDashboardToast("Lengkapi semua nilai dan catatan sebelum diajukan.");
    }
  });
}

document.querySelectorAll("[data-user-store-row]").forEach((row) => {
  const checkbox = row.querySelector("[data-user-store-checkbox]");
  const role = row.querySelector("[data-user-store-role]");
  const defaultRadio = row.querySelector("[data-user-store-default]");
  checkbox?.addEventListener("change", () => {
    const enabled = checkbox.checked;
    if (role) role.disabled = !enabled;
    if (defaultRadio) {
      defaultRadio.disabled = !enabled;
      if (!enabled) defaultRadio.checked = false;
    }
    if (enabled && !document.querySelector("[data-user-store-default]:checked")) defaultRadio.checked = true;
  });
});

document.querySelectorAll("form[data-confirm]").forEach((form) => {
  const deleteButton = form.querySelector("button");
  if (deleteButton && !deleteButton.title) deleteButton.title = deleteButton.disabled ? "Data dilindungi" : "Hapus";
  form.addEventListener("submit", async (event) => {
    event.preventDefault();

    if (typeof Swal === "undefined") {
      if (window.confirm(form.dataset.confirm)) form.submit();
      return;
    }

    const result = await Swal.fire({
      title: "Hapus data?",
      text: form.dataset.confirm,
      icon: "warning",
      showCancelButton: true,
      confirmButtonText: "Ya, hapus",
      cancelButtonText: "Batal",
      confirmButtonColor: "#ef4444",
      cancelButtonColor: "#64748b",
      reverseButtons: true,
      focusCancel: true,
    });

    if (result.isConfirmed) form.submit();
  });
});

document.querySelectorAll("[data-edit-role], [data-edit-permission], [data-edit-user], [data-edit-store], [data-edit-aspect], [data-edit-criterion], [data-edit-rule], [data-edit-organization], [data-edit-employee], [data-edit-profile], [data-edit-assignment], [data-edit-reporting]").forEach((button) => {
  button.title = "Edit";
});

if (["/master/employees", "/master/employee-assignments"].includes(window.location.pathname)) {
  const table = document.querySelector("[data-filter-table] table");
  const heading = table?.querySelector("thead tr");
  if (table && heading && !heading.querySelector("[data-generated-number-heading]")) {
    const th = document.createElement("th");
    th.dataset.generatedNumberHeading = "";
    th.className = "w-16 px-5 py-3 text-center";
    th.textContent = "No.";
    heading.prepend(th);
    table.querySelectorAll("[data-filter-row]").forEach((row) => {
      const td = document.createElement("td");
      td.className = "px-5 py-4 text-center";
      td.innerHTML = '<span data-row-number class="inline-grid size-7 place-items-center rounded-lg bg-slate-100 text-xs font-bold text-slate-600"></span>';
      row.prepend(td);
    });
    table.querySelectorAll("td[colspan]").forEach((cell) => {
      const colspan = Number.parseInt(cell.getAttribute("colspan") ?? "", 10);
      if (Number.isFinite(colspan)) cell.setAttribute("colspan", String(colspan + 1));
    });
  }
}

document.querySelectorAll("[data-access-filter]").forEach((filter) => {
  const nextElement = filter.nextElementSibling;
  const table = nextElement?.matches("[data-filter-table]")
    ? nextElement
    : filter.parentElement?.querySelector("[data-filter-table]");
  if (!table?.matches("[data-filter-table]")) return;

  const search = filter.querySelector("[data-filter-search]");
  const guard = filter.querySelector("[data-filter-guard]");
  const extra = filter.querySelector("[data-filter-extra]");
  const department = filter.querySelector("[data-filter-department]");
  const status = filter.querySelector("[data-filter-status]");
  const employment = filter.querySelector("[data-filter-employment]");
  const supervisor = filter.querySelector("[data-filter-supervisor]");
  const count = filter.querySelector("[data-filter-count]");
  const rows = [...table.querySelectorAll("[data-filter-row]")];
  const empty = table.querySelector("[data-filter-empty]");
  const previousSection = table.previousElementSibling;
  const mobileList = previousSection?.querySelector("[data-filter-row]") ? previousSection : null;
  const mobileRows = [...mobileList?.querySelectorAll("[data-filter-row]") ?? []];
  const mobileEmpty = mobileList?.querySelector("[data-filter-empty]");
  const pageSize = 25;
  let currentPage = 1;

  const pagination = document.createElement("nav");
  pagination.setAttribute("aria-label", "Navigasi halaman tabel");
  pagination.className = "mt-4 flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white px-4 py-3 shadow-sm sm:flex-row sm:items-center sm:justify-between";
  pagination.innerHTML = '<p data-pagination-info class="text-center text-xs text-slate-500 sm:text-left"></p><div data-pagination-buttons class="flex flex-wrap items-center justify-center gap-1.5"></div>';
  table.insertAdjacentElement("afterend", pagination);
  const paginationInfo = pagination.querySelector("[data-pagination-info]");
  const paginationButtons = pagination.querySelector("[data-pagination-buttons]");

  const existingOptions = new Set([...guard?.options ?? []].map((option) => option.value));
  [...new Set(rows.map((row) => row.dataset.filterGuard).filter((value) => value && !existingOptions.has(value)))]
    .sort((left, right) => left.localeCompare(right))
    .forEach((value) => guard?.add(new Option(value, value)));

  const populateFilterOptions = (select, dataKey) => {
    if (!select) return;
    const existing = new Set([...select.options].map((option) => option.value));
    [...new Set(rows.map((row) => row.dataset[dataKey]).filter((value) => value && !existing.has(value)))]
      .sort((left, right) => left.localeCompare(right))
      .forEach((value) => select.add(new Option(value, value)));
  };
  populateFilterOptions(extra, "filterExtra");
  populateFilterOptions(department, "filterDepartment");
  populateFilterOptions(status, "filterStatus");
  if (supervisor?.matches("select")) populateFilterOptions(supervisor, "filterSupervisor");

  const extraDependsOnDepartment = extra?.getAttribute("aria-label") === "Filter jabatan";
  const updateDependentExtra = () => {
    if (!extraDependsOnDepartment || !extra) return;
    const selectedDepartment = department?.value ?? "";
    const available = new Set(rows
      .filter((row) => !selectedDepartment || row.dataset.filterDepartment === selectedDepartment)
      .map((row) => row.dataset.filterExtra)
      .filter(Boolean));
    [...extra.options].slice(1).forEach((option) => {
      const visible = !selectedDepartment || available.has(option.value);
      option.hidden = !visible;
      option.disabled = !visible;
    });
    if (extra.value && !available.has(extra.value)) extra.value = "";
  };
  updateDependentExtra();

  function renderPagination(totalItems, totalPages) {
    const start = totalItems === 0 ? 0 : ((currentPage - 1) * pageSize) + 1;
    const end = Math.min(currentPage * pageSize, totalItems);
    if (paginationInfo) paginationInfo.textContent = `Menampilkan ${start}–${end} dari ${totalItems} data`;
    if (!paginationButtons) return;
    paginationButtons.replaceChildren();

    const makeButton = (label, page, options = {}) => {
      const button = document.createElement("button");
      button.type = "button";
      button.disabled = options.disabled ?? false;
      button.setAttribute("aria-label", options.ariaLabel ?? `Halaman ${page}`);
      if (options.current) button.setAttribute("aria-current", "page");
      button.className = options.current
        ? "grid min-w-9 place-items-center rounded-lg bg-orange-500 px-2.5 py-2 text-xs font-bold text-white shadow-sm"
        : "grid min-w-9 place-items-center rounded-lg border border-slate-200 bg-white px-2.5 py-2 text-xs font-bold text-slate-600 transition hover:border-orange-200 hover:bg-orange-50 hover:text-orange-700 disabled:cursor-not-allowed disabled:opacity-40";
      button.innerHTML = label;
      button.addEventListener("click", () => { currentPage = page; applyAccessFilter(false); });
      return button;
    };

    paginationButtons.append(makeButton('<i class="fa-solid fa-chevron-left" aria-hidden="true"></i>', Math.max(1, currentPage - 1), { disabled: currentPage === 1, ariaLabel: "Halaman sebelumnya" }));
    const pages = [];
    for (let page = 1; page <= totalPages; page += 1) {
      if (page === 1 || page === totalPages || Math.abs(page - currentPage) <= 1) pages.push(page);
    }
    let previousPage = 0;
    pages.forEach((page) => {
      if (page - previousPage > 1) {
        const separator = document.createElement("span");
        separator.className = "px-1 text-xs text-slate-400";
        separator.textContent = "…";
        paginationButtons.append(separator);
      }
      paginationButtons.append(makeButton(String(page), page, { current: page === currentPage }));
      previousPage = page;
    });
    paginationButtons.append(makeButton('<i class="fa-solid fa-chevron-right" aria-hidden="true"></i>', Math.min(totalPages, currentPage + 1), { disabled: currentPage === totalPages, ariaLabel: "Halaman berikutnya" }));
  }

  function applyAccessFilter(resetPage = false) {
    if (resetPage) currentPage = 1;
    const query = search?.value.trim().toLowerCase() ?? "";
    const selectedGuard = guard?.value ?? "";
    const selectedExtra = extra?.value ?? "";
    const selectedDepartment = department?.value ?? "";
    const selectedStatus = status?.value ?? "";
    const selectedSupervisor = supervisor?.value.trim().toLowerCase() ?? "";
    const filteredRows = rows.filter((row) => {
      const matchesSearch = (row.dataset.filterSearch ?? "").toLowerCase().includes(query);
      const matchesGuard = !selectedGuard || row.dataset.filterGuard === selectedGuard;
      const matchesExtra = !selectedExtra || row.dataset.filterExtra === selectedExtra;
      const matchesDepartment = !selectedDepartment || row.dataset.filterDepartment === selectedDepartment;
      const matchesStatus = !selectedStatus || row.dataset.filterStatus === selectedStatus;
      const matchesSupervisor = !selectedSupervisor || (row.dataset.filterSupervisor ?? "").toLowerCase().includes(selectedSupervisor);
      const matchesEmployment = !employment?.value || row.dataset.filterEmployment === employment.value;
      return matchesEmployment && matchesSearch && matchesGuard && matchesExtra && matchesDepartment && matchesStatus && matchesSupervisor;
    });
    const totalPages = Math.max(1, Math.ceil(filteredRows.length / pageSize));
    currentPage = Math.min(currentPage, totalPages);
    const pageStart = (currentPage - 1) * pageSize;
    const pageRows = new Set(filteredRows.slice(pageStart, pageStart + pageSize));
    rows.forEach((row) => row.classList.toggle("hidden", !pageRows.has(row)));
    const visibleIndexes = new Set(rows.flatMap((row, index) => pageRows.has(row) ? [index] : []));
    mobileRows.forEach((row, index) => row.classList.toggle("hidden", !visibleIndexes.has(index)));
    filteredRows.forEach((row, index) => {
      const number = row.querySelector("[data-row-number]") ?? (mobileRows.length > 0 ? row.cells?.[0] : null);
      if (number) number.textContent = String(index + 1);
      const mobileRow = mobileRows[rows.indexOf(row)];
      const mobileNumber = mobileRow?.querySelector("[data-row-number]") ?? mobileRow?.querySelector(".flex > span:first-child");
      if (mobileNumber) mobileNumber.textContent = String(index + 1);
    });

    if (count) count.textContent = `${filteredRows.length} data`;
    empty?.classList.toggle("hidden", filteredRows.length !== 0 || rows.length === 0);
    mobileEmpty?.classList.toggle("hidden", filteredRows.length !== 0 || mobileRows.length === 0);
    renderPagination(filteredRows.length, totalPages);
  }

  search?.addEventListener("input", () => applyAccessFilter(true));
  guard?.addEventListener("change", () => applyAccessFilter(true));
  extra?.addEventListener("change", () => applyAccessFilter(true));
  department?.addEventListener("change", () => {
    updateDependentExtra();
    applyAccessFilter(true);
  });
  employment?.addEventListener("change", () => applyAccessFilter(true));
  status?.addEventListener("change", () => applyAccessFilter(true));
  supervisor?.addEventListener(supervisor.matches("select") ? "change" : "input", () => applyAccessFilter(true));
  applyAccessFilter();
});

document.querySelectorAll("[data-permission-select]").forEach((button) => {
  button.addEventListener("click", () => {
    const checked = button.dataset.permissionSelect === "all";
    document.querySelectorAll('input[name="permission_ids"]').forEach((input) => { input.checked = checked; });
    updatePermissionCounters();
  });
});

function updatePermissionCounters() {
  const inputs = [...document.querySelectorAll('input[name="permission_ids"]')];
  const total = inputs.filter((input) => input.checked).length;
  const totalElement = document.getElementById("assignedPermissionCount");
  if (totalElement) totalElement.textContent = String(total);

  document.querySelectorAll("[data-permission-group]").forEach((group) => {
    const assigned = [...group.querySelectorAll('input[name="permission_ids"]')].filter((input) => input.checked).length;
    const counter = group.querySelector("[data-group-assigned]");
    if (counter) counter.textContent = String(assigned);
  });
}

document.querySelectorAll('input[name="permission_ids"]').forEach((input) => {
  input.addEventListener("change", updatePermissionCounters);
});

document.querySelectorAll("[data-group-select]").forEach((button) => {
  button.addEventListener("click", () => {
    const group = button.closest("[data-permission-group]");
    const checked = button.dataset.groupSelect === "all";
    group?.querySelectorAll('input[name="permission_ids"]').forEach((input) => { input.checked = checked; });
    updatePermissionCounters();
  });
});

document.getElementById("permissionSearch")?.addEventListener("input", (event) => {
  const query = event.target.value.trim().toLowerCase();
  let visibleTotal = 0;
  document.querySelectorAll("[data-permission-group]").forEach((group) => {
    let groupVisible = 0;
    group.querySelectorAll("[data-permission-item]").forEach((item) => {
      const visible = (item.dataset.search ?? "").toLowerCase().includes(query);
      item.classList.toggle("hidden", !visible);
      if (visible) groupVisible += 1;
    });
    group.classList.toggle("hidden", groupVisible === 0);
    visibleTotal += groupVisible;
  });
  document.getElementById("permissionEmpty")?.classList.toggle("hidden", visibleTotal !== 0);
});

const profileMenuButton = document.querySelector("[data-profile-menu-button]");
const profileMenu = document.querySelector("[data-profile-menu]");
const notificationWrapper = document.querySelector("[data-notification-wrapper]");
const notificationButton = document.querySelector("[data-notification-button]");
const notificationMenu = document.querySelector("[data-notification-menu]");

notificationWrapper?.addEventListener("toggle", () => {
  if (!notificationWrapper.open) return;
  profileMenu?.classList.add("hidden");
  profileMenuButton?.setAttribute("aria-expanded", "false");
});

document.querySelector("[data-notification-read-all]")?.addEventListener("click", () => {
  document.querySelectorAll("[data-notification-dot]").forEach((dot) => dot.classList.add("hidden"));
  document.querySelector("[data-notification-badge]")?.classList.add("hidden");
  showDashboardToast("Semua notifikasi ditandai sudah dibaca.");
});

profileMenuButton?.addEventListener("click", (event) => {
  event.stopPropagation();
  const opening = profileMenu?.classList.contains("hidden");
  profileMenu?.classList.toggle("hidden", !opening);
  profileMenuButton.setAttribute("aria-expanded", String(opening));
  notificationWrapper?.removeAttribute("open");
});

document.addEventListener("click", (event) => {
  if (profileMenu && profileMenuButton && !profileMenu.contains(event.target) && !profileMenuButton.contains(event.target)) {
    profileMenu.classList.add("hidden");
    profileMenuButton.setAttribute("aria-expanded", "false");
  }
  if (notificationWrapper && !notificationWrapper.contains(event.target)) notificationWrapper.removeAttribute("open");
});

document.querySelectorAll("[data-edit-category]").forEach((button) => {
  button.textContent = "";
  button.title = "Edit kategori";
  button.setAttribute("aria-label", "Edit kategori");
  button.classList.remove("px-3", "py-2");
  button.classList.add("grid", "size-9", "cursor-pointer", "place-items-center", "p-0");
  const icon = document.createElement("i");
  icon.className = "fa-solid fa-pen";
  icon.setAttribute("aria-hidden", "true");
  button.append(icon);
  button.addEventListener("click", () => {
    const dialog = document.querySelector("#editCategoryDialog");
    const form = document.querySelector("#editCategoryForm");
    if (!dialog || !form) return;
    form.action = `/master/assessment-result-categories/${button.dataset.id}/update`;
    form.elements.name.value = button.dataset.name ?? "";
    form.elements.minimum_score.value = button.dataset.minimum ?? "";
    form.elements.maximum_score.value = button.dataset.maximum ?? "";
    form.elements.deduction_percentage.value = button.dataset.deduction ?? "";
    form.elements.sort_order.value = button.dataset.order ?? "";
    form.elements.is_active.checked = button.dataset.active === "true";
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-rubric]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.querySelector("#editRubricDialog");
    const form = document.querySelector("#editRubricForm");
    if (!dialog || !form) return;
    form.action = `/master/assessment-rubrics/${button.dataset.id}/update`;
    form.elements.position_id.value = button.dataset.positionId ?? "";
    form.elements.month_from.value = button.dataset.monthFrom ?? "";
    form.elements.month_to.value = button.dataset.monthTo ?? "";
    form.elements.is_active.checked = button.dataset.active === "true";
    dialog.showModal();
  });
});

document.querySelectorAll('form[action^="/master/assessment-result-categories/"][action$="/delete"] button').forEach((button) => {
  button.textContent = "";
  button.title = "Hapus kategori";
  button.setAttribute("aria-label", "Hapus kategori");
  button.classList.remove("px-3", "py-2");
  button.classList.add("grid", "size-9", "cursor-pointer", "place-items-center", "p-0");
  const icon = document.createElement("i");
  icon.className = "fa-solid fa-trash";
  icon.setAttribute("aria-hidden", "true");
  button.append(icon);
});

const assessmentCategoryNodes = [...document.querySelectorAll("[data-result-category]")];
if (assessmentCategoryNodes.length) {
  const categoryHeading = [...document.querySelectorAll("h2")].find((item) => item.textContent.includes("Keterangan hasil penilaian"));
  const categoryGrid = categoryHeading?.parentElement?.querySelector(".mt-3.grid");
  if (categoryGrid) {
    categoryGrid.replaceChildren(...assessmentCategoryNodes.map((item, index) => {
      const entry = document.createElement("span");
      const number = document.createElement("strong");
      number.textContent = `${index + 1}. `;
      const detail = document.createElement("small");
      detail.textContent = `${Number(item.dataset.minimum).toFixed(2)}–${Number(item.dataset.maximum).toFixed(2)} · Potongan ${Number(item.dataset.deduction).toFixed(0)}%`;
      entry.append(number, document.createTextNode(item.dataset.name ?? ""), document.createElement("br"), detail);
      return entry;
    }));
  }
}

if (["/master/assessment-criteria", "/master/assessment-score-rules"].includes(window.location.pathname)) {
  document.querySelectorAll("th, label, option, [aria-label]").forEach((element) => {
    if (element.childElementCount === 0 && element.textContent) {
      element.textContent = element.textContent
        .replaceAll("Semua bagian", "Semua jabatan")
        .replaceAll("Pilih bagian", "Pilih jabatan")
        .replaceAll("Pilih bagian dan periode", "Pilih jabatan dan periode")
        .replaceAll("Pilih bagian, periode, dan aspek", "Pilih jabatan, periode, dan aspek")
        .replaceAll("Bagian / Periode", "Jabatan / Periode")
        .replace(/^Bagian$/, "Jabatan");
    }
    const label = element.getAttribute("aria-label");
    if (label) element.setAttribute("aria-label", label.replaceAll("bagian", "jabatan").replaceAll("Bagian", "Jabatan"));
  });
  document.querySelectorAll('select[name="department_id"]').forEach((select) => {
    const label = select.closest("label");
    if (label?.firstChild?.nodeType === Node.TEXT_NODE) label.firstChild.textContent = "Jabatan";
  });
  document.querySelectorAll("[data-rule-rubric]").forEach((select) => {
    const label = select.closest("label");
    if (label?.firstChild?.nodeType === Node.TEXT_NODE) label.firstChild.textContent = "Jabatan / Periode";
  });
}

if (window.location.pathname === "/placement-change-requests") {
  document.querySelector("table")?.closest("section")?.classList.add("hidden", "md:block");
  document.querySelectorAll('select[name="employee_id"] option').forEach((option) => {
    if (option.textContent.trim() === "Pilih bawahan") option.textContent = "Pilih karyawan";
  });
}

if (window.location.pathname === "/my-team") document.querySelector("[data-filter-count]")?.classList.add("hidden");
