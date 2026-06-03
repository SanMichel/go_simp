package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
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
func (a *App) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /style.css", a.style)
	mux.HandleFunc("GET /admin.css", a.adminStyle)
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
	
	mux.HandleFunc("GET /api/admin/users", a.requireAPI(a.apiAdminUsersList))
	mux.HandleFunc("POST /api/admin/users", a.requireAPI(a.apiAdminUserCreate))
	mux.HandleFunc("PATCH /api/admin/users/{id}", a.requireAPI(a.apiAdminUserUpdate))
	
	mux.HandleFunc("GET /api/dashboard/activities/filters", a.requireAPI(a.apiDashboardFilters))
	mux.HandleFunc("GET /api/dashboard/activities", a.requireAPI(a.apiDashboardActivities))
	mux.HandleFunc("GET /api/dashboard/activities/{id}", a.requireAPI(a.apiDashboardActivityDetails))
	mux.HandleFunc("POST /api/dashboard/activities/bulk", a.requireAPI(a.apiDashboardBulkDetails))
	mux.HandleFunc("PATCH /api/dashboard/activities/bulk/print", a.requireAPI(a.apiDashboardBulkPrint))
}
func (a *App) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.tpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func (a *App) style(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	b, _ := templatesFS.ReadFile("templates/style.css")
	_, _ = w.Write(b)
}
func (a *App) adminStyle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	b, _ := templatesFS.ReadFile("templates/admin.css")
	_, _ = w.Write(b)
}
func (a *App) serveJS(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		b, _ := templatesFS.ReadFile("templates/" + filename)
		_, _ = w.Write(b)
	}
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
	return template.Must(template.New("app").Funcs(funcs).ParseFS(templatesFS, "templates/*.html", "templates/components/*.html"))
}

//go:embed templates/*.html templates/components/*.html templates/*.css templates/*.js
var templatesFS embed.FS
