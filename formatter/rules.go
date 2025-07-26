package formatter

import (
	"fmt"
	"regexp"
	"strings"

	antlr "github.com/antlr4-go/antlr/v4"
)

var keywordSet = map[string]struct{}{
	"select":  {},
	"from":    {},
	"where":   {},
	"group":   {},
	"having":  {},
	"order":   {},
	"qualify": {},
	"insert":  {},
	"update":  {},
	"delete":  {},
	"create":  {},
	"drop":    {},
	"merge":   {},
	"use":     {},
	"show":    {},
	"describe": {},
}

func isKeyword(token antlr.Token) bool {
	_, ok := keywordSet[strings.ToLower(token.GetText())]
	return ok
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
		text := tok.GetText()
		if identPattern.MatchString(text) {
			unquoted := strings.Trim(text, `"`)
			tok.(*antlr.CommonToken).SetText(unquoted)
		}
	}
}

var alignKeywords = map[string]bool{
	"SELECT": true,
	"FROM":   true,
	"WHERE":  true,
	"GROUP":  true,
	"HAVING": true,
	"ORDER":  true,
	"QUALIFY": true,
}

func rightAlignClauses(tokens antlr.TokenStream) string {
	var b strings.Builder

	for i := 0; i < tokens.Size(); i++ {
		tok := tokens.Get(i)
		text := tok.GetText()
		upper := strings.ToUpper(text)

		// Insert newline before aligning clause keyword
		if alignKeywords[upper] {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			// Add clause right-aligned to column 16 (adjustable)
			padded := fmt.Sprintf("%16s", upper)
			b.WriteString(padded)
		} else {
			// Basic spacing for others
			if b.Len() > 0 && !strings.HasSuffix(b.String(), "\n") {
				b.WriteString(" ")
			}
			b.WriteString(text)
		}
	}

	b.WriteString("\n<EOF>")
	return b.String()
}

func tokensToText(tokens antlr.TokenStream) string {
	var b strings.Builder
	for i := 0; i < tokens.Size(); i++ {
		b.WriteString(tokens.Get(i).GetText())
	}
	return b.String()
}

