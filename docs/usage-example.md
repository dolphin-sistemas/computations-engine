# Exemplo de Uso

Este documento mostra um exemplo prático de como integrar a engine em seu projeto.

## Passo a Passo

### 1. Adicionar Dependência

No arquivo `go.mod` do seu projeto:

```go
require (
    // ... outras dependências existentes
    github.com/dolphin-sistemas/computations-engine v0.0.0
)

replace github.com/dolphin-sistemas/computations-engine => ../engine
```

Execute:
```bash
go mod tidy
```

### 2. Criar Client

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
	stateData, rulePackData map[string]interface{},
) (*ValidateOrderResult, error) {
	// Converter map para core.State
	state, err := convertMapToState(stateData)
	if err != nil {
		return nil, fmt.Errorf("convert state: %w", err)
	}

	// Converter map para core.RulePack
	pack, err := convertMapToRulePack(rulePackData)
	if err != nil {
		return nil, fmt.Errorf("convert rule pack: %w", err)
	}

	// Executar motor usando a biblioteca
	result, err := engine.RunEngine(
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
	reasonsOut := make([]Reason, len(result.Reasons))
	for i, r := range result.Reasons {
		reasonsOut[i] = Reason{
			RuleID:  r.RuleID,
			Phase:   r.Phase,
			Message: r.Message,
		}
	}

	// Converter violations
	violations := make([]Violation, len(result.Violations))
	for i, v := range result.Violations {
		violations[i] = Violation{
			Field:   v.Field,
			Code:    v.Code,
			Message: v.Message,
		}
	}

	return &ValidateOrderResult{
		StateFragment: result.StateFragment,
		ServerDelta:   result.ServerDelta,
		Reasons:       reasonsOut,
		RulesVersion:  result.RulesVersion,
		Violations:    violations,
	}, nil
}

func (c *Client) Close() error {
	return nil
}

// convertMapToState converte map[string]interface{} para core.State
func convertMapToState(m map[string]interface{}) (core.State, error) {
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

### 3. Testar

```bash
go test ./internal/rulesengine/...
go run cmd/main.go
```

