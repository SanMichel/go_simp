package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type APIActivity struct {
	ID       int        `json:"id"`
	Empresa  string     `json:"empresa"`
	SeqLocal int        `json:"seqlocal"`
	UserID   int        `json:"user_id"`
	Username string     `json:"username"`
	DataFim  *time.Time `json:"dataFim"`
	Impresso bool       `json:"impresso"`
	Rua      string     `json:"rua"`
	Predio   string     `json:"predio"`
	Predios  []string   `json:"predios"`
}

type APIProductVerification struct {
	ID             int     `json:"id"`
	AtividadeID    int     `json:"atividade_id"`
	SeqProduto     int     `json:"seqproduto"`
	Empresa        string  `json:"empresa"`
	RuaLida        *string `json:"rua"`
	PredioLido     *string `json:"predio"`
	RuaEsperada    *string `json:"expectedRua"`
	PredioEsperado *string `json:"expectedPredio"`
	Status         string  `json:"status"`
	Reposicao      bool    `json:"reposicao"`
	Estoque        int     `json:"estoque"`
	DataEntrada    *string `json:"data_entrada"`
	DescCompleta   *string `json:"desccompleta"`
	MDV            *float64 `json:"mdv"`
	DDV            *float64 `json:"ddv"`
}

type APIUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func mapActivity(a Activity) APIActivity {
	var dataFim *time.Time
	if !a.DataFim.IsZero() {
		dataFim = &a.DataFim
	}
	return APIActivity{
		ID: a.ID, Empresa: a.Empresa, SeqLocal: a.SeqLocal, UserID: a.UserID, Username: a.Username,
		DataFim: dataFim, Impresso: a.Impresso, Rua: a.Rua, Predio: a.Predio, Predios: a.Predios,
	}
}

func mapProduct(p ProductVerification) APIProductVerification {
	var rua, predio, expRua, expPredio, desc, data *string
	if p.RuaLida.Valid {
		rua = &p.RuaLida.String
	}
	if p.PredioLido.Valid {
		predio = &p.PredioLido.String
	}
	if p.RuaEsperada.Valid {
		expRua = &p.RuaEsperada.String
	}
	if p.PredioEsperado.Valid {
		expPredio = &p.PredioEsperado.String
	}
	if p.DescCompleta.Valid {
		desc = &p.DescCompleta.String
	}
	if p.DataEntrada.Valid {
		d := p.DataEntrada.Time.Format(time.RFC3339)
		data = &d
	}
	var mdv, ddv *float64
	if p.MDV.Valid {
		m := p.MDV.Float64
		mdv = &m
	}
	if p.DDV.Valid {
		d := p.DDV.Float64
		ddv = &d
	}
	return APIProductVerification{
		ID: p.ID, AtividadeID: p.AtividadeID, SeqProduto: p.SeqProduto, Empresa: p.Empresa,
		RuaLida: rua, PredioLido: predio, RuaEsperada: expRua, PredioEsperado: expPredio,
		Status: p.Status, Reposicao: p.Reposicao, Estoque: p.Estoque,
		DataEntrada: data, DescCompleta: desc, MDV: mdv, DDV: ddv,
	}
}

func mapUser(u User) APIUser {
	return APIUser{ID: u.ID, Username: u.Username, Role: u.Role}
}

func (a *App) apiAdminUsersList(w http.ResponseWriter, r *http.Request, u *User) {
	users, err := a.listUsers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	apiUsers := make([]APIUser, len(users))
	for i, usr := range users {
		apiUsers[i] = mapUser(usr)
	}
	writeJSON(w, http.StatusOK, apiUsers)
}

func (a *App) apiAdminUserCreate(w http.ResponseWriter, r *http.Request, u *User) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Username == "" || len(req.Password) < 4 || !validRole(req.Role) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	_, err := a.pg.ExecContext(r.Context(), `INSERT INTO users (username, password, role) VALUES ($1,$2,$3)`, req.Username, string(hash), req.Role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "OK"})
}

func (a *App) apiAdminUserUpdate(w http.ResponseWriter, r *http.Request, u *User) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	var req struct {
		Role     string `json:"role"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if !validRole(req.Role) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Role inválido"})
		return
	}
	if req.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		_, _ = a.pg.ExecContext(r.Context(), `UPDATE users SET role=$1, password=$2 WHERE id=$3`, req.Role, string(hash), id)
	} else {
		_, _ = a.pg.ExecContext(r.Context(), `UPDATE users SET role=$1 WHERE id=$2`, req.Role, id)
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "OK"})
}

func (a *App) apiDashboardFilters(w http.ResponseWriter, r *http.Request, u *User) {
	options, err := a.listFilterOptions(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	// FilterOptions can just have lowercase JSON tags
	type APIFilterOptions struct {
		Username []string `json:"username"`
		Empresa  []string `json:"empresa"`
		Rua      []string `json:"rua"`
		Predio   []string `json:"predio"`
		Impresso []string `json:"impresso"`
		ID       []string `json:"id"`
	}
	apiOptions := APIFilterOptions{
		Username: options.Username, Empresa: options.Empresa, Rua: options.Rua,
		Predio: options.Predio, Impresso: options.Impresso, ID: options.ID,
	}
	writeJSON(w, http.StatusOK, apiOptions)
}

func (a *App) apiDashboardActivities(w http.ResponseWriter, r *http.Request, u *User) {
	activities, err := a.listActivities(r.Context(), parseFilters(r), 200)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	apiActivities := make([]APIActivity, len(activities))
	for i, act := range activities {
		apiActivities[i] = mapActivity(act)
	}
	writeJSON(w, http.StatusOK, apiActivities)
}

func (a *App) apiDashboardActivityDetails(w http.ResponseWriter, r *http.Request, u *User) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	act, items, err := a.activityDetailsData(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Não encontrado"})
		return
	}
	apiItems := make([]APIProductVerification, len(items))
	for i, item := range items {
		apiItems[i] = mapProduct(item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"activity": mapActivity(act), "items": apiItems})
}

func (a *App) apiDashboardBulkDetails(w http.ResponseWriter, r *http.Request, u *User) {
	var req struct {
		Ids []int `json:"ids"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	type Bundle struct {
		Activity APIActivity              `json:"activity"`
		Items    []APIProductVerification `json:"items"`
	}
	var bundles []Bundle
	for _, id := range req.Ids {
		act, items, err := a.activityDetailsData(r.Context(), id)
		if err == nil {
			apiItems := make([]APIProductVerification, len(items))
			for i, item := range items {
				apiItems[i] = mapProduct(item)
			}
			bundles = append(bundles, Bundle{Activity: mapActivity(act), Items: apiItems})
		}
	}
	if bundles == nil {
		bundles = []Bundle{}
	}
	writeJSON(w, http.StatusOK, bundles)
}

func (a *App) apiDashboardBulkPrint(w http.ResponseWriter, r *http.Request, u *User) {
	var req struct {
		Ids []int `json:"ids"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	for _, id := range req.Ids {
		a.pg.ExecContext(r.Context(), `UPDATE atividades SET impresso=true WHERE id=$1`, id)
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "OK"})
}
