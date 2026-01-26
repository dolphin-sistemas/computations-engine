package guards

import (
	"fmt"

	"github.com/dolphin-sistemas/engine/core"
	"github.com/dolphin-sistemas/engine/operators"
)

// ValidateMaxRatio valida se uma razão não excede um máximo (genérico)
func ValidateMaxRatio(state *core.State, numerator, denominator string, maxPercent float64) error {
	num, ok1 := getNumericValue(state, numerator)
	den, ok2 := getNumericValue(state, denominator)
	
	if !ok1 || !ok2 {
		return fmt.Errorf("cannot validate ratio: fields not found")
	}
	
	if den <= 0 {
		return nil // Não pode calcular razão
	}
	
	ratio := (num / den) * 100
	if ratio > maxPercent {
		return fmt.Errorf("ratio exceeds maximum of %.2f%%", maxPercent)
	}
	
	return nil
}

// ValidateRequiredIDs valida se há IDs obrigatórios nos items
func ValidateRequiredIDs(state *core.State, requiredIDs []string, idField string) error {
	if len(requiredIDs) == 0 {
		return nil
	}

	idMap := make(map[string]bool)
	for _, item := range state.Items {
		var id string
		if idField == "id" {
			id = item.ID
		} else if v, ok := item.Fields[idField].(string); ok {
			id = v
		}
		if id != "" {
			idMap[id] = true
		}
	}

	for _, requiredID := range requiredIDs {
		if !idMap[requiredID] {
			return fmt.Errorf("required ID %s is missing", requiredID)
		}
	}

	return nil
}

// getNumericValue obtém valor numérico de um campo do estado
func getNumericValue(state *core.State, fieldPath string) (float64, bool) {
	// Tentar totals primeiro
	if fieldPath == "totals.discount" {
		return state.Totals.Discount, true
	}
	if fieldPath == "totals.subtotal" {
		return state.Totals.Subtotal, true
	}
	if fieldPath == "totals.tax" {
		return state.Totals.Tax, true
	}
	if fieldPath == "totals.total" {
		return state.Totals.Total, true
	}
	
	// Tentar fields
	if v, ok := state.Fields[fieldPath]; ok {
		switch n := v.(type) {
		case float64:
			return n, true
		case float32:
			return float64(n), true
		case int:
			return float64(n), true
		case int64:
			return float64(n), true
		}
	}
	
	return 0, false
}

// ValidateMinValue valida valor mínimo
func ValidateMinValue(value float64, minValue float64, fieldName string) error {
	if value < minValue {
		return fmt.Errorf("%s must be at least %.2f", fieldName, minValue)
	}
	return nil
}

// ValidateMaxValue valida valor máximo
func ValidateMaxValue(value float64, maxValue float64, fieldName string) error {
	if value > maxValue {
		return fmt.Errorf("%s must be at most %.2f", fieldName, maxValue)
	}
	return nil
}

// EvaluateGuardCondition avalia uma condição de guard usando JsonLogic
func EvaluateGuardCondition(logic map[string]interface{}, ctx *core.EngineContext) (bool, error) {
	// Usar BuildEvaluationData do pipeline
	evalData := buildEvaluationDataForGuard(ctx)
	result, err := operators.EvaluateJsonLogic(logic, evalData)
	if err != nil {
		return false, err
	}

	if boolResult, ok := result.(bool); ok {
		return boolResult, nil
	}

	// Truthy check
	return result != nil && result != false && result != 0 && result != "", nil
}

// buildEvaluationDataForGuard monta contexto de dados para guards (similar ao pipeline)
func buildEvaluationDataForGuard(ctx *core.EngineContext) map[string]interface{} {
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

	// Items
	itemsData := make([]map[string]interface{}, len(state.Items))
	for i, item := range state.Items {
		itemData := map[string]interface{}{
			"id":     item.ID,
			"amount": item.Amount,
		}
		for k, v := range item.Fields {
			itemData[k] = v
		}
		itemsData[i] = itemData
	}
	data["items"] = itemsData

	return data
}
