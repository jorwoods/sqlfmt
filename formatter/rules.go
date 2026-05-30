package formatter

import (
	"strings"

	antlr "github.com/antlr4-go/antlr/v4"
	"github.com/jorwoods/sqlfmt/parser"
)

var operatorSymbols = map[string]bool{
	"=": true, ">": true, ">=": true, "<": true, "<=": true,
	"!=": true, "+": true, "-": true, "/": true,
	// "*" excluded: doubles as SELECT * wildcard
}

var operatorTokenTypes = map[int]bool{
	parser.SnowflakeLexerEQ:     true,
	parser.SnowflakeLexerGT:     true,
	parser.SnowflakeLexerGE:     true,
	parser.SnowflakeLexerLT:     true,
	parser.SnowflakeLexerLE:     true,
	parser.SnowflakeLexerNE:     true,
	parser.SnowflakeLexerPLUS:   true,
	parser.SnowflakeLexerMINUS:  true,
	parser.SnowflakeLexerDIVIDE: true,
}

var joinQualifierTypes = map[int]bool{
	parser.SnowflakeLexerINNER:   true,
	parser.SnowflakeLexerLEFT:    true,
	parser.SnowflakeLexerRIGHT:   true,
	parser.SnowflakeLexerFULL:    true,
	parser.SnowflakeLexerCROSS:   true,
	parser.SnowflakeLexerNATURAL: true,
}

var wherelikeClauseTypes = map[int]bool{
	parser.SnowflakeLexerWHERE:   true,
	parser.SnowflakeLexerHAVING:  true,
	parser.SnowflakeLexerQUALIFY: true,
}

var booleanOpTokens = map[int]bool{
	parser.SnowflakeLexerAND: true,
	parser.SnowflakeLexerOR:  true,
}

// booleanOpIndent right-aligns AND/OR to the same 6-char column as SELECT.
var booleanOpIndent = map[int]string{
	parser.SnowflakeLexerAND: "   ",  // 3 spaces + "AND" = 6 chars
	parser.SnowflakeLexerOR:  "    ", // 4 spaces + "OR"  = 6 chars
}

func operatorSpacingEnabled(cfg *Config) bool {
	if cfg == nil {
		return true
	}
	return cfg.Rules.OperatorSpacing
}

// Clause keyword token types from the generated lexer
var clauseTokenTypes = map[int]bool{
	parser.SnowflakeLexerSELECT:  true,
	parser.SnowflakeLexerFROM:    true,
	parser.SnowflakeLexerWHERE:   true,
	parser.SnowflakeLexerGROUP:   true,
	parser.SnowflakeLexerHAVING:  true,
	parser.SnowflakeLexerORDER:   true,
	parser.SnowflakeLexerQUALIFY: true,
}

func isKeyword(token antlr.Token) bool {
	if clauseTokenTypes[token.GetTokenType()] {
		return true
	}
	ttype := token.GetTokenType()
	switch ttype {
	case parser.SnowflakeLexerAS,
		parser.SnowflakeLexerAND,
		parser.SnowflakeLexerOR,
		parser.SnowflakeLexerCASE,
		parser.SnowflakeLexerWHEN,
		parser.SnowflakeLexerTHEN,
		parser.SnowflakeLexerELSE,
		parser.SnowflakeLexerEND,
		parser.SnowflakeLexerJOIN,
		parser.SnowflakeLexerINNER,
		parser.SnowflakeLexerLEFT,
		parser.SnowflakeLexerRIGHT,
		parser.SnowflakeLexerFULL,
		parser.SnowflakeLexerCROSS,
		parser.SnowflakeLexerOUTER,
		parser.SnowflakeLexerNATURAL,
		parser.SnowflakeLexerON:
		return true
	}
	return false
}

func uppercaseKeywords(tokens antlr.TokenStream) {
	       for i := 0; i < tokens.Size(); i++ {
		       tok := tokens.Get(i)
		       if isKeyword(tok) {
			       tok.(*antlr.CommonToken).SetText(strings.ToUpper(tok.GetText()))
		       }
	       }
}

func stripQuotesIfSafe(tokens antlr.TokenStream) {
       for i := 0; i < tokens.Size(); i++ {
	       tok := tokens.Get(i)
	       text := tok.GetText()
	       if len(text) > 1 && text[0] == '"' && text[len(text)-1] == '"' {
		       unquoted := text[1 : len(text)-1]
		       tok.(*antlr.CommonToken).SetText(unquoted)
	       }
       }
}

// rightAlignClausesWithConfig aligns both clauses and SELECT lists (for both rules enabled)
func rightAlignClausesWithConfig(tokens antlr.TokenStream, cfg *Config) string {
	return alignClausesAndSelectList(tokens, cfg)
}

// alignClausesOnly aligns only the clauses, leaves select list formatting flat
func alignClausesOnly(tokens antlr.TokenStream, cfg *Config) string {
	var out []string
	var line []string
	var currentIndent string
	currentClause := 0
	newlineAndOr := cfg != nil && cfg.Rules.NewlineBeforeAndOr
	newlineOn := cfg != nil && cfg.Rules.NewlineBeforeOn
	inJoin := false
	joinStarts := map[int]bool{}
	if cfg != nil && cfg.Rules.NewlineBeforeJoin {
		joinStarts = scanJoinStarts(tokens)
	}
	clauseIndent := map[int]string{
		parser.SnowflakeLexerSELECT:  "",
		parser.SnowflakeLexerFROM:    "  ",
		parser.SnowflakeLexerWHERE:   " ",
		parser.SnowflakeLexerGROUP:   " ",
		parser.SnowflakeLexerHAVING:  " ",
		parser.SnowflakeLexerORDER:   " ",
		parser.SnowflakeLexerQUALIFY: " ",
	}
	flushLine := func() {
		if len(line) > 0 {
			joined := joinTokens(line, operatorSpacingEnabled(cfg))
			if currentIndent != "" {
				joined = currentIndent + joined
			}
			out = append(out, joined)
			line = nil
		}
	}
	for i := 0; i < tokens.Size(); i++ {
		tok := tokens.Get(i)
		if tok.GetChannel() != antlr.TokenDefaultChannel {
			continue
		}
		text := tok.GetText()
		ttype := tok.GetTokenType()
		if strings.ToUpper(text) == "<EOF>" {
			break
		}
		if clauseTokenTypes[ttype] {
			flushLine()
			currentClause = ttype
			currentIndent = clauseIndent[ttype]
			inJoin = false
			line = []string{text}
			continue
		}
		if ttype == parser.SnowflakeLexerSEMI {
			flushLine()
			out = append(out, ";")
			currentIndent = ""
			currentClause = 0
			inJoin = false
			continue
		}
		if newlineAndOr && booleanOpTokens[ttype] && wherelikeClauseTypes[currentClause] {
			flushLine()
			currentIndent = booleanOpIndent[ttype]
			line = []string{text}
			continue
		}
		if joinStarts[i] {
			flushLine()
			currentIndent = "  "
			inJoin = true
			line = []string{text}
			continue
		}
		if newlineOn && inJoin && ttype == parser.SnowflakeLexerON {
			flushLine()
			currentIndent = "    "
			line = []string{text}
			continue
		}
		if text == "," && len(line) > 0 {
			line[len(line)-1] += ","
			continue
		}
		line = append(line, text)
	}
	flushLine()
	return strings.Join(out, "\n")
}

// formatSelectListOnly formats only the SELECT list (if >3 items), leaves clauses unaligned
func formatSelectListOnly(tokens antlr.TokenStream, cfg *Config) string {
	// If only select list formatting is enabled, do not align clauses, just format select list if >3 items
	// Always use the mutated token stream for output, so uppercasing and quote stripping are reflected
	var out []string
	var selectIdents []string
	var inSelect bool
	var selectWord string
	var afterSelect []string
	selectIndent := ""
	itemIndent := "       "
	for i := 0; i < tokens.Size(); i++ {
		tok := tokens.Get(i)
		if tok.GetChannel() != antlr.TokenDefaultChannel {
			continue
		}
		text := tok.GetText()
		ttype := tok.GetTokenType()
		if strings.ToUpper(text) == "<EOF>" {
			break
		}
		if ttype == parser.SnowflakeLexerSEMI {
			if inSelect {
				if len(selectIdents) > 3 {
					for j, ident := range selectIdents {
						comma := ","
						if j == len(selectIdents)-1 {
							comma = ""
						}
						if j == 0 {
							out = append(out, selectIndent+joinTokens([]string{selectWord, ident + comma}, operatorSpacingEnabled(cfg)))
						} else {
							out = append(out, itemIndent+ident+comma)
						}
					}
				} else if len(selectIdents) > 0 {
					out = append(out, selectIndent+selectWord+" "+strings.Join(selectIdents, ", "))
				}
				selectIdents = nil
				inSelect = false
			} else if len(afterSelect) > 0 {
				if !(len(afterSelect) == 1 && strings.ToUpper(afterSelect[0]) == "SELECT") {
					out = append(out, joinTokens(afterSelect, operatorSpacingEnabled(cfg)))
				}
				afterSelect = nil
			}
			out = append(out, ";")
			continue
		}
		if ttype == parser.SnowflakeLexerSELECT {
			inSelect = true
			selectWord = text
			continue
		}
		if inSelect {
			if text == "," {
				continue
			}
			if clauseTokenTypes[ttype] && ttype != parser.SnowflakeLexerSELECT {
				// End of select list
				if len(selectIdents) > 3 {
					for j, ident := range selectIdents {
						comma := ","
						if j == len(selectIdents)-1 {
							comma = ""
						}
						if j == 0 {
							out = append(out, selectIndent+joinTokens([]string{selectWord, ident + comma}, operatorSpacingEnabled(cfg)))
						} else {
							out = append(out, itemIndent+ident+comma)
						}
					}
				   } else if len(selectIdents) > 0 {
					   // Join identifiers with commas for single-line SELECT
					   out = append(out, selectIndent+selectWord+" "+strings.Join(selectIdents, ", "))
				   }
				selectIdents = nil
				inSelect = false
				afterSelect = append(afterSelect, text)
				continue
			}
			if !clauseTokenTypes[ttype] || ttype == parser.SnowflakeLexerSELECT {
				selectIdents = append(selectIdents, text)
			}
			continue
		}
		if !inSelect {
			afterSelect = append(afterSelect, text)
		}
	}
	if inSelect && len(selectIdents) > 0 {
		if len(selectIdents) > 3 {
			for j, ident := range selectIdents {
				comma := ","
				if j == len(selectIdents)-1 {
					comma = ""
				}
				if j == 0 {
					out = append(out, selectIndent+joinTokens([]string{selectWord, ident + comma}, operatorSpacingEnabled(cfg)))
				} else {
					out = append(out, itemIndent+ident+comma)
				}
			}
		   } else {
			   // Join identifiers with commas for single-line SELECT
			   out = append(out, selectIndent+selectWord+" "+strings.Join(selectIdents, ", "))
		   }
	}
	   if len(afterSelect) > 0 {
		   // Only add the clause line if it is not SELECT (to avoid duplicate SELECT)
		   if !(len(afterSelect) == 1 && strings.ToUpper(afterSelect[0]) == "SELECT") {
			   out = append(out, joinTokens(afterSelect, operatorSpacingEnabled(cfg)))
		   }
	   }
	return strings.Join(out, "\n")
}

// alignClausesAndSelectList applies both clause alignment and select list formatting
func alignClausesAndSelectList(tokens antlr.TokenStream, cfg *Config) string {
	var out []string
	var line []string
	var currentIndent string
	currentClause := 0
	newlineAndOr := cfg != nil && cfg.Rules.NewlineBeforeAndOr
	newlineOn := cfg != nil && cfg.Rules.NewlineBeforeOn
	inJoin := false
	joinStarts := map[int]bool{}
	if cfg != nil && cfg.Rules.NewlineBeforeJoin {
		joinStarts = scanJoinStarts(tokens)
	}
	clauseIndent := map[int]string{
		parser.SnowflakeLexerSELECT:  "",
		parser.SnowflakeLexerFROM:    "  ",
		parser.SnowflakeLexerWHERE:   " ",
		parser.SnowflakeLexerGROUP:   " ",
		parser.SnowflakeLexerHAVING:  " ",
		parser.SnowflakeLexerORDER:   " ",
		parser.SnowflakeLexerQUALIFY: " ",
	}
	var inSelect bool
	var selectIdents []string
	var lastClauseText string
	selectIndent := ""
	itemIndent := "       "
	for i := 0; i < tokens.Size(); i++ {
		tok := tokens.Get(i)
		if tok.GetChannel() != antlr.TokenDefaultChannel {
			continue
		}
		text := tok.GetText()
		ttype := tok.GetTokenType()
		if strings.ToUpper(text) == "<EOF>" {
			break
		}
		if ttype == parser.SnowflakeLexerSEMI {
			if inSelect && len(selectIdents) > 0 {
				selectWord := lastClauseText
				if selectWord == "" {
					selectWord = "SELECT"
				}
				if len(selectIdents) > 3 {
					for j, ident := range selectIdents {
						comma := ","
						if j == len(selectIdents)-1 {
							comma = ""
						}
						if j == 0 {
							out = append(out, selectIndent+joinTokens([]string{selectWord, ident + comma}, operatorSpacingEnabled(cfg)))
						} else {
							out = append(out, itemIndent+ident+comma)
						}
					}
				} else {
					out = append(out, selectIndent+selectWord+" "+strings.Join(selectIdents, ", "))
				}
				selectIdents = nil
				inSelect = false
			}
			if len(line) > 0 {
				if !(currentIndent == clauseIndent[parser.SnowflakeLexerSELECT] && len(line) == 1 && strings.ToUpper(line[0]) == "SELECT") {
					joined := joinTokens(line, operatorSpacingEnabled(cfg))
					if currentIndent != "" {
						joined = currentIndent + joined
					}
					out = append(out, joined)
				}
				line = nil
			}
			out = append(out, ";")
			currentIndent = ""
			currentClause = 0
			inJoin = false
			continue
		}
		if clauseTokenTypes[ttype] {
			if inSelect && len(selectIdents) > 0 {
				selectWord := lastClauseText
				if selectWord == "" {
					selectWord = "SELECT"
				}
				if len(selectIdents) > 3 {
					for j, ident := range selectIdents {
						comma := ","
						if j == len(selectIdents)-1 {
							comma = ""
						}
						if j == 0 {
							out = append(out, selectIndent+joinTokens([]string{selectWord, ident + comma}, operatorSpacingEnabled(cfg)))
						} else {
							out = append(out, itemIndent+ident+comma)
						}
					}
				   } else if len(selectIdents) > 0 {
					   // Join identifiers with commas for single-line SELECT
					   out = append(out, selectIndent+selectWord+" "+strings.Join(selectIdents, ", "))
				   }
				selectIdents = nil
				inSelect = false
			}
			   if len(line) > 0 {
				   // Only add the clause line if it is not SELECT (to avoid duplicate SELECT)
				   if !(currentIndent == clauseIndent[parser.SnowflakeLexerSELECT] && len(line) == 1 && strings.ToUpper(line[0]) == "SELECT") {
					   joined := joinTokens(line, operatorSpacingEnabled(cfg))
					   if currentIndent != "" {
						   joined = currentIndent + joined
					   }
					   out = append(out, joined)
				   }
				   line = nil
			   }
			   currentClause = ttype
			   currentIndent = clauseIndent[ttype]
			   inJoin = false
			   line = []string{text}
			   if ttype == parser.SnowflakeLexerSELECT {
				   inSelect = true
				   lastClauseText = text
			   }
			   continue
		}
		if newlineAndOr && booleanOpTokens[ttype] && wherelikeClauseTypes[currentClause] {
			if len(line) > 0 {
				joined := joinTokens(line, operatorSpacingEnabled(cfg))
				if currentIndent != "" {
					joined = currentIndent + joined
				}
				out = append(out, joined)
				line = nil
			}
			currentIndent = booleanOpIndent[ttype]
			line = []string{text}
			continue
		}
		if joinStarts[i] {
			if len(line) > 0 {
				joined := joinTokens(line, operatorSpacingEnabled(cfg))
				if currentIndent != "" {
					joined = currentIndent + joined
				}
				out = append(out, joined)
				line = nil
			}
			currentIndent = "  "
			inJoin = true
			line = []string{text}
			continue
		}
		if newlineOn && inJoin && ttype == parser.SnowflakeLexerON {
			if len(line) > 0 {
				joined := joinTokens(line, operatorSpacingEnabled(cfg))
				if currentIndent != "" {
					joined = currentIndent + joined
				}
				out = append(out, joined)
				line = nil
			}
			currentIndent = "    "
			line = []string{text}
			continue
		}
		if inSelect {
			if text == "," {
				continue
			}
			if clauseTokenTypes[tok.GetTokenType()] && tok.GetTokenType() != parser.SnowflakeLexerSELECT {
				inSelect = false
			}
			if !clauseTokenTypes[tok.GetTokenType()] || tok.GetTokenType() == parser.SnowflakeLexerSELECT {
				selectIdents = append(selectIdents, text)
			}
			continue
		}
		if text == "," && len(line) > 0 {
			line[len(line)-1] += ","
			continue
		}
		line = append(line, text)
	}
	if inSelect && len(selectIdents) > 0 {
		selectWord := lastClauseText
		if selectWord == "" {
			selectWord = "SELECT"
		}
		if len(selectIdents) > 3 {
			for j, ident := range selectIdents {
				comma := ","
				if j == len(selectIdents)-1 {
					comma = ""
				}
				if j == 0 {
					out = append(out, selectIndent+joinTokens([]string{selectWord, ident + comma}, operatorSpacingEnabled(cfg)))
				} else {
					out = append(out, itemIndent+ident+comma)
				}
			}
		   } else if len(selectIdents) > 0 {
			   // Join identifiers with commas for single-line SELECT
			   out = append(out, selectIndent+selectWord+" "+strings.Join(selectIdents, ", "))
		   }
		selectIdents = nil
		inSelect = false
	}
	if len(line) > 0 {
		joined := joinTokens(line, operatorSpacingEnabled(cfg))
		if currentIndent != "" {
			joined = currentIndent + joined
		}
		out = append(out, joined)
	}
	return strings.Join(out, "\n")
}

// joinTokens joins tokens with a single space, but avoids spaces before commas and after opening parens.
// When operatorSpacing is false, spaces around arithmetic and comparison operators are also suppressed.
func joinTokens(tokens []string, operatorSpacing bool) string {
	var out strings.Builder
	prev := ""
	for _, text := range tokens {
		if prev != "" {
			if text == "," || text == ")" || text == ";" {
				// no space
			} else if prev == "(" {
				// no space
			} else if text == "(" && isFunctionCallParen(prev) {
				// no space before ( after function name
			} else if !operatorSpacing && (operatorSymbols[text] || operatorSymbols[prev]) {
				// compact mode: no space around operators
			} else {
				out.WriteString(" ")
			}
		}
		out.WriteString(text)
		prev = text
	}
	return out.String()
}
// rightAlignClauses uses all rules enabled (for legacy calls)
func rightAlignClauses(tokens antlr.TokenStream) string {
	return rightAlignClausesWithConfig(tokens, nil)
}

func tokensToText(tokens antlr.TokenStream, cfg *Config) string {
	operatorSpacing := operatorSpacingEnabled(cfg)
	newlineAndOr := cfg != nil && cfg.Rules.NewlineBeforeAndOr
	newlineOn := cfg != nil && cfg.Rules.NewlineBeforeOn
	joinStarts := map[int]bool{}
	if cfg != nil && cfg.Rules.NewlineBeforeJoin {
		joinStarts = scanJoinStarts(tokens)
	}
	currentClause := 0
	inJoin := false
	var out strings.Builder
	prev := ""
	prevTtype := -1
	for i := 0; i < tokens.Size(); i++ {
		tok := tokens.Get(i)
		text := strings.TrimSpace(tok.GetText())
		ttype := tok.GetTokenType()
		if strings.ToUpper(text) == "<EOF>" || text == "" {
			continue
		}
		if clauseTokenTypes[ttype] {
			currentClause = ttype
			inJoin = false
		}
		// Semicolons always go on their own line.
		if text == ";" {
			if prev != "" {
				out.WriteString("\n")
			}
			out.WriteString(";")
			if hasMoreNonEOFTokens(tokens, i) {
				out.WriteString("\n")
			}
			prev = "\n"
			prevTtype = ttype
			currentClause = 0
			inJoin = false
			continue
		}
		// JOIN clause on its own line.
		if joinStarts[i] {
			if prev != "" && prev != "\n" {
				out.WriteString("\n")
			}
			prev = "\n"
			inJoin = true
		}
		// ON on its own line inside a JOIN clause.
		if newlineOn && inJoin && ttype == parser.SnowflakeLexerON {
			if prev != "" && prev != "\n" {
				out.WriteString("\n")
			}
			out.WriteString("    " + text) // 4 spaces + "ON" = 6 chars, aligning with AND/OR/WHERE
			prev = text
			prevTtype = ttype
			continue
		}
		// AND/OR on their own line inside WHERE/HAVING/QUALIFY.
		if newlineAndOr && booleanOpTokens[ttype] && wherelikeClauseTypes[currentClause] {
			if prev != "" {
				out.WriteString("\n")
			}
			out.WriteString(booleanOpIndent[ttype] + text)
			prev = text
			prevTtype = ttype
			continue
		}
		if prev != "" {
			if text == "," || text == ")" {
				// no space
			} else if prev == "(" || prev == "\n" {
				// no space
			} else if text == "(" && isFunctionCallParen(prev) {
				// no space before ( after function name
			} else if !operatorSpacing && (operatorTokenTypes[ttype] || operatorTokenTypes[prevTtype]) {
				// compact mode
			} else {
				out.WriteString(" ")
			}
		}
		out.WriteString(text)
		prev = text
		prevTtype = ttype
	}
	return out.String()
}

func isPunctuation(s string) bool {
       return s == "," || s == "." || s == "(" || s == ")" || s == ";"
}

func isWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

// spaceBeforeParenKeywords lists SQL keywords that precede ( but are not function calls.
var spaceBeforeParenKeywords = func() map[string]bool {
	words := []string{"over", "in", "from", "select", "not", "on", "join"}
	m := make(map[string]bool, len(words)*2)
	for _, w := range words {
		m[w] = true
		m[strings.ToUpper(w)] = true
	}
	return m
}()

func isFunctionCallParen(prev string) bool {
	return len(prev) > 0 && isWordChar(prev[len(prev)-1]) && !spaceBeforeParenKeywords[prev]
}

func normalizeNotEqual(tokens antlr.TokenStream) {
	for i := 0; i < tokens.Size(); i++ {
		tok := tokens.Get(i)
		if tok.GetTokenType() == parser.SnowflakeLexerLTGT {
			tok.(*antlr.CommonToken).SetText("!=")
		}
	}
}

var booleanLiteralTypes = map[int]bool{
	parser.SnowflakeLexerTRUE:  true,
	parser.SnowflakeLexerFALSE: true,
	parser.SnowflakeLexerNULL_: true,
}

func normalizeBooleans(tokens antlr.TokenStream) {
	for i := 0; i < tokens.Size(); i++ {
		tok := tokens.Get(i)
		if booleanLiteralTypes[tok.GetTokenType()] {
			tok.(*antlr.CommonToken).SetText(strings.ToUpper(tok.GetText()))
		}
	}
}

// scanJoinStarts returns a set of stream indices where a JOIN clause begins.
// For qualified joins (LEFT JOIN, INNER JOIN, etc.) the index points to the
// qualifier; for bare JOIN it points to the JOIN token itself.
func scanJoinStarts(tokens antlr.TokenStream) map[int]bool {
	starts := map[int]bool{}
	for i := 0; i < tokens.Size(); i++ {
		tok := tokens.Get(i)
		if tok.GetChannel() != antlr.TokenDefaultChannel {
			continue
		}
		if tok.GetTokenType() != parser.SnowflakeLexerJOIN {
			continue
		}
		startIdx := i
		for j := i - 1; j >= 0; j-- {
			prev := tokens.Get(j)
			if prev.GetChannel() != antlr.TokenDefaultChannel {
				continue
			}
			pt := prev.GetTokenType()
			if pt == parser.SnowflakeLexerOUTER || joinQualifierTypes[pt] {
				startIdx = j
			} else {
				break
			}
		}
		starts[startIdx] = true
	}
	return starts
}

func uppercaseFunctions(tokens antlr.TokenStream) {
	for i := 0; i < tokens.Size()-1; i++ {
		tok := tokens.Get(i)
		if tok.GetChannel() != antlr.TokenDefaultChannel {
			continue
		}
		for j := i + 1; j < tokens.Size(); j++ {
			next := tokens.Get(j)
			if next.GetChannel() != antlr.TokenDefaultChannel {
				continue
			}
			if next.GetTokenType() == parser.SnowflakeLexerLR_BRACKET {
				tok.(*antlr.CommonToken).SetText(strings.ToUpper(tok.GetText()))
			}
			break
		}
	}
}

// blankLinesBetweenStatements inserts a blank line after lines that are exactly ";".
// All renderers guarantee semicolons appear on their own line, so an exact match is safe
// and avoids false positives from semicolons inside string literals.
func blankLinesBetweenStatements(sql string) string {
	lines := strings.Split(sql, "\n")
	var out []string
	for i, line := range lines {
		out = append(out, line)
		if strings.TrimSpace(line) == ";" && i < len(lines)-1 && strings.TrimSpace(lines[i+1]) != "" {
			out = append(out, "")
		}
	}
	return strings.Join(out, "\n")
}

// formatCaseExpressions reformats CASE expressions so that WHEN, ELSE, and END
// each start on their own line, indented 4 spaces past the column of the CASE keyword.
// END aligns with CASE. Nested CASEs are handled via a stack.
// The function operates on an already-rendered SQL string and respects string literals
// and quoted identifiers to avoid misidentifying keywords inside them.
func formatCaseExpressions(sql string) string {
	type frame struct{ caseCol int }
	var stack []frame

	var out strings.Builder
	col := 0      // columns written since last newline
	pending := "" // whitespace buffered between words

	flushPending := func() {
		if pending == "" {
			return
		}
		out.WriteString(pending)
		if idx := strings.LastIndex(pending, "\n"); idx >= 0 {
			col = len(pending) - idx - 1
		} else {
			col += len(pending)
		}
		pending = ""
	}

	writeDirect := func(s string) {
		out.WriteString(s)
		if idx := strings.LastIndex(s, "\n"); idx >= 0 {
			col = len(s) - idx - 1
		} else {
			col += len(s)
		}
	}

	i := 0
	for i < len(sql) {
		ch := sql[i]

		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			j := i
			for j < len(sql) && (sql[j] == ' ' || sql[j] == '\t' || sql[j] == '\n' || sql[j] == '\r') {
				j++
			}
			pending += sql[i:j]
			i = j
			continue
		}

		// Single-quoted string literal: copy verbatim, handle '' escapes.
		if ch == '\'' {
			flushPending()
			j := i + 1
			for j < len(sql) {
				if sql[j] == '\'' {
					j++
					if j < len(sql) && sql[j] == '\'' {
						j++ // escaped ''
						continue
					}
					break
				}
				j++
			}
			writeDirect(sql[i:j])
			i = j
			continue
		}

		// Double-quoted identifier: copy verbatim.
		if ch == '"' {
			flushPending()
			j := i + 1
			for j < len(sql) && sql[j] != '"' {
				j++
			}
			if j < len(sql) {
				j++
			}
			writeDirect(sql[i:j])
			i = j
			continue
		}

		// Word token (keyword or identifier).
		if isWordChar(ch) {
			j := i
			for j < len(sql) && isWordChar(sql[j]) {
				j++
			}
			word := sql[i:j]
			upper := strings.ToUpper(word)

			switch upper {
			case "CASE":
				flushPending()
				caseStartCol := col
				writeDirect(word)
				stack = append(stack, frame{caseCol: caseStartCol})
			case "WHEN":
				if len(stack) > 0 {
					f := stack[len(stack)-1]
					pending = ""
					writeDirect("\n" + strings.Repeat(" ", f.caseCol+4))
				} else {
					flushPending()
				}
				writeDirect(word)
			case "ELSE":
				if len(stack) > 0 {
					f := stack[len(stack)-1]
					pending = ""
					writeDirect("\n" + strings.Repeat(" ", f.caseCol+4))
				} else {
					flushPending()
				}
				writeDirect(word)
			case "END":
				if len(stack) > 0 {
					f := stack[len(stack)-1]
					pending = ""
					writeDirect("\n" + strings.Repeat(" ", f.caseCol))
					stack = stack[:len(stack)-1]
				} else {
					flushPending()
				}
				writeDirect(word)
			default:
				flushPending()
				writeDirect(word)
			}
			i = j
			continue
		}

		// Operators, punctuation, numbers.
		flushPending()
		writeDirect(string(ch))
		i++
	}
	flushPending()
	return out.String()
}

func hasMoreNonEOFTokens(tokens antlr.TokenStream, from int) bool {
	for j := from + 1; j < tokens.Size(); j++ {
		text := strings.TrimSpace(tokens.Get(j).GetText())
		if text != "" && !strings.EqualFold(text, "<EOF>") {
			return true
		}
	}
	return false
}

// applyLeadingCommas moves trailing commas to the start of the following line.
// It only fires when a comma ends a line and the next line is indented by at least 2
// spaces (i.e. it is a continuation item, not a new clause). The comma is placed 2
// columns before the item so that column names stay aligned with the first item.
func applyLeadingCommas(sql string) string {
	lines := strings.Split(sql, "\n")
	for i := 0; i < len(lines)-1; i++ {
		trimmed := strings.TrimRight(lines[i], " \t")
		if !strings.HasSuffix(trimmed, ",") {
			continue
		}
		nextLine := lines[i+1]
		nextIndent := 0
		for nextIndent < len(nextLine) && nextLine[nextIndent] == ' ' {
			nextIndent++
		}
		if nextIndent < 2 {
			continue
		}
		lines[i] = trimmed[:len(trimmed)-1]
		lines[i+1] = strings.Repeat(" ", nextIndent-2) + ", " + nextLine[nextIndent:]
	}
	return strings.Join(lines, "\n")
}

func stripTrailingWhitespace(sql string) string {
	lines := strings.Split(sql, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

func ensureTrailingSemicolon(sql string) string {
	trimmed := strings.TrimRight(sql, " \t\n\r")
	trimmed = strings.TrimRight(trimmed, ";")
	trimmed = strings.TrimRight(trimmed, " \t\n\r")
	return trimmed + "\n;"
}
