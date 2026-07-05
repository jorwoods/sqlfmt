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
	RequireExplicitAS             bool `yaml:"require_explicit_as"`
	TrailingSemicolon             bool `yaml:"trailing_semicolon"`
	StripTrailingWhitespace       bool `yaml:"strip_trailing_whitespace"`
	NormalizeNotEqual             bool `yaml:"normalize_not_equal"`
	OperatorSpacing               bool `yaml:"operator_spacing"`
	BlankLinesBetweenStatements   bool `yaml:"blank_lines_between_statements"`
	NewlineBeforeAndOr            bool `yaml:"newline_before_and_or"`
	NormalizeBoolean              bool `yaml:"normalize_boolean"`
	UppercaseFunctions            bool `yaml:"uppercase_functions"`
	NewlineBeforeJoin             bool `yaml:"newline_before_join"`
	NewlineBeforeOn               bool `yaml:"newline_before_on"`
	IndentCaseWhen                bool `yaml:"indent_case_when"`
	LeadingComma                  bool `yaml:"leading_comma"`
	NormalizeNullComparison       bool `yaml:"normalize_null_comparison"`
	TrailingNewline               bool `yaml:"trailing_newline"`
	NewlineBeforeLimit            bool `yaml:"newline_before_limit"`
	NormalizeOrderDirection       bool `yaml:"normalize_order_direction"`
	CTEFormatting                 bool `yaml:"cte_formatting"`
	LeadingCommaCTE               bool `yaml:"leading_comma_cte"`
	RemoveRedundantParens         bool `yaml:"remove_redundant_parens"`
	NewlineBeforeSetOp            bool `yaml:"newline_before_set_op"`
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
