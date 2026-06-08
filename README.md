# SIMP — Sistema Integrado de Monitoramento de Prateleiras

App de conferência de prateleiras, focado em atividades de leitura de produtos por EAN (código de barras), descoberta de rupturas de estoque e divergências de reposição em lojas de varejo.

Conferentes registram produtos lidos em determinados endereços (rua/predio). O sistema compara com o estoque esperado (Oracle) e sinaliza:
- **OK** — produto presente no local esperado
- **RUPTURA** — produto esperado mas não lido (contém estoque e não está resposto na prateleira)
- **DIVERGENTE** — produto lido em local diferente do esperado
- **Reposição** — produto com estoque baixo ou nulo, aguardando reposição

Gerentes e admins acompanham atividades via dashboard com filtros e relatórios impressos.

## Stack

- **Linguagem:** Go (1.23)
- **HTTP:** `net/http` padrão (Go 1.22+ routing)
- **Templates:** `html/template` com HTMX (server-rendered, sem SPA)
- **Postgres:** `pgx` via `database/sql` — dados da aplicação (usuários, atividades, verificações)
- **Oracle:** `go-ora/v2` — consultas **read-only** a banco legado (produtos, empresas, locais)
- **Frontend:** HTMX + Go templates + CSS vanilla (sem framework JS)
- **Migrações:** Auto-migração na inicialização (`CREATE TABLE IF NOT EXISTS`)
- **Hot reload:** [Air](https://github.com/air-verse/air) (`.air.toml`)

## Estrutura

```
go-simp/
├── cmd/server/                    # Pacote único main
│   ├── main.go                    # Entrypoint, routes, go:embed
│   ├── models.go                  # Structs (User, Activity, ProductVerification, etc.)
│   ├── handlers.go                # Handlers de página e API mistos
│   ├── api_handlers.go            # Handlers JSON das APIs REST
│   ├── auth.go                    # Sessão HMAC, RBAC, CSRF
│   ├── db.go                      # Conexões Postgres/Oracle, queries, migrações
│   ├── utils.go                   # Config, .env loader, rate limiter, helpers
│   ├── main_test.go               # Testes
│   └── templates/                 # Templates HTML + assets (embedded via go:embed)
│       ├── style.css
│       ├── admin.css
│       ├── *.html                 # Páginas (login, home, atividades, dashboard, admin, print)
│       ├── *.js                   # JS (app, dashboard, admin, login, shared, htmx.min)
│       └── components/            # Parciais HTMX
│           ├── head.html
│           ├── nav.html
│           ├── activities_table.html
│           ├── activity_modal.html
│           ├── user_row.html
│           ├── user_edit_row.html
│           └── users_section.html
├── tmp/                           # Ref. da arquitetura antiga — NÃO MODIFICAR
├── bin/                           # Build artifacts (gitignored)
├── .tmp/                          # Build artifacts (gitignored)
├── .air.toml
├── .env.example
└── go.mod
```

## Arquitetura

### Conexões

- **Postgres** (`pgx`): dados locais da aplicação — usuários, sessões, atividades, produto_verificação. Esquema auto-migrado na inicialização.
- **Oracle** (`go-ora`, **read-only**): consulta em tempo real ao banco legado do cliente — empresas, locais, produtos, estoque, preços, médias de venda. Toda query passa por `isReadOnlySQL()` que só permite `SELECT`/`WITH`.

### Sessão e Segurança

- Token HMAC customizado (não JWT) armazenado em cookie HttpOnly
- CSRF via cookie + header `X-CSRF-Token` para endpoints POST/PATCH
- Rate limiter de login por IP (5 tentativas/minuto)
- Roles: `conferente`, `gerente`, `sysadmin`

### Banco Postgres

| Tabela | Finalidade |
|--------|-----------|
| `users` | Usuários e hash de senha |
| `atividades` | Registro de atividade de conferência |
| `atividade_enderecos` | Endereços (rua/predio) por atividade |
| `produto_verificacao` | Produtos lidos/esperados com status |

### Consultas Oracle

Lookups em tabelas `CONSINCO.*` para dados mestres de produto, empresa e local — via `OracleReader` que wrappa `database/sql` com proteção read-only.

## Quick start

```bash
cp .env.example .env   # EDITAR .env — POSTGRES_URL é obrigatório
go mod tidy
go run ./cmd/server    # ou: air (hot reload)
```

Na primeira execução, migra tabelas e cria `admin` como `sysadmin` com senha aleatória (exibida no log).

## Testes

```bash
go test ./cmd/server
```

Testes em `main_test.go`. Sem dependências externas ou framework.
