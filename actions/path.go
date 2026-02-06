package actions

import (
	"fmt"
	"strconv"
	"strings"
)

// PathStep represents one step in a path (e.g. "items[*]", "items[0]" or "percent").
type PathStep struct {
	Key      string
	Wildcard bool // true if [*]
	HasIndex bool
	Index    int
}

// ParsePath parses a target path like "items[*].fields.negotiations[*].percent" into steps.
// Supports arbitrary nesting: key, key[*], key.key[*].key, etc.
func ParsePath(target string) ([]PathStep, error) {
	if target == "" {
		return nil, nil
	}
	var steps []PathStep
	remaining := target
	for remaining != "" {
		dot := strings.Index(remaining, ".")
		var seg string
		if dot < 0 {
			seg = remaining
			remaining = ""
		} else {
			seg = remaining[:dot]
			remaining = remaining[dot+1:]
		}
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}

		step := PathStep{Key: seg}
		if open := strings.Index(seg, "["); open >= 0 {
			if !strings.HasSuffix(seg, "]") {
				return nil, fmt.Errorf("invalid path segment %q (missing closing ])", seg)
			}
			key := strings.TrimSpace(seg[:open])
			if key == "" {
				return nil, fmt.Errorf("invalid path segment %q (empty key)", seg)
			}
			raw := strings.TrimSpace(seg[open+1 : len(seg)-1])
			step.Key = key
			switch raw {
			case "*":
				step.Wildcard = true
			default:
				i, err := strconv.Atoi(raw)
				if err != nil {
					return nil, fmt.Errorf("invalid path segment %q (index must be number or *): %w", seg, err)
				}
				if i < 0 {
					return nil, fmt.Errorf("invalid path segment %q (negative index)", seg)
				}
				step.HasIndex = true
				step.Index = i
			}
		}
		steps = append(steps, step)
	}
	return steps, nil
}

// HasWildcard returns true if any step has Wildcard
func HasWildcard(steps []PathStep) bool {
	for _, s := range steps {
		if s.Wildcard {
			return true
		}
	}
	return false
}

// LeafKey returns the last segment key (for setting the final value)
func LeafKey(steps []PathStep) string {
	if len(steps) == 0 {
		return ""
	}
	return steps[len(steps)-1].Key
}
