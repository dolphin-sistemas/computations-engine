package guards

import (
	"github.com/dolphin-sistemas/computations-engine/core"
	"github.com/dolphin-sistemas/computations-engine/pipeline"
)

// ExecuteGuards executa a fase de guards (validações finais)
func ExecuteGuards(ctx *core.EngineContext, phase core.RulePhase) error {
	// Guards são executados como uma fase normal do pipeline
	// A diferença é que violações aqui podem bloquear o processamento
	return pipeline.RunPhase(ctx, phase)
}
