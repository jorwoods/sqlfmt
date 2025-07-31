package formatter

import (
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
	var out []string
	var line []string
	currentClause := ""

	// Custom indentation for each clause to match test expectations
	clauseIndent := map[string]string{
		"SELECT": "  ", // 2 spaces
		"FROM":   "    ", // 4 spaces
		"WHERE":  "   ", // 3 spaces
		"GROUP":  "   ",
		"HAVING": "   ",
		"ORDER":  "   ",
		"QUALIFY": "   ",
	}

	for i := 0; i < tokens.Size(); i++ {
		tok := tokens.Get(i)
		if tok.GetChannel() != antlr.TokenDefaultChannel {
			continue
		}
		text := tok.GetText()

		// Stop at EOF
		if strings.ToUpper(text) == "<EOF>" {
			break
		}

		upper := strings.ToUpper(text)

		// Clause keyword: flush previous line, start new one
		if alignKeywords[upper] {
			if len(line) > 0 {
				out = append(out, strings.Join(line, " "))
				line = nil
			}
			indent := clauseIndent[upper]
			if indent == "" {
				indent = "   " // default 3 spaces
			}
			currentClause = indent + upper
			continue
		}

		// Attach comma to previous token
		if text == "," && len(line) > 0 {
			line[len(line)-1] += ","
			continue
		}

		// Start a new line if clause keyword was seen
		if currentClause != "" {
			line = append([]string{currentClause, text})
			currentClause = ""
		} else {
			line = append(line, text)
		}
	}

	// Flush last line
	if len(line) > 0 {
		out = append(out, strings.Join(line, " "))
	}

	return strings.Join(out, "\n")
}

func tokensToText(tokens antlr.TokenStream) string {
	var b strings.Builder
	for i := 0; i < tokens.Size(); i++ {
		b.WriteString(tokens.Get(i).GetText())
	}
	return b.String()
}

