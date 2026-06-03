package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplatesParse(t *testing.T) {
	if parseTemplates() == nil {
		t.Fatal("parseTemplates returned nil")
	}
}

func TestOracleReadOnlySQLGuard(t *testing.T) {
	cases := map[string]bool{
		"SELECT * FROM dual": true,
		"\n\twith x as (select 1 from dual) select * from x": true,
		" INSERT INTO x VALUES (1)":                          false,
		"UPDATE x SET y=1":                                   false,
		"DELETE FROM x":                                      false,
		"":                                                   false,
	}
	for query, want := range cases {
		if got := isReadOnlySQL(query); got != want {
			t.Fatalf("isReadOnlySQL(%q)=%v want %v", query, got, want)
		}
	}
}

func TestLoadDotEnv(t *testing.T) {
	key := "GO_SIMP_TEST_ENV"
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("GO_SIMP_TEST_ENV=\"ok\"\n# ignored\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := loadDotEnv(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv(key); got != "ok" {
		t.Fatalf("env=%q want ok", got)
	}
}
