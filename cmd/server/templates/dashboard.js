// src/client/admin/state.ts
var token = null;
var user = JSON.parse(localStorage.getItem("simp_user") || "null");
var selectedActivities = new Set;
var filterState = {
  username: [],
  empresa: [],
  rua: [],
  predio: [],
  impresso: [],
  id: []
};
var dateFilterState = { dataFimStart: "", dataFimEnd: "" };
var sortState = { column: null, direction: null };
var filterOptions = {
  username: [],
  empresa: [],
  rua: [],
  predio: [],
  impresso: [],
  id: []
};
var openDropdownColumn = null;
var currentModalItems = [];
var modalSort = {
  col: null,
  dir: null
};
function setToken(val) {
  token = val;
}
function setUser(val) {
  user = val;
}
function setFilterOptions(val) {
  filterOptions = val;
}
function setOpenDropdownColumn(val) {
  openDropdownColumn = val;
}
function setCurrentModalItems(val) {
  currentModalItems = val;
}
function setModalSort(val) {
  modalSort = val;
}
var COLUMN_LABELS = {
  username: "Usuário",
  empresa: "Empresa",
  rua: "Rua",
  predio: "Prédio",
  impresso: "Impresso",
  id: "ID",
  dataFim: "Data Fim"
};

// src/client/admin/helpers.ts
var _logoutCb = null;
function setLogoutCb(cb) {
  _logoutCb = cb;
}
async function api(path, options = {}) {
  return apiCall(path, options, () => _logoutCb?.());
}
function rolePt(role) {
  const map = {
    sysadmin: "Sysadmin",
    gerente: "Gerente",
    conferente: "Conferente"
  };
  return map[role] || role;
}
function fmtDate(iso) {
  if (!iso)
    return '<span style="color:var(--muted)">—</span>';
  try {
    return new Date(iso).toLocaleDateString("pt-BR");
  } catch {
    return iso;
  }
}
function fmtDateShort(dateStr) {
  if (!dateStr)
    return "";
  const parts = dateStr.split("-");
  if (parts.length === 3)
    return `${parts[2]}/${parts[1]}/${parts[0]}`;
  return dateStr;
}
function setMsg(el, text2, ok) {
  el.textContent = text2;
  el.className = `msg ${ok ? "msg-ok" : "msg-err"}`;
  if (ok)
    setTimeout(() => {
      el.textContent = "";
      el.className = "msg";
    }, 3000);
}
async function loadFilterOptions() {
  const result = await api("/api/dashboard/activities/filters");
  if (result.ok && result.data) {
    setFilterOptions(result.data);
  }
}
function buildQueryString() {
  const params = [];
  params.push("limit=100");
  const keys = ["username", "empresa", "rua", "predio", "impresso", "id"];
  for (const key of keys) {
    for (const val of (filterState[key] || [])) {
      params.push(`filter_${key}[]=${encodeURIComponent(val)}`);
    }
  }
  if (dateFilterState.dataFimStart) {
    params.push(`filter_dataFimStart=${encodeURIComponent(dateFilterState.dataFimStart)}`);
  }
  if (dateFilterState.dataFimEnd) {
    params.push(`filter_dataFimEnd=${encodeURIComponent(dateFilterState.dataFimEnd)}`);
  }
  if (sortState.column && sortState.direction) {
    params.push(`sort=${sortState.column}`);
    params.push(`order=${sortState.direction}`);
  }
  return params.join("&");
}
function updateFilterActiveClasses() {
  const ths = document.querySelectorAll(".th-filterable[data-column]");
  ths.forEach((th) => {
    const col = th.getAttribute("data-column");
    if (!col)
      return;
    if (col === "dataFim") {
      if (dateFilterState.dataFimStart || dateFilterState.dataFimEnd) {
        th.classList.add("filter-active");
      } else {
        th.classList.remove("filter-active");
      }
    } else if (filterState[col] && filterState[col].length > 0) {
      th.classList.add("filter-active");
    } else {
      th.classList.remove("filter-active");
    }
  });
}
function updateActiveFiltersBar() {
  const bar = document.getElementById("active-filters-bar");
  if (!bar)
    return;
  const keys = ["username", "empresa", "rua", "predio", "impresso", "id"];
  let hasAny = false;
  let html2 = '<span class="active-filter-label">Filtros:</span>';
  for (const key of keys) {
    if (filterState[key] && filterState[key].length > 0) {
      hasAny = true;
      for (const val of filterState[key]) {
        html2 += '<span class="active-filter-tag">' + (COLUMN_LABELS[key] || key) + ": " + escHtml(val) + ' <span class="tag-remove" data-filter-action="remove-value" data-filter-column="' + key + '" data-filter-value="' + escHtml(val) + '">&times;</span>' + "</span>";
      }
    }
  }
  if (dateFilterState.dataFimStart || dateFilterState.dataFimEnd) {
    hasAny = true;
    let dateLabel = "Data Fim: ";
    if (dateFilterState.dataFimStart && dateFilterState.dataFimEnd) {
      dateLabel += fmtDateShort(dateFilterState.dataFimStart) + " → " + fmtDateShort(dateFilterState.dataFimEnd);
    } else if (dateFilterState.dataFimStart) {
      dateLabel += `a partir de ${fmtDateShort(dateFilterState.dataFimStart)}`;
    } else {
      dateLabel += `até ${fmtDateShort(dateFilterState.dataFimEnd)}`;
    }
    html2 += '<span class="active-filter-tag">' + dateLabel + ' <span class="tag-remove" data-filter-action="clear-date">&times;</span>' + "</span>";
  }
  if (hasAny) {
    html2 += '<button class="btn-clear-filters" data-filter-action="clear-all">Limpar Filtros</button>';
    bar.innerHTML = sanitizeHtml(html2);
    bar.classList.remove("hidden");
  } else {
    bar.classList.add("hidden");
    bar.innerHTML = "";
  }
  updateFilterActiveClasses();
}
function closeFilterDropdown() {
  const existing = document.getElementById("filter-dropdown-active");
  if (existing) {
    const parent = existing.parentElement;
    existing.remove();
    if (parent) {
      const col = parent.getAttribute("data-column");
      if (col && (!filterState[col] || filterState[col].length === 0)) {
        parent.classList.remove("filter-active");
      }
    }
  }
  setOpenDropdownColumn(null);
}

// src/client/admin/activities.ts
async function loadActivities() {
  const tbody = document.getElementById("activities-tbody");
  if (!tbody)
    return;
  tbody.innerHTML = "";
  const loadingTr = document.createElement("tr");
  const loadingTd = document.createElement("td");
  loadingTd.colSpan = 9;
  const loadingDiv = document.createElement("div");
  loadingDiv.className = "empty";
  loadingDiv.textContent = "Carregando…";
  loadingTd.appendChild(loadingDiv);
  loadingTr.appendChild(loadingTd);
  tbody.appendChild(loadingTr);
  selectedActivities.clear();
  updateBulkPrintButton();
  const selectAllCheckbox = document.getElementById("select-all-activities");
  if (selectAllCheckbox)
    selectAllCheckbox.checked = false;
  if (filterOptions.username.length === 0 && filterOptions.id.length === 0) {
    await loadFilterOptions();
  }
  updateActiveFiltersBar();
  const qs = buildQueryString();
  const result = await api(`/api/dashboard/activities?${qs}`);
  if (!result.ok || !Array.isArray(result.data)) {
    tbody.innerHTML = "";
    const errTr = document.createElement("tr");
    const errTd = document.createElement("td");
    errTd.colSpan = 9;
    const errDiv = document.createElement("div");
    errDiv.className = "empty";
    errDiv.style.color = "var(--danger)";
    errDiv.textContent = "Erro ao carregar atividades.";
    errTd.appendChild(errDiv);
    errTr.appendChild(errTd);
    tbody.appendChild(errTr);
    return;
  }
  const data = result.data;
  const statTotalAct = document.getElementById("stat-total-act");
  if (statTotalAct)
    statTotalAct.textContent = data.length.toString();
  const statFinalizadas = document.getElementById("stat-finalizadas");
  if (statFinalizadas)
    statFinalizadas.textContent = data.filter((a) => a.dataFim).length.toString();
  const empresaSet = {};
  for (const act of data) {
    empresaSet[act.empresa] = true;
  }
  const statEmpresasAct = document.getElementById("stat-empresas-act");
  if (statEmpresasAct)
    statEmpresasAct.textContent = Object.keys(empresaSet).length.toString();
  if (data.length === 0) {
    tbody.innerHTML = "";
    const emptyTr = document.createElement("tr");
    const emptyTd = document.createElement("td");
    emptyTd.colSpan = 9;
    const emptyDiv = document.createElement("div");
    emptyDiv.className = "empty";
    emptyDiv.textContent = "Nenhuma atividade encontrada.";
    emptyTd.appendChild(emptyDiv);
    emptyTr.appendChild(emptyTd);
    tbody.appendChild(emptyTr);
    return;
  }
  tbody.innerHTML = "";
  for (const a of data) {
    const tr = buildActivityRow(a);
    tbody.appendChild(tr);
  }
}
function buildActivityRow(a) {
  const tr = document.createElement("tr");
  tr.setAttribute("data-activity-id", String(a.id));
  const cbCell = document.createElement("td");
  cbCell.className = "checkbox-col";
  const cb = document.createElement("input");
  cb.type = "checkbox";
  cb.className = "activity-checkbox";
  cb.setAttribute("data-activity-id", String(a.id));
  cbCell.appendChild(cb);
  tr.appendChild(cbCell);
  const idCell = document.createElement("td");
  idCell.style.color = "var(--muted)";
  idCell.style.fontSize = "12px";
  idCell.textContent = `#${a.id}`;
  tr.appendChild(idCell);
  const userCell = document.createElement("td");
  const userStrong = document.createElement("strong");
  userStrong.textContent = a.username || "—";
  userCell.appendChild(userStrong);
  tr.appendChild(userCell);
  const empCell = document.createElement("td");
  empCell.textContent = a.empresa || "—";
  tr.appendChild(empCell);
  const ruaCell = document.createElement("td");
  ruaCell.textContent = a.rua || "—";
  tr.appendChild(ruaCell);
  const predioCell = document.createElement("td");
  predioCell.textContent = a.predio || "—";
  tr.appendChild(predioCell);
  const dateCell = document.createElement("td");
  dateCell.style.color = "var(--muted)";
  dateCell.style.fontSize = "12px";
  dateCell.textContent = fmtDate(a.dataFim);
  tr.appendChild(dateCell);
  const imprCell = document.createElement("td");
  imprCell.style.textAlign = "center";
  imprCell.style.fontWeight = "bold";
  imprCell.style.color = a.impresso ? "#10b981" : "var(--muted)";
  imprCell.textContent = a.impresso ? "S" : "N";
  tr.appendChild(imprCell);
  const actionCell = document.createElement("td");
  const printBtn = document.createElement("button");
  printBtn.className = "btn btn-ghost btn-print";
  printBtn.setAttribute("data-activity-id", String(a.id));
  printBtn.setAttribute("data-action", "print");
  printBtn.title = "Imprimir";
  printBtn.textContent = "\uD83D\uDDA8️";
  actionCell.appendChild(printBtn);
  tr.appendChild(actionCell);
  return tr;
}
function toggleActivitySelection(id, checked) {
  if (checked) {
    selectedActivities.add(id);
  } else {
    selectedActivities.delete(id);
  }
  updateBulkPrintButton();
}
function toggleAllActivities(checked) {
  const checkboxes = document.querySelectorAll(".activity-checkbox");
  checkboxes.forEach((cb) => {
    cb.checked = checked;
    const id = parseInt(cb.getAttribute("data-activity-id") ?? "0", 10);
    if (checked)
      selectedActivities.add(id);
    else
      selectedActivities.delete(id);
  });
  updateBulkPrintButton();
}
function updateBulkPrintButton() {
  const btn = document.getElementById("btn-bulk-print");
  if (!btn)
    return;
  if (selectedActivities.size > 0) {
    btn.classList.remove("hidden");
    btn.textContent = `\uD83D\uDDA8️ Imprimir (${selectedActivities.size})`;
  } else {
    btn.classList.add("hidden");
  }
}

// src/client/admin/auth.ts
var _loadUsersCb = null;
function setLoadUsersCb(cb) {
  _loadUsersCb = cb;
}
var _allowedRoles = null;
function setAllowedRoles(roles) {
  _allowedRoles = roles;
}
function updateAuthState() {
  const loginScreen = document.getElementById("admin-login-screen");
  const appScreen = document.getElementById("app");
  const allowed = _allowedRoles || ["sysadmin"];
  if (!token || !user || !allowed.includes(user.role)) {
    loginScreen?.classList.remove("hidden");
    appScreen?.classList.add("hidden");
  } else {
    loginScreen?.classList.add("hidden");
    appScreen?.classList.remove("hidden");
    initDashboard();
  }
}
function initDashboard() {
  if (!user)
    return;
  const sidebarUsername = document.getElementById("sidebar-username");
  if (sidebarUsername)
    sidebarUsername.textContent = user.username;
  const sidebarRole = document.getElementById("sidebar-role");
  if (sidebarRole)
    sidebarRole.textContent = rolePt(user.role);
  const userAvatar = document.getElementById("user-avatar");
  if (userAvatar)
    userAvatar.textContent = user.username.charAt(0).toUpperCase();
  const userStatsRow = document.getElementById("user-stats-row");
  if (user.role !== "sysadmin" && userStatsRow) {
    userStatsRow.style.display = "none";
  }
  _loadUsersCb?.();
}
function logout() {
  apiCall("/api/auth/logout", { method: "POST" }).catch(() => {});
  setToken(null);
  setUser(null);
  localStorage.removeItem("simp_user");
  updateAuthState();
}

// src/client/admin/filters.ts
function handleColumnClick(column, event) {
  event.stopPropagation();
  if (openDropdownColumn === column) {
    closeFilterDropdown();
    return;
  }
  closeFilterDropdown();
  openFilterDropdown(column, event.currentTarget);
}
function handleSort(column, event) {
  event.stopPropagation();
  closeFilterDropdown();
  cycleSort(column);
}
function cycleSort(column) {
  if (sortState.column === column) {
    if (sortState.direction === "asc") {
      sortState.direction = "desc";
    } else if (sortState.direction === "desc") {
      sortState.column = null;
      sortState.direction = null;
    }
  } else {
    sortState.column = column;
    sortState.direction = "asc";
  }
  updateSortIndicators();
  loadActivities();
}
function updateSortIndicators() {
  const allCols = [
    "id",
    "username",
    "empresa",
    "rua",
    "predio",
    "dataFim",
    "impresso"
  ];
  for (const col of allCols) {
    const el = document.getElementById(`sort-${col}`);
    if (!el)
      continue;
    if (sortState.column === col) {
      el.classList.add("active");
      el.textContent = sortState.direction === "asc" ? "↑" : "↓";
    } else {
      el.classList.remove("active");
      el.textContent = "⇅";
    }
  }
}
function openFilterDropdown(column, thElement) {
  setOpenDropdownColumn(column);
  if (column === "dataFim") {
    openDateFilterDropdown(thElement);
    return;
  }
  const options = filterOptions[column] || [];
  const selected = filterState[column] || [];
  const dropdown = document.createElement("div");
  dropdown.className = "filter-dropdown";
  dropdown.id = "filter-dropdown-active";
  dropdown.setAttribute("data-filter-column", column);
  const searchHtml = '<div class="filter-search-wrap">' + '<input type="text" class="filter-search" placeholder="Pesquisar..." id="filter-search-input">' + "</div>";
  const actionsHtml = '<div class="filter-actions">' + '<button data-filter-action="select-all">Todos</button>' + '<button data-filter-action="select-none">Nenhum</button>' + '<button data-filter-action="sort">Ordenar</button>' + "</div>";
  let listHtml = '<div class="filter-list" id="filter-list-items">';
  for (const opt of options) {
    const isChecked = selected.indexOf(opt) === -1 ? "" : " checked";
    listHtml += '<label class="filter-item" data-value="' + escHtml(opt.toLowerCase()) + '">' + '<input type="checkbox" value="' + escHtml(opt) + '"' + isChecked + ' data-filter-column="' + column + '" data-filter-value="' + escHtml(opt) + '">' + '<span class="filter-item-label">' + escHtml(opt) + "</span>" + "</label>";
  }
  listHtml += "</div>";
  dropdown.innerHTML = sanitizeHtml(searchHtml + actionsHtml + listHtml);
  thElement.appendChild(dropdown);
  thElement.classList.add("filter-active");
  const searchInput = document.getElementById("filter-search-input");
  if (searchInput) {
    searchInput.addEventListener("input", function() {
      const term = this.value.toLowerCase();
      const items = document.querySelectorAll("#filter-list-items .filter-item");
      for (let j = 0;j < items.length; j++) {
        const val = items[j].getAttribute("data-value") || "";
        if (val.indexOf(term) !== -1) {
          items[j].classList.remove("hidden-by-search");
        } else {
          items[j].classList.add("hidden-by-search");
        }
      }
    });
    setTimeout(() => searchInput.focus(), 50);
  }
}
function openDateFilterDropdown(thElement) {
  const dropdown = document.createElement("div");
  dropdown.className = "filter-dropdown";
  dropdown.id = "filter-dropdown-active";
  dropdown.setAttribute("data-filter-column", "dataFim");
  const startVal = dateFilterState.dataFimStart || "";
  const endVal = dateFilterState.dataFimEnd || "";
  const html2 = '<div class="filter-date-wrap">' + '<div class="filter-date-field">' + "<label>De</label>" + '<input type="date" id="date-filter-start" lang="pt-BR" value="' + startVal + '">' + "</div>" + '<div class="filter-date-field">' + "<label>Até</label>" + '<input type="date" id="date-filter-end" lang="pt-BR" value="' + endVal + '">' + "</div>" + '<div class="filter-date-actions">' + '<button class="btn-apply" data-filter-action="apply-date">Aplicar</button>' + '<button class="btn-clear" data-filter-action="clear-date">Limpar</button>' + "</div>" + '<div class="filter-actions" style="border-top: 1px solid var(--border); border-bottom: none;">' + '<button data-filter-action="sort">Ordenar</button>' + "</div>" + "</div>";
  dropdown.innerHTML = sanitizeHtml(html2);
  thElement.appendChild(dropdown);
  thElement.classList.add("filter-active");
}
function applyDateFilter() {
  const startEl = document.getElementById("date-filter-start");
  const endEl = document.getElementById("date-filter-end");
  dateFilterState.dataFimStart = startEl ? startEl.value : "";
  dateFilterState.dataFimEnd = endEl ? endEl.value : "";
  closeFilterDropdown();
  updateActiveFiltersBar();
  loadActivities();
}
function clearDateFilter() {
  dateFilterState.dataFimStart = "";
  dateFilterState.dataFimEnd = "";
  const startEl = document.getElementById("date-filter-start");
  const endEl = document.getElementById("date-filter-end");
  if (startEl)
    startEl.value = "";
  if (endEl)
    endEl.value = "";
  updateActiveFiltersBar();
  loadActivities();
}
function handleFilterCheck(column, value, checked) {
  if (!filterState[column])
    filterState[column] = [];
  if (checked) {
    if (filterState[column].indexOf(value) === -1) {
      filterState[column].push(value);
    }
  } else {
    const idx = filterState[column].indexOf(value);
    if (idx !== -1)
      filterState[column].splice(idx, 1);
  }
  updateActiveFiltersBar();
  loadActivities();
}
function filterSelectAll(column) {
  filterState[column] = [];
  const checkboxes = document.querySelectorAll('#filter-list-items input[type="checkbox"]');
  checkboxes.forEach((cb) => {
    cb.checked = false;
  });
  updateActiveFiltersBar();
  loadActivities();
}
function filterSelectNone(column) {
  const options = filterOptions[column] || [];
  filterState[column] = options.slice();
  const checkboxes = document.querySelectorAll('#filter-list-items input[type="checkbox"]');
  checkboxes.forEach((cb) => {
    cb.checked = true;
  });
  updateActiveFiltersBar();
  loadActivities();
}
function cycleSortFromDropdown(column) {
  closeFilterDropdown();
  cycleSort(column);
}
function clearAllFilters() {
  const keys = ["username", "empresa", "rua", "predio", "impresso", "id"];
  for (const key of keys) {
    filterState[key] = [];
  }
  dateFilterState.dataFimStart = "";
  dateFilterState.dataFimEnd = "";
  updateActiveFiltersBar();
  updateFilterActiveClasses();
  loadActivities();
}
function removeFilterValue(column, value) {
  if (!filterState[column])
    return;
  const idx = filterState[column].indexOf(value);
  if (idx !== -1)
    filterState[column].splice(idx, 1);
  updateActiveFiltersBar();
  updateFilterActiveClasses();
  loadActivities();
}

// src/client/admin/modal.ts
async function showActivityDetails(id) {
  showLoader(true);
  const { ok, data } = await api(`/api/dashboard/activities/${id}`);
  showLoader(false);
  if (!ok) {
    alert(data.error || "Erro ao carregar detalhes da atividade.");
    return;
  }
  setCurrentModalItems(data.items || []);
  setModalSort({ col: null, dir: null });
  const modalTitle = document.getElementById("modal-title");
  if (modalTitle)
    modalTitle.textContent = `Detalhes da Atividade`;
  const modalSubtitle = document.getElementById("modal-subtitle");
  if (modalSubtitle)
    modalSubtitle.textContent = `#${data.id} — ${fmtDate(data.dataFim)}`;
  const modalUser = document.getElementById("modal-user");
  if (modalUser)
    modalUser.textContent = data.username || "—";
  const modalLocal = document.getElementById("modal-local");
  if (modalLocal)
    modalLocal.textContent = `Empresa ${data.empresa} / Rua ${data.rua} / Prédio ${data.predio}`;
  renderModalItems();
  updateModalSortIndicators();
  document.getElementById("modal-history")?.classList.remove("hidden");
}
function renderModalItems() {
  const tbody = document.getElementById("modal-history-tbody");
  if (!tbody)
    return;
  tbody.innerHTML = "";
  if (currentModalItems.length === 0) {
    const tr = document.createElement("tr");
    const td = document.createElement("td");
    td.colSpan = 7;
    const div = document.createElement("div");
    div.className = "empty";
    div.textContent = "Nenhum item verificado.";
    td.appendChild(div);
    tr.appendChild(td);
    tbody.appendChild(tr);
    return;
  }
  for (const item of currentModalItems) {
    const tr = document.createElement("tr");
    const nameTd = document.createElement("td");
    const nameDiv = document.createElement("div");
    nameDiv.style.fontWeight = "600";
    nameDiv.textContent = item.desccompleta || "Sem descrição";
    nameTd.appendChild(nameDiv);
    const seqDiv = document.createElement("div");
    seqDiv.style.fontSize = "11px";
    seqDiv.style.color = "var(--muted)";
    seqDiv.textContent = `Seq: ${item.seqproduto}`;
    nameTd.appendChild(seqDiv);
    tr.appendChild(nameTd);
    const statusTd = document.createElement("td");
    statusTd.style.textAlign = "center";
    const statusWrap = document.createElement("div");
    statusWrap.style.display = "flex";
    statusWrap.style.flexDirection = "column";
    statusWrap.style.alignItems = "center";
    statusWrap.style.gap = "2px";
    const badge = document.createElement("span");
    const statusClass = item.status === "OK" ? "badge-ok" : item.status === "ERRO" || item.status === "DIVERGENTE" ? "badge-warning" : "badge-error";
    badge.className = `badge ${statusClass}`;
    badge.textContent = item.status || "";
    statusWrap.appendChild(badge);
    if (item.reposicao) {
      const repBadge = document.createElement("span");
      repBadge.className = "badge";
      repBadge.style.background = "var(--warning)";
      repBadge.style.color = "#000";
      repBadge.style.fontSize = "10px";
      repBadge.textContent = "\uD83D\uDCE6 REPOSIÇÃO";
      statusWrap.appendChild(repBadge);
    }
    statusTd.appendChild(statusWrap);
    tr.appendChild(statusTd);
    const expectedAddr = item.expectedRua != null && item.expectedPredio != null ? `${item.expectedRua}/${item.expectedPredio}` : "N/A";
    const expTd = document.createElement("td");
    expTd.style.textAlign = "center";
    expTd.style.fontSize = "12px";
    expTd.style.color = "var(--muted)";
    expTd.textContent = expectedAddr;
    tr.appendChild(expTd);
    const readAddr = item.rua != null && item.predio != null ? `${item.rua}/${item.predio}` : "N/A";
    const readTd = document.createElement("td");
    readTd.style.textAlign = "center";
    readTd.style.fontSize = "12px";
    readTd.style.fontWeight = "600";
    readTd.textContent = readAddr;
    tr.appendChild(readTd);
    const estTd = document.createElement("td");
    estTd.style.textAlign = "right";
    estTd.style.fontWeight = "600";
    estTd.textContent = item.estoque != null ? String(item.estoque) : "";
    tr.appendChild(estTd);
    const mdv = item.mdv !== null ? item.mdv : "—";
    const mdvTd = document.createElement("td");
    mdvTd.style.textAlign = "right";
    mdvTd.style.fontSize = "12px";
    mdvTd.textContent = mdv;
    tr.appendChild(mdvTd);
    const ddv = item.ddv !== null ? item.ddv : "—";
    const ddvTd = document.createElement("td");
    ddvTd.style.textAlign = "right";
    ddvTd.style.fontSize = "12px";
    ddvTd.style.fontWeight = "600";
    ddvTd.textContent = ddv;
    tr.appendChild(ddvTd);
    tbody.appendChild(tr);
  }
}
function handleModalSort(col) {
  if (modalSort.col === col) {
    modalSort.dir = modalSort.dir === "asc" ? "desc" : "asc";
  } else {
    modalSort.col = col;
    modalSort.dir = "asc";
  }
  currentModalItems.sort((a, b) => {
    let valA, valB;
    if (col === "produto") {
      valA = a.desccompleta || "";
      valB = b.desccompleta || "";
    } else if (col === "status") {
      valA = a.status || "";
      valB = b.status || "";
    } else if (col === "esperado") {
      valA = `${a.expectedRua || ""}/${a.expectedPredio || ""}`;
      valB = `${b.expectedRua || ""}/${b.expectedPredio || ""}`;
    } else if (col === "lido") {
      valA = `${a.rua || ""}/${a.predio || ""}`;
      valB = `${b.rua || ""}/${b.predio || ""}`;
    } else if (col === "estoque" || col === "mdv" || col === "ddv") {
      const itemA = a;
      const itemB = b;
      valA = itemA[col] ?? 0;
      valB = itemB[col] ?? 0;
    } else {
      valA = 0;
      valB = 0;
    }
    if (typeof valA === "string" && typeof valB === "string") {
      return modalSort.dir === "asc" ? valA.localeCompare(valB, undefined, { numeric: true }) : valB.localeCompare(valA, undefined, { numeric: true });
    }
    return modalSort.dir === "asc" ? valA - valB : valB - valA;
  });
  renderModalItems();
  updateModalSortIndicators();
}
function updateModalSortIndicators() {
  const allCols = [
    "produto",
    "status",
    "esperado",
    "lido",
    "estoque",
    "mdv",
    "ddv"
  ];
  for (const col of allCols) {
    const el = document.getElementById(`modal-sort-${col}`);
    if (!el)
      continue;
    if (modalSort.col === col) {
      el.classList.add("active");
      el.textContent = modalSort.dir === "asc" ? "↑" : "↓";
    } else {
      el.classList.remove("active");
      el.textContent = "⇅";
    }
  }
}
function closeModal() {
  document.getElementById("modal-history")?.classList.add("hidden");
}

// src/client/admin/printing.ts
async function handleBulkPrint() {
  if (selectedActivities.size === 0)
    return;
  showLoader(true);
  const ids = Array.from(selectedActivities);
  const { ok, data } = await api("/api/dashboard/activities/bulk", {
    method: "POST",
    body: JSON.stringify({ ids })
  });
  showLoader(false);
  if (!ok) {
    alert(data.error || "Erro ao carregar dados para impressão em lote.");
    return;
  }
  printMultipleActivities(data);
}
function printMultipleActivities(activities) {
  const printWindow = window.open("", "_blank");
  if (!printWindow) {
    alert("Bloqueador de pop-up impediu a impressão.");
    return;
  }
  let allReportsHtml = "";
  activities.forEach((activity, index) => {
    allReportsHtml += `<div class="activity-container">${generateActivityReportHtml(activity)}</div>`;
    if (index < activities.length - 1) {
      allReportsHtml += '<hr class="report-divider">';
    }
  });
  const finalHtml = `
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="UTF-8">
        <title>SIMP - Relatórios de Impressão</title>
        ${getReportStyles()}
    </head>
    <body>
        ${allReportsHtml}
    </body>
    </html>
    `;
  printWindow.document.write(finalHtml);
  printWindow.document.close();
  printWindow.focus();
  setTimeout(() => {
    printWindow.print();
    printWindow.close();
    const ids = activities.map((a) => a.id);
    api(`/api/dashboard/activities/bulk/print`, {
      method: "PATCH",
      body: JSON.stringify({ ids })
    }).then(() => loadActivities());
  }, 250);
}
function generateActivityReportHtml(data) {
  const itemsToPrint = (data.items || []).filter((item) => item.status === "RUPTURA" || item.reposicao).sort((a, b) => {
    const nameA = (a.desccompleta || "").toUpperCase();
    const nameB = (b.desccompleta || "").toUpperCase();
    if (nameA < nameB)
      return -1;
    if (nameA > nameB)
      return 1;
    return 0;
  });
  const rows = itemsToPrint.map((item) => {
    const expectedAddr = item.expectedRua != null && item.expectedPredio != null ? `${item.expectedRua}/${item.expectedPredio}` : "N/A";
    const formatNum = (v) => v !== null && v !== undefined ? v.toLocaleString("pt-BR", {
      minimumFractionDigits: 0,
      maximumFractionDigits: 2
    }) : "—";
    const mdv = formatNum(item.mdv);
    const ddv = formatNum(item.ddv);
    const dtOpts = {
      day: "2-digit",
      month: "2-digit",
      year: "2-digit"
    };
    const dtEntrada = item.dtaultentrada ? new Date(item.dtaultentrada).toLocaleDateString("pt-BR", dtOpts) : "—";
    let daysSinceLastSale = "—";
    if (item.dtaultvenda) {
      const lastSale = new Date(item.dtaultvenda);
      const today = new Date;
      const diffTime = Math.abs(today.getTime() - lastSale.getTime());
      daysSinceLastSale = Math.floor(diffTime / (1000 * 60 * 60 * 24));
    }
    return `
        <tr>
            <td class="col-seq">${item.seqproduto}</td>
            <td>
                ${item.desccompleta || "Sem descrição"}
                ${item.reposicao ? '<span style="margin-left:4px;" title="Reposição">\uD83D\uDCE6</span>' : ""}
            </td>
            <td class="col-addr">${expectedAddr}</td>
            <td style="text-align:right;">${item.estoque}</td>
            <td style="text-align:right;">${mdv}</td>
            <td style="text-align:right;">${ddv}</td>
            <td class="col-date">${dtEntrada}</td>
            <td class="col-date">${daysSinceLastSale}</td>
        </tr>
        `;
  }).join("");
  return `
        <div class="report-header">
            <div class="header-left">
                <h1>Relatório de Atividade #${data.id}</h1>
                <div class="subtitle">Gerado em ${new Date().toLocaleString("pt-BR")}</div>
            </div>

            <div class="header-right info">
                <div class="info-row">
                    <div>
                        <span><span class="info-label">Usuário:</span> ${data.username || "—"}</span>
                        <span><span class="info-label">Empresa:</span> ${data.empresa}</span>
                    </div>
                    <div>
                        <span><span class="info-label">Rua:</span> ${data.rua}</span>
                        <span><span class="info-label">Prédio:</span> ${data.predio}</span>
                        <span><span class="info-label">Data:</span> ${fmtDate(data.dataFim)}</span>
                    </div>
                </div>
            </div>
        </div>

        <table>
            <thead>
                <tr>
                    <th class="col-seq">Cod.</th>
                    <th>Descrição</th>
                    <th class="col-addr">End. Esp.</th>
                    <th style="text-align:right; width: 40px;">Estq</th>
                    <th style="text-align:right; width: 45px;">MDV</th>
                    <th style="text-align:right; width: 45px;">DDV</th>
                    <th class="col-date">Últ. Ent.</th>
                    <th class="col-date">Dias s/ Vnd.</th>
                </tr>
            </thead>
            <tbody>
                ${rows.length > 0 ? rows : '<tr><td colspan="8" style="text-align:center; padding: 20px;">Nenhuma ruptura ou reposição encontrada nesta atividade.</td></tr>'}
            </tbody>
        </table>

        <div class="footer">
            SIMP - Sistema de Monitoramento de Prateleiras
        </div>
    `;
}
function getReportStyles() {
  return `
        <style>
            * { box-sizing: border-box; margin: 0; padding: 0; }
            body { font-family: Arial, sans-serif; font-size: 11px; color: #000; padding: 20px; }
            h1 { font-size: 18px; margin-bottom: 4px; }
            .subtitle { color: #666; font-size: 12px; }
            .report-header {
                display: flex;
                justify-content: space-between;
                align-items: flex-start;
                margin-bottom: 15px;
            }
            .header-right { text-align: right; }
            .info { margin-bottom: 0; }
            .info-row { display: flex; flex-direction: column; align-items: flex-end; gap: 2px; }
            .info-row > div { display: flex; gap: 15px; }
            .info-label { font-weight: bold; color: #444; }
            table { width: 100%; border-collapse: collapse; margin-top: 10px; table-layout: auto; }
            th, td { border: 1px solid #ccc; padding: 4px 6px; text-align: left; word-wrap: break-word; }
            th { background: #f5f5f5; font-weight: bold; }
            tr { page-break-inside: avoid; }
            .footer { margin-top: 20px; font-size: 10px; color: #888; text-align: center; }
            .col-seq { width: 40px; }
            .col-addr { width: 60px; text-align: center; }
            .col-date { width: 65px; text-align: center; }
            .activity-container { margin-bottom: 20px; }
            .report-divider {
                border: none;
                border-top: 2px dashed #000;
                margin: 40px 0;
                page-break-inside: avoid;
            }
            @media print {
                body { padding: 0; }
                .report-divider { border-top-color: #aaa; }
            }
        </style>
    `;
}
async function printActivity(id) {
  showLoader(true);
  const { ok, data } = await api(`/api/dashboard/activities/${id}`);
  showLoader(false);
  if (!ok) {
    alert(data.error || "Erro ao carregar dados para impressão.");
    return;
  }
  printMultipleActivities([data]);
}

// src/client/dashboard/index.ts
setLogoutCb(logout);
window.filterSelectAll = filterSelectAll;
window.filterSelectNone = filterSelectNone;
window.cycleSortFromDropdown = cycleSortFromDropdown;
window.removeFilterValue = removeFilterValue;
window.clearAllFilters = clearAllFilters;
window.applyDateFilter = applyDateFilter;
window.clearDateFilter = clearDateFilter;
window.handleFilterCheck = handleFilterCheck;
window.cycleSort = cycleSort;
window.handleSort = handleSort;
window.handleColumnClick = handleColumnClick;
window.printActivity = printActivity;
window.showActivityDetails = showActivityDetails;
window.toggleActivitySelection = toggleActivitySelection;
window.toggleAllActivities = toggleAllActivities;
window.handleBulkPrint = handleBulkPrint;
window.handleModalSort = handleModalSort;
window.closeModal = closeModal;
window.loadActivities = loadActivities;
document.getElementById("form-admin-login")?.addEventListener("submit", async (e) => {
  e.preventDefault();
  showLoader(true);
  const errEl = document.getElementById("admin-login-error");
  errEl?.classList.add("hidden");
  const username = document.getElementById("admin-username").value;
  const password = document.getElementById("admin-password").value;
  const res = await apiCall("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password })
  });
  showLoader(false);
  if (res.ok) {
    if (!["sysadmin", "gerente"].includes(res.data.user.role)) {
      if (errEl) {
        errEl.textContent = "Acesso restrito";
        errEl.classList.remove("hidden");
      }
      return;
    }
    setToken("cookie");
    setUser(res.data.user);
    localStorage.setItem("simp_user", JSON.stringify(user));
    updateAuthState();
  } else if (errEl) {
    errEl.textContent = res.data.error || "Erro ao logar";
    errEl.classList.remove("hidden");
  }
});
document.getElementById("btn-logout")?.addEventListener("click", logout);
document.addEventListener("click", (e) => {
  if (openDropdownColumn && !e.target.closest(".filter-dropdown") && !e.target.closest(".th-filterable")) {
    closeFilterDropdown();
  }
});
document.addEventListener("click", (e) => {
  const filterBtn = e.target.closest("[data-filter-action]");
  if (!filterBtn)
    return;
  const dropdown = filterBtn.closest(".filter-dropdown");
  const column = dropdown?.getAttribute("data-filter-column") || "";
  const action = filterBtn.getAttribute("data-filter-action");
  if (action === "select-all")
    filterSelectAll(column);
  else if (action === "select-none")
    filterSelectNone(column);
  else if (action === "sort")
    cycleSortFromDropdown(column);
  else if (action === "apply-date")
    applyDateFilter();
  else if (action === "clear-date")
    clearDateFilter();
  else if (action === "clear-all")
    clearAllFilters();
  else if (action === "remove-value") {
    const val = filterBtn.getAttribute("data-filter-value") || "";
    const col = filterBtn.getAttribute("data-filter-column") || "";
    removeFilterValue(col, val);
  }
});
document.addEventListener("change", (e) => {
  const cb = e.target.closest("input[data-filter-column][data-filter-value]");
  if (!cb)
    return;
  const column = cb.getAttribute("data-filter-column") || "";
  const value = cb.getAttribute("data-filter-value") || "";
  handleFilterCheck(column, value, cb.checked);
});
document.getElementById("activities-tbody")?.addEventListener("click", (e) => {
  const printBtn = e.target.closest('button[data-action="print"]');
  if (printBtn) {
    const id = Number(printBtn.getAttribute("data-activity-id"));
    if (id)
      printActivity(id);
    return;
  }
  const row = e.target.closest("tr[data-activity-id]");
  if (row) {
    const id = Number(row.getAttribute("data-activity-id"));
    if (id)
      showActivityDetails(id);
  }
});
document.getElementById("activities-tbody")?.addEventListener("change", (e) => {
  const cb = e.target.closest(".activity-checkbox");
  if (!cb)
    return;
  const id = Number(cb.getAttribute("data-activity-id"));
  if (id)
    toggleActivitySelection(id, cb.checked);
});
document.addEventListener("DOMContentLoaded", async () => {
  const sessionRes = await apiCall("/api/auth/me", {}, () => {});
  if (sessionRes.ok && sessionRes.data?.user && ["sysadmin", "gerente"].includes(sessionRes.data.user.role)) {
    setToken("cookie");
    setUser(sessionRes.data.user);
    localStorage.setItem("simp_user", JSON.stringify(user));
  }
  setAllowedRoles(["sysadmin", "gerente"]);
  updateAuthState();
  loadFilterOptions();
  loadActivities();
});
