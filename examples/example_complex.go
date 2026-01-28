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
	fmt.Println("=== Exemplo Complexo: Pipeline Completo ===\n")

	// Criar RulePack complexo com todas as fases
	rulePack := core.RulePack{
		ID:      "complex-rules",
		Version: "v2.0.0",
		Phases: []core.RulePhase{
			// Fase 1: Baseline - Inicialização
			{
				Name: "baseline",
				Rules: []core.Rule{
					{
						ID:        "init-item-totals",
						Phase:     "baseline",
						Priority:  1,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "compute",
								Target: "items[*].fields.itemTotal",
								Logic: map[string]interface{}{
									"*": []interface{}{
										map[string]interface{}{"var": "basePrice"},
										map[string]interface{}{"var": "amount"},
									},
								},
							},
						},
					},
					{
						ID:        "init-subtotal",
						Phase:     "baseline",
						Priority:  2,
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
				},
			},
			// Fase 2: Allocation - Distribuição de descontos
			{
				Name: "allocation",
				Rules: []core.Rule{
					{
						ID:       "apply-customer-discount",
						Phase:    "allocation",
						Priority: 1,
						Enabled:  true,
						Condition: map[string]interface{}{
							"and": []interface{}{
								map[string]interface{}{">": []interface{}{
									map[string]interface{}{"var": "customerDiscountPercent"},
									0,
								}},
								map[string]interface{}{">": []interface{}{
									map[string]interface{}{"var": "totals.subtotal"},
									100,
								}},
							},
						},
						Actions: []core.Action{
							{
								Type:   "compute",
								Target: "totals.discount",
								Logic: map[string]interface{}{
									"round2": []interface{}{
										map[string]interface{}{
											"*": []interface{}{
												map[string]interface{}{"var": "totals.subtotal"},
												map[string]interface{}{
													"/": []interface{}{
														map[string]interface{}{"var": "customerDiscountPercent"},
														100,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			// Fase 3: Taxes - Cálculo de impostos
			{
				Name: "taxes",
				Rules: []core.Rule{
					{
						ID:       "calculate-tax",
						Phase:    "taxes",
						Priority: 1,
						Enabled:  true,
						Condition: map[string]interface{}{
							">": []interface{}{
								map[string]interface{}{"var": "taxRate"},
								0,
							},
						},
						Actions: []core.Action{
							{
								Type:   "compute",
								Target: "totals.tax",
								Logic: map[string]interface{}{
									"round2": []interface{}{
										map[string]interface{}{
											"*": []interface{}{
												map[string]interface{}{
													"-": []interface{}{
														map[string]interface{}{"var": "totals.subtotal"},
														map[string]interface{}{"var": "totals.discount"},
													},
												},
												map[string]interface{}{
													"/": []interface{}{
														map[string]interface{}{"var": "taxRate"},
														100,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			// Fase 4: Totals - Cálculo final
			{
				Name: "totals",
				Rules: []core.Rule{
					{
						ID:        "calculate-total",
						Phase:     "totals",
						Priority:  1,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "compute",
								Target: "totals.total",
								Logic: map[string]interface{}{
									"round2": []interface{}{
										map[string]interface{}{
											"+": []interface{}{
												map[string]interface{}{
													"-": []interface{}{
														map[string]interface{}{"var": "totals.subtotal"},
														map[string]interface{}{"var": "totals.discount"},
													},
												},
												map[string]interface{}{"var": "totals.tax"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			// Fase 5: Guards - Validações
			{
				Name: "guards",
				Rules: []core.Rule{
					{
						ID:       "validate-max-discount",
						Phase:    "guards",
						Priority: 1,
						Enabled:  true,
						Condition: map[string]interface{}{
							">": []interface{}{
								map[string]interface{}{"var": "totals.discount"},
								0,
							},
						},
						Actions: []core.Action{
							{
								Type:   "validate",
								Target: "",
								Logic: map[string]interface{}{
									">": []interface{}{
										map[string]interface{}{
											"/": []interface{}{
												map[string]interface{}{"var": "totals.discount"},
												map[string]interface{}{"var": "totals.subtotal"},
											},
										},
										0.3, // 30% máximo
									},
								},
								Params: map[string]interface{}{
									"field":   "totals.discount",
									"code":    "MAX_DISCOUNT_EXCEEDED",
									"message": "Desconto não pode exceder 30% do subtotal",
								},
							},
						},
					},
					{
						ID:        "validate-min-total",
						Phase:     "guards",
						Priority:  2,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "validate",
								Target: "",
								Logic: map[string]interface{}{
									"<": []interface{}{
										map[string]interface{}{"var": "totals.total"},
										0,
									},
								},
								Params: map[string]interface{}{
									"field":   "totals.total",
									"code":    "NEGATIVE_TOTAL",
									"message": "Total não pode ser negativo",
								},
							},
						},
					},
				},
			},
		},
	}

	// Criar estado complexo
	state := core.State{
		TenantID: "tenant-123",
		Items: []core.Item{
			{
				ID:     "item-1",
				Amount: 3,
				Fields: map[string]interface{}{
					"basePrice": 150.50,
					"value":     451.50, // Será calculado
				},
			},
			{
				ID:     "item-2",
				Amount: 2,
				Fields: map[string]interface{}{
					"basePrice": 200.00,
					"value":     400.00, // Será calculado
				},
			},
		},
		Fields: map[string]interface{}{
			"customerDiscountPercent": 15.0,
			"taxRate":                 10.0,
		},
		Totals: core.Totals{},
	}

	// Executar engine
	result, err := engine.RunEngine(
		context.Background(),
		state,
		rulePack,
		core.ContextMeta{
			TenantID: "tenant-123",
			UserID:   "user-456",
			Locale:   "pt-BR",
		},
	)

	if err != nil {
		log.Fatalf("❌ Erro na execução: %v", err)
	}

	// Exibir resultados
	fmt.Println("✅ Execução bem-sucedida!\n")
	fmt.Printf("📋 Versão das Regras: %s\n", result.RulesVersion)
	fmt.Printf("📊 Regras Executadas: %d\n", len(result.Reasons))
	fmt.Printf("⚠️  Violações: %d\n\n", len(result.Violations))

	// Mostrar reasons
	if len(result.Reasons) > 0 {
		fmt.Println("📝 Regras Executadas:")
		for _, reason := range result.Reasons {
			fmt.Printf("  - [%s] %s: %s\n", reason.Phase, reason.RuleID, reason.Message)
		}
		fmt.Println()
	}

	// Mostrar violations
	if len(result.Violations) > 0 {
		fmt.Println("❌ Violações Encontradas:")
		for _, violation := range result.Violations {
			fmt.Printf("  - Campo: %s\n", violation.Field)
			fmt.Printf("    Código: %s\n", violation.Code)
			fmt.Printf("    Mensagem: %s\n\n", violation.Message)
		}
	}

	// Mostrar state fragment
	fmt.Println("📦 State Fragment (campos que mudaram):")
	stateJSON, _ := json.MarshalIndent(result.StateFragment, "", "  ")
	fmt.Println(string(stateJSON))
	fmt.Println()

	// Mostrar server delta
	fmt.Println("🔄 Server Delta (diferenças para sincronização):")
	deltaJSON, _ := json.MarshalIndent(result.ServerDelta, "", "  ")
	fmt.Println(string(deltaJSON))
}
