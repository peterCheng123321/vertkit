package rules

import (
	"context"
	"testing"

	"github.com/vertkit/vertkit/internal/domain"
)

func TestEvaluateMatchesDocumentedJSONFieldPaths(t *testing.T) {
	amount, err := domain.NewMoney(12_000_000, domain.CurrencyUSD)
	if err != nil {
		t.Fatalf("new money: %v", err)
	}
	opportunity, err := domain.NewOpportunity("opp_001", "tenant_a", "acc_001", "Enterprise", amount)
	if err != nil {
		t.Fatalf("new opportunity: %v", err)
	}
	opportunity.CustomFields = map[string]any{"region": "EMEA"}

	engine := NewEngine()
	results := engine.Evaluate(context.Background(), []Rule{{
		ID:         "large_emea_deal",
		Name:       "Large EMEA deal",
		EntityType: "opportunity",
		IsActive:   true,
		Conditions: []Condition{
			{Field: "amount", Operator: "gte", Value: float64(10_000_000)},
			{Field: "currency", Operator: "eq", Value: "USD"},
			{Field: "custom_fields.region", Operator: "in", Value: []any{"EMEA", "EU", "UK"}},
		},
		Actions: []Action{{Type: "warn", Params: map[string]any{"message": "review required"}}},
	}}, EvaluationContext{
		Entity:     opportunity,
		EntityType: "opportunity",
		TenantID:   "tenant_a",
	})

	if len(results) != 1 {
		t.Fatalf("expected one result, got %d", len(results))
	}
	if !results[0].Matched {
		t.Fatalf("expected rule to match: %#v", results[0])
	}
	if got := len(results[0].Warnings); got != 1 {
		t.Fatalf("expected one warning, got %d", got)
	}
}
