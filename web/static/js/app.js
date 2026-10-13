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
  button.addEventListener("click", () => document.getElementById(button.dataset.openDialog)?.showModal());
});
document.querySelectorAll("[data-close-dialog]").forEach((button) => {
  button.addEventListener("click", () => button.closest("dialog")?.close());
});
document.querySelectorAll("dialog").forEach((dialog) => {
  dialog.addEventListener("click", (event) => {
    if (event.target === dialog) dialog.close();
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
  if (!criterion) return;
  const available = Boolean(aspectId);
  criterion.disabled = !available;
  const placeholder = criterion.options[0];
  if (placeholder) placeholder.textContent = available ? "Pilih subaspek" : "Pilih aspek terlebih dahulu";
  [...criterion.options].slice(1).forEach((option) => {
    const visible = option.dataset.aspectId === String(aspectId);
    option.hidden = !visible;
    option.disabled = !visible;
  });
  criterion.value = selectedCriterion;
  if (!criterion.value) criterion.value = "";
}

document.querySelectorAll("[data-rule-form]").forEach((form) => {
  const aspect = form.querySelector("[data-rule-aspect]");
  aspect?.addEventListener("change", () => filterRuleCriteria(form, aspect.value));
  filterRuleCriteria(form, aspect?.value ?? "");
});

document.querySelectorAll("[data-edit-rule]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editRuleDialog");
    const form = document.getElementById("editRuleForm");
    if (!dialog || !form) return;
    form.action = `/master/assessment-score-rules/${button.dataset.criterionId}/${button.dataset.score}/update`;
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

document.querySelectorAll("[data-edit-role], [data-edit-permission], [data-edit-user], [data-edit-store], [data-edit-aspect], [data-edit-criterion], [data-edit-rule]").forEach((button) => {
  button.title = "Edit";
});

document.querySelectorAll("[data-access-filter]").forEach((filter) => {
  const table = filter.nextElementSibling;
  if (!table?.matches("[data-filter-table]")) return;

  const search = filter.querySelector("[data-filter-search]");
  const guard = filter.querySelector("[data-filter-guard]");
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
    const filteredRows = rows.filter((row) => {
      const matchesSearch = (row.dataset.filterSearch ?? "").toLowerCase().includes(query);
      const matchesGuard = !selectedGuard || row.dataset.filterGuard === selectedGuard;
      return matchesSearch && matchesGuard;
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
