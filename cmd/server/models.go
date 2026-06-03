package main

import (
	"database/sql"
	"html/template"
	"time"
)

type Config struct {
	Port          string
	AppEnv        string
	SessionSecret []byte
	PostgresURL   string
	OracleURL     string
}

type App struct {
	cfg Config
	pg  *sql.DB
	ora *OracleReader
	tpl *template.Template
}

type OracleReader struct {
	db *sql.DB
}

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Role         string
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
	SEQPRODUTO    int             `json:"SEQPRODUTO"`
	CODACESSO     sql.NullString  `json:"CODACESSO,omitempty"`
	NRORUA        sql.NullString  `json:"NRORUA,omitempty"`
	NROPREDIO     sql.NullString  `json:"NROPREDIO,omitempty"`
	DESCCOMPLETA  sql.NullString  `json:"DESCCOMPLETA,omitempty"`
	NROEMPRESA    int             `json:"NROEMPRESA,omitempty"`
	DTAULTENTRADA sql.NullTime    `json:"DTAULTENTRADA,omitempty"`
	DTAULTVENDA   sql.NullTime    `json:"DTAULTVENDA,omitempty"`
	ESTQLOJA      int             `json:"ESTQLOJA,omitempty"`
	MEDVDIAGERAL  sql.NullFloat64 `json:"MEDVDIAGERAL,omitempty"`
	MARCA         sql.NullString  `json:"MARCA,omitempty"`
	PRECO_VENDA   sql.NullFloat64 `json:"PRECO_VENDA,omitempty"`
	CODIGOS       sql.NullString  `json:"CODIGOS,omitempty"`
}
type finalizeReq struct {
	Empresa      any      `json:"empresa"`
	SeqLocal     int      `json:"seqlocal"`
	Rua          string   `json:"rua"`
	Predio       []string `json:"predio"`
	ReadProducts []struct {
		SeqProduto int    `json:"seqproduto"`
		EAN        string `json:"ean"`
		Rua        string `json:"rua"`
		Predio     string `json:"predio"`
		Status     string `json:"status"`
		Reposicao  bool   `json:"reposicao"`
	} `json:"readProducts"`
	ExpectedProducts []struct {
		SeqProduto int `json:"seqproduto"`
	} `json:"expectedProducts"`
}
