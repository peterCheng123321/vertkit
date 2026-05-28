package postgres

import (
	"os"
	"strings"
	"testing"
)

func TestSchemaDefinesTenantScopedTablesAndRLS(t *testing.T) {
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	text := string(schema)
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS tenants",
		"CREATE TABLE IF NOT EXISTS accounts",
		"CREATE TABLE IF NOT EXISTS contacts",
		"CREATE TABLE IF NOT EXISTS opportunities",
		"CREATE TABLE IF NOT EXISTS products",
		"CREATE TABLE IF NOT EXISTS rules",
		"PRIMARY KEY (tenant_id, id)",
		"ENABLE ROW LEVEL SECURITY",
		"current_setting('vertkit.tenant_id'",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("schema missing %q", want)
		}
	}
}
