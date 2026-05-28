package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vertkit/vertkit/internal/storage/memory"
)

func TestTenantScopedRoutesRequireServiceToken(t *testing.T) {
	server := NewServer(memory.NewStores(), WithServiceToken("test-token"))

	req := httptest.NewRequest(http.MethodGet, "/tenants/acme/accounts", nil)
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestCannotCreateAccountForMissingTenant(t *testing.T) {
	server := NewServer(memory.NewStores(), WithServiceToken("test-token"))

	rec := requestJSON(t, server, http.MethodPost, "/tenants/missing/accounts", map[string]any{
		"id":   "acc_001",
		"name": "Acme",
	})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestAccountCRUD(t *testing.T) {
	server := NewServer(memory.NewStores(), WithServiceToken("test-token"))

	rec := requestJSON(t, server, http.MethodPost, "/tenants", map[string]any{
		"id":               "acme",
		"name":             "Acme Global",
		"default_currency": "USD",
		"default_locale":   "en-US",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create tenant status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	rec = requestJSON(t, server, http.MethodPost, "/tenants/acme/accounts", map[string]any{
		"id":            "acc_001",
		"name":          "Acme",
		"industry":      "Software",
		"custom_fields": map[string]any{"segment": "enterprise"},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create account status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	rec = requestJSON(t, server, http.MethodGet, "/tenants/acme/accounts/acc_001", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get account status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	rec = requestJSON(t, server, http.MethodPut, "/tenants/acme/accounts/acc_001", map[string]any{
		"name":          "Acme Updated",
		"status":        "active",
		"custom_fields": map[string]any{"segment": "midmarket"},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("update account status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	rec = requestJSON(t, server, http.MethodDelete, "/tenants/acme/accounts/acc_001", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete account status = %d, want %d; body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}

	rec = requestJSON(t, server, http.MethodGet, "/tenants/acme/accounts/acc_001", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get deleted account status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func requestJSON(t *testing.T, server *Server, method, path string, payload any) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode request: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	return rec
}
