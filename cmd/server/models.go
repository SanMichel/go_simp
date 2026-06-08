package main

import (
	"database/sql"
	"html/template"
	"time"
)

type Config struct {
	Port            string
	AppEnv          string
	SessionSecret   []byte
	PostgresURL     string
	OracleURL       string
	SessionTTL      time.Duration
	PGMaxConns      int
	OracleMaxConns  int
	OracleIdleConns int
	OracleIdleTime  time.Duration
}

type App struct {
	cfg          Config
	pg           *sql.DB
	ora          *OracleReader
	tpl          *template.Template
	loginLimiter *rateLimiter
}

type OracleReader struct {
	db *sql.DB
}

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Role         string
	LastTokenAt  sql.NullTime
}

type UserRow struct {
	ID          int
	Username    string
	Role        string
	LastTokenAt sql.NullTime
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
	MDV            sql.NullFloat64
	DDV            sql.NullFloat64
	Reincidencia   int
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
	SEQPRODUTO    int             `json:"seqproduto"`
	CODACESSO     sql.NullString  `json:"codacesso,omitempty"`
	NRORUA        sql.NullString  `json:"nrorua,omitempty"`
	NROPREDIO     sql.NullString  `json:"nropredio,omitempty"`
	DESCCOMPLETA  sql.NullString  `json:"desccompleta,omitempty"`
	NROEMPRESA    int             `json:"nroempresa,omitempty"`
	DTAULTENTRADA sql.NullTime    `json:"dtaUltEntrada,omitempty"`
	DTAULTVENDA   sql.NullTime    `json:"dtaUltVenda,omitempty"`
	ESTQLOJA      int             `json:"estoque,omitempty"`
	MEDVDIAGERAL  sql.NullFloat64 `json:"mdv,omitempty"`
	MARCA         sql.NullString  `json:"marca,omitempty"`
	PRECO_VENDA   sql.NullFloat64 `json:"precoVenda,omitempty"`
	CODIGOS       sql.NullString  `json:"codigos,omitempty"`
	DiasEstoque   sql.NullFloat64 `json:"diasEstoque,omitempty"`
}
type finalizeReq struct {
	Empresa      int      `json:"empresa"`
	SeqLocal     int      `json:"seqlocal"`
	Rua          string   `json:"rua"`
	Predio       []string `json:"predio"`
	ReadProducts []struct {
		SeqProduto   int    `json:"seqproduto"`
		EAN          string `json:"ean"`
		Rua          string `json:"rua"`
		Predio       string `json:"predio"`
		Status       string `json:"status"`
		Reposicao    bool   `json:"reposicao"`
		Desccompleta string `json:"desccompleta"`
	} `json:"readProducts"`
	ExpectedProducts []struct {
		SeqProduto int `json:"seqproduto"`
	} `json:"expectedProducts"`
}
