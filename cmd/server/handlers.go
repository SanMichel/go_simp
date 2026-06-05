package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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
		http.Error(w, "session error", http.StatusInternalServerError)
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
		log.Printf("error: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
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
		}
	}
	for _, id := range ids {
		_, _ = a.pg.ExecContext(r.Context(), `UPDATE atividades SET impresso=true WHERE id=$1`, id)
	}
	a.render(w, "print", map[string]any{"Bundles": bundles})
}

func (a *App) adminPage(w http.ResponseWriter, r *http.Request) {
	u, _ := a.currentUser(r)
	users, err := a.listUsers(r.Context())
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
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
		log.Printf("error: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusBadRequest)
		return
	}
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	role := r.FormValue("role")
	if username == "" || len(password) < 8 || !validRole(role) {
		users, _ := a.listUsers(r.Context())
		a.render(w, "users_section", map[string]any{"Users": users, "Message": "Dados inválidos.", "Error": true})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_, err = a.pg.ExecContext(r.Context(), `INSERT INTO users (username, password, role) VALUES ($1,$2,$3)`, username, string(hash), role)
	users, _ := a.listUsers(r.Context())
	if err != nil {
		log.Printf("error: %v", err)
		a.render(w, "users_section", map[string]any{"Users": users, "Message": "Erro interno do servidor", "Error": true})
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
		log.Printf("error: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusBadRequest)
		return
	}
	role := r.FormValue("role")
	password := r.FormValue("password")
	if !validRole(role) {
		http.Error(w, "role inválido", http.StatusBadRequest)
		return
	}
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		_, _ = a.pg.ExecContext(r.Context(), `UPDATE users SET role=$1, password=$2, last_token_at=now() WHERE id=$3`, role, string(hash), id)
	} else {
		_, _ = a.pg.ExecContext(r.Context(), `UPDATE users SET role=$1, last_token_at=now() WHERE id=$2`, role, id)
	}
	u, _ := a.findUserByID(r.Context(), id)
	a.render(w, "user_row", map[string]any{"RowUser": u})
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
	if !a.loginLimiter.allow(r.RemoteAddr) {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "Muitas tentativas. Aguarde 1 minuto."})
		return
	}
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
	http.SetCookie(w, &http.Cookie{Name: "token", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: int(a.cfg.SessionTTL.Seconds()), Secure: true})
	a.setCSRFCookie(w)
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  map[string]any{"id": u.ID, "username": u.Username, "role": u.Role},
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
