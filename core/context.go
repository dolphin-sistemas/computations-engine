package core

import (
	"encoding/json"
)

// EngineContext mantém o estado interno do motor durante a execução
type EngineContext struct {
	State      *State
	Original   State  // Cópia original para calcular deltas
	Context    ContextMeta
	Reasons    []Reason
	Violations []Violation
	PhaseIndex int // Índice da fase atual
}

// NewEngineContext cria um novo contexto do motor
func NewEngineContext(state State, context ContextMeta) (*EngineContext, error) {
	// Clonar estado original
	original, err := cloneState(state)
	if err != nil {
		return nil, err
	}

	// Inicializar campos vazios
	if state.Items == nil {
		state.Items = []Item{}
	}
	if state.Fields == nil {
		state.Fields = make(map[string]interface{})
	}

	return &EngineContext{
		State:      &state,
		Original:   original,
		Context:    context,
		Reasons:    []Reason{},
		Violations: []Violation{},
		PhaseIndex: 0,
	}, nil
}

// cloneState faz uma cópia profunda do State
func cloneState(s State) (State, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return State{}, err
	}

	var cloned State
	if err := json.Unmarshal(data, &cloned); err != nil {
		return State{}, err
	}

	return cloned, nil
}
