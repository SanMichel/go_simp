package main

import (
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
	}
	if err := json.Unmarshal(payloadBytes, &p); err != nil {
		return nil, err
	}
	if time.Now().Unix() > p.Exp {
		return nil, errors.New("expired")
	}
	return a.findUserByID(r.Context(), p.ID)
}

func (a *App) makeToken(userID int) (string, error) {
	payload, err := json.Marshal(struct {
		ID    int    `json:"id"`
		Exp   int64  `json:"exp"`
		Nonce string `json:"nonce"`
	}{ID: userID, Exp: time.Now().Add(8 * time.Hour).Unix(), Nonce: randomString(12)})
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
