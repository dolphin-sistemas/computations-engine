# Quick Start - Usar engine no teste-opa-jsonlogic-next

## Passos Rápidos

### 1. Adicionar Dependência

No arquivo `teste-opa-jsonlogic-next/go.mod`:

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

### 2. Atualizar Client

Substitua `internal/infra/rulesengine/client.go` pelo código em [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md).

**Mudança principal:**
- Antigo: `enginecore.Compute(input)`
- Novo: `engine.RunEngine(ctx, state, pack, contextMeta)`

### 3. Converter Dados

A função `convertOrderStateToState()` já está incluída no exemplo e faz:
- `Quantity` → `Amount`
- `ProductID`/`ProductName` → preservados em `Fields`

### 4. Testar

```bash
cd teste-opa-jsonlogic-next
go test ./internal/infra/rulesengine/...
```

## API da Biblioteca

```go
stateFragment, serverDelta, reasons, violations, rulesVersion, err := engine.RunEngine(
    ctx,
    state,      // core.State
    pack,       // core.RulePack
    contextMeta, // core.ContextMeta
)
```

## Retornos

- `stateFragment`: Campos que mudaram (para UI)
- `serverDelta`: Diferenças para sincronização
- `reasons`: Regras que executaram
- `violations`: Violações de validação
- `rulesVersion`: Versão das regras usadas

## Compatibilidade

- Tipos antigos (`OrderState`, `OrderItem`) funcionam via aliases
- Campos antigos (`ProductID`, `Quantity`) são convertidos automaticamente
- RulePacks existentes funcionam (apenas atualizar `itemTotals` → `itemValues` se necessário)
