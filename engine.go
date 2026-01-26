package engine

import (
	"context"
	"fmt"

	"github.com/dolphin-sistemas/computations-engine/actions"
	"github.com/dolphin-sistemas/computations-engine/core"
	"github.com/dolphin-sistemas/computations-engine/diff"
	"github.com/dolphin-sistemas/computations-engine/pipeline"
)

// RunEngine é a função principal pública do motor de regras
// Executa o pipeline completo e retorna os resultados
func RunEngine(ctx context.Context, state core.State, rules core.RulePack, contextMeta core.ContextMeta) (
	stateFragment map[string]interface{},
	serverDelta map[string]interface{},
	reasons []core.Reason,
	violations []core.Violation,
	rulesVersion string,
	err error,
) {
	// Inicializar referência de actions no pipeline (evitar import circular)
	pipeline.ExecuteActions = actions.ExecuteActions

	// Criar contexto do motor
	engineCtx, err := core.NewEngineContext(state, contextMeta)
	if err != nil {
		return nil, nil, nil, nil, "", fmt.Errorf("failed to create engine context: %w", err)
	}

	// Validar RulePack
	if rules.ID == "" {
		return nil, nil, nil, nil, "", fmt.Errorf("rulePack.id is required")
	}

	// Executar pipeline
	if err := pipeline.RunPipeline(engineCtx, rules); err != nil {
		return nil, nil, nil, nil, "", fmt.Errorf("pipeline execution failed: %w", err)
	}

	// Gerar outputs
	stateFragment = diff.BuildStateFragment(engineCtx)
	serverDelta = diff.BuildServerDelta(engineCtx)
	reasons = engineCtx.Reasons
	violations = engineCtx.Violations
	rulesVersion = rules.Version

	return stateFragment, serverDelta, reasons, violations, rulesVersion, nil
}
