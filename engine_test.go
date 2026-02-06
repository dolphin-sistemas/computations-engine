package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dolphin-sistemas/computations-engine/core"
	"github.com/dolphin-sistemas/computations-engine/loader"
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
	result, err := RunEngine(
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

	if result.RulesVersion != "v1.0.0" {
		t.Errorf("Expected rulesVersion v1.0.0, got %s", result.RulesVersion)
	}

	if len(result.Reasons) == 0 {
		t.Error("Expected at least one reason")
	}

	if result.StateFragment == nil {
		t.Error("Expected stateFragment")
	}

	if result.ServerDelta == nil {
		t.Error("Expected serverDelta")
	}

	if result.Violations == nil {
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
				Name  string `json:"name"`
				Input struct {
					Order    core.State       `json:"order"`
					RulePack core.RulePack    `json:"rulePack"`
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
			result, err := RunEngine(
				context.Background(),
				vector.Input.Order,
				vector.Input.RulePack,
				vector.Input.Context,
			)
			if err != nil {
				t.Fatalf("RunEngine failed: %v", err)
			}

			// Verificar rulesVersion
			if result.RulesVersion != vector.Expected.RulesVersion {
				t.Errorf("rulesVersion mismatch: got %s, expected %s", result.RulesVersion, vector.Expected.RulesVersion)
			}

			// Verificar violations
			if len(vector.Expected.Violations) != len(result.Violations) {
				t.Errorf("violations count mismatch: got %d, expected %d", len(result.Violations), len(vector.Expected.Violations))
				t.Logf("Expected violations: %+v", vector.Expected.Violations)
				t.Logf("Got violations: %+v", result.Violations)
			}

			// Verificar stateFragment (expected is treated as a subset of actual)
			if vector.Expected.StateFragment != nil && len(vector.Expected.StateFragment) > 0 {
				actualNorm, err := normalizeJSON(result.StateFragment)
				if err != nil {
					t.Fatalf("failed to normalize actual stateFragment: %v", err)
				}
				expectedNorm, err := normalizeJSON(vector.Expected.StateFragment)
				if err != nil {
					t.Fatalf("failed to normalize expected stateFragment: %v", err)
				}
				assertSubset(t, expectedNorm, actualNorm, "stateFragment")
			}

			// Log para debug
			if len(result.Reasons) > 0 {
				t.Logf("Executed %d rules", len(result.Reasons))
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

			result, err := RunEngine(
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

			if fields, ok := result.StateFragment["fields"].(map[string]interface{}); ok {
				if resultValue, ok := fields["result"].(float64); ok {
					diff := resultValue - tt.expected
					if diff < 0 {
						diff = -diff
					}
					if diff > 0.001 {
						t.Errorf("Expected %.3f, got %.3f", tt.expected, resultValue)
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

// TestRunEngine_ErrorCases testa cenários de erro
func TestRunEngine_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		state       core.State
		rulePack    core.RulePack
		contextMeta core.ContextMeta
		wantErr     bool
		errContains string
	}{
		{
			name: "missing_rulepack_id",
			state: core.State{
				TenantID: "test-tenant",
				Items:    []core.Item{},
				Fields:   make(map[string]interface{}),
				Totals:   core.Totals{},
			},
			rulePack: core.RulePack{
				ID:      "",
				Version: "v1.0.0",
				Phases:  []core.RulePhase{},
			},
			contextMeta: core.ContextMeta{
				TenantID: "test-tenant",
			},
			wantErr:     true,
			errContains: "rulePack.id is required",
		},
		{
			name: "invalid_jsonlogic_operator",
			state: core.State{
				TenantID: "test-tenant",
				Items:    []core.Item{},
				Fields:   make(map[string]interface{}),
				Totals:   core.Totals{},
			},
			rulePack: core.RulePack{
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
											"invalid_operator_xyz": []interface{}{1, 2, 3},
										},
									},
								},
							},
						},
					},
				},
			},
			contextMeta: core.ContextMeta{
				TenantID: "test-tenant",
			},
			wantErr:     true,
			errContains: "failed to apply jsonlogic",
		},
		{
			name: "validate_missing_params",
			state: core.State{
				TenantID: "test-tenant",
				Items:    []core.Item{},
				Fields:   make(map[string]interface{}),
				Totals:   core.Totals{},
			},
			rulePack: core.RulePack{
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
										Params: make(map[string]interface{}),
									},
								},
							},
						},
					},
				},
			},
			contextMeta: core.ContextMeta{
				TenantID: "test-tenant",
			},
			wantErr:     true,
			errContains: "validate action requires field and code in params",
		},
		{
			name: "validate_missing_logic",
			state: core.State{
				TenantID: "test-tenant",
				Items:    []core.Item{},
				Fields:   make(map[string]interface{}),
				Totals:   core.Totals{},
			},
			rulePack: core.RulePack{
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
			},
			contextMeta: core.ContextMeta{
				TenantID: "test-tenant",
			},
			wantErr:     true,
			errContains: "validate action requires logic",
		},
		{
			name: "invalid_condition",
			state: core.State{
				TenantID: "test-tenant",
				Items:    []core.Item{},
				Fields:   make(map[string]interface{}),
				Totals:   core.Totals{},
			},
			rulePack: core.RulePack{
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
									"unknown_operator_xyz": []interface{}{1, 2},
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
			},
			contextMeta: core.ContextMeta{
				TenantID: "test-tenant",
			},
			wantErr:     true,
			errContains: "failed to evaluate condition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RunEngine(
				context.Background(),
				tt.state,
				tt.rulePack,
				tt.contextMeta,
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errContains, err)
				}
				t.Logf("✓ Expected error occurred: %v", err)
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result == nil {
					t.Error("Expected result, got nil")
				}
			}
		})
	}
}

// TestRunEngine_ErrorVectors testa vectors de erro do diretório testdata/errors
func TestRunEngine_ErrorVectors(t *testing.T) {
	errorsDir := "testdata/errors"
	entries, err := os.ReadDir(errorsDir)
	if err != nil {
		t.Skipf("testdata/errors directory not found: %v", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			path := filepath.Join(errorsDir, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read error vector file: %v", err)
			}

			var errorVector struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Input       struct {
					Order    core.State       `json:"order"`
					RulePack core.RulePack    `json:"rulePack"`
					Context  core.ContextMeta `json:"context"`
				} `json:"input"`
				ExpectedError string `json:"expectedError"`
				Note          string `json:"note"`
			}

			if err := json.Unmarshal(data, &errorVector); err != nil {
				t.Fatalf("failed to unmarshal error vector: %v", err)
			}

			// Executar motor - deve retornar erro
			result, err := RunEngine(
				context.Background(),
				errorVector.Input.Order,
				errorVector.Input.RulePack,
				errorVector.Input.Context,
			)

			if err == nil {
				t.Errorf("Expected error '%s', but got nil. Result: %+v", errorVector.ExpectedError, result)
				return
			}

			if errorVector.ExpectedError != "" {
				if !contains(err.Error(), errorVector.ExpectedError) {
					t.Errorf("Expected error to contain '%s', got: %v", errorVector.ExpectedError, err)
				} else {
					t.Logf("✓ Expected error occurred: %v", err)
				}
			}

			if errorVector.Note != "" {
				t.Logf("Note: %s", errorVector.Note)
			}
		})
	}
}

// contains verifica se uma string contém uma substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func normalizeJSON(v interface{}) (interface{}, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func assertSubset(t *testing.T, expected, actual interface{}, path string) {
	t.Helper()

	if expected == nil {
		if actual != nil {
			t.Fatalf("%s: expected nil, got %T", path, actual)
		}
		return
	}

	switch e := expected.(type) {
	case map[string]interface{}:
		a, ok := actual.(map[string]interface{})
		if !ok {
			t.Fatalf("%s: expected object, got %T", path, actual)
		}
		for k, ev := range e {
			av, exists := a[k]
			if !exists {
				t.Fatalf("%s.%s: missing key", path, k)
			}
			assertSubset(t, ev, av, path+"."+k)
		}
		return

	case []interface{}:
		a, ok := actual.([]interface{})
		if !ok {
			t.Fatalf("%s: expected array, got %T", path, actual)
		}
		if len(a) != len(e) {
			t.Fatalf("%s: array length mismatch: got %d, expected %d", path, len(a), len(e))
		}
		for i := range e {
			assertSubset(t, e[i], a[i], fmt.Sprintf("%s[%d]", path, i))
		}
		return

	case float64:
		af, ok := actual.(float64)
		if !ok {
			t.Fatalf("%s: expected number, got %T", path, actual)
		}
		diff := af - e
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.01 {
			t.Fatalf("%s: number mismatch: got %.6f, expected %.6f", path, af, e)
		}
		return

	default:
		if expected != actual {
			t.Fatalf("%s: mismatch: got %#v, expected %#v", path, actual, expected)
		}
	}
}
