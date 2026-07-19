package formatter

import (
	"strings"

	antlr "github.com/antlr4-go/antlr/v4"
	"github.com/jorwoods/sqlfmt/parser"
)

func nextDefaultTokenIndex(tokens antlr.TokenStream, i int) int {
	for j := i + 1; j < tokens.Size(); j++ {
		tok := tokens.Get(j)
		if tok.GetChannel() == antlr.TokenDefaultChannel {
			return j
		}
	}
	return -1
}

func prevDefaultTokenIndex(tokens antlr.TokenStream, i int) int {
	for j := i - 1; j >= 0; j-- {
		tok := tokens.Get(j)
		if tok.GetChannel() == antlr.TokenDefaultChannel {
			return j
		}
	}
	return -1
}

// nonAliasKeywordTypes are keyword token types that can never be part of an
// implicit alias pair -- either as the token immediately before the alias
// or as the alias candidate itself. clauseTokenTypes (SELECT/FROM/WHERE/
// GROUP/HAVING/ORDER/QUALIFY/UNION/INTERSECT/EXCEPT) are excluded via a
// separate check in isWordLikeToken; this covers everything else that can
// appear inside a SELECT list or JOIN...ON without being wrapped in parens
// (a CASE...END expression has no parens at all) or with parens that don't
// open a nested statement (window function OVER(...)/PARTITION BY).
var nonAliasKeywordTypes = map[int]bool{
	parser.SnowflakeLexerJOIN:      true,
	parser.SnowflakeLexerON:        true,
	parser.SnowflakeLexerAND:       true,
	parser.SnowflakeLexerOR:        true,
	parser.SnowflakeLexerNOT:       true,
	parser.SnowflakeLexerCASE:      true,
	parser.SnowflakeLexerWHEN:      true,
	parser.SnowflakeLexerTHEN:      true,
	parser.SnowflakeLexerELSE:      true,
	parser.SnowflakeLexerEND:       true,
	parser.SnowflakeLexerIS:        true,
	parser.SnowflakeLexerNULL_:     true,
	parser.SnowflakeLexerTRUE:      true,
	parser.SnowflakeLexerFALSE:     true,
	parser.SnowflakeLexerIN:        true,
	parser.SnowflakeLexerLIKE:      true,
	parser.SnowflakeLexerBETWEEN:   true,
	parser.SnowflakeLexerASC:       true,
	parser.SnowflakeLexerDESC:      true,
	parser.SnowflakeLexerDISTINCT:  true,
	parser.SnowflakeLexerALL:       true,
	parser.SnowflakeLexerANY:       true,
	parser.SnowflakeLexerSOME:      true,
	parser.SnowflakeLexerEXISTS:    true,
	parser.SnowflakeLexerOVER:      true,
	parser.SnowflakeLexerPARTITION: true,
	parser.SnowflakeLexerWITHIN:    true,
	parser.SnowflakeLexerROWS:      true,
	parser.SnowflakeLexerCURRENT:   true,
	parser.SnowflakeLexerBY:        true,
}

// nonAliasOperatorTypes are comparison/arithmetic operator token types.
// These are never alias candidates, and a token immediately followed by one
// is never an alias anchor either -- there's an operator in between.
var nonAliasOperatorTypes = map[int]bool{
	parser.SnowflakeLexerEQ:     true,
	parser.SnowflakeLexerGT:     true,
	parser.SnowflakeLexerGE:     true,
	parser.SnowflakeLexerLT:     true,
	parser.SnowflakeLexerLE:     true,
	parser.SnowflakeLexerNE:     true,
	parser.SnowflakeLexerLTGT:   true,
	parser.SnowflakeLexerPLUS:   true,
	parser.SnowflakeLexerMINUS:  true,
	parser.SnowflakeLexerDIVIDE: true,
	parser.SnowflakeLexerSTAR:   true,
	parser.SnowflakeLexerMODULE: true,
}

func isWordLikeToken(tok antlr.Token) bool {
	text := strings.TrimSpace(tok.GetText())
	if text == "" || strings.EqualFold(text, "<EOF>") {
		return false
	}
	if isPunctuation(text) {
		return false
	}
	ttype := tok.GetTokenType()
	if clauseTokenTypes[ttype] || nonAliasKeywordTypes[ttype] || nonAliasOperatorTypes[ttype] {
		return false
	}
	// Join qualifiers (LEFT JOIN, INNER JOIN, ...) precede JOIN itself, not
	// an alias -- without this, "FROM t f LEFT JOIN ..." would treat "f"
	// and "LEFT" as an implicit alias pair and double up the AS already
	// inserted for "f".
	if joinQualifierTypes[ttype] || ttype == parser.SnowflakeLexerOUTER {
		return false
	}
	return true
}

// isAliasAnchorToken reports whether tok can be the token immediately
// before an implicit alias -- either a plain word-like token (identifier,
// literal, ...) or a ')'/END, which are the two ways a SELECT-list
// expression or a FROM/JOIN table reference can end without a trailing
// word (a function call, a parenthesized subquery, or a CASE expression).
func isAliasAnchorToken(tok antlr.Token) bool {
	ttype := tok.GetTokenType()
	if ttype == parser.SnowflakeLexerRR_BRACKET || ttype == parser.SnowflakeLexerEND {
		return true
	}
	return isWordLikeToken(tok)
}

// requireExplicitAS mutates token *text* in-place so that aliases become explicit by suffixing " AS" onto the aliased token.
// This operates purely on the token stream (no re-tokenization) and always inserts uppercase AS.
//
// Alias insertion is only considered while currentClause is SELECT, FROM, or
// JOIN, and only for a bare pair of adjacent word-like tokens with nothing
// else between them (real SQL grammar: that's exactly what an implicit,
// no-AS alias looks like). currentClause is saved/restored across a paren
// pair so that a nested statement's own clause state (a subquery in the
// SELECT list or FROM clause) doesn't leak out and suppress alias detection
// once the parens close.
func requireExplicitAS(tokens antlr.TokenStream, cfg *Config) {
	if cfg == nil || !cfg.Rules.RequireExplicitAS {
		return
	}

	currentClause := 0
	var clauseStack []int
	for i := 0; i < tokens.Size(); i++ {
		tok := tokens.Get(i)
		if tok.GetChannel() != antlr.TokenDefaultChannel {
			continue
		}
		text := tok.GetText()
		ttype := tok.GetTokenType()
		if strings.EqualFold(text, "<EOF>") {
			break
		}

		if ttype == parser.SnowflakeLexerLR_BRACKET {
			clauseStack = append(clauseStack, currentClause)
			continue
		}
		if ttype == parser.SnowflakeLexerRR_BRACKET {
			if len(clauseStack) > 0 {
				currentClause = clauseStack[len(clauseStack)-1]
				clauseStack = clauseStack[:len(clauseStack)-1]
			}
			// A ')' can itself be the token immediately before an alias --
			// "FROM (SELECT ...) s" or "SELECT max(x) total" -- so fall
			// through to the eligibility checks below instead of skipping
			// it outright.
		} else if clauseTokenTypes[ttype] || ttype == parser.SnowflakeLexerJOIN || ttype == parser.SnowflakeLexerON {
			currentClause = ttype
			continue
		}

		// Only consider alias insertion within SELECT / FROM / JOIN clauses.
		if currentClause != parser.SnowflakeLexerSELECT && currentClause != parser.SnowflakeLexerFROM && currentClause != parser.SnowflakeLexerJOIN {
			continue
		}
		if ttype == parser.SnowflakeLexerAS {
			continue
		}
		if !isAliasAnchorToken(tok) {
			continue
		}

		nextIdx := nextDefaultTokenIndex(tokens, i)
		if nextIdx == -1 {
			continue
		}
		nextTok := tokens.Get(nextIdx)
		if nextTok.GetTokenType() == parser.SnowflakeLexerAS {
			continue
		}
		if !isWordLikeToken(nextTok) {
			continue
		}

		prevIdx := prevDefaultTokenIndex(tokens, i)
		if prevIdx != -1 {
			prevTok := tokens.Get(prevIdx)
			if prevTok.GetTokenType() == parser.SnowflakeLexerAS {
				continue
			}
		}

		// Suffix AS onto the current token, unless it's already there.
		if strings.HasSuffix(text, " AS") {
			continue
		}
		tok.(*antlr.CommonToken).SetText(text + " AS")
	}
}

// needsSpace determines if a space is needed between prev and curr token types
func needsSpace(prev, curr int) bool {
	   // No space before comma, right paren, or after left paren
	   if curr == parser.SnowflakeLexerCOMMA || curr == parser.SnowflakeLexerRR_BRACKET {
			   return false
	   }
	   if prev == parser.SnowflakeLexerLR_BRACKET {
			   return false
	   }
	   // Space after comma (before identifier/keyword/number)
	   if prev == parser.SnowflakeLexerCOMMA && (curr == parser.SnowflakeLexerIDENTIFIER || curr == parser.SnowflakeLexerAS || curr == parser.SnowflakeLexerNUMBER) {
			   return true
	   }
	   // Space between identifiers, AS, numbers
	   if (prev == parser.SnowflakeLexerIDENTIFIER || prev == parser.SnowflakeLexerAS || prev == parser.SnowflakeLexerNUMBER) &&
		  (curr == parser.SnowflakeLexerIDENTIFIER || curr == parser.SnowflakeLexerAS || curr == parser.SnowflakeLexerNUMBER) {
			   return true
	   }
	   // Space between keywords and identifiers/AS
	   if (prev == parser.SnowflakeLexerSELECT || prev == parser.SnowflakeLexerFROM || prev == parser.SnowflakeLexerWHERE) &&
		  (curr == parser.SnowflakeLexerIDENTIFIER || curr == parser.SnowflakeLexerAS) {
			   return true
	   }
	   // Space after right paren before identifier/keyword/number
	   if prev == parser.SnowflakeLexerRR_BRACKET && (curr == parser.SnowflakeLexerIDENTIFIER || curr == parser.SnowflakeLexerAS || curr == parser.SnowflakeLexerNUMBER) {
			   return true
	   }
	   // Default: no space
	   return false
}


