package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	engine "github.com/dolphin-sistemas/computations-engine"
	"github.com/dolphin-sistemas/computations-engine/core"
)

func main() {
	// Exemplo 1: Criar RulePack diretamente
	rulePack := core.RulePack{
		ID:      "example-rules",
		Version: "v1.0.0",
		Phases: []core.RulePhase{
			{
				Name: "baseline",
				Rules: []core.Rule{
					{
						ID:        "calc-subtotal",
						Phase:     "baseline",
						Priority:  1,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "compute",
								Target: "totals.subtotal",
								Logic: map[string]interface{}{
									"sum": []interface{}{
										map[string]interface{}{
											"var": []interface{}{"itemValues", []interface{}{}},
										},
									},
								},
							},
						},
					},
					{
						ID:        "calc-total",
						Phase:     "baseline",
						Priority:  2,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "compute",
								Target: "totals.total",
								Logic: map[string]interface{}{
									"var": []interface{}{"totals.subtotal", 0},
								},
							},
						},
					},
				},
			},
		},
	}

	// Exemplo 2: Criar um estado
	state := core.State{
		TenantID: "test-tenant",
		Items: []core.Item{
			{
				ID:     "item-1",
				Amount: 2,
				Fields: map[string]interface{}{
					"value": 100.0,
				},
			},
			{
				ID:     "item-2",
				Amount: 1,
				Fields: map[string]interface{}{
					"value": 50.0,
				},
			},
		},
		Fields: make(map[string]interface{}),
		Totals: core.Totals{},
	}

	// Exemplo 3: Executar motor de regras
	result, err := engine.RunEngine(
		context.Background(),
		state,
		rulePack,
		core.ContextMeta{
			TenantID: "test-tenant",
			UserID:   "user-1",
			Locale:   "pt-BR",
		},
	)
	if err != nil {
		log.Fatalf("Engine execution failed: %v", err)
	}

	// Exemplo 4: Exibir resultados
	fmt.Printf("Rules Version: %s\n", result.RulesVersion)
	fmt.Printf("Reasons: %d\n", len(result.Reasons))
	fmt.Printf("Violations: %d\n", len(result.Violations))

	stateJSON, _ := json.MarshalIndent(result.StateFragment, "", "  ")
	fmt.Printf("State Fragment:\n%s\n", stateJSON)

	deltaJSON, _ := json.MarshalIndent(result.ServerDelta, "", "  ")
	fmt.Printf("Server Delta:\n%s\n", deltaJSON)
}
