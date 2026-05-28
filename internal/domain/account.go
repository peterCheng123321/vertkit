package domain

import (
	"errors"
	"strings"
	"time"
)

// AccountID is the unique identifier for a customer account (company/organization).
type AccountID string

// Account represents a customer organization (B2B) or sometimes a large individual customer.
// This is one of the central entities in both CRM and ERP contexts.
//
// Global considerations built in:
// - BillingAddress and ShippingAddress use the global Address type.
// - Currency preference per account (overrides tenant default for quotes/orders).
type Account struct {
	ID              AccountID      `json:"id"`
	TenantID        TenantID       `json:"tenant_id"`
	Name            string         `json:"name"`
	Website         string         `json:"website,omitempty"`
	Industry        string         `json:"industry,omitempty"`
	Employees       int            `json:"employees,omitempty"`
	AnnualRevenue   *Money         `json:"annual_revenue,omitempty"` // pointer so it can be null
	BillingAddress  Address        `json:"billing_address"`
	ShippingAddress Address        `json:"shipping_address,omitempty"`
	Phone           string         `json:"phone,omitempty"`
	Email           string         `json:"email,omitempty"`
	Status          string         `json:"status"` // e.g. "active", "churned", "prospect"
	OwnerID         string         `json:"owner_id,omitempty"`
	CustomFields    map[string]any `json:"custom_fields,omitempty"` // Extension point for verticals
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// NewAccount performs minimal validation. More sophisticated rules belong in services or policies.
func NewAccount(id AccountID, tenantID TenantID, name string) (Account, error) {
	if id == "" {
		return Account{}, errors.New("account id is required")
	}
	if tenantID == "" {
		return Account{}, errors.New("tenant_id is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Account{}, errors.New("account name is required")
	}

	now := time.Now().UTC()
	return Account{
		ID:        id,
		TenantID:  tenantID,
		Name:      name,
		Status:    "prospect",
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
