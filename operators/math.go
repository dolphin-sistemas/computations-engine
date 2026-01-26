package operators

import (
	"encoding/json"
	"math"

	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/dolphin-sistemas/engine/pkg"
)

func init() {
	// Registrar operador "sum" para somar arrays de números
	jsonlogic.AddOperator("sum", func(values, data interface{}) interface{} {
		var arr []interface{}

		switch v := values.(type) {
		case []interface{}:
			if len(v) == 0 {
				return 0.0
			}
			// O primeiro argumento deve ser o array a ser somado
			if nestedArr, ok := v[0].([]interface{}); ok {
				arr = nestedArr
			} else if nestedArr, ok := v[0].([]float64); ok {
				// Converter []float64 para []interface{}
				arr = make([]interface{}, len(nestedArr))
				for i, f := range nestedArr {
					arr[i] = f
				}
			} else {
				// Se não é array, tentar somar os valores diretamente
				arr = v
			}
		default:
			return 0.0
		}

		var sum float64
		for _, item := range arr {
			switch n := item.(type) {
			case float64:
				sum += n
			case float32:
				sum += float64(n)
			case int:
				sum += float64(n)
			case int64:
				sum += float64(n)
			case json.Number:
				if f, err := n.Float64(); err == nil {
					sum += f
				}
			}
		}

		return sum
	})

	// Registrar operador "round2" para arredondar para 2 casas decimais
	jsonlogic.AddOperator("round2", func(values, data interface{}) interface{} {
		val := extractFloat64(values)
		return math.Round(val*100) / 100.0
	})

	// Registrar operador "round" genérico: {"round": [value, decimals]}
	jsonlogic.AddOperator("round", func(values, data interface{}) interface{} {
		var val float64
		var decimals float64 = 0

		switch v := values.(type) {
		case []interface{}:
			if len(v) == 0 {
				return 0.0
			}
			val = pkg.ExtractFloat64FromValue(v[0])
			if len(v) > 1 {
				decimals = pkg.ExtractFloat64FromValue(v[1])
			}
		default:
			return 0.0
		}

		multiplier := math.Pow(10, decimals)
		return math.Round(val*multiplier) / multiplier
	})
}

// extractFloat64 extrai um float64 de um valor interface{}
func extractFloat64(values interface{}) float64 {
	return pkg.ExtractFloat64FromValue(values)
}
