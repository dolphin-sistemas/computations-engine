package diff

import (
	"github.com/dolphin-sistemas/engine/core"
)

// BuildStateFragment extrai apenas os campos que mudaram (otimização para UI)
func BuildStateFragment(ctx *core.EngineContext) map[string]interface{} {
	fragment := make(map[string]interface{})
	state := ctx.State

	// Totais sempre expor
	if state.Totals != (core.Totals{}) {
		fragment["totals"] = state.Totals
	}

	// Fields customizados
	if len(state.Fields) > 0 {
		fragment["fields"] = state.Fields
	}

	// Items com seus campos
	itemsFragment := make([]map[string]interface{}, len(state.Items))
	for i, item := range state.Items {
		itemsFragment[i] = map[string]interface{}{
			"id":     item.ID,
			"fields": item.Fields,
		}
	}
	if len(itemsFragment) > 0 {
		fragment["items"] = itemsFragment
	}

	return fragment
}

// BuildServerDelta calcula diferenças entre original e resultado (para sincronização)
func BuildServerDelta(ctx *core.EngineContext) map[string]interface{} {
	delta := make(map[string]interface{})
	original := ctx.Original
	current := ctx.State

	// Comparar totais
	if original.Totals != current.Totals {
		delta["totals"] = current.Totals
	}

	// Comparar fields (simplificado: sempre envia se mudou ou é novo)
	if len(current.Fields) > 0 {
		delta["fields"] = current.Fields
	}

	// Comparar items (simplificado: sempre envia todos se houver mudanças)
	if len(current.Items) > 0 {
		delta["items"] = current.Items
	}

	return delta
}
