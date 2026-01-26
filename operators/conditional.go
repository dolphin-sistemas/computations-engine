package operators

import (
	"github.com/diegoholiveira/jsonlogic/v3"
)

func init() {
	// Registrar operador "if": {"if": [cond, trueVal, falseVal]}
	// Nota: cond, trueVal e falseVal podem ser valores literais ou expressões JsonLogic
	jsonlogic.AddOperator("if", func(values, data interface{}) interface{} {
		var cond interface{}
		var trueVal interface{}
		var falseVal interface{}

		switch v := values.(type) {
		case []interface{}:
			if len(v) == 0 {
				return nil
			}
			cond = v[0]
			if len(v) > 1 {
				trueVal = v[1]
			}
			if len(v) > 2 {
				falseVal = v[2]
			}
		default:
			return nil
		}

		// Avaliar condição (pode ser JsonLogic ou valor literal)
		condResult := evaluateCondition(cond, data)

		if condResult {
			// Retornar trueVal (pode ser JsonLogic ou literal)
			return evaluateValue(trueVal, data)
		}

		// Retornar falseVal (pode ser JsonLogic ou literal)
		return evaluateValue(falseVal, data)
	})
}

// evaluateCondition avalia uma condição (JsonLogic ou valor literal)
func evaluateCondition(cond interface{}, data interface{}) bool {
	// Se for JsonLogic, avaliar
	if logic, ok := cond.(map[string]interface{}); ok {
		if dataMap, ok := data.(map[string]interface{}); ok {
			result, err := EvaluateJsonLogic(logic, dataMap)
			if err == nil {
				return isTruthy(result)
			}
		}
		return false
	}

	// Valor literal - verificar truthiness
	return isTruthy(cond)
}

// evaluateValue avalia um valor (JsonLogic ou literal)
func evaluateValue(val interface{}, data interface{}) interface{} {
	// Se for JsonLogic, avaliar
	if logic, ok := val.(map[string]interface{}); ok {
		if dataMap, ok := data.(map[string]interface{}); ok {
			result, err := EvaluateJsonLogic(logic, dataMap)
			if err == nil {
				return result
			}
		}
		// Em caso de erro, retornar valor original
		return val
	}

	// Valor literal - retornar como está
	return val
}

// isTruthy verifica se um valor é truthy
func isTruthy(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case int:
		return val != 0
	case string:
		return val != ""
	case []interface{}:
		return len(val) > 0
	case map[string]interface{}:
		return len(val) > 0
	}
	return true
}
