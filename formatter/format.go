package formatter

import (
	antlr "github.com/antlr4-go/antlr/v4"
	"github.com/jorwoods/sqlfmt/parser"
)

// FormatSQLWithConfig formats SQL using the provided config.
func FormatSQLWithConfig(input string, cfg *Config) string {
	is := antlr.NewInputStream(input)
	lexer := parser.NewSnowflakeLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := parser.NewSnowflakeParser(stream)
	p.BuildParseTrees = true
	p.Snowflake_file()

	if cfg == nil || cfg.Rules.UppercaseKeywords {
		uppercaseKeywords(stream)
	}
	if cfg == nil || cfg.Rules.StripQuotes {
		stripQuotesIfSafe(stream)
	}
	return rightAlignClausesWithConfig(stream, cfg)
}

// FormatSQL formats SQL using default rules (all enabled).
func FormatSQL(input string) string {
	return FormatSQLWithConfig(input, nil)
}

