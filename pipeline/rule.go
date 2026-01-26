package pipeline

import (
	"fmt"

	"github.com/dolphin-sistemas/computations-engine/core"
	"github.com/dolphin-sistemas/computations-engine/operators"
)

// ExecuteActions é uma referência para actions.ExecuteActions para evitar import circular
var ExecuteActions func(ctx *core.EngineContext, actions []core.Action) ([]core.Reason, []core.Violation, error)

// RunRule avalia a condition de uma regra e executa as actions se verdadeira
func RunRule(ctx *core.EngineContext, rule core.Rule) ([]core.Reason, []core.Violation, error) {
	// Se não tem condition, sempre executa
	if rule.Condition == nil || len(rule.Condition) == 0 {
		if ExecuteActions == nil {
			return nil, nil, fmt.Errorf("ExecuteActions not initialized")
		}
		reasons, violations, err := ExecuteActions(ctx, rule.Actions)
		if err != nil {
			return nil, nil, err
		}
		for i := range reasons {
			reasons[i].RuleID = rule.ID
			reasons[i].Phase = rule.Phase
		}
		return reasons, violations, nil
	}

	// Avaliar condition com JsonLogic
	evalData := BuildEvaluationData(ctx)
	shouldExecute, err := operators.EvaluateJsonLogic(rule.Condition, evalData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to evaluate condition for rule %s: %w", rule.ID, err)
	}

	// Se condition retornou true, executar actions
	if shouldExecuteBool, ok := shouldExecute.(bool); ok && shouldExecuteBool {
		if ExecuteActions == nil {
			return nil, nil, fmt.Errorf("ExecuteActions not initialized")
		}
		reasons, violations, err := ExecuteActions(ctx, rule.Actions)
		if err != nil {
			return nil, nil, err
		}
		for i := range reasons {
			reasons[i].RuleID = rule.ID
			reasons[i].Phase = rule.Phase
		}
		return reasons, violations, nil
	}

	return nil, nil, nil
}

// BuildEvaluationData monta o contexto de dados para avaliação JsonLogic
func BuildEvaluationData(ctx *core.EngineContext) map[string]interface{} {
	data := make(map[string]interface{})
	state := ctx.State

	// Context
	data["context"] = map[string]interface{}{
		"tenantId": ctx.Context.TenantID,
		"userId":   ctx.Context.UserID,
		"locale":   ctx.Context.Locale,
	}

	// Fields do estado
	for k, v := range state.Fields {
		data[k] = v
	}

	// Totals
	if state.Totals != (core.Totals{}) {
		data["totals"] = map[string]interface{}{
			"subtotal": state.Totals.Subtotal,
			"discount": state.Totals.Discount,
			"tax":      state.Totals.Tax,
			"total":    state.Totals.Total,
		}
	}

	// Items - criar array para acesso por índice
	itemsData := make([]map[string]interface{}, len(state.Items))
	for i, item := range state.Items {
		itemData := map[string]interface{}{
			"id":     item.ID,
			"amount": item.Amount,
		}
		// Fields do item
		for k, v := range item.Fields {
			itemData[k] = v
		}
		itemsData[i] = itemData
	}
	data["items"] = itemsData

	// Helper: itemValues (array de valores dos itens) para facilitar sum(itemValues)
	// Procura por campo "value" ou "total" nos fields do item
	itemValues := make([]float64, len(state.Items))
	for i, item := range state.Items {
		var value float64
		// Tentar encontrar valor em fields["value"] ou fields["total"]
		if v, ok := item.Fields["value"]; ok {
			value = toFloat64(v)
		} else if v, ok := item.Fields["total"]; ok {
			value = toFloat64(v)
		} else {
			// Fallback para amount
			value = item.Amount
		}
		itemValues[i] = value
	}
	data["itemValues"] = itemValues

	return data
}

// toFloat64 converte interface{} para float64
func toFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case int32:
		return float64(n)
	default:
		return 0.0
	}
}
