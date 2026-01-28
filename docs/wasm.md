# Build WASM - Engine

## Resumo

A engine pode ser compilada para WASM usando os scripts fornecidos. O WASM expõe a função `runEngine` para JavaScript.

## Como Buildar

### Opção 1: Usando Makefile

```bash
cd engine
make wasm
```

### Opção 2: Usando Script Diretamente

```bash
cd engine
bash scripts/build-wasm.sh
```

## O que é Gerado

1. **`client/wasm/order_engine.wasm`**: Binário WASM compilado
2. **`client/wasm/wasm_exec.js`**: Runtime JavaScript do Go (copiado do GOROOT)

## Como Usar no JavaScript

```javascript
// Carregar WASM
const go = new Go();
const wasmModule = await WebAssembly.instantiateStreaming(
    fetch('client/wasm/order_engine.wasm'),
    go.importObject
);
go.run(wasmModule.instance);

// Usar a função
const input = {
    state: {
        tenantId: "tenant-1",
        items: [
            {
                id: "item-1",
                amount: 2,
                fields: {
                    basePrice: 100.0
                }
            }
        ],
        fields: {},
        totals: {}
    },
    rulePack: {
        id: "rules-1",
        version: "1.0",
        phases: [...]
    },
    context: {
        tenantId: "tenant-1",
        userId: "user-1",
        locale: "pt-BR"
    }
};

const resultJSON = runEngine(JSON.stringify(input));
const result = JSON.parse(resultJSON);

// Verificar erros
if (result.error) {
    console.error("Error:", result.error);
} else {
    console.log("State Fragment:", result.stateFragment);
    console.log("Violations:", result.violations);
}
```

## Estrutura do Retorno

### Sucesso

```json
{
  "stateFragment": {...},
  "serverDelta": {...},
  "reasons": [...],
  "violations": [],
  "rulesVersion": "1.0"
}
```

### Erro

```json
{
  "error": "error message"
}
```

## API da Função WASM

### `runEngine(inputJSON: string): string`

- **Input**: JSON string com `state`, `rulePack`, `context`
- **Output**: JSON string com `stateFragment`, `serverDelta`, `reasons`, `violations`, `rulesVersion` ou `error`

### Exemplo Completo

```javascript
// 1. Carregar wasm_exec.js
<script src="client/wasm/wasm_exec.js"></script>

// 2. Carregar e inicializar WASM
async function initWASM() {
    const go = new Go();
    const wasmModule = await WebAssembly.instantiateStreaming(
        fetch('client/wasm/order_engine.wasm'),
        go.importObject
    );
    go.run(wasmModule.instance);
    return true;
}

// 3. Usar após inicialização
await initWASM();

const input = {
    state: {
        tenantId: "tenant-1",
        items: [
            {
                id: "item-1",
                amount: 2,
                fields: { basePrice: 100.0 }
            }
        ],
        fields: {},
        totals: {}
    },
    rulePack: {
        id: "rules-1",
        version: "1.0",
        phases: [
            {
                name: "baseline",
                rules: [
                    {
                        id: "calc-total",
                        phase: "baseline",
                        condition: null,
                        actions: [
                            {
                                type: "compute",
                                target: "items[0].fields.total",
                                logic: {
                                    "*": [
                                        {"var": "items[0].fields.basePrice"},
                                        {"var": "items[0].amount"}
                                    ]
                                }
                            }
                        ]
                    }
                ]
            }
        ]
    },
    context: {
        tenantId: "tenant-1",
        userId: "user-1",
        locale: "pt-BR"
    }
};

const resultJSON = runEngine(JSON.stringify(input));
const result = JSON.parse(resultJSON);

if (result.error) {
    console.error("Error:", result.error);
} else {
    console.log("State Fragment:", result.stateFragment);
    console.log("Server Delta:", result.serverDelta);
    console.log("Reasons:", result.reasons);
    console.log("Violations:", result.violations);
    console.log("Rules Version:", result.rulesVersion);
}
```

## Tratamento de Erros

```javascript
try {
    const resultJSON = runEngine(JSON.stringify(input));
    const result = JSON.parse(resultJSON);
    
    if (result.error) {
        // Erro na execução da engine
        console.error("Engine error:", result.error);
        return;
    }
    
    if (result.violations && result.violations.length > 0) {
        // Há violações de validação
        result.violations.forEach(v => {
            console.error(`Field: ${v.field}, Code: ${v.code}, Message: ${v.message}`);
        });
        return;
    }
    
    // Sucesso
    console.log("Success:", result);
} catch (error) {
    // Erro ao parsear JSON ou chamar função
    console.error("Error:", error);
}
```

## Estrutura do Entry Point

O entry point WASM está em `cmd/wasm/main.go`:

```go
func RunEngineWASM(this js.Value, args []js.Value) interface{} {
    // 1. Parse input JSON
    // 2. Executar engine.RunEngine
    // 3. Retornar JSON string
}
```

## Notas

- O WASM expõe apenas a função `runEngine`
- Input e output são JSON strings
- O contexto é criado com `context.Background()` (não usa cancelamento)
- O WASM mantém todas as capacidades da engine (cálculos, validações, guards)
- O tamanho do WASM pode variar, mas geralmente fica entre 2-5MB

## Troubleshooting

### Erro: "runEngine is not defined"

- Certifique-se de que o WASM foi carregado e inicializado antes de chamar `runEngine`
- Verifique se `go.run(wasmModule.instance)` foi executado

### Erro: "failed to parse input"

- Verifique se o JSON de input está correto
- Certifique-se de que todos os campos obrigatórios estão presentes (`state`, `rulePack`, `context`)

### WASM não carrega

- Verifique se o servidor está servindo o arquivo `.wasm` com o content-type correto: `application/wasm`
- Verifique se o caminho do arquivo está correto
