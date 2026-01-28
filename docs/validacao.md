# Validação na Engine

## Capacidades de Validação

A engine pode validar **qualquer coisa** usando JsonLogic:

### 1. **Cálculos** (Fase `baseline`, `allocation`, `taxes`, `totals`)
- Operações matemáticas: `+`, `-`, `*`, `/`, `%`
- Operadores customizados: `sum`, `round`, `round2`, `allocate`, `if`, `foreach`
- Ações: `set`, `compute`, `add`, `multiply`

### 2. **Validações de Campos** (Fase `guards`)
- Validações condicionais com JsonLogic
- Validações de layout dinâmico (required, max, min, pattern)
- Validações de negócio (desconto máximo, itens obrigatórios, etc.)

## Como Funciona a Validação

### 1. Validação via Actions (Tipo `validate`)

```json
{
  "type": "validate",
  "logic": {
    "==": [{"var": "fields.discountPercent"}, null]
  },
  "params": {
    "field": "fields.discountPercent",
    "code": "REQUIRED",
    "message": "Discount percent is required"
  }
}
```

**Como funciona:**
- Avalia `logic` (JsonLogic)
- Se retornar `true` → cria `Violation`
- Se retornar `false` → não cria violação

### 2. Validação via Guards (Fase `guards`)

```json
{
  "name": "guards",
  "rules": [
    {
      "id": "max-discount",
      "condition": {
        ">": [
          {"var": "totals.discount"},
          {"*": [{"var": "totals.subtotal"}, 0.5]}
        ]
      },
      "actions": [
        {
          "type": "validate",
          "logic": true,
          "params": {
            "field": "totals.discount",
            "code": "MAX_DISCOUNT_EXCEEDED",
            "message": "Discount cannot exceed 50%"
          }
        }
      ]
    }
  ]
}
```

**Como funciona:**
- Avalia `condition` (JsonLogic)
- Se `true` → executa `actions`
- Actions do tipo `validate` criam `Violation`

### 3. Validação de Layout Dinâmico

Para validar campos required, max, min, pattern, você pode criar regras:

```json
{
  "id": "validate-email",
  "phase": "guards",
  "condition": {
    "and": [
      {"var": "fields.customerEmail"},
      {"!": [{"var": "fields.customerEmail"}]}
    ]
  },
  "actions": [
    {
      "type": "validate",
      "logic": {
        "!": [
          {"var": "fields.customerEmail"}
        ]
      },
      "params": {
        "field": "fields.customerEmail",
        "code": "INVALID_EMAIL",
        "message": "Email is required"
      }
    }
  ]
}
```

## Retorno de Erro

### Estrutura de Violation

```go
type Violation struct {
    Field   string `json:"field"`   // Campo que falhou (ex: "fields.discountPercent")
    Code    string `json:"code"`    // Código do erro (ex: "REQUIRED", "MAX_DISCOUNT")
    Message string `json:"message"` // Mensagem de erro
}
```

### Exemplo de Retorno com Erros

```go
stateFragment, serverDelta, reasons, violations, rulesVersion, err := engine.RunEngine(
    ctx,
    state,
    rulePack,
    contextMeta,
)

// Se houver violations
if len(violations) > 0 {
    // violations contém todas as violações encontradas
    for _, v := range violations {
        fmt.Printf("Field: %s, Code: %s, Message: %s\n", 
            v.Field, v.Code, v.Message)
    }
}
```

### JSON de Retorno (com erros)

```json
{
  "state_fragment": {...},
  "server_delta": {...},
  "reasons": [...],
  "rules_version": "1.0",
  "violations": [
    {
      "field": "fields.discountPercent",
      "code": "REQUIRED",
      "message": "Discount percent is required"
    },
    {
      "field": "totals.discount",
      "code": "MAX_DISCOUNT_EXCEEDED",
      "message": "Discount cannot exceed 50%"
    }
  ]
}
```

## Retorno de Sucesso

### Quando Não Há Violations

```go
stateFragment, serverDelta, reasons, violations, rulesVersion, err := engine.RunEngine(
    ctx,
    state,
    rulePack,
    contextMeta,
)

// Sucesso: err == nil && len(violations) == 0
if err == nil && len(violations) == 0 {
    // stateFragment: campos que mudaram (para atualizar UI)
    // serverDelta: diferenças para sincronização
    // reasons: regras que executaram
    // rulesVersion: versão das regras usadas
}
```

### JSON de Retorno (sucesso)

```json
{
  "state_fragment": {
    "totals": {
      "subtotal": 100.0,
      "discount": 10.0,
      "tax": 9.0,
      "total": 99.0
    },
    "fields": {
      "calculatedAt": "2026-01-26T10:00:00Z"
    }
  },
  "server_delta": {
    "totals.total": 99.0,
    "totals.discount": 10.0
  },
  "reasons": [
    {
      "ruleId": "apply-discount",
      "phase": "allocation",
      "message": "Applied 10% discount"
    }
  ],
  "rules_version": "1.0",
  "violations": []
}
```

## Fluxo de Validação

1. **Pipeline executa fases** em ordem: `baseline → allocation → taxes → totals → guards`
2. **Cada fase** executa regras em ordem de prioridade
3. **Cada regra** avalia `condition` (JsonLogic)
4. **Se condition = true**, executa `actions`
5. **Actions do tipo `validate`** criam `Violation` se `logic` retornar `true`
6. **Violations são acumuladas** no `EngineContext`
7. **Ao final**, `RunEngine` retorna todas as violations

## Exemplos de Validações

### Required Field

```json
{
  "type": "validate",
  "logic": {
    "==": [{"var": "fields.customerEmail"}, null]
  },
  "params": {
    "field": "fields.customerEmail",
    "code": "REQUIRED",
    "message": "Customer email is required"
  }
}
```

### Max Value

```json
{
  "type": "validate",
  "logic": {
    ">": [
      {"var": "totals.discount"},
      1000.0
    ]
  },
  "params": {
    "field": "totals.discount",
    "code": "MAX_VALUE_EXCEEDED",
    "message": "Discount cannot exceed 1000.00"
  }
}
```

### Pattern (Regex via JsonLogic)

```json
{
  "type": "validate",
  "logic": {
    "if": [
      {"var": "fields.customerEmail"},
      {
        "!": [
          {"var": ["fields.customerEmail", {"regex": "^[^@]+@[^@]+\\.[^@]+$"}]}
        ]
      },
      false
    ]
  },
  "params": {
    "field": "fields.customerEmail",
    "code": "INVALID_FORMAT",
    "message": "Invalid email format"
  }
}
```

### Conditional Validation

```json
{
  "type": "validate",
  "logic": {
    "and": [
      {"==": [{"var": "fields.paymentMethod"}, "BOLETO"]},
      {"==": [{"var": "fields.paymentTermDays"}, null]}
    ]
  },
  "params": {
    "field": "fields.paymentTermDays",
    "code": "REQUIRED_FOR_PAYMENT_METHOD",
    "message": "Payment term days is required for BOLETO"
  }
}
```

## Notas Importantes

1. **Violations não param a execução**: Todas as regras são executadas, violations são acumuladas
2. **Ordem importa**: Fases executam em ordem, regras dentro de uma fase executam por prioridade
3. **JsonLogic é poderoso**: Qualquer lógica pode ser expressa em JsonLogic
4. **Violations são retornadas mesmo em sucesso**: Verifique `len(violations) == 0` para sucesso
