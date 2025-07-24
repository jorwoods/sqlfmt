//go:build generate
// +build generate

package parser

//go:generate antlr4 -Dlanguage=Go -o . SnowflakeLexer.g4 SnowflakeParser.g4

