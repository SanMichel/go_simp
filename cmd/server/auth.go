package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (a *App) currentUser(r *http.Request) (*User, error) {
	c, err := r.Cookie("token")
	if err != nil {
		return nil, err
	}
	parts := strings.Split(c.Value, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid token")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, a.cfg.SessionSecret)
	mac.Write([]byte(parts[0]))
	want := mac.Sum(nil)
	got, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(got, want) {
		return nil, errors.New("bad signature")
	}
	var p struct {
		ID  int   `json:"id"`
		Exp int64 `json:"exp"`
		Iat int64 `json:"iat"`
	}
	if err := json.Unmarshal(payloadBytes, &p); err != nil {
		return nil, err
	}
	if time.Now().Unix() > p.Exp {
		return nil, errors.New("expired")
	}
	u, err := a.findUserByID(r.Context(), p.ID)
	if err != nil {
		return nil, err
	}
	if u.LastTokenAt.Valid && p.Iat < u.LastTokenAt.Time.Unix() {
		return nil, errors.New("session revoked")
	}
	return u, nil
}

func (a *App) makeToken(userID int) (string, error) {
	now := time.Now()
	payload, err := json.Marshal(struct {
		ID  int   `json:"id"`
		Exp int64 `json:"exp"`
		Iat int64 `json:"iat"`
	}{ID: userID, Exp: now.Add(a.cfg.SessionTTL).Unix(), Iat: now.Unix()})
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, a.cfg.SessionSecret)
	mac.Write([]byte(body))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return body + "." + sig, nil
}

func randomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
func (a *App) requireRole(roles string, next http.HandlerFunc) http.HandlerFunc {
	allowed := map[string]bool{}
	for _, r := range strings.Split(roles, ",") {
		if r != "" {
			allowed[r] = true
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := a.currentUser(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if len(allowed) > 0 && !allowed[u.Role] {
			if u.Role == "conferente" {
				http.Redirect(w, r, "/atividades", http.StatusFound)
			} else {
				http.Redirect(w, r, "/dashboard", http.StatusFound)
			}
			return
		}
		next(w, r)
	}
}

func (a *App) requireAPI(next func(http.ResponseWriter, *http.Request, *User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := a.currentUser(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Não autorizado"})
			return
		}
		next(w, r, u)
	}
}
func (a *App) redirectByRole(w http.ResponseWriter, r *http.Request, role string) {
	dest := "/atividades"
	if role == "sysadmin" {
		dest = "/admin"
	} else if role == "gerente" {
		dest = "/dashboard"
	}
	http.Redirect(w, r, dest, http.StatusFound)
}

func (a *App) setCSRFCookie(w http.ResponseWriter) {
	token := randomString(32)
	http.SetCookie(w, &http.Cookie{Name: "csrf_token", Value: token, Path: "/", SameSite: http.SameSiteLaxMode, MaxAge: int(a.cfg.SessionTTL.Seconds()), Secure: true})
}

func (a *App) clearCSRFCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "csrf_token", Value: "", Path: "/", MaxAge: -1, SameSite: http.SameSiteLaxMode, Secure: true})
}

func (a *App) revokeSession(ctx context.Context, userID int) {
	_, _ = a.pg.ExecContext(ctx, `UPDATE users SET last_token_at=now() WHERE id=$1`, userID)
}

func (a *App) csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PATCH" || r.Method == "DELETE" {
			skipCSRF := r.URL.Path == "/login" || r.URL.Path == "/api/auth/login"
			if !skipCSRF {
				origin := r.Header.Get("Origin")
				if origin != "" {
					if !strings.Contains(origin, "://"+r.Host) {
						http.Error(w, "403 Forbidden", http.StatusForbidden)
						return
					}
				}
				if strings.HasPrefix(r.URL.Path, "/api/") {
					cookieCSRF, err := r.Cookie("csrf_token")
					if err != nil || cookieCSRF.Value == "" {
						writeJSON(w, http.StatusForbidden, map[string]string{"error": "CSRF token ausente"})
						return
					}
					headerCSRF := r.Header.Get("X-CSRF-Token")
					if headerCSRF == "" || !hmac.Equal([]byte(cookieCSRF.Value), []byte(headerCSRF)) {
						writeJSON(w, http.StatusForbidden, map[string]string{"error": "CSRF token inválido"})
						return
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
