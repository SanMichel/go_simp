package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

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
