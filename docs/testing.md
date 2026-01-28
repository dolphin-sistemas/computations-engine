# Guia de Testes

Este documento explica como testar a biblioteca engine.

## Executar Testes Unitários

### Todos os testes

```bash
cd engine
go test ./...
```

### Testes com verbose

```bash
go test -v ./...
```

### Testes específicos

```bash
# Teste básico
go test -v -run TestRunEngine_Basic

# Testes com test vectors
go test -v -run TestRunEngine_WithTestVectors

# Testes de operações matemáticas
go test -v -run TestRunEngine_MathOperations
```

### Testes com cobertura

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Vectors

Os test vectors estão em `testdata/vectors/` e contêm cenários determinísticos:

- `vector1_baseline.json` - Inicialização básica
- `vector2_allocation.json` - Distribuição de valores
- `vector3_taxes.json` - Cálculo de impostos
- `vector4_totals.json` - Cálculo de totais complexo
- `vector5_guards.json` - Validações e bloqueios
- `vector6_dynamic_layout.json` - Validações de layout dinâmico (required, min, max, pattern, condicionais)

Cada vector contém:
- `input`: State + RulePack + Context
- `expected`: StateFragment, RulesVersion, Violations

## Test Vectors de Erro

Os test vectors de erro estão em `testdata/errors/` e testam cenários de erro:

- `error1_missing_rulepack_id.json` - RulePack sem ID
- `error2_invalid_jsonlogic.json` - JsonLogic inválido
- `error3_validate_missing_params.json` - Ação validate sem params obrigatórios
- `error4_division_by_zero.json` - Divisão por zero
- `error5_invalid_condition.json` - Condição com JsonLogic inválido
- `error6_validate_missing_logic.json` - Ação validate sem logic

Cada error vector contém:
- `input`: State + RulePack + Context
- `expectedError`: Mensagem de erro esperada
- `description`: Descrição do cenário de erro
- `note`: Notas adicionais (opcional)

### Executar Testes de Erro

```bash
# Todos os testes de erro
go test -v -run TestRunEngine_ErrorCases

# Testes de erro via vectors
go test -v -run TestRunEngine_ErrorVectors
```

## Testar Manualmente

### 1. Usando o exemplo

```bash
cd engine
go run examples/example.go
```

### 2. Criar um teste simples

Crie um arquivo `test_manual.go`:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/dolphin-sistemas/computations-engine"
	"github.com/dolphin-sistemas/computations-engine/core"
)

func main() {
	// Criar estado
	state := core.State{
		TenantID: "test",
		Items: []core.Item{
			{
				ID:     "item-1",
				Amount: 2,
				Fields: map[string]interface{}{
					"price": 100.0,
				},
			},
		},
		Fields: make(map[string]interface{}),
		Totals: core.Totals{},
	}

	// Criar RulePack
	rulePack := core.RulePack{
		ID:      "test",
		Version: "v1.0.0",
		Phases: []core.RulePhase{
			{
				Name: "baseline",
				Rules: []core.Rule{
					{
						ID:        "calc-total",
						Phase:     "baseline",
						Priority:  1,
						Enabled:   true,
						Condition: nil,
						Actions: []core.Action{
							{
								Type:   "compute",
								Target: "totals.total",
								Logic: map[string]interface{}{
									"*": []interface{}{
										map[string]interface{}{"var": []interface{}{"items", []interface{}{}}},
										2.0,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Executar
	result, err := engine.RunEngine(
		context.Background(),
		state,
		rulePack,
		core.ContextMeta{TenantID: "test"},
	)

	if err != nil {
		log.Fatal(err)
	}

	// Exibir resultados
	fmt.Printf("Version: %s\n", result.RulesVersion)
	fmt.Printf("Reasons: %d\n", len(result.Reasons))
	fmt.Printf("Violations: %d\n", len(result.Violations))
	
	resultJSON, _ := json.MarshalIndent(result.StateFragment, "", "  ")
	fmt.Printf("State Fragment:\n%s\n", resultJSON)

	deltaJSON, _ := json.MarshalIndent(result.ServerDelta, "", "  ")
	fmt.Printf("Server Delta:\n%s\n", deltaJSON)
}
```

Execute:
```bash
go run test_manual.go
```

## Testar Operações Matemáticas

### Soma
```json
{
  "type": "compute",
  "target": "fields.result",
  "logic": {
    "+": [10, 5]
  }
}
```

### Subtração
```json
{
  "type": "compute",
  "target": "fields.result",
  "logic": {
    "-": [10, 3]
  }
}
```

### Multiplicação
```json
{
  "type": "compute",
  "target": "fields.result",
  "logic": {
    "*": [5, 4]
  }
}
```

### Divisão
```json
{
  "type": "compute",
  "target": "fields.result",
  "logic": {
    "/": [20, 4]
  }
}
```

### Expressão Complexa
```json
{
  "type": "compute",
  "target": "fields.result",
  "logic": {
    "+": [
      {"*": [2, 3]},
      {"/": [10, 2]}
    ]
  }
}
```

## Verificar Resultados

Os testes verificam:
- ✅ `rulesVersion` corresponde ao esperado
- ✅ `stateFragment` contém os campos calculados
- ✅ `serverDelta` contém as diferenças
- ✅ `reasons` contém as regras executadas
- ✅ Operações matemáticas produzem resultados corretos
- ✅ Tolerância de 0.01 para arredondamento

## Troubleshooting

### Erro: "testdata/vectors directory not found"
- Certifique-se de executar os testes a partir do diretório `engine/`
- Verifique se os arquivos em `testdata/vectors/` existem

### Erro: "failed to load rule pack"
- Verifique o formato JSON do RulePack
- Certifique-se de que `id` e `version` estão presentes

### Resultados diferentes do esperado
- Verifique a ordem das fases no pipeline
- Confirme que as regras estão habilitadas (`enabled: true`)
- Verifique a prioridade das regras
