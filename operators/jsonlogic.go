package operators

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/diegoholiveira/jsonlogic/v3"
)

const (
	MaxLogicSize = 50 * 1024 // 50KB
	MaxDepth     = 20
)

// EvaluateJsonLogic avalia uma expressão JsonLogic e retorna o resultado
func EvaluateJsonLogic(logic map[string]interface{}, data map[string]interface{}) (interface{}, error) {
	// Validar tamanho
	logicJSON, err := json.Marshal(logic)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal logic: %w", err)
	}
	if len(logicJSON) > MaxLogicSize {
		return nil, fmt.Errorf("logic exceeds maximum size of %d bytes", MaxLogicSize)
	}

	// Validar profundidade
	if err := validateDepth(logic, 0); err != nil {
		return nil, fmt.Errorf("logic exceeds maximum depth: %w", err)
	}

	// Avaliar JsonLogic
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	logicReader := bytes.NewReader(logicJSON)
	dataReader := bytes.NewReader(dataJSON)

	var resultBuffer bytes.Buffer
	err = jsonlogic.Apply(logicReader, dataReader, &resultBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to apply jsonlogic: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(resultBuffer.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return result, nil
}

// validateDepth valida a profundidade máxima da lógica (prevenir DoS)
func validateDepth(logic interface{}, currentDepth int) error {
	if currentDepth > MaxDepth {
		return fmt.Errorf("maximum depth of %d exceeded", MaxDepth)
	}

	switch v := logic.(type) {
	case map[string]interface{}:
		for _, val := range v {
			if err := validateDepth(val, currentDepth+1); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, item := range v {
			if err := validateDepth(item, currentDepth+1); err != nil {
				return err
			}
		}
	}

	return nil
}
