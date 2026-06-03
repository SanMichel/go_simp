package main

import (
	"context"
	"database/sql"
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
func (a *App) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.tpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
