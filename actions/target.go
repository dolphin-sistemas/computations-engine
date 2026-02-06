package actions

import (
	"encoding/json"
	"fmt"

	"github.com/dolphin-sistemas/computations-engine/core"
)

type selectedValue struct {
	Key   string
	Value interface{}
}

type leafRef struct {
	Get func() (interface{}, error)
	Set func(interface{}) error
}

// SetValue defines a value at a target path.
//
// Supported (examples):
// - paymentTermDays (stored in state.Fields)
// - fields.customerType
// - totals.total
// - items[*].unitPriceOnScreen
// - items[*].negotiations[*].percent
// - items[*].foo[*].bar.baz
//
// Notes:
// - For backward compatibility, "items[*].fields.x" is treated as an alias of "items[*].x".
// - Unknown keys on state/items are stored under their respective Fields maps.
func SetValue(state *core.State, target string, value interface{}) error {
	steps, err := ParsePath(target)
	if err != nil {
		return err
	}
	if len(steps) == 0 {
		return fmt.Errorf("invalid target: %q", target)
	}

	setCount, err := visitLeaves(state, steps, true, func(ref leafRef, _ []selectedValue) error {
		return ref.Set(value)
	})
	if err != nil {
		return err
	}

	// Wildcard targets that match nothing are treated as no-ops.
	_ = setCount
	return nil
}

// GetValue obtains a value at a target path.
//
// If the path contains wildcards, it returns a []interface{} with one entry per match.
// If the path matches nothing, it returns (nil, nil).
func GetValue(state *core.State, target string) (interface{}, error) {
	steps, err := ParsePath(target)
	if err != nil {
		return nil, err
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("invalid target: %q", target)
	}

	if !HasWildcard(steps) {
		var out interface{}
		var seen bool
		_, err := visitLeaves(state, steps, false, func(ref leafRef, _ []selectedValue) error {
			v, err := ref.Get()
			if err != nil {
				return err
			}
			out = v
			seen = true
			return nil
		})
		if err != nil {
			return nil, err
		}
		if !seen {
			return nil, nil
		}
		return out, nil
	}

	values := make([]interface{}, 0, 8)
	_, err = visitLeaves(state, steps, false, func(ref leafRef, _ []selectedValue) error {
		v, err := ref.Get()
		if err != nil {
			return err
		}
		values = append(values, v)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return values, nil
}

// SetValueAtIndices sets a value at a specific index combination for wildcard steps.
//
// Example: for "items[*].negotiations[*].percent" the indices slice is [itemIdx, negotiationIdx].
func SetValueAtIndices(state *core.State, target string, indices []int, value interface{}) error {
	steps, err := ParsePath(target)
	if err != nil {
		return err
	}
	if len(steps) == 0 {
		return fmt.Errorf("invalid target: %q", target)
	}

	wildIdx := 0
	current := interface{}(state)
	for stepIdx := 0; stepIdx < len(steps)-1; stepIdx++ {
		step := steps[stepIdx]
		var next *PathStep
		if stepIdx+1 < len(steps) {
			next = &steps[stepIdx+1]
		}

		child, childSetter, err := getChild(current, step.Key, true)
		if err != nil {
			return err
		}
		if child == nil {
			// Missing path: create a container if possible, otherwise fail (this is a targeted set).
			child = makeContainerForNext(next)
			if child == nil {
				return fmt.Errorf("path %q does not exist at %q", target, step.Key)
			}
			if err := childSetter(child); err != nil {
				return err
			}
		}

		switch {
		case step.HasIndex:
			elem, err := selectIndex(child, childSetter, step.Index, true, next)
			if err != nil {
				return err
			}
			current = elem
		case step.Wildcard:
			if wildIdx >= len(indices) {
				return fmt.Errorf("missing wildcard index %d for target %q", wildIdx, target)
			}
			elem, err := selectIndex(child, childSetter, indices[wildIdx], true, next)
			if err != nil {
				return err
			}
			wildIdx++
			current = elem
		default:
			current = child
		}
	}

	ref, err := makeLeafRef(current, steps[len(steps)-1], true, nil)
	if err != nil {
		return err
	}
	return ref.Set(value)
}

func visitLeaves(
	state *core.State,
	steps []PathStep,
	createMissing bool,
	onLeaf func(ref leafRef, selections []selectedValue) error,
) (int, error) {
	if len(steps) == 0 {
		return 0, nil
	}
	return visitAt(interface{}(state), steps, 0, createMissing, nil, onLeaf)
}

func visitAt(
	current interface{},
	steps []PathStep,
	stepIdx int,
	createMissing bool,
	selections []selectedValue,
	onLeaf func(ref leafRef, selections []selectedValue) error,
) (int, error) {
	if stepIdx >= len(steps) {
		return 0, nil
	}

	var next *PathStep
	if stepIdx+1 < len(steps) {
		next = &steps[stepIdx+1]
	}

	// Last step: yield a leaf reference.
	if stepIdx == len(steps)-1 {
		ref, err := makeLeafRef(current, steps[stepIdx], createMissing, next)
		if err != nil {
			return 0, err
		}
		if ref.Get == nil || ref.Set == nil {
			return 0, fmt.Errorf("invalid target at %q", steps[stepIdx].Key)
		}
		if err := onLeaf(ref, selections); err != nil {
			return 0, err
		}
		return 1, nil
	}

	step := steps[stepIdx]
	child, childSetter, err := getChild(current, step.Key, createMissing)
	if err != nil {
		return 0, err
	}

	if child == nil {
		if !createMissing {
			return 0, nil
		}
		// Don't create containers for wildcard steps (no implicit element creation).
		if step.Wildcard {
			return 0, nil
		}
		// For explicit indexes (e.g. foo[0]) we can create an array.
		if step.HasIndex {
			created := makeContainerForNext(&PathStep{HasIndex: true, Index: step.Index})
			if err := childSetter(created); err != nil {
				return 0, err
			}
			child = created
		} else {
			created := makeContainerForNext(next)
			if created == nil {
				return 0, nil
			}
			if err := childSetter(created); err != nil {
				return 0, err
			}
			child = created
		}
	}

	switch {
	case step.HasIndex:
		elem, err := selectIndex(child, childSetter, step.Index, createMissing, next)
		if err != nil {
			return 0, err
		}
		nextSelections := append(selections, selectedValue{Key: step.Key, Value: elem})
		return visitAt(elem, steps, stepIdx+1, createMissing, nextSelections, onLeaf)

	case step.Wildcard:
		total := 0
		switch arr := child.(type) {
		case []core.Item:
			for i := range arr {
				elem := &arr[i]
				nextSelections := append(selections, selectedValue{Key: step.Key, Value: elem})
				n, err := visitAt(elem, steps, stepIdx+1, createMissing, nextSelections, onLeaf)
				if err != nil {
					return 0, err
				}
				total += n
			}
			return total, nil
		case []interface{}:
			for i := 0; i < len(arr); i++ {
				elem, err := selectIndex(child, childSetter, i, createMissing, next)
				if err != nil {
					return 0, err
				}
				nextSelections := append(selections, selectedValue{Key: step.Key, Value: elem})
				n, err := visitAt(elem, steps, stepIdx+1, createMissing, nextSelections, onLeaf)
				if err != nil {
					return 0, err
				}
				total += n
			}
			return total, nil
		default:
			return 0, nil
		}

	default:
		return visitAt(child, steps, stepIdx+1, createMissing, selections, onLeaf)
	}
}

func makeLeafRef(parent interface{}, leaf PathStep, createMissing bool, next *PathStep) (leafRef, error) {
	if leaf.Wildcard || leaf.HasIndex {
		return leafRef{
			Get: func() (interface{}, error) {
				child, _, err := getChild(parent, leaf.Key, false)
				if err != nil {
					return nil, err
				}
				switch arr := child.(type) {
				case []interface{}:
					if leaf.Wildcard {
						return arr, nil
					}
					if leaf.Index < 0 || leaf.Index >= len(arr) {
						return nil, nil
					}
					return arr[leaf.Index], nil
				default:
					return nil, nil
				}
			},
			Set: func(v interface{}) error {
				child, childSetter, err := getChild(parent, leaf.Key, true)
				if err != nil {
					return err
				}
				if child == nil {
					// Don't create containers for wildcard leaf targets.
					if leaf.Wildcard {
						return nil
					}
					created := makeContainerForNext(&PathStep{HasIndex: leaf.HasIndex, Index: leaf.Index, Wildcard: leaf.Wildcard})
					if err := childSetter(created); err != nil {
						return err
					}
					child = created
				}

				arr, ok := child.([]interface{})
				if !ok {
					return nil
				}

				if leaf.Wildcard {
					for i := range arr {
						arr[i] = v
					}
					return nil
				}

				if leaf.Index < 0 {
					return fmt.Errorf("negative index")
				}
				if leaf.Index >= len(arr) {
					expanded := make([]interface{}, leaf.Index+1)
					copy(expanded, arr)
					arr = expanded
					if childSetter != nil {
						if err := childSetter(arr); err != nil {
							return err
						}
					}
				}

				arr[leaf.Index] = v
				return nil
			},
		}, nil
	}

	_, childSetter, err := getChild(parent, leaf.Key, createMissing)
	if err != nil {
		return leafRef{}, err
	}

	return leafRef{
		Get: func() (interface{}, error) {
			v, _, err := getChild(parent, leaf.Key, false)
			if err != nil {
				return nil, err
			}
			return v, nil
		},
		Set: func(v interface{}) error {
			if childSetter == nil {
				return nil
			}
			return childSetter(v)
		},
	}, nil
}

func getChild(current interface{}, key string, createMissing bool) (interface{}, func(interface{}) error, error) {
	switch c := current.(type) {
	case *core.State:
		switch key {
		case "id":
			return c.ID, func(v interface{}) error {
				if s, ok := v.(string); ok {
					c.ID = s
					return nil
				}
				return fmt.Errorf("state.id must be string, got %T", v)
			}, nil
		case "tenantId":
			return c.TenantID, func(v interface{}) error {
				if s, ok := v.(string); ok {
					c.TenantID = s
					return nil
				}
				return fmt.Errorf("state.tenantId must be string, got %T", v)
			}, nil
		case "items":
			return c.Items, func(v interface{}) error {
				items, ok := v.([]core.Item)
				if !ok {
					return fmt.Errorf("state.items must be []core.Item, got %T", v)
				}
				c.Items = items
				return nil
			}, nil
		case "totals":
			return &c.Totals, func(v interface{}) error {
				t, ok := v.(core.Totals)
				if !ok {
					return fmt.Errorf("state.totals must be core.Totals, got %T", v)
				}
				c.Totals = t
				return nil
			}, nil
		case "fields":
			if c.Fields == nil && createMissing {
				c.Fields = make(map[string]interface{})
			}
			return c.Fields, func(v interface{}) error {
				m, ok := v.(map[string]interface{})
				if !ok {
					return fmt.Errorf("state.fields must be object, got %T", v)
				}
				c.Fields = m
				return nil
			}, nil
		case "meta":
			if c.Meta == nil && createMissing {
				c.Meta = make(map[string]interface{})
			}
			return c.Meta, func(v interface{}) error {
				m, ok := v.(map[string]interface{})
				if !ok {
					return fmt.Errorf("state.meta must be object, got %T", v)
				}
				c.Meta = m
				return nil
			}, nil
		default:
			if c.Fields == nil && createMissing {
				c.Fields = make(map[string]interface{})
			}
			if c.Fields == nil {
				return nil, func(interface{}) error { return nil }, nil
			}
			return c.Fields[key], func(v interface{}) error {
				c.Fields[key] = v
				return nil
			}, nil
		}

	case *core.Item:
		switch key {
		case "id":
			return c.ID, func(v interface{}) error {
				if s, ok := v.(string); ok {
					c.ID = s
					return nil
				}
				return fmt.Errorf("item.id must be string, got %T", v)
			}, nil
		case "amount":
			return c.Amount, func(v interface{}) error {
				f, err := toFloat64(v)
				if err != nil {
					return err
				}
				c.Amount = f
				return nil
			}, nil
		case "fields":
			if c.Fields == nil && createMissing {
				c.Fields = make(map[string]interface{})
			}
			return c.Fields, func(v interface{}) error {
				m, ok := v.(map[string]interface{})
				if !ok {
					return fmt.Errorf("item.fields must be object, got %T", v)
				}
				c.Fields = m
				return nil
			}, nil
		default:
			if c.Fields == nil && createMissing {
				c.Fields = make(map[string]interface{})
			}
			if c.Fields == nil {
				return nil, func(interface{}) error { return nil }, nil
			}
			return c.Fields[key], func(v interface{}) error {
				c.Fields[key] = v
				return nil
			}, nil
		}

	case *core.Totals:
		switch key {
		case "subtotal":
			return c.Subtotal, func(v interface{}) error {
				f, err := toFloat64(v)
				if err != nil {
					return err
				}
				c.Subtotal = f
				return nil
			}, nil
		case "discount":
			return c.Discount, func(v interface{}) error {
				f, err := toFloat64(v)
				if err != nil {
					return err
				}
				c.Discount = f
				return nil
			}, nil
		case "tax":
			return c.Tax, func(v interface{}) error {
				f, err := toFloat64(v)
				if err != nil {
					return err
				}
				c.Tax = f
				return nil
			}, nil
		case "total":
			return c.Total, func(v interface{}) error {
				f, err := toFloat64(v)
				if err != nil {
					return err
				}
				c.Total = f
				return nil
			}, nil
		default:
			return nil, func(interface{}) error { return nil }, nil
		}

	case map[string]interface{}:
		return c[key], func(v interface{}) error {
			c[key] = v
			return nil
		}, nil
	default:
		return nil, func(interface{}) error { return nil }, nil
	}
}

func selectIndex(child interface{}, childSetter func(interface{}) error, idx int, createMissing bool, next *PathStep) (interface{}, error) {
	switch arr := child.(type) {
	case []core.Item:
		if idx < 0 || idx >= len(arr) {
			return nil, fmt.Errorf("index out of range: %d", idx)
		}
		return &arr[idx], nil

	case []interface{}:
		// Expand slice if needed and allowed.
		if idx >= len(arr) {
			if !createMissing {
				return nil, nil
			}
			expanded := make([]interface{}, idx+1)
			copy(expanded, arr)
			arr = expanded
			if childSetter != nil {
				if err := childSetter(arr); err != nil {
					return nil, err
				}
			}
		}

		elem := arr[idx]
		if elem == nil && createMissing {
			created := makeContainerForNext(next)
			if created != nil {
				arr[idx] = created
				elem = created
			}
		}
		return elem, nil

	default:
		return nil, nil
	}
}

func makeContainerForNext(next *PathStep) interface{} {
	if next == nil {
		return make(map[string]interface{})
	}
	if next.HasIndex {
		return make([]interface{}, next.Index+1)
	}
	if next.Wildcard {
		// Wildcard with no index hint: create an empty slice (iteration yields no matches).
		return []interface{}{}
	}
	return make(map[string]interface{})
}

func toFloat64(v interface{}) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case float32:
		return float64(n), nil
	case int:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case int32:
		return float64(n), nil
	case uint:
		return float64(n), nil
	case uint64:
		return float64(n), nil
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0, fmt.Errorf("numeric field must be numeric, got %T", v)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("numeric field must be numeric, got %T", v)
	}
}
