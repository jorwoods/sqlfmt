package formatter

import (
  "os"
  "testing"
  "strings"
)

func TestLoadConfig(t *testing.T) {
  // Write a temporary config file
  configYAML := `rules:
  uppercase_keywords: false
  align_clauses: true
  strip_quotes: false
  format_select_list: true
`
  tmp := "test_sqlfmt.yaml"
  if err := os.WriteFile(tmp, []byte(configYAML), 0644); err != nil {
    t.Fatalf("failed to write temp config: %v", err)
  }
  defer os.Remove(tmp)

  // Change working dir to current for config search
  cwd, _ := os.Getwd()
  defer os.Chdir(cwd)
  os.Chdir(".")

  // Rename to match loader expectation
  os.Rename(tmp, "sqlfmt.yaml")
  defer os.Remove("sqlfmt.yaml")

  cfg, err := LoadConfig()
  if err != nil {
    t.Fatalf("LoadConfig failed: %v", err)
  }
  if cfg == nil {
    t.Fatal("expected config, got nil")
  }
  if cfg.Rules.UppercaseKeywords {
    t.Error("expected UppercaseKeywords to be false")
  }
  if !cfg.Rules.AlignClauses {
    t.Error("expected AlignClauses to be true")
  }
  if cfg.Rules.StripQuotes {
    t.Error("expected StripQuotes to be false")
  }
  if !cfg.Rules.FormatSelectList {
    t.Error("expected FormatSelectList to be true")
  }
}

func TestFormatSQLWithConfig(t *testing.T) {
  cfg := &Config{
    Rules: RulesConfig{
      UppercaseKeywords: false,
      AlignClauses:      true,
      StripQuotes:       false,
      FormatSelectList:  true,
    },
  }
  input := `select "ID", "NAME" from "USERS" where "AGE" > 30`
  got := FormatSQLWithConfig(input, cfg)
  if strings.Contains(got, "SELECT") {
    t.Error("expected keywords to not be uppercased")
  }
  if strings.Contains(got, "\"") {
    // Quotes should remain
  } else {
    t.Error("expected quoted identifiers to remain")
  }
}
