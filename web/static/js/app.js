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
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-store]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editStoreDialog");
    const form = document.getElementById("editStoreForm");
    if (!dialog || !form) return;
    form.action = `/master/stores/${button.dataset.id}/update`;
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
    form.elements.department_id.value = button.dataset.departmentId ?? "";
    form.elements.aspect_id.value = button.dataset.aspectId ?? "";
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
    form.elements.code.value = button.dataset.code ?? "";
    form.elements.name.value = button.dataset.name ?? "";
    if (form.elements.rank_order) form.elements.rank_order.value = button.dataset.rankOrder ?? "";
    if (form.elements.job_level_id) form.elements.job_level_id.value = button.dataset.jobLevelId ?? "";
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
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-profile]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editProfileDialog");
    const form = document.getElementById("editProfileForm");
    if (!dialog || !form) return;
    form.action = `/master/employee-profiles/${button.dataset.id}/update`;
    const values = { birth_place: "birthPlace", birth_date: "birthDate", gender: "gender", address: "address", joined_date: "joinedDate", employment_status: "employmentStatus", photo_path: "photoPath", emergency_contact_name: "emergencyName", emergency_contact_phone: "emergencyPhone", emergency_contact_relationship: "emergencyRelationship" };
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
    const values = { employee_id: "employeeId", position_id: "positionId", department_id: "departmentId", store_id: "storeId", started_at: "startedAt", ended_at: "endedAt" };
    Object.entries(values).forEach(([field, data]) => { form.elements[field].value = button.dataset[data] ?? ""; });
    dialog.showModal();
  });
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
    const rating = average <= 0 ? "Belum lengkap" : average <= 1.8 ? "Sangat Kurang Sekali" : average <= 2.6 ? "Kurang Sekali" : average <= 3.4 ? "Kurang" : average <= 4.2 ? "Mendekati Standar" : "Sesuai Standar";
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
    const matches = rows.filter(row => row.dataset.departmentId === department && month >= Number(row.dataset.monthFrom) && (!Number(row.dataset.monthTo) || month <= Number(row.dataset.monthTo)));
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
		if (message) message.textContent = !trainee?.value ? 'Pilih trainee dan bulan pelatihan untuk menampilkan rubrik.' : valid ? `Rubrik aktif: ${matches[0].dataset.rubricName}` : 'Rubrik bagian dan bulan ini belum tersedia atau belum lengkap. Hubungi pengelola master penilaian.';
    assessmentForm.querySelector('[data-submit-final]').disabled = !valid;
    updateAssessmentTotal();
  };
  trainee?.addEventListener('change', updateRubric);
  assessmentForm.elements.training_month.addEventListener('input', updateRubric);
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

document.querySelectorAll("[data-access-filter]").forEach((filter) => {
  const table = filter.nextElementSibling;
  if (!table?.matches("[data-filter-table]")) return;

  const search = filter.querySelector("[data-filter-search]");
  const guard = filter.querySelector("[data-filter-guard]");
  const extra = filter.querySelector("[data-filter-extra]");
  const department = filter.querySelector("[data-filter-department]");
  const count = filter.querySelector("[data-filter-count]");
  const rows = [...table.querySelectorAll("[data-filter-row]")];
  const empty = table.querySelector("[data-filter-empty]");
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
    const filteredRows = rows.filter((row) => {
      const matchesSearch = (row.dataset.filterSearch ?? "").toLowerCase().includes(query);
      const matchesGuard = !selectedGuard || row.dataset.filterGuard === selectedGuard;
      const matchesExtra = !selectedExtra || row.dataset.filterExtra === selectedExtra;
      const matchesDepartment = !selectedDepartment || row.dataset.filterDepartment === selectedDepartment;
      return matchesSearch && matchesGuard && matchesExtra && matchesDepartment;
    });
    const totalPages = Math.max(1, Math.ceil(filteredRows.length / pageSize));
    currentPage = Math.min(currentPage, totalPages);
    const pageStart = (currentPage - 1) * pageSize;
    const pageRows = new Set(filteredRows.slice(pageStart, pageStart + pageSize));
    rows.forEach((row) => row.classList.toggle("hidden", !pageRows.has(row)));
    filteredRows.forEach((row, index) => {
      const number = row.querySelector("[data-row-number]");
      if (number) number.textContent = String(index + 1);
    });

    if (count) count.textContent = `${filteredRows.length} data`;
    empty?.classList.toggle("hidden", filteredRows.length !== 0 || rows.length === 0);
    renderPagination(filteredRows.length, totalPages);
  }

  search?.addEventListener("input", () => applyAccessFilter(true));
  guard?.addEventListener("change", () => applyAccessFilter(true));
  extra?.addEventListener("change", () => applyAccessFilter(true));
  department?.addEventListener("change", () => applyAccessFilter(true));
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
