package pipeline

import (
	"fmt"
	"sort"

	"github.com/dolphin-sistemas/computations-engine/core"
)

// RunPhase executa todas as regras de uma fase em ordem de prioridade
func RunPhase(ctx *core.EngineContext, phase core.RulePhase) error {
	// Ordenar regras por prioridade (menor = primeiro)
	rules := make([]core.Rule, len(phase.Rules))
	copy(rules, phase.Rules)

	sort.Slice(rules, func(i, j int) bool {
		// Se prioridade não especificada, assume 0
		priI := rules[i].Priority
		priJ := rules[j].Priority
		return priI < priJ
	})

	// Executar cada regra
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		ruleReasons, ruleViolations, err := RunRule(ctx, rule)
		if err != nil {
			return fmt.Errorf("error executing rule %s: %w", rule.ID, err)
		}

		if len(ruleReasons) > 0 {
			ctx.Reasons = append(ctx.Reasons, ruleReasons...)
		}
		if len(ruleViolations) > 0 {
			ctx.Violations = append(ctx.Violations, ruleViolations...)
		}
	}

	return nil
}
