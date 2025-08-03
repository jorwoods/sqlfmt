package formatter

import (
	"os"
	"strings"
	"testing"
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
	if cfg.Rules.RefactorLongSubqueriesToCTE {
		t.Error("expected RefactorLongSubqueriesToCTE to be false")
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

func TestCTERefactorEnabled(t *testing.T) {
	cfg := &Config{Rules: RulesConfig{RefactorLongSubqueriesToCTE: true}}
	if !CTERefactorEnabled(cfg) {
		t.Error("expected CTERefactorEnabled to be true when config is true")
	}
	cfg = &Config{Rules: RulesConfig{RefactorLongSubqueriesToCTE: false}}
	if CTERefactorEnabled(cfg) {
		t.Error("expected CTERefactorEnabled to be false when config is false")
	}
	if !CTERefactorEnabled(nil) {
		t.Error("expected CTERefactorEnabled to be true when config is nil (default)")
	}
}

func TestRefactorLongSubqueriesToCTE(t *testing.T) {
	cfg := &Config{Rules: RulesConfig{RefactorLongSubqueriesToCTE: true}}
	// Long subquery (should be refactored)
	input := `SELECT * FROM (SELECT id, name, email, country, age, salary, department, hire_date, status, manager_id, region, office, phone, address, zip, state, country_code FROM employees)`
	got := RefactorLongSubqueriesToCTE(input, cfg)
	if !strings.HasPrefix(got, "WITH cte_1 AS ") {
		t.Errorf("expected CTE refactor, got: %s", got)
	}
	if !strings.Contains(got, "FROM cte_1") {
		t.Errorf("expected subquery replaced with CTE name, got: %s", got)
	}

	// Short subquery (should not be refactored)
	input2 := `SELECT * FROM (SELECT id FROM employees)`
	got2 := RefactorLongSubqueriesToCTE(input2, cfg)
	if got2 != input2 {
		t.Errorf("expected no refactor for short subquery, got: %s", got2)
	}

	// Correlated subquery (should not be refactored, but our stub always says not correlated)
	// Placeholder for future: input3 := ...
}

func TestCTERefactorOutputValidation(t *testing.T) {
	   cfg := &Config{Rules: RulesConfig{RefactorLongSubqueriesToCTE: true}}
	   // Valid input, but refactor would produce invalid output (simulate by breaking the output)
	   isValidSnowflakeSQLTestHook = func(sql string) bool {
			   // Only allow the original input string to parse, not the output
			   if strings.HasPrefix(sql, "WITH cte_1 AS") {
					   return false
			   }
			   return true
	   }
	   oldExitFunc := exitFunc
	   defer func() {
			   isValidSnowflakeSQLTestHook = nil
			   exitFunc = oldExitFunc
	   }()
	   input := `SELECT * FROM (SELECT id, name, email, country, age, salary, department, hire_date, status, manager_id, region, office, phone, address, zip, state, country_code FROM employees)`
	   exitCalled := false
	   exitFunc = func(code int) { exitCalled = true; panic("exit called") }
	   defer func() {
			   if r := recover(); r != nil {
					   if r != "exit called" {
							   t.Errorf("unexpected panic: %v", r)
					   }
			   }
	   }()
	   got := RefactorLongSubqueriesToCTE(input, cfg)
	   if !exitCalled {
			   t.Errorf("expected exitFunc to be called on output validation failure")
	   }
	   if got != input {
			   t.Errorf("expected original input to be returned if output is invalid, got: %s", got)
	   }

	   // If input is invalid, should also return input (no exit, as input is checked before output)
	   isValidSnowflakeSQLTestHook = func(sql string) bool { return false }
	   exitCalled = false
	   got2 := RefactorLongSubqueriesToCTE(input, cfg)
	   if exitCalled {
			   t.Errorf("did not expect exitFunc to be called when input is invalid")
	   }
	   if got2 != input {
			   t.Errorf("expected original input to be returned if input is invalid, got: %s", got2)
	   }
}
