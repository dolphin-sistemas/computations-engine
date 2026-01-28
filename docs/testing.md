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

Cada vector contém:
- `input`: State + RulePack + Context
- `expected`: StateFragment, RulesVersion, Violations

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
	result, delta, reasons, violations, version, err := engine.RunEngine(
		context.Background(),
		state,
		rulePack,
		core.ContextMeta{TenantID: "test"},
	)

	if err != nil {
		log.Fatal(err)
	}

	// Exibir resultados
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Reasons: %d\n", len(reasons))
	fmt.Printf("Violations: %d\n", len(violations))
	
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Result:\n%s\n", resultJSON)

	deltaJSON, _ := json.MarshalIndent(delta, "", "  ")
	fmt.Printf("Delta:\n%s\n", deltaJSON)
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
