// State management (ported from app.js lines 2-30)
var state = {
  screen: localStorage.getItem("simp_screen") || "login",
  user: JSON.parse(localStorage.getItem("simp_user") || "null"),
  token: null,
  atividade: JSON.parse(localStorage.getItem("simp_atividade") || "null"),
  scannedProducts: JSON.parse(localStorage.getItem("simp_scanned") || "[]"),
  expectedProducts: JSON.parse(localStorage.getItem("simp_expected") || "[]"),
  lastScanned: null,
  allowKeyboard: false
};

function saveState() {
  localStorage.setItem("simp_screen", state.screen);
  if (state.user)
    localStorage.setItem("simp_user", JSON.stringify(state.user));
  else
    localStorage.removeItem("simp_user");
  if (state.atividade)
    localStorage.setItem("simp_atividade", JSON.stringify(state.atividade));
  else
    localStorage.removeItem("simp_atividade");
  localStorage.setItem("simp_scanned", JSON.stringify(state.scannedProducts));
  localStorage.setItem("simp_expected", JSON.stringify(state.expectedProducts));
}

function resetActivityState() {
  state.atividade = null;
  state.scannedProducts = [];
  state.expectedProducts = [];
  saveState();
}

// UI helpers (ported from app.js lines 33-168)
function focusScannerInput(input) {
  if (!input) return;
  if (state.allowKeyboard) {
    input.readOnly = false;
    input.focus();
    return;
  }
  input.readOnly = true;
  input.focus();
  setTimeout(function() {
    input.readOnly = false;
  }, 50);
}

function refocusInput() {
  if (state.screen === "scanning" || state.screen === "consulta") {
    var inputId = state.screen === "scanning" ? "scan-input" : "consulta-input";
    var input = document.getElementById(inputId);
    if (state.allowKeyboard && document.activeElement === input)
      return;
    if (input && document.activeElement !== input) {
      var activeTag = document.activeElement ? document.activeElement.tagName.toLowerCase() : "";
      if (["select", "textarea", "input"].indexOf(activeTag) === -1) {
        focusScannerInput(input);
      }
    }
  }
}

function syncKeyboardUI() {
  var screens = [
    { btnId: "btn-toggle-keyboard", inputId: "scan-input" },
    { btnId: "btn-consulta-toggle-keyboard", inputId: "consulta-input" }
  ];
  for (var i = 0; i < screens.length; i++) {
    var s = screens[i];
    var btn = document.getElementById(s.btnId);
    var input = document.getElementById(s.inputId);
    if (btn)
      btn.classList.toggle("active", state.allowKeyboard);
    if (input) {
      input.inputMode = state.allowKeyboard ? "tel" : "none";
      if (state.allowKeyboard && state.screen === (s.inputId === "scan-input" ? "scanning" : "consulta")) {
        input.focus();
      } else if (!state.allowKeyboard && state.screen === (s.inputId === "scan-input" ? "scanning" : "consulta")) {
        focusScannerInput(input);
      }
    }
  }
}

function renderHistory() {
  var histEl = document.getElementById("scan-history");
  if (!histEl) return;
  var unique = [];
  var reversed = state.scannedProducts.slice(0).reverse();
  for (var ri = 0; ri < reversed.length; ri++) {
    var p = reversed[ri];
    var found = false;
    for (var ui = 0; ui < unique.length; ui++) {
      if (unique[ui].seqproduto === p.seqproduto) {
        found = true;
        break;
      }
    }
    if (!found) unique.push(p);
  }
  var html = "";
  for (var di = 0; di < unique.length; di++) {
    var item = unique[di];
    var reposicaoHtml = item.reposicao ? '<span title="Reposição">\uD83D\uDCE6</span>' : "";
    var statusClass = item.status === "OK" ? "status-ok" : "status-warn";
    var statusText = item.status === "OK" ? "&#x2714;&#xFE0F;" : "&#x26A0;&#xFE0F;";
    html += '<div class="history-item" data-seqproduto="' + item.seqproduto + '">' +
      '<span class="truncate" style="flex:1; margin-right: 8px;">' + sanitizeHtml(item.desccompleta || item.ean) + '</span>' +
      '<div style="display: flex; gap: 4px; align-items: center;">' +
        reposicaoHtml +
        '<span class="' + statusClass + '">' + statusText + '</span>' +
      '</div>' +
    '</div>';
  }
  histEl.innerHTML = html;
}

function showProductDetailModal(seqproduto) {
  var product = null;
  for (var si = 0; si < state.scannedProducts.length; si++) {
    if (state.scannedProducts[si].seqproduto === seqproduto) {
      product = state.scannedProducts[si];
      break;
    }
  }
  if (!product) return;
  var modal = document.getElementById("modal-product-detail");
  if (!modal) return;
  var descEl = document.getElementById("product-detail-desc");
  var eanEl = document.getElementById("product-detail-ean");
  var seqEl = document.getElementById("product-detail-seq");
  var statusEl = document.getElementById("product-detail-status");
  var localEl = document.getElementById("product-detail-local");
  if (descEl) descEl.innerText = product.desccompleta || "Sem descrição";
  if (eanEl) eanEl.innerText = product.ean || "N/A";
  if (seqEl) seqEl.innerText = String(product.seqproduto);
  if (statusEl) statusEl.innerText = product.status || "N/A";
  if (localEl) localEl.innerText = product.rua != null && product.predio != null ? product.rua + "/" + product.predio : "N/A";
  var toggleBtn = document.getElementById("btn-modal-toggle-reposicao");
  if (toggleBtn) {
    if (product.reposicao) {
      toggleBtn.innerHTML = sanitizeHtml('Remover Reposição <span style="filter: grayscale(0);">\uD83D\uDCE6</span>');
      toggleBtn.style.backgroundColor = "#fef3c7";
      toggleBtn.style.color = "#92400e";
      toggleBtn.style.borderColor = "#f59e0b";
    } else {
      toggleBtn.innerHTML = sanitizeHtml('Marcar Reposição <span style="filter: grayscale(1); opacity: 0.5;">\uD83D\uDCE6</span>');
      toggleBtn.style.backgroundColor = "#f8fafc";
      toggleBtn.style.color = "#64748b";
      toggleBtn.style.borderColor = "#e2e8f0";
    }
    toggleBtn.onclick = function() {
      product.reposicao = !product.reposicao;
      saveState();
      renderHistory();
      showProductDetailModal(seqproduto);
    };
  }
  var removeBtn = document.getElementById("btn-modal-remove");
  if (removeBtn) {
    var canRemove = state.atividade !== null && state.user !== null;
    removeBtn.style.display = canRemove ? "block" : "none";
    if (canRemove) {
      removeBtn.onclick = function() {
        if (!confirm("Remover este produto da leitura?")) {
          return;
        }
        var newScanned = [];
        for (var fi = 0; fi < state.scannedProducts.length; fi++) {
          if (state.scannedProducts[fi].seqproduto !== seqproduto) {
            newScanned.push(state.scannedProducts[fi]);
          }
        }
        state.scannedProducts = newScanned;
        saveState();
        renderHistory();
        closeProductDetailModal();
      };
    }
  }
  modal.classList.remove("hidden");
}

function closeProductDetailModal() {
  var el = document.getElementById("modal-product-detail");
  if (el) el.classList.add("hidden");
}

// Navigation (ported from app.js lines 170-236)
var _loadEmpresasCb = null;

function showScreen(screenId) {
  if (screenId === "consulta") {
    state.previousScreen = state.screen;
  }
  state.screen = screenId;
  saveState();
  var screens = document.querySelectorAll(".screen");
  for (var si = 0; si < screens.length; si++) {
    screens[si].classList.add("hidden");
  }
  var targetScreen = document.getElementById("screen-" + screenId);
  if (targetScreen) targetScreen.classList.remove("hidden");
  syncKeyboardUI();
  var isAuth = screenId !== "login";
  var hideGlobalHeader = screenId === "scanning" || screenId === "divergence" || screenId === "predio-switch" || screenId === "consulta";
  var headerActions = document.getElementById("header-actions");
  if (headerActions) headerActions.classList.toggle("hidden", !isAuth);
  var header = document.querySelector(".header");
  if (header) header.classList.toggle("hidden", hideGlobalHeader);
  if (isAuth) {
    var userName = (state.user && state.user.nome) || (state.user && state.user.username) || "Usuário";
    var displayUser = userName.length > 12 ? userName.substring(0, 10) + ".." : userName;
    var userEl = document.getElementById("header-user");
    if (userEl) userEl.innerText = displayUser;
  }
  if (screenId === "start") {
    var infoEl = document.getElementById("last-activity-info");
    if (infoEl) infoEl.innerText = "";
    if (_loadEmpresasCb) _loadEmpresasCb();
  }
  if (screenId === "scanning" && state.atividade) {
    var currentPredio = state.atividade.currentPredio || state.atividade.predio;
    var userName2 = (state.user && state.user.nome) || (state.user && state.user.username) || "Usuário";
    var displayUser2 = userName2.length > 12 ? userName2.substring(0, 10) + ".." : userName2;
    var scanAddress = document.getElementById("scan-address");
    if (scanAddress)
      scanAddress.innerText = state.atividade.rua + " | " + currentPredio + " \u2022 " + displayUser2;
    var scanFeedback = document.getElementById("scan-feedback");
    if (scanFeedback) scanFeedback.innerHTML = "";
    renderHistory();
    setTimeout(refocusInput, 100);
  }
}

function confirmExit(callback) {
  if (state.scannedProducts.length > 0 && (state.screen === "scanning" || state.screen === "divergence" || state.screen === "predio-switch")) {
    if (!confirm("Você tem produtos lidos. Deseja realmente sair SEM salvar a atividade? Os dados lidos serão perdidos.")) {
      return;
    }
  }
  callback();
}

// Auth (ported from app.js lines 239-263)
function logout() {
  confirmExit(function() {
    var xhr = new XMLHttpRequest();
    xhr.open("POST", "/api/auth/logout", true);
    xhr.withCredentials = true;
    xhr.send();
    state.user = null;
    state.token = null;
    saveState();
    window.location.href = "/atividades/login";
  });
}

// API loading functions (ported from app.js lines 266-312)
function loadEmpresas() {
  apiGet("/api/empresas", function(ok, status, data) {
    if (ok && Array.isArray(data)) {
      var select = document.getElementById("start-empresa");
      if (select) {
        var html = "";
        for (var ei = 0; ei < data.length; ei++) {
          var e = data[ei];
          html += '<option value="' + String(e.NROEMPRESA) + '">' + sanitizeHtml(String(e.NROEMPRESA) + " - " + String(e.NOMEREDUZIDO)) + '</option>';
        }
        select.innerHTML = sanitizeHtml(html);
        if (data.length > 0) {
          loadLocais(data[0].NROEMPRESA);
        }
      }
    }
  }, function() {
    window.location.href = "/atividades/login";
  });
}

function loadLocais(empresaId) {
  apiGet("/api/locais?empresa=" + encodeURIComponent(empresaId), function(ok, status, data) {
    var select = document.getElementById("start-local");
    if (select) {
      if (ok && Array.isArray(data) && data.length > 0) {
        var html = "";
        for (var li = 0; li < data.length; li++) {
          var loc = data[li];
          html += '<option value="' + sanitizeHtml(String(loc.SEQLOCAL)) + '">' + sanitizeHtml(String(loc.LOCAL)) + '</option>';
        }
        select.innerHTML = sanitizeHtml(html);
      } else {
        select.innerHTML = sanitizeHtml('<option value="">Nenhum local ativo</option>');
      }
    }
  }, function() {
    window.location.href = "/atividades/login";
  });
}

function fetchLastActivityInfo() {
  var empresa = document.getElementById("start-empresa") ? document.getElementById("start-empresa").value : null;
  var seqlocal = document.getElementById("start-local") ? document.getElementById("start-local").value : null;
  var rua = document.getElementById("start-rua") ? document.getElementById("start-rua").value : null;
  var predio = document.getElementById("start-predio") ? document.getElementById("start-predio").value : null;
  var infoEl = document.getElementById("last-activity-info");
  if (!infoEl) return;
  if (!state.token) return;
  if (!empresa || !seqlocal || !rua || !predio) {
    infoEl.innerText = "";
    return;
  }
  infoEl.innerText = "Buscando hist\u00F3rico...";
  apiGet("/api/atividades/last-info?empresa=" + encodeURIComponent(empresa) + "&seqlocal=" + encodeURIComponent(seqlocal) + "&rua=" + encodeURIComponent(rua) + "&predio=" + encodeURIComponent(predio), function(ok, status, data) {
    if (ok && data && data.dataFim) {
      infoEl.innerHTML = sanitizeHtml('Data \u00FAltima atividade: <strong style="color: #4f46e5;">' + formatDate(data.dataFim) + '</strong>');
    } else if (ok && !data) {
      infoEl.innerText = "Nenhuma atividade encontrada";
    } else {
      infoEl.innerText = "";
    }
  }, function() {
    window.location.href = "/atividades/login";
  });
}

// Scan functions (ported from app.js lines 315-418)
function startActivity() {
  var empSelect = document.getElementById("start-empresa");
  var empresa = empSelect.value;
  var empresaNome = empSelect.options[empSelect.selectedIndex].text.split(" - ")[1] || empSelect.options[empSelect.selectedIndex].text;
  var seqlocal = document.getElementById("start-local").value;
  var rua = document.getElementById("start-rua").value;
  var predio = document.getElementById("start-predio").value;
  var errEl = document.getElementById("start-error");
  if (!seqlocal) {
    if (errEl) {
      errEl.innerText = "Selecione um local v\u00E1lido";
      errEl.classList.remove("hidden");
    }
    return;
  }
  showLoader(true);
  apiGet("/api/produtos/local?empresa=" + encodeURIComponent(empresa) + "&seqlocal=" + encodeURIComponent(seqlocal) + "&rua=" + encodeURIComponent(rua) + "&predio=" + encodeURIComponent(predio), function(ok, status, data) {
    showLoader(false);
    if (ok && Array.isArray(data) && data.length > 0) {
      state.expectedProducts = data;
      state.atividade = {
        id: 0,
        empresa: empresa,
        empresaNome: empresaNome,
        seqlocal: seqlocal,
        rua: rua,
        predio: predio,
        predios: [predio],
        currentPredio: predio,
        status: "aberta",
        dataInicio: new Date().toISOString()
      };
      state.scannedProducts = [];
      saveState();
      showScreen("scanning");
    } else {
      if (errEl) {
        errEl.innerText = "Endere\u00E7o n\u00E3o encontrado ou sem produtos";
        errEl.classList.remove("hidden");
      }
    }
  }, function() {
    showLoader(false);
    window.location.href = "/atividades/login";
  });
}

function finalizeActivity() {
  if (!state.atividade) return;
  if (!confirm("Tem certeza que deseja finalizar a atividade?")) return;
  showLoader(true);
  var predios = state.atividade.predios || [state.atividade.predio];
  var payload = {
    empresa: Number(state.atividade.empresa),
    seqlocal: Number(state.atividade.seqlocal),
    rua: state.atividade.rua,
    predio: predios,
    readProducts: state.scannedProducts,
    expectedProducts: state.expectedProducts
  };
  apiPost("/api/atividades/finalizar", payload, function(ok, status, data) {
    showLoader(false);
    if (ok) {
      resetActivityState();
      var rp = data || {};
      var divergences = rp.divergences || [];
      var ruptures = rp.ruptures || [];
      var replenishments = rp.replenishments || [];
      var reportIdEl = document.getElementById("report-id");
      var countDivEl = document.getElementById("count-div");
      var countRupEl = document.getElementById("count-rup");
      var countRepEl = document.getElementById("count-rep");
      if (reportIdEl) reportIdEl.innerText = rp.atividadeId || "--";
      if (countDivEl) countDivEl.innerText = divergences.length.toString();
      if (countRupEl) countRupEl.innerText = ruptures.length.toString();
      if (countRepEl) countRepEl.innerText = replenishments.length.toString();
      var divEl = document.getElementById("report-divergences");
      var rupEl = document.getElementById("report-ruptures");
      var repEl = document.getElementById("report-replenishments");
      function itemHtml(p) {
        return '<div style="padding: 0.5rem 0; border-bottom: 1px solid rgba(0,0,0,0.05);">' +
          '<strong style="color: #334155;">SEQ: ' + sanitizeHtml(String(p.seqproduto || p.ean || "-")) + '</strong>' +
          '<p style="color: #64748b; margin-top: 0.25rem;">' + sanitizeHtml(p.desccompleta || "-") + '</p>' +
        '</div>';
      }
      if (divEl)
        divEl.innerHTML = sanitizeHtml(divergences.map(itemHtml).join("") || '<p style="color: #94a3b8; font-style: italic;">Nenhuma diverg\u00EAncia</p>');
      if (rupEl)
        rupEl.innerHTML = sanitizeHtml(ruptures.map(itemHtml).join("") || '<p style="color: #94a3b8; font-style: italic;">Nenhuma ruptura</p>');
      if (repEl)
        repEl.innerHTML = sanitizeHtml(replenishments.map(itemHtml).join("") || '<p style="color: #94a3b8; font-style: italic;">Nenhuma reposi\u00E7\u00E3o</p>');
      showScreen("report");
    } else {
      if (status === 401) return;
      alert("Erro ao finalizar: " + ((data && data.error) ? data.error : "Erro de conex\u00E3o") + "\n\nSeus dados est\u00E3o salvos localmente. Tente novamente quando recuperar o sinal.");
    }
  }, function() {
    showLoader(false);
    window.location.href = "/atividades/login";
  });
}

// DOMContentLoaded
document.addEventListener("DOMContentLoaded", function() {
  // Session check
  apiGet("/api/auth/me", function(ok, status, data) {
    if (ok && data && data.user) {
      state.token = "cookie";
      state.user = data.user;
      saveState();
    }
    // Resolve entry screen and show
    var hasUser = state.user !== null;
    var hasAtividade = state.atividade !== null;
    var entryScreen;
    if (!hasUser) entryScreen = "login";
    else if (hasAtividade) entryScreen = "scanning";
    else if (state.screen === "login") entryScreen = "start";
    else entryScreen = state.screen;
    showScreen(entryScreen);
  }, function() {
    showScreen("login");
  });

  // Global exports
  window.showProductDetailModal = showProductDetailModal;
  window.closeProductDetailModal = closeProductDetailModal;
  _loadEmpresasCb = loadEmpresas;

  // History click handler
  var histEl = document.getElementById("scan-history");
  if (histEl) {
    histEl.addEventListener("click", function(e) {
      var target = e.target.closest(".history-item");
      if (target) {
        var seq = target.getAttribute("data-seqproduto");
        if (seq) showProductDetailModal(Number(seq));
      }
    });
  }

  // Logout buttons
  var btnLogout = document.getElementById("btn-logout");
  if (btnLogout) btnLogout.addEventListener("click", logout);
  var logouts = document.querySelectorAll(".js-btn-logout");
  for (var loi = 0; loi < logouts.length; loi++) {
    logouts[loi].addEventListener("click", logout);
  }

  // Keyboard prevention
  function preventKeyboard(e) {
    if (!state.allowKeyboard) {
      var target = e.currentTarget;
      if (target && (state.screen === "scanning" || state.screen === "consulta")) {
        e.preventDefault();
        focusScannerInput(target);
      }
    }
  }

  // Scanner input handlers
  var scanInput = document.getElementById("scan-input");
  if (scanInput) {
    scanInput.addEventListener("blur", function() {
      setTimeout(refocusInput, 300);
    });
    scanInput.addEventListener("mousedown", preventKeyboard);
    scanInput.addEventListener("touchstart", preventKeyboard);
  }

  // Consulta input handlers
  var consultaInput = document.getElementById("consulta-input");
  if (consultaInput) {
    consultaInput.addEventListener("blur", function() {
      setTimeout(refocusInput, 300);
    });
    consultaInput.addEventListener("mousedown", preventKeyboard);
    consultaInput.addEventListener("touchstart", preventKeyboard);
  }

  // Global focus lock
  var handleGlobalFocusLock = function(e) {
    if (!state.allowKeyboard && (state.screen === "scanning" || state.screen === "consulta")) {
      var target = e.target;
      var tag = target.tagName;
      var isInteractive = tag === "BUTTON" || tag === "INPUT" || tag === "SELECT" || tag === "TEXTAREA" || tag === "A" || tag === "SPAN" || tag === "I";
      isInteractive = isInteractive || target.closest("button") || target.closest(".history-item") || target.closest(".modal-content") || target.closest(".scan-history");
      if (!isInteractive && e.type !== "touchmove") {
        e.preventDefault();
      }
    }
  };
  document.addEventListener("touchstart", handleGlobalFocusLock, { passive: true });
  document.addEventListener("mousedown", handleGlobalFocusLock);

  // Keyboard toggle
  var toggleKeyboard = function() {
    state.allowKeyboard = !state.allowKeyboard;
    syncKeyboardUI();
    saveState();
  };
  var btnToggle = document.getElementById("btn-toggle-keyboard");
  if (btnToggle) {
    btnToggle.addEventListener("click", function(e) {
      toggleKeyboard();
      if (e.currentTarget instanceof HTMLElement) e.currentTarget.blur();
      setTimeout(refocusInput, 100);
    });
  }
  var btnConsultaToggle = document.getElementById("btn-consulta-toggle-keyboard");
  if (btnConsultaToggle) {
    btnConsultaToggle.addEventListener("click", function(e) {
      toggleKeyboard();
      if (e.currentTarget instanceof HTMLElement) e.currentTarget.blur();
      setTimeout(refocusInput, 100);
    });
  }

  // Password toggles
  var toggles = document.querySelectorAll(".password-toggle");
  for (var ti = 0; ti < toggles.length; ti++) {
    (function(toggle) {
      toggle.addEventListener("click", function(e) {
        var input = e.target.previousElementSibling;
        if (input && input.tagName === "INPUT") {
          if (input.type === "password") {
            input.type = "text";
            e.target.innerText = "\uD83D\uDE48";
          } else {
            input.type = "password";
            e.target.innerText = "\uD83D\uDC41\uFE0F";
          }
        }
      });
    })(toggles[ti]);
  }

  // Navigation buttons
  var btnBack = document.getElementById("btn-back-to-start");
  if (btnBack) {
    btnBack.addEventListener("click", function() {
      confirmExit(function() { showScreen("start"); });
    });
  }

  // Form handlers
  var startEmpresa = document.getElementById("start-empresa");
  if (startEmpresa) {
    startEmpresa.addEventListener("change", function(e) {
      loadLocais(e.target.value);
      fetchLastActivityInfo();
    });
  }
  var startLocal = document.getElementById("start-local");
  if (startLocal) {
    startLocal.addEventListener("change", fetchLastActivityInfo);
  }
  var startRua = document.getElementById("start-rua");
  if (startRua) {
    startRua.addEventListener("blur", fetchLastActivityInfo);
  }
  var startPredio = document.getElementById("start-predio");
  if (startPredio) {
    startPredio.addEventListener("blur", fetchLastActivityInfo);
  }

  // Start form
  var formStart = document.getElementById("form-start");
  if (formStart) {
    formStart.addEventListener("submit", function(e) {
      e.preventDefault();
      if (state.atividade && state.scannedProducts.length > 0) {
        if (!confirm("Voc\u00EA j\u00E1 possui uma atividade em andamento com produtos lidos.\n\nDeseja DESCARTAR os dados e iniciar uma nova?\n\n\u2022 Clique OK para descartar e come\u00E7ar nova.\n\u2022 Clique CANCELAR para voltar \u00E0 atividade em andamento.")) {
          showScreen("scanning");
          return;
        }
        resetActivityState();
      }
      startActivity();
    });
  }

  // Scan form
  var formScan = document.getElementById("form-scan");
  if (formScan) {
    formScan.addEventListener("submit", function(e) {
      e.preventDefault();
      var input = document.getElementById("scan-input");
      var code = input.value.trim();
      if (!code) return;
      if (!state.atividade) return;
      showLoader(true);
      apiGet("/api/produtos/ean/" + encodeURIComponent(code) + "?empresa=" + state.atividade.empresa + "&seqlocal=" + state.atividade.seqlocal, function(ok, status, data) {
        showLoader(false);
        var feedback = document.getElementById("scan-feedback");
        if (!feedback) return;
        if (!ok) {
          playBeep("error");
          feedback.innerHTML = sanitizeHtml('<div style="color: #ef4444; font-weight: bold;">\u274C Produto n\u00E3o encontrado</div>');
          input.select();
          return;
        }
        input.value = "";
        focusScannerInput(input);
        var currentPredio = state.atividade.currentPredio || state.atividade.predio;
        var isNullAddress = data.rua == null || data.predio == null;
        var sameRua = data.rua === state.atividade.rua;
        var samePredio = String(data.predio) === String(currentPredio);
        var alreadyScanned = false;
        for (var asi = 0; asi < state.scannedProducts.length; asi++) {
          if (state.scannedProducts[asi].seqproduto === data.seqproduto) {
            alreadyScanned = true;
            break;
          }
        }
        if (alreadyScanned) {
          playBeep("warning");
          feedback.innerHTML = sanitizeHtml('<div style="color: #f59e0b; font-weight: bold;">\u26A0\uFE0F Produto j\u00E1 lido nesta atividade!</div>');
          return;
        }
        var scanStatus = "OK";
        if (isNullAddress || !sameRua || !samePredio) {
          scanStatus = "DIVERGENTE";
        }
        state.scannedProducts.push({
          seqproduto: data.seqproduto,
          ean: code,
          rua: data.rua,
          predio: String(currentPredio),
          desccompleta: data.desccompleta,
          status: scanStatus,
          reposicao: false
        });
        saveState();
        state.lastScanned = data;
        if (isNullAddress || !sameRua) {
          playBeep("warning");
          var reason = isNullAddress ? "S/ Endere\u00E7o" : "Rua Divergente";
          feedback.innerHTML = sanitizeHtml('<div style="color: #f59e0b; font-weight: bold;">\u26A0\uFE0F ' + reason + ': ' + sanitizeHtml(data.desccompleta) + '</div>');
          renderHistory();
        } else if (!samePredio) {
          playBeep("warning");
          var displayPredio = data.predio != null ? data.predio : "N/A";
          var predioSwitchDesc = document.getElementById("predio-switch-desc");
          if (predioSwitchDesc)
            predioSwitchDesc.innerText = data.desccompleta + " pertence ao Pr\u00E9dio " + displayPredio + " (mesma rua).";
          var predioSwitchNew = document.getElementById("predio-switch-new");
          if (predioSwitchNew)
            predioSwitchNew.innerText = displayPredio.toString();
          var predioSwitchCurrent = document.getElementById("predio-switch-current");
          if (predioSwitchCurrent)
            predioSwitchCurrent.innerText = currentPredio.toString();
          renderHistory();
          showScreen("predio-switch");
        } else {
          playBeep("success");
          feedback.innerHTML = sanitizeHtml('<div style="color: #10b981; font-weight: bold;">\u2705 Lido: ' + sanitizeHtml(data.desccompleta) + '</div>');
          renderHistory();
        }
      }, function() {
        showLoader(false);
        window.location.href = "/atividades/login";
      });
    });
  }

  // Finalize button
  var btnFinalize = document.getElementById("btn-finalize");
  if (btnFinalize) {
    btnFinalize.addEventListener("click", finalizeActivity);
  }

  // Building switch: Yes
  var btnPredioYes = document.getElementById("btn-predio-switch-yes");
  if (btnPredioYes) {
    btnPredioYes.addEventListener("click", function() {
      if (!state.atividade || !state.lastScanned) return;
      try {
        var newPredio = String(state.lastScanned.predio);
        var predios = state.atividade.predios || [state.atividade.predio];
        var isNewBuilding = predios.indexOf(newPredio) === -1;
        if (isNewBuilding) {
          predios.push(newPredio);
          state.atividade.predios = predios;
        }
        state.atividade.currentPredio = newPredio;
        if (isNewBuilding) {
          showLoader(true);
          apiGet("/api/produtos/local?empresa=" + state.atividade.empresa + "&seqlocal=" + state.atividade.seqlocal + "&rua=" + state.atividade.rua + "&predio=" + encodeURIComponent(newPredio), function(ok, result, resultData) {
            showLoader(false);
            if (ok && Array.isArray(resultData) && resultData.length > 0) {
              var existingSeqs = [];
              for (var esi = 0; esi < state.expectedProducts.length; esi++) {
                existingSeqs.push(state.expectedProducts[esi].seqproduto);
              }
              for (var npi = 0; npi < resultData.length; npi++) {
                var np = resultData[npi];
                var seqFound = false;
                for (var esj = 0; esj < existingSeqs.length; esj++) {
                  if (existingSeqs[esj] === np.seqproduto) {
                    seqFound = true;
                    break;
                  }
                }
                if (!seqFound) {
                  state.expectedProducts.push(np);
                }
              }
            }
            finishPredioSwitch(newPredio);
          }, function() {
            showLoader(false);
            window.location.href = "/atividades/login";
          });
        } else {
          finishPredioSwitch(newPredio);
        }
      } catch (e) {
        console.error("Error during building switch:", e);
        showScreen("scanning");
        var fb = document.getElementById("scan-feedback");
        if (fb) fb.innerHTML = sanitizeHtml('<div style="color: #ef4444; font-weight: bold;">\u274C Erro ao trocar de pr\u00E9dio</div>');
      }
    });
  }

  function finishPredioSwitch(newPredio) {
    var isExpected = false;
    for (var epi = 0; epi < state.expectedProducts.length; epi++) {
      if (state.expectedProducts[epi].seqproduto === state.lastScanned.seqproduto) {
        isExpected = true;
        break;
      }
    }
    var newStatus = isExpected ? "OK" : "DIVERGENTE";
    var lastIdx = -1;
    for (var lsi = state.scannedProducts.length - 1; lsi >= 0; lsi--) {
      if (state.scannedProducts[lsi].seqproduto === state.lastScanned.seqproduto) {
        lastIdx = lsi;
        break;
      }
    }
    if (lastIdx >= 0) {
      state.scannedProducts[lastIdx].status = newStatus;
      state.scannedProducts[lastIdx].predio = newPredio;
    }
    saveState();
    showScreen("scanning");
    var feedback = document.getElementById("scan-feedback");
    if (feedback) {
      if (isExpected) {
        playBeep("success");
        feedback.innerHTML = sanitizeHtml('<div style="color: #10b981; font-weight: bold;">\u2705 Pr\u00E9dio ' + sanitizeHtml(newPredio) + ' agora \u00E9 o pr\u00E9dio atual. Produto OK.</div>');
      } else {
        playBeep("warning");
        feedback.innerHTML = sanitizeHtml('<div style="color: #f59e0b; font-weight: bold;">\u26A0\uFE0F Pr\u00E9dio ' + sanitizeHtml(newPredio) + ' agora \u00E9 o pr\u00E9dio atual, por\u00E9m produto n\u00E3o esperado!</div>');
      }
    }
    state.lastScanned = null;
    renderHistory();
  }

  // Building switch: No
  var btnPredioNo = document.getElementById("btn-predio-switch-no");
  if (btnPredioNo) {
    btnPredioNo.addEventListener("click", function() {
      showScreen("scanning");
      var feedback = document.getElementById("scan-feedback");
      if (feedback && state.lastScanned) {
        feedback.innerHTML = sanitizeHtml('<div style="color: #f59e0b; font-weight: bold;">\u26A0\uFE0F Divergente: ' + sanitizeHtml(state.lastScanned.desccompleta) + '</div>');
      }
      renderHistory();
    });
  }

  // Report OK button
  var btnReportOk = document.getElementById("btn-report-ok");
  if (btnReportOk) {
    btnReportOk.addEventListener("click", function() {
      showScreen("start");
    });
  }

  // Open consulta
  var openConsulta = function() {
    state.lastScanned = null;
    saveState();
    setConsultaMode("codigo");
    showScreen("consulta");
    var lojaNome = "";
    if (state.previousScreen === "scanning" && state.atividade) {
      lojaNome = "\u2022 " + (state.atividade.empresaNome || "Loja " + state.atividade.empresa);
    } else {
      var empSelect = document.getElementById("start-empresa");
      if (empSelect && empSelect.selectedIndex >= 0) {
        var fullText = empSelect.options[empSelect.selectedIndex].text;
        lojaNome = "\u2022 " + (fullText.split(" - ")[1] || fullText);
      }
    }
    var headerLoja = document.getElementById("consulta-header-loja");
    if (headerLoja) headerLoja.innerText = lojaNome;
    var input = document.getElementById("consulta-input");
    if (input) {
      input.value = "";
      input.focus();
    }
    var cr = document.getElementById("consulta-result");
    if (cr) cr.classList.add("hidden");
    var crl = document.getElementById("consulta-result-list");
    if (crl) {
      crl.classList.add("hidden");
      crl.innerHTML = "";
    }
    var ce = document.getElementById("consulta-empty");
    if (ce) ce.classList.remove("hidden");
  };

  var btnGoConsulta = document.getElementById("btn-go-consulta");
  if (btnGoConsulta) btnGoConsulta.addEventListener("click", openConsulta);
  var btnStartConsulta = document.getElementById("btn-start-consulta");
  if (btnStartConsulta) btnStartConsulta.addEventListener("click", openConsulta);
});
