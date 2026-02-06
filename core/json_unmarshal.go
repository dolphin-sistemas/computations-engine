package core

import "encoding/json"

// UnmarshalJSON preserves arbitrary nested state fields.
//
// Known top-level keys are decoded into strongly-typed fields (id, tenantId, items, totals, fields, meta).
// Any other top-level keys are preserved into State.Fields (without overriding an existing key).
func (s *State) UnmarshalJSON(data []byte) error {
	type alias State
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*s = State(decoded)

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	delete(raw, "id")
	delete(raw, "tenantId")
	delete(raw, "items")
	delete(raw, "totals")
	delete(raw, "fields")
	delete(raw, "meta")

	if len(raw) == 0 {
		return nil
	}
	if s.Fields == nil {
		s.Fields = make(map[string]interface{}, len(raw))
	}
	for k, v := range raw {
		if k == "" {
			continue
		}
		if _, exists := s.Fields[k]; exists {
			continue
		}
		s.Fields[k] = v
	}
	return nil
}

// UnmarshalJSON preserves arbitrary nested item fields.
//
// Known keys:
// - id (string)
// - amount (number) [also accepts "quantity" as fallback]
// - fields (object)
//
// Any other keys are preserved into Item.Fields.
func (it *Item) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var fields map[string]interface{}
	if fm, ok := raw["fields"].(map[string]interface{}); ok && fm != nil {
		fields = make(map[string]interface{}, len(fm)+8)
		for k, v := range fm {
			if k != "" {
				fields[k] = v
			}
		}
	}

	if id, ok := raw["id"].(string); ok {
		it.ID = id
	}

	if v, ok := raw["amount"]; ok {
		if f, ok2 := asFloat64(v); ok2 {
			it.Amount = f
		}
	} else if v, ok := raw["quantity"]; ok {
		if f, ok2 := asFloat64(v); ok2 {
			it.Amount = f
		}
	}

	// Preserve unknown keys into Fields.
	for k, v := range raw {
		switch k {
		case "id", "amount", "quantity", "fields":
			continue
		default:
			if k == "" {
				continue
			}
			if fields == nil {
				fields = make(map[string]interface{}, 8)
			}
			fields[k] = v
		}
	}

	it.Fields = fields
	return nil
}

func asFloat64(v interface{}) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case int32:
		return float64(t), true
	case uint:
		return float64(t), true
	case uint64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}
