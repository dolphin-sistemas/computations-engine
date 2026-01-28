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
func RunEngine(ctx context.Context, state core.State, rules core.RulePack, contextMeta core.ContextMeta) (*core.RunEngineResult, error) {
	// Inicializar referência de actions no pipeline (evitar import circular)
	pipeline.ExecuteActions = actions.ExecuteActions

	// Criar contexto do motor
	engineCtx, err := core.NewEngineContext(state, contextMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to create engine context: %w", err)
	}

	// Validar RulePack
	if rules.ID == "" {
		return nil, fmt.Errorf("rulePack.id is required")
	}

	// Executar pipeline
	if err := pipeline.RunPipeline(engineCtx, rules); err != nil {
		return nil, fmt.Errorf("pipeline execution failed: %w", err)
	}

	// Gerar outputs
	return &core.RunEngineResult{
		StateFragment: diff.BuildStateFragment(engineCtx),
		ServerDelta:   diff.BuildServerDelta(engineCtx),
		Reasons:       engineCtx.Reasons,
		Violations:    engineCtx.Violations,
		RulesVersion:  rules.Version,
	}, nil
}
