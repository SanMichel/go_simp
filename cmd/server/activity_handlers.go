package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func (a *App) apiEmpresas(w http.ResponseWriter, r *http.Request, u *User) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	rows, err := a.ora.QueryContext(ctx, `SELECT me.NROEMPRESA, me.NOMEREDUZIDO FROM CONSINCO.MAX_EMPRESA me WHERE me.STATUS = 'A' ORDER BY me.NROEMPRESA`)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeInternal, Message: "Erro ao buscar empresas", HTTPStatus: http.StatusInternalServerError, Err: err})
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
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	empresa, _ := strconv.Atoi(r.URL.Query().Get("empresa"))
	rows, err := a.ora.QueryContext(ctx, `SELECT ml.SEQLOCAL, ml.NROEMPRESA, ml.LOCAL FROM CONSINCO.MRL_LOCAL ml WHERE ml.STATUS = 'A' AND ml.NROEMPRESA = :1 ORDER BY ml.SEQLOCAL`, empresa)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeInternal, Message: "Erro ao buscar locais", HTTPStatus: http.StatusInternalServerError, Err: err})
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
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	codigo := r.PathValue("codigo")
	empresa, _ := strconv.Atoi(r.URL.Query().Get("empresa"))
	seqlocal, _ := strconv.Atoi(r.URL.Query().Get("seqlocal"))
	p, err := a.findAddressByCode(ctx, codigo, empresa, seqlocal)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeNotFound, Message: "Produto não encontrado", HTTPStatus: http.StatusNotFound, Err: err})
		return
	}
	writeJSON(w, http.StatusOK, mapOracleProduct(p))
}

func (a *App) apiProdutoConsulta(w http.ResponseWriter, r *http.Request, u *User) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	codigo := r.PathValue("codigo")
	empresa, _ := strconv.Atoi(r.URL.Query().Get("empresa"))
	seqlocal, _ := strconv.Atoi(r.URL.Query().Get("seqlocal"))
	p, err := a.findFullProductByCode(ctx, codigo, empresa, seqlocal)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeNotFound, Message: "Produto não encontrado", HTTPStatus: http.StatusNotFound, Err: err})
		return
	}
	writeJSON(w, http.StatusOK, mapOracleProduct(p))
}

func (a *App) apiProdutoConsultaDescricao(w http.ResponseWriter, r *http.Request, u *User) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	q := r.URL.Query().Get("q")
	empresa, _ := strconv.Atoi(r.URL.Query().Get("empresa"))
	seqlocal, _ := strconv.Atoi(r.URL.Query().Get("seqlocal"))
	if q == "" {
		a.handleError(w, r, &AppError{Code: ErrCodeBadRequest, Message: "Parâmetro 'q' é obrigatório", HTTPStatus: http.StatusBadRequest})
		return
	}
	products, err := a.findProductsByDescription(ctx, q, empresa, seqlocal)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeInternal, Message: "Erro na consulta", HTTPStatus: http.StatusInternalServerError, Err: err})
		return
	}
	resp := make([]OracleProductResponse, len(products))
	for i, p := range products {
		resp[i] = mapOracleProduct(p)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (a *App) apiProdutosLocal(w http.ResponseWriter, r *http.Request, u *User) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	empresa, _ := strconv.Atoi(r.URL.Query().Get("empresa"))
	seqlocal, _ := strconv.Atoi(r.URL.Query().Get("seqlocal"))
	rua := r.URL.Query().Get("rua")
	predio := r.URL.Query().Get("predio")
	rows, err := a.ora.QueryContext(ctx, `SELECT mrlp.SEQPRODUTO, mrlp.NRORUA, mrlp.NROPREDIO, mrlp.ESTOQUE FROM CONSINCO.MRL_PRODLOCAL mrlp WHERE mrlp.SEQLOCAL = :1 AND mrlp.ESTOQUE > 0 AND mrlp.NRORUA = :2 AND mrlp.NROPREDIO = :3 AND mrlp.NROEMPRESA = :4`, seqlocal, rua, predio, empresa)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeInternal, Message: "Erro ao buscar produtos", HTTPStatus: http.StatusInternalServerError, Err: err})
		return
	}
	defer rows.Close()
	type row struct {
		SEQPRODUTO int    `json:"seqproduto"`
		NRORUA     string `json:"nrorua"`
		NROPREDIO  string `json:"nropredio"`
		ESTOQUE    int    `json:"estoque"`
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

func (a *App) finalizeActivity(ctx context.Context, req finalizeReq, userID int) (*FinalizarResult, error) {
	if len(req.Predio) == 0 {
		req.Predio = []string{""}
	}
	empresa := strconv.Itoa(req.Empresa)
	tx, err := a.pg.BeginTx(ctx, nil)
	if err != nil {
		return nil, &AppError{Code: ErrCodeInternal, Message: "Erro ao iniciar transação", HTTPStatus: http.StatusInternalServerError, Err: err}
	}
	defer tx.Rollback()

	var activityID int
	err = tx.QueryRowContext(ctx, `INSERT INTO atividades (empresa, seqlocal, usuario_id, data_fim) VALUES ($1,$2,$3,now()) RETURNING id`, empresa, req.SeqLocal, userID).Scan(&activityID)
	if err != nil {
		return nil, &AppError{Code: ErrCodeInternal, Message: "Erro ao salvar atividade", HTTPStatus: http.StatusInternalServerError, Err: err}
	}
	for _, p := range req.Predio {
		if _, err := tx.ExecContext(ctx, `INSERT INTO atividade_enderecos (atividade_id, rua, predio) VALUES ($1,$2,$3)`, activityID, req.Rua, p); err != nil {
			return nil, &AppError{Code: ErrCodeInternal, Message: "Erro ao salvar endereço", HTTPStatus: http.StatusInternalServerError, Err: err}
		}
	}

	read := map[int]struct {
		Status       string
		Reposicao    bool
		Predio       string
		Desccompleta string
		EAN          string
	}{}
	seqSet := map[int]bool{}
	expectedSeqs := map[int]bool{}
	for _, p := range req.ReadProducts {
		read[p.SeqProduto] = struct {
			Status       string
			Reposicao    bool
			Predio       string
			Desccompleta string
			EAN          string
		}{p.Status, p.Reposicao, p.Predio, p.Desccompleta, p.EAN}
		seqSet[p.SeqProduto] = true
	}
	for _, p := range req.ExpectedProducts {
		expectedSeqs[p.SeqProduto] = true
		seqSet[p.SeqProduto] = true
	}
	result := &FinalizarResult{
		Divergences:    make([]map[string]any, 0),
		Ruptures:       make([]map[string]any, 0),
		Replenishments: make([]map[string]any, 0),
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
			if rp.Status == "DIVERGENTE" || rp.Status == "ERRO" {
				result.Divergences = append(result.Divergences, map[string]any{
					"seqproduto":   seq,
					"ean":          rp.EAN,
					"desccompleta": rp.Desccompleta,
				})
			}
			if rp.Reposicao {
				result.Replenishments = append(result.Replenishments, map[string]any{
					"seqproduto":   seq,
					"ean":          rp.EAN,
					"desccompleta": rp.Desccompleta,
				})
			}
		} else if expectedSeqs[seq] {
			result.Ruptures = append(result.Ruptures, map[string]any{"seqproduto": seq})
		}
		desc := sql.NullString{}
		if rp, ok := read[seq]; ok && rp.Desccompleta != "" {
			desc = sql.NullString{String: rp.Desccompleta, Valid: true}
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO produto_verificacao (atividade_id, seqproduto, empresa, rua_lida, predio_lido, status, reposicao, estoque, desccompleta) VALUES ($1,$2,$3,$4,$5,$6,$7,0,$8)`, activityID, seq, empresa, ruaLida, predioLido, status, reposicao, desc)
		if err != nil {
			return nil, &AppError{Code: ErrCodeInternal, Message: "Erro ao salvar produtos", HTTPStatus: http.StatusInternalServerError, Err: err}
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, &AppError{Code: ErrCodeInternal, Message: "Erro ao finalizar atividade", HTTPStatus: http.StatusInternalServerError, Err: err}
	}
	result.ActivityID = activityID
	result.DataFim = time.Now()
	return result, nil
}

func (a *App) apiFinalizar(w http.ResponseWriter, r *http.Request, u *User) {
	var req finalizeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeBadRequest, Message: "JSON inválido", HTTPStatus: http.StatusBadRequest})
		return
	}
	if req.Empresa == 0 || req.Rua == "" || req.SeqLocal == 0 {
		a.handleError(w, r, &AppError{Code: ErrCodeBadRequest, Message: "Campos obrigatórios ausentes", HTTPStatus: http.StatusBadRequest})
		return
	}
	result, err := a.finalizeActivity(r.Context(), req, u.ID)
	if err != nil {
		a.handleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "atividadeId": result.ActivityID,
		"dataFim": result.DataFim, "divergences": result.Divergences,
		"ruptures": result.Ruptures, "replenishments": result.Replenishments,
	})
}
