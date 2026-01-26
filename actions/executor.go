package actions

import (
	"fmt"

	"github.com/dolphin-sistemas/computations-engine/core"
	"github.com/dolphin-sistemas/computations-engine/pipeline"
)

// ExecuteActions executa uma lista de ações sobre o State
func ExecuteActions(ctx *core.EngineContext, actions []core.Action) ([]core.Reason, []core.Violation, error) {
	var reasons []core.Reason
	var violations []core.Violation

	for _, action := range actions {
		reason, violation, err := ExecuteAction(ctx, action)
		if err != nil {
			return nil, nil, fmt.Errorf("error executing action %s: %w", action.Type, err)
		}

		if reason != nil {
			reasons = append(reasons, *reason)
		}
		if violation != nil {
			violations = append(violations, *violation)
		}
	}

	return reasons, violations, nil
}

// ExecuteAction executa uma única ação
func ExecuteAction(ctx *core.EngineContext, action core.Action) (*core.Reason, *core.Violation, error) {
	evalData := pipeline.BuildEvaluationData(ctx)

	switch action.Type {
	case "set":
		return ExecuteSetAction(ctx, action, evalData)
	case "compute":
		return ExecuteComputeAction(ctx, action, evalData)
	case "validate":
		return ExecuteValidateAction(ctx, action, evalData)
	case "add":
		return ExecuteAddAction(ctx, action, evalData)
	case "multiply":
		return ExecuteMultiplyAction(ctx, action, evalData)
	default:
		return nil, nil, fmt.Errorf("unknown action type: %s", action.Type)
	}
}
