# Guia de Migração - teste-opa-jsonlogic-next

## Resumo

Substituir `internal/infra/rulesengine/enginecore` pela biblioteca `github.com/dolphin-sistemas/computations-engine`.

## Passos

### 1. Atualizar go.mod

Adicione no `teste-opa-jsonlogic-next/go.mod`:

```go
require (
    github.com/dolphin-sistemas/computations-engine v0.0.0
)

replace github.com/dolphin-sistemas/computations-engine => ../engine
```

Execute:
```bash
cd teste-opa-jsonlogic-next
go mod tidy
```

### 2. Substituir client.go

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

// ValidateOrderResult representa o resultado da validação
type ValidateOrderResult struct {
	StateFragment map[string]interface{} `json:"state_fragment"`
	ServerDelta   map[string]interface{} `json:"server_delta"`
	Reasons       []Reason               `json:"reasons"`
	RulesVersion  string                 `json:"rules_version"`
	Violations    []Violation            `json:"violations"`
}

// Violation representa uma violação de validação
type Violation struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Reason representa uma razão de execução de regra
type Reason struct {
	RuleID  string `json:"rule_id"`
	Phase   string `json:"phase"`
	Message string `json:"message"`
}

// Client wraps the engine library
type Client struct {
	logger *slog.Logger
}

// NewLocalClient cria um client que executa o motor em-processo
func NewLocalClient() *Client {
	return &Client{logger: slog.Default()}
}

// ValidateOrder valida/computa um pedido usando a biblioteca engine
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
	stateFragment, serverDelta, reasons, violationsOut, rulesVersionOut, err := engine.RunEngine(
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

	// Converter violations
	violations := make([]Violation, len(violationsOut))
	for i, v := range violationsOut {
		violations[i] = Violation{
			Field:   v.Field,
			Code:    v.Code,
			Message: v.Message,
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

// Close fecha o client (no-op para local)
func (c *Client) Close() error {
	return nil
}

// convertOrderStateToState converte map[string]interface{} (formato antigo) para core.State
// Mantém compatibilidade com campos antigos (ProductID, Quantity, etc.)
func convertOrderStateToState(m map[string]interface{}) (core.State, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return core.State{}, err
	}

	// Estrutura temporária para ler formato antigo (com ProductID, Quantity)
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
		// Preservar campos antigos em Fields para compatibilidade com regras existentes
		if oldItem.ProductID != "" {
			fields["productId"] = oldItem.ProductID
		}
		if oldItem.ProductName != "" {
			fields["productName"] = oldItem.ProductName
		}
		// Preservar quantity também para regras que ainda usam
		if oldItem.Quantity > 0 {
			fields["quantity"] = oldItem.Quantity
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

### 3. Remover enginecore (Opcional)

Após validar que tudo funciona, você pode remover:
- `internal/infra/rulesengine/enginecore/` (código antigo)

### 4. Atualizar RulePacks (se necessário)

Se seus RulePacks JSON usam `itemTotals`, atualize para `itemValues`:

**Buscar e substituir:**
```json
"itemTotals" → "itemValues"
```

### 5. Testar

```bash
cd teste-opa-jsonlogic-next
go test ./internal/infra/rulesengine/...
go run cmd/main.go
```

## Compatibilidade

A biblioteca mantém compatibilidade com tipos antigos via aliases:
- `OrderState = State`
- `OrderItem = Item`  
- `OrderTotals = Totals`

Mas os campos `ProductID`, `ProductName`, `Quantity` precisam ser convertidos:
- `Quantity` → `Amount`
- `ProductID`/`ProductName` → preservados em `Fields` para compatibilidade

## Diferenças de API

**Antigo:**
```go
input := enginecore.ComputeInput{
    Order:    order,
    RulePack: pack,
    Context:  context,
}
out, err := enginecore.Compute(input)
```

**Novo:**
```go
stateFragment, serverDelta, reasons, violations, rulesVersion, err := engine.RunEngine(
    ctx,
    state,
    pack,
    contextMeta,
)
```

## Suporte a Violations

Se você precisa de violations, pode precisar modificar a biblioteca engine para expor violations do contexto, ou implementar uma lógica adicional no adapter.
