package engine

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dolphin-sistemas/engine/core"
	"github.com/dolphin-sistemas/engine/loader"
)

func TestRunEngine_Basic(t *testing.T) {
	// Criar RulePack simples
	rulePack := core.RulePack{
		ID:      "test-pack",
		Version: "v1.0.0",
		Phases: []core.RulePhase{
			{
				Name: "baseline",
				Rules: []core.Rule{
					{
						ID:        "init-total",
						Phase:     "baseline",
						Priority:  1,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "set",
								Target: "totals.total",
								Value:  100.0,
							},
						},
					},
				},
			},
		},
	}

	// Criar estado
	state := core.State{
		TenantID: "test-tenant",
		Items:    []core.Item{},
		Fields:   make(map[string]interface{}),
		Totals:   core.Totals{},
	}

	// Executar motor
	stateFragment, serverDelta, reasons, violations, rulesVersion, err := RunEngine(
		context.Background(),
		state,
		rulePack,
		core.ContextMeta{
			TenantID: "test-tenant",
			UserID:   "test-user",
			Locale:   "pt-BR",
		},
	)

	if err != nil {
		t.Fatalf("RunEngine failed: %v", err)
	}

	if rulesVersion != "v1.0.0" {
		t.Errorf("Expected rulesVersion v1.0.0, got %s", rulesVersion)
	}

	if len(reasons) == 0 {
		t.Error("Expected at least one reason")
	}

	if stateFragment == nil {
		t.Error("Expected stateFragment")
	}

	if serverDelta == nil {
		t.Error("Expected serverDelta")
	}

	if violations == nil {
		t.Error("Expected violations (can be empty)")
	}
}

func TestLoadRulePackFromJSON(t *testing.T) {
	jsonData := `{
		"id": "test-pack",
		"version": "v1.0.0",
		"phases": []
	}`

	rulePack, err := loader.LoadRulePackFromJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("LoadRulePackFromJSON failed: %v", err)
	}

	if rulePack.ID != "test-pack" {
		t.Errorf("Expected ID test-pack, got %s", rulePack.ID)
	}

	if rulePack.Version != "v1.0.0" {
		t.Errorf("Expected version v1.0.0, got %s", rulePack.Version)
	}
}

func TestRunEngine_WithTestVectors(t *testing.T) {
	vectorsDir := "testdata/vectors"
	entries, err := os.ReadDir(vectorsDir)
	if err != nil {
		t.Skipf("testdata/vectors directory not found: %v", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			path := filepath.Join(vectorsDir, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read vector file: %v", err)
			}

			var vector struct {
				Name     string       `json:"name"`
				Input    struct {
					Order    core.State   `json:"order"`
					RulePack core.RulePack `json:"rulePack"`
					Context  core.ContextMeta `json:"context"`
				} `json:"input"`
				Expected struct {
					StateFragment map[string]interface{} `json:"stateFragment"`
					RulesVersion  string                 `json:"rulesVersion"`
					Violations    []core.Violation       `json:"violations"`
				} `json:"expected"`
			}

			if err := json.Unmarshal(data, &vector); err != nil {
				t.Fatalf("failed to unmarshal vector: %v", err)
			}

			// Executar motor
			stateFragment, _, reasons, violations, rulesVersion, err := RunEngine(
				context.Background(),
				vector.Input.Order,
				vector.Input.RulePack,
				vector.Input.Context,
			)
			if err != nil {
				t.Fatalf("RunEngine failed: %v", err)
			}

			// Verificar rulesVersion
			if rulesVersion != vector.Expected.RulesVersion {
				t.Errorf("rulesVersion mismatch: got %s, expected %s", rulesVersion, vector.Expected.RulesVersion)
			}

			// Verificar violations
			if len(vector.Expected.Violations) != len(violations) {
				t.Errorf("violations count mismatch: got %d, expected %d", len(violations), len(vector.Expected.Violations))
			}

			// Verificar totals.total (se presente no expected)
			if expectedTotals, ok := vector.Expected.StateFragment["totals"].(map[string]interface{}); ok {
				if expectedTotal, ok := expectedTotals["total"].(float64); ok {
					if stateFragment["totals"] == nil {
						t.Error("stateFragment.totals is missing")
					} else if totals, ok := stateFragment["totals"].(core.Totals); ok {
						// Comparar com tolerância de 0.01 (erro de arredondamento)
						diff := totals.Total - expectedTotal
						if diff < 0 {
							diff = -diff
						}
						if diff > 0.01 {
							t.Errorf("totals.total mismatch: got %.2f, expected %.2f", totals.Total, expectedTotal)
						}
					}
				}
			}

			// Log para debug
			if len(reasons) > 0 {
				t.Logf("Executed %d rules", len(reasons))
			}
		})
	}
}

func TestRunEngine_MathOperations(t *testing.T) {
	tests := []struct {
		name     string
		logic    map[string]interface{}
		expected float64
	}{
		{
			name: "addition",
			logic: map[string]interface{}{
				"+": []interface{}{10.0, 5.0},
			},
			expected: 15.0,
		},
		{
			name: "subtraction",
			logic: map[string]interface{}{
				"-": []interface{}{10.0, 3.0},
			},
			expected: 7.0,
		},
		{
			name: "multiplication",
			logic: map[string]interface{}{
				"*": []interface{}{5.0, 4.0},
			},
			expected: 20.0,
		},
		{
			name: "division",
			logic: map[string]interface{}{
				"/": []interface{}{20.0, 4.0},
			},
			expected: 5.0,
		},
		{
			name: "complex_expression",
			logic: map[string]interface{}{
				"+": []interface{}{
					map[string]interface{}{
						"*": []interface{}{2.0, 3.0},
					},
					map[string]interface{}{
						"/": []interface{}{10.0, 2.0},
					},
				},
			},
			expected: 11.0, // (2*3) + (10/2) = 6 + 5 = 11
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulePack := core.RulePack{
				ID:      "math-test",
				Version: "v1.0.0",
				Phases: []core.RulePhase{
					{
						Name: "baseline",
						Rules: []core.Rule{
							{
								ID:        "compute-result",
								Phase:     "baseline",
								Priority:  1,
								Enabled:   true,
								Condition: nil,
								Actions: []core.Action{
									{
										Type:   "compute",
										Target: "fields.result",
										Logic:  tt.logic,
									},
								},
							},
						},
					},
				},
			}

			state := core.State{
				TenantID: "test-tenant",
				Items:    []core.Item{},
				Fields:   make(map[string]interface{}),
				Totals:   core.Totals{},
			}

			stateFragment, _, _, _, _, err := RunEngine(
				context.Background(),
				state,
				rulePack,
				core.ContextMeta{
					TenantID: "test-tenant",
				},
			)

			if err != nil {
				t.Fatalf("RunEngine failed: %v", err)
			}

			if fields, ok := stateFragment["fields"].(map[string]interface{}); ok {
				if result, ok := fields["result"].(float64); ok {
					diff := result - tt.expected
					if diff < 0 {
						diff = -diff
					}
					if diff > 0.001 {
						t.Errorf("Expected %.3f, got %.3f", tt.expected, result)
					}
				} else {
					t.Errorf("Result not found or not a float64: %v", fields["result"])
				}
			} else {
				t.Error("Fields not found in stateFragment")
			}
		})
	}
}
