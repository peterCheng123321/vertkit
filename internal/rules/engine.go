package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/vertkit/vertkit/internal/domain"
)

// Engine evaluates business rules against CRM/ERP entities.
// It is intentionally simple and side-effect free in v0.
// The caller decides what to do with the EvaluationResults (block save, show warnings, etc.).
type Engine struct {
	// In a real system this would be injected with a RuleStore.
	// For the first slice we pass rules in explicitly so the API layer owns loading.
}

// NewEngine creates a rules engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Evaluate runs the provided active rules against the context and returns results.
// Only rules whose EntityType matches are considered.
func (e *Engine) Evaluate(ctx context.Context, rules []Rule, ec EvaluationContext) []EvaluationResult {
	results := make([]EvaluationResult, 0, len(rules))

	for _, r := range rules {
		if !r.IsActive {
			continue
		}
		if r.EntityType != "" && r.EntityType != ec.EntityType {
			continue
		}

		res := EvaluationResult{
			RuleID:   r.ID,
			RuleName: r.Name,
		}

		matched := true
		for _, cond := range r.Conditions {
			if !e.evaluateCondition(cond, ec.Entity) {
				matched = false
				break
			}
		}
		res.Matched = matched

		if matched {
			for _, act := range r.Actions {
				e.applyAction(act, &res, ec)
			}
		}

		results = append(results, res)
	}

	return results
}

func (e *Engine) evaluateCondition(c Condition, entity any) bool {
	val := getFieldValue(entity, c.Field)
	if val == nil {
		return false // missing field never matches for safety
	}

	switch c.Operator {
	case "eq":
		return valuesEqual(val, c.Value)
	case "neq":
		return !valuesEqual(val, c.Value)
	case "gt", "gte", "lt", "lte":
		return compareNumbers(val, c.Value, c.Operator)
	case "contains":
		return contains(val, c.Value)
	case "in":
		return inList(val, c.Value)
	default:
		return false
	}
}

func (e *Engine) applyAction(a Action, res *EvaluationResult, ec EvaluationContext) {
	switch a.Type {
	case "block":
		msg := "Rule triggered: action blocked"
		if m, ok := a.Params["message"].(string); ok {
			msg = m
		}
		res.Errors = append(res.Errors, msg)

	case "warn":
		msg := "Rule warning"
		if m, ok := a.Params["message"].(string); ok {
			msg = m
		}
		res.Warnings = append(res.Warnings, msg)

	case "set_field":
		// For v0 we only record the suggestion. The caller decides whether to apply it.
		res.Suggestions = append(res.Suggestions, map[string]any{
			"field":  a.Params["field"],
			"value":  a.Params["value"],
			"reason": "rule:" + res.RuleName,
		})

	case "require_field":
		field := ""
		if f, ok := a.Params["field"].(string); ok {
			field = f
		}
		res.Warnings = append(res.Warnings, fmt.Sprintf("required field missing or rule-enforced: %s", field))
	}
}

// --- Value extraction and comparison helpers (kept minimal) ---

func getFieldValue(entity any, path string) any {
	if entity == nil {
		return nil
	}
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	parts := strings.Split(path, ".")
	current := v

	for _, part := range parts {
		if !current.IsValid() {
			return nil
		}
		if current.Kind() == reflect.Ptr {
			if current.IsNil() {
				return nil
			}
			current = current.Elem()
		}
		if current.Kind() == reflect.Interface {
			if current.IsNil() {
				return nil
			}
			current = current.Elem()
		}
		if current.Kind() == reflect.Map {
			key := reflect.ValueOf(part)
			if key.Type().AssignableTo(current.Type().Key()) {
				current = current.MapIndex(key)
			} else {
				return nil
			}
			continue
		}
		if current.Kind() != reflect.Struct {
			// If we have a Money-like struct and asking for "amount", drill in
			if current.Type().Name() == "Money" {
				if strings.EqualFold(part, "amount") {
					return current.FieldByName("Amount").Interface()
				}
			}
			return nil
		}
		current = fieldByExternalName(current, part)
	}

	if !current.IsValid() {
		return nil
	}
	return current.Interface()
}

func fieldByExternalName(v reflect.Value, name string) reflect.Value {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonName == "-" {
			continue
		}
		if jsonName == name || strings.EqualFold(field.Name, name) || normalizeName(field.Name) == normalizeName(name) {
			return v.Field(i)
		}
	}
	return reflect.Value{}
}

func normalizeName(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), "_", "")
}

func valuesEqual(a, b any) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}
	if af, aok := toFloat64(a); aok {
		if bf, bok := toFloat64(b); bok {
			return af == bf
		}
	}
	if as, aok := toString(a); aok {
		if bs, bok := toString(b); bok {
			return as == bs
		}
	}
	return false
}

func toString(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case domain.Currency:
		return string(x), true
	case fmt.Stringer:
		return x.String(), true
	default:
		return "", false
	}
}

func compareNumbers(a, b any, op string) bool {
	af, aok := toFloat64(a)
	bf, bok := toFloat64(b)
	if !aok || !bok {
		return false
	}
	switch op {
	case "gt":
		return af > bf
	case "gte":
		return af >= bf
	case "lt":
		return af < bf
	case "lte":
		return af <= bf
	}
	return false
}

func toFloat64(v any) (float64, bool) {
	switch x := v.(type) {
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case json.Number:
		f, err := strconv.ParseFloat(string(x), 64)
		return f, err == nil
	case domain.Money:
		return float64(x.Amount), true // compare on minor units
	case *domain.Money:
		if x != nil {
			return float64(x.Amount), true
		}
	}
	return 0, false
}

func contains(container, needle any) bool {
	switch c := container.(type) {
	case string:
		if s, ok := needle.(string); ok {
			return strings.Contains(strings.ToLower(c), strings.ToLower(s))
		}
	case []any:
		for _, item := range c {
			if reflect.DeepEqual(item, needle) {
				return true
			}
		}
	}
	return false
}

func inList(val, list any) bool {
	rv := reflect.ValueOf(list)
	if rv.Kind() != reflect.Slice {
		return false
	}
	for i := 0; i < rv.Len(); i++ {
		if valuesEqual(val, rv.Index(i).Interface()) {
			return true
		}
	}
	return false
}
