package actions

import (
	"fmt"
	"strings"

	"github.com/dolphin-sistemas/computations-engine/core"
	"github.com/dolphin-sistemas/computations-engine/operators"
)

// ExecuteComputeAction executa ação "compute": calcula usando JsonLogic e define em target
func ExecuteComputeAction(ctx *core.EngineContext, action core.Action, evalData map[string]interface{}) (*core.Reason, *core.Violation, error) {
	if action.Target == "" {
		return nil, nil, fmt.Errorf("compute action requires target")
	}

	if action.Logic == nil || len(action.Logic) == 0 {
		return nil, nil, fmt.Errorf("compute action requires logic")
	}

	// Se o target é items[*].fields.x, processar cada item individualmente
	if strings.HasPrefix(action.Target, "items[*].") {
		return executeComputeActionForAllItems(ctx, action, evalData)
	}

	// Avaliar JsonLogic uma vez para targets não-iterativos
	result, err := operators.EvaluateJsonLogic(action.Logic, evalData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to evaluate compute logic: %w", err)
	}

	// Aplicar no target
	if err := SetValue(ctx.State, action.Target, result); err != nil {
		return nil, nil, err
	}

	return &core.Reason{Message: fmt.Sprintf("computed %s = %v", action.Target, result)}, nil, nil
}

// executeComputeActionForAllItems executa compute action para cada item individualmente
func executeComputeActionForAllItems(ctx *core.EngineContext, action core.Action, evalData map[string]interface{}) (*core.Reason, *core.Violation, error) {
	// Extrair o nome do campo: items[*].fields.x -> x
	parts := strings.Split(action.Target, ".")
	if len(parts) != 3 || !strings.HasPrefix(parts[0], "items[") || !strings.HasSuffix(parts[0], "]") || parts[1] != "fields" {
		return nil, nil, fmt.Errorf("invalid items[*] target format: %s", action.Target)
	}

	// Verificar se é items[*] e extrair o campo
	itemsPart := parts[0] // "items[*]"
	if itemsPart != "items[*]" {
		return nil, nil, fmt.Errorf("invalid items[*] target format: %s (expected items[*], got %s)", action.Target, itemsPart)
	}

	fieldName := parts[2] // "x"

	// Para cada item, criar contexto específico e avaliar
	for i := range ctx.State.Items {
		// Criar contexto de avaliação para este item específico
		itemEvalData := make(map[string]interface{})

		// Copiar dados globais (context, fields do estado, totals)
		for k, v := range evalData {
			if k != "items" && k != "itemValues" {
				itemEvalData[k] = v
			}
		}

		// Adicionar dados do item atual ao contexto
		item := ctx.State.Items[i]
		itemEvalData["id"] = item.ID
		itemEvalData["amount"] = item.Amount

		// Adicionar todos os campos do item diretamente no contexto
		for k, v := range item.Fields {
			itemEvalData[k] = v
		}

		// Avaliar JsonLogic com contexto deste item
		result, err := operators.EvaluateJsonLogic(action.Logic, itemEvalData)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to evaluate compute logic for item %d: %w", i, err)
		}

		// Aplicar resultado no campo do item
		if ctx.State.Items[i].Fields == nil {
			ctx.State.Items[i].Fields = make(map[string]interface{})
		}
		ctx.State.Items[i].Fields[fieldName] = result
	}

	return &core.Reason{Message: fmt.Sprintf("computed %s for %d items", action.Target, len(ctx.State.Items))}, nil, nil
}
