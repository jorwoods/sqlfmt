
package formatter

import (
	"strings"
	"testing"
)


func TestFormatSQL_NoRulesEnabled_PassThrough(t *testing.T) {
	       cfg := &Config{Rules: RulesConfig{
		       UppercaseKeywords: false,
		       AlignClauses: false,
		       StripQuotes: false,
		       FormatSelectList: false,
	       }}
       input := `select "id", "name" from users where age > 30`
       got := FormatSQLWithConfig(input, cfg)
       if got != input {
	       t.Errorf("expected input to pass through unchanged, got: %q", got)
       }
}

type formatTestCase struct {
	name     string
	input    string
	expected string
	rules    RulesConfig
}

// All test cases now expect keywords and function names to be uppercased and clause detection to be grammar-driven.
var testCases = []formatTestCase{
	       {
		       name: "require explicit AS for aliases",
		       input: `select id user_id, name as user_name from users u` ,
		       expected: "SELECT id AS user_id, name AS user_name FROM users AS u",
		       rules: RulesConfig{UppercaseKeywords: true, RequireExplicitAS: true},
	       },
	       {
		       name: "do not require explicit AS (default)",
		       input: `select id user_id, name as user_name from users u` ,
		       expected: "SELECT id user_id, name AS user_name FROM users u",
		       rules: RulesConfig{UppercaseKeywords: true, RequireExplicitAS: false},
	       },
       {
	       name: "uppercase keywords only",
	       input: `select id, name from users`,
	       expected: "SELECT id, name FROM users",
	       rules: RulesConfig{UppercaseKeywords: true},
       },
       {
	       name: "align clauses only",
	       input: `select id, name from users where age > 30`,
	       expected: "select id, name\n  from users\n where age > 30",
	       rules: RulesConfig{AlignClauses: true},
       },
       {
	       name: "strip quotes only",
	       input: `select "id", "name" from "users"`,
	       expected: "select id, name from users",
	       rules: RulesConfig{StripQuotes: true},
       },
       {
	       name: "format select list only",
	       input: `select id, name, age, email, country from users`,
	       expected: "select id,\n       name,\n       age,\n       email,\n       country\nfrom users",
	       rules: RulesConfig{FormatSelectList: true},
       },
	       {
		       name: "all rules enabled",
		       input: `select "id", "name", age from "users" where age > 30`,
		       expected: "SELECT id, name, age\n  FROM users\n WHERE age > 30",
		       rules: RulesConfig{
			       UppercaseKeywords: true,
			       AlignClauses: true,
			       StripQuotes: true,
			       FormatSelectList: true,
		       },
	       },
}

func TestFormatSQL(t *testing.T) {
		       for _, tc := range testCases {
			       t.Run(tc.name, func(t *testing.T) {
				       cfg := &Config{Rules: tc.rules}
				       output := strings.TrimSpace(FormatSQLWithConfig(tc.input, cfg))
				       if output != tc.expected {
					       t.Errorf("unexpected format output\n--- got:\n%s\n--- want:\n%s", output, tc.expected)
				       }
			       })
		       }
}
