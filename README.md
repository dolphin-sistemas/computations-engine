# Engine - Motor de Regras

Biblioteca Go completa e reutilizável para processamento de regras de negócio usando JsonLogic.

## Características

- **Pipeline de Fases**: Execução sequencial de fases (baseline → allocation → taxes → totals → guards)
- **JsonLogic + Operadores Customizados**: Suporte a operadores como `if`, `foreach`, `round`, `allocate`
- **Determinístico**: Mesmo input sempre produz mesmo output
- **Modular**: Arquitetura organizada em pastas por responsabilidade
- **Extensível**: Fácil adicionar novos operadores, ações e validadores

## Instalação

```bash
go get github.com/dolphin-sistemas/engine
```

## Uso Básico

```go
package main

import (
    "context"
    "github.com/dolphin-sistemas/engine/core"
    "github.com/dolphin-sistemas/engine"
    "github.com/dolphin-sistemas/engine/loader"
)

func main() {
    // Carregar RulePack
    rulePack, err := loader.LoadRulePackFromFile("rules.json")
    if err != nil {
        panic(err)
    }

    // Criar estado
    state := core.State{
        TenantID: "tenant-1",
        Items: []core.Item{
            {
                ID:     "item-1",
                Amount: 2,
                Fields: map[string]interface{}{
                    "basePrice": 100.0,
                },
            },
        },
        Fields: make(map[string]interface{}),
        Totals: core.Totals{},
    }

    // Executar motor
    stateFragment, serverDelta, reasons, violations, rulesVersion, err := engine.RunEngine(
        context.Background(),
        state,
        rulePack,
        core.ContextMeta{
            TenantID: "tenant-1",
            UserID:   "user-1",
            Locale:   "pt-BR",
        },
    )
    if err != nil {
        panic(err)
    }

    // Usar resultados
    fmt.Printf("Version: %s\n", rulesVersion)
    fmt.Printf("Reasons: %+v\n", reasons)
    fmt.Printf("Violations: %+v\n", violations)
}
```

## Estrutura do Projeto

```
engine/
├── core/           # Tipos e estruturas principais
├── pipeline/       # Execução de fases
├── operators/      # Operadores JsonLogic customizados
├── actions/        # Execução de ações
├── guards/         # Validações e guards
├── loader/         # Carregamento de RulePacks
├── diff/           # Geração de deltas
└── examples/       # Exemplos de uso
```

## Pipeline de Fases

O motor executa fases na seguinte ordem:

1. **baseline**: Inicialização de valores base
2. **allocation**: Distribuição/alocação de valores
3. **taxes**: Cálculo de impostos
4. **totals**: Cálculo de totais
5. **guards**: Validações finais e bloqueios

## Operadores Customizados

### `if`
Condicional ternário:
```json
{"if": [condição, valorSeVerdadeiro, valorSeFalso]}
```

### `foreach`
Iteração sobre arrays:
```json
{"foreach": [array, lógica]}
```

### `round`
Arredondamento genérico:
```json
{"round": [valor, casasDecimais]}
```

### `allocate`
Distribuição proporcional:
```json
{"allocate": [total, pesos]}
```

## Tipos de Ações

- **set**: Define valor literal
- **compute**: Calcula valor usando JsonLogic
- **validate**: Valida condição e cria violação se falsa
- **add**: Incrementa valor existente
- **multiply**: Multiplica valor existente

## Formato de RulePack

```json
{
  "id": "rule-pack-id",
  "version": "v1.0.0",
  "phases": [
    {
      "name": "baseline",
      "rules": [
        {
          "id": "rule-id",
          "phase": "baseline",
          "priority": 1,
          "enabled": true,
          "condition": null,
          "actions": [
            {
              "type": "compute",
              "target": "totals.subtotal",
              "logic": {
                "sum": [{"var": ["itemValues", []]}]
              }
            }
          ]
        }
      ]
    }
  ]
}
```

## Testes

### Executar Todos os Testes

```bash
cd engine
go test ./...
```

### Testes com Verbose

```bash
go test -v ./...
```

### Testes Específicos

```bash
# Teste básico
go test -v -run TestRunEngine_Basic

# Testes com test vectors
go test -v -run TestRunEngine_WithTestVectors

# Testes de operações matemáticas
go test -v -run TestRunEngine_MathOperations
```

### Testar Manualmente

```bash
# Executar exemplo
go run examples/example.go
```

### Operações Matemáticas Suportadas

A biblioteca suporta todas as operações matemáticas básicas via JsonLogic:

- **`+`** (soma): `{"+": [10, 5]}` → `15`
- **`-`** (subtração): `{"-": [10, 3]}` → `7`
- **`*`** (multiplicação): `{"*": [5, 4]}` → `20`
- **`/`** (divisão): `{"/": [20, 4]}` → `5`
- **`%`** (módulo): `{"%": [10, 3]}` → `1`

### Test Vectors

Test vectors determinísticos estão em `testdata/vectors/`:
- `vector1_baseline.json` - Inicialização básica
- `vector2_allocation.json` - Distribuição de valores
- `vector3_taxes.json` - Cálculo de impostos
- `vector4_totals.json` - Cálculo de totais complexo
- `vector5_guards.json` - Validações e bloqueios

Para mais detalhes, veja [TESTING.md](TESTING.md).

## Licença

MIT
