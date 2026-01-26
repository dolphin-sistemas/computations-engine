package loader

import (
	"encoding/json"
	"fmt"

	"github.com/dolphin-sistemas/engine/core"
)

// LoadRulePackFromJSON carrega um RulePack de dados JSON
func LoadRulePackFromJSON(data []byte) (core.RulePack, error) {
	var rulePack core.RulePack
	if err := json.Unmarshal(data, &rulePack); err != nil {
		return core.RulePack{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validar RulePack
	if err := validateRulePack(rulePack); err != nil {
		return core.RulePack{}, fmt.Errorf("invalid rule pack: %w", err)
	}

	return rulePack, nil
}

// validateRulePack valida um RulePack
func validateRulePack(rulePack core.RulePack) error {
	if rulePack.ID == "" {
		return fmt.Errorf("rulePack.id is required")
	}
	if rulePack.Version == "" {
		return fmt.Errorf("rulePack.version is required")
	}
	return nil
}
