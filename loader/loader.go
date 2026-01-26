package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dolphin-sistemas/computations-engine/core"
)

// LoadRulePackFromFile carrega um RulePack de um arquivo (JSON ou YAML)
func LoadRulePackFromFile(path string) (core.RulePack, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return core.RulePack{}, fmt.Errorf("failed to read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return LoadRulePackFromJSON(data)
	case ".yaml", ".yml":
		return LoadRulePackFromYAML(data)
	default:
		return core.RulePack{}, fmt.Errorf("unsupported file format: %s (use .json or .yaml)", ext)
	}
}
