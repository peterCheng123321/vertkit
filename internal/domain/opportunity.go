package domain

import (
	"errors"
	"strings"
	"time"
)

// OpportunityID is the unique identifier for a sales opportunity / deal.
type OpportunityID string

// Opportunity represents a potential sale or project. This bridges pure CRM (pipeline)
// and ERP (will eventually become Quote → Order).
//
// Global: Amount is Money. CloseDate, Stage, and Probability are core to forecasting
// across different countries and sales cultures.
type Opportunity struct {
	ID              OpportunityID  `json:"id"`
	TenantID        TenantID       `json:"tenant_id"`
	AccountID       AccountID      `json:"account_id"`
	ContactID       ContactID      `json:"contact_id,omitempty"`
	Name            string         `json:"name"`
	Description     string         `json:"description,omitempty"`
	Amount          Money          `json:"amount"`      // Expected or committed value
	Currency        Currency       `json:"currency"`    // Explicit for clarity even if same as Amount
	Stage           string         `json:"stage"`       // e.g. "qualification", "proposal", "negotiation", "closed-won"
	Probability     int            `json:"probability"` // 0-100
	CloseDate       *time.Time     `json:"close_date,omitempty"`
	ExpectedRevenue *Money         `json:"expected_revenue,omitempty"` // often Amount * Probability / 100
	Source          string         `json:"source,omitempty"`
	OwnerID         string         `json:"owner_id,omitempty"`
	CustomFields    map[string]any `json:"custom_fields,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// NewOpportunity creates a validated opportunity.
func NewOpportunity(id OpportunityID, tenantID TenantID, accountID AccountID, name string, amount Money) (Opportunity, error) {
	if id == "" {
		return Opportunity{}, errors.New("opportunity id is required")
	}
	if tenantID == "" {
		return Opportunity{}, errors.New("tenant_id is required")
	}
	if accountID == "" {
		return Opportunity{}, errors.New("account_id is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Opportunity{}, errors.New("name is required")
	}
	if amount.IsZero() {
		return Opportunity{}, errors.New("amount must be greater than zero for a new opportunity")
	}

	now := time.Now().UTC()
	return Opportunity{
		ID:          id,
		TenantID:    tenantID,
		AccountID:   accountID,
		Name:        name,
		Amount:      amount,
		Currency:    amount.Currency,
		Stage:       "qualification",
		Probability: 10,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
