package operators

// RegisterAllOperators registra todos os operadores customizados
func RegisterAllOperators() {
	// Operadores matemáticos
	registerMathOperators()

	// Operador condicional
	registerConditionalOperator()

	// Operador de iteração
	registerIterationOperator()

	// Operador de alocação
	registerAllocationOperator()
}

// registerMathOperators registra operadores matemáticos (sum, round, round2)
func registerMathOperators() {
	// sum já está implementado em math.go
	// round e round2 já estão implementados em math.go
	// Esta função é chamada por init() em math.go
}

// registerConditionalOperator registra operador "if"
func registerConditionalOperator() {
	// Implementado em conditional.go via init()
}

// registerIterationOperator registra operador "foreach"
func registerIterationOperator() {
	// Implementado em iteration.go via init()
}

// registerAllocationOperator registra operador "allocate"
func registerAllocationOperator() {
	// Implementado em allocation.go via init()
}

// init registra todos os operadores quando o pacote é importado
func init() {
	RegisterAllOperators()
}
