
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
	       rules: RulesConfig{AlignClauses: true, OperatorSpacing: true},
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
		       expected: "SELECT id, name, age\n  FROM users\n WHERE age > 30\n;",
		       rules: RulesConfig{
			       UppercaseKeywords: true,
			       AlignClauses:      true,
			       StripQuotes:       true,
			       FormatSelectList:  true,
			       RequireExplicitAS: true,
			       TrailingSemicolon: true,
			       OperatorSpacing:   true,
		       },
	       },
	       {
		       name: "trailing semicolon added when missing",
		       input: `select id from users`,
		       expected: "select id from users\n;",
		       rules: RulesConfig{TrailingSemicolon: true},
	       },
	       {
		       name: "trailing semicolon normalized to new line when present inline",
		       input: `select id from users;`,
		       expected: "select id from users\n;",
		       rules: RulesConfig{TrailingSemicolon: true},
	       },
	       {
		       name: "trailing semicolon with other rules",
		       input: `select id from users`,
		       expected: "SELECT id\n  FROM users\n;",
		       rules: RulesConfig{UppercaseKeywords: true, AlignClauses: true, TrailingSemicolon: true},
	       },
	       {
		       name: "normalize not equal rewrites <> to !=",
		       input: `select id from users where age <> 30`,
		       expected: "SELECT id FROM users WHERE age != 30",
		       rules: RulesConfig{UppercaseKeywords: true, NormalizeNotEqual: true, OperatorSpacing: true},
	       },
	       {
		       name: "normalize not equal leaves != unchanged",
		       input: `select id from users where age != 30`,
		       expected: "SELECT id FROM users WHERE age != 30",
		       rules: RulesConfig{UppercaseKeywords: true, NormalizeNotEqual: true, OperatorSpacing: true},
	       },
	       {
		       name: "operator_spacing true ensures spaces around operators",
		       input: `select id from users where age>=30`,
		       expected: "SELECT id FROM users WHERE age >= 30",
		       rules: RulesConfig{UppercaseKeywords: true, OperatorSpacing: true},
	       },
	       {
		       name: "operator_spacing false produces compact operators",
		       input: `select id from users where age = 30`,
		       expected: "SELECT id FROM users WHERE age=30",
		       rules: RulesConfig{UppercaseKeywords: true, OperatorSpacing: false},
	       },
	       {
		       name: "blank lines between statements with inline semicolons",
		       input: `select id from users; select name from products`,
		       expected: "SELECT id FROM users\n;\n\nSELECT name FROM products",
		       rules: RulesConfig{UppercaseKeywords: true, BlankLinesBetweenStatements: true},
	       },
	       {
		       name: "blank lines between statements with trailing semicolon rule",
		       input: `select id from users; select name from products`,
		       expected: "SELECT id FROM users\n;\n\nSELECT name FROM products\n;",
		       rules: RulesConfig{UppercaseKeywords: true, TrailingSemicolon: true, BlankLinesBetweenStatements: true},
	       },
	       {
		       name: "blank lines between statements single statement unchanged",
		       input: `select id from users`,
		       expected: "SELECT id FROM users",
		       rules: RulesConfig{UppercaseKeywords: true, BlankLinesBetweenStatements: true},
	       },
	       {
		       name: "strip trailing whitespace does not alter clean output",
		       input: "select id, name from users where age > 30",
		       expected: "SELECT id, name\n  FROM users\n WHERE age > 30",
		       rules: RulesConfig{UppercaseKeywords: true, AlignClauses: true, StripTrailingWhitespace: true, OperatorSpacing: true},
	       },
	       {
		       name: "newline_before_and_or: AND in WHERE (flat path)",
		       input: `select id from users where age > 30 and name = 'foo'`,
		       expected: "SELECT id FROM users WHERE age > 30\n   AND name = 'foo'",
		       rules: RulesConfig{UppercaseKeywords: true, NewlineBeforeAndOr: true, OperatorSpacing: true},
	       },
	       {
		       name: "newline_before_and_or: OR in WHERE (flat path)",
		       input: `select id from users where age > 30 or name = 'foo'`,
		       expected: "SELECT id FROM users WHERE age > 30\n    OR name = 'foo'",
		       rules: RulesConfig{UppercaseKeywords: true, NewlineBeforeAndOr: true, OperatorSpacing: true},
	       },
	       {
		       name: "newline_before_and_or: multiple conditions (flat path)",
		       input: `select id from users where a = 1 and b = 2 or c = 3`,
		       expected: "SELECT id FROM users WHERE a = 1\n   AND b = 2\n    OR c = 3",
		       rules: RulesConfig{UppercaseKeywords: true, NewlineBeforeAndOr: true, OperatorSpacing: true},
	       },
	       {
		       name: "newline_before_and_or disabled leaves AND OR inline",
		       input: `select id from users where age > 30 and name = 'foo'`,
		       expected: "SELECT id FROM users WHERE age > 30 AND name = 'foo'",
		       rules: RulesConfig{UppercaseKeywords: true, NewlineBeforeAndOr: false, OperatorSpacing: true},
	       },
	       {
		       name: "newline_before_and_or with align_clauses",
		       input: `select id from users where age > 30 and name = 'foo'`,
		       expected: "SELECT id\n  FROM users\n WHERE age > 30\n   AND name = 'foo'",
		       rules: RulesConfig{UppercaseKeywords: true, AlignClauses: true, NewlineBeforeAndOr: true, OperatorSpacing: true},
	       },
	       {
		       name: "newline_before_and_or with align_clauses multiple conditions",
		       input: `select id from users where a = 1 and b = 2 or c = 3`,
		       expected: "SELECT id\n  FROM users\n WHERE a = 1\n   AND b = 2\n    OR c = 3",
		       rules: RulesConfig{UppercaseKeywords: true, AlignClauses: true, NewlineBeforeAndOr: true, OperatorSpacing: true},
	       },
	       {
		       name: "uppercase_functions uppercases built-in aggregates",
		       input: `select count(*), sum(val), avg(val), max(val), min(val) from t`,
		       expected: `select COUNT(*), SUM(val), AVG(val), MAX(val), MIN(val) from t`,
		       rules: RulesConfig{UppercaseFunctions: true},
	       },
	       {
		       name: "uppercase_functions uppercases window and scalar functions",
		       input: `select coalesce(a, b), row_number() over (partition by x order by y) from t`,
		       expected: `select COALESCE(a, b), ROW_NUMBER() OVER (partition by x order by y) from t`,
		       rules: RulesConfig{UppercaseFunctions: true, OperatorSpacing: true},
	       },
	       {
		       name: "uppercase_functions with uppercase_keywords",
		       input: `select count(*) from users`,
		       expected: `SELECT COUNT(*) FROM users`,
		       rules: RulesConfig{UppercaseKeywords: true, UppercaseFunctions: true},
	       },
	       {
		       name: "normalize_boolean uppercases true/false/null",
		       input: `select id from users where active = true and deleted = false and name != null`,
		       expected: `select id from users where active = TRUE and deleted = FALSE and name != NULL`,
		       rules: RulesConfig{NormalizeBoolean: true, OperatorSpacing: true},
	       },
	       {
		       name: "normalize_boolean with uppercase_keywords",
		       input: `select id from users where active = true`,
		       expected: `SELECT id FROM users WHERE active = TRUE`,
		       rules: RulesConfig{UppercaseKeywords: true, NormalizeBoolean: true, OperatorSpacing: true},
	       },
	       {
		       name: "newline_before_and_or: AND not injected outside WHERE (SELECT clause)",
		       input: `select id, case when a = 1 and b = 2 then 'y' else 'n' end from users where c = 3`,
		       expected: "SELECT id, CASE WHEN a = 1 AND b = 2 THEN 'y' ELSE 'n' END FROM users WHERE c = 3",
		       rules: RulesConfig{UppercaseKeywords: true, NewlineBeforeAndOr: true, OperatorSpacing: true},
	       },
}

func TestStripTrailingWhitespace(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"line one   \nline two\t\nline three", "line one\nline two\nline three"},
		{"no trailing whitespace", "no trailing whitespace"},
		{"  leading spaces", "  leading spaces"},
		{"mixed   \n  indented   \nclean", "mixed\n  indented\nclean"},
	}
	for _, tc := range cases {
		got := stripTrailingWhitespace(tc.input)
		if got != tc.expected {
			t.Errorf("stripTrailingWhitespace(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
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
