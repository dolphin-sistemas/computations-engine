package actions

import (
	"fmt"

	"github.com/dolphin-sistemas/engine/core"
	"github.com/dolphin-sistemas/engine/operators"
	"github.com/dolphin-sistemas/engine/pkg"
)

// ExecuteAddAction executa ação "add": incrementa valor existente
func ExecuteAddAction(ctx *core.EngineContext, action core.Action, evalData map[string]interface{}) (*core.Reason, *core.Violation, error) {
	if action.Target == "" {
		return nil, nil, fmt.Errorf("add action requires target")
	}

	// Obter valor atual
	currentValue, err := GetValue(ctx.State, action.Target)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current value for target %s: %w", action.Target, err)
	}

	// Calcular incremento
	var increment float64
	if action.Logic != nil && len(action.Logic) > 0 {
		// Incremento calculado via JsonLogic
		result, err := operators.EvaluateJsonLogic(action.Logic, evalData)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to evaluate add logic: %w", err)
		}
		increment = pkg.ToFloat64(result)
	} else if action.Value != nil {
		// Incremento literal
		increment = pkg.ToFloat64(action.Value)
	} else {
		return nil, nil, fmt.Errorf("add action requires either logic or value")
	}

	// Somar valores
	newValue := pkg.ToFloat64(currentValue) + increment

	// Aplicar novo valor
	if err := SetValue(ctx.State, action.Target, newValue); err != nil {
		return nil, nil, err
	}

	return &core.Reason{Message: fmt.Sprintf("added %v to %s (result: %v)", increment, action.Target, newValue)}, nil, nil
}
