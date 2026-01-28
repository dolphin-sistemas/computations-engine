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
	stateData, rulePackData map[string]interface{},
) (*ValidateOrderResult, error) {
	// Converter map para core.State
	stateJSON, err := json.Marshal(stateData)
	if err != nil {
		return nil, fmt.Errorf("marshal state: %w", err)
	}
	
	var state core.State
	if err := json.Unmarshal(stateJSON, &state); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
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
		c.logger.Error("RulesEngine RunEngine failed", "error", err)
		return nil, fmt.Errorf("rules engine: %w", err)
	}

	// Converter violations e reasons
	violations := make([]Violation, len(result.Violations))
	for i, v := range result.Violations {
		violations[i] = Violation{
			Field:   v.Field,
			Code:    v.Code,
			Message: v.Message,
		}
	}
	reasonsOut := make([]Reason, len(result.Reasons))
	for i, r := range result.Reasons {
		reasonsOut[i] = Reason{
			RuleID:  r.RuleID,
			Phase:   r.Phase,
			Message: r.Message,
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
```

## Opção 2: Adapter

Criar um adapter que usa a biblioteca engine:

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

// Adapter adapta a biblioteca engine para a interface do projeto
type Adapter struct {
	// Pode adicionar cache, logger, etc.
}

// ValidateOrder adapta a chamada para usar engine.RunEngine
func (a *Adapter) ValidateOrder(
	ctx context.Context,
	tenantID, userID, rulesVersion string,
	stateData, rulePackData map[string]interface{},
) (*ValidateOrderResult, error) {
	// Converter map[string]interface{} para core.State
	state, err := mapToState(stateData)
	if err != nil {
		return nil, fmt.Errorf("convert state: %w", err)
	}

	// Converter map[string]interface{} para core.RulePack
	pack, err := mapToRulePack(rulePack)
	if err != nil {
		return nil, fmt.Errorf("convert rule pack: %w", err)
	}

	// Executar motor
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
		return nil, fmt.Errorf("engine execution: %w", err)
	}

	// Converter resultados
	return &ValidateOrderResult{
		StateFragment: result.StateFragment,
		ServerDelta:   result.ServerDelta,
		Reasons:       convertReasons(result.Reasons),
		RulesVersion:  result.RulesVersion,
		Violations:    convertViolations(result.Violations),
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

## Opção 3: Implementação Completa (Recomendado)

Implementação completa usando a biblioteca engine:

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
	stateData, rulePackData map[string]interface{},
) (*ValidateOrderResult, error) {
	// Converter map para core.State
	state, err := convertMapToState(stateData)
	if err != nil {
		return nil, fmt.Errorf("convert state: %w", err)
	}

	pack, err := convertMapToRulePack(rulePack)
	if err != nil {
		return nil, fmt.Errorf("convert rule pack: %w", err)
	}

	// Executar motor
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
		Reasons:       convertReasons(result.Reasons),
		RulesVersion:  result.RulesVersion,
		Violations:    violations,
	}, nil
}

func (c *Client) Close() error {
	return nil
}

// Funções auxiliares de conversão
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

### 3. Testar

```bash
go test ./internal/rulesengine/...
go run cmd/main.go
```

## API da Engine

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
