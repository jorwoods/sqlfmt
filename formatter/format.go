package formatter

import (
	antlr "github.com/antlr4-go/antlr/v4"
	"github.com/jorwoods/sqlfmt/parser"
)

func rulesAllDisabled(rules RulesConfig) bool {
	return !rules.UppercaseKeywords &&
		!rules.AlignClauses &&
		!rules.StripQuotes &&
		!rules.FormatSelectList &&
		!rules.RequireExplicitAS &&
		!rules.TrailingSemicolon &&
		!rules.StripTrailingWhitespace &&
		!rules.NormalizeNotEqual &&
		!rules.BlankLinesBetweenStatements &&
		!rules.NewlineBeforeAndOr &&
		!rules.NormalizeBoolean &&
		!rules.UppercaseFunctions &&
		!rules.NewlineBeforeJoin &&
		!rules.NewlineBeforeOn &&
		!rules.IndentCaseWhen &&
		!rules.LeadingComma &&
		!rules.NormalizeNullComparison &&
		!rules.TrailingNewline &&
		!rules.NewlineBeforeLimit &&
		!rules.NormalizeOrderDirection &&
		!rules.CTEFormatting &&
		!rules.LeadingCommaCTE
}

func effectiveRules(cfg *Config) RulesConfig {
	if cfg != nil {
		return cfg.Rules
	}
	return RulesConfig{
		UppercaseKeywords:           true,
		AlignClauses:                true,
		StripQuotes:                 true,
		FormatSelectList:            true,
		RequireExplicitAS:           false,
		OperatorSpacing:             true,
	}
}

func lexAndParse(input string) antlr.TokenStream {
	is := antlr.NewInputStream(input)
	lexer := parser.NewSnowflakeLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := parser.NewSnowflakeParser(stream)
	p.BuildParseTrees = true
	p.Snowflake_file()

	return stream
}

func applyTokenRules(stream antlr.TokenStream, rules RulesConfig) {
	// Apply rules in the correct order for combined effects.
	if rules.UppercaseKeywords {
		uppercaseKeywords(stream)
	}
	if rules.StripQuotes {
		stripQuotesIfSafe(stream)
	}
	if rules.RequireExplicitAS {
		// Explicit aliasing operates on the token stream by mutating token text.
		requireExplicitAS(stream, &Config{Rules: rules})
	}
	if rules.NormalizeNotEqual {
		normalizeNotEqual(stream)
	}
	if rules.NormalizeBoolean {
		normalizeBooleans(stream)
	}
	if rules.NormalizeNullComparison {
		normalizeNullComparisons(stream)
	}
	if rules.UppercaseFunctions {
		uppercaseFunctions(stream)
	}
}

func render(stream antlr.TokenStream, rules RulesConfig) string {
	if rules.AlignClauses && rules.FormatSelectList {
		return rightAlignClausesWithConfig(stream, &Config{Rules: rules})
	}
	if rules.AlignClauses {
		return alignClausesOnly(stream, &Config{Rules: rules})
	}
	if rules.FormatSelectList {
		return formatSelectListOnly(stream, &Config{Rules: rules})
	}
	return tokensToText(stream, &Config{Rules: rules})
}

// FormatSQLWithConfig formats SQL using the provided config.
func FormatSQLWithConfig(input string, cfg *Config) string {
	rules := effectiveRules(cfg)
	if cfg != nil && rulesAllDisabled(rules) {
		return input
	}

	stream := lexAndParse(input)
	applyTokenRules(stream, rules)
	result := render(stream, rules)
	if rules.NormalizeOrderDirection {
		result = normalizeOrderDirectionExplicit(result)
	}
	if rules.CTEFormatting {
		result = formatCTEClosingParens(result)
	}
	if rules.LeadingCommaCTE {
		result = applyLeadingCommasCTE(result)
	}
	if rules.IndentCaseWhen {
		result = formatCaseExpressions(result)
	}
	if rules.LeadingComma {
		result = applyLeadingCommas(result)
	}
	if rules.TrailingSemicolon {
		result = ensureTrailingSemicolon(result)
	}
	if rules.StripTrailingWhitespace {
		result = stripTrailingWhitespace(result)
	}
	if rules.BlankLinesBetweenStatements {
		result = blankLinesBetweenStatements(result)
	}
	if rules.TrailingNewline {
		result = ensureTrailingNewline(result)
	}
	return result
}

// FormatSQL formats SQL using default rules (all enabled).
func FormatSQL(input string) string {
	return FormatSQLWithConfig(input, nil)
}
