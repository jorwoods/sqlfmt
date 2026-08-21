package formatter

import (
	"strings"

	antlr "github.com/antlr4-go/antlr/v4"
	"github.com/jorwoods/sqlfmt/parser"
)

// lexAndParseTree parses input the same way lexAndParse does, but also
// exposes the root parse tree instead of discarding it. p.BuildParseTrees
// is already set by lexAndParse's twin below; kept as a separate entry
// point so the existing token-stream pipeline (lexAndParse) is untouched
// while this tree-based renderer is developed on a branch.
func lexAndParseTree(input string) (antlr.TokenStream, parser.ISnowflake_fileContext) {
	is := antlr.NewInputStream(input)
	lexer := parser.NewSnowflakeLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := parser.NewSnowflakeParser(stream)
	p.BuildParseTrees = true
	tree := p.Snowflake_file()

	return stream, tree
}

// treeEmitter walks a parse tree and produces formatted SQL text, using the
// grammar's own structure (select_list_elem, join_clause, search_condition,
// order_item, common_table_expression, create_table_as_select, ...) instead
// of scanning the flat token stream for clause/scope boundaries.
type treeEmitter struct {
	tokens antlr.TokenStream
	rules  RulesConfig
}

// renderTree is the entry point: format the whole parsed script.
func renderTree(tokens antlr.TokenStream, tree parser.ISnowflake_fileContext, rules RulesConfig) string {
	e := &treeEmitter{tokens: tokens, rules: rules}
	if tree == nil || tree.Batch() == nil {
		return ""
	}
	var stmts []string
	for _, cmd := range tree.Batch().AllSql_command() {
		stmts = append(stmts, strings.Join(e.emitSqlCommand(cmd), "\n"))
	}
	return strings.Join(stmts, "\n;\n\n")
}

// --- token-range fallback (also the leaf-level expr/predicate renderer) ---

// emitTokenRange renders every default-channel token between startIdx and
// stopIdx (inclusive) as one flat, correctly-spaced string, reusing the
// exact pairwise spacing rules joinTokens already applies (dot spacing,
// unary sign handling, comma/paren spacing, operator spacing). This is the
// leaf-level renderer for any expr/predicate subtree, and also the fallback
// for statement types this emitter doesn't specially format.
func (e *treeEmitter) emitTokenRange(startIdx, stopIdx int) string {
	var texts []string
	for i := startIdx; i <= stopIdx && i < e.tokens.Size(); i++ {
		tok := e.tokens.Get(i)
		if tok.GetChannel() != antlr.TokenDefaultChannel {
			continue
		}
		text := tok.GetText()
		if text == "" || strings.EqualFold(text, "<EOF>") {
			continue
		}
		texts = append(texts, text)
	}
	return joinTokens(texts, e.rules.OperatorSpacing)
}

func (e *treeEmitter) emitRuleRange(ctx antlr.ParserRuleContext) string {
	if ctx == nil {
		return ""
	}
	return e.emitTokenRange(ctx.GetStart().GetTokenIndex(), ctx.GetStop().GetTokenIndex())
}

// --- top-level statement dispatch ---

// emitSqlCommand formats one statement, returning its output as a slice of
// lines. query_statement and CREATE TABLE AS get real structural formatting;
// everything else (ALTER/DROP/INSERT/UPDATE/DELETE/MERGE/other CREATE
// variants/...) falls back to a flat token-range rendering, matching this
// engine's stance that full coverage of every DDL/DML statement type is out
// of scope -- only what the formatter already promises to format well.
func (e *treeEmitter) emitSqlCommand(cmd parser.ISql_commandContext) []string {
	if dml := cmd.Dml_command(); dml != nil {
		if qs := dml.Query_statement(); qs != nil {
			return e.emitQueryStatement(qs, "")
		}
	}
	if ddl := cmd.Ddl_command(); ddl != nil {
		if create := ddl.Create_command(); create != nil {
			if ctas := create.Create_table_as_select(); ctas != nil {
				return e.emitCreateTableAsSelect(ctas)
			}
		}
	}
	return []string{e.emitRuleRange(cmd.(antlr.ParserRuleContext))}
}

// --- CREATE TABLE AS ---

func (e *treeEmitter) emitCreateTableAsSelect(ctx parser.ICreate_table_as_selectContext) []string {
	// Header: CREATE [OR REPLACE] [TEMPORARY|TRANSIENT] TABLE <name> AS
	headerStop := ctx.AS().GetSymbol().GetTokenIndex()
	header := e.emitTokenRange(ctx.GetStart().GetTokenIndex(), headerStop)
	lines := []string{header}
	body := e.emitQueryStatement(ctx.Query_statement(), "  ")
	lines = append(lines, body...)
	return lines
}

// --- query_statement: with_expression? select_statement_in_parentheses set_operators* ---

func (e *treeEmitter) emitQueryStatement(ctx parser.IQuery_statementContext, indent string) []string {
	var lines []string
	if with := ctx.With_expression(); with != nil {
		lines = append(lines, e.emitWithExpression(with, indent)...)
	}
	lines = append(lines, e.emitSelectStatementInParens(ctx.Select_statement_in_parentheses(), indent)...)
	for _, setOp := range ctx.AllSet_operators() {
		lines = append(lines, e.emitSetOperator(setOp, indent)...)
	}
	return lines
}

func (e *treeEmitter) emitSetOperator(ctx parser.ISet_operatorsContext, indent string) []string {
	op := "UNION"
	switch {
	case ctx.UNION() != nil:
		op = upperKeywordText(ctx.UNION())
		if ctx.ALL() != nil {
			op += " " + upperKeywordText(ctx.ALL())
		}
	case ctx.EXCEPT() != nil:
		op = upperKeywordText(ctx.EXCEPT())
	case ctx.MINUS_() != nil:
		op = upperKeywordText(ctx.MINUS_())
	case ctx.INTERSECT() != nil:
		op = upperKeywordText(ctx.INTERSECT())
	}
	lines := []string{indent + op}
	lines = append(lines, e.emitSelectStatementInParens(ctx.Select_statement_in_parentheses(), indent)...)
	return lines
}

// upperKeywordText returns a terminal's already-mutated text (uppercaseKeywords
// runs as a token-mutation pass before the tree is walked, same as the rest
// of applyTokenRules, so this just reads whatever that pass produced).
func upperKeywordText(n antlr.TerminalNode) string {
	if n == nil {
		return ""
	}
	return n.GetText()
}

// --- with_expression / common_table_expression ---

func (e *treeEmitter) emitWithExpression(ctx parser.IWith_expressionContext, indent string) []string {
	ctes := ctx.AllCommon_table_expression()
	var lines []string
	withKw := upperKeywordText(ctx.WITH())
	for i, cte := range ctes {
		prefix := indent + withKw + " "
		if i > 0 {
			prefix = indent + ","
		}
		if i == 0 && ctx.RECURSIVE() != nil {
			prefix = indent + withKw + " " + upperKeywordText(ctx.RECURSIVE()) + " "
		}
		lines = append(lines, e.emitCommonTableExpression(cte, prefix, indent)...)
		if i < len(ctes)-1 {
			lines = append(lines, "")
		}
	}
	if len(ctes) > 0 {
		lines = append(lines, "")
	}
	return lines
}

func (e *treeEmitter) emitCommonTableExpression(ctx parser.ICommon_table_expressionContext, headPrefix, indent string) []string {
	name := e.emitRuleRange(ctx.Id_().(antlr.ParserRuleContext))
	if cols := ctx.Column_list(); cols != nil {
		name += " (" + e.emitRuleRange(cols.(antlr.ParserRuleContext)) + ")"
	}
	asKw := upperKeywordText(ctx.AS())
	lines := []string{headPrefix + name + " " + asKw + " ("}
	body := e.emitSelectStatementInParensBody(ctx.Select_statement_in_parentheses(), indent+"  ")
	lines = append(lines, body...)
	lines = append(lines, indent+")")
	return lines
}

// --- select_statement_in_parentheses ---

func (e *treeEmitter) emitSelectStatementInParens(ctx parser.ISelect_statement_in_parenthesesContext, indent string) []string {
	if ctx == nil {
		return nil
	}
	if ctx.LR_BRACKET() != nil {
		lines := []string{indent + "("}
		lines = append(lines, e.emitSelectStatementInParens(ctx.Select_statement_in_parentheses(), indent+"  ")...)
		lines = append(lines, indent+")")
		return lines
	}
	return e.emitSelectStatementInParensBody(ctx, indent)
}

// emitSelectStatementInParensBody emits the content without adding an extra
// wrapping paren pair -- used for a CTE body ("name AS (" already supplies
// the paren) and the redundant select_statement_in_parentheses wrapper the
// grammar allows.
func (e *treeEmitter) emitSelectStatementInParensBody(ctx parser.ISelect_statement_in_parenthesesContext, indent string) []string {
	if with := ctx.With_expression(); with != nil {
		return e.emitWithExpression(with, indent)
	}
	if stmt := ctx.Select_statement(); stmt != nil {
		return e.emitSelectStatement(stmt, indent)
	}
	if ctx.LR_BRACKET() != nil {
		// The caller (e.g. a CTE's "name AS (") already supplies the
		// wrapping parens for this level; unwrap one redundant
		// select_statement_in_parentheses layer without adding another pair.
		return e.emitSelectStatementInParensBody(ctx.Select_statement_in_parentheses(), indent)
	}
	if setOp := ctx.Set_operators(); setOp != nil {
		lines := e.emitSelectStatementInParensBody(ctx.Select_statement_in_parentheses(), indent)
		lines = append(lines, e.emitSetOperator(setOp, indent)...)
		return lines
	}
	return nil
}

// --- select_statement: select_clause select_optional_clauses limit_clause? ---

func (e *treeEmitter) emitSelectStatement(ctx parser.ISelect_statementContext, indent string) []string {
	var lines []string
	lines = append(lines, e.emitSelectClauseLines(ctx.Select_clause(), indent)...)
	if opt := ctx.Select_optional_clauses(); opt != nil {
		lines = append(lines, e.emitSelectOptionalClauses(opt, indent)...)
	}
	if lim := ctx.Limit_clause(); lim != nil {
		lines = append(lines, indent+" "+e.emitRuleRange(lim.(antlr.ParserRuleContext)))
	}
	return lines
}

func (e *treeEmitter) emitSelectClauseLines(ctx parser.ISelect_clauseContext, indent string) []string {
	selectKw := upperKeywordText(ctx.SELECT())
	listNoTop := ctx.Select_list_no_top()
	prefix := selectKw
	if ad := listNoTop.All_distinct(); ad != nil {
		prefix += " " + e.emitRuleRange(ad.(antlr.ParserRuleContext))
	}
	elems := listNoTop.Select_list().AllSelect_list_elem()
	items := make([]string, len(elems))
	for i, el := range elems {
		items[i] = e.emitSelectListElem(el)
	}
	if e.rules.FormatSelectList && len(items) > 3 {
		itemIndent := indent + strings.Repeat(" ", len(prefix)+1)
		lines := make([]string, 0, len(items))
		lines = append(lines, indent+prefix+" "+items[0]+",")
		for i := 1; i < len(items)-1; i++ {
			lines = append(lines, itemIndent+items[i]+",")
		}
		lines = append(lines, itemIndent+items[len(items)-1])
		return lines
	}
	return []string{indent + prefix + " " + strings.Join(items, ", ")}
}

// emitSelectListElem renders one SELECT-list item. Unlike the old
// paren-depth item-grouping hack, the grammar hands us the expression and
// its optional alias as two separate, already-correct children -- no
// scanning needed to find where one item ends and the next begins, and no
// scanning needed to detect whether an alias is present or already has AS.
func (e *treeEmitter) emitSelectListElem(ctx parser.ISelect_list_elemContext) string {
	if ctx.Column_elem_star() != nil {
		return e.emitRuleRange(ctx.(antlr.ParserRuleContext))
	}
	var base string
	var aliasCtx parser.IAs_aliasContext
	switch {
	case ctx.Column_elem() != nil:
		base = e.emitRuleRange(ctx.Column_elem().(antlr.ParserRuleContext))
		aliasCtx = ctx.As_alias()
	case ctx.Expression_elem() != nil:
		base = e.emitExpressionElem(ctx.Expression_elem())
		aliasCtx = ctx.As_alias()
	default:
		return e.emitRuleRange(ctx.(antlr.ParserRuleContext))
	}
	if aliasCtx == nil {
		return base
	}
	aliasText := e.emitRuleRange(aliasCtx.Alias().(antlr.ParserRuleContext))
	if a := aliasCtx.AS(); a != nil {
		return base + " " + a.GetText() + " " + aliasText
	}
	if e.rules.RequireExplicitAS {
		return base + " AS " + aliasText
	}
	return base + " " + aliasText
}

// --- select_optional_clauses ---

func (e *treeEmitter) emitSelectOptionalClauses(ctx parser.ISelect_optional_clausesContext, indent string) []string {
	var lines []string
	if from := ctx.From_clause(); from != nil {
		lines = append(lines, e.emitFromClause(from, indent)...)
	}
	if where := ctx.Where_clause(); where != nil {
		lines = append(lines, e.emitClauseWithCondition(indent+" ", upperKeywordText(where.WHERE()), where.Search_condition())...)
	}
	if grp := ctx.Group_by_clause(); grp != nil {
		lines = append(lines, e.emitGroupByClause(grp, indent)...)
	}
	if hav := ctx.Having_clause(); hav != nil {
		lines = append(lines, e.emitClauseWithCondition(indent+" ", upperKeywordText(hav.HAVING()), hav.Search_condition())...)
	}
	if qual := ctx.Qualify_clause(); qual != nil {
		lines = append(lines, indent+" "+upperKeywordText(qual.QUALIFY())+" "+e.emitExpr(qual.Expr()))
	}
	if ord := ctx.Order_by_clause(); ord != nil {
		lines = append(lines, e.emitOrderByClause(ord, indent)...)
	}
	return lines
}

func (e *treeEmitter) emitClauseWithCondition(indent, keyword string, cond parser.ISearch_conditionContext) []string {
	condLines := e.emitSearchCondition(cond)
	condLines[0] = indent + keyword + " " + condLines[0]
	return condLines
}

// emitSearchCondition recursively splits a search_condition on its
// top-level AND/OR (search_condition is a real left-recursive binary tree
// in the grammar, so there's no currentClause bookkeeping needed to know
// whether we're "inside" a WHERE/HAVING/QUALIFY the way the old
// booleanOpTokens+wherelikeClauseTypes token scan required). Anything else
// (a bare predicate, or a parenthesized/NOT-prefixed condition) is left as
// one flat line -- matching existing behavior of never breaking inside parens.
func (e *treeEmitter) emitSearchCondition(ctx parser.ISearch_conditionContext) []string {
	if e.rules.NewlineBeforeAndOr {
		if ctx.AND() != nil || ctx.OR() != nil {
			conds := ctx.AllSearch_condition()
			return e.joinAndOrLines(e.emitSearchCondition(conds[0]), e.emitSearchCondition(conds[1]), ctx.AND(), ctx.OR())
		}
		// predicate's catch-all alternative is bare `expr`, and AND/OR are
		// themselves expr alternatives (not just search_condition's) -- for
		// an ordinary boolean condition like "a = 1 AND b = 2", the whole
		// thing parses as one predicate wrapping one expr, with the AND/OR
		// split living inside that expr tree instead of at this level.
		if pred := ctx.Predicate(); pred != nil {
			if expr := solePredicateExpr(pred); expr != nil {
				return e.emitExprCondition(expr)
			}
		}
	}
	return []string{e.emitRuleRange(ctx.(antlr.ParserRuleContext))}
}

// solePredicateExpr returns predicate's expr child when predicate matched
// its catch-all "| expr" alternative (no comparison/BETWEEN/IN/LIKE/IS/...
// set), nil otherwise.
func solePredicateExpr(p parser.IPredicateContext) parser.IExprContext {
	if p.Comparison_operator() != nil || p.EXISTS() != nil || p.BETWEEN() != nil ||
		p.IN() != nil || p.LIKE() != nil || p.ILIKE() != nil || p.RLIKE() != nil || p.IS() != nil {
		return nil
	}
	exprs := p.AllExpr()
	if len(exprs) != 1 {
		return nil
	}
	return exprs[0]
}

func (e *treeEmitter) emitExprCondition(ctx parser.IExprContext) []string {
	if ctx.AND() != nil || ctx.OR() != nil {
		exprs := ctx.AllExpr()
		return e.joinAndOrLines(e.emitExprCondition(exprs[0]), e.emitExprCondition(exprs[1]), ctx.AND(), ctx.OR())
	}
	return []string{e.emitExpr(ctx)}
}

// joinAndOrLines appends right's lines onto left's, with right's first line
// prefixed by the AND/OR keyword and right-aligned to the existing 6-column
// booleanOpIndent convention (3 spaces + "AND", 4 spaces + "OR").
func (e *treeEmitter) joinAndOrLines(left, right []string, and, or antlr.TerminalNode) []string {
	kw, pad := upperKeywordText(and), "   "
	if or != nil {
		kw, pad = upperKeywordText(or), "    "
	}
	lines := append([]string{}, left...)
	lines = append(lines, pad+kw+" "+strings.TrimLeft(right[0], " "))
	lines = append(lines, right[1:]...)
	return lines
}

// --- GROUP BY / ORDER BY ---

func (e *treeEmitter) emitGroupByClause(ctx parser.IGroup_by_clauseContext, indent string) []string {
	head := indent + " " + upperKeywordText(ctx.GROUP()) + " " + upperKeywordText(ctx.BY())
	if ctx.ALL() != nil {
		return []string{head + " " + upperKeywordText(ctx.ALL())}
	}
	if list := ctx.Group_by_list(); list != nil {
		items := make([]string, 0)
		for _, el := range list.AllGroup_by_elem() {
			items = append(items, e.emitRuleRange(el.(antlr.ParserRuleContext)))
		}
		lines := []string{head + " " + strings.Join(items, ", ")}
		if hav := ctx.Having_clause(); hav != nil {
			lines = append(lines, e.emitClauseWithCondition(indent+" ", upperKeywordText(hav.HAVING()), hav.Search_condition())...)
		}
		return lines
	}
	// CUBE/GROUPING SETS/ROLLUP(...) form: fall back to flattening the rest.
	return []string{e.emitRuleRange(ctx.(antlr.ParserRuleContext))}
}

func (e *treeEmitter) emitOrderByClause(ctx parser.IOrder_by_clauseContext, indent string) []string {
	head := indent + " " + upperKeywordText(ctx.ORDER()) + " " + upperKeywordText(ctx.BY())
	items := make([]string, 0)
	for _, item := range ctx.AllOrder_item() {
		items = append(items, e.emitOrderItem(item))
	}
	return []string{head + " " + strings.Join(items, ", ")}
}

// emitOrderItem reads ASC()/DESC() directly off the tree -- normalize_order_direction
// no longer needs to text-scan for "ORDER BY " and guess where the item list
// ends (the bug that let a window function's nested ORDER BY get corrupted);
// the grammar already knows exactly which items belong to this ORDER BY and
// whether each one already has an explicit direction.
func (e *treeEmitter) emitOrderItem(ctx parser.IOrder_itemContext) string {
	var base string
	switch {
	case ctx.Id_() != nil:
		base = e.emitRuleRange(ctx.Id_().(antlr.ParserRuleContext))
	case ctx.Num() != nil:
		base = e.emitRuleRange(ctx.Num().(antlr.ParserRuleContext))
	case ctx.Expr() != nil:
		base = e.emitExpr(ctx.Expr())
	}
	dir := ""
	if ctx.ASC() != nil {
		dir = upperKeywordText(ctx.ASC())
	} else if ctx.DESC() != nil {
		dir = upperKeywordText(ctx.DESC())
	} else if e.rules.NormalizeOrderDirection {
		dir = "ASC"
	}
	if dir != "" {
		base += " " + dir
	}
	if ctx.NULLS() != nil {
		nulls := upperKeywordText(ctx.NULLS())
		if ctx.FIRST() != nil {
			base += " " + nulls + " " + upperKeywordText(ctx.FIRST())
		} else if ctx.LAST() != nil {
			base += " " + nulls + " " + upperKeywordText(ctx.LAST())
		}
	}
	return base
}

// --- FROM / JOIN ---

func (e *treeEmitter) emitFromClause(ctx parser.IFrom_clauseContext, indent string) []string {
	sources := ctx.Table_sources().AllTable_source()
	srcTexts := make([]string, len(sources))
	var joinLines []string
	for i, src := range sources {
		joined := src.Table_source_item_joined()
		srcTexts[i] = e.emitObjectRefWithAlias(joined.Object_ref())
		joinLines = append(joinLines, e.emitJoinClauses(joined.AllJoin_clause(), indent)...)
	}
	lines := []string{indent + "  " + upperKeywordText(ctx.FROM()) + " " + strings.Join(srcTexts, ", ")}
	lines = append(lines, joinLines...)
	return lines
}

func (e *treeEmitter) emitJoinClauses(joins []parser.IJoin_clauseContext, indent string) []string {
	var lines []string
	for _, j := range joins {
		head := indent + "  "
		if qt := j.Join_type(); qt != nil {
			head += e.emitRuleRange(qt.(antlr.ParserRuleContext)) + " "
		} else if j.NATURAL() != nil {
			head += upperKeywordText(j.NATURAL()) + " "
			if oj := j.Outer_join(); oj != nil {
				head += e.emitRuleRange(oj.(antlr.ParserRuleContext)) + " "
			}
		} else if j.CROSS() != nil {
			head += upperKeywordText(j.CROSS()) + " "
		}
		head += upperKeywordText(j.JOIN()) + " " + e.emitObjectRefWithAlias(j.Object_ref())
		if cond := j.Search_condition(); cond != nil {
			condLines := e.emitSearchCondition(cond)
			condLines[0] = indent + "    " + upperKeywordText(j.ON()) + " " + condLines[0]
			lines = append(lines, head)
			lines = append(lines, condLines...)
			continue
		}
		if j.USING() != nil {
			head += " " + upperKeywordText(j.USING()) + " (" + e.emitRuleRange(j.Column_list().(antlr.ParserRuleContext)) + ")"
		}
		lines = append(lines, head)
	}
	return lines
}

func (e *treeEmitter) emitObjectRefWithAlias(ctx parser.IObject_refContext) string {
	return e.emitRuleRange(ctx.(antlr.ParserRuleContext))
}

// --- expr / OVER (window functions) ---

// emitExpressionElem renders a select_list_elem/group_by_elem's expression_elem,
// which wraps either a bare expr or a predicate (whose own catch-all
// alternative is itself a bare expr).
func (e *treeEmitter) emitExpressionElem(ctx parser.IExpression_elemContext) string {
	if expr := ctx.Expr(); expr != nil {
		return e.emitExpr(expr)
	}
	if pred := ctx.Predicate(); pred != nil {
		if expr := solePredicateExpr(pred); expr != nil {
			return e.emitExpr(expr)
		}
		return e.emitRuleRange(pred.(antlr.ParserRuleContext))
	}
	return e.emitRuleRange(ctx.(antlr.ParserRuleContext))
}

// emitExpr renders an expr subtree. Every expr shape still falls back to a
// flat token range (emitTokenRange) except one that carries a window
// function's over_clause, e.g. ROW_NUMBER() OVER (...) or SUM(x) OVER (...):
// PARTITION inside partition_by is a grammar-ambiguous token -- Snowflake
// allows a column literally named "partition" (see non_reserved_words in
// SnowflakeParser.g4) -- so the blanket uppercaseKeywords token-mutation
// pass in rules.go deliberately never touches it. Only here, where the tree
// confirms this token sits inside a real partition_by node and not an
// identifier position, is it safe to uppercase.
func (e *treeEmitter) emitExpr(ctx parser.IExprContext) string {
	if over := exprOverClause(ctx); over != nil {
		base := e.emitTokenRange(ctx.GetStart().GetTokenIndex(), over.GetStart().GetTokenIndex()-1)
		return base + " " + e.emitOverClause(over)
	}
	// Recurse through comparisons (l comparison_operator r) so a windowed
	// call on either side -- e.g. QUALIFY row_number() OVER (...) = 1 --
	// still gets the PARTITION BY treatment above instead of falling
	// straight to the flat fallback.
	if ctx.Comparison_operator() != nil {
		if operands := ctx.AllExpr(); len(operands) == 2 {
			return e.emitExpr(operands[0]) + " " +
				e.emitRuleRange(ctx.Comparison_operator().(antlr.ParserRuleContext)) + " " +
				e.emitExpr(operands[1])
		}
	}
	return e.emitRuleRange(ctx.(antlr.ParserRuleContext))
}

// exprOverClause finds the over_clause governing this expr, if any. Ranking
// functions (ROW_NUMBER, RANK, DENSE_RANK, NTILE, LEAD, LAG, FIRST_VALUE,
// LAST_VALUE) carry their over_clause inside function_call/ranking_windowed_function;
// any other windowed call (e.g. SUM(x) OVER (...)) uses expr's own
// "expr over_clause" alternative instead.
func exprOverClause(ctx parser.IExprContext) parser.IOver_clauseContext {
	if over := ctx.Over_clause(); over != nil {
		return over
	}
	if fc := ctx.Function_call(); fc != nil {
		if rwf := fc.Ranking_windowed_function(); rwf != nil {
			return rwf.Over_clause()
		}
	}
	return nil
}

func (e *treeEmitter) emitOverClause(ctx parser.IOver_clauseContext) string {
	var inner []string
	if pb := ctx.Partition_by(); pb != nil {
		inner = append(inner, e.emitPartitionBy(pb))
	}
	if ob := ctx.Order_by_expr(); ob != nil {
		inner = append(inner, e.emitRuleRange(ob.(antlr.ParserRuleContext)))
	}
	return upperKeywordText(ctx.OVER()) + " (" + strings.Join(inner, " ") + ")"
}

// emitPartitionBy explicitly uppercases PARTITION (see emitExpr) since the
// global token-mutation pass can't safely do it for this ambiguous token.
func (e *treeEmitter) emitPartitionBy(ctx parser.IPartition_byContext) string {
	kw := ctx.PARTITION().GetText()
	if e.rules.UppercaseKeywords {
		kw = strings.ToUpper(kw)
	}
	return kw + " " + upperKeywordText(ctx.BY()) + " " + e.emitRuleRange(ctx.Expr_list().(antlr.ParserRuleContext))
}
