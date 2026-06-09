package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/sijms/go-ora/v2"
)

func main() {
	if err := loadDotEnv(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal("load .env: ", err)
	}

	cfg := loadConfig()

	pg, err := sql.Open("pgx", cfg.PostgresURL)
	if err != nil {
		log.Fatal(err)
	}
	pg.SetMaxOpenConns(cfg.PGMaxConns)
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
	oracleReader.db.SetMaxOpenConns(cfg.OracleMaxConns)
	oracleReader.db.SetMaxIdleConns(cfg.OracleIdleConns)
	oracleReader.db.SetConnMaxIdleTime(cfg.OracleIdleTime)
	oracleReader.db.SetConnMaxLifetime(time.Hour)
	if err := oracleReader.db.PingContext(ctx); err != nil {
		slog.Warn("oracle ping failed", "error", err)
	}

	app := &App{cfg: cfg, pg: pg, ora: oracleReader, tpl: parseTemplates(), loginLimiter: newRateLimiter()}
	if err := app.migrate(ctx); err != nil {
		log.Fatal(err)
	}
	if err := app.seedAdmin(ctx); err != nil {
		log.Fatal(err)
	}

	var slogHandler slog.Handler
	if cfg.AppEnv == "production" {
		slogHandler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})
	} else {
		slogHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(slogHandler))

	mux := http.NewServeMux()
	app.routes(mux)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      app.csrfMiddleware(app.securityHeaders(app.log(recoveryMiddleware(requestIDMiddleware(mux))))),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		slog.Info("server started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}
func (a *App) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", a.healthCheck)
	mux.HandleFunc("GET /style.css", a.style)
	mux.HandleFunc("GET /admin.css", a.adminStyle)
	mux.HandleFunc("GET /shared.js", a.serveJS("shared.js"))
	mux.HandleFunc("GET /htmx.min.js", a.serveJS("htmx.min.js"))
	mux.HandleFunc("GET /app.js", a.serveJS("app.js"))
	mux.HandleFunc("GET /dashboard.js", a.serveJS("dashboard.js"))
	mux.HandleFunc("GET /admin.js", a.serveJS("admin.js"))
	mux.HandleFunc("GET /login.js", a.serveJS("login.js"))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/login", http.StatusFound) })
	mux.HandleFunc("GET /index.html", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/home", http.StatusFound) })
	mux.HandleFunc("GET /home", a.home)
	mux.HandleFunc("GET /login", a.loginPage)
	mux.HandleFunc("POST /login", a.loginPost)
	mux.HandleFunc("POST /logout", a.logout)
	mux.HandleFunc("GET /atividades", a.requireRole("", a.atividadesPage)) // empty = any authenticated user
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
	mux.HandleFunc("GET /api/empresas", a.requireAPIRole("conferente,gerente,sysadmin", a.apiEmpresas))
	mux.HandleFunc("GET /api/locais", a.requireAPIRole("conferente,gerente,sysadmin", a.apiLocais))
	mux.HandleFunc("GET /api/produtos/ean/{codigo}", a.requireAPIRole("conferente,gerente,sysadmin", a.apiProdutoEAN))
	mux.HandleFunc("GET /api/produtos/consulta/{codigo}", a.requireAPIRole("conferente,gerente,sysadmin", a.apiProdutoConsulta))
	mux.HandleFunc("GET /api/produtos/consulta", a.requireAPIRole("conferente,gerente,sysadmin", a.apiProdutoConsultaDescricao))
	mux.HandleFunc("GET /api/produtos/local", a.requireAPIRole("conferente,gerente,sysadmin", a.apiProdutosLocal))
	mux.HandleFunc("POST /api/atividades/finalizar", a.requireAPIRole("conferente,gerente,sysadmin", a.apiFinalizar))
	mux.HandleFunc("GET /api/atividades/last-info", a.requireAPIRole("conferente,gerente,sysadmin", a.apiLastInfo))

	mux.HandleFunc("GET /api/admin/users", a.requireAPIRole("sysadmin", a.apiAdminUsersList))
	mux.HandleFunc("POST /api/admin/users", a.requireAPIRole("sysadmin", a.apiAdminUserCreate))
	mux.HandleFunc("PATCH /api/admin/users/{id}", a.requireAPIRole("sysadmin", a.apiAdminUserUpdate))

	mux.HandleFunc("GET /api/dashboard/activities/filters", a.requireAPIRole("gerente,sysadmin", a.apiDashboardFilters))
	mux.HandleFunc("GET /api/dashboard/activities", a.requireAPIRole("gerente,sysadmin", a.apiDashboardActivities))
	mux.HandleFunc("GET /api/dashboard/activities/{id}", a.requireAPIRole("gerente,sysadmin", a.apiDashboardActivityDetails))
	mux.HandleFunc("POST /api/dashboard/activities/bulk", a.requireAPIRole("gerente,sysadmin", a.apiDashboardBulkDetails))
	mux.HandleFunc("PATCH /api/dashboard/activities/bulk/print", a.requireAPIRole("gerente,sysadmin", a.apiDashboardBulkPrint))
}
func (a *App) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.tpl.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("template render failed", "template", name, "error", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
	}
}
func (a *App) style(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	b, err := templatesFS.ReadFile("templates/style.css")
	if err != nil {
		slog.Error("failed to read static file", "file", "style.css", "error", err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(b)
}
func (a *App) adminStyle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	b, err := templatesFS.ReadFile("templates/admin.css")
	if err != nil {
		slog.Error("failed to read static file", "file", "admin.css", "error", err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(b)
}
func (a *App) serveJS(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		b, err := templatesFS.ReadFile("templates/" + filename)
		if err != nil {
			slog.Error("failed to read static file", "file", filename, "error", err)
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(b)
	}
}

func parseTemplates() *template.Template {
	funcs := template.FuncMap{
		"rowUser": func(u UserRow) map[string]any {
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
	return template.Must(template.New("app").Funcs(funcs).ParseFS(templatesFS, "templates/*.html", "templates/components/*.html"))
}

//go:embed templates/*.html templates/components/*.html templates/*.css templates/*.js
var templatesFS embed.FS
