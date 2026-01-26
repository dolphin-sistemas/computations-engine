package loader

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/dolphin-sistemas/computations-engine/core"
)

// LoadRulePackFromYAML carrega um RulePack de dados YAML
func LoadRulePackFromYAML(data []byte) (core.RulePack, error) {
	var rulePack core.RulePack
	if err := yaml.Unmarshal(data, &rulePack); err != nil {
		return core.RulePack{}, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Validar RulePack
	if err := validateRulePack(rulePack); err != nil {
		return core.RulePack{}, fmt.Errorf("invalid rule pack: %w", err)
	}

	return rulePack, nil
}
