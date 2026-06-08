// src/client/app/state.ts
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

// src/client/app/ui.ts
function focusScannerInput(input) {
  if (!input)
    return;
  if (state.allowKeyboard) {
    input.readOnly = false;
    input.focus();
    return;
  }
  input.readOnly = true;
  input.focus();
  setTimeout(() => {
    input.readOnly = false;
  }, 50);
}
function refocusInput() {
  if (state.screen === "scanning" || state.screen === "consulta") {
    const inputId = state.screen === "scanning" ? "scan-input" : "consulta-input";
    const input = document.getElementById(inputId);
    const reauthModal = document.getElementById("modal-reauth");
    const isReauthVisible = reauthModal && !reauthModal.classList.contains("hidden");
    if (state.allowKeyboard && document.activeElement === input)
      return;
    if (input && document.activeElement !== input && !isReauthVisible) {
      const activeTag = document.activeElement ? document.activeElement.tagName.toLowerCase() : "";
      if (!["select", "textarea", "input"].includes(activeTag)) {
        focusScannerInput(input);
      }
    }
  }
}
function syncKeyboardUI() {
  const screens = [
    { btnId: "btn-toggle-keyboard", inputId: "scan-input" },
    { btnId: "btn-consulta-toggle-keyboard", inputId: "consulta-input" }
  ];
  for (const s of screens) {
    const btn = document.getElementById(s.btnId);
    const input = document.getElementById(s.inputId);
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
function syncReposicaoUI() {}
function renderHistory() {
  const histEl = document.getElementById("scan-history");
  if (!histEl)
    return;
  const unique = [];
  const reversed = [...state.scannedProducts].reverse();
  for (const p of reversed) {
    if (!unique.find((x) => x.seqproduto === p.seqproduto))
      unique.push(p);
  }
  const displayList = unique;
  histEl.innerHTML = sanitizeHtml(displayList.map((p) => `
        <div class="history-item" data-seqproduto="${p.seqproduto}">
            <span class="truncate" style="flex:1; margin-right: 8px;">${p.desccompleta || p.ean}</span>
            <div style="display: flex; gap: 4px; align-items: center;">
                ${p.reposicao ? '<span title="Reposição">\uD83D\uDCE6</span>' : ""}
                <span class="${p.status === "OK" ? "status-ok" : "status-warn"}">${p.status === "OK" ? "✔️" : "⚠️"}</span>
            </div>
        </div>
    `).join(""));
}
function showProductDetailModal(seqproduto) {
  const product = state.scannedProducts.find((p) => p.seqproduto === seqproduto);
  if (!product)
    return;
  const modal = document.getElementById("modal-product-detail");
  if (!modal)
    return;
  const descEl = document.getElementById("product-detail-desc");
  const eanEl = document.getElementById("product-detail-ean");
  const seqEl = document.getElementById("product-detail-seq");
  const statusEl = document.getElementById("product-detail-status");
  const localEl = document.getElementById("product-detail-local");
  if (descEl)
    descEl.innerText = product.desccompleta || "Sem descrição";
  if (eanEl)
    eanEl.innerText = product.ean || "N/A";
  if (seqEl)
    seqEl.innerText = String(product.seqproduto);
  if (statusEl)
    statusEl.innerText = product.status || "N/A";
  if (localEl)
    localEl.innerText = product.rua != null && product.predio != null ? `${product.rua}/${product.predio}` : "N/A";
  const toggleBtn = document.getElementById("btn-modal-toggle-reposicao");
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
    toggleBtn.onclick = () => {
      product.reposicao = !product.reposicao;
      saveState();
      renderHistory();
      showProductDetailModal(seqproduto);
    };
  }
  const removeBtn = document.getElementById("btn-modal-remove");
  if (removeBtn) {
    const canRemove = state.atividade !== null && state.user !== null;
    removeBtn.style.display = canRemove ? "block" : "none";
    if (canRemove) {
      removeBtn.onclick = () => {
        if (!confirm("Remover este produto da leitura?")) {
          return;
        }
        state.scannedProducts = state.scannedProducts.filter((p) => p.seqproduto !== seqproduto);
        saveState();
        renderHistory();
        closeProductDetailModal();
      };
    }
  }
  modal.classList.remove("hidden");
}
function closeProductDetailModal() {
  document.getElementById("modal-product-detail")?.classList.add("hidden");
}

// src/client/app/navigation.ts
var _loadEmpresasCb = null;
function setLoadEmpresasCb(cb) {
  _loadEmpresasCb = cb;
}
function showScreen(screenId) {
  if (screenId === "consulta") {
    state.previousScreen = state.screen;
  }
  state.screen = screenId;
  saveState();
  const screens = document.querySelectorAll(".screen");
  for (let i = 0;i < screens.length; i++) {
    screens[i].classList.add("hidden");
  }
  const targetScreen = document.getElementById(`screen-${screenId}`);
  if (targetScreen)
    targetScreen.classList.remove("hidden");
  syncKeyboardUI();
  syncReposicaoUI();
  const isAuth = screenId !== "login";
  const hideGlobalHeader = [
    "scanning",
    "divergence",
    "predio-switch",
    "consulta"
  ].includes(screenId);
  const headerActions = document.getElementById("header-actions");
  if (headerActions)
    headerActions.classList.toggle("hidden", !isAuth);
  const header = document.querySelector(".header");
  if (header)
    header.classList.toggle("hidden", hideGlobalHeader);
  if (isAuth) {
    const userName = state.user?.nome || state.user?.username || "Usuário";
    const displayUser = userName.length > 12 ? `${userName.substring(0, 10)}..` : userName;
    const userEl = document.getElementById("header-user");
    if (userEl)
      userEl.innerText = displayUser;
  }
  if (screenId === "start") {
    const infoEl = document.getElementById("last-activity-info");
    if (infoEl)
      infoEl.innerText = "";
    _loadEmpresasCb?.();
  }
  if (screenId === "scanning" && state.atividade) {
    const currentPredio = state.atividade.currentPredio || state.atividade.predio;
    const userName = state.user?.nome || state.user?.username || "Usuário";
    const displayUser = userName.length > 12 ? `${userName.substring(0, 10)}..` : userName;
    const scanAddress = document.getElementById("scan-address");
    if (scanAddress)
      scanAddress.innerText = `${state.atividade.rua} | ${currentPredio} • ${displayUser}`;
    const scanFeedback = document.getElementById("scan-feedback");
    if (scanFeedback)
      scanFeedback.innerHTML = "";
    renderHistory();
    setTimeout(refocusInput, 100);
  }
}
async function confirmExit(callback) {
  if (state.scannedProducts.length > 0 && ["scanning", "divergence", "predio-switch"].includes(state.screen)) {
    if (!confirm("Você tem produtos lidos. Deseja realmente sair SEM salvar a atividade? Os dados lidos serão perdidos.")) {
      return;
    }
  }
  callback();
}

// src/client/app/auth.ts
function logout() {
  confirmExit(() => {
    fetch("/api/auth/logout", { method: "POST", credentials: "include" }).catch(() => {});
    state.user = null;
    state.token = null;
    saveState();
    showScreen("login");
  });
}
function showReauthModal(show) {
  const modal = document.getElementById("modal-reauth");
  if (!modal)
    return;
  modal.classList.toggle("hidden", !show);
  if (show) {
    const passInput = document.getElementById("reauth-password");
    if (passInput) {
      passInput.value = "";
      passInput.focus();
    }
    const errEl = document.getElementById("reauth-error");
    if (errEl)
      errEl.classList.add("hidden");
  }
}

// src/client/app/api.ts
async function loadEmpresas() {
  const { ok, data } = await apiCall("/api/empresas", {}, () => showReauthModal(true));
  if (ok && Array.isArray(data)) {
    const select = document.getElementById("start-empresa");
    if (select) {
      select.innerHTML = sanitizeHtml(data.map((e) => `<option value="${String(e.NROEMPRESA)}">${String(e.NROEMPRESA)} - ${String(e.NOMEREDUZIDO)}</option>`).join(""));
      if (data.length > 0) {
        loadLocais(data[0].NROEMPRESA);
      }
    }
  }
}
async function loadLocais(empresaId) {
  const { ok, data } = await apiCall(`/api/locais?empresa=${empresaId}`, {}, () => showReauthModal(true));
  const select = document.getElementById("start-local");
  if (select) {
    if (ok && Array.isArray(data) && data.length > 0) {
      select.innerHTML = sanitizeHtml(data.map((e) => `<option value="${String(e.SEQLOCAL)}">${String(e.LOCAL)}</option>`).join(""));
    } else {
      select.innerHTML = sanitizeHtml('<option value="">Nenhum local ativo</option>');
    }
  }
}
async function fetchLastActivityInfo() {
  const empresa = document.getElementById("start-empresa")?.value;
  const seqlocal = document.getElementById("start-local")?.value;
  const rua = document.getElementById("start-rua")?.value;
  const predio = document.getElementById("start-predio")?.value;
  const infoEl = document.getElementById("last-activity-info");
  if (!infoEl)
    return;
  if (!state.token)
    return;
  if (!empresa || !seqlocal || !rua || !predio) {
    infoEl.innerText = "";
    return;
  }
  infoEl.innerText = "Buscando histórico...";
  const res = await apiCall(`/api/atividades/last-info?empresa=${empresa}&seqlocal=${seqlocal}&rua=${rua}&predio=${predio}`, {}, () => showReauthModal(true));
  if (res.ok && res.data?.dataFim) {
    infoEl.innerHTML = sanitizeHtml(`Data última atividade: <strong style="color: #4f46e5;">${formatDate(res.data.dataFim)}</strong>`);
  } else if (res.ok && !res.data) {
    infoEl.innerText = "Nenhuma atividade encontrada";
  } else {
    infoEl.innerText = "";
  }
}

// src/client/app/scan.ts
async function startActivity() {
  const empSelect = document.getElementById("start-empresa");
  const empresa = empSelect.value;
  const empresaNome = empSelect.options[empSelect.selectedIndex].text.split(" - ")[1] || empSelect.options[empSelect.selectedIndex].text;
  const seqlocal = document.getElementById("start-local").value;
  const rua = document.getElementById("start-rua").value;
  const predio = document.getElementById("start-predio").value;
  const errEl = document.getElementById("start-error");
  if (!seqlocal) {
    if (errEl) {
      errEl.innerText = "Selecione um local válido";
      errEl.classList.remove("hidden");
    }
    return;
  }
  showLoader(true);
  const { ok, data } = await apiCall(`/api/produtos/local?empresa=${empresa}&seqlocal=${seqlocal}&rua=${rua}&predio=${predio}`, {}, () => showReauthModal(true));
  showLoader(false);
  if (ok && Array.isArray(data) && data.length > 0) {
    state.expectedProducts = data;
    state.atividade = {
      id: 0,
      empresa,
      empresaNome,
      seqlocal,
      rua,
      predio,
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
      errEl.innerText = "Endereço não encontrado ou sem produtos";
      errEl.classList.remove("hidden");
    }
  }
}
async function finalizeActivity() {
  if (!state.atividade)
    return;
  if (!confirm("Tem certeza que deseja finalizar a atividade?")) {
    return;
  }
  showLoader(true);
  const predios = state.atividade.predios || [state.atividade.predio];
  const payload = {
    empresa: Number(state.atividade.empresa),
    seqlocal: Number(state.atividade.seqlocal),
    rua: state.atividade.rua,
    predio: predios,
    readProducts: state.scannedProducts,
    expectedProducts: state.expectedProducts
  };
  const result = await apiCall("/api/atividades/finalizar", {
    method: "POST",
    body: JSON.stringify(payload)
  }, () => showReauthModal(true));
  showLoader(false);
  if (result.ok) {
    resetActivityState();
    const rp = result.data || {};
    const divergences = rp.divergences || [];
    const ruptures = rp.ruptures || [];
    const replenishments = rp.replenishments || [];
    const reportIdEl = document.getElementById("report-id");
    const countDivEl = document.getElementById("count-div");
    const countRupEl = document.getElementById("count-rup");
    const countRepEl = document.getElementById("count-rep");
    if (reportIdEl)
      reportIdEl.innerText = rp.atividadeId || "--";
    if (countDivEl)
      countDivEl.innerText = divergences.length.toString();
    if (countRupEl)
      countRupEl.innerText = ruptures.length.toString();
    if (countRepEl)
      countRepEl.innerText = replenishments.length.toString();
    const divEl = document.getElementById("report-divergences");
    const rupEl = document.getElementById("report-ruptures");
    const repEl = document.getElementById("report-replenishments");
    const itemHtml = (p) => `<div style="padding: 0.5rem 0; border-bottom: 1px solid rgba(0,0,0,0.05);">
            <strong style="color: #334155;">SEQ: ${p.seqproduto || p.ean || "-"}</strong>
            <p style="color: #64748b; margin-top: 0.25rem;">${p.desccompleta || "-"}</p>
        </div>`;
    if (divEl)
      divEl.innerHTML = sanitizeHtml(divergences.map(itemHtml).join("") || '<p style="color: #94a3b8; font-style: italic;">Nenhuma divergência</p>');
    if (rupEl)
      rupEl.innerHTML = sanitizeHtml(ruptures.map(itemHtml).join("") || '<p style="color: #94a3b8; font-style: italic;">Nenhuma ruptura</p>');
    if (repEl)
      repEl.innerHTML = sanitizeHtml(replenishments.map(itemHtml).join("") || '<p style="color: #94a3b8; font-style: italic;">Nenhuma reposição</p>');
    showScreen("report");
  } else {
    if (result.status === 401) {
      return;
    }
    alert("Erro ao finalizar: " + (result.data.error || "Erro de conexão") + `

Seus dados estão salvos localmente. Tente novamente quando recuperar o sinal.`);
  }
}

// src/client/app/session.ts
function resolveAtividadesEntryScreen(hasUser, screen, hasAtividade) {
  if (!hasUser)
    return "login";
  if (hasAtividade)
    return "scanning";
  if (screen === "login")
    return "start";
  return screen;
}

// src/client/app/index.ts
setLoadEmpresasCb(loadEmpresas);
window.showProductDetailModal = showProductDetailModal;
window.closeProductDetailModal = closeProductDetailModal;
document.getElementById("scan-history")?.addEventListener("click", (e) => {
  const target = e.target.closest(".history-item");
  if (target) {
    const seq = target.getAttribute("data-seqproduto");
    if (seq)
      showProductDetailModal(Number(seq));
  }
});
document.addEventListener("DOMContentLoaded", async () => {
  const sessionRes = await apiCall("/api/auth/me", {}, () => {});
  if (sessionRes.ok && sessionRes.data?.user) {
    state.token = "cookie";
    state.user = sessionRes.data.user;
    saveState();
  }
  showScreen(resolveAtividadesEntryScreen(Boolean(state.user), state.screen, Boolean(state.atividade)));
  const btnLogout = document.getElementById("btn-logout");
  if (btnLogout)
    btnLogout.addEventListener("click", logout);
  const logouts = document.querySelectorAll(".js-btn-logout");
  for (let i = 0;i < logouts.length; i++) {
    logouts[i].addEventListener("click", logout);
  }
  function preventKeyboard(e) {
    if (!state.allowKeyboard) {
      const target = e.currentTarget;
      if (target && (state.screen === "scanning" || state.screen === "consulta")) {
        e.preventDefault();
        focusScannerInput(target);
      }
    }
  }
  const scanInput = document.getElementById("scan-input");
  if (scanInput) {
    scanInput.addEventListener("blur", () => {
      setTimeout(refocusInput, 300);
    });
    scanInput.addEventListener("mousedown", preventKeyboard);
    scanInput.addEventListener("touchstart", preventKeyboard);
  }
  const consultaInput = document.getElementById("consulta-input");
  if (consultaInput) {
    consultaInput.addEventListener("blur", () => {
      setTimeout(refocusInput, 300);
    });
    consultaInput.addEventListener("mousedown", preventKeyboard);
    consultaInput.addEventListener("touchstart", preventKeyboard);
  }
  const handleGlobalFocusLock = (e) => {
    if (!state.allowKeyboard && (state.screen === "scanning" || state.screen === "consulta")) {
      const target = e.target;
      const isInteractive = ["BUTTON", "INPUT", "SELECT", "TEXTAREA", "A", "SPAN", "I"].includes(target.tagName) || target.closest("button") || target.closest(".history-item") || target.closest(".modal-content") || target.closest(".scan-history");
      if (!isInteractive && e.type !== "touchmove") {
        e.preventDefault();
      }
    }
  };
  document.addEventListener("touchstart", handleGlobalFocusLock, {
    passive: true
  });
  document.addEventListener("mousedown", handleGlobalFocusLock);
  const toggleKeyboard = () => {
    state.allowKeyboard = !state.allowKeyboard;
    syncKeyboardUI();
    saveState();
  };
  document.getElementById("btn-toggle-keyboard")?.addEventListener("click", (e) => {
    toggleKeyboard();
    if (e.currentTarget instanceof HTMLElement)
      e.currentTarget.blur();
    setTimeout(refocusInput, 100);
  });
  document.getElementById("btn-consulta-toggle-keyboard")?.addEventListener("click", (e) => {
    toggleKeyboard();
    if (e.currentTarget instanceof HTMLElement)
      e.currentTarget.blur();
    setTimeout(refocusInput, 100);
  });
  const toggles = document.querySelectorAll(".password-toggle");
  for (let i = 0;i < toggles.length; i++) {
    toggles[i].addEventListener("click", (e) => {
      const input = e.target.previousElementSibling;
      if (input && input.tagName === "INPUT") {
        if (input.type === "password") {
          input.type = "text";
          e.target.innerText = "\uD83D\uDE48";
        } else {
          input.type = "password";
          e.target.innerText = "\uD83D\uDC41️";
        }
      }
    });
  }
  document.getElementById("btn-back-to-start")?.addEventListener("click", () => {
    confirmExit(() => showScreen("start"));
  });
  document.getElementById("start-empresa")?.addEventListener("change", (e) => {
    loadLocais(e.target.value);
    fetchLastActivityInfo();
  });
  document.getElementById("start-local")?.addEventListener("change", fetchLastActivityInfo);
  document.getElementById("start-rua")?.addEventListener("blur", fetchLastActivityInfo);
  document.getElementById("start-predio")?.addEventListener("blur", fetchLastActivityInfo);
  document.getElementById("form-login")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    showLoader(true);
    const errEl = document.getElementById("login-error");
    if (errEl)
      errEl.classList.add("hidden");
    const { ok, data } = await apiCall("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({
        username: document.getElementById("login-username")?.value,
        password: document.getElementById("login-password")?.value
      })
    });
    showLoader(false);
    if (ok) {
      state.user = data.user;
      state.token = "cookie";
      saveState();
      if (state.atividade) {
        showScreen("scanning");
      } else {
        showScreen("start");
      }
    } else {
      if (errEl) {
        errEl.innerText = data.error || "Erro ao logar";
        errEl.classList.remove("hidden");
      }
    }
  });
  document.getElementById("form-reauth")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    showLoader(true);
    const errEl = document.getElementById("reauth-error");
    if (errEl)
      errEl.classList.add("hidden");
    const { ok, data } = await apiCall("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({
        username: state.user?.username,
        password: document.getElementById("reauth-password")?.value
      })
    });
    showLoader(false);
    if (ok) {
      state.token = "cookie";
      saveState();
      showReauthModal(false);
      alert("Sessão revalidada! Você pode continuar.");
    } else {
      if (errEl) {
        errEl.innerText = data.error || "Senha incorreta";
        errEl.classList.remove("hidden");
      }
    }
  });
  document.getElementById("btn-reauth-cancel")?.addEventListener("click", () => {
    if (confirm("Ao sair agora, todos os produtos lidos nesta atividade serão PERDIDOS. Tem certeza?")) {
      showReauthModal(false);
      state.user = null;
      state.token = null;
      saveState();
      showScreen("login");
    }
  });
  document.getElementById("form-start")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    if (state.atividade && state.scannedProducts.length > 0) {
      if (!confirm(`Você já possui uma atividade em andamento com produtos lidos.

Deseja DESCARTAR os dados e iniciar uma nova?

• Clique OK para descartar e começar nova.
• Clique CANCELAR para voltar à atividade em andamento.`)) {
        showScreen("scanning");
        return;
      }
      resetActivityState();
    }
    await startActivity();
  });
  document.getElementById("form-scan")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    const input = document.getElementById("scan-input");
    const code = input.value.trim();
    if (!code)
      return;
    if (!state.atividade)
      return;
    showLoader(true);
    const { ok, data } = await apiCall(`/api/produtos/ean/${code}?empresa=${state.atividade.empresa}&seqlocal=${state.atividade.seqlocal}`, {}, () => showReauthModal(true));
    showLoader(false);
    const feedback = document.getElementById("scan-feedback");
    if (!feedback)
      return;
    if (!ok) {
      playBeep("error");
      feedback.innerHTML = sanitizeHtml(`<div style="color: #ef4444; font-weight: bold;">❌ Produto não encontrado</div>`);
      input.select();
      return;
    }
    input.value = "";
    focusScannerInput(input);
    const currentPredio = state.atividade.currentPredio || state.atividade.predio;
    const isNullAddress = data.rua == null || data.predio == null;
    const sameRua = data.rua === state.atividade.rua;
    const samePredio = String(data.predio) === String(currentPredio);
    const alreadyScanned = state.scannedProducts.some((p) => p.seqproduto === data.seqproduto);
    if (alreadyScanned) {
      playBeep("warning");
      feedback.innerHTML = sanitizeHtml(`<div style="color: #f59e0b; font-weight: bold;">⚠️ Produto já lido nesta atividade!</div>`);
      return;
    }
    let status = "OK";
    if (isNullAddress || !sameRua || !samePredio) {
      status = "DIVERGENTE";
    }
    state.scannedProducts.push({
      seqproduto: data.seqproduto,
      ean: code,
      rua: data.rua,
      predio: String(currentPredio),
      desccompleta: data.desccompleta,
      status,
      reposicao: false
    });
    saveState();
    state.lastScanned = data;
    if (isNullAddress || !sameRua) {
      playBeep("warning");
      const reason = isNullAddress ? "S/ Endereço" : "Rua Divergente";
      feedback.innerHTML = sanitizeHtml(`<div style="color: #f59e0b; font-weight: bold;">⚠️ ${reason}: ${data.desccompleta}</div>`);
      renderHistory();
    } else if (!samePredio) {
      playBeep("warning");
      const displayPredio = data.predio != null ? data.predio : "N/A";
      const predioSwitchDesc = document.getElementById("predio-switch-desc");
      if (predioSwitchDesc)
        predioSwitchDesc.innerText = `${data.desccompleta} pertence ao Prédio ${displayPredio} (mesma rua).`;
      const predioSwitchNew = document.getElementById("predio-switch-new");
      if (predioSwitchNew)
        predioSwitchNew.innerText = displayPredio.toString();
      const predioSwitchCurrent = document.getElementById("predio-switch-current");
      if (predioSwitchCurrent)
        predioSwitchCurrent.innerText = currentPredio.toString();
      renderHistory();
      showScreen("predio-switch");
    } else {
      playBeep("success");
      feedback.innerHTML = sanitizeHtml(`<div style="color: #10b981; font-weight: bold;">✅ Lido: ${data.desccompleta}</div>`);
      renderHistory();
    }
  });
  document.getElementById("btn-finalize")?.addEventListener("click", finalizeActivity);

  document.getElementById("btn-predio-switch-yes")?.addEventListener("click", async () => {
    if (!state.atividade || !state.lastScanned)
      return;
    try {
      const newPredio = String(state.lastScanned.predio);
      const predios = state.atividade.predios || [
        state.atividade.predio
      ];
      const isNewBuilding = !predios.includes(newPredio);
      if (isNewBuilding) {
        predios.push(newPredio);
        state.atividade.predios = predios;
      }
      state.atividade.currentPredio = newPredio;
      if (isNewBuilding) {
        showLoader(true);
        const result = await apiCall(`/api/produtos/local?empresa=${state.atividade.empresa}&seqlocal=${state.atividade.seqlocal}&rua=${state.atividade.rua}&predio=${newPredio}`, {}, () => showReauthModal(true));
        showLoader(false);
        if (result.ok && Array.isArray(result.data) && result.data.length > 0) {
          const existingSeqs = new Set(state.expectedProducts.map((p) => p.seqproduto));
          const newProducts = result.data.filter((p) => !existingSeqs.has(p.seqproduto));
          state.expectedProducts.push(...newProducts);
        }
      }
      const isExpected = state.expectedProducts.some((p) => p.seqproduto === state.lastScanned.seqproduto);
      const newStatus = isExpected ? "OK" : "DIVERGENTE";
      let lastIdx = -1;
      for (let i = state.scannedProducts.length - 1;i >= 0; i--) {
        if (state.scannedProducts[i].seqproduto === state.lastScanned.seqproduto) {
          lastIdx = i;
          break;
        }
      }
      if (lastIdx >= 0) {
        state.scannedProducts[lastIdx].status = newStatus;
        state.scannedProducts[lastIdx].predio = newPredio;
      }
      saveState();
      showScreen("scanning");
      const feedback = document.getElementById("scan-feedback");
      if (feedback) {
        if (isExpected) {
          playBeep("success");
          feedback.innerHTML = sanitizeHtml(`<div style="color: #10b981; font-weight: bold;">✅ Prédio ${newPredio} agora é o prédio atual. Produto OK.</div>`);
        } else {
          playBeep("warning");
          feedback.innerHTML = sanitizeHtml(`<div style="color: #f59e0b; font-weight: bold;">⚠️ Prédio ${newPredio} agora é o prédio atual, porém produto não esperado!</div>`);
        }
      }
      state.lastScanned = null;
      renderHistory();
    } catch (e) {
      console.error("Error during building switch:", e);
      showScreen("scanning");
      const feedback = document.getElementById("scan-feedback");
      if (feedback)
        feedback.innerHTML = sanitizeHtml('<div style="color: #ef4444; font-weight: bold;">❌ Erro ao trocar de prédio</div>');
    }
  });
  document.getElementById("btn-predio-switch-no")?.addEventListener("click", () => {
    showScreen("scanning");
    const feedback = document.getElementById("scan-feedback");
    if (feedback && state.lastScanned) {
      feedback.innerHTML = sanitizeHtml(`<div style="color: #f59e0b; font-weight: bold;">⚠️ Divergente: ${state.lastScanned.desccompleta}</div>`);
    }
    renderHistory();
  });
  document.getElementById("btn-report-ok")?.addEventListener("click", () => {
    showScreen("start");
  });
  const openConsulta = () => {
    state.lastScanned = null;
    saveState();
    setConsultaMode("codigo");
    showScreen("consulta");
    let lojaNome = "";
    if (state.previousScreen === "scanning" && state.atividade) {
      lojaNome = `• ${state.atividade.empresaNome || `Loja ${state.atividade.empresa}`}`;
    } else {
      const empSelect = document.getElementById("start-empresa");
      if (empSelect?.selectedIndex >= 0) {
        const fullText = empSelect.options[empSelect.selectedIndex].text;
        lojaNome = `• ${fullText.split(" - ")[1] || fullText}`;
      }
    }
    const headerLoja = document.getElementById("consulta-header-loja");
    if (headerLoja)
      headerLoja.innerText = lojaNome;
    const input = document.getElementById("consulta-input");
    if (input) {
      input.value = "";
      input.focus();
    }
    document.getElementById("consulta-result")?.classList.add("hidden");
    document.getElementById("consulta-result-list")?.classList.add("hidden");
    document.getElementById("consulta-result-list").innerHTML = "";
    document.getElementById("consulta-empty")?.classList.remove("hidden");
  };
  document.getElementById("btn-go-consulta")?.addEventListener("click", openConsulta);
  document.getElementById("btn-start-consulta")?.addEventListener("click", openConsulta);
  // ── Consulta search mode toggle ─────────────────────────────────────
  let consultaMode = "codigo";
  function setConsultaMode(mode) {
    consultaMode = mode;
    const input = document.getElementById("consulta-input");
    const btnCodigo = document.getElementById("btn-consulta-mode-codigo");
    const btnDescricao = document.getElementById("btn-consulta-mode-descricao");
    if (mode === "codigo") {
      input.placeholder = "Escanear ou digitar EAN/DUN";
      input.type = "tel";
      input.inputMode = "none";
      btnCodigo.className = "btn btn-sm btn-primary";
      btnCodigo.style.cssText = "flex: 1; font-size: 0.75rem; padding: 0.35rem;";
      btnDescricao.className = "btn btn-sm";
      btnDescricao.style.cssText = "flex: 1; font-size: 0.75rem; padding: 0.35rem; background: #e2e8f0; color: #475569;";
    } else {
      input.placeholder = "Digitar descrição do produto";
      input.type = "text";
      input.inputMode = "text";
      btnDescricao.className = "btn btn-sm btn-primary";
      btnDescricao.style.cssText = "flex: 1; font-size: 0.75rem; padding: 0.35rem;";
      btnCodigo.className = "btn btn-sm";
      btnCodigo.style.cssText = "flex: 1; font-size: 0.75rem; padding: 0.35rem; background: #e2e8f0; color: #475569;";
    }
    document.getElementById("consulta-result")?.classList.add("hidden");
    document.getElementById("consulta-result-list")?.classList.add("hidden");
    document.getElementById("consulta-result-list").innerHTML = "";
    input.value = "";
    input.focus();
  }
  document.getElementById("btn-consulta-mode-codigo")?.addEventListener("click", () => setConsultaMode("codigo"));
  document.getElementById("btn-consulta-mode-descricao")?.addEventListener("click", () => setConsultaMode("descricao"));

  document.getElementById("btn-consulta-back")?.addEventListener("click", () => {
    const backTo = state.previousScreen || (state.atividade ? "scanning" : "start");
    showScreen(backTo === "consulta" ? "start" : backTo);
  });
  document.getElementById("form-consulta")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    const input = document.getElementById("consulta-input");
    const code = input.value.trim();
    if (!code)
      return;
    let empresa = null;
    let seqlocal = null;
    let lojaNome = "";
    if (state.previousScreen === "scanning" && state.atividade) {
      empresa = state.atividade.empresa;
      seqlocal = state.atividade.seqlocal;
      lojaNome = state.atividade.empresaNome || `Loja ${empresa}`;
    } else {
      const empSelect = document.getElementById("start-empresa");
      const locSelect = document.getElementById("start-local");
      empresa = empSelect?.value;
      seqlocal = locSelect?.value;
      if (empSelect?.selectedIndex >= 0) {
        const fullText = empSelect.options[empSelect.selectedIndex].text;
        lojaNome = fullText.split(" - ")[1] || fullText;
      }
    }
    if (!empresa || !seqlocal) {
      alert("Selecione uma empresa e local primeiro");
      return;
    }
    showLoader(true);
    document.getElementById("consulta-result")?.classList.add("hidden");
    document.getElementById("consulta-result-list")?.classList.add("hidden");
    document.getElementById("consulta-result-list").innerHTML = "";
    if (consultaMode === "descricao") {
      const { ok, data } = await apiCall(`/api/produtos/consulta?q=${encodeURIComponent(code)}&empresa=${empresa}&seqlocal=${seqlocal}`, {}, () => showReauthModal(true));
      showLoader(false);
      if (ok && Array.isArray(data) && data.length > 0) {
        playBeep("success");
        document.getElementById("consulta-empty")?.classList.add("hidden");
        const listEl = document.getElementById("consulta-result-list");
        listEl.innerHTML = data.map((p) => {
          const mdv = p.mdv != null ? Number(p.mdv).toFixed(2).replace(".", ",") : "—";
          const preco = p.precoVenda ? `R$ ${Number(p.precoVenda).toFixed(2).replace(".", ",")}` : "N/A";
          const ultEntrada = p.dtaUltEntrada ? formatDate(p.dtaUltEntrada) : "N/A";
          const ultVenda = p.dtaUltVenda ? formatDate(p.dtaUltVenda) : "N/A";
          const codigosHtml = p.codigos ? p.codigos.split("|").map((c) => `<span class="ean-badge">${sanitizeHtml(c)}</span>`).join(" ") : "";
          return `<div class="card" style="margin-bottom: 0.5rem;">
            <div style="display: flex; justify-content: space-between; align-items: flex-start;">
              <div style="flex: 1;">
                <div style="font-weight: 600; color: #4f46e5; font-size: 0.9rem;">${sanitizeHtml(p.desccompleta || "Sem descrição")}</div>
                <div style="font-size: 0.7rem; color: #64748b; margin-top: 2px;">SEQ ${p.seqproduto} · ${sanitizeHtml(p.marca || "—")}</div>
              </div>
              <span style="background: #f1f5f9; color: #475569; padding: 2px 6px; border-radius: 4px; font-size: 0.625rem; font-weight: 700; white-space: nowrap;">${sanitizeHtml(lojaNome)}</span>
            </div>
            <div class="grid-details" style="margin-top: 0.5rem;">
              <div class="detail-item"><span class="detail-label">ESTOQUE</span><span class="detail-value">${p.estoque}</span></div>
              <div class="detail-item"><span class="detail-label">DIAS</span><span class="detail-value">${p.diasEstoque != null ? Number(p.diasEstoque).toFixed(1).replace(".", ",") : "—"}</span></div>
              <div class="detail-item"><span class="detail-label">MDV</span><span class="detail-value">${mdv}</span></div>
              <div class="detail-item"><span class="detail-label">PREÇO</span><span class="detail-value">${preco}</span></div>
            </div>
            <div class="mt-2 text-xs text-slate-500" style="border-top: 1px solid #e2e8f0; padding-top: 0.5rem;">
              <p>Última Entrada: <span style="font-weight: 600;">${ultEntrada}</span></p>
              <p>Última Venda: <span style="font-weight: 600;">${ultVenda}</span></p>
              ${codigosHtml ? `<p style="margin-top: 0.35rem; word-break: break-all;"><strong>CÓDIGOS:</strong> ${codigosHtml}</p>` : ""}
            </div>
          </div>`;
        }).join("");
        listEl.classList.remove("hidden");
      } else {
        playBeep("error");
        alert("Nenhum produto encontrado");
      }
      input.select();
      return;
    }
    const { ok, data } = await apiCall(`/api/produtos/consulta/${code}?empresa=${empresa}&seqlocal=${seqlocal}`, {}, () => showReauthModal(true));
    showLoader(false);
    if (ok) {
      playBeep("success");
      document.getElementById("consulta-empty")?.classList.add("hidden");
      const resultEl = document.getElementById("consulta-result");
      if (resultEl)
        resultEl.classList.remove("hidden");
      const setVal = (id, val) => {
        const el = document.getElementById(id);
        if (el)
          el.innerText = String(val ?? "");
      };
      setVal("consulta-nome", data.desccompleta);
      setVal("consulta-loja-name", lojaNome);
      setVal("consulta-seq", data.seqproduto);
      setVal("consulta-marca", data.marca);
      setVal("consulta-estoque", data.estoque);
      setVal("consulta-dias", data.diasEstoque != null ? Number(data.diasEstoque).toFixed(1).replace(".", ",") : "—");
      setVal("consulta-mdv", data.mdv != null ? Number(data.mdv).toFixed(2).replace(".", ",") : "—");
      setVal("consulta-preco", data.precoVenda ? `R$ ${Number(data.precoVenda).toFixed(2).replace(".", ",")}` : "N/A");
      setVal("consulta-entrada", data.dtaUltEntrada ? formatDate(data.dtaUltEntrada) : "N/A");
      setVal("consulta-venda", data.dtaUltVenda ? formatDate(data.dtaUltVenda) : "N/A");
      const codigosEl = document.getElementById("consulta-codigos");
      if (codigosEl && data.codigos) {
        codigosEl.innerHTML = sanitizeHtml(data.codigos.split("|").map((c) => `<span class="ean-badge">${c}</span>`).join(" "));
      }
      input.select();
    } else {
      playBeep("error");
      alert("Produto não encontrado");
      input.select();
    }
  });
});
function bulkPrint(){let ids=[...document.querySelectorAll('.bulk:checked')].map(x=>x.value).join(',');if(ids) location.href='/dashboard/activities/print-view/bulk?ids='+ids}
