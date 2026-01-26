package actions

import (
	"fmt"

	"github.com/dolphin-sistemas/engine/core"
	"github.com/dolphin-sistemas/engine/operators"
	"github.com/dolphin-sistemas/engine/pkg"
)

// ExecuteMultiplyAction executa ação "multiply": multiplica valor existente
func ExecuteMultiplyAction(ctx *core.EngineContext, action core.Action, evalData map[string]interface{}) (*core.Reason, *core.Violation, error) {
	if action.Target == "" {
		return nil, nil, fmt.Errorf("multiply action requires target")
	}

	// Obter valor atual
	currentValue, err := GetValue(ctx.State, action.Target)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current value for target %s: %w", action.Target, err)
	}

	// Calcular multiplicador
	var multiplier float64
	if action.Logic != nil && len(action.Logic) > 0 {
		// Multiplicador calculado via JsonLogic
		result, err := operators.EvaluateJsonLogic(action.Logic, evalData)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to evaluate multiply logic: %w", err)
		}
		multiplier = pkg.ToFloat64(result)
	} else if action.Value != nil {
		// Multiplicador literal
		multiplier = pkg.ToFloat64(action.Value)
	} else {
		return nil, nil, fmt.Errorf("multiply action requires either logic or value")
	}

	// Multiplicar valores
	newValue := pkg.ToFloat64(currentValue) * multiplier

	// Aplicar novo valor
	if err := SetValue(ctx.State, action.Target, newValue); err != nil {
		return nil, nil, err
	}

	return &core.Reason{Message: fmt.Sprintf("multiplied %s by %v (result: %v)", action.Target, multiplier, newValue)}, nil, nil
}
