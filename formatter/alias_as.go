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

func isWordLikeToken(tok antlr.Token) bool {
	text := strings.TrimSpace(tok.GetText())
	if text == "" || strings.EqualFold(text, "<EOF>") {
		return false
	}
	if isPunctuation(text) {
		return false
	}
	return true
}

// requireExplicitAS mutates token *text* in-place so that aliases become explicit by suffixing " AS" onto the aliased token.
// This operates purely on the token stream (no re-tokenization) and always inserts uppercase AS.
func requireExplicitAS(tokens antlr.TokenStream, cfg *Config) {
	if cfg == nil || !cfg.Rules.RequireExplicitAS {
		return
	}

	currentClause := 0
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

		if clauseTokenTypes[ttype] || ttype == parser.SnowflakeLexerJOIN {
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
		if !isWordLikeToken(tok) {
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
		if clauseTokenTypes[nextTok.GetTokenType()] || nextTok.GetTokenType() == parser.SnowflakeLexerJOIN {
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


