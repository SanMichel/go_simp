document.addEventListener("DOMContentLoaded", function() {
  apiGet("/api/auth/me", function(ok, status, data) {
    if (ok && data && data.user && data.user.role) {
      window.location.href = "/atividades";
      return;
    }
  });

  document.getElementById("form-atividades-login").addEventListener("submit", function(e) {
    e.preventDefault();
    showLoader(true);
    var errorEl = document.getElementById("login-error");
    if (errorEl) errorEl.classList.add("hidden");
    var username = document.getElementById("login-username").value;
    var password = document.getElementById("login-password").value;
    apiPost("/api/auth/login", { username: username, password: password }, function(ok, status, data) {
      showLoader(false);
      if (ok && data && data.user) {
        window.location.href = "/atividades";
      } else {
        if (errorEl) {
          errorEl.innerText = data && data.error ? data.error : "Erro ao logar";
          errorEl.classList.remove("hidden");
        }
      }
    }, function() {
      showLoader(false);
      if (errorEl) {
        errorEl.innerText = "Sess\u00e3o expirada";
        errorEl.classList.remove("hidden");
      }
    });
  });
});
