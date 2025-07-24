package formatter

import (
	"strings"
	"testing"
)

type formatTestCase struct {
	name     string
	input    string
	expected string
}

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
