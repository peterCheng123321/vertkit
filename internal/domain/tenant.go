package domain

import (
	"errors"
	"strings"
	"time"
)

// TenantID is the unique identifier for a tenant (customer organization using the system).
type TenantID string

// Tenant represents a customer of the platform (the "global" multi-tenant boundary).
// Every piece of business data belongs to exactly one tenant.
//
// Design notes for agent extensibility:
//   - Keep this struct small and obvious.
//   - All customization (custom fields, workflows, modules) should live in
//     extension mechanisms, not by bloating the core Tenant.
type Tenant struct {
	ID              TenantID  `json:"id"`
	Name            string    `json:"name"`
	DefaultCurrency Currency  `json:"default_currency"`
	DefaultLocale   string    `json:"default_locale"` // e.g. "en-US", "zh-CN", "pt-BR"
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// NewTenant creates a validated tenant.
func NewTenant(id TenantID, name string, defaultCurrency Currency, defaultLocale string) (Tenant, error) {
	if id == "" {
		return Tenant{}, errors.New("tenant id is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Tenant{}, errors.New("tenant name is required")
	}
	if !isValidCurrency(defaultCurrency) {
		return Tenant{}, errors.New("invalid default currency")
	}
	if defaultLocale == "" {
		defaultLocale = "en-US"
	}

	now := time.Now().UTC()
	return Tenant{
		ID:              id,
		Name:            name,
		DefaultCurrency: defaultCurrency,
		DefaultLocale:   defaultLocale,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}
