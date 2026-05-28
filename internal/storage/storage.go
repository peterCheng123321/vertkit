package storage

import (
	"context"
	"time"

	"github.com/vertkit/vertkit/internal/domain"
)

// TenantStore defines persistence operations for tenants.
type TenantStore interface {
	Create(ctx context.Context, t domain.Tenant) error
	Get(ctx context.Context, id domain.TenantID) (*domain.Tenant, error)
	List(ctx context.Context) ([]*domain.Tenant, error)
	Update(ctx context.Context, t domain.Tenant) error
	Delete(ctx context.Context, id domain.TenantID) error
}

// AccountStore defines multi-tenant persistence for accounts.
type AccountStore interface {
	Create(ctx context.Context, tenantID domain.TenantID, a domain.Account) error
	Get(ctx context.Context, tenantID domain.TenantID, id domain.AccountID) (*domain.Account, error)
	List(ctx context.Context, tenantID domain.TenantID) ([]*domain.Account, error)
	Update(ctx context.Context, tenantID domain.TenantID, a domain.Account) error
	Delete(ctx context.Context, tenantID domain.TenantID, id domain.AccountID) error
}

// ContactStore defines multi-tenant persistence for contacts.
type ContactStore interface {
	Create(ctx context.Context, tenantID domain.TenantID, c domain.Contact) error
	Get(ctx context.Context, tenantID domain.TenantID, id domain.ContactID) (*domain.Contact, error)
	List(ctx context.Context, tenantID domain.TenantID) ([]*domain.Contact, error)
	ListByAccount(ctx context.Context, tenantID domain.TenantID, accountID domain.AccountID) ([]*domain.Contact, error)
	Update(ctx context.Context, tenantID domain.TenantID, c domain.Contact) error
	Delete(ctx context.Context, tenantID domain.TenantID, id domain.ContactID) error
}

// OpportunityStore defines multi-tenant persistence for opportunities.
type OpportunityStore interface {
	Create(ctx context.Context, tenantID domain.TenantID, o domain.Opportunity) error
	Get(ctx context.Context, tenantID domain.TenantID, id domain.OpportunityID) (*domain.Opportunity, error)
	List(ctx context.Context, tenantID domain.TenantID) ([]*domain.Opportunity, error)
	ListByAccount(ctx context.Context, tenantID domain.TenantID, accountID domain.AccountID) ([]*domain.Opportunity, error)
	Update(ctx context.Context, tenantID domain.TenantID, o domain.Opportunity) error
	Delete(ctx context.Context, tenantID domain.TenantID, id domain.OpportunityID) error
}

// ProductStore defines multi-tenant persistence for products.
type ProductStore interface {
	Create(ctx context.Context, tenantID domain.TenantID, p domain.Product) error
	Get(ctx context.Context, tenantID domain.TenantID, id domain.ProductID) (*domain.Product, error)
	List(ctx context.Context, tenantID domain.TenantID) ([]*domain.Product, error)
	ListActive(ctx context.Context, tenantID domain.TenantID) ([]*domain.Product, error)
	Update(ctx context.Context, tenantID domain.TenantID, p domain.Product) error
	Delete(ctx context.Context, tenantID domain.TenantID, id domain.ProductID) error
}

// Rule represents a business rule definition (stored separately from domain).
// We keep the struct here so storage can reference it without circular imports.
type Rule struct {
	ID          string
	TenantID    domain.TenantID
	Name        string
	Description string
	EntityType  string          // "opportunity", "account", etc.
	Conditions  []RuleCondition `json:"conditions"`
	Actions     []RuleAction    `json:"actions"`
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RuleCondition struct {
	Field    string `json:"field"`    // e.g. "amount", "stage", "probability", "custom_fields.deal_type"
	Operator string `json:"operator"` // eq, neq, gt, gte, lt, lte, contains, in
	Value    any    `json:"value"`
}

type RuleAction struct {
	Type   string         `json:"type"` // "block", "warn", "set_field", "add_note"
	Params map[string]any `json:"params"`
}

// RuleStore defines multi-tenant persistence for business rules.
type RuleStore interface {
	Create(ctx context.Context, tenantID domain.TenantID, r Rule) error
	Get(ctx context.Context, tenantID domain.TenantID, id string) (*Rule, error)
	List(ctx context.Context, tenantID domain.TenantID) ([]*Rule, error)
	ListActive(ctx context.Context, tenantID domain.TenantID, entityType string) ([]*Rule, error)
	Update(ctx context.Context, tenantID domain.TenantID, r Rule) error
	Delete(ctx context.Context, tenantID domain.TenantID, id string) error
}

// Stores aggregates all stores for convenience when wiring the API layer.
type Stores struct {
	Tenants       TenantStore
	Accounts      AccountStore
	Contacts      ContactStore
	Opportunities OpportunityStore
	Products      ProductStore
	Rules         RuleStore
}
