package actions

import (
	"fmt"

	"github.com/dolphin-sistemas/engine/core"
	"github.com/dolphin-sistemas/engine/operators"
)

// ExecuteValidateAction executa ação "validate": valida condição e cria violação se falsa
func ExecuteValidateAction(ctx *core.EngineContext, action core.Action, evalData map[string]interface{}) (*core.Reason, *core.Violation, error) {
	if action.Logic == nil || len(action.Logic) == 0 {
		return nil, nil, fmt.Errorf("validate action requires logic")
	}

	// Avaliar JsonLogic - se retornar true, significa violação
	violated, err := operators.EvaluateJsonLogic(action.Logic, evalData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to evaluate validate logic: %w", err)
	}

	if violatedBool, ok := violated.(bool); ok && violatedBool {
		field, _ := action.Params["field"].(string)
		code, _ := action.Params["code"].(string)
		message, _ := action.Params["message"].(string)

		if field == "" || code == "" {
			return nil, nil, fmt.Errorf("validate action requires field and code in params")
		}

		return nil, &core.Violation{
			Field:   field,
			Code:    code,
			Message: message,
		}, nil
	}

	return nil, nil, nil
}
