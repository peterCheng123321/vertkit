package rules

import (
	"time"

	"github.com/vertkit/vertkit/internal/domain"
)

// Rule is a tenant-scoped business rule that can be evaluated against entities.
// Designed to be simple enough for agents to generate and reason about, while
// powerful enough for real CRM/ERP business logic (approvals, gates, alerts, etc.).
type Rule struct {
	ID          string
	TenantID    domain.TenantID
	Name        string
	Description string
	EntityType  string // "opportunity", "account", "contact", etc.
	Conditions  []Condition
	Actions     []Action
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Condition represents a single predicate. All conditions in a rule are ANDed.
type Condition struct {
	// Field path into the entity. Examples:
	//   "amount", "stage", "probability", "custom_fields.deal_size"
	//   "billing_address.country_code"
	Field    string `json:"field"`
	Operator string `json:"operator"` // eq | neq | gt | gte | lt | lte | contains | in
	Value    any    `json:"value"`
}

// Action describes what should happen when the rule matches.
type Action struct {
	// Supported types in v0:
	//   "block"        - prevent the operation (with message)
	//   "warn"         - return a warning but allow
	//   "set_field"    - suggest or apply a field change
	//   "require_field" - entity must have this custom field populated
	Type   string         `json:"type"`
	Params map[string]any `json:"params"`
}

// EvaluationContext provides the data the engine evaluates against.
type EvaluationContext struct {
	Entity     any // the domain entity (Opportunity, Account, etc.)
	EntityType string
	TenantID   domain.TenantID
	Operation  string // "create", "update", "stage_change", etc.
	Now        time.Time
}

// EvaluationResult captures everything that happened during rule evaluation.
type EvaluationResult struct {
	RuleID      string
	RuleName    string
	Matched     bool
	Errors      []string // blocking issues
	Warnings    []string
	Suggestions []map[string]any
}
