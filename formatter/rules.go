package formatter

import (
  "regexp"
  "strings"

  antlr "github.com/antlr4-go/antlr/v4"
  "github.com/jorwoods/sqlfmt/parser"
)


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
  return clauseTokenTypes[token.GetTokenType()]
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

// rightAlignClausesWithConfig aligns clauses and SELECT lists, respecting config for uppercasing.
func rightAlignClausesWithConfig(tokens antlr.TokenStream, cfg *Config) string {
  var out []string
  var line []string
  currentClause := ""
  clauseIndent := map[int]string{
    parser.SnowflakeLexerSELECT:  "  ",
    parser.SnowflakeLexerFROM:    "    ",
    parser.SnowflakeLexerWHERE:   "   ",
    parser.SnowflakeLexerGROUP:   "   ",
    parser.SnowflakeLexerHAVING:  "   ",
    parser.SnowflakeLexerORDER:   "   ",
    parser.SnowflakeLexerQUALIFY: "   ",
  }
  var inSelect bool
  var selectIdents []string
  var lastClauseText string
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
        // Use lastClauseText for SELECT in output
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
              out = append(out, "  "+selectWord+" "+ident+comma)
            } else {
              out = append(out, "         "+ident+comma)
            }
          }
        } else if len(selectIdents) > 0 {
          out = append(out, "  "+selectWord+" "+strings.Join(selectIdents, ", "))
        }
        selectIdents = nil
        inSelect = false
      }
      if len(line) > 0 {
        out = append(out, strings.Join(line, " "))
        line = nil
      }
      indent := clauseIndent[ttype]
      clauseText := text
      if cfg == nil || cfg.Rules.UppercaseKeywords {
        clauseText = strings.ToUpper(text)
      }
      currentClause = indent + clauseText
      if ttype == parser.SnowflakeLexerSELECT {
        inSelect = true
        lastClauseText = clauseText
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
    if currentClause != "" {
      line = append([]string{currentClause, text})
      currentClause = ""
    } else {
      line = append(line, text)
    }
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
          out = append(out, "  "+selectWord+" "+ident+comma)
        } else {
          out = append(out, "         "+ident+comma)
        }
      }
    } else if len(selectIdents) > 0 {
      out = append(out, "  "+selectWord+" "+strings.Join(selectIdents, ", "))
    }
    selectIdents = nil
    inSelect = false
  }
  if len(line) > 0 {
    out = append(out, strings.Join(line, " "))
  }
  return strings.Join(out, "\n")
}

// rightAlignClauses uses all rules enabled (for legacy calls)
func rightAlignClauses(tokens antlr.TokenStream) string {
  return rightAlignClausesWithConfig(tokens, nil)
}

func tokensToText(tokens antlr.TokenStream) string {
  var b strings.Builder
  for i := 0; i < tokens.Size(); i++ {
    b.WriteString(tokens.Get(i).GetText())
  }
  return b.String()
}

