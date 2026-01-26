package pipeline

import (
	"fmt"

	"github.com/dolphin-sistemas/engine/core"
)

// PhaseOrder define a ordem das fases do pipeline
var PhaseOrder = []string{
	"baseline",
	"allocation",
	"taxes",
	"totals",
	"guards",
}

// RunPipeline executa o pipeline completo de fases
func RunPipeline(ctx *core.EngineContext, rulePack core.RulePack) error {
	// Criar mapa de fases por nome para acesso rápido
	phaseMap := make(map[string]core.RulePhase)
	for _, phase := range rulePack.Phases {
		phaseMap[phase.Name] = phase
	}

	// Executar fases na ordem definida
	for phaseIndex, phaseName := range PhaseOrder {
		phase, exists := phaseMap[phaseName]
		if !exists {
			// Fase não existe no RulePack, pular
			continue
		}

		ctx.PhaseIndex = phaseIndex
		if err := RunPhase(ctx, phase); err != nil {
			return fmt.Errorf("error in phase %s: %w", phaseName, err)
		}
	}

	// Executar fases adicionais que não estão na ordem padrão
	for _, phase := range rulePack.Phases {
		found := false
		for _, standardPhase := range PhaseOrder {
			if phase.Name == standardPhase {
				found = true
				break
			}
		}
		if !found {
			// Fase customizada, executar no final
			if err := RunPhase(ctx, phase); err != nil {
				return fmt.Errorf("error in custom phase %s: %w", phase.Name, err)
			}
		}
	}

	return nil
}
