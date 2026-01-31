package formatter

import (
	antlr "github.com/antlr4-go/antlr/v4"
	"github.com/jorwoods/sqlfmt/parser"
	"strings"
)

// enforceExplicitAliasAS returns a new SQL string with explicit AS for all aliases in SELECT and FROM
// enforceExplicitAliasASWithConfig returns a new SQL string with explicit AS for all aliases in SELECT and FROM, respecting config
func enforceExplicitAliasASWithConfig(tokens antlr.TokenStream, cfg *Config) string {
       var out strings.Builder
       total := tokens.Size()
       isWordType := func(ttype int) bool {
	       return ttype == parser.SnowflakeLexerID ||
		      ttype == parser.SnowflakeLexerID2 ||
		      ttype == parser.SnowflakeLexerDOUBLE_QUOTE_ID ||
		      ttype == parser.SnowflakeLexerIDENTIFIER ||
		      ttype == parser.SnowflakeLexerAS ||
		      ttype == parser.SnowflakeLexerNUMBER
       }
       i := 0
       for i < total {
	       tok := tokens.Get(i)
	       text := tok.GetText()
	       ttype := tok.GetTokenType()
	       if tok.GetChannel() != antlr.TokenDefaultChannel || text == "<EOF>" {
		       i++
		       continue
	       }
	       // Only apply AS insertion in SELECT/ FROM if FormatSelectList is false or not in a select list
	       if i < total-1 {
		       next := tokens.Get(i+1)
		       if isWordType(ttype) && isWordType(next.GetTokenType()) && isSelectOrFromAliasPattern(tok, next, tokens, i) {
			       // Only insert AS if FormatSelectList is false or not in a select list
			       if cfg == nil || !cfg.Rules.FormatSelectList {
				       if out.Len() > 0 && ttype != parser.SnowflakeLexerCOMMA && ttype != parser.SnowflakeLexerRR_BRACKET {
					       out.WriteString(" ")
				       }
				       out.WriteString(text)
				       out.WriteString(" AS ")
				       out.WriteString(next.GetText())
				       i += 2
				       if i < total && tokens.Get(i).GetTokenType() == parser.SnowflakeLexerCOMMA {
					       out.WriteString(",")
					       i++
				       }
				       continue
			       }
		       }
	       }
	       if out.Len() > 0 && ttype != parser.SnowflakeLexerCOMMA && ttype != parser.SnowflakeLexerRR_BRACKET {
		       out.WriteString(" ")
	       }
	       out.WriteString(text)
	       i++
       }
       return out.String()
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

// isSelectOrFromAliasPattern detects col alias or table alias without AS
func isSelectOrFromAliasPattern(tok, next antlr.Token, tokens antlr.TokenStream, idx int) bool {
	   // Use enums for all identifier token types present in the generated lexer
	   isIdent := func(ttype int) bool {
			   return ttype == parser.SnowflakeLexerID ||
					  ttype == parser.SnowflakeLexerID2 ||
					  ttype == parser.SnowflakeLexerDOUBLE_QUOTE_ID ||
					  ttype == parser.SnowflakeLexerIDENTIFIER
	   }
	   if isIdent(tok.GetTokenType()) && isIdent(next.GetTokenType()) {
			   // Check previous token is not AS
			   if idx > 0 {
					   prev := tokens.Get(idx-1)
					   if strings.EqualFold(prev.GetText(), "AS") {
							   return false
					   }
			   }
			   return true
	   }
	   return false
}
