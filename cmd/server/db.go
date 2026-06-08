package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
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
	for {
		trimmed := strings.TrimSpace(q)
		if trimmed == "" {
			return false
		}
		upper := strings.ToUpper(trimmed)
		if strings.HasPrefix(upper, "SELECT") || strings.HasPrefix(upper, "WITH") {
			break
		}
		if strings.HasPrefix(upper, "/*") {
			end := strings.Index(trimmed[2:], "*/")
			if end < 0 {
				return false
			}
			q = trimmed[end+4:]
			continue
		}
		return false
	}
	dml := []string{
		" INSERT", " UPDATE", " DELETE", " DROP", " ALTER", " CREATE",
		" TRUNCATE", " EXEC", " EXECUTE", " MERGE", " REPLACE",
		" GRANT", " REVOKE", " CALL", " LOCK", " RENAME",
		" COMMIT", " ROLLBACK", " BEGIN", " DECLARE",
	}
	sq := removeSQLComments(query)
	upper := " " + strings.ToUpper(sq)
	for _, kw := range dml {
		if strings.Contains(upper, kw+" ") || strings.HasSuffix(upper, kw) {
			return false
		}
	}
	return true
}

func removeSQLComments(q string) string {
	var out strings.Builder
	out.Grow(len(q))
	for i := 0; i < len(q); i++ {
		if q[i] == '-' && i+1 < len(q) && q[i+1] == '-' {
			for i < len(q) && q[i] != '\n' {
				i++
			}
			out.WriteByte(' ')
			continue
		}
		if q[i] == '/' && i+1 < len(q) && q[i+1] == '*' {
			i += 2
			for i+1 < len(q) && !(q[i] == '*' && q[i+1] == '/') {
				i++
			}
			i += 2
			out.WriteByte(' ')
			continue
		}
		if q[i] == '\'' || q[i] == '"' {
			quote := q[i]
			out.WriteByte(quote)
			i++
			for i < len(q) && q[i] != quote {
				if q[i] == '\\' && i+1 < len(q) {
					out.WriteByte(q[i])
					i++
				}
				if i < len(q) {
					out.WriteByte(q[i])
					i++
				}
			}
			if i < len(q) {
				out.WriteByte(q[i])
			}
			continue
		}
		out.WriteByte(q[i])
	}
	return strings.NewReplacer("\n", " ", "\r", " ", "\t", " ").Replace(out.String())
}

func (a *App) migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'conferente',
			last_token_at TIMESTAMPTZ
		)`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS last_token_at TIMESTAMPTZ`,
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
		`ALTER TABLE produto_verificacao ADD COLUMN IF NOT EXISTS desccompleta TEXT`,
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
	pass := randomString(16)
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = a.pg.ExecContext(ctx, `INSERT INTO users (username, password, role) VALUES ($1, $2, 'sysadmin')`, "admin", string(hash))
	if err == nil {
		log.Printf("⚠️  FIRST RUN — admin password: %s  (change immediately)", pass)
	}
	return err
}
func (a *App) findUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := a.pg.QueryRowContext(ctx, `SELECT id, username, password, role, last_token_at FROM users WHERE username=$1`, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.LastTokenAt)
	return &u, err
}

func (a *App) findUserByID(ctx context.Context, id int) (*User, error) {
	var u User
	err := a.pg.QueryRowContext(ctx, `SELECT id, username, password, role, last_token_at FROM users WHERE id=$1`, id).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.LastTokenAt)
	return &u, err
}

func (a *App) listUsers(ctx context.Context) ([]UserRow, error) {
	rows, err := a.pg.QueryContext(ctx, `SELECT id, username, role, last_token_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []UserRow
	for rows.Next() {
		var u UserRow
		if rows.Scan(&u.ID, &u.Username, &u.Role, &u.LastTokenAt) == nil {
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
		WHERE a.id IN (
			SELECT a2.id FROM atividades a2
			LEFT JOIN users u2 ON u2.id=a2.usuario_id
			LEFT JOIN atividade_enderecos e2 ON e2.atividade_id=a2.id
			WHERE %s
			ORDER BY %s %s
			LIMIT $%d
		)
		ORDER BY %s %s`, strings.Join(where, " AND "), sort, order, len(args), sort, order)
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
	rows, err := a.pg.QueryContext(ctx, `SELECT id, atividade_id, seqproduto, empresa, rua_lida, predio_lido, rua_esperada, predio_esperado, status, reposicao, estoque, data_entrada, desccompleta FROM produto_verificacao WHERE atividade_id=$1 ORDER BY id`, id)
	if err != nil {
		return Activity{}, nil, err
	}
	defer rows.Close()
	var items []ProductVerification
	for rows.Next() {
		var p ProductVerification
		if rows.Scan(&p.ID, &p.AtividadeID, &p.SeqProduto, &p.Empresa, &p.RuaLida, &p.PredioLido, &p.RuaEsperada, &p.PredioEsperado, &p.Status, &p.Reposicao, &p.Estoque, &p.DataEntrada, &p.DescCompleta) == nil {
			// Query Oracle for DescCompleta, MDV, etc.
			var desc sql.NullString
			var estq int
			var mdv sql.NullFloat64
			var dtaUltVenda sql.NullTime
			errOra := a.ora.QueryRowContext(ctx, `
				SELECT mp.DESCCOMPLETA, mpe.ESTQLOJA, mpe.MEDVDIAGERAL, mpe.DTAULTVENDA
				FROM CONSINCO.MRL_PRODUTOEMPRESA mpe
				LEFT JOIN CONSINCO.MAP_PRODUTO mp ON mp.SEQPRODUTO=mpe.SEQPRODUTO
				WHERE mpe.NROEMPRESA=:1 AND mpe.SEQPRODUTO=:2
			`, p.Empresa, p.SeqProduto).Scan(&desc, &estq, &mdv, &dtaUltVenda)

			if errOra == nil {
				p.DescCompleta = desc
				p.Estoque = estq
				p.MDV = mdv
				if dtaUltVenda.Valid && mdv.Valid && mdv.Float64 > 0 {
					p.DDV = sql.NullFloat64{Float64: float64(estq) / mdv.Float64, Valid: true}
				}
			}

			// Count reincidencia — how many times this product was RUPTURA in other activities
			a.pg.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM produto_verificacao WHERE seqproduto=$1 AND empresa=$2 AND status='RUPTURA' AND atividade_id!=$3`,
				p.SeqProduto, p.Empresa, id).Scan(&p.Reincidencia)

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

func (a *App) findProductsByDescription(ctx context.Context, descricao string, empresa, seqlocal int) ([]OracleProduct, error) {
	// Split input into words and build pattern: "COCA LT 350" → "%COCA%LT%350%"
	words := strings.Fields(descricao)
	var parts []string
	for _, w := range words {
		if w != "" {
			parts = append(parts, strings.ToUpper(w))
		}
	}
	pattern := "%" + strings.Join(parts, "%") + "%"
	rows, err := a.ora.QueryContext(ctx, `
		SELECT mpe.SEQPRODUTO, mpe.NROEMPRESA, mpe.DTAULTENTRADA, mpe.DTAULTVENDA, mpe.ESTQLOJA, mpe.MEDVDIAGERAL,
		       mp.DESCCOMPLETA, mrlp.NRORUA, mrlp.NROPREDIO,
		       (SELECT MAX(marca) FROM CONSINCO.ETLV_PRODUTO WHERE SEQPRODUTO = mpe.SEQPRODUTO) AS MARCA,
		       CONSINCO.fBuscaPrecoAtualPdv(mpe.SEQPRODUTO, 1, mpe.NROEMPRESA) AS PRECO_VENDA,
		       (SELECT LISTAGG(CODACESSO, '|') WITHIN GROUP (ORDER BY CODACESSO) FROM CONSINCO.MAP_PRODCODIGO WHERE SEQPRODUTO = mpe.SEQPRODUTO) AS CODIGOS
		FROM CONSINCO.MRL_PRODUTOEMPRESA mpe
		LEFT JOIN CONSINCO.MAP_PRODUTO mp ON mp.SEQPRODUTO=mpe.SEQPRODUTO
		LEFT JOIN CONSINCO.MRL_PRODLOCAL mrlp ON mrlp.SEQPRODUTO=mpe.SEQPRODUTO AND mrlp.NROEMPRESA=mpe.NROEMPRESA AND mrlp.SEQLOCAL=:1
		WHERE mpe.NROEMPRESA=:2 AND UPPER(mp.DESCCOMPLETA) LIKE :3 AND ROWNUM <= 20
		ORDER BY mp.DESCCOMPLETA`, seqlocal, empresa, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var products []OracleProduct
	for rows.Next() {
		var p OracleProduct
		if err := rows.Scan(&p.SEQPRODUTO, &p.NROEMPRESA, &p.DTAULTENTRADA, &p.DTAULTVENDA, &p.ESTQLOJA, &p.MEDVDIAGERAL, &p.DESCCOMPLETA, &p.NRORUA, &p.NROPREDIO, &p.MARCA, &p.PRECO_VENDA, &p.CODIGOS); err == nil {
			if p.MEDVDIAGERAL.Valid && p.MEDVDIAGERAL.Float64 > 0 {
				p.DiasEstoque = sql.NullFloat64{Float64: float64(p.ESTQLOJA) / p.MEDVDIAGERAL.Float64, Valid: true}
			}
			products = append(products, p)
		}
	}
	return products, nil
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
	if err == nil && p.MEDVDIAGERAL.Valid && p.MEDVDIAGERAL.Float64 > 0 {
		p.DiasEstoque = sql.NullFloat64{Float64: float64(p.ESTQLOJA) / p.MEDVDIAGERAL.Float64, Valid: true}
	}
	return p, err
}
