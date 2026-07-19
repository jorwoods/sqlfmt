
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
		       name: "newline_before_join bare JOIN, flat path",
		       input: `select id from users join orders on uid = oid`,
		       expected: "SELECT id FROM users\nJOIN orders ON uid = oid",
		       rules: RulesConfig{UppercaseKeywords: true, NewlineBeforeJoin: true, OperatorSpacing: true},
	       },
	       {
		       name: "newline_before_join INNER JOIN with align_clauses",
		       input: `select id from users inner join orders on uid = oid`,
		       expected: "SELECT id\n  FROM users\n  INNER JOIN orders ON uid = oid",
		       rules: RulesConfig{UppercaseKeywords: true, AlignClauses: true, NewlineBeforeJoin: true, OperatorSpacing: true},
	       },
	       {
		       name: "newline_before_join LEFT OUTER JOIN with align_clauses",
		       input: `select id from users left outer join orders on uid = oid`,
		       expected: "SELECT id\n  FROM users\n  LEFT OUTER JOIN orders ON uid = oid",
		       rules: RulesConfig{UppercaseKeywords: true, AlignClauses: true, NewlineBeforeJoin: true, OperatorSpacing: true},
	       },
	       {
		       name: "newline_before_join multiple joins with align_clauses",
		       input: `select id from a inner join b on a_id = b_id left join c on b_id = c_id where age > 30`,
		       expected: "SELECT id\n  FROM a\n  INNER JOIN b ON a_id = b_id\n  LEFT JOIN c ON b_id = c_id\n WHERE age > 30",
		       rules: RulesConfig{UppercaseKeywords: true, AlignClauses: true, NewlineBeforeJoin: true, OperatorSpacing: true},
	       },
	       {
		       name: "newline_before_on puts ON on its own line (flat path)",
		       input: `select id from users inner join orders on uid = oid`,
		       expected: "SELECT id FROM users\nINNER JOIN orders\n    ON uid = oid",
		       rules: RulesConfig{UppercaseKeywords: true, NewlineBeforeJoin: true, NewlineBeforeOn: true, OperatorSpacing: true},
	       },
	       {
		       name: "newline_before_on with align_clauses",
		       input: `select id from users inner join orders on uid = oid`,
		       expected: "SELECT id\n  FROM users\n  INNER JOIN orders\n    ON uid = oid",
		       rules: RulesConfig{UppercaseKeywords: true, AlignClauses: true, NewlineBeforeJoin: true, NewlineBeforeOn: true, OperatorSpacing: true},
	       },
	       {
		       name: "newline_before_on multiple joins",
		       input: `select id from a inner join b on a_id = b_id left join c on b_id = c_id`,
		       expected: "SELECT id\n  FROM a\n  INNER JOIN b\n    ON a_id = b_id\n  LEFT JOIN c\n    ON b_id = c_id",
		       rules: RulesConfig{UppercaseKeywords: true, AlignClauses: true, NewlineBeforeJoin: true, NewlineBeforeOn: true, OperatorSpacing: true},
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
		// CASE at col 7 ("SELECT " = 7), so WHEN/ELSE indent = 11, END indent = 7.
		name:     "indent_case_when: searched CASE, flat path",
		input:    `select case when a = 1 then 'x' else 'y' end from t`,
		expected: "SELECT CASE\n           WHEN a = 1 THEN 'x'\n           ELSE 'y'\n       END FROM t",
		rules:    RulesConfig{UppercaseKeywords: true, IndentCaseWhen: true, OperatorSpacing: true},
	},
	{
		// CASE at col 11 ("SELECT id, " = 11), WHEN/ELSE indent = 15, END indent = 11.
		name:     "indent_case_when: valued CASE with align_clauses",
		input:    `select id, case status when 1 then 'active' when 2 then 'inactive' else 'unknown' end from users`,
		expected: "SELECT id, CASE status\n               WHEN 1 THEN 'active'\n               WHEN 2 THEN 'inactive'\n               ELSE 'unknown'\n           END\n  FROM users",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, IndentCaseWhen: true, OperatorSpacing: true},
	},
	{
		name:     "indent_case_when: CASE in WHERE clause",
		input:    `select id from t where case a when 1 then true else false end = true`,
		expected: "SELECT id\n  FROM t\n WHERE CASE a\n           WHEN 1 THEN TRUE\n           ELSE FALSE\n       END = TRUE",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, NormalizeBoolean: true, IndentCaseWhen: true, OperatorSpacing: true},
	},
	{
		// Outer CASE at col 7; inner CASE at col 27 (after "           WHEN a = 1 THEN ").
		name:     "indent_case_when: nested CASE",
		input:    `select case when a = 1 then case when b = 2 then 'x' else 'y' end else 'z' end from t`,
		expected: "SELECT CASE\n           WHEN a = 1 THEN CASE\n                               WHEN b = 2 THEN 'x'\n                               ELSE 'y'\n                           END\n           ELSE 'z'\n       END FROM t",
		rules:    RulesConfig{UppercaseKeywords: true, IndentCaseWhen: true, OperatorSpacing: true},
	},
	{
		name:     "leading_comma: multi-column SELECT with format_select_list",
		input:    `select col_a, col_b, col_c, col_d from t`,
		expected: "SELECT col_a\n     , col_b\n     , col_c\n     , col_d\n  FROM t",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, FormatSelectList: true, LeadingComma: true, OperatorSpacing: true},
	},
	{
		name:     "leading_comma: single-line SELECT unchanged",
		input:    `select col_a, col_b from t`,
		expected: "SELECT col_a, col_b FROM t",
		rules:    RulesConfig{UppercaseKeywords: true, LeadingComma: true, OperatorSpacing: true},
	},
	{
		name:     "leading_comma: with trailing_semicolon",
		input:    `select col_a, col_b, col_c, col_d from t`,
		expected: "SELECT col_a\n     , col_b\n     , col_c\n     , col_d\n  FROM t\n;",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, FormatSelectList: true, LeadingComma: true, TrailingSemicolon: true, OperatorSpacing: true},
	},
	{
		name:     "normalize_null_comparison: = NULL becomes IS NULL",
		input:    `select id from users where name = null`,
		expected: `SELECT id FROM users WHERE name IS NULL`,
		rules:    RulesConfig{UppercaseKeywords: true, NormalizeNullComparison: true, OperatorSpacing: true},
	},
	{
		name:     "normalize_null_comparison: != NULL becomes IS NOT NULL",
		input:    `select id from users where name != null`,
		expected: `SELECT id FROM users WHERE name IS NOT NULL`,
		rules:    RulesConfig{UppercaseKeywords: true, NormalizeNullComparison: true, OperatorSpacing: true},
	},
	{
		name:     "normalize_null_comparison: <> NULL becomes IS NOT NULL",
		input:    `select id from users where name <> null`,
		expected: `SELECT id FROM users WHERE name IS NOT NULL`,
		rules:    RulesConfig{UppercaseKeywords: true, NormalizeNullComparison: true, OperatorSpacing: true},
	},
	{
		name:     "normalize_null_comparison: leaves IS NULL unchanged",
		input:    `select id from users where name is null`,
		expected: `SELECT id FROM users WHERE name IS NULL`,
		rules:    RulesConfig{UppercaseKeywords: true, NormalizeBoolean: true, NormalizeNullComparison: true, OperatorSpacing: true},
	},
	{
		name:     "normalize_null_comparison: leaves non-null comparisons unchanged",
		input:    `select id from users where age = 30`,
		expected: `SELECT id FROM users WHERE age = 30`,
		rules:    RulesConfig{UppercaseKeywords: true, NormalizeNullComparison: true, OperatorSpacing: true},
	},
	{
		name:     "trailing_newline: appended when missing",
		input:    `select id from users`,
		expected: "SELECT id FROM users\n",
		rules:    RulesConfig{UppercaseKeywords: true, TrailingNewline: true},
	},
	{
		name:     "trailing_newline: not doubled when already present",
		input:    "select id from users\n",
		expected: "SELECT id FROM users\n",
		rules:    RulesConfig{UppercaseKeywords: true, TrailingNewline: true},
	},
	{
		name:     "trailing_newline: works with trailing_semicolon",
		input:    `select id from users`,
		expected: "SELECT id FROM users\n;\n",
		rules:    RulesConfig{UppercaseKeywords: true, TrailingSemicolon: true, TrailingNewline: true},
	},
	{
		name:     "newline_before_limit: LIMIT on its own line (flat path)",
		input:    `select id from users limit 10`,
		expected: "SELECT id FROM users\nLIMIT 10",
		rules:    RulesConfig{UppercaseKeywords: true, NewlineBeforeLimit: true, OperatorSpacing: true},
	},
	{
		name:     "newline_before_limit: LIMIT and OFFSET on their own lines with align_clauses",
		input:    `select id from users order by id limit 10 offset 20`,
		expected: "SELECT id\n  FROM users\n ORDER BY id\n LIMIT 10\n OFFSET 20",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, NewlineBeforeLimit: true, OperatorSpacing: true},
	},
	{
		name:     "newline_before_limit: LIMIT with both align_clauses and format_select_list",
		input:    `select id, name, age, email from users order by name limit 10 offset 5`,
		expected: "SELECT id,\n       name,\n       age,\n       email\n  FROM users\n ORDER BY name\n LIMIT 10\n OFFSET 5",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, FormatSelectList: true, NewlineBeforeLimit: true, OperatorSpacing: true},
	},
	{
		name:     "newline_before_limit: LIMIT only, no OFFSET",
		input:    `select id from users where age > 30 limit 5`,
		expected: "SELECT id\n  FROM users\n WHERE age > 30\n LIMIT 5",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, NewlineBeforeLimit: true, OperatorSpacing: true},
	},
	{
		name:     "uppercase_keywords: ASC and DESC are uppercased",
		input:    `select id from users order by name desc, id asc`,
		expected: "SELECT id FROM users ORDER BY name DESC, id ASC",
		rules:    RulesConfig{UppercaseKeywords: true},
	},
	{
		name:     "set operations: UNION on its own line, uppercased",
		input:    `select id from t union select id from s union all select id from r`,
		expected: "SELECT id\n  FROM t\nUNION\nSELECT id\n  FROM s\nUNION ALL\nSELECT id\n  FROM r",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true},
	},
	{
		name:     "set operations: INTERSECT and EXCEPT",
		input:    `select a from x intersect select a from y except select a from z`,
		expected: "SELECT a\n  FROM x\nINTERSECT\nSELECT a\n  FROM y\nEXCEPT\nSELECT a\n  FROM z",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true},
	},
	{
		name:     "cte: WITH uppercased, AS followed by space before paren",
		input:    `with cte as (select id from users) select * from cte`,
		expected: "WITH cte AS (\nSELECT id\n  FROM users)\nSELECT *\n  FROM cte",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true},
	},
	{
		name:     "normalize_order_direction: adds ASC to items without direction",
		input:    `select id from t order by name, created_at desc, id`,
		expected: "SELECT id\n  FROM t\n ORDER BY name ASC, created_at DESC, id ASC",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, NormalizeOrderDirection: true},
	},
	{
		name:     "normalize_order_direction: function call items get ASC",
		input:    `select id from t order by upper(name), id desc`,
		expected: "SELECT id\n  FROM t\n ORDER BY upper(name) ASC, id DESC",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, NormalizeOrderDirection: true},
	},
	{
		name:     "normalize_order_direction: already explicit stays unchanged",
		input:    `select id from t order by name asc, id desc`,
		expected: "SELECT id\n  FROM t\n ORDER BY name ASC, id DESC",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, NormalizeOrderDirection: true},
	},
	{
		name:     "normalize_order_direction: with LIMIT suffix preserved",
		input:    `select id from t order by name limit 5`,
		expected: "SELECT id\n  FROM t\n ORDER BY name ASC\n LIMIT 5",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, NormalizeOrderDirection: true, NewlineBeforeLimit: true},
	},
	{
		name:     "cte_formatting: closing paren on its own line",
		input:    `with cte as (select id from users where active = true) select * from cte`,
		expected: "WITH cte AS (\n  SELECT id\n    FROM users\n   WHERE active = TRUE\n)\n\nSELECT *\n  FROM cte",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, NormalizeBoolean: true, CTEFormatting: true, OperatorSpacing: true},
	},
	{
		name:     "cte_formatting: multiple CTEs",
		input:    `with a as (select id from t), b as (select id from s) select * from a join b on a.id = b.id`,
		expected: "WITH a AS (\n  SELECT id\n    FROM t\n)\n\n,b AS (\n  SELECT id\n    FROM s\n)\n\nSELECT *\n  FROM a\n  JOIN b\n    ON a.id = b.id",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, CTEFormatting: true, NewlineBeforeJoin: true, NewlineBeforeOn: true, OperatorSpacing: true},
	},
	{
		name:     "leading_comma_cte: comma moves to start of next CTE line",
		input:    `with a as (select id from t), b as (select name from s) select * from a, b`,
		expected: "WITH a AS (\n  SELECT id\n    FROM t\n)\n\n,b AS (\n  SELECT name\n    FROM s\n)\n\nSELECT *\n  FROM a, b",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, CTEFormatting: true, LeadingCommaCTE: true, OperatorSpacing: true},
	},
	       {
		       name: "newline_before_and_or: AND not injected outside WHERE (SELECT clause)",
		       input: `select id, case when a = 1 and b = 2 then 'y' else 'n' end from users where c = 3`,
		       expected: "SELECT id, CASE WHEN a = 1 AND b = 2 THEN 'y' ELSE 'n' END FROM users WHERE c = 3",
		       rules: RulesConfig{UppercaseKeywords: true, NewlineBeforeAndOr: true, OperatorSpacing: true},
	       },
	{
		name:     "remove_redundant_parens: simple WHERE condition",
		input:    `select id from users where (id = 1)`,
		expected: "SELECT id FROM users WHERE id = 1",
		rules:    RulesConfig{UppercaseKeywords: true, RemoveRedundantParens: true, OperatorSpacing: true},
	},
	{
		name:     "remove_redundant_parens: multiple AND conditions each wrapped",
		input:    `select id from users where (a = 1) and (b = 2)`,
		expected: "SELECT id FROM users WHERE a = 1 AND b = 2",
		rules:    RulesConfig{UppercaseKeywords: true, RemoveRedundantParens: true, OperatorSpacing: true},
	},
	{
		name:     "remove_redundant_parens: preserves function call parens",
		input:    `select count(id) from users where upper(name) = 'ALICE'`,
		expected: "SELECT count(id) FROM users WHERE upper(name) = 'ALICE'",
		rules:    RulesConfig{UppercaseKeywords: true, RemoveRedundantParens: true, OperatorSpacing: true},
	},
	{
		name:     "remove_redundant_parens: preserves subquery parens",
		input:    `select id from users where id in (select id from admins)`,
		expected: "SELECT id FROM users WHERE id IN (SELECT id FROM admins)",
		rules:    RulesConfig{UppercaseKeywords: true, RemoveRedundantParens: true, OperatorSpacing: true},
	},
	{
		name:     "uppercase IN keyword",
		input:    `select id from users where id in (1, 2, 3)`,
		expected: "SELECT id FROM users WHERE id IN (1, 2, 3)",
		rules:    RulesConfig{UppercaseKeywords: true},
	},
	{
		name:     "newline_before_set_op: UNION on its own line (flat path)",
		input:    `select id from t union select id from s`,
		expected: "SELECT id FROM t\nUNION\nSELECT id FROM s",
		rules:    RulesConfig{UppercaseKeywords: true, NewlineBeforeSetOp: true},
	},
	{
		name:     "newline_before_set_op: UNION ALL on its own line",
		input:    `select id from t union all select id from s`,
		expected: "SELECT id FROM t\nUNION ALL\nSELECT id FROM s",
		rules:    RulesConfig{UppercaseKeywords: true, NewlineBeforeSetOp: true},
	},
	{
		name:     "newline_before_set_op: disabled keeps UNION inline",
		input:    `select id from t union select id from s`,
		expected: "SELECT id FROM t UNION SELECT id FROM s",
		rules:    RulesConfig{UppercaseKeywords: true, NewlineBeforeSetOp: false},
	},
	{
		name:     "indent_subquery: FROM clause subquery",
		input:    `select * from (select id from users)`,
		expected: "SELECT * FROM (\n  SELECT id FROM users\n)",
		rules:    RulesConfig{UppercaseKeywords: true, IndentSubquery: true},
	},
	{
		name:     "indent_subquery: WHERE IN subquery",
		input:    `select id from users where id in (select id from admins)`,
		expected: "SELECT id FROM users WHERE id IN (\n  SELECT id FROM admins\n)",
		rules:    RulesConfig{UppercaseKeywords: true, IndentSubquery: true},
	},
	{
		name:     "indent_subquery: nested subqueries get extra indentation per level",
		input:    `select * from (select id from (select id from base) as t)`,
		expected: "SELECT * FROM (\n  SELECT id FROM (\n    SELECT id FROM base\n  ) AS t\n)",
		rules:    RulesConfig{UppercaseKeywords: true, IndentSubquery: true},
	},
	{
		name:     "indent_subquery: function call parens unchanged",
		input:    `select count(id) from users`,
		expected: "SELECT count(id) FROM users",
		rules:    RulesConfig{UppercaseKeywords: true, IndentSubquery: true},
	},
	{
		name:     "newline_before_group_by: places GROUP BY on its own line",
		input:    `select id, count(*) from users group by id`,
		expected: "SELECT id, count(*) FROM users\nGROUP BY id",
		rules:    RulesConfig{UppercaseKeywords: true, NewlineBeforeGroupBy: true},
	},
	{
		name:     "newline_before_group_by: no extra newline when already on new line",
		input:    "select id from t\ngroup by id",
		expected: "SELECT id FROM t\nGROUP BY id",
		rules:    RulesConfig{UppercaseKeywords: true, NewlineBeforeGroupBy: true},
	},
	{
		name:     "newline_before_order_by: places ORDER BY on its own line",
		input:    `select id from users order by id`,
		expected: "SELECT id FROM users\nORDER BY id",
		rules:    RulesConfig{UppercaseKeywords: true, NewlineBeforeOrderBy: true},
	},
	{
		name:     "newline_before_having: places HAVING on its own line",
		input:    `select id, count(*) from users group by id having count(*) > 1`,
		expected: "SELECT id, count(*) FROM users GROUP BY id\nHAVING count(*) > 1",
		rules:    RulesConfig{UppercaseKeywords: true, NewlineBeforeHaving: true, OperatorSpacing: true},
	},
	{
		name:     "inline_override: simple single-table query stays on one line",
		input:    "select id from t where x = 1 group by id having id > 1",
		expected: "SELECT id FROM t WHERE x = 1 GROUP BY id HAVING id > 1",
		rules:    RulesConfig{UppercaseKeywords: true, NewlineBeforeGroupBy: true, NewlineBeforeHaving: true, OperatorSpacing: true, InlineOverride: true},
	},
	{
		name:     "inline_override: query with JOIN is not inlined",
		input:    "select id from a join b on a_id = b_id group by id",
		expected: "SELECT id FROM a JOIN b ON a_id = b_id\nGROUP BY id",
		rules:    RulesConfig{UppercaseKeywords: true, NewlineBeforeGroupBy: true, OperatorSpacing: true, InlineOverride: true},
	},
	{
		name:     "inline_override: collapses simple subquery content for indent_subquery",
		input:    "select * from (select id from t where x = 1 group by id) as sub",
		expected: "SELECT * FROM (\n  SELECT id FROM t WHERE x = 1 GROUP BY id\n) AS sub",
		rules:    RulesConfig{UppercaseKeywords: true, NewlineBeforeGroupBy: true, IndentSubquery: true, OperatorSpacing: true, InlineOverride: true},
	},
	{
		name:  "inline_override: long query stays multi-line with aligned formatting",
		input: "select a, b, c, d, e, f, g, h, i from t where alpha = 1 and beta = 2 and gamma = 3 and delta = 4 and epsilon = 5 group by a, b, c, d, e, f, g, h, i",
		// collapsed would be ~150 chars, over the 120-char threshold
		expected: "SELECT a, b, c, d, e, f, g, h, i\n  FROM t\n WHERE alpha = 1\n   AND beta = 2\n   AND gamma = 3\n   AND delta = 4\n   AND epsilon = 5\n GROUP BY a, b, c, d, e, f, g, h, i",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, NewlineBeforeAndOr: true, OperatorSpacing: true, InlineOverride: true},
	},
	{
		name:     "remove_redundant_parens: preserves precedence-changing parens",
		input:    `select id from users where (a = 1 or b = 2) and c = 3`,
		expected: "SELECT id FROM users WHERE (a = 1 OR b = 2) AND c = 3",
		rules:    RulesConfig{UppercaseKeywords: true, RemoveRedundantParens: true, OperatorSpacing: true},
	},
	{
		name:     "remove_redundant_parens: ON clause",
		input:    `select id from a join b on (a_id = b_id)`,
		expected: "SELECT id FROM a JOIN b ON a_id = b_id",
		rules:    RulesConfig{UppercaseKeywords: true, RemoveRedundantParens: true, OperatorSpacing: true},
	},
	{
		name:     "select_list_function_call: function call with alias not split across items",
		input:    `select customer_id, customer_name, customer_group, max(signup_date) signup_date from t`,
		expected: "SELECT customer_id,\n       customer_name,\n       customer_group,\n       max(signup_date) signup_date\n  FROM t",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, FormatSelectList: true, OperatorSpacing: true},
	},
	{
		name:     "format_select_list: function call with alias not split across items (select-list-only path)",
		input:    `select customer_id, customer_name, customer_group, max(signup_date) signup_date from t`,
		expected: "select customer_id,\n       customer_name,\n       customer_group,\n       max(signup_date) signup_date\nfrom t",
		rules:    RulesConfig{FormatSelectList: true, OperatorSpacing: true},
	},
	{
		name:     "select_list_function_call: qualified columns not split across items",
		input:    `select a.id, a.name, b.id, b.name from a join b on a.id = b.id`,
		expected: "SELECT a.id,\n       a.name,\n       b.id,\n       b.name\n  FROM a JOIN b ON a.id = b.id",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, FormatSelectList: true, OperatorSpacing: true},
	},
	{
		name:     "join_on: function calls in ON clause stay on one line",
		input:    `select id from t join s on coalesce(t.id, 0) = coalesce(s.id, 0)`,
		expected: "SELECT id FROM t JOIN s ON coalesce(t.id, 0) = coalesce(s.id, 0)",
		rules:    RulesConfig{UppercaseKeywords: true, OperatorSpacing: true},
	},
	{
		name:     "where: function calls in WHERE clause stay on one line",
		input:    `select id from t where round(amount, 2) > 10 and datediff('day', a, b) < 5`,
		expected: "SELECT id\n  FROM t\n WHERE round(amount, 2) > 10 AND datediff('day', a, b) < 5",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, OperatorSpacing: true},
	},
	{
		name:     "group_by: function calls in GROUP BY stay on one line",
		input:    `select id from t group by date_trunc('month', signup_date), coalesce(region, 'unknown')`,
		expected: "SELECT id\n  FROM t\n GROUP BY date_trunc('month', signup_date), coalesce(region, 'unknown')",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, OperatorSpacing: true},
	},
	{
		name:     "having: function calls in HAVING stay on one line",
		input:    `select id from t having sum(coalesce(amount, 0)) > 100`,
		expected: "SELECT id\n  FROM t\n HAVING sum(coalesce(amount, 0)) > 100",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, OperatorSpacing: true},
	},
	{
		name:     "order_by: function calls in ORDER BY stay on one line",
		input:    `select id from t order by upper(name), round(score, 1) desc`,
		expected: "SELECT id\n  FROM t\n ORDER BY upper(name), round(score, 1) DESC",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, OperatorSpacing: true},
	},
	{
		name:     "qualify: ORDER BY inside window OVER() clause is not a clause boundary",
		input:    `select id from t qualify row_number() over (partition by coalesce(a,b) order by c) = 1`,
		expected: "SELECT id\n  FROM t\n QUALIFY row_number() over (partition BY coalesce(a, b) ORDER BY c) = 1",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, OperatorSpacing: true},
	},
	{
		name:     "format_select_list: window function with ORDER BY inside OVER() not split as separate items",
		input:    `select row_number() over (order by x), y, z, w from t`,
		expected: "select row_number() over (order by x),\n       y,\n       z,\n       w\nfrom t",
		rules:    RulesConfig{FormatSelectList: true, OperatorSpacing: true},
	},
	{
		name:     "tokens_to_text: ORDER BY inside OVER() does not trigger newline_before_order_by",
		input:    `select id, row_number() over (order by x) as rn from t order by id`,
		expected: "SELECT id, row_number() over (ORDER BY x) AS rn FROM t\nORDER BY id",
		rules:    RulesConfig{UppercaseKeywords: true, NewlineBeforeOrderBy: true, OperatorSpacing: true},
	},
	{
		name:     "nested scope: subquery's own WHERE still breaks while its window function's ORDER BY stays inline",
		input:    `select id from (select id, row_number() over (order by x) as rn from t where coalesce(a,b) = 1) s where s.rn = 1`,
		expected: "SELECT id\n  FROM (\nSELECT id, row_number() over (ORDER BY x) AS rn\n  FROM t\n WHERE coalesce(a, b) = 1) s\n WHERE s.rn = 1",
		rules:    RulesConfig{UppercaseKeywords: true, AlignClauses: true, OperatorSpacing: true},
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
				       output := FormatSQLWithConfig(tc.input, cfg)
			       if !tc.rules.TrailingNewline {
				       output = strings.TrimSpace(output)
			       }
				       if output != tc.expected {
					       t.Errorf("unexpected format output\n--- got:\n%s\n--- want:\n%s", output, tc.expected)
				       }
			       })
		       }
}
