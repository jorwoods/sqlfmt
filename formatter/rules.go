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
	// Use the token type from the generated lexer
	if clauseTokenTypes[token.GetTokenType()] {
		return true
	}
	return token.GetTokenType() == parser.SnowflakeLexerAS
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
	// If only clause alignment is enabled, just align clauses, do not format select list or uppercase/strip unless those rules are also enabled (handled by token mutation before this call)
	// Always use the mutated token stream for output, so uppercasing and quote stripping are reflected
	var out []string
	var line []string
	var currentIndent string
	clauseIndent := map[int]string{
		parser.SnowflakeLexerSELECT:  "",
		parser.SnowflakeLexerFROM:    "  ",
		parser.SnowflakeLexerWHERE:   " ",
		parser.SnowflakeLexerGROUP:   " ",
		parser.SnowflakeLexerHAVING:  " ",
		parser.SnowflakeLexerORDER:   " ",
		parser.SnowflakeLexerQUALIFY: " ",
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
			if len(line) > 0 {
				joined := joinTokens(line, operatorSpacingEnabled(cfg))
				if currentIndent != "" {
					joined = currentIndent + joined
				}
				out = append(out, joined)
				line = nil
			}
			currentIndent = clauseIndent[ttype]
			line = []string{text}
			continue
		}
		if text == "," && len(line) > 0 {
			line[len(line)-1] += ","
			continue
		}
		line = append(line, text)
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
	// Always use the mutated token stream for output, so uppercasing and quote stripping are reflected
	var out []string
	var line []string
	var currentIndent string
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
			   currentIndent = clauseIndent[ttype]
			   line = []string{text}
			   if ttype == parser.SnowflakeLexerSELECT {
				   inSelect = true
				   lastClauseText = text
			   }
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

func tokensToText(tokens antlr.TokenStream, operatorSpacing bool) string {
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
		if prev != "" {
			if text == "," || text == ")" || text == ";" {
				// no space
			} else if prev == "(" {
				// no space
			} else if !operatorSpacing && (operatorTokenTypes[ttype] || operatorTokenTypes[prevTtype]) {
				// compact mode: no space around operators
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

func normalizeNotEqual(tokens antlr.TokenStream) {
	for i := 0; i < tokens.Size(); i++ {
		tok := tokens.Get(i)
		if tok.GetTokenType() == parser.SnowflakeLexerLTGT {
			tok.(*antlr.CommonToken).SetText("!=")
		}
	}
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
