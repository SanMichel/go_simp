// src/client/shared/auth-routing.ts
function routeForRole(role) {
  if (role === "conferente")
    return "/atividades";
  if (role === "gerente")
    return "/dashboard";
  return "/admin";
}
function isLoginResponse(data) {
  if (!data || typeof data !== "object")
    return false;
  const candidate = data;
  if (typeof candidate.token !== "string")
    return false;
  if (!candidate.user || typeof candidate.user !== "object")
    return false;
  const user = candidate.user;
  return user.role === "sysadmin" || user.role === "gerente" || user.role === "conferente";
}
function isMeResponse(data) {
  if (!data || typeof data !== "object")
    return false;
  const candidate = data;
  if (!candidate.user || typeof candidate.user !== "object")
    return false;
  const user = candidate.user;
  return user.role === "sysadmin" || user.role === "gerente" || user.role === "conferente";
}
// src/client/login/index.ts
function showError(message) {
  const errorEl = document.getElementById("login-error");
  if (!errorEl)
    return;
  errorEl.textContent = message;
  errorEl.classList.remove("hidden");
}
function hideError() {
  document.getElementById("login-error")?.classList.add("hidden");
}
function redirectForRole(role) {
  window.location.href = routeForRole(role);
}
document.addEventListener("DOMContentLoaded", async () => {
  const sessionRes = await apiCall("/api/auth/me", {}, () => {});
  if (sessionRes.ok && isMeResponse(sessionRes.data)) {
    redirectForRole(sessionRes.data.user.role);
    return;
  }
  document.getElementById("form-login-page")?.addEventListener("submit", async (event) => {
    event.preventDefault();
    hideError();
    showLoader(true);
    const username = document.getElementById("login-username").value;
    const password = document.getElementById("login-password").value;
    const res = await apiCall("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password })
    });
    showLoader(false);
    if (res.ok && isLoginResponse(res.data)) {
      localStorage.setItem("simp_user", JSON.stringify(res.data.user));
      if (res.data.user.role === "conferente") {
        localStorage.setItem("simp_screen", "start");
      }
      redirectForRole(res.data.user.role);
      return;
    }
    const error = res.data && typeof res.data === "object" && "error" in res.data ? String(res.data.error) : "Erro ao logar";
    showError(error);
  });
});
