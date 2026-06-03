package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	go_ora "github.com/sijms/go-ora/v2"
)

func loadConfig() Config {
	port := getenv("PORT", "3000")
	secret := getenv("SESSION_SECRET", "dev-secret-change-me")
	pgURL := getenv("POSTGRES_URL", os.Getenv("DATABASE_URL"))
	if pgURL == "" {
		log.Fatal("POSTGRES_URL or DATABASE_URL required")
	}

	oracleURL := os.Getenv("ORACLE_URL")
	if oracleURL == "" {
		portNum, _ := strconv.Atoi(getenv("ORACLE_PORT", "1521"))
		oracleURL = go_ora.BuildUrl(
			getenv("ORACLE_HOST", "localhost"),
			portNum,
			getenv("ORACLE_SERVICE", "xe"),
			os.Getenv("ORACLE_USER"),
			os.Getenv("ORACLE_PASSWORD"),
			map[string]string{"TIMEOUT": "30", "client charset": "UTF8"},
		)
	}

	return Config{
		Port:          port,
		AppEnv:        getenv("APP_ENV", "development"),
		SessionSecret: []byte(secret),
		PostgresURL:   pgURL,
		OracleURL:     oracleURL,
	}
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for lineNo, raw := range strings.Split(string(b), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("%s:%d: expected KEY=VALUE", path, lineNo+1)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return fmt.Errorf("%s:%d: empty key", path, lineNo+1)
		}
		if len(value) >= 2 {
			quote := value[0]
			if (quote == '"' || quote == '\'') && value[len(value)-1] == quote {
				value = value[1 : len(value)-1]
			}
		}
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return nil
}
func (a *App) log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &logWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, lw.status, time.Since(start).Truncate(time.Millisecond))
	})
}

type logWriter struct {
	http.ResponseWriter
	status int
}

func (w *logWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func validRole(role string) bool {
	return role == "sysadmin" || role == "gerente" || role == "conferente"
}
func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func intQuery(r *http.Request, name string) int {
	v, _ := strconv.Atoi(r.URL.Query().Get(name))
	return v
}
func parseFilters(r *http.Request) ActivityFilters {
	q := r.URL.Query()
	return ActivityFilters{
		Username:     q["filter_username[]"],
		Empresa:      q["filter_empresa[]"],
		Rua:          q["filter_rua[]"],
		Predio:       q["filter_predio[]"],
		Impresso:     q["filter_impresso[]"],
		ID:           q["filter_id[]"],
		DataFimStart: q.Get("filter_dataFimStart"),
		DataFimEnd:   q.Get("filter_dataFimEnd"),
		Sort:         q.Get("sort"),
		Order:        q.Get("order"),
	}
}
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
