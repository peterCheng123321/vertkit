package memory

import (
	"context"
	"testing"

	"github.com/vertkit/vertkit/internal/domain"
)

func TestAccountIDsAreScopedPerTenant(t *testing.T) {
	ctx := context.Background()
	stores := NewStores()

	for _, tenantID := range []domain.TenantID{"tenant_a", "tenant_b"} {
		tenant, err := domain.NewTenant(tenantID, string(tenantID), domain.CurrencyUSD, "en-US")
		if err != nil {
			t.Fatalf("new tenant: %v", err)
		}
		if err := stores.Tenants.Create(ctx, tenant); err != nil {
			t.Fatalf("create tenant %s: %v", tenantID, err)
		}
		account, err := domain.NewAccount("acc_001", tenantID, "Acme")
		if err != nil {
			t.Fatalf("new account: %v", err)
		}
		if err := stores.Accounts.Create(ctx, tenantID, account); err != nil {
			t.Fatalf("create account for %s: %v", tenantID, err)
		}
	}

	for _, tenantID := range []domain.TenantID{"tenant_a", "tenant_b"} {
		account, err := stores.Accounts.Get(ctx, tenantID, "acc_001")
		if err != nil {
			t.Fatalf("get account for %s: %v", tenantID, err)
		}
		if account.TenantID != tenantID {
			t.Fatalf("expected account tenant %s, got %s", tenantID, account.TenantID)
		}
	}
}

func TestReturnedAccountsCannotMutateStoredCustomFields(t *testing.T) {
	ctx := context.Background()
	stores := NewStores()
	tenantID := domain.TenantID("tenant_a")
	account, err := domain.NewAccount("acc_001", tenantID, "Acme")
	if err != nil {
		t.Fatalf("new account: %v", err)
	}
	account.CustomFields = map[string]any{"tier": "gold"}

	if err := stores.Accounts.Create(ctx, tenantID, account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	got, err := stores.Accounts.Get(ctx, tenantID, "acc_001")
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	got.CustomFields["tier"] = "platinum"

	again, err := stores.Accounts.Get(ctx, tenantID, "acc_001")
	if err != nil {
		t.Fatalf("get account again: %v", err)
	}
	if again.CustomFields["tier"] != "gold" {
		t.Fatalf("stored custom field was mutated, got %v", again.CustomFields["tier"])
	}
}
