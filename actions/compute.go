package actions

import (
	"fmt"

	"github.com/dolphin-sistemas/computations-engine/core"
	"github.com/dolphin-sistemas/computations-engine/operators"
)

// ExecuteComputeAction executa aÃ§Ã£o "compute": calcula usando JsonLogic e define em target.
// Supports nested paths: items[*].x, items[*].negotiations[*].percent, items[*].foo[*].bar[*].baz, etc.
func ExecuteComputeAction(ctx *core.EngineContext, action core.Action, evalData map[string]interface{}) (*core.Reason, *core.Violation, error) {
	if action.Target == "" {
		return nil, nil, fmt.Errorf("compute action requires target")
	}

	if action.Logic == nil || len(action.Logic) == 0 {
		return nil, nil, fmt.Errorf("compute action requires logic")
	}

	steps, err := ParsePath(action.Target)
	if err != nil {
		return nil, nil, err
	}

	if HasWildcard(steps) {
		return executeComputeActionIterative(ctx, action, evalData, steps)
	}

	// Non-iterative: evaluate once and set
	result, err := operators.EvaluateJsonLogic(action.Logic, evalData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to evaluate compute logic: %w", err)
	}
	if err := SetValue(ctx.State, action.Target, result); err != nil {
		return nil, nil, err
	}
	return &core.Reason{Message: fmt.Sprintf("computed %s = %v", action.Target, result)}, nil, nil
}

// executeComputeActionIterative iterates over all wildcard matches and evaluates logic per-element.
func executeComputeActionIterative(ctx *core.EngineContext, action core.Action, evalData map[string]interface{}, steps []PathStep) (*core.Reason, *core.Violation, error) {
	count := 0
	_, err := visitLeaves(ctx.State, steps, true, func(ref leafRef, selections []selectedValue) error {
		itemEvalData := buildEvalDataForSelections(evalData, selections)
		result, err := operators.EvaluateJsonLogic(action.Logic, itemEvalData)
		if err != nil {
			return fmt.Errorf("failed to evaluate compute logic: %w", err)
		}
		if err := ref.Set(result); err != nil {
			return err
		}
		count++
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if count == 0 {
		return &core.Reason{Message: fmt.Sprintf("computed %s (no elements)", action.Target)}, nil, nil
	}

	return &core.Reason{Message: fmt.Sprintf("computed %s for %d elements", action.Target, count)}, nil, nil
}

func buildEvalDataForSelections(base map[string]interface{}, selections []selectedValue) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range base {
		if k != "items" && k != "itemValues" {
			out[k] = v
		}
	}

	for _, sel := range selections {
		switch v := sel.Value.(type) {
		case *core.Item:
			out["id"] = v.ID
			out["amount"] = v.Amount
			for k2, v2 := range v.Fields {
				out[k2] = v2
			}
		case map[string]interface{}:
			for k2, v2 := range v {
				out[k2] = v2
			}
		}
	}

	return out
}
