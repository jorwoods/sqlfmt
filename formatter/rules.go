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
	parser.SnowflakeLexerSELECT:    true,
	parser.SnowflakeLexerFROM:      true,
	parser.SnowflakeLexerWHERE:     true,
	parser.SnowflakeLexerGROUP:     true,
	parser.SnowflakeLexerHAVING:    true,
	parser.SnowflakeLexerORDER:     true,
	parser.SnowflakeLexerQUALIFY:   true,
	parser.SnowflakeLexerUNION:     true,
	parser.SnowflakeLexerINTERSECT: true,
	parser.SnowflakeLexerEXCEPT:    true,
}

// setOperationTypes is the subset of clauseTokenTypes that are set operators.
// Used to force a newline before them in the flat rendering path.
var setOperationTypes = map[int]bool{
	parser.SnowflakeLexerUNION:     true,
	parser.SnowflakeLexerINTERSECT: true,
	parser.SnowflakeLexerEXCEPT:    true,
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
		parser.SnowflakeLexerIN,
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
		parser.SnowflakeLexerON,
		parser.SnowflakeLexerIS,
		parser.SnowflakeLexerNOT,
		parser.SnowflakeLexerBY,
		parser.SnowflakeLexerLIMIT,
		parser.SnowflakeLexerOFFSET,
		parser.SnowflakeLexerASC,
		parser.SnowflakeLexerDESC,
		parser.SnowflakeLexerDISTINCT,
		parser.SnowflakeLexerUNION,
		parser.SnowflakeLexerINTERSECT,
		parser.SnowflakeLexerEXCEPT,
		parser.SnowflakeLexerALL,
		parser.SnowflakeLexerWITH:
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
	newlineLimit := cfg != nil && cfg.Rules.NewlineBeforeLimit
	inJoin := false
	joinStarts := map[int]bool{}
	if cfg != nil && cfg.Rules.NewlineBeforeJoin {
		joinStarts = scanJoinStarts(tokens)
	}
	statementParens := scanStatementParens(tokens)
	var scopeStack []bool
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
		if strings.TrimSpace(text) == "" {
			continue
		}
		if ttype == parser.SnowflakeLexerLR_BRACKET {
			scopeStack = append(scopeStack, statementParens[i])
		} else if ttype == parser.SnowflakeLexerRR_BRACKET {
			if len(scopeStack) > 0 {
				scopeStack = scopeStack[:len(scopeStack)-1]
			}
		}
		atStatementLevel := len(scopeStack) == 0 || scopeStack[len(scopeStack)-1]
		if clauseTokenTypes[ttype] && atStatementLevel {
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
		if newlineAndOr && booleanOpTokens[ttype] && wherelikeClauseTypes[currentClause] && atStatementLevel {
			flushLine()
			currentIndent = booleanOpIndent[ttype]
			line = []string{text}
			continue
		}
		if joinStarts[i] && atStatementLevel {
			flushLine()
			currentIndent = "  "
			inJoin = true
			line = []string{text}
			continue
		}
		if newlineOn && inJoin && ttype == parser.SnowflakeLexerON && atStatementLevel {
			flushLine()
			currentIndent = "    "
			line = []string{text}
			continue
		}
		if newlineLimit && (ttype == parser.SnowflakeLexerLIMIT || ttype == parser.SnowflakeLexerOFFSET) && atStatementLevel {
			flushLine()
			currentIndent = " "
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
	var itemTokens []string
	var itemParenDepth int
	var inSelect bool
	var selectWord string
	var afterSelect []string
	selectIndent := ""
	itemIndent := "       "
	flushItem := func() {
		if len(itemTokens) > 0 {
			selectIdents = append(selectIdents, joinTokens(itemTokens, operatorSpacingEnabled(cfg)))
			itemTokens = nil
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
		if strings.TrimSpace(text) == "" {
			continue
		}
		if ttype == parser.SnowflakeLexerSEMI {
			flushItem()
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
			if ttype == parser.SnowflakeLexerLR_BRACKET {
				itemParenDepth++
				itemTokens = append(itemTokens, text)
				continue
			}
			if ttype == parser.SnowflakeLexerRR_BRACKET {
				itemParenDepth--
				itemTokens = append(itemTokens, text)
				continue
			}
			if text == "," && itemParenDepth == 0 {
				flushItem()
				continue
			}
			if clauseTokenTypes[ttype] && ttype != parser.SnowflakeLexerSELECT && itemParenDepth == 0 {
				// End of select list
				flushItem()
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
			itemTokens = append(itemTokens, text)
			continue
		}
		if !inSelect {
			afterSelect = append(afterSelect, text)
		}
	}
	flushItem()
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
	newlineLimit := cfg != nil && cfg.Rules.NewlineBeforeLimit
	inJoin := false
	joinStarts := map[int]bool{}
	if cfg != nil && cfg.Rules.NewlineBeforeJoin {
		joinStarts = scanJoinStarts(tokens)
	}
	statementParens := scanStatementParens(tokens)
	var scopeStack []bool
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
	var itemTokens []string
	var itemParenDepth int
	var lastClauseText string
	selectIndent := ""
	itemIndent := "       "
	flushItem := func() {
		if len(itemTokens) > 0 {
			selectIdents = append(selectIdents, joinTokens(itemTokens, operatorSpacingEnabled(cfg)))
			itemTokens = nil
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
		if strings.TrimSpace(text) == "" {
			continue
		}
		if ttype == parser.SnowflakeLexerLR_BRACKET {
			scopeStack = append(scopeStack, statementParens[i])
		} else if ttype == parser.SnowflakeLexerRR_BRACKET {
			if len(scopeStack) > 0 {
				scopeStack = scopeStack[:len(scopeStack)-1]
			}
		}
		atStatementLevel := len(scopeStack) == 0 || scopeStack[len(scopeStack)-1]
		if ttype == parser.SnowflakeLexerSEMI {
			flushItem()
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
		if clauseTokenTypes[ttype] && atStatementLevel {
			flushItem()
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
		if newlineAndOr && booleanOpTokens[ttype] && wherelikeClauseTypes[currentClause] && atStatementLevel {
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
		if joinStarts[i] && atStatementLevel {
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
		if newlineOn && inJoin && ttype == parser.SnowflakeLexerON && atStatementLevel {
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
		if newlineLimit && (ttype == parser.SnowflakeLexerLIMIT || ttype == parser.SnowflakeLexerOFFSET) && atStatementLevel {
			if len(line) > 0 {
				joined := joinTokens(line, operatorSpacingEnabled(cfg))
				if currentIndent != "" {
					joined = currentIndent + joined
				}
				out = append(out, joined)
				line = nil
			}
			currentIndent = " "
			inJoin = false
			line = []string{text}
			continue
		}
		if inSelect {
			if ttype == parser.SnowflakeLexerLR_BRACKET {
				itemParenDepth++
				itemTokens = append(itemTokens, text)
				continue
			}
			if ttype == parser.SnowflakeLexerRR_BRACKET {
				itemParenDepth--
				itemTokens = append(itemTokens, text)
				continue
			}
			if text == "," && itemParenDepth == 0 {
				flushItem()
				continue
			}
			itemTokens = append(itemTokens, text)
			continue
		}
		if text == "," && len(line) > 0 {
			line[len(line)-1] += ","
			continue
		}
		line = append(line, text)
	}
	flushItem()
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
			if text == "," || text == ";" || strings.HasPrefix(text, ")") {
				// no space before ) or ),  or );
			} else if prev == "(" {
				// no space
			} else if text == "." || prev == "." {
				// no space around qualifier dot, e.g. a.id
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
	newlineLimit := cfg != nil && cfg.Rules.NewlineBeforeLimit
	newlineSetOp := cfg != nil && cfg.Rules.NewlineBeforeSetOp
	newlineGroupBy := cfg != nil && cfg.Rules.NewlineBeforeGroupBy
	newlineOrderBy := cfg != nil && cfg.Rules.NewlineBeforeOrderBy
	newlineHaving := cfg != nil && cfg.Rules.NewlineBeforeHaving
	joinStarts := map[int]bool{}
	if cfg != nil && cfg.Rules.NewlineBeforeJoin {
		joinStarts = scanJoinStarts(tokens)
	}
	statementParens := scanStatementParens(tokens)
	var scopeStack []bool
	currentClause := 0
	inJoin := false
	afterSetOp := false
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
		if ttype == parser.SnowflakeLexerLR_BRACKET {
			scopeStack = append(scopeStack, statementParens[i])
		} else if ttype == parser.SnowflakeLexerRR_BRACKET {
			if len(scopeStack) > 0 {
				scopeStack = scopeStack[:len(scopeStack)-1]
			}
		}
		atStatementLevel := len(scopeStack) == 0 || scopeStack[len(scopeStack)-1]
		if clauseTokenTypes[ttype] && atStatementLevel {
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
			afterSetOp = false
			continue
		}
		// Set operations (UNION / INTERSECT / EXCEPT) on their own line when enabled.
		if newlineSetOp && setOperationTypes[ttype] && atStatementLevel {
			if prev != "" && prev != "\n" {
				out.WriteString("\n")
			}
			prev = "\n"
			afterSetOp = true
		}
		// The SELECT that follows a set-op also starts on its own line.
		if newlineSetOp && afterSetOp && ttype == parser.SnowflakeLexerSELECT && atStatementLevel {
			if prev != "" && prev != "\n" {
				out.WriteString("\n")
			}
			prev = "\n"
			afterSetOp = false
		}
		// JOIN clause on its own line.
		if joinStarts[i] && atStatementLevel {
			if prev != "" && prev != "\n" {
				out.WriteString("\n")
			}
			prev = "\n"
			inJoin = true
		}
		// LIMIT / OFFSET on their own line.
		if newlineLimit && (ttype == parser.SnowflakeLexerLIMIT || ttype == parser.SnowflakeLexerOFFSET) && atStatementLevel {
			if prev != "" && prev != "\n" {
				out.WriteString("\n")
			}
			prev = "\n"
		}
		// GROUP BY, ORDER BY, HAVING on their own lines when respective rules are enabled.
		if newlineGroupBy && ttype == parser.SnowflakeLexerGROUP && atStatementLevel {
			if prev != "" && prev != "\n" {
				out.WriteString("\n")
			}
			prev = "\n"
		}
		if newlineOrderBy && ttype == parser.SnowflakeLexerORDER && atStatementLevel {
			if prev != "" && prev != "\n" {
				out.WriteString("\n")
			}
			prev = "\n"
		}
		if newlineHaving && ttype == parser.SnowflakeLexerHAVING && atStatementLevel {
			if prev != "" && prev != "\n" {
				out.WriteString("\n")
			}
			prev = "\n"
		}
		// ON on its own line inside a JOIN clause.
		if newlineOn && inJoin && ttype == parser.SnowflakeLexerON && atStatementLevel {
			if prev != "" && prev != "\n" {
				out.WriteString("\n")
			}
			out.WriteString("    " + text) // 4 spaces + "ON" = 6 chars, aligning with AND/OR/WHERE
			prev = text
			prevTtype = ttype
			continue
		}
		// AND/OR on their own line inside WHERE/HAVING/QUALIFY.
		if newlineAndOr && booleanOpTokens[ttype] && wherelikeClauseTypes[currentClause] && atStatementLevel {
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
			} else if text == "." || prev == "." {
				// no space around qualifier dot, e.g. a.id
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
	words := []string{
		"over", "in", "from", "select", "not", "on", "join", "as",
		"where", "and", "or", "having",
	}
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

// normalizeNullComparisons rewrites = NULL to IS NULL and != / <> NULL to IS NOT NULL.
// Comparing with = or != against NULL is always incorrect SQL (always returns UNKNOWN),
// so this rewrite produces the semantically intended form.
func normalizeNullComparisons(tokens antlr.TokenStream) {
	eqOrNeq := map[int]bool{
		parser.SnowflakeLexerEQ:   true,
		parser.SnowflakeLexerNE:   true,
		parser.SnowflakeLexerLTGT: true,
	}
	size := tokens.Size()
	for i := 0; i < size; i++ {
		tok := tokens.Get(i)
		if tok.GetChannel() != antlr.TokenDefaultChannel {
			continue
		}
		if !eqOrNeq[tok.GetTokenType()] {
			continue
		}
		for j := i + 1; j < size; j++ {
			next := tokens.Get(j)
			if next.GetChannel() != antlr.TokenDefaultChannel {
				continue
			}
			if next.GetTokenType() == parser.SnowflakeLexerNULL_ {
				if tok.GetTokenType() == parser.SnowflakeLexerEQ {
					tok.(*antlr.CommonToken).SetText("IS")
				} else {
					tok.(*antlr.CommonToken).SetText("IS NOT")
				}
				next.(*antlr.CommonToken).SetText("NULL")
			}
			break
		}
	}
}

// removeRedundantParens blanks out parentheses that wrap a simple expression
// (no subquery, no top-level AND/OR/NOT) in a boolean context (after WHERE,
// AND, OR, ON, HAVING, SELECT, or COMMA).  Parens whose matching ) is
// immediately followed by an arithmetic operator are preserved to avoid
// changing precedence.
func removeRedundantParens(tokens antlr.TokenStream) {
	arithmeticOps := map[int]bool{
		parser.SnowflakeLexerSTAR:   true,
		parser.SnowflakeLexerDIVIDE: true,
		parser.SnowflakeLexerMODULE: true,
		parser.SnowflakeLexerPLUS:   true,
		parser.SnowflakeLexerMINUS:  true,
	}
	booleanContext := map[int]bool{
		parser.SnowflakeLexerWHERE:  true,
		parser.SnowflakeLexerAND:    true,
		parser.SnowflakeLexerOR:     true,
		parser.SnowflakeLexerON:     true,
		parser.SnowflakeLexerHAVING: true,
		parser.SnowflakeLexerCOMMA:  true,
		parser.SnowflakeLexerSELECT: true,
	}
	topLevelBoolOps := map[int]bool{
		parser.SnowflakeLexerAND: true,
		parser.SnowflakeLexerOR:  true,
		parser.SnowflakeLexerNOT: true,
	}

	size := tokens.Size()
	for i := 0; i < size; i++ {
		tok := tokens.Get(i)
		if tok.GetChannel() != antlr.TokenDefaultChannel {
			continue
		}
		if tok.GetTokenType() != parser.SnowflakeLexerLR_BRACKET {
			continue
		}

		// Find the matching closing paren.
		depth := 1
		j := i + 1
		for j < size && depth > 0 {
			t := tokens.Get(j)
			if t.GetChannel() == antlr.TokenDefaultChannel {
				tt := t.GetTokenType()
				if tt == parser.SnowflakeLexerLR_BRACKET {
					depth++
				} else if tt == parser.SnowflakeLexerRR_BRACKET {
					depth--
				}
			}
			if depth > 0 {
				j++
			}
		}
		if depth != 0 {
			continue
		}
		rparenIdx := j

		// Only remove in boolean/list contexts; this also filters out function calls.
		prevIdx := prevDefaultTokenIndex(tokens, i)
		if prevIdx < 0 || !booleanContext[tokens.Get(prevIdx).GetTokenType()] {
			continue
		}

		// Scan interior for a subquery or top-level boolean operator.
		unsafe := false
		innerDepth := 0
		for k := i + 1; k < rparenIdx; k++ {
			t := tokens.Get(k)
			if t.GetChannel() != antlr.TokenDefaultChannel {
				continue
			}
			tt := t.GetTokenType()
			if tt == parser.SnowflakeLexerLR_BRACKET {
				innerDepth++
				continue
			}
			if tt == parser.SnowflakeLexerRR_BRACKET {
				innerDepth--
				continue
			}
			if innerDepth > 0 {
				continue
			}
			if tt == parser.SnowflakeLexerSELECT || topLevelBoolOps[tt] {
				unsafe = true
				break
			}
		}
		if unsafe {
			continue
		}

		// Preserve parens needed for arithmetic precedence: (a+b)*c.
		nextIdx := nextDefaultTokenIndex(tokens, rparenIdx)
		if nextIdx >= 0 && arithmeticOps[tokens.Get(nextIdx).GetTokenType()] {
			continue
		}

		tokens.Get(i).(*antlr.CommonToken).SetText(" ")
		tokens.Get(rparenIdx).(*antlr.CommonToken).SetText(" ")
	}
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

// scanStatementParens identifies, by token index, which '(' tokens open a
// nested SELECT statement (a subquery or CTE body) as opposed to a plain
// expression grouping (function call args, window OVER(...), IN (...) value
// lists, arithmetic grouping). A '(' opens a statement scope iff the next
// default-channel token is SELECT. Clause keywords encountered inside a
// non-statement paren (e.g. ORDER BY inside OVER(...)) are not real clause
// boundaries and must not trigger clause formatting.
func scanStatementParens(tokens antlr.TokenStream) map[int]bool {
	isStatement := map[int]bool{}
	for i := 0; i < tokens.Size(); i++ {
		tok := tokens.Get(i)
		if tok.GetChannel() != antlr.TokenDefaultChannel || tok.GetTokenType() != parser.SnowflakeLexerLR_BRACKET {
			continue
		}
		for j := i + 1; j < tokens.Size(); j++ {
			next := tokens.Get(j)
			if next.GetChannel() != antlr.TokenDefaultChannel {
				continue
			}
			isStatement[i] = next.GetTokenType() == parser.SnowflakeLexerSELECT
			break
		}
	}
	return isStatement
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

// normalizeOrderDirectionExplicit adds explicit ASC to every ORDER BY item that has
// no direction keyword. Works on both aligned (ORDER BY on its own line) and flat
// (inline) output. Must run before formatCaseExpressions.
func normalizeOrderDirectionExplicit(sql string) string {
	lines := strings.Split(sql, "\n")
	for i, line := range lines {
		upper := strings.ToUpper(line)
		obIdx := strings.Index(upper, "ORDER BY ")
		if obIdx < 0 {
			continue
		}
		prefix := line[:obIdx+len("ORDER BY ")]
		rest := line[obIdx+len("ORDER BY "):]
		// Locate any inline terminator (LIMIT/OFFSET) so we don't consume it.
		itemStr, suffix := splitOrderBySuffix(rest)
		items := splitAtTopLevelCommas(itemStr)
		for j, item := range items {
			items[j] = ensureExplicitOrderDirection(item)
		}
		lines[i] = prefix + strings.Join(items, ", ") + suffix
	}
	return strings.Join(lines, "\n")
}

// splitOrderBySuffix splits an ORDER BY remainder into the item list and any
// trailing clause/LIMIT/OFFSET that follows (e.g. " LIMIT 10").
func splitOrderBySuffix(s string) (items, suffix string) {
	upper := strings.ToUpper(s)
	for _, kw := range []string{" LIMIT ", " LIMIT\t", " OFFSET ", " OFFSET\t"} {
		if idx := strings.Index(upper, kw); idx >= 0 {
			return strings.TrimRight(s[:idx], " \t"), s[idx:]
		}
	}
	for _, kw := range []string{" LIMIT", " OFFSET"} {
		if strings.HasSuffix(upper, kw) {
			idx := len(s) - len(kw)
			return strings.TrimRight(s[:idx], " \t"), s[idx:]
		}
	}
	return s, ""
}

// splitAtTopLevelCommas splits s by commas that are not inside parentheses or
// string literals.
func splitAtTopLevelCommas(s string) []string {
	var items []string
	depth := 0
	inStr := false
	start := 0
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inStr {
			if ch == '\'' {
				if i+1 < len(s) && s[i+1] == '\'' {
					i++ // escaped ''
				} else {
					inStr = false
				}
			}
			continue
		}
		switch ch {
		case '\'':
			inStr = true
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				items = append(items, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	items = append(items, strings.TrimSpace(s[start:]))
	return items
}

// ensureExplicitOrderDirection appends ASC to an ORDER BY item if it has no
// explicit direction. Items already ending in ASC, DESC, FIRST, or LAST are
// returned unchanged (FIRST/LAST covers NULLS FIRST / NULLS LAST).
func ensureExplicitOrderDirection(item string) string {
	trimmed := strings.TrimSpace(item)
	lastSpace := strings.LastIndex(trimmed, " ")
	var lastWord string
	if lastSpace >= 0 {
		lastWord = trimmed[lastSpace+1:]
	} else {
		lastWord = trimmed
	}
	switch strings.ToUpper(lastWord) {
	case "ASC", "DESC", "FIRST", "LAST":
		return item
	}
	return item + " ASC"
}

// formatCTEClosingParens ensures each CTE subquery's closing ) sits on its own
// line, indents the body 4 spaces, and adds a blank line after each CTE closer.
// It detects CTE subqueries by the pattern <word> AS (, tracks paren depth, and
// emits a newline+indent before the ) that brings depth back to zero.
func formatCTEClosingParens(sql string) string {
	const cteIndent = "  "
	var out strings.Builder
	pending := ""
	seenAS := false
	inCTE := false
	cteDepth := 0

	flushPending := func() {
		if pending != "" {
			out.WriteString(pending)
			pending = ""
		}
	}

	i := 0
	for i < len(sql) {
		ch := sql[i]

		// Whitespace — buffer it so we can discard before a `)` closer.
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			j := i
			for j < len(sql) && (sql[j] == ' ' || sql[j] == '\t' || sql[j] == '\n' || sql[j] == '\r') {
				j++
			}
			ws := sql[i:j]
			if inCTE && strings.Contains(ws, "\n") {
				// Inject cteIndent after each newline so CTE body lines are indented.
				var sb strings.Builder
				parts := strings.Split(ws, "\n")
				for k, part := range parts {
					if k > 0 {
						sb.WriteByte('\n')
						sb.WriteString(cteIndent)
					}
					sb.WriteString(part)
				}
				pending += sb.String()
			} else {
				pending += ws
			}
			i = j
			continue
		}

		// Single-quoted string literal — copy verbatim.
		if ch == '\'' {
			flushPending()
			seenAS = false
			j := i + 1
			for j < len(sql) {
				if sql[j] == '\'' {
					j++
					if j < len(sql) && sql[j] == '\'' {
						j++
						continue
					}
					break
				}
				j++
			}
			out.WriteString(sql[i:j])
			i = j
			continue
		}

		// Double-quoted identifier — copy verbatim.
		if ch == '"' {
			flushPending()
			seenAS = false
			j := i + 1
			for j < len(sql) && sql[j] != '"' {
				j++
			}
			if j < len(sql) {
				j++
			}
			out.WriteString(sql[i:j])
			i = j
			continue
		}

		// Opening paren.
		if ch == '(' {
			if seenAS && !inCTE {
				inCTE = true
				cteDepth = 1
			} else if inCTE {
				cteDepth++
			}
			seenAS = false
			flushPending()
			out.WriteByte('(')
			i++
			continue
		}

		// Closing paren.
		if ch == ')' {
			if inCTE {
				cteDepth--
				if cteDepth == 0 {
					pending = "" // discard trailing whitespace before closer
					out.WriteString("\n)\n\n")
					inCTE = false
					seenAS = false
					i++
					// Skip any whitespace immediately following the CTE closer so
					// we don't double-emit the newline that was already there.
					for i < len(sql) && (sql[i] == ' ' || sql[i] == '\t' || sql[i] == '\n' || sql[i] == '\r') {
						i++
					}
					// If a CTE separator comma follows, emit it directly and skip
					// the whitespace after it too, so the next CTE name hugs the
					// comma (e.g. ",b AS (" rather than ", b AS (").
					if i < len(sql) && sql[i] == ',' {
						out.WriteByte(',')
						i++
						for i < len(sql) && (sql[i] == ' ' || sql[i] == '\t' || sql[i] == '\n' || sql[i] == '\r') {
							i++
						}
					}
					continue
				}
			}
			seenAS = false
			flushPending()
			out.WriteByte(')')
			i++
			continue
		}

		// Word token.
		if isWordChar(ch) {
			j := i
			for j < len(sql) && isWordChar(sql[j]) {
				j++
			}
			word := sql[i:j]
			seenAS = strings.ToUpper(word) == "AS"
			flushPending()
			out.WriteString(word)
			i = j
			continue
		}

		// Everything else.
		seenAS = false
		flushPending()
		out.WriteByte(ch)
		i++
	}
	flushPending()
	return out.String()
}

// applyLeadingCommasCTE moves the comma that separates CTE definitions to the
// start of the next line. It only fires on lines produced by formatCTEClosingParens
// that begin with ")," — the CTE closing paren immediately followed by the separator.
func applyLeadingCommasCTE(sql string) string {
	lines := strings.Split(sql, "\n")
	var out []string
	for _, line := range lines {
		if strings.HasPrefix(line, "),") {
			out = append(out, ")")
			out = append(out, ","+line[2:])
		} else {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
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

func ensureTrailingNewline(sql string) string {
	if strings.HasSuffix(sql, "\n") {
		return sql
	}
	return sql + "\n"
}

const inlineMaxLength = 120

// inlineSimpleStatements collapses multi-line SQL statements back to a single
// line when they have no JOINs, no CTEs, and the collapsed line stays under
// inlineMaxLength characters. Statements that exceed the threshold or contain
// JOINs/CTEs are left unchanged, allowing other formatting rules to apply.
// Running before indentSubqueries ensures simple subquery content is already
// on one line when the subquery formatter picks it up.
func inlineSimpleStatements(sql string) string {
	blocks := strings.Split(sql, "\n\n")
	for i, block := range blocks {
		blocks[i] = maybeInlineBlock(block)
	}
	return strings.Join(blocks, "\n\n")
}

func maybeInlineBlock(block string) string {
	if !strings.Contains(block, "\n") {
		return block
	}
	upper := strings.ToUpper(block)
	// Don't inline multi-table queries.
	for _, w := range strings.Fields(upper) {
		if w == "JOIN" {
			return block
		}
	}
	// Don't inline CTEs.
	if strings.HasPrefix(strings.TrimSpace(upper), "WITH ") {
		return block
	}
	trailingNewline := strings.HasSuffix(block, "\n")
	collapsed := strings.Join(strings.Fields(block), " ")
	if len(collapsed) > inlineMaxLength {
		return block
	}
	if trailingNewline {
		collapsed += "\n"
	}
	return collapsed
}

// extractBalancedContent returns the text between the '(' at openIdx and its
// matching ')', the index of that ')', and whether it succeeded.
func extractBalancedContent(sql string, openIdx int) (content string, closeIdx int, ok bool) {
	depth := 1
	i := openIdx + 1
	for i < len(sql) {
		c := sql[i]
		switch c {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return sql[openIdx+1 : i], i, true
			}
		case '\'':
			i++
			for i < len(sql) {
				if sql[i] == '\'' {
					i++
					if i < len(sql) && sql[i] == '\'' {
						i++ // escaped ''
						continue
					}
					break
				}
				i++
			}
			continue
		}
		i++
	}
	return "", openIdx, false
}

// indentSubqueries reformats inline (SELECT ...) subqueries to multi-line with
// 2-space indentation per nesting level. The function is applied recursively so
// that nested subqueries each receive 2 additional spaces relative to their
// containing subquery.
func indentSubqueries(sql string) string {
	var out strings.Builder
	i := 0
	for i < len(sql) {
		ch := sql[i]

		if ch == '\'' {
			out.WriteByte(ch)
			i++
			for i < len(sql) {
				c := sql[i]
				out.WriteByte(c)
				i++
				if c == '\'' {
					if i < len(sql) && sql[i] == '\'' {
						out.WriteByte(sql[i])
						i++ // escaped ''
						continue
					}
					break
				}
			}
			continue
		}

		if ch == '(' {
			j := i + 1
			for j < len(sql) && sql[j] == ' ' {
				j++
			}
			isSubquery := j+6 <= len(sql) &&
				strings.EqualFold(sql[j:j+6], "SELECT") &&
				(j+6 >= len(sql) || !isWordChar(sql[j+6]))

			if isSubquery {
				content, endIdx, ok := extractBalancedContent(sql, i)
				if ok && !strings.Contains(content, "\n") {
					// Recursively apply to inner content so nested subqueries
					// are formatted first with relative indentation.
					processed := indentSubqueries(strings.TrimSpace(content))
					// Indent every line of the processed content by 2 spaces.
					lines := strings.Split(processed, "\n")
					for k, line := range lines {
						if len(line) > 0 {
							lines[k] = "  " + line
						}
					}
					out.WriteString("(\n")
					out.WriteString(strings.Join(lines, "\n"))
					out.WriteString("\n)")
					i = endIdx + 1
					continue
				}
			}
		}

		out.WriteByte(ch)
		i++
	}
	return out.String()
}

func ensureTrailingSemicolon(sql string) string {
	trimmed := strings.TrimRight(sql, " \t\n\r")
	trimmed = strings.TrimRight(trimmed, ";")
	trimmed = strings.TrimRight(trimmed, " \t\n\r")
	return trimmed + "\n;"
}
