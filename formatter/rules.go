package formatter

import (
	"regexp"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/jorwoods/sqlfmt/parser"
)

func isKeyword(token antlr.Token) bool {
	tType := token.GetTokenType()
	return tType >= parser.SnowflakeParserK_SELECT &&
		tType <= parser.SnowflakeParserK_QUALIFY
}

func uppercaseKeywords(tokens antlr.TokenStream) {
	for i := 0; i < tokens.Size(); i++ {
		tok := tokens.Get(i)
		if isKeyword(tok) {
			tok.(*antlr.CommonToken).SetText(strings.ToUpper(tok.GetText()))
		}
	}
}

var identPattern = regexp.MustCompile(`^"[A-Z]+"$`)

func stripQuotesIfSafe(tokens antlr.TokenStream) {
	for i := 0; i < tokens.Size(); i++ {
		tok := tokens.Get(i)
		if tok.GetTokenType() == parser.SnowflakeParserIDENTIFIER {
			text := tok.GetText()
			if identPattern.MatchString(text) {
				tok.(*antlr.CommonToken).SetText(strings.Trim(text, `"`))
			}
		}
	}
}

var alignKeywords = map[string]struct{}{
	"SELECT": {}, "FROM": {}, "WHERE": {}, "GROUP": {}, "HAVING": {}, "ORDER": {}, "QUALIFY": {},
}

func rightAlignClauses(tokens antlr.TokenStream) string {
	lines := strings.Split(tokensToText(tokens), "\n")

	max := 0
	for _, line := range lines {
		for k := range alignKeywords {
			if idx := strings.Index(line, k); idx > max {
				max = idx
			}
		}
	}

	var b strings.Builder
	for _, line := range lines {
		for k := range alignKeywords {
			if idx := strings.Index(line, k); idx != -1 {
				pad := strings.Repeat(" ", max-idx)
				line = strings.Replace(line, k, pad+k, 1)
				break
			}
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func tokensToText(tokens antlr.TokenStream) string {
	var b strings.Builder
	for i := 0; i < tokens.Size(); i++ {
		b.WriteString(tokens.Get(i).GetText())
	}
	return b.String()
}

