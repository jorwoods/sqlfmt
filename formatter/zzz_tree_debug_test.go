package formatter

import (
	"fmt"
	"testing"
)

func TestTreeRenderDebug(t *testing.T) {
	cases := []struct {
		name  string
		input string
		rules RulesConfig
	}{
		{"simple", `select id, name from users where age > 30`, RulesConfig{UppercaseKeywords: true, AlignClauses: true, OperatorSpacing: true}},
		{"select_list_format", `select id, name, age, email, country from users`, RulesConfig{FormatSelectList: true}},
		{"join", `select id from a join b on a.id = b.id`, RulesConfig{UppercaseKeywords: true, AlignClauses: true, OperatorSpacing: true}},
		{"and_or", `select id from t where a = 1 and b = 2 or c = 3`, RulesConfig{UppercaseKeywords: true, AlignClauses: true, NewlineBeforeAndOr: true, OperatorSpacing: true}},
		{"cte", `with a as (select id from t), b as (select id from s) select * from a join b on a.id = b.id`, RulesConfig{UppercaseKeywords: true, AlignClauses: true, CTEFormatting: true, NewlineBeforeJoin: true, NewlineBeforeOn: true, OperatorSpacing: true}},
		{"order_direction", `select id from t order by name, created_at desc, id`, RulesConfig{UppercaseKeywords: true, AlignClauses: true, NormalizeOrderDirection: true}},
		{"require_as", `select id user_id, name as user_name from users u`, RulesConfig{UppercaseKeywords: true, RequireExplicitAS: true}},
		{"window_fn", `select row_number() over (partition by x order by y desc) as rn from t`, RulesConfig{UppercaseKeywords: true, AlignClauses: true, OperatorSpacing: true}},
		{"create_table_as", `create or replace temporary table foo as select id from t`, RulesConfig{UppercaseKeywords: true, AlignClauses: true, OperatorSpacing: true}},
		{"group_having", `select id from t group by id having count(*) > 1`, RulesConfig{UppercaseKeywords: true, AlignClauses: true, OperatorSpacing: true}},
		{"qualify_window", `select id from t qualify row_number() over (partition by id order by ts desc) = 1`, RulesConfig{UppercaseKeywords: true, AlignClauses: true, OperatorSpacing: true}},
		{"partition_as_identifier", `select partition from t`, RulesConfig{UppercaseKeywords: true, AlignClauses: true, OperatorSpacing: true}},
	}
	for _, tc := range cases {
		stream, tree := lexAndParseTree(tc.input)
		applyTokenRules(stream, tc.rules)
		out := renderTree(stream, tree, tc.rules)
		fmt.Printf("=== %s ===\nIN:  %s\nOUT: %q\n\n", tc.name, tc.input, out)
	}
}
