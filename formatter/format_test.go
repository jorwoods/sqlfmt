
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
		       RefactorLongSubqueriesToCTE: false,
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
}

// All test cases now expect keywords and function names to be uppercased and clause detection to be grammar-driven.
var testCases = []formatTestCase{
	{
		name: "basic select with quoted identifiers and lowercase keywords",
		input: `
select "ID", "NAME" from "USERS" where "AGE" > 30
`,
		expected: strings.TrimSpace(`
  SELECT ID, NAME
    FROM USERS
   WHERE AGE > 30
`),
	},
	{
		name: "no changes needed",
		input: `
  SELECT ID, NAME
    FROM USERS
   WHERE AGE > 30
`,
		expected: strings.TrimSpace(`
  SELECT ID, NAME
    FROM USERS
   WHERE AGE > 30
`),
	},
	{
		name: "select with more than 3 identifiers",
		input: `
select "ID", "NAME", "AGE", "EMAIL", "COUNTRY" from "USERS"
`,
		expected: strings.TrimSpace(`
  SELECT ID,
         NAME,
         AGE,
         EMAIL,
         COUNTRY
    FROM USERS
`),
	},
	{
		name: "select with more than 3 identifiers, trailing comma",
		input: `
select "ID", "NAME", "AGE", "EMAIL", "COUNTRY" from "USERS"
`,
		expected: strings.TrimSpace(`
  SELECT ID,
         NAME,
         AGE,
         EMAIL,
         COUNTRY
    FROM USERS
`),
	},
}

func TestFormatSQL(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := strings.TrimSpace(FormatSQL(tc.input))
			if output != tc.expected {
				t.Errorf("unexpected format output\n--- got:\n%s\n--- want:\n%s", output, tc.expected)
			}
		})
	}
}
