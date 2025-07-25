package formatter

import (
	antlr "github.com/antlr4-go/antlr/v4"
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

