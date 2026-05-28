package domain

import (
	"errors"
	"strings"
	"time"
)

// ProductID is the unique identifier for a sellable item or service.
type ProductID string

// Product represents a catalog item. This is the bridge between CRM (what we sell in opportunities)
// and ERP (inventory, pricing, fulfillment).
//
// Global design:
// - Price is Money (never raw float).
// - Supports simple recurring vs one-time distinction for global SaaS / subscription businesses.
type Product struct {
	ID                ProductID      `json:"id"`
	TenantID          TenantID       `json:"tenant_id"`
	SKU               string         `json:"sku,omitempty"`
	Name              string         `json:"name"`
	Description       string         `json:"description,omitempty"`
	Price             Money          `json:"price"`
	Unit              string         `json:"unit,omitempty"` // e.g. "each", "month", "user", "license"
	IsRecurring       bool           `json:"is_recurring"`
	RecurringInterval string         `json:"recurring_interval,omitempty"` // "monthly", "annual"
	Category          string         `json:"category,omitempty"`
	Active            bool           `json:"active"`
	CustomFields      map[string]any `json:"custom_fields,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// NewProduct creates a product with validation.
func NewProduct(id ProductID, tenantID TenantID, name string, price Money) (Product, error) {
	if id == "" {
		return Product{}, errors.New("product id is required")
	}
	if tenantID == "" {
		return Product{}, errors.New("tenant_id is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Product{}, errors.New("name is required")
	}
	if price.IsZero() {
		return Product{}, errors.New("price must be greater than zero")
	}

	now := time.Now().UTC()
	return Product{
		ID:        id,
		TenantID:  tenantID,
		Name:      name,
		Price:     price,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
