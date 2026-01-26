package operators

import (
	"github.com/diegoholiveira/jsonlogic/v3"
)

func init() {
	// Registrar operador "foreach": {"foreach": [array, logic]}
	// Itera sobre array aplicando logic a cada elemento
	jsonlogic.AddOperator("foreach", func(values, data interface{}) interface{} {
		var arr []interface{}
		var logic map[string]interface{}

		switch v := values.(type) {
		case []interface{}:
			if len(v) < 2 {
				return []interface{}{}
			}
			// Primeiro argumento é o array
			switch a := v[0].(type) {
			case []interface{}:
				arr = a
			default:
				return []interface{}{}
			}
			// Segundo argumento é a lógica
			if l, ok := v[1].(map[string]interface{}); ok {
				logic = l
			} else {
				return []interface{}{}
			}
		default:
			return []interface{}{}
		}

		// Converter data para map se necessário
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			dataMap = make(map[string]interface{})
		}

		// Iterar sobre array aplicando logic
		result := make([]interface{}, len(arr))
		for i, item := range arr {
			// Criar contexto para este item
			itemData := make(map[string]interface{})
			for k, v := range dataMap {
				itemData[k] = v
			}
			itemData["item"] = item
			itemData["index"] = float64(i)

			// Avaliar logic com contexto do item
			evaluated, err := EvaluateJsonLogic(logic, itemData)
			if err != nil {
				// Em caso de erro, manter valor original
				result[i] = item
			} else {
				result[i] = evaluated
			}
		}

		return result
	})
}
