package main

import (
	"encoding/json"
	"log/slog"
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
	ID             int      `json:"id"`
	AtividadeID    int      `json:"atividade_id"`
	SeqProduto     int      `json:"seqproduto"`
	Empresa        string   `json:"empresa"`
	RuaLida        *string  `json:"rua"`
	PredioLido     *string  `json:"predio"`
	RuaEsperada    *string  `json:"expectedRua"`
	PredioEsperado *string  `json:"expectedPredio"`
	Status         string   `json:"status"`
	Reposicao      bool     `json:"reposicao"`
	Estoque        int      `json:"estoque"`
	DataEntrada    *string  `json:"data_entrada"`
	DescCompleta   *string  `json:"desccompleta"`
	MDV            *float64 `json:"mdv"`
	DDV            *float64 `json:"ddv"`
	Reincidencia   int      `json:"reincidencia"`
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
		Reincidencia: p.Reincidencia,
	}
}

type OracleProductResponse struct {
	SeqProduto    int      `json:"seqproduto"`
	DescCompleta  *string  `json:"desccompleta,omitempty"`
	Marca         *string  `json:"marca,omitempty"`
	Rua           *string  `json:"rua,omitempty"`
	Predio        *string  `json:"predio,omitempty"`
	Estoque       int      `json:"estoque"`
	Mdv           *float64 `json:"mdv,omitempty"`
	DiasEstoque   *float64 `json:"diasEstoque,omitempty"`
	PrecoVenda    *float64 `json:"precoVenda,omitempty"`
	DtaUltEntrada *string  `json:"dtaUltEntrada,omitempty"`
	DtaUltVenda   *string  `json:"dtaUltVenda,omitempty"`
	Codigos       *string  `json:"codigos,omitempty"`
	Codacesso     *string  `json:"codacesso,omitempty"`
}

func mapOracleProduct(p OracleProduct) OracleProductResponse {
	var desc, marca, codigos, codacesso, rua, predio *string
	var mdv, diasEstoque, precoVenda *float64
	var dtaUltEntrada, dtaUltVenda *string
	if p.DESCCOMPLETA.Valid {
		s := p.DESCCOMPLETA.String
		desc = &s
	}
	if p.MARCA.Valid {
		s := p.MARCA.String
		marca = &s
	}
	if p.CODIGOS.Valid {
		s := p.CODIGOS.String
		codigos = &s
	}
	if p.CODACESSO.Valid {
		s := p.CODACESSO.String
		codacesso = &s
	}
	if p.NRORUA.Valid {
		s := p.NRORUA.String
		rua = &s
	}
	if p.NROPREDIO.Valid {
		s := p.NROPREDIO.String
		predio = &s
	}
	if p.MEDVDIAGERAL.Valid {
		f := p.MEDVDIAGERAL.Float64
		mdv = &f
	}
	if p.DiasEstoque.Valid {
		f := p.DiasEstoque.Float64
		diasEstoque = &f
	}
	if p.PRECO_VENDA.Valid {
		f := p.PRECO_VENDA.Float64
		precoVenda = &f
	}
	if p.DTAULTENTRADA.Valid {
		s := p.DTAULTENTRADA.Time.Format(time.RFC3339)
		dtaUltEntrada = &s
	}
	if p.DTAULTVENDA.Valid {
		s := p.DTAULTVENDA.Time.Format(time.RFC3339)
		dtaUltVenda = &s
	}
	return OracleProductResponse{
		SeqProduto: p.SEQPRODUTO, DescCompleta: desc, Marca: marca,
		Rua: rua, Predio: predio, Estoque: p.ESTQLOJA, Mdv: mdv, DiasEstoque: diasEstoque,
		PrecoVenda: precoVenda, DtaUltEntrada: dtaUltEntrada, DtaUltVenda: dtaUltVenda,
		Codigos: codigos, Codacesso: codacesso,
	}
}

func mapUser(u UserRow) APIUser {
	return APIUser{ID: u.ID, Username: u.Username, Role: u.Role}
}

func (a *App) apiAdminUsersList(w http.ResponseWriter, r *http.Request, u *User) {
	users, err := a.listUsers(r.Context())
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeInternal, Message: "Erro interno do servidor", HTTPStatus: http.StatusInternalServerError, Err: err})
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
	if req.Username == "" || len(req.Password) < 8 || !validRole(req.Role) {
		a.handleError(w, r, &AppError{Code: ErrCodeValidation, Message: "Dados inválidos", HTTPStatus: http.StatusBadRequest})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeInternal, Message: "Erro interno do servidor", HTTPStatus: http.StatusInternalServerError})
		return
	}
	_, err = a.pg.ExecContext(r.Context(), `INSERT INTO users (username, password, role) VALUES ($1,$2,$3)`, req.Username, string(hash), req.Role)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeInternal, Message: "Erro interno do servidor", HTTPStatus: http.StatusInternalServerError, Err: err})
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
	_, err := a.updateUserAdmin(r.Context(), u.ID, id, req.Role, req.Password)
	if err != nil {
		a.handleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "OK"})
}

func (a *App) apiDashboardFilters(w http.ResponseWriter, r *http.Request, u *User) {
	options, err := a.listFilterOptions(r.Context())
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeInternal, Message: "Erro interno do servidor", HTTPStatus: http.StatusInternalServerError, Err: err})
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
		a.handleError(w, r, &AppError{Code: ErrCodeInternal, Message: "Erro interno do servidor", HTTPStatus: http.StatusInternalServerError, Err: err})
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
		a.handleError(w, r, &AppError{Code: ErrCodeNotFound, Message: "Atividade não encontrada", HTTPStatus: http.StatusNotFound})
		return
	}
	apiAct := mapActivity(act)
	apiItems := make([]APIProductVerification, len(items))
	for i, item := range items {
		apiItems[i] = mapProduct(item)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":       apiAct.ID,
		"empresa":  apiAct.Empresa,
		"username": apiAct.Username,
		"dataFim":  apiAct.DataFim,
		"rua":      apiAct.Rua,
		"predio":   apiAct.Predio,
		"impresso": apiAct.Impresso,
		"items":    apiItems,
	})
}

func (a *App) apiDashboardBulkDetails(w http.ResponseWriter, r *http.Request, u *User) {
	var req struct {
		Ids []int `json:"ids"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	type FlatBundle struct {
		ID       int                      `json:"id"`
		Empresa  string                   `json:"empresa"`
		Username string                   `json:"username"`
		DataFim  *time.Time               `json:"dataFim"`
		Rua      string                   `json:"rua"`
		Predio   string                   `json:"predio"`
		Impresso bool                     `json:"impresso"`
		Items    []APIProductVerification `json:"items"`
	}
	var bundles []FlatBundle
	for _, id := range req.Ids {
		act, items, err := a.activityDetailsData(r.Context(), id)
		if err == nil {
			apiAct := mapActivity(act)
			apiItems := make([]APIProductVerification, len(items))
			for i, item := range items {
				apiItems[i] = mapProduct(item)
			}
			bundles = append(bundles, FlatBundle{
				ID: apiAct.ID, Empresa: apiAct.Empresa, Username: apiAct.Username,
				DataFim: apiAct.DataFim, Rua: apiAct.Rua, Predio: apiAct.Predio,
				Impresso: apiAct.Impresso, Items: apiItems,
			})
		}
	}
	if bundles == nil {
		bundles = []FlatBundle{}
	}
	writeJSON(w, http.StatusOK, bundles)
}

func (a *App) apiDashboardBulkPrint(w http.ResponseWriter, r *http.Request, u *User) {
	var req struct {
		Ids []int `json:"ids"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	for _, id := range req.Ids {
		if _, err := a.pg.ExecContext(r.Context(), `UPDATE atividades SET impresso=true WHERE id=$1`, id); err != nil {
			slog.WarnContext(r.Context(), "failed to set impresso", "activity_id", id, "error", err)
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "OK"})
}
