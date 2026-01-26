# Guia de Integração - Usar engine no teste-opa-jsonlogic-next

Este guia mostra como integrar a biblioteca `engine` no projeto `teste-opa-jsonlogic-next`.

## Opção 1: Usar como Módulo Local (Desenvolvimento)

### 1. Adicionar replace no go.mod

No arquivo `teste-opa-jsonlogic-next/go.mod`, adicione:

```go
module github.com/dolphin-sistemas/template-api-go

go 1.24.4

require (
    // ... outras dependências
    github.com/dolphin-sistemas/computations-engine v0.0.0
)

replace github.com/dolphin-sistemas/computations-engine => ../engine
```

### 2. Atualizar o Client

Modifique `internal/infra/rulesengine/client.go`:

```go
package rulesengine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/dolphin-sistemas/computations-engine"
	"github.com/dolphin-sistemas/computations-engine/core"
)

// ValidateOrder valida/computa um pedido usando a biblioteca engine
func (c *Client) ValidateOrder(
	ctx context.Context,
	tenantID, userID, rulesVersion string,
	orderState, rulePack map[string]interface{},
) (*ValidateOrderResult, error) {
	// Converter orderState para core.State
	orderStateJSON, err := json.Marshal(orderState)
	if err != nil {
		return nil, fmt.Errorf("marshal order state: %w", err)
	}
	
	var state core.State
	if err := json.Unmarshal(orderStateJSON, &state); err != nil {
		return nil, fmt.Errorf("unmarshal order state: %w", err)
	}

	// Converter rulePack para core.RulePack
	rulePackJSON, err := json.Marshal(rulePack)
	if err != nil {
		return nil, fmt.Errorf("marshal rule pack: %w", err)
	}
	
	var pack core.RulePack
	if err := json.Unmarshal(rulePackJSON, &pack); err != nil {
		return nil, fmt.Errorf("unmarshal rule pack: %w", err)
	}

	// Executar motor usando a nova biblioteca
	stateFragment, serverDelta, reasons, rulesVersionOut, err := engine.RunEngine(
		ctx,
		state,
		pack,
		core.ContextMeta{
			TenantID: tenantID,
			UserID:   userID,
			Locale:   "pt-BR",
		},
	)
	if err != nil {
		c.logger.Error("RulesEngine RunEngine failed", "error", err)
		return nil, fmt.Errorf("rules engine: %w", err)
	}

	// Converter violations e reasons
	// Nota: você precisará obter violations do contexto se necessário
	violations := []Violation{} // Adicionar lógica para obter violations
	reasonsOut := make([]Reason, len(reasons))
	for i, r := range reasons {
		reasonsOut[i] = Reason{
			RuleID:  r.RuleID,
			Phase:   r.Phase,
			Message: r.Message,
		}
	}

	return &ValidateOrderResult{
		StateFragment: stateFragment,
		ServerDelta:   serverDelta,
		Reasons:       reasonsOut,
		RulesVersion:  rulesVersionOut,
		Violations:    violations,
	}, nil
}
```

## Opção 2: Adapter com Compatibilidade

Criar um adapter que mantém a interface atual mas usa a nova biblioteca:

### Criar `internal/infra/rulesengine/adapter.go`

```go
package rulesengine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dolphin-sistemas/computations-engine"
	"github.com/dolphin-sistemas/computations-engine/core"
)

// Adapter adapta a nova biblioteca engine para a interface atual
type Adapter struct {
	// Pode adicionar cache, logger, etc.
}

// ValidateOrder adapta a chamada para usar engine.RunEngine
func (a *Adapter) ValidateOrder(
	ctx context.Context,
	tenantID, userID, rulesVersion string,
	orderState, rulePack map[string]interface{},
) (*ValidateOrderResult, error) {
	// Converter map[string]interface{} para core.State
	state, err := mapToState(orderState)
	if err != nil {
		return nil, fmt.Errorf("convert order state: %w", err)
	}

	// Converter map[string]interface{} para core.RulePack
	pack, err := mapToRulePack(rulePack)
	if err != nil {
		return nil, fmt.Errorf("convert rule pack: %w", err)
	}

	// Executar motor
	stateFragment, serverDelta, reasons, rulesVersionOut, err := engine.RunEngine(
		ctx,
		state,
		pack,
		core.ContextMeta{
			TenantID: tenantID,
			UserID:   userID,
			Locale:   "pt-BR",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("engine execution: %w", err)
	}

	// Converter resultados
	return &ValidateOrderResult{
		StateFragment: stateFragment,
		ServerDelta:   serverDelta,
		Reasons:       convertReasons(reasons),
		RulesVersion:  rulesVersionOut,
		Violations:    []Violation{}, // Adicionar lógica se necessário
	}, nil
}

// mapToState converte map[string]interface{} para core.State
func mapToState(m map[string]interface{}) (core.State, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return core.State{}, err
	}
	
	var state core.State
	// Converter OrderState antigo para State novo
	var oldState struct {
		ID       string                 `json:"id,omitempty"`
		TenantID string                 `json:"tenantId"`
		Items    []struct {
			ID          string                 `json:"id,omitempty"`
			ProductID   string                 `json:"productId"`
			ProductName string                 `json:"productName,omitempty"`
			Quantity    float64                `json:"quantity"`
			Fields      map[string]interface{} `json:"fields"`
		} `json:"items"`
		Totals struct {
			Subtotal float64 `json:"subtotal,omitempty"`
			Discount float64 `json:"discount,omitempty"`
			Tax      float64 `json:"tax,omitempty"`
			Total    float64 `json:"total,omitempty"`
		} `json:"totals"`
		Fields map[string]interface{} `json:"fields"`
		Meta   map[string]interface{} `json:"meta,omitempty"`
	}
	
	if err := json.Unmarshal(data, &oldState); err != nil {
		return core.State{}, err
	}

	// Converter para novo formato
	state = core.State{
		ID:       oldState.ID,
		TenantID: oldState.TenantID,
		Fields:   oldState.Fields,
		Meta:     oldState.Meta,
		Totals: core.Totals{
			Subtotal: oldState.Totals.Subtotal,
			Discount: oldState.Totals.Discount,
			Tax:      oldState.Totals.Tax,
			Total:    oldState.Totals.Total,
		},
	}

	// Converter items
	state.Items = make([]core.Item, len(oldState.Items))
	for i, oldItem := range oldState.Items {
		state.Items[i] = core.Item{
			ID:     oldItem.ID,
			Amount: oldItem.Quantity, // Quantity → Amount
			Fields: oldItem.Fields,
		}
		// Preservar productId e productName em Fields se necessário
		if oldItem.ProductID != "" {
			if state.Items[i].Fields == nil {
				state.Items[i].Fields = make(map[string]interface{})
			}
			state.Items[i].Fields["productId"] = oldItem.ProductID
			if oldItem.ProductName != "" {
				state.Items[i].Fields["productName"] = oldItem.ProductName
			}
		}
	}

	return state, nil
}

// mapToRulePack converte map[string]interface{} para core.RulePack
func mapToRulePack(m map[string]interface{}) (core.RulePack, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return core.RulePack{}, err
	}
	
	var pack core.RulePack
	if err := json.Unmarshal(data, &pack); err != nil {
		return core.RulePack{}, err
	}
	
	return pack, nil
}

// convertReasons converte core.Reason para rulesengine.Reason
func convertReasons(reasons []core.Reason) []Reason {
	result := make([]Reason, len(reasons))
	for i, r := range reasons {
		result[i] = Reason{
			RuleID:  r.RuleID,
			Phase:   r.Phase,
			Message: r.Message,
		}
	}
	return result
}
```

## Opção 3: Migração Completa (Recomendado)

Substituir completamente o código antigo pela nova biblioteca:

### 1. Atualizar go.mod

```go
require (
    github.com/dolphin-sistemas/computations-engine v0.0.0
)

replace github.com/dolphin-sistemas/computations-engine => ../engine
```

### 2. Atualizar client.go

Substituir `internal/infra/rulesengine/client.go` completamente:

```go
package rulesengine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/dolphin-sistemas/computations-engine"
	"github.com/dolphin-sistemas/computations-engine/core"
)

type Client struct {
	logger *slog.Logger
}

func NewLocalClient() *Client {
	return &Client{logger: slog.Default()}
}

func (c *Client) ValidateOrder(
	ctx context.Context,
	tenantID, userID, rulesVersion string,
	orderState, rulePack map[string]interface{},
) (*ValidateOrderResult, error) {
	// Converter para tipos da nova biblioteca
	state, err := convertMapToState(orderState)
	if err != nil {
		return nil, fmt.Errorf("convert order state: %w", err)
	}

	pack, err := convertMapToRulePack(rulePack)
	if err != nil {
		return nil, fmt.Errorf("convert rule pack: %w", err)
	}

	// Executar motor
	stateFragment, serverDelta, reasons, rulesVersionOut, err := engine.RunEngine(
		ctx,
		state,
		pack,
		core.ContextMeta{
			TenantID: tenantID,
			UserID:   userID,
			Locale:   "pt-BR",
		},
	)
	if err != nil {
		c.logger.Error("RunEngine failed", "error", err)
		return nil, fmt.Errorf("rules engine: %w", err)
	}

	// Converter violations (se necessário obter do contexto)
	violations := []Violation{} // Implementar lógica se necessário

	return &ValidateOrderResult{
		StateFragment: stateFragment,
		ServerDelta:   serverDelta,
		Reasons:       convertReasons(reasons),
		RulesVersion:  rulesVersionOut,
		Violations:    violations,
	}, nil
}

func (c *Client) Close() error {
	return nil
}

// Funções auxiliares de conversão
func convertMapToState(m map[string]interface{}) (core.State, error) {
	// Implementar conversão de map para core.State
	// Tratar campos antigos (ProductID, Quantity) → novos (Amount)
	data, _ := json.Marshal(m)
	var state core.State
	json.Unmarshal(data, &state)
	return state, nil
}

func convertMapToRulePack(m map[string]interface{}) (core.RulePack, error) {
	data, _ := json.Marshal(m)
	var pack core.RulePack
	json.Unmarshal(data, &pack)
	return pack, nil
}

func convertReasons(reasons []core.Reason) []Reason {
	result := make([]Reason, len(reasons))
	for i, r := range reasons {
		result[i] = Reason{
			RuleID:  r.RuleID,
			Phase:   r.Phase,
			Message: r.Message,
		}
	}
	return result
}
```

## Diferenças Importantes

### Campos Renomeados

- `OrderState` → `State`
- `OrderItem` → `Item`
- `OrderTotals` → `Totals`
- `Quantity` → `Amount`
- Removidos: `ProductID`, `ProductName` (podem ser mantidos em `Fields`)

### API Mudou

- Antigo: `enginecore.Compute(input ComputeInput)`
- Novo: `engine.RunEngine(ctx, state State, rules RulePack, contextMeta ContextMeta)`

### itemTotals → itemValues

Nos RulePacks, atualizar referências:
- `{"var": ["itemTotals", []]}` → `{"var": ["itemValues", []]}`

## Testar a Integração

1. Executar testes do projeto:
```bash
cd teste-opa-jsonlogic-next
go test ./internal/infra/rulesengine/...
```

2. Testar endpoint:
```bash
curl -X POST http://localhost:8080/v1/orders \
  -H "Content-Type: application/json" \
  -d @test_order.json
```

## Próximos Passos

1. Adicionar replace no go.mod
2. Atualizar client.go para usar engine.RunEngine
3. Converter dados de entrada (map → State)
4. Atualizar RulePacks para usar `itemValues` ao invés de `itemTotals`
5. Testar e validar resultados
