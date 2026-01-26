package actions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dolphin-sistemas/engine/core"
)

// SetValue define um valor em um target (fields.x ou items[i].fields.y)
func SetValue(state *core.State, target string, value interface{}) error {
	parts := strings.Split(target, ".")
	if len(parts) == 0 {
		return fmt.Errorf("invalid target: %s", target)
	}

	// Caso 1: fields.x (campo do estado)
	if parts[0] == "fields" && len(parts) == 2 {
		if state.Fields == nil {
			state.Fields = make(map[string]interface{})
		}
		state.Fields[parts[1]] = value
		return nil
	}

	// Caso 2: totals.x (total/sumário)
	if parts[0] == "totals" && len(parts) == 2 {
		// Converter para float64
		var val float64
		switch v := value.(type) {
		case float64:
			val = v
		case float32:
			val = float64(v)
		case int:
			val = float64(v)
		case int64:
			val = float64(v)
		case int32:
			val = float64(v)
		default:
			return fmt.Errorf("totals field must be numeric, got %T", value)
		}

		switch parts[1] {
		case "subtotal":
			state.Totals.Subtotal = val
		case "discount":
			state.Totals.Discount = val
		case "tax":
			state.Totals.Tax = val
		case "total":
			state.Totals.Total = val
		default:
			return fmt.Errorf("unknown total field: %s", parts[1])
		}
		return nil
	}

	// Caso 3: items[i].fields.x (campo de um item específico)
	if parts[0] == "items" && len(parts) >= 3 {
		if len(parts) < 2 {
			return fmt.Errorf("invalid item target: %s", target)
		}

		// Parse do índice: [0]
		idxStr := strings.Trim(parts[1], "[]")
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return fmt.Errorf("invalid item index in target: %s", target)
		}

		if idx < 0 || idx >= len(state.Items) {
			return fmt.Errorf("item index out of range: %d", idx)
		}

		// Se for items[i].fields.x
		if parts[2] == "fields" && len(parts) == 4 {
			if state.Items[idx].Fields == nil {
				state.Items[idx].Fields = make(map[string]interface{})
			}
			state.Items[idx].Fields[parts[3]] = value
			return nil
		}

		return fmt.Errorf("unsupported item target format: %s", target)
	}

	// Caso 4: items[*].fields.x (aplicar a todos os itens com mesmo valor)
	if parts[0] == "items" && len(parts) >= 2 {
		idxStr := strings.Trim(parts[1], "[]")
		if idxStr == "*" && len(parts) >= 3 {
			if parts[2] == "fields" && len(parts) == 4 {
				for i := range state.Items {
					if state.Items[i].Fields == nil {
						state.Items[i].Fields = make(map[string]interface{})
					}
					state.Items[i].Fields[parts[3]] = value
				}
				return nil
			}
			return fmt.Errorf("unsupported items[*] target format: %s", target)
		}
	}

	return fmt.Errorf("unsupported target format: %s", target)
}

// GetValue obtém um valor de um target
func GetValue(state *core.State, target string) (interface{}, error) {
	parts := strings.Split(target, ".")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid target: %s", target)
	}

	// Caso 1: fields.x
	if parts[0] == "fields" && len(parts) == 2 {
		if state.Fields == nil {
			return nil, nil
		}
		return state.Fields[parts[1]], nil
	}

	// Caso 2: totals.x
	if parts[0] == "totals" && len(parts) == 2 {
		switch parts[1] {
		case "subtotal":
			return state.Totals.Subtotal, nil
		case "discount":
			return state.Totals.Discount, nil
		case "tax":
			return state.Totals.Tax, nil
		case "total":
			return state.Totals.Total, nil
		}
		return nil, fmt.Errorf("unknown total field: %s", parts[1])
	}

	// Caso 3: items[i].fields.x
	if parts[0] == "items" && len(parts) >= 3 {
		idxStr := strings.Trim(parts[1], "[]")
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return nil, fmt.Errorf("invalid item index: %s", target)
		}

		if idx < 0 || idx >= len(state.Items) {
			return nil, fmt.Errorf("item index out of range: %d", idx)
		}

		if parts[2] == "fields" && len(parts) == 4 {
			if state.Items[idx].Fields == nil {
				return nil, nil
			}
			return state.Items[idx].Fields[parts[3]], nil
		}
	}

	return nil, fmt.Errorf("unsupported target format: %s", target)
}
