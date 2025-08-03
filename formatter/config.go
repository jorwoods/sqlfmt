package formatter

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type RulesConfig struct {
	UppercaseKeywords             bool `yaml:"uppercase_keywords"`
	AlignClauses                  bool `yaml:"align_clauses"`
	StripQuotes                   bool `yaml:"strip_quotes"`
	FormatSelectList              bool `yaml:"format_select_list"`
	RefactorLongSubqueriesToCTE   bool `yaml:"refactor_long_subqueries_to_cte"`
}

type Config struct {
	Rules RulesConfig `yaml:"rules"`
}

// LoadConfig searches for sqlfmt.yaml in the current or parent directories.
func LoadConfig() (*Config, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	for {
		configPath := filepath.Join(dir, "sqlfmt.yaml")
		if _, err := os.Stat(configPath); err == nil {
			f, err := os.Open(configPath)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			var cfg Config
			dec := yaml.NewDecoder(f)
			if err := dec.Decode(&cfg); err != nil {
				return nil, err
			}
			return &cfg, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil, nil // not found
}
