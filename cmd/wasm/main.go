//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"encoding/json"
	"syscall/js"

	"github.com/dolphin-sistemas/computations-engine"
	"github.com/dolphin-sistemas/computations-engine/core"
)

// RunEngineWASM é a função exposta para JavaScript
// Recebe JSON string e retorna JSON string
func RunEngineWASM(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		result, _ := json.Marshal(map[string]interface{}{
			"error": "missing input argument",
		})
		return string(result)
	}

	inputJSON := args[0].String()

	// Parse input
	var input struct {
		State    core.State       `json:"state"`
		RulePack core.RulePack    `json:"rulePack"`
		Context  core.ContextMeta `json:"context"`
	}

	if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
		result, _ := json.Marshal(map[string]interface{}{
			"error": "failed to parse input: " + err.Error(),
		})
		return string(result)
	}

	// Executar engine
	result, err := engine.RunEngine(
		context.Background(),
		input.State,
		input.RulePack,
		input.Context,
	)

	if err != nil {
		result, _ := json.Marshal(map[string]interface{}{
			"error": err.Error(),
		})
		return string(result)
	}

	// Converter resultado para JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		result, _ := json.Marshal(map[string]interface{}{
			"error": "failed to marshal result: " + err.Error(),
		})
		return string(result)
	}

	return string(resultJSON)
}

func main() {
	// Registrar função global
	js.Global().Set("runEngine", js.FuncOf(RunEngineWASM))

	// Manter o programa rodando
	select {}
}
