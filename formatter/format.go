package formatter

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/jorwoods/sqlfmt/parser"
)

func FormatSQL(input string) string {
	is := antlr.NewInputStream(input)
	lexer := parser.NewSnowflakeSqlLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := parser.NewSnowflakeSqlParser(stream)
	p.BuildParseTrees = true
	p.SingleStatement() // parse tree ignored for now

	uppercaseKeywords(stream)
	stripQuotesIfSafe(stream)

	return rightAlignClauses(stream)
}

