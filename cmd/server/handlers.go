package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	a.render(w, "home", nil)
}

func (a *App) loginPage(w http.ResponseWriter, r *http.Request) {
	if u, err := a.currentUser(r); err == nil {
		a.redirectByRole(w, r, u.Role)
		return
	}
	a.render(w, "login", nil)
}

func (a *App) loginPost(w http.ResponseWriter, r *http.Request) {
	if !a.loginLimiter.allow(r.RemoteAddr) {
		a.render(w, "login", map[string]string{"Error": "Muitas tentativas. Aguarde 1 minuto."})
		return
	}
	if err := r.ParseForm(); err != nil {
		a.render(w, "login", map[string]string{"Error": "Erro ao processar formulário."})
		return
	}
	u, err := a.findUserByUsername(r.Context(), r.FormValue("username"))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(r.FormValue("password"))) != nil {
		a.render(w, "login", map[string]string{"Error": "Usuário ou senha incorretos."})
		return
	}
	token, err := a.makeToken(u.ID)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeInternal, Message: "Erro ao criar sessão", HTTPStatus: http.StatusInternalServerError})
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "token", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: int(a.cfg.SessionTTL.Seconds()), Secure: true})
	a.setCSRFCookie(w)
	a.redirectByRole(w, r, u.Role)
}

func (a *App) healthCheck(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	if a.ora != nil && a.ora.db != nil {
		if err := a.ora.db.PingContext(r.Context()); err != nil {
			status = "oracle_down"
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": status})
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	if u, err := a.currentUser(r); err == nil {
		a.revokeSession(r.Context(), u.ID)
	}
	http.SetCookie(w, &http.Cookie{Name: "token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode})
	a.clearCSRFCookie(w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (a *App) atividadesPage(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(ctxUser).(*User)
	a.render(w, "atividades", map[string]any{"User": u})
}

func (a *App) atividadesLoginPage(w http.ResponseWriter, r *http.Request) {
	u, err := a.currentUser(r)
	if err == nil {
		a.redirectByRole(w, r, u.Role)
		return
	}
	a.render(w, "atividades-login", nil)
}

func (a *App) apiMe(w http.ResponseWriter, r *http.Request) {
	u, err := a.currentUser(r)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeUnauthorized, Message: "Não autenticado", HTTPStatus: http.StatusUnauthorized})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": map[string]any{"id": u.ID, "username": u.Username, "role": u.Role}})
}

func (a *App) apiLogin(w http.ResponseWriter, r *http.Request) {
	if !a.loginLimiter.allow(r.RemoteAddr) {
		a.handleError(w, r, &AppError{Code: ErrCodeRateLimited, Message: "Muitas tentativas. Aguarde 1 minuto.", HTTPStatus: http.StatusTooManyRequests})
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeBadRequest, Message: "JSON inválido", HTTPStatus: http.StatusBadRequest})
		return
	}
	u, err := a.findUserByUsername(r.Context(), body.Username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.Password)) != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeUnauthorized, Message: "Credenciais inválidas", HTTPStatus: http.StatusUnauthorized})
		return
	}
	token, err := a.makeToken(u.ID)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeInternal, Message: "Erro ao criar sessão", HTTPStatus: http.StatusInternalServerError, Err: err})
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "token", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: int(a.cfg.SessionTTL.Seconds()), Secure: true})
	a.setCSRFCookie(w)
	writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{"id": u.ID, "username": u.Username, "role": u.Role},
	})
}

func (a *App) apiLogout(w http.ResponseWriter, r *http.Request) {
	if u, err := a.currentUser(r); err == nil {
		a.revokeSession(r.Context(), u.ID)
	}
	http.SetCookie(w, &http.Cookie{Name: "token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode})
	a.clearCSRFCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Logout realizado com sucesso"})
}
