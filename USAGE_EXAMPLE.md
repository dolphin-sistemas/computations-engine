# Exemplo de Uso - Integração com teste-opa-jsonlogic-next

## Passo a Passo

### 1. Adicionar Dependência Local

No arquivo `teste-opa-jsonlogic-next/go.mod`:

```go
require (
    // ... outras dependências existentes
    github.com/dolphin-sistemas/computations-engine v0.0.0
)

replace github.com/dolphin-sistemas/computations-engine => ../engine
```

Execute:
```bash
cd teste-opa-jsonlogic-next
go mod tidy
```

### 2. Atualizar Client

Substitua o conteúdo de `internal/infra/rulesengine/client.go`:

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

// Client wraps the engine library
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
	// Converter orderState (map) para core.State
	state, err := convertOrderStateToState(orderState)
	if err != nil {
		return nil, fmt.Errorf("convert order state: %w", err)
	}

	// Converter rulePack (map) para core.RulePack
	pack, err := convertMapToRulePack(rulePack)
	if err != nil {
		return nil, fmt.Errorf("convert rule pack: %w", err)
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
		c.logger.Error("RunEngine failed", "error", err)
		return nil, fmt.Errorf("rules engine: %w", err)
	}

	// Converter reasons
	reasonsOut := make([]Reason, len(reasons))
	for i, r := range reasons {
		reasonsOut[i] = Reason{
			RuleID:  r.RuleID,
			Phase:   r.Phase,
			Message: r.Message,
		}
	}

	// Nota: Violations precisam ser obtidas de outra forma
	// Por enquanto, retornar vazio ou implementar lógica adicional
	violations := []Violation{}

	return &ValidateOrderResult{
		StateFragment: stateFragment,
		ServerDelta:   serverDelta,
		Reasons:       reasonsOut,
		RulesVersion:  rulesVersionOut,
		Violations:    violations,
	}, nil
}

func (c *Client) Close() error {
	return nil
}

// convertOrderStateToState converte map[string]interface{} (formato antigo) para core.State
func convertOrderStateToState(m map[string]interface{}) (core.State, error) {
	// Converter via JSON para manter compatibilidade
	data, err := json.Marshal(m)
	if err != nil {
		return core.State{}, err
	}

	// Estrutura temporária para ler formato antigo
	var oldFormat struct {
		ID       string                 `json:"id,omitempty"`
		TenantID string                 `json:"tenantId"`
		Items    []struct {
			ID          string                 `json:"id,omitempty"`
			ProductID   string                 `json:"productId,omitempty"`
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

	if err := json.Unmarshal(data, &oldFormat); err != nil {
		return core.State{}, err
	}

	// Converter para novo formato
	state := core.State{
		ID:       oldFormat.ID,
		TenantID: oldFormat.TenantID,
		Fields:   oldFormat.Fields,
		Meta:     oldFormat.Meta,
		Totals: core.Totals{
			Subtotal: oldFormat.Totals.Subtotal,
			Discount: oldFormat.Totals.Discount,
			Tax:      oldFormat.Totals.Tax,
			Total:    oldFormat.Totals.Total,
		},
		Items: make([]core.Item, len(oldFormat.Items)),
	}

	// Converter items: Quantity → Amount, preservar ProductID/ProductName em Fields
	for i, oldItem := range oldFormat.Items {
		fields := make(map[string]interface{})
		if oldItem.Fields != nil {
			for k, v := range oldItem.Fields {
				fields[k] = v
			}
		}
		// Preservar campos antigos em Fields para compatibilidade
		if oldItem.ProductID != "" {
			fields["productId"] = oldItem.ProductID
		}
		if oldItem.ProductName != "" {
			fields["productName"] = oldItem.ProductName
		}
		if oldItem.Quantity > 0 {
			fields["quantity"] = oldItem.Quantity // Preservar para regras antigas
		}

		state.Items[i] = core.Item{
			ID:     oldItem.ID,
			Amount: oldItem.Quantity, // Quantity → Amount
			Fields: fields,
		}
	}

	return state, nil
}

// convertMapToRulePack converte map para core.RulePack
func convertMapToRulePack(m map[string]interface{}) (core.RulePack, error) {
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
```

### 3. Atualizar RulePacks (se necessário)

Se seus RulePacks usam `itemTotals`, atualize para `itemValues`:

**Antes:**
```json
{
  "logic": {
    "sum": [{"var": ["itemTotals", []]}]
  }
}
```

**Depois:**
```json
{
  "logic": {
    "sum": [{"var": ["itemValues", []]}]
  }
}
```

### 4. Testar

```bash
cd teste-opa-jsonlogic-next
go test ./internal/infra/rulesengine/...
go run cmd/main.go
```

## Compatibilidade

A biblioteca mantém tipos alias para compatibilidade:
- `OrderState = State`
- `OrderItem = Item`
- `OrderTotals = Totals`

Mas recomenda-se migrar para os novos nomes genéricos.
