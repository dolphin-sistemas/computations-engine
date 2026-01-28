# Guia de Integração

Este guia mostra como integrar a biblioteca `engine` em seu projeto Go.

## Opção 1: Usar como Módulo Local (Desenvolvimento)

### 1. Adicionar dependência no go.mod

No arquivo `go.mod` do seu projeto, adicione:

```go
module seu-projeto

go 1.24.0

require (
    // ... outras dependências
    github.com/dolphin-sistemas/computations-engine v0.0.0
)

// Para desenvolvimento local, use replace:
replace github.com/dolphin-sistemas/computations-engine => ../engine
```

### 2. Criar ou Atualizar o Client

Crie um client que usa a engine. Exemplo em `internal/rulesengine/client.go`:

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

// ValidateOrder valida/computa dados usando a biblioteca engine
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
		c.logger.Error("RulesEngine RunEngine failed", "error", err)
		return nil, fmt.Errorf("rules engine: %w", err)
	}

	// Converter violations e reasons
	violations := make([]Violation, len(violationsOut))
	for i, v := range violationsOut {
		violations[i] = Violation{
			Field:   v.Field,
			Code:    v.Code,
			Message: v.Message,
		}
	}
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

### Criar um Adapter

Exemplo em `internal/rulesengine/adapter.go`:

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
		return nil, fmt.Errorf("engine execution: %w", err)
	}

	// Converter resultados
	return &ValidateOrderResult{
		StateFragment: stateFragment,
		ServerDelta:   serverDelta,
		Reasons:       convertReasons(reasons),
		RulesVersion:  rulesVersionOut,
		Violations:    convertViolations(violationsOut),
	}, nil
}

// mapToState converte map[string]interface{} para core.State
func mapToState(m map[string]interface{}) (core.State, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return core.State{}, err
	}
	
	var state core.State
	if err := json.Unmarshal(data, &state); err != nil {
		return core.State{}, err
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

// convertViolations converte core.Violation para rulesengine.Violation
func convertViolations(violations []core.Violation) []Violation {
	result := make([]Violation, len(violations))
	for i, v := range violations {
		result[i] = Violation{
			Field:   v.Field,
			Code:    v.Code,
			Message: v.Message,
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

### 2. Criar Client

Exemplo completo de client em `internal/rulesengine/client.go`:

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

### 3. Remover código antigo (Opcional)

Após validar que tudo funciona, você pode remover qualquer implementação antiga da engine que não seja mais necessária.

### 4. Atualizar RulePacks (se necessário)

Se seus RulePacks JSON usam `itemTotals`, atualize para `itemValues`:

**Buscar e substituir:**
```json
"itemTotals" → "itemValues"
```

### 5. Testar

```bash
go test ./internal/rulesengine/...
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

Violations são retornadas diretamente pela função `RunEngine`:

```go
violations := make([]Violation, len(violationsOut))
for i, v := range violationsOut {
    violations[i] = Violation{
        Field:   v.Field,
        Code:    v.Code,
        Message: v.Message,
    }
}
```
