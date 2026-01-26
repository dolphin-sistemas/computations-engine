package operators

import (
	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/dolphin-sistemas/computations-engine/pkg"
)

func init() {
	// Registrar operador "allocate": {"allocate": [total, weights]}
	// Distribui total proporcionalmente baseado em weights
	jsonlogic.AddOperator("allocate", func(values, data interface{}) interface{} {
		var total float64
		var weights []interface{}

		switch v := values.(type) {
		case []interface{}:
			if len(v) < 2 {
				return []interface{}{}
			}
			// Primeiro argumento é o total
			total = pkg.ExtractFloat64FromValue(v[0])
			// Segundo argumento são os pesos
			switch w := v[1].(type) {
			case []interface{}:
				weights = w
			default:
				return []interface{}{}
			}
		default:
			return []interface{}{}
		}

		if len(weights) == 0 {
			return []interface{}{}
		}

		// Calcular soma dos pesos
		var sumWeights float64
		weightValues := make([]float64, len(weights))
		for i, w := range weights {
			val := pkg.ExtractFloat64FromValue(w)
			weightValues[i] = val
			sumWeights += val
		}

		if sumWeights == 0 {
			// Se soma é zero, distribuir igualmente
			equalValue := total / float64(len(weights))
			result := make([]interface{}, len(weights))
			for i := range result {
				result[i] = equalValue
			}
			return result
		}

		// Distribuir proporcionalmente
		result := make([]interface{}, len(weights))
		var allocated float64
		for i, weight := range weightValues {
			if i == len(weights)-1 {
				// Último item recebe o restante para evitar erros de arredondamento
				result[i] = total - allocated
			} else {
				value := (total * weight) / sumWeights
				result[i] = value
				allocated += value
			}
		}

		return result
	})
}
