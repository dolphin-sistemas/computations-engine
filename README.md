# Engine - Motor de Regras

Biblioteca Go completa e reutilizável para processamento de regras de negócio usando JsonLogic.

## Características

- **Pipeline de Fases**: Execução sequencial de fases (baseline → allocation → taxes → totals → guards)
- **JsonLogic + Operadores Customizados**: Suporte a operadores como `sum`, `round`, `round2`, `if`, `foreach`, `allocate`
- **Determinístico**: Mesmo input sempre produz mesmo output
- **Modular**: Arquitetura organizada em pastas por responsabilidade
- **Extensível**: Fácil adicionar novos operadores, ações e validadores
- **WASM Support**: Pode ser compilado para WebAssembly

## Instalação

```bash
go get github.com/dolphin-sistemas/computations-engine
```

Para desenvolvimento local:

```bash
# No go.mod do seu projeto
replace github.com/dolphin-sistemas/computations-engine => ../engine
```

## Uso Básico

A engine processa um `State` (estado dos dados) usando um `RulePack` (pacote de regras) e retorna os resultados.

### Formato do RulePack

O RulePack deve estar no formato JSON (ou YAML) conforme a estrutura abaixo. Você pode carregá-lo de qualquer fonte (arquivo, banco de dados, API, etc.):

```json
{
  "id": "rule-pack-id",
  "version": "v1.0.0",
  "phases": [
    {
      "name": "baseline",
      "rules": [
        {
          "id": "calc-subtotal",
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

### Exemplo de Código

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
	// 1. Carregar RulePack (de arquivo, banco de dados, API, etc.)
	rulePackJSON := `{
		"id": "rules-1",
		"version": "v1.0.0",
		"phases": [
			{
				"name": "baseline",
				"rules": [
					{
						"id": "calc-subtotal",
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
	}`

	var rulePack core.RulePack
	if err := json.Unmarshal([]byte(rulePackJSON), &rulePack); err != nil {
		log.Fatal(err)
	}

	// 2. Criar estado inicial
	state := core.State{
		TenantID: "tenant-1",
		Items: []core.Item{
			{
				ID:     "item-1",
				Amount: 2,
				Fields: map[string]interface{}{
					"value": 100.0,
				},
			},
		},
		Fields: make(map[string]interface{}),
		Totals: core.Totals{},
	}

	// 3. Executar motor
	result, err := engine.RunEngine(
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
		log.Fatal(err)
	}

	// 4. Usar resultados
	fmt.Printf("Version: %s\n", result.RulesVersion)
	fmt.Printf("Reasons: %d\n", len(result.Reasons))
	fmt.Printf("Violations: %d\n", len(result.Violations))
	
	// result.StateFragment: campos que mudaram (para atualizar UI)
	// result.ServerDelta: diferenças para sincronização
	// result.Reasons: regras que executaram
	// result.Violations: violações de validação (se houver)
}
```

**Nota**: O exemplo acima mostra RulePack como JSON inline. Na prática, você pode carregá-lo de qualquer fonte (arquivo, banco de dados, API, etc.) e fazer `json.Unmarshal` para obter o `core.RulePack`.

## Estrutura do Projeto

```
engine/
├── core/           # Tipos e estruturas principais (State, RulePack, etc.)
├── pipeline/       # Execução de fases e regras
├── operators/      # Operadores JsonLogic customizados
├── actions/        # Execução de ações (set, compute, validate, add, multiply)
├── guards/         # Validações e guards
├── loader/         # Carregamento de RulePacks (JSON/YAML)
├── diff/           # Geração de deltas (stateFragment, serverDelta)
├── pkg/            # Utilitários compartilhados
├── cmd/            # Entry points (WASM)
│   └── wasm/       # WASM entry point
├── examples/       # Exemplos de uso
├── testdata/       # Test vectors
└── docs/           # Documentação completa
```

## Pipeline de Fases

O motor executa fases na seguinte ordem:

1. **baseline**: Inicialização de valores base
2. **allocation**: Distribuição/alocação de valores
3. **taxes**: Cálculo de impostos
4. **totals**: Cálculo de totais
5. **guards**: Validações finais e bloqueios

Cada fase executa suas regras em ordem de prioridade (menor = primeiro).

## Operadores Customizados

A engine inclui os seguintes operadores customizados além dos operadores nativos do JsonLogic:

### `sum`
Soma todos os valores de um array:
```json
{"sum": [[10, 20, 30]]}  // → 60
{"sum": [{"var": ["itemValues", []]}]}  // Soma valores de itemValues
```

### `round2`
Arredonda para 2 casas decimais:
```json
{"round2": [10.456]}  // → 10.46
{"round2": [{"var": "totals.total"}]}
```

### `round`
Arredondamento genérico com casas decimais configuráveis:
```json
{"round": [10.456, 1]}  // → 10.5 (1 casa decimal)
{"round": [10.456, 0]}  // → 10 (0 casas decimais)
{"round": [10.456, 2]}  // → 10.46 (2 casas decimais)
```

### `if`
Condicional ternário (suporta JsonLogic aninhado):
```json
{"if": [condição, valorSeVerdadeiro, valorSeFalso]}
{"if": [
  {">": [{"var": "totals.total"}, 100]},
  {"*": [{"var": "totals.total"}, 0.1]},
  0
]}
```

### `foreach`
Iteração sobre arrays aplicando lógica a cada elemento:
```json
{"foreach": [array, lógica]}
{"foreach": [
  {"var": "items"},
  {"*": [{"var": "item.amount"}, {"var": "item.fields.price"}]}
]}
```
O contexto de cada iteração inclui:
- `item`: elemento atual do array
- `index`: índice do elemento (float64)

### `allocate`
Distribuição proporcional de um total baseado em pesos:
```json
{"allocate": [total, pesos]}
{"allocate": [100, [1, 2, 3]]}  // → [16.67, 33.33, 50.0]
```
Distribui o total proporcionalmente aos pesos. Se a soma dos pesos for zero, distribui igualmente.

## Operações Matemáticas Nativas

A biblioteca suporta todas as operações matemáticas básicas via JsonLogic nativo:

- **`+`** (soma): `{"+": [10, 5]}` → `15`
- **`-`** (subtração): `{"-": [10, 3]}` → `7`
- **`*`** (multiplicação): `{"*": [5, 4]}` → `20`
- **`/`** (divisão): `{"/": [20, 4]}` → `5`
- **`%`** (módulo): `{"%": [10, 3]}` → `1`

## Tipos de Ações

### `set`
Define valor literal em um target:
```json
{
  "type": "set",
  "target": "fields.status",
  "value": "active"
}
```

### `compute`
Calcula valor usando JsonLogic:
```json
{
  "type": "compute",
  "target": "totals.subtotal",
  "logic": {
    "sum": [{"var": ["itemValues", []]}]
  }
}
```

### `validate`
Valida condição e cria violação se a lógica retornar `true`:
```json
{
  "type": "validate",
  "logic": {
    ">": [{"var": "totals.discount"}, 1000]
  },
  "params": {
    "field": "totals.discount",
    "code": "MAX_DISCOUNT_EXCEEDED",
    "message": "Discount cannot exceed 1000"
  }
}
```

### `add`
Incrementa valor existente:
```json
{
  "type": "add",
  "target": "totals.total",
  "value": 10.5
}
```

### `multiply`
Multiplica valor existente:
```json
{
  "type": "multiply",
  "target": "totals.total",
  "value": 1.1
}
```

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

### Estrutura de uma Regra

- **id**: Identificador único da regra
- **phase**: Nome da fase (baseline, allocation, taxes, totals, guards)
- **priority**: Prioridade (menor = executa primeiro, padrão: 0)
- **enabled**: Se a regra está habilitada (padrão: true)
- **condition**: JsonLogic para avaliar se a regra deve executar (null = sempre executa)
- **actions**: Lista de ações a executar se condition for verdadeira

## Retorno da Engine

A função `RunEngine` retorna um único objeto `RunEngineResult`:

```go
result, err := engine.RunEngine(ctx, state, rulePack, contextMeta)
```

### `result.StateFragment`
Mapa com apenas os campos que mudaram (útil para atualizar UI):
```json
{
  "totals": {
    "subtotal": 100.0,
    "total": 99.0
  },
  "fields": {
    "calculatedAt": "2026-01-26T10:00:00Z"
  }
}
```

### `result.ServerDelta`
Mapa com diferenças no formato chave-valor (útil para sincronização):
```json
{
  "totals.total": 99.0,
  "totals.discount": 10.0
}
```

### `result.Reasons`
Array de regras que executaram:
```json
[
  {
    "ruleId": "apply-discount",
    "phase": "allocation",
    "message": "Applied 10% discount"
  }
]
```

### `result.Violations`
Array de violações de validação (vazio em sucesso):
```json
[
  {
    "field": "totals.discount",
    "code": "MAX_DISCOUNT_EXCEEDED",
    "message": "Discount cannot exceed 50%"
  }
]
```

### `result.RulesVersion`
Versão das regras usadas (do RulePack.version)

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

### Test Vectors

Test vectors determinísticos estão em `testdata/vectors/`:
- `vector1_baseline.json` - Inicialização básica
- `vector2_allocation.json` - Distribuição de valores
- `vector3_taxes.json` - Cálculo de impostos
- `vector4_totals.json` - Cálculo de totais complexo
- `vector5_guards.json` - Validações e bloqueios
- `vector6_dynamic_layout.json` - Validações de layout dinâmico (required, min, max, pattern, condicionais)

Test vectors de erro estão em `testdata/errors/`:
- `error1_missing_rulepack_id.json` - RulePack sem ID
- `error2_invalid_jsonlogic.json` - JsonLogic inválido
- `error3_validate_missing_params.json` - Ação validate sem params
- `error4_division_by_zero.json` - Divisão por zero
- `error5_invalid_condition.json` - Condição inválida
- `error6_validate_missing_logic.json` - Ação validate sem logic

Para mais detalhes, veja [docs/testing.md](docs/testing.md).

## Build WASM

Para compilar a engine para WebAssembly:

```bash
make wasm
# ou
bash scripts/build-wasm.sh
```

Isso gera:
- `client/wasm/order_engine.wasm`
- `client/wasm/wasm_exec.js`

Para mais detalhes, veja [docs/wasm.md](docs/wasm.md).

## Documentação

Documentação completa disponível na pasta [`docs/`](docs/):

- **[Integração](docs/integracao.md)** - Como integrar a engine em seu projeto
- **[Validação](docs/validacao.md)** - Como funciona a validação
- **[WASM](docs/wasm.md)** - Build e uso do WASM
- **[Testes](docs/testing.md)** - Guia completo de testes
- **[Exemplos](docs/usage-example.md)** - Exemplos práticos de uso

## Licença

MIT
