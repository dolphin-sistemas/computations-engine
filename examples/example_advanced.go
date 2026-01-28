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
	fmt.Println("=== Exemplo Avançado: Cálculos Complexos e Condicionais ===\n")

	// RulePack com cálculos avançados usando operadores customizados
	rulePack := core.RulePack{
		ID:      "advanced-rules",
		Version: "v3.0.0",
		Phases: []core.RulePhase{
			{
				Name: "baseline",
				Rules: []core.Rule{
					// Calcular valor de cada item
					{
						ID:        "calc-item-value",
						Phase:     "baseline",
						Priority:  1,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "compute",
								Target: "items[*].fields.value",
								Logic: map[string]interface{}{
									"*": []interface{}{
										map[string]interface{}{"var": "basePrice"},
										map[string]interface{}{"var": "amount"},
									},
								},
							},
						},
					},
					// Calcular subtotal usando sum
					{
						ID:        "calc-subtotal",
						Phase:     "baseline",
						Priority:  2,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "compute",
								Target: "totals.subtotal",
								Logic: map[string]interface{}{
									"round2": []interface{}{
										map[string]interface{}{
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
				},
			},
			{
				Name: "allocation",
				Rules: []core.Rule{
					// Desconto progressivo baseado no valor total
					{
						ID:        "progressive-discount",
						Phase:     "allocation",
						Priority:  1,
						Enabled:   true,
						Condition: map[string]interface{}{
							">": []interface{}{
								map[string]interface{}{"var": []interface{}{"totals.subtotal", 0}},
								500,
							},
						},
						Actions: []core.Action{
							{
								Type:   "compute",
								Target: "totals.discount",
								Logic: map[string]interface{}{
									"round2": []interface{}{
										map[string]interface{}{
											"if": []interface{}{
												// Condição: subtotal > 1000
												map[string]interface{}{
													">": []interface{}{
														map[string]interface{}{"var": []interface{}{"totals.subtotal", 0}},
														1000,
													},
												},
												// Se verdadeiro: 20% de desconto
												map[string]interface{}{
													"*": []interface{}{
														map[string]interface{}{"var": []interface{}{"totals.subtotal", 0}},
														0.20,
													},
												},
												// Se falso: 10% de desconto
												map[string]interface{}{
													"*": []interface{}{
														map[string]interface{}{"var": []interface{}{"totals.subtotal", 0}},
														0.10,
													},
												},
											},
										},
									},
								},
							},
						},
					},
					// Distribuir desconto proporcionalmente entre itens usando allocate
					{
						ID:        "allocate-discount",
						Phase:     "allocation",
						Priority:  2,
						Enabled:   true,
						Condition: map[string]interface{}{
							">": []interface{}{
								map[string]interface{}{"var": []interface{}{"totals.discount", 0}},
								0,
							},
						},
						Actions: []core.Action{
							{
								Type:   "compute",
								Target: "items[*].fields.discount",
								Logic: map[string]interface{}{
									"foreach": []interface{}{
										map[string]interface{}{"var": "items"},
										map[string]interface{}{
											"*": []interface{}{
												map[string]interface{}{"var": "item.fields.value"},
												map[string]interface{}{
													"/": []interface{}{
														map[string]interface{}{"var": []interface{}{"totals.discount", 0}},
														map[string]interface{}{"var": []interface{}{"totals.subtotal", 0}},
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
			{
				Name: "taxes",
				Rules: []core.Rule{
					// Imposto condicional baseado no tipo de cliente
					{
						ID:        "conditional-tax",
						Phase:     "taxes",
						Priority:  1,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "compute",
								Target: "totals.tax",
								Logic: map[string]interface{}{
									"round2": []interface{}{
										map[string]interface{}{
											"if": []interface{}{
												// Se customerType == "PF"
												map[string]interface{}{
													"==": []interface{}{
														map[string]interface{}{"var": "customerType"},
														"PF",
													},
												},
												// Taxa de 10%
												map[string]interface{}{
													"*": []interface{}{
														map[string]interface{}{
															"-": []interface{}{
																map[string]interface{}{"var": []interface{}{"totals.subtotal", 0}},
																map[string]interface{}{"var": []interface{}{"totals.discount", 0}},
															},
														},
														0.10,
													},
												},
												// Taxa de 20% para PJ
												map[string]interface{}{
													"*": []interface{}{
														map[string]interface{}{
															"-": []interface{}{
																map[string]interface{}{"var": []interface{}{"totals.subtotal", 0}},
																map[string]interface{}{"var": []interface{}{"totals.discount", 0}},
															},
														},
														0.20,
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
			{
				Name: "totals",
				Rules: []core.Rule{
					{
						ID:        "final-total",
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
														map[string]interface{}{"var": []interface{}{"totals.subtotal", 0}},
														map[string]interface{}{"var": []interface{}{"totals.discount", 0}},
													},
												},
												map[string]interface{}{"var": []interface{}{"totals.tax", 0}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "guards",
				Rules: []core.Rule{
					// Validação: desconto máximo
					{
						ID:        "validate-max-discount",
						Phase:     "guards",
						Priority:  1,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "validate",
								Target: "",
								Logic: map[string]interface{}{
									">": []interface{}{
										map[string]interface{}{
											"/": []interface{}{
												map[string]interface{}{"var": []interface{}{"totals.discount", 0}},
												map[string]interface{}{"var": []interface{}{"totals.subtotal", 0}},
											},
										},
										0.30, // 30% máximo
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
					// Validação: total mínimo
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
										map[string]interface{}{"var": []interface{}{"totals.total", 0}},
										10.0, // Mínimo R$ 10,00
									},
								},
								Params: map[string]interface{}{
									"field":   "totals.total",
									"code":    "MIN_TOTAL_NOT_MET",
									"message": "Total mínimo da compra é R$ 10,00",
								},
							},
						},
					},
				},
			},
		},
	}

	// Estado com múltiplos itens e campos condicionais
	state := core.State{
		TenantID: "tenant-789",
		Items: []core.Item{
			{
				ID:     "item-1",
				Amount: 5,
				Fields: map[string]interface{}{
					"basePrice": 100.00,
				},
			},
			{
				ID:     "item-2",
				Amount: 3,
				Fields: map[string]interface{}{
					"basePrice": 150.00,
				},
			},
			{
				ID:     "item-3",
				Amount: 2,
				Fields: map[string]interface{}{
					"basePrice": 200.00,
				},
			},
		},
		Fields: map[string]interface{}{
			"customerType": "PF", // PF = 10% tax, PJ = 20% tax
		},
		Totals: core.Totals{},
	}

	// Executar engine
	result, err := engine.RunEngine(
		context.Background(),
		state,
		rulePack,
		core.ContextMeta{
			TenantID: "tenant-789",
			UserID:   "user-999",
			Locale:   "pt-BR",
		},
	)

	if err != nil {
		log.Fatalf("❌ Erro na execução: %v", err)
	}

	// Exibir resultados formatados
	fmt.Println("✅ Execução Completa!\n")
	fmt.Println("📊 Resumo:")
	fmt.Printf("  Versão: %s\n", result.RulesVersion)
	fmt.Printf("  Regras Executadas: %d\n", len(result.Reasons))
	fmt.Printf("  Violações: %d\n\n", len(result.Violations))

	// Mostrar cálculos passo a passo
	if len(result.Reasons) > 0 {
		fmt.Println("📝 Regras Executadas (por fase):")
		phases := make(map[string][]core.Reason)
		for _, reason := range result.Reasons {
			phases[reason.Phase] = append(phases[reason.Phase], reason)
		}

		phaseOrder := []string{"baseline", "allocation", "taxes", "totals", "guards"}
		for _, phase := range phaseOrder {
			if reasons, ok := phases[phase]; ok {
				fmt.Printf("\n  %s:\n", phase)
				for _, reason := range reasons {
					fmt.Printf("    - %s: %s\n", reason.RuleID, reason.Message)
				}
			}
		}
		fmt.Println()
	}

	// Mostrar violations se houver
	if len(result.Violations) > 0 {
		fmt.Println("⚠️  Violações:")
		for _, v := range result.Violations {
			fmt.Printf("  ❌ %s: %s (%s)\n", v.Field, v.Message, v.Code)
		}
		fmt.Println()
	}

	// Mostrar resultados finais
	fmt.Println("💰 Resultados Finais:")
	stateJSON, _ := json.MarshalIndent(result.StateFragment, "", "  ")
	fmt.Println(string(stateJSON))
}
