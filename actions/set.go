package actions

import (
	"fmt"

	"github.com/dolphin-sistemas/engine/core"
)

// ExecuteSetAction executa ação "set": define um valor literal em target
func ExecuteSetAction(ctx *core.EngineContext, action core.Action, evalData map[string]interface{}) (*core.Reason, *core.Violation, error) {
	if action.Target == "" {
		return nil, nil, fmt.Errorf("set action requires target")
	}

	value := action.Value
	if err := SetValue(ctx.State, action.Target, value); err != nil {
		return nil, nil, err
	}

	return &core.Reason{Message: fmt.Sprintf("set %s = %v", action.Target, value)}, nil, nil
}
