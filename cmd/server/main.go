package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/sijms/go-ora/v2"
	go_ora "github.com/sijms/go-ora/v2"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	Port          string
	AppEnv        string
	SessionSecret []byte
	PostgresURL   string
	OracleURL     string
}

type App struct {
	cfg Config
	pg  *sql.DB
	ora *OracleReader
	tpl *template.Template
}

type OracleReader struct {
	db *sql.DB
}

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Role         string
}

type Activity struct {
	ID       int
	Empresa  string
	SeqLocal int
	UserID   int
	Username string
	DataFim  time.Time
	Impresso bool
	Rua      string
	Predio   string
	Predios  []string
}

type ActivityFilters struct {
	Username     []string
	Empresa      []string
	Rua          []string
	Predio       []string
	Impresso     []string
	ID           []string
	DataFimStart string
	DataFimEnd   string
	Sort         string
	Order        string
}

type FilterOptions struct {
	Username []string
	Empresa  []string
	Rua      []string
	Predio   []string
	Impresso []string
	ID       []string
}

type ProductVerification struct {
	ID             int
	AtividadeID    int
	SeqProduto     int
	Empresa        string
	RuaLida        sql.NullString
	PredioLido     sql.NullString
	RuaEsperada    sql.NullString
	PredioEsperado sql.NullString
	Status         string
	Reposicao      bool
	Estoque        int
	DataEntrada    sql.NullTime
	DescCompleta   sql.NullString
}

type OracleEmpresa struct {
	NROEMPRESA   int    `json:"NROEMPRESA"`
	NOMEREDUZIDO string `json:"NOMEREDUZIDO"`
}

type OracleLocal struct {
	SEQLOCAL   int    `json:"SEQLOCAL"`
	NROEMPRESA int    `json:"NROEMPRESA"`
	LOCAL      string `json:"LOCAL"`
}

type OracleProduct struct {
	SEQPRODUTO    int             `json:"SEQPRODUTO"`
	CODACESSO     sql.NullString  `json:"CODACESSO,omitempty"`
	NRORUA        sql.NullString  `json:"NRORUA,omitempty"`
	NROPREDIO     sql.NullString  `json:"NROPREDIO,omitempty"`
	DESCCOMPLETA  sql.NullString  `json:"DESCCOMPLETA,omitempty"`
	NROEMPRESA    int             `json:"NROEMPRESA,omitempty"`
	DTAULTENTRADA sql.NullTime    `json:"DTAULTENTRADA,omitempty"`
	DTAULTVENDA   sql.NullTime    `json:"DTAULTVENDA,omitempty"`
	ESTQLOJA      int             `json:"ESTQLOJA,omitempty"`
	MEDVDIAGERAL  sql.NullFloat64 `json:"MEDVDIAGERAL,omitempty"`
	MARCA         sql.NullString  `json:"MARCA,omitempty"`
	PRECO_VENDA   sql.NullFloat64 `json:"PRECO_VENDA,omitempty"`
	CODIGOS       sql.NullString  `json:"CODIGOS,omitempty"`
}

func main() {
	if err := loadDotEnv(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal("load .env: ", err)
	}

	cfg := loadConfig()

	pg, err := sql.Open("pgx", cfg.PostgresURL)
	if err != nil {
		log.Fatal(err)
	}
	pg.SetMaxOpenConns(10)
	pg.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := pg.PingContext(ctx); err != nil {
		log.Fatal("postgres ping: ", err)
	}

	ora, err := sql.Open("oracle", cfg.OracleURL)
	if err != nil {
		log.Fatal(err)
	}
	oracleReader := &OracleReader{db: ora}
	oracleReader.db.SetMaxOpenConns(5)
	oracleReader.db.SetMaxIdleConns(1)
	oracleReader.db.SetConnMaxLifetime(time.Hour)
	if err := oracleReader.db.PingContext(ctx); err != nil {
		log.Printf("warning: oracle ping failed: %v", err)
	}

	app := &App{cfg: cfg, pg: pg, ora: oracleReader, tpl: parseTemplates()}
	if err := app.migrate(ctx); err != nil {
		log.Fatal(err)
	}
	if err := app.seedAdmin(ctx); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	app.routes(mux)

	addr := ":" + cfg.Port
	log.Printf("server ready on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, app.log(mux)))
}

func loadConfig() Config {
	port := getenv("PORT", "3000")
	secret := getenv("SESSION_SECRET", "dev-secret-change-me")
	pgURL := getenv("POSTGRES_URL", os.Getenv("DATABASE_URL"))
	if pgURL == "" {
		log.Fatal("POSTGRES_URL or DATABASE_URL required")
	}

	oracleURL := os.Getenv("ORACLE_URL")
	if oracleURL == "" {
		portNum, _ := strconv.Atoi(getenv("ORACLE_PORT", "1521"))
		oracleURL = go_ora.BuildUrl(
			getenv("ORACLE_HOST", "localhost"),
			portNum,
			getenv("ORACLE_SERVICE", "xe"),
			os.Getenv("ORACLE_USER"),
			os.Getenv("ORACLE_PASSWORD"),
			map[string]string{"TIMEOUT": "30", "client charset": "UTF8"},
		)
	}

	return Config{
		Port:          port,
		AppEnv:        getenv("APP_ENV", "development"),
		SessionSecret: []byte(secret),
		PostgresURL:   pgURL,
		OracleURL:     oracleURL,
	}
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for lineNo, raw := range strings.Split(string(b), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("%s:%d: expected KEY=VALUE", path, lineNo+1)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return fmt.Errorf("%s:%d: empty key", path, lineNo+1)
		}
		if len(value) >= 2 {
			quote := value[0]
			if (quote == '"' || quote == '\'') && value[len(value)-1] == quote {
				value = value[1 : len(value)-1]
			}
		}
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *App) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /style.css", a.style)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/login", http.StatusFound) })
	mux.HandleFunc("GET /index.html", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/home", http.StatusFound) })
	mux.HandleFunc("GET /home", a.home)
	mux.HandleFunc("GET /login", a.loginPage)
	mux.HandleFunc("POST /login", a.loginPost)
	mux.HandleFunc("POST /logout", a.logout)
	mux.HandleFunc("GET /atividades", a.requireRole("", a.atividadesPage))
	mux.HandleFunc("GET /dashboard", a.requireRole("gerente,sysadmin", a.dashboardPage))
	mux.HandleFunc("GET /dashboard/table", a.requireRole("gerente,sysadmin", a.dashboardTable))
	mux.HandleFunc("GET /dashboard/activities/{id}/details", a.requireRole("gerente,sysadmin", a.activityDetails))
	mux.HandleFunc("GET /dashboard/activities/{id}/print-view", a.requireRole("gerente,sysadmin", a.printOne))
	mux.HandleFunc("GET /dashboard/activities/print-view/bulk", a.requireRole("gerente,sysadmin", a.printBulk))
	mux.HandleFunc("GET /admin", a.requireRole("sysadmin", a.adminPage))
	mux.HandleFunc("GET /admin/users/section", a.requireRole("sysadmin", a.adminUsersSection))
	mux.HandleFunc("POST /admin/users", a.requireRole("sysadmin", a.adminCreateUser))
	mux.HandleFunc("GET /admin/users/{id}/edit", a.requireRole("sysadmin", a.adminEditUserRow))
	mux.HandleFunc("GET /admin/users/{id}/row", a.requireRole("sysadmin", a.adminUserRow))
	mux.HandleFunc("POST /admin/users/{id}", a.requireRole("sysadmin", a.adminUpdateUser))
	mux.HandleFunc("GET /api/auth/me", a.apiMe)
	mux.HandleFunc("POST /api/auth/login", a.apiLogin)
	mux.HandleFunc("POST /api/auth/logout", a.apiLogout)
	mux.HandleFunc("GET /api/empresas", a.requireAPI(a.apiEmpresas))
	mux.HandleFunc("GET /api/locais", a.requireAPI(a.apiLocais))
	mux.HandleFunc("GET /api/produtos/ean/{codigo}", a.requireAPI(a.apiProdutoEAN))
	mux.HandleFunc("GET /api/produtos/consulta/{codigo}", a.requireAPI(a.apiProdutoConsulta))
	mux.HandleFunc("GET /api/produtos/local", a.requireAPI(a.apiProdutosLocal))
	mux.HandleFunc("POST /api/atividades/finalizar", a.requireAPI(a.apiFinalizar))
	mux.HandleFunc("GET /api/atividades/last-info", a.requireAPI(a.apiLastInfo))
}

func (a *App) log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &logWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, lw.status, time.Since(start).Truncate(time.Millisecond))
	})
}

type logWriter struct {
	http.ResponseWriter
	status int
}

func (w *logWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (o *OracleReader) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if !isReadOnlySQL(query) {
		return nil, errors.New("oracle connection is read-only; only SELECT/WITH queries are allowed")
	}
	return o.db.QueryContext(ctx, query, args...)
}

func (o *OracleReader) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if !isReadOnlySQL(query) {
		return o.db.QueryRowContext(ctx, "SELECT 1 FROM dual WHERE 1=0")
	}
	return o.db.QueryRowContext(ctx, query, args...)
}

func isReadOnlySQL(query string) bool {
	q := strings.TrimSpace(strings.TrimPrefix(query, "\ufeff"))
	q = strings.TrimLeft(q, "(\n\r\t ")
	fields := strings.Fields(q)
	if len(fields) == 0 {
		return false
	}
	first := strings.ToUpper(fields[0])
	return first == "SELECT" || first == "WITH"
}

func (a *App) migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'conferente'
		)`,
		`CREATE TABLE IF NOT EXISTS atividades (
			id BIGSERIAL PRIMARY KEY,
			empresa TEXT NOT NULL,
			seqlocal INTEGER,
			usuario_id BIGINT REFERENCES users(id),
			data_fim TIMESTAMPTZ NOT NULL DEFAULT now(),
			impresso BOOLEAN NOT NULL DEFAULT false
		)`,
		`CREATE TABLE IF NOT EXISTS atividade_enderecos (
			id BIGSERIAL PRIMARY KEY,
			atividade_id BIGINT NOT NULL REFERENCES atividades(id) ON DELETE CASCADE,
			rua TEXT NOT NULL,
			predio TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS produto_verificacao (
			id BIGSERIAL PRIMARY KEY,
			atividade_id BIGINT REFERENCES atividades(id) ON DELETE CASCADE,
			seqproduto INTEGER NOT NULL,
			empresa TEXT NOT NULL,
			rua_lida TEXT,
			predio_lido TEXT,
			rua_esperada TEXT,
			predio_esperado TEXT,
			status TEXT NOT NULL,
			reposicao BOOLEAN NOT NULL DEFAULT false,
			estoque INTEGER NOT NULL DEFAULT 0,
			data_entrada TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_atividades_usuario ON atividades(usuario_id)`,
		`CREATE INDEX IF NOT EXISTS idx_atividades_data_fim ON atividades(data_fim DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_produto_verificacao_atividade ON produto_verificacao(atividade_id)`,
	}
	for _, stmt := range stmts {
		if _, err := a.pg.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) seedAdmin(ctx context.Context) error {
	var id int
	err := a.pg.QueryRowContext(ctx, `SELECT id FROM users WHERE username='admin'`).Scan(&id)
	if err == nil {
		_, err = a.pg.ExecContext(ctx, `UPDATE users SET role='sysadmin' WHERE username='admin' AND role <> 'sysadmin'`)
		return err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = a.pg.ExecContext(ctx, `INSERT INTO users (username, password, role) VALUES ($1, $2, 'sysadmin')`, "admin", string(hash))
	return err
}

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
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "token", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: 8 * 60 * 60, Secure: a.cfg.AppEnv == "production"})
	a.redirectByRole(w, r, u.Role)
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

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (a *App) atividadesPage(w http.ResponseWriter, r *http.Request) {
	u, _ := a.currentUser(r)
	a.render(w, "atividades", map[string]any{"User": u})
}

func (a *App) dashboardPage(w http.ResponseWriter, r *http.Request) {
	u, _ := a.currentUser(r)
	activities, _ := a.listActivities(r.Context(), parseFilters(r), 50)
	options, _ := a.listFilterOptions(r.Context())
	a.render(w, "dashboard", map[string]any{"User": u, "Activities": activities, "Options": options, "Filters": parseFilters(r)})
}

func (a *App) dashboardTable(w http.ResponseWriter, r *http.Request) {
	activities, err := a.listActivities(r.Context(), parseFilters(r), 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	options, _ := a.listFilterOptions(r.Context())
	a.render(w, "activities_table", map[string]any{"Activities": activities, "Options": options, "Filters": parseFilters(r)})
}

func (a *App) activityDetails(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	act, items, err := a.activityDetailsData(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	a.render(w, "activity_modal", map[string]any{"Activity": act, "Items": items})
}

func (a *App) printOne(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	a.printActivities(w, r, []int{id})
}

func (a *App) printBulk(w http.ResponseWriter, r *http.Request) {
	var ids []int
	for _, part := range strings.Split(r.URL.Query().Get("ids"), ",") {
		if id, err := strconv.Atoi(strings.TrimSpace(part)); err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		http.Error(w, "IDs inválidos", http.StatusBadRequest)
		return
	}
	a.printActivities(w, r, ids)
}

func (a *App) printActivities(w http.ResponseWriter, r *http.Request, ids []int) {
	type Bundle struct {
		Activity Activity
		Items    []ProductVerification
	}
	var bundles []Bundle
	for _, id := range ids {
		act, items, err := a.activityDetailsData(r.Context(), id)
		if err == nil {
			bundles = append(bundles, Bundle{Activity: act, Items: items})
			_, _ = a.pg.ExecContext(r.Context(), `UPDATE atividades SET impresso=true WHERE id=$1`, id)
		}
	}
	a.render(w, "print", map[string]any{"Bundles": bundles})
}

func (a *App) adminPage(w http.ResponseWriter, r *http.Request) {
	u, _ := a.currentUser(r)
	users, err := a.listUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.render(w, "admin", map[string]any{"User": u, "Users": users})
}

func (a *App) adminUsersSection(w http.ResponseWriter, r *http.Request) {
	users, _ := a.listUsers(r.Context())
	a.render(w, "users_section", map[string]any{"Users": users})
}

func (a *App) adminCreateUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	role := r.FormValue("role")
	if username == "" || len(password) < 4 || !validRole(role) {
		users, _ := a.listUsers(r.Context())
		a.render(w, "users_section", map[string]any{"Users": users, "Message": "Dados inválidos.", "Error": true})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	_, err := a.pg.ExecContext(r.Context(), `INSERT INTO users (username, password, role) VALUES ($1,$2,$3)`, username, string(hash), role)
	users, _ := a.listUsers(r.Context())
	if err != nil {
		a.render(w, "users_section", map[string]any{"Users": users, "Message": err.Error(), "Error": true})
		return
	}
	a.render(w, "users_section", map[string]any{"Users": users, "Message": "Usuário criado com sucesso."})
}

func (a *App) adminEditUserRow(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	u, err := a.findUserByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	a.render(w, "user_edit_row", map[string]any{"RowUser": u})
}

func (a *App) adminUserRow(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	u, err := a.findUserByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	a.render(w, "user_row", map[string]any{"RowUser": u})
}

func (a *App) adminUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	role := r.FormValue("role")
	password := r.FormValue("password")
	if !validRole(role) {
		http.Error(w, "role inválido", http.StatusBadRequest)
		return
	}
	if password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		_, _ = a.pg.ExecContext(r.Context(), `UPDATE users SET role=$1, password=$2 WHERE id=$3`, role, string(hash), id)
	} else {
		_, _ = a.pg.ExecContext(r.Context(), `UPDATE users SET role=$1 WHERE id=$2`, role, id)
	}
	u, _ := a.findUserByID(r.Context(), id)
	a.render(w, "user_row", map[string]any{"RowUser": u})
}

func validRole(role string) bool {
	return role == "sysadmin" || role == "gerente" || role == "conferente"
}

func (a *App) apiMe(w http.ResponseWriter, r *http.Request) {
	u, err := a.currentUser(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Não autenticado"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": map[string]any{"id": u.ID, "username": u.Username, "role": u.Role}})
}

func (a *App) apiLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return
	}
	u, err := a.findUserByUsername(r.Context(), body.Username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.Password)) != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Credenciais inválidas"})
		return
	}
	token, err := a.makeToken(u.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erro ao criar sessão"})
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "token", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: 8 * 60 * 60, Secure: a.cfg.AppEnv == "production"})
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  map[string]any{"id": u.ID, "username": u.Username, "role": u.Role},
	})
}

func (a *App) apiLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Logout realizado com sucesso"})
}

func (a *App) apiEmpresas(w http.ResponseWriter, r *http.Request, u *User) {
	rows, err := a.ora.QueryContext(r.Context(), `SELECT me.NROEMPRESA, me.NOMEREDUZIDO FROM CONSINCO.MAX_EMPRESA me WHERE me.STATUS = 'A' ORDER BY me.NROEMPRESA`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar empresas"})
		return
	}
	defer rows.Close()
	var out []OracleEmpresa
	for rows.Next() {
		var e OracleEmpresa
		if rows.Scan(&e.NROEMPRESA, &e.NOMEREDUZIDO) == nil {
			out = append(out, e)
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) apiLocais(w http.ResponseWriter, r *http.Request, u *User) {
	empresa, _ := strconv.Atoi(r.URL.Query().Get("empresa"))
	rows, err := a.ora.QueryContext(r.Context(), `SELECT ml.SEQLOCAL, ml.NROEMPRESA, ml.LOCAL FROM CONSINCO.MRL_LOCAL ml WHERE ml.STATUS = 'A' AND ml.NROEMPRESA = :1 ORDER BY ml.SEQLOCAL`, empresa)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar locais"})
		return
	}
	defer rows.Close()
	var out []OracleLocal
	for rows.Next() {
		var l OracleLocal
		if rows.Scan(&l.SEQLOCAL, &l.NROEMPRESA, &l.LOCAL) == nil {
			out = append(out, l)
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) apiProdutoEAN(w http.ResponseWriter, r *http.Request, u *User) {
	codigo := r.PathValue("codigo")
	empresa, _ := strconv.Atoi(r.URL.Query().Get("empresa"))
	seqlocal, _ := strconv.Atoi(r.URL.Query().Get("seqlocal"))
	p, err := a.findAddressByCode(r.Context(), codigo, empresa, seqlocal)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Produto não encontrado"})
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (a *App) apiProdutoConsulta(w http.ResponseWriter, r *http.Request, u *User) {
	codigo := r.PathValue("codigo")
	empresa, _ := strconv.Atoi(r.URL.Query().Get("empresa"))
	seqlocal, _ := strconv.Atoi(r.URL.Query().Get("seqlocal"))
	p, err := a.findFullProductByCode(r.Context(), codigo, empresa, seqlocal)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Produto não encontrado"})
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (a *App) apiProdutosLocal(w http.ResponseWriter, r *http.Request, u *User) {
	empresa, _ := strconv.Atoi(r.URL.Query().Get("empresa"))
	seqlocal, _ := strconv.Atoi(r.URL.Query().Get("seqlocal"))
	rua := r.URL.Query().Get("rua")
	predio := r.URL.Query().Get("predio")
	rows, err := a.ora.QueryContext(r.Context(), `SELECT mrlp.SEQPRODUTO, mrlp.NRORUA, mrlp.NROPREDIO, mrlp.ESTOQUE FROM CONSINCO.MRL_PRODLOCAL mrlp WHERE mrlp.SEQLOCAL = :1 AND mrlp.ESTOQUE > 0 AND mrlp.NRORUA = :2 AND mrlp.NROPREDIO = :3 AND mrlp.NROEMPRESA = :4`, seqlocal, rua, predio, empresa)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar produtos"})
		return
	}
	defer rows.Close()
	type row struct {
		SEQPRODUTO int    `json:"SEQPRODUTO"`
		NRORUA     string `json:"NRORUA"`
		NROPREDIO  string `json:"NROPREDIO"`
		ESTOQUE    int    `json:"ESTOQUE"`
	}
	var out []row
	for rows.Next() {
		var x row
		if rows.Scan(&x.SEQPRODUTO, &x.NRORUA, &x.NROPREDIO, &x.ESTOQUE) == nil {
			out = append(out, x)
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) apiLastInfo(w http.ResponseWriter, r *http.Request, u *User) {
	var dataFim sql.NullTime
	err := a.pg.QueryRowContext(r.Context(), `
		SELECT a.data_fim
		FROM atividades a
		JOIN atividade_enderecos e ON e.atividade_id=a.id
		WHERE a.empresa=$1 AND a.seqlocal=$2 AND e.rua=$3 AND e.predio=$4
		ORDER BY a.data_fim DESC LIMIT 1`,
		r.URL.Query().Get("empresa"), intQuery(r, "seqlocal"), r.URL.Query().Get("rua"), r.URL.Query().Get("predio"),
	).Scan(&dataFim)
	if err != nil {
		writeJSON(w, http.StatusOK, nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"dataFim": dataFim.Time})
}

type finalizeReq struct {
	Empresa      any      `json:"empresa"`
	SeqLocal     int      `json:"seqlocal"`
	Rua          string   `json:"rua"`
	Predio       []string `json:"predio"`
	ReadProducts []struct {
		SeqProduto int    `json:"seqproduto"`
		EAN        string `json:"ean"`
		Rua        string `json:"rua"`
		Predio     string `json:"predio"`
		Status     string `json:"status"`
		Reposicao  bool   `json:"reposicao"`
	} `json:"readProducts"`
	ExpectedProducts []struct {
		SeqProduto int `json:"seqproduto"`
	} `json:"expectedProducts"`
}

func (a *App) apiFinalizar(w http.ResponseWriter, r *http.Request, u *User) {
	var req finalizeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return
	}
	if len(req.Predio) == 0 {
		req.Predio = []string{""}
	}
	empresa := fmt.Sprint(req.Empresa)
	tx, err := a.pg.BeginTx(r.Context(), nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erro ao iniciar transação"})
		return
	}
	defer tx.Rollback()

	var activityID int
	err = tx.QueryRowContext(r.Context(), `INSERT INTO atividades (empresa, seqlocal, usuario_id, data_fim) VALUES ($1,$2,$3,now()) RETURNING id`, empresa, req.SeqLocal, u.ID).Scan(&activityID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar atividade"})
		return
	}
	for _, p := range req.Predio {
		if _, err := tx.ExecContext(r.Context(), `INSERT INTO atividade_enderecos (atividade_id, rua, predio) VALUES ($1,$2,$3)`, activityID, req.Rua, p); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar endereço"})
			return
		}
	}

	read := map[int]struct {
		Status    string
		Reposicao bool
		Predio    string
	}{}
	seqSet := map[int]bool{}
	for _, p := range req.ReadProducts {
		read[p.SeqProduto] = struct {
			Status    string
			Reposicao bool
			Predio    string
		}{p.Status, p.Reposicao, p.Predio}
		seqSet[p.SeqProduto] = true
	}
	for _, p := range req.ExpectedProducts {
		seqSet[p.SeqProduto] = true
	}
	for seq := range seqSet {
		status := "RUPTURA"
		reposicao := false
		predioLido := sql.NullString{}
		ruaLida := sql.NullString{}
		if rp, ok := read[seq]; ok {
			status = rp.Status
			reposicao = rp.Reposicao
			ruaLida = sql.NullString{String: req.Rua, Valid: true}
			predioLido = sql.NullString{String: firstNonEmpty(rp.Predio, req.Predio[0]), Valid: true}
		}
		_, err = tx.ExecContext(r.Context(), `INSERT INTO produto_verificacao (atividade_id, seqproduto, empresa, rua_lida, predio_lido, status, reposicao, estoque) VALUES ($1,$2,$3,$4,$5,$6,$7,0)`, activityID, seq, empresa, ruaLida, predioLido, status, reposicao)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar produtos"})
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erro ao finalizar atividade"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "atividadeId": activityID, "dataFim": time.Now()})
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func intQuery(r *http.Request, name string) int {
	v, _ := strconv.Atoi(r.URL.Query().Get(name))
	return v
}

func (a *App) findUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := a.pg.QueryRowContext(ctx, `SELECT id, username, password, role FROM users WHERE username=$1`, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role)
	return &u, err
}

func (a *App) findUserByID(ctx context.Context, id int) (*User, error) {
	var u User
	err := a.pg.QueryRowContext(ctx, `SELECT id, username, password, role FROM users WHERE id=$1`, id).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role)
	return &u, err
}

func (a *App) listUsers(ctx context.Context) ([]User, error) {
	rows, err := a.pg.QueryContext(ctx, `SELECT id, username, '' AS password, role FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role) == nil {
			users = append(users, u)
		}
	}
	return users, rows.Err()
}

func parseFilters(r *http.Request) ActivityFilters {
	q := r.URL.Query()
	return ActivityFilters{
		Username:     q["filter_username[]"],
		Empresa:      q["filter_empresa[]"],
		Rua:          q["filter_rua[]"],
		Predio:       q["filter_predio[]"],
		Impresso:     q["filter_impresso[]"],
		ID:           q["filter_id[]"],
		DataFimStart: q.Get("filter_dataFimStart"),
		DataFimEnd:   q.Get("filter_dataFimEnd"),
		Sort:         q.Get("sort"),
		Order:        q.Get("order"),
	}
}

func (a *App) listActivities(ctx context.Context, f ActivityFilters, limit int) ([]Activity, error) {
	var args []any
	where := []string{"1=1"}
	addIn := func(col string, vals []string) {
		if len(vals) == 0 {
			return
		}
		var marks []string
		for _, v := range vals {
			args = append(args, v)
			marks = append(marks, fmt.Sprintf("$%d", len(args)))
		}
		where = append(where, col+" IN ("+strings.Join(marks, ",")+")")
	}
	addIn("u.username", f.Username)
	addIn("a.empresa", f.Empresa)
	addIn("e.rua", f.Rua)
	addIn("e.predio", f.Predio)
	if len(f.ID) > 0 {
		addIn("a.id::text", f.ID)
	}
	if len(f.Impresso) == 1 {
		args = append(args, f.Impresso[0] == "S")
		where = append(where, fmt.Sprintf("a.impresso=$%d", len(args)))
	}
	if f.DataFimStart != "" {
		args = append(args, f.DataFimStart)
		where = append(where, fmt.Sprintf("a.data_fim >= $%d::date", len(args)))
	}
	if f.DataFimEnd != "" {
		args = append(args, f.DataFimEnd)
		where = append(where, fmt.Sprintf("a.data_fim < ($%d::date + interval '1 day')", len(args)))
	}
	sortMap := map[string]string{"id": "a.id", "username": "u.username", "empresa": "a.empresa", "rua": "e.rua", "predio": "e.predio", "dataFim": "a.data_fim", "impresso": "a.impresso"}
	sort := "a.data_fim"
	if col := sortMap[f.Sort]; col != "" {
		sort = col
	}
	order := "DESC"
	if f.Order == "asc" {
		order = "ASC"
	}
	args = append(args, limit)
	query := fmt.Sprintf(`
		SELECT a.id, a.empresa, COALESCE(a.seqlocal,0), COALESCE(a.usuario_id,0), COALESCE(u.username,''), a.data_fim, a.impresso, COALESCE(e.rua,''), COALESCE(e.predio,'')
		FROM atividades a
		LEFT JOIN users u ON u.id=a.usuario_id
		LEFT JOIN atividade_enderecos e ON e.atividade_id=a.id
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d`, strings.Join(where, " AND "), sort, order, len(args))
	rows, err := a.pg.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byID := map[int]*Activity{}
	var orderIDs []int
	for rows.Next() {
		var x Activity
		if err := rows.Scan(&x.ID, &x.Empresa, &x.SeqLocal, &x.UserID, &x.Username, &x.DataFim, &x.Impresso, &x.Rua, &x.Predio); err != nil {
			return nil, err
		}
		if cur := byID[x.ID]; cur != nil {
			if x.Predio != "" {
				cur.Predios = append(cur.Predios, x.Predio)
				cur.Predio = strings.Join(cur.Predios, ", ")
			}
			continue
		}
		if x.Predio != "" {
			x.Predios = []string{x.Predio}
		}
		byID[x.ID] = &x
		orderIDs = append(orderIDs, x.ID)
	}
	var out []Activity
	for _, id := range orderIDs {
		out = append(out, *byID[id])
	}
	return out, rows.Err()
}

func (a *App) listFilterOptions(ctx context.Context) (FilterOptions, error) {
	read := func(query string) []string {
		rows, err := a.pg.QueryContext(ctx, query)
		if err != nil {
			return nil
		}
		defer rows.Close()
		var out []string
		for rows.Next() {
			var v sql.NullString
			if rows.Scan(&v) == nil && v.Valid && v.String != "" {
				out = append(out, v.String)
			}
		}
		return out
	}
	return FilterOptions{
		Username: read(`SELECT DISTINCT u.username FROM atividades a LEFT JOIN users u ON u.id=a.usuario_id WHERE u.username IS NOT NULL ORDER BY 1`),
		Empresa:  read(`SELECT DISTINCT empresa FROM atividades ORDER BY 1`),
		Rua:      read(`SELECT DISTINCT rua FROM atividade_enderecos ORDER BY 1`),
		Predio:   read(`SELECT DISTINCT predio FROM atividade_enderecos ORDER BY 1`),
		ID:       read(`SELECT DISTINCT id::text FROM atividades ORDER BY 1`),
		Impresso: []string{"S", "N"},
	}, nil
}

func (a *App) activityDetailsData(ctx context.Context, id int) (Activity, []ProductVerification, error) {
	acts, err := a.listActivities(ctx, ActivityFilters{ID: []string{strconv.Itoa(id)}}, 1)
	if err != nil || len(acts) == 0 {
		return Activity{}, nil, sql.ErrNoRows
	}
	rows, err := a.pg.QueryContext(ctx, `SELECT id, atividade_id, seqproduto, empresa, rua_lida, predio_lido, rua_esperada, predio_esperado, status, reposicao, estoque, data_entrada FROM produto_verificacao WHERE atividade_id=$1 ORDER BY id`, id)
	if err != nil {
		return Activity{}, nil, err
	}
	defer rows.Close()
	var items []ProductVerification
	for rows.Next() {
		var p ProductVerification
		if rows.Scan(&p.ID, &p.AtividadeID, &p.SeqProduto, &p.Empresa, &p.RuaLida, &p.PredioLido, &p.RuaEsperada, &p.PredioEsperado, &p.Status, &p.Reposicao, &p.Estoque, &p.DataEntrada) == nil {
			items = append(items, p)
		}
	}
	return acts[0], items, nil
}

func (a *App) findAddressByCode(ctx context.Context, codigo string, empresa, seqlocal int) (OracleProduct, error) {
	var p OracleProduct
	err := a.ora.QueryRowContext(ctx, `
		SELECT mpc.SEQPRODUTO, mpc.CODACESSO, mrlp.NRORUA, mrlp.NROPREDIO, mp.DESCCOMPLETA
		FROM CONSINCO.MAP_PRODCODIGO mpc
		LEFT JOIN CONSINCO.MRL_PRODLOCAL mrlp ON mrlp.SEQPRODUTO=mpc.SEQPRODUTO AND mrlp.SEQLOCAL=:1 AND mrlp.NROEMPRESA=:2
		LEFT JOIN CONSINCO.MAP_PRODUTO mp ON mp.SEQPRODUTO=mpc.SEQPRODUTO
		WHERE mpc.CODACESSO=:3`, seqlocal, empresa, codigo,
	).Scan(&p.SEQPRODUTO, &p.CODACESSO, &p.NRORUA, &p.NROPREDIO, &p.DESCCOMPLETA)
	return p, err
}

func (a *App) findFullProductByCode(ctx context.Context, codigo string, empresa, seqlocal int) (OracleProduct, error) {
	addr, err := a.findAddressByCode(ctx, codigo, empresa, seqlocal)
	if err != nil {
		return OracleProduct{}, err
	}
	var p OracleProduct
	err = a.ora.QueryRowContext(ctx, `
		SELECT mpe.SEQPRODUTO, mpe.NROEMPRESA, mpe.DTAULTENTRADA, mpe.DTAULTVENDA, mpe.ESTQLOJA, mpe.MEDVDIAGERAL,
		       mp.DESCCOMPLETA, mrlp.NRORUA, mrlp.NROPREDIO,
		       (SELECT MAX(marca) FROM CONSINCO.ETLV_PRODUTO WHERE SEQPRODUTO = mpe.SEQPRODUTO) AS MARCA,
		       CONSINCO.fBuscaPrecoAtualPdv(mpe.SEQPRODUTO, 1, mpe.NROEMPRESA) AS PRECO_VENDA,
		       (SELECT LISTAGG(CODACESSO, '|') WITHIN GROUP (ORDER BY CODACESSO) FROM CONSINCO.MAP_PRODCODIGO WHERE SEQPRODUTO = mpe.SEQPRODUTO) AS CODIGOS
		FROM CONSINCO.MRL_PRODUTOEMPRESA mpe
		LEFT JOIN CONSINCO.MAP_PRODUTO mp ON mp.SEQPRODUTO=mpe.SEQPRODUTO
		LEFT JOIN CONSINCO.MRL_PRODLOCAL mrlp ON mrlp.SEQPRODUTO=mpe.SEQPRODUTO AND mrlp.NROEMPRESA=mpe.NROEMPRESA AND mrlp.SEQLOCAL=:1
		WHERE mpe.NROEMPRESA=:2 AND mpe.SEQPRODUTO=:3`, seqlocal, empresa, addr.SEQPRODUTO,
	).Scan(&p.SEQPRODUTO, &p.NROEMPRESA, &p.DTAULTENTRADA, &p.DTAULTVENDA, &p.ESTQLOJA, &p.MEDVDIAGERAL, &p.DESCCOMPLETA, &p.NRORUA, &p.NROPREDIO, &p.MARCA, &p.PRECO_VENDA, &p.CODIGOS)
	return p, err
}

func (a *App) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.tpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (a *App) style(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write([]byte(css))
}

func parseTemplates() *template.Template {
	funcs := template.FuncMap{
		"rowUser": func(u User) map[string]any {
			return map[string]any{"RowUser": u}
		},
		"date": func(t time.Time) string {
			if t.IsZero() {
				return "-"
			}
			return t.Format("02/01/2006 15:04")
		},
		"rolePt": func(r string) string {
			switch r {
			case "sysadmin":
				return "Administrador"
			case "gerente":
				return "Gerente"
			default:
				return "Conferente"
			}
		},
		"checked": func(v bool) string {
			if v {
				return "S"
			}
			return "N"
		},
	}
	return template.Must(template.New("app").Funcs(funcs).Parse(templates))
}

const css = `
*{box-sizing:border-box}body{margin:0;font-family:Inter,system-ui,-apple-system,Segoe UI,sans-serif;background:#f6f7fb;color:#111827}a{color:#2563eb;text-decoration:none}.top{height:56px;background:#fff;border-bottom:1px solid #e5e7eb;display:flex;align-items:center;justify-content:space-between;padding:0 20px;position:sticky;top:0;z-index:5}.brand{font-weight:800}.wrap{max-width:1180px;margin:0 auto;padding:22px}.panel{background:#fff;border:1px solid #e5e7eb;border-radius:8px;padding:18px}.grid{display:grid;gap:14px}.grid-3{grid-template-columns:repeat(3,minmax(0,1fr))}.row{display:flex;gap:10px;align-items:center;flex-wrap:wrap}.btn,button{border:0;border-radius:6px;background:#2563eb;color:#fff;padding:9px 12px;font-weight:700;cursor:pointer}.btn.secondary,button.secondary{background:#e5e7eb;color:#111827}.btn.ghost,button.ghost{background:transparent;color:#2563eb}input,select{border:1px solid #d1d5db;border-radius:6px;padding:9px 10px;background:#fff;min-height:38px}label{font-size:12px;font-weight:800;color:#6b7280;text-transform:uppercase;display:block;margin-bottom:5px}.field{display:flex;flex-direction:column;gap:4px}.table{width:100%;border-collapse:collapse}.table th,.table td{border-bottom:1px solid #e5e7eb;padding:10px;text-align:left;font-size:14px}.table th{font-size:12px;color:#6b7280;text-transform:uppercase;background:#fafafa}.muted{color:#6b7280}.danger{color:#dc2626}.ok{color:#059669}.message{padding:10px 12px;border-radius:6px;margin-bottom:12px;background:#ecfdf5;color:#047857}.message.err{background:#fef2f2;color:#b91c1c}.stats{display:grid;grid-template-columns:repeat(3,1fr);gap:12px;margin-bottom:16px}.stat{background:#fff;border:1px solid #e5e7eb;border-radius:8px;padding:14px}.stat strong{font-size:24px}.modal{position:fixed;inset:0;background:rgba(17,24,39,.45);display:grid;place-items:center;padding:20px}.modal>div{max-width:720px;width:100%;background:#fff;border-radius:8px;padding:18px;max-height:90vh;overflow:auto}.scan-shell{max-width:520px;margin:0 auto}.screen-title{font-size:22px;margin:0 0 12px}.print-page{background:#fff}.print-card{break-inside:avoid;border:1px solid #ddd;margin:0 0 18px;padding:16px}@media(max-width:760px){.grid-3,.stats{grid-template-columns:1fr}.wrap{padding:14px}.top{padding:0 12px}.table{font-size:12px}}@media print{.top,.noprint,.btn,button{display:none}.wrap{max-width:none}.panel{border:0}.print-card{page-break-inside:avoid}}`

const templates = `
{{define "head"}}<meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><link rel="stylesheet" href="/style.css"><script src="https://unpkg.com/htmx.org@2.0.4"></script>{{end}}

{{define "home"}}<!doctype html><html lang="pt-BR"><head><title>SIMP</title>{{template "head" .}}</head><body><header class="top"><div class="brand">SIMP</div><a class="btn" href="/login">Entrar</a></header><main class="wrap"><section class="panel"><h1>Controle de Atividades</h1><p class="muted">Verificação de produtos, dashboard e administração.</p></section></main></body></html>{{end}}

{{define "login"}}<!doctype html><html lang="pt-BR"><head><title>Login - SIMP</title>{{template "head" .}}</head><body><main class="wrap" style="max-width:420px"><section class="panel"><h1 class="screen-title">Acesso SIMP</h1>{{with .Error}}<p class="message err">{{.}}</p>{{end}}<form method="post" action="/login" class="grid"><div class="field"><label>Usuário</label><input name="username" required autofocus></div><div class="field"><label>Senha</label><input type="password" name="password" required></div><button>Entrar</button></form></section></main></body></html>{{end}}

{{define "nav"}}<header class="top"><div class="brand">SIMP</div><nav class="row"><a href="/atividades">Atividades</a>{{if or (eq .User.Role "gerente") (eq .User.Role "sysadmin")}}<a href="/dashboard">Dashboard</a>{{end}}{{if eq .User.Role "sysadmin"}}<a href="/admin">Admin</a>{{end}}<form method="post" action="/logout"><button class="secondary">Sair</button></form></nav></header>{{end}}

{{define "atividades"}}<!doctype html><html lang="pt-BR"><head><title>Atividades - SIMP</title>{{template "head" .}}</head><body>{{template "nav" .}}<main class="wrap scan-shell"><section class="panel grid"><h1 class="screen-title">Nova Verificação</h1><div class="grid grid-3"><div class="field"><label>Empresa</label><select id="empresa"></select></div><div class="field"><label>Local</label><select id="local"></select></div><div class="field"><label>Rua</label><input id="rua" inputmode="numeric"></div><div class="field"><label>Prédio</label><input id="predio" inputmode="numeric"></div></div><div class="row"><button type="button" onclick="loadExpected()">Carregar Produtos</button><button type="button" class="secondary" onclick="finish()">Finalizar</button></div><div class="field"><label>EAN / Código</label><input id="codigo" inputmode="numeric" onkeydown="if(event.key==='Enter'){event.preventDefault();scan()}"></div><div id="scan-status" class="muted">Carregue empresa/local e escaneie produtos.</div><table class="table"><thead><tr><th>SEQ</th><th>Status</th><th>Rua</th><th>Prédio</th></tr></thead><tbody id="read"></tbody></table></section></main><script>
let expected=[],read=[];
async function j(url,opt){let r=await fetch(url,opt);if(!r.ok)throw new Error(await r.text());return r.json()}
async function boot(){try{let es=await j('/api/empresas');empresa.innerHTML=es.map(e=>'<option value="'+e.NROEMPRESA+'">'+e.NROEMPRESA+' - '+e.NOMEREDUZIDO+'</option>').join('');await locais()}catch(e){scanStatus('Erro ao carregar empresas')}}
async function locais(){let ls=await j('/api/locais?empresa='+empresa.value);local.innerHTML=ls.map(l=>'<option value="'+l.SEQLOCAL+'">'+l.SEQLOCAL+' - '+l.LOCAL+'</option>').join('')}
empresa?.addEventListener('change',locais);
function scanStatus(s){document.getElementById('scan-status').textContent=s}
async function loadExpected(){expected=await j('/api/produtos/local?empresa='+empresa.value+'&seqlocal='+local.value+'&rua='+rua.value+'&predio='+predio.value);scanStatus(expected.length+' produtos esperados')}
async function scan(){let c=codigo.value.trim();if(!c)return;codigo.value='';try{let p=await j('/api/produtos/ean/'+encodeURIComponent(c)+'?empresa='+empresa.value+'&seqlocal='+local.value);let st=(p.NRORUA?.String||p.NRORUA)==rua.value&&(p.NROPREDIO?.String||p.NROPREDIO)==predio.value?'OK':'DIVERGENTE';let item={seqproduto:p.SEQPRODUTO,ean:c,rua:p.NRORUA?.String||p.NRORUA,predio:p.NROPREDIO?.String||p.NROPREDIO,status:st,reposicao:false};read.push(item);renderRead();scanStatus(st)}catch(e){scanStatus('Produto não encontrado')}}
function renderRead(){document.getElementById('read').innerHTML=read.map(p=>'<tr><td>'+p.seqproduto+'</td><td>'+p.status+'</td><td>'+p.rua+'</td><td>'+p.predio+'</td></tr>').join('')}
async function finish(){let body={empresa:empresa.value,seqlocal:Number(local.value),rua:rua.value,predio:[predio.value],readProducts:read,expectedProducts:expected.map(p=>({seqproduto:p.SEQPRODUTO}))};let res=await j('/api/atividades/finalizar',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});scanStatus('Atividade #'+res.atividadeId+' finalizada');read=[];renderRead()}
boot();
</script></body></html>{{end}}

{{define "dashboard"}}<!doctype html><html lang="pt-BR"><head><title>Dashboard - SIMP</title>{{template "head" .}}</head><body>{{template "nav" .}}<main class="wrap"><section class="stats"><div class="stat"><span class="muted">Atividades</span><br><strong>{{len .Activities}}</strong></div><div class="stat"><span class="muted">Finalizadas</span><br><strong>{{len .Activities}}</strong></div><div class="stat"><span class="muted">Usuário</span><br><strong>{{.User.Username}}</strong></div></section><section class="panel"><form class="row noprint" hx-get="/dashboard/table" hx-target="#activity-table" hx-push-url="true"><input name="filter_id[]" placeholder="ID"><input name="filter_empresa[]" placeholder="Empresa"><input name="filter_username[]" placeholder="Usuário"><button>Filtrar</button><a class="btn secondary" href="/dashboard">Limpar</a></form><div id="activity-table">{{template "activities_table" .}}</div></section><div id="modal"></div></main></body></html>{{end}}

{{define "activities_table"}}<table class="table"><thead><tr><th></th><th>ID</th><th>Usuário</th><th>Empresa</th><th>Rua</th><th>Prédio</th><th>Data</th><th>Impresso</th><th></th></tr></thead><tbody>{{range .Activities}}<tr><td><input type="checkbox" class="bulk" value="{{.ID}}"></td><td>#{{.ID}}</td><td>{{.Username}}</td><td>{{.Empresa}}</td><td>{{.Rua}}</td><td>{{.Predio}}</td><td>{{date .DataFim}}</td><td>{{checked .Impresso}}</td><td class="row"><button class="secondary" hx-get="/dashboard/activities/{{.ID}}/details" hx-target="#modal">Detalhes</button><a class="btn" href="/dashboard/activities/{{.ID}}/print-view">Imprimir</a></td></tr>{{else}}<tr><td colspan="9" class="muted">Nenhuma atividade encontrada.</td></tr>{{end}}</tbody></table><button class="secondary noprint" onclick="bulkPrint()">Imprimir selecionadas</button><script>function bulkPrint(){let ids=[...document.querySelectorAll('.bulk:checked')].map(x=>x.value).join(',');if(ids) location.href='/dashboard/activities/print-view/bulk?ids='+ids}</script>{{end}}

{{define "activity_modal"}}<div class="modal" onclick="if(event.target===this)this.remove()"><div><div class="row" style="justify-content:space-between"><h2>Atividade #{{.Activity.ID}}</h2><button class="secondary" onclick="document.getElementById('modal').innerHTML=''">Fechar</button></div><p class="muted">{{.Activity.Username}} · {{.Activity.Empresa}} · Rua {{.Activity.Rua}} / Prédio {{.Activity.Predio}} · {{date .Activity.DataFim}}</p><table class="table"><thead><tr><th>Produto</th><th>Status</th><th>Lido</th><th>Esperado</th><th>Estoque</th></tr></thead><tbody>{{range .Items}}<tr><td>{{.SeqProduto}}</td><td>{{.Status}}</td><td>{{.RuaLida.String}}/{{.PredioLido.String}}</td><td>{{.RuaEsperada.String}}/{{.PredioEsperado.String}}</td><td>{{.Estoque}}</td></tr>{{end}}</tbody></table></div></div>{{end}}

{{define "print"}}<!doctype html><html lang="pt-BR"><head><title>Impressão - SIMP</title>{{template "head" .}}</head><body><main class="wrap print-page"><button class="noprint" onclick="print()">Imprimir</button>{{range .Bundles}}<section class="print-card"><h2>Atividade #{{.Activity.ID}}</h2><p>{{.Activity.Empresa}} · Rua {{.Activity.Rua}} / Prédio {{.Activity.Predio}} · {{date .Activity.DataFim}}</p><table class="table"><thead><tr><th>Produto</th><th>Status</th><th>Estoque</th></tr></thead><tbody>{{range .Items}}<tr><td>{{.SeqProduto}}</td><td>{{.Status}}</td><td>{{.Estoque}}</td></tr>{{end}}</tbody></table></section>{{end}}</main></body></html>{{end}}

{{define "admin"}}<!doctype html><html lang="pt-BR"><head><title>Admin - SIMP</title>{{template "head" .}}</head><body>{{template "nav" .}}<main class="wrap"><section class="panel"><h1 class="screen-title">Usuários</h1><div id="users-section">{{template "users_section" .}}</div></section></main></body></html>{{end}}

{{define "users_section"}}{{with .Message}}<p class="message {{if $.Error}}err{{end}}">{{.}}</p>{{end}}<form class="row" hx-post="/admin/users" hx-target="#users-section"><input name="username" placeholder="Usuário" required><input type="password" name="password" placeholder="Senha" required><select name="role"><option value="conferente">Conferente</option><option value="gerente">Gerente</option><option value="sysadmin">Administrador</option></select><button>Criar</button></form><table class="table"><thead><tr><th>ID</th><th>Usuário</th><th>Função</th><th></th></tr></thead><tbody>{{range .Users}}{{template "user_row" (rowUser .)}}{{end}}</tbody></table>{{end}}

{{define "user_row"}}<tr id="user-{{.RowUser.ID}}"><td>{{.RowUser.ID}}</td><td>{{.RowUser.Username}}</td><td>{{rolePt .RowUser.Role}}</td><td><button class="secondary" hx-get="/admin/users/{{.RowUser.ID}}/edit" hx-target="#user-{{.RowUser.ID}}" hx-swap="outerHTML">Editar</button></td></tr>{{end}}

{{define "user_edit_row"}}<tr id="user-{{.RowUser.ID}}"><td>{{.RowUser.ID}}</td><td>{{.RowUser.Username}}</td><td colspan="2"><form class="row" hx-post="/admin/users/{{.RowUser.ID}}" hx-target="#user-{{.RowUser.ID}}" hx-swap="outerHTML"><select name="role"><option value="conferente">Conferente</option><option value="gerente">Gerente</option><option value="sysadmin">Administrador</option></select><input type="password" name="password" placeholder="Nova senha opcional"><button>Salvar</button><button type="button" class="secondary" hx-get="/admin/users/{{.RowUser.ID}}/row" hx-target="#user-{{.RowUser.ID}}" hx-swap="outerHTML">Cancelar</button></form></td></tr>{{end}}
`
