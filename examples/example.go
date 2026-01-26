package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/dolphin-sistemas/engine"
	"github.com/dolphin-sistemas/engine/core"
	"github.com/dolphin-sistemas/engine/loader"
)

func main() {
	// Exemplo 1: Carregar RulePack de arquivo JSON
	rulePack, err := loader.LoadRulePackFromFile("testdata/vectors/vector1_baseline.json")
	if err != nil {
		log.Fatalf("Failed to load rule pack: %v", err)
	}

	// Exemplo 2: Criar um estado
	state := core.State{
		TenantID: "test-tenant",
		Items: []core.Item{
			{
				ID:     "item-1",
				Amount: 2,
				Fields: map[string]interface{}{
					"basePrice": 2500.0,
				},
			},
		},
		Fields: map[string]interface{}{
			"customerType": "PF",
		},
		Totals: core.Totals{},
	}

	// Exemplo 3: Executar motor de regras
	stateFragment, serverDelta, reasons, violations, rulesVersion, err := engine.RunEngine(
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
	fmt.Printf("Rules Version: %s\n", rulesVersion)
	fmt.Printf("Reasons: %d\n", len(reasons))
	fmt.Printf("Violations: %d\n", len(violations))

	stateJSON, _ := json.MarshalIndent(stateFragment, "", "  ")
	fmt.Printf("State Fragment:\n%s\n", stateJSON)

	deltaJSON, _ := json.MarshalIndent(serverDelta, "", "  ")
	fmt.Printf("Server Delta:\n%s\n", deltaJSON)
}
