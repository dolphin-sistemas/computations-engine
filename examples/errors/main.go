package main

import (
	"context"
	"encoding/json"
	"fmt"

	engine "github.com/dolphin-sistemas/computations-engine"
	"github.com/dolphin-sistemas/computations-engine/core"
)

func main() {
	fmt.Println("=== Exemplos de Tratamento de Erros ===")
	fmt.Println()

	// Exemplo 1: RulePack sem ID
	fmt.Println("1️⃣  Erro: RulePack sem ID")
	testMissingRulePackID()

	// Exemplo 2: JsonLogic inválido
	fmt.Println("\n2️⃣  Erro: Operador JsonLogic inválido")
	testInvalidJsonLogic()

	// Exemplo 3: Validate sem params
	fmt.Println("\n3️⃣  Erro: Ação validate sem params obrigatórios")
	testValidateMissingParams()

	// Exemplo 4: Validate sem logic
	fmt.Println("\n4️⃣  Erro: Ação validate sem logic")
	testValidateMissingLogic()

	// Exemplo 5: Condição inválida
	fmt.Println("\n5️⃣  Erro: Condição com JsonLogic inválido")
	testInvalidCondition()

	// Exemplo 6: Divisão por zero
	fmt.Println("\n6️⃣  Erro: Divisão por zero")
	testDivisionByZero()
}

func testMissingRulePackID() {
	state := core.State{
		TenantID: "test-tenant",
		Items:    []core.Item{},
		Fields:   make(map[string]interface{}),
		Totals:   core.Totals{},
	}

	rulePack := core.RulePack{
		ID:      "", // ❌ ID vazio
		Version: "v1.0.0",
		Phases:  []core.RulePhase{},
	}

	result, err := engine.RunEngine(
		context.Background(),
		state,
		rulePack,
		core.ContextMeta{TenantID: "test-tenant"},
	)

	handleError("RulePack sem ID", err, result)
}

func testInvalidJsonLogic() {
	state := core.State{
		TenantID: "test-tenant",
		Items:    []core.Item{},
		Fields:   make(map[string]interface{}),
		Totals:   core.Totals{},
	}

	rulePack := core.RulePack{
		ID:      "error-test",
		Version: "v1.0.0",
		Phases: []core.RulePhase{
			{
				Name: "baseline",
				Rules: []core.Rule{
					{
						ID:        "invalid-logic",
						Phase:     "baseline",
						Priority:  1,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "compute",
								Target: "fields.result",
								Logic: map[string]interface{}{
									"operador_inexistente_xyz": []interface{}{1, 2, 3}, // ❌ Operador inválido
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := engine.RunEngine(
		context.Background(),
		state,
		rulePack,
		core.ContextMeta{TenantID: "test-tenant"},
	)

	handleError("JsonLogic inválido", err, result)
}

func testValidateMissingParams() {
	state := core.State{
		TenantID: "test-tenant",
		Items:    []core.Item{},
		Fields:   make(map[string]interface{}),
		Totals:   core.Totals{},
	}

	rulePack := core.RulePack{
		ID:      "error-test",
		Version: "v1.0.0",
		Phases: []core.RulePhase{
			{
				Name: "guards",
				Rules: []core.Rule{
					{
						ID:        "validate-missing-params",
						Phase:     "guards",
						Priority:  1,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "validate",
								Target: "",
								Logic: map[string]interface{}{
									"==": []interface{}{map[string]interface{}{"var": "test"}, nil},
								},
								Params: make(map[string]interface{}), // ❌ Params vazio
							},
						},
					},
				},
			},
		},
	}

	result, err := engine.RunEngine(
		context.Background(),
		state,
		rulePack,
		core.ContextMeta{TenantID: "test-tenant"},
	)

	handleError("Validate sem params", err, result)
}

func testValidateMissingLogic() {
	state := core.State{
		TenantID: "test-tenant",
		Items:    []core.Item{},
		Fields:   make(map[string]interface{}),
		Totals:   core.Totals{},
	}

	rulePack := core.RulePack{
		ID:      "error-test",
		Version: "v1.0.0",
		Phases: []core.RulePhase{
			{
				Name: "guards",
				Rules: []core.Rule{
					{
						ID:        "validate-missing-logic",
						Phase:     "guards",
						Priority:  1,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "validate",
								Target: "",
								// ❌ Logic ausente
								Params: map[string]interface{}{
									"field":   "fields.test",
									"code":    "TEST",
									"message": "Test error",
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := engine.RunEngine(
		context.Background(),
		state,
		rulePack,
		core.ContextMeta{TenantID: "test-tenant"},
	)

	handleError("Validate sem logic", err, result)
}

func testInvalidCondition() {
	state := core.State{
		TenantID: "test-tenant",
		Items:    []core.Item{},
		Fields:   make(map[string]interface{}),
		Totals:   core.Totals{},
	}

	rulePack := core.RulePack{
		ID:      "error-test",
		Version: "v1.0.0",
		Phases: []core.RulePhase{
			{
				Name: "baseline",
				Rules: []core.Rule{
					{
						ID:       "invalid-condition",
						Phase:    "baseline",
						Priority: 1,
						Enabled:  true,
						Condition: map[string]interface{}{
							"operador_inexistente_xyz": []interface{}{1, 2}, // ❌ Condição inválida
						},
						Actions: []core.Action{
							{
								Type:   "set",
								Target: "fields.test",
								Value:  1,
							},
						},
					},
				},
			},
		},
	}

	result, err := engine.RunEngine(
		context.Background(),
		state,
		rulePack,
		core.ContextMeta{TenantID: "test-tenant"},
	)

	handleError("Condição inválida", err, result)
}

func testDivisionByZero() {
	state := core.State{
		TenantID: "test-tenant",
		Items:    []core.Item{},
		Fields: map[string]interface{}{
			"value": 100.0,
		},
		Totals: core.Totals{},
	}

	rulePack := core.RulePack{
		ID:      "error-test",
		Version: "v1.0.0",
		Phases: []core.RulePhase{
			{
				Name: "baseline",
				Rules: []core.Rule{
					{
						ID:        "divide-by-zero",
						Phase:     "baseline",
						Priority:  1,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "compute",
								Target: "fields.result",
								Logic: map[string]interface{}{
									"/": []interface{}{
										map[string]interface{}{"var": "value"},
										0, // ❌ Divisão por zero
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := engine.RunEngine(
		context.Background(),
		state,
		rulePack,
		core.ContextMeta{TenantID: "test-tenant"},
	)

	handleError("Divisão por zero", err, result)
}

// handleError trata e exibe erros de forma formatada
func handleError(scenario string, err error, result *core.RunEngineResult) {
	if err != nil {
		fmt.Printf("  ❌ Erro capturado em '%s':\n", scenario)
		fmt.Printf("     %v\n", err)

		// Tentar exibir como JSON
		errorJSON := map[string]interface{}{
			"scenario": scenario,
			"error":    err.Error(),
		}

		jsonBytes, jsonErr := json.MarshalIndent(errorJSON, "     ", "  ")
		if jsonErr == nil {
			fmt.Println("     JSON:")
			fmt.Println(string(jsonBytes))
		}
	} else {
		fmt.Printf("  ⚠️  Não retornou erro (inesperado para '%s')\n", scenario)
		if result != nil {
			fmt.Printf("     Result: %+v\n", result)
		}
	}
}
