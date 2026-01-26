package diff

import (
	"encoding/json"
	"reflect"

	"github.com/dolphin-sistemas/computations-engine/core"
)

// CompareStates compara dois estados e retorna se são diferentes
func CompareStates(original, current core.State) bool {
	// Comparação simples usando JSON marshaling
	origJSON, _ := json.Marshal(original)
	currJSON, _ := json.Marshal(current)
	return !reflect.DeepEqual(origJSON, currJSON)
}

// GetChangedFields retorna lista de campos que mudaram
func GetChangedFields(original, current core.State) []string {
	var changed []string

	// Comparar totals
	if original.Totals != current.Totals {
		changed = append(changed, "totals")
	}

	// Comparar fields
	if !reflect.DeepEqual(original.Fields, current.Fields) {
		changed = append(changed, "fields")
	}

	// Comparar items
	if len(original.Items) != len(current.Items) {
		changed = append(changed, "items")
	} else {
		for i := range original.Items {
			if !reflect.DeepEqual(original.Items[i], current.Items[i]) {
				changed = append(changed, "items")
				break
			}
		}
	}

	return changed
}
