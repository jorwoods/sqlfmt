package formatter

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/jorwoods/sqlfmt/parser"
)

func FormatSQL(input string) string {
	is := antlr.NewInputStream(input)
	lexer := parser.NewSnowflakeLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := parser.NewSnowflakeParser(stream)
	p.BuildParseTrees = true

	p.Snowflake_file()

	uppercaseKeywords(stream)
	stripQuotesIfSafe(stream)

	return rightAlignClauses(stream)
}

