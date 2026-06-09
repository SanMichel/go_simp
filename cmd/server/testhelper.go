package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("TEST_POSTGRES_URL")
	if url == "" {
		t.Skip("set TEST_POSTGRES_URL to run database-dependent tests")
	}
	pg, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("test DB open: %v", err)
	}
	if err := pg.PingContext(context.Background()); err != nil {
		t.Fatalf("test DB ping: %v", err)
	}
	t.Cleanup(func() { pg.Close() })
	return pg
}

func testApp(t *testing.T) *App {
	t.Helper()
	pg := testDB(t)
	app := &App{
		cfg: Config{
			SessionSecret: []byte("test-secret-32-chars-minimum-length!"),
			SessionTTL:    8 * time.Hour,
			AppEnv:        "test",
		},
		pg:           pg,
		tpl:          parseTemplates(),
		loginLimiter: newRateLimiter(),
	}
	ctx := context.Background()
	if err := app.migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := app.seedAdmin(ctx); err != nil {
		t.Fatalf("seedAdmin: %v", err)
	}
	t.Cleanup(func() { cleanupTestData(t, pg) })
	return app
}

func testToken(app *App, userID int) string {
	now := time.Now()
	payload, _ := json.Marshal(struct {
		ID  int   `json:"id"`
		Exp int64 `json:"exp"`
		Iat int64 `json:"iat"`
	}{ID: userID, Exp: now.Add(app.cfg.SessionTTL).Unix(), Iat: now.Unix()})
	body := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, app.cfg.SessionSecret)
	mac.Write([]byte(body))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return body + "." + sig
}

func testUser(t *testing.T, app *App, username, role, password string) *User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	var u User
	err = app.pg.QueryRowContext(context.Background(),
		`INSERT INTO users (username, password, role) VALUES ($1,$2,$3)
		 RETURNING id, username, password, role, last_token_at`,
		username, string(hash), role,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.LastTokenAt)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	return &u
}

func cleanupTestData(t *testing.T, pg *sql.DB) {
	ctx := context.Background()
	_, _ = pg.ExecContext(ctx, `DELETE FROM produto_verificacao`)
	_, _ = pg.ExecContext(ctx, `DELETE FROM atividade_enderecos`)
	_, _ = pg.ExecContext(ctx, `DELETE FROM atividades`)
	_, _ = pg.ExecContext(ctx, `DELETE FROM users WHERE username LIKE 'test_%'`)
}
