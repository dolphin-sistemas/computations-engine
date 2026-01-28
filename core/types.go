package core

// State representa o estado genérico dos dados sendo processados
type State struct {
	ID       string                 `json:"id,omitempty"`
	TenantID string                 `json:"tenantId,omitempty"`
	Items    []Item                 `json:"items,omitempty"`  // Coleção de itens
	Totals   Totals                 `json:"totals,omitempty"`  // Totais/sumário
	Fields   map[string]interface{} `json:"fields,omitempty"`  // Campos customizáveis
	Meta     map[string]interface{} `json:"meta,omitempty"`    // Metadados (não afetam cálculo)
}

// Item representa um item genérico em uma coleção
type Item struct {
	ID     string                 `json:"id,omitempty"`
	Amount float64                `json:"amount,omitempty"`  // Quantidade/valor base
	Fields map[string]interface{} `json:"fields,omitempty"`  // Campos customizáveis
}

// Totals representa totais/sumário calculados
type Totals struct {
	Subtotal float64 `json:"subtotal,omitempty"`
	Discount float64 `json:"discount,omitempty"`
	Tax      float64 `json:"tax,omitempty"`
	Total    float64 `json:"total,omitempty"`
}

// RulePack representa um pacote de regras versionado
type RulePack struct {
	ID      string     `json:"id"`
	Version string     `json:"version"`
	Phases  []RulePhase `json:"phases"`
}

// RulePhase representa uma fase de processamento (baseline, allocation, taxes, totals, validations, guards, etc.)
type RulePhase struct {
	Name  string `json:"name"`
	Rules []Rule `json:"rules"`
}

// Rule representa uma regra individual com condição e ações
type Rule struct {
	ID        string                 `json:"id"`
	Phase     string                 `json:"phase"`
	Condition map[string]interface{} `json:"condition"` // JsonLogic para avaliar se a regra deve executar
	Actions   []Action               `json:"actions"`   // Lista de ações a executar se condition for true
	Priority  int                    `json:"priority,omitempty"` // Prioridade dentro da phase (menor = executa antes)
	Enabled   bool                   `json:"enabled,omitempty"`
}

// Action representa uma ação a ser executada (DSL simples)
type Action struct {
	Type   string                 `json:"type"`   // "set", "compute", "validate", "add", etc.
	Target string                 `json:"target"` // Ex: "fields.totalPrice", "items[0].fields.value", "totals.total"
	Value  interface{}            `json:"value,omitempty"`  // Valor literal ou JsonLogic
	Logic  map[string]interface{} `json:"logic,omitempty"`  // JsonLogic para calcular o valor
	Params map[string]interface{} `json:"params,omitempty"` // Parâmetros adicionais da ação
}

// ContextMeta representa metadados de contexto (não afeta cálculo, apenas disponível para regras)
type ContextMeta struct {
	TenantID string `json:"tenantId"`
	UserID   string `json:"userId,omitempty"`
	Locale   string `json:"locale,omitempty"`
}

// Reason rastreia qual regra executou e por quê
type Reason struct {
	RuleID  string `json:"ruleId"`
	Phase   string `json:"phase"`
	Message string `json:"message,omitempty"`
}

// Violation representa uma violação de validação
type Violation struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RunEngineResult representa o resultado da execução do motor
type RunEngineResult struct {
	StateFragment map[string]interface{} `json:"stateFragment"` // Campos que mudaram (para atualizar UI)
	ServerDelta   map[string]interface{} `json:"serverDelta"`    // Diferenças para sincronização
	Reasons       []Reason               `json:"reasons"`         // Regras que executaram
	Violations    []Violation            `json:"violations"`      // Violações de validação
	RulesVersion  string                 `json:"rulesVersion"`    // Versão das regras usadas
}
