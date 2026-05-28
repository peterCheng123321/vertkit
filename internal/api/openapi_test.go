package api

import (
	"os"
	"strings"
	"testing"
)

func TestOpenAPIContractDocumentsFoundationRoutes(t *testing.T) {
	spec, err := os.ReadFile("../../openapi.yaml")
	if err != nil {
		t.Fatalf("read openapi.yaml: %v", err)
	}
	text := string(spec)
	for _, want := range []string{
		"openapi: 3.1.0",
		"bearerAuth:",
		"/tenants/{tenant_id}/accounts/{id}:",
		"put:",
		"delete:",
		"/tenants/{tenant_id}/rules/evaluate/opportunity:",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("OpenAPI contract missing %q", want)
		}
	}
}
