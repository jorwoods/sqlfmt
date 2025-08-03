package formatter

import (
	"fmt"
	"strings"

	antlr "github.com/antlr4-go/antlr/v4"
	"github.com/jorwoods/sqlfmt/parser"
)

// CTERefactorEnabled returns true if the config enables the CTE refactor rule.
func CTERefactorEnabled(cfg *Config) bool {
	return cfg == nil || cfg.Rules.RefactorLongSubqueriesToCTE
}

// RefactorLongSubqueriesToCTE rewrites long/non-correlated subqueries as CTEs.
func RefactorLongSubqueriesToCTE(sql string, cfg *Config) string {
	if !CTERefactorEnabled(cfg) {
		return sql
	}
	is := antlr.NewInputStream(sql)
	lexer := parser.NewSnowflakeLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewSnowflakeParser(stream)
	p.BuildParseTrees = true
	tree := p.Snowflake_file()

	// Find subqueries using a custom listener
	listener := &subqueryListener{
		tokens:     stream,
		subqueries: make([]*subqueryInfo, 0),
	}
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	if len(listener.subqueries) == 0 {
		return sql
	}

	// Only refactor subqueries that are long and not correlated
	cteList := make([]string, 0)
	type replacement struct {
		start, stop int
		cteName     string
	}
	replacements := make([]replacement, 0)
	for i, sub := range listener.subqueries {
		if sub.isLong && !sub.isCorrelated {
			cteName := fmt.Sprintf("cte_%d", i+1)
			cteList = append(cteList, fmt.Sprintf("%s AS %s", cteName, sub.text))
			replacements = append(replacements, replacement{start: sub.start, stop: sub.stop, cteName: cteName})
		}
	}
	if len(cteList) == 0 {
		return sql
	}
	// Sort replacements by start index descending to avoid offset issues
	for i := 0; i < len(replacements)-1; i++ {
		for j := i + 1; j < len(replacements); j++ {
			if replacements[i].start < replacements[j].start {
				replacements[i], replacements[j] = replacements[j], replacements[i]
			}
		}
	}
	rewritten := sql
	for _, rep := range replacements {
		if rep.start >= 0 && rep.stop > rep.start && rep.stop <= len(rewritten) {
			rewritten = rewritten[:rep.start] + rep.cteName + rewritten[rep.stop:]
		}
	}
	// Prepend CTEs to the rewritten query
	cteClause := "WITH " + strings.Join(cteList, ", ") + "\n"
	return cteClause + rewritten
}

type subqueryInfo struct {
	text         string
	isLong       bool
	isCorrelated bool
	start        int // start byte offset
	stop         int // stop byte offset (exclusive)
}

type subqueryListener struct {
	*parser.BaseSnowflakeParserListener
	tokens     antlr.TokenStream
	subqueries []*subqueryInfo
}

// Enter every subquery in FROM (SELECT ... inside parens)
func (l *subqueryListener) EnterSelect_statement_in_parentheses(ctx *parser.Select_statement_in_parenthesesContext) {
	parent := ctx.GetParent()
	found := false
	for parent != nil {
		if _, ok := parent.(*parser.Table_source_item_joinedContext); ok {
			found = true
			break
		}
		parent = parent.GetParent()
	}
	if found {
		text := l.tokens.GetTextFromRuleContext(ctx)
		isLong := strings.Count(text, "\n") > 0 || len(strings.Fields(text)) > 15
		startToken := ctx.GetStart()
		stopToken := ctx.GetStop()
		start := startToken.GetStart()
		stop := stopToken.GetStop() + 1 // exclusive
		// Use token stream to check for surrounding parens
		startIdx := startToken.GetTokenIndex()
		stopIdx := stopToken.GetTokenIndex()
		if startIdx > 0 {
			prev := l.tokens.Get(startIdx - 1)
			if prev.GetTokenType() == parser.SnowflakeParserLR_BRACKET {
				start = prev.GetStart()
			}
		}
		if stopIdx+1 < l.tokens.Size() {
			next := l.tokens.Get(stopIdx + 1)
			if next.GetTokenType() == parser.SnowflakeParserRR_BRACKET {
				stop = next.GetStop() + 1
			}
		}
		l.subqueries = append(l.subqueries, &subqueryInfo{
			text:         text,
			isLong:       isLong,
			isCorrelated: false,
			start:        start,
			stop:         stop,
		})
	}
}
