# sqlfmt

A future-proof SQL formatter for Snowflake SQL, driven by ANTLR grammar.

## Overview

**sqlfmt** is a command-line tool and Go library for formatting SQL code written for the Snowflake data platform. It uses a parser and lexer generated from the official Snowflake SQL grammar (ANTLR), ensuring that all keyword, function, and clause logic is always in sync with the grammar—no hand-maintained keyword lists or brittle regexes.

- **Grammar-driven:** All keyword and function detection is based on the generated parser/lexer, not custom slices or maps.
- **Consistent formatting:** Uppercases keywords and function names, aligns clauses, and formats SELECT lists for readability.
- **Spaces only:** All output uses spaces for alignment and indentation (never tabs). See `.editorconfig` for enforcement.
- **Robust and maintainable:** Changes to the grammar are automatically reflected in formatting logic, making the tool future-proof.

## Formatting Rules and Config

You can enable or disable individual formatting rules via `sqlfmt.yaml`:

| Rule                           | Config Key                         | Description                                                          |
|--------------------------------|------------------------------------|----------------------------------------------------------------------|
| Uppercase Keywords             | `uppercase_keywords`               | Uppercase all SQL keywords                                           |
| Uppercase Functions            | `uppercase_functions`              | Uppercase all SQL built-in function names                            |
| Align Clauses                  | `align_clauses`                    | Align major SQL clauses for readability                              |
| Strip Quotes                   | `strip_quotes`                     | Remove quotes from identifiers when safe                             |
| Format SELECT List             | `format_select_list`               | Format long SELECT lists vertically and aligned                      |
| Require Explicit AS            | `require_explicit_as`              | Require all column and table aliases to use the AS keyword           |
| Strip Trailing Whitespace      | `strip_trailing_whitespace`        | Remove trailing spaces and tabs from each line of output             |
| Normalize Not Equal            | `normalize_not_equal`              | Rewrite `<>` to `!=`                                                 |
| Normalize Boolean              | `normalize_boolean`                | Normalize boolean literals to a consistent case (e.g. `TRUE`/`FALSE`) |
| Operator Spacing               | `operator_spacing`                 | Ensure spaces around operators; `false` enables compact mode (`a=b`) |
| Blank Lines Between Statements | `blank_lines_between_statements`   | Insert a blank line between SQL statements                           |
| Newline Before AND/OR          | `newline_before_and_or`            | Place AND/OR at the start of a new line in WHERE/HAVING clauses      |
| Newline Before JOIN            | `newline_before_join`              | Place each JOIN clause on a new line                                 |
| Newline Before ON              | `newline_before_on`                | Place the ON condition of a JOIN on a new line                       |
| Indent CASE WHEN               | `indent_case_when`                 | Place WHEN, ELSE, and END each on their own indented line in CASE expressions |
| Leading Comma                  | `leading_comma`                    | Place commas at the start of each item line rather than the end of the previous line |
| Normalize NULL Comparison      | `normalize_null_comparison`        | Rewrite `= NULL` to `IS NULL` and `!= NULL` / `<> NULL` to `IS NOT NULL`            |
| Trailing Newline               | `trailing_newline`                 | Ensure the output ends with exactly one newline character                            |
| Newline Before LIMIT/OFFSET    | `newline_before_limit`             | Place LIMIT and OFFSET each on their own line                                        |
| Normalize Order Direction      | `normalize_order_direction`        | Add explicit `ASC` to every ORDER BY item that has no direction keyword (requires `align_clauses`) |
| CTE Formatting                 | `cte_formatting`                   | Place the closing `)` of each CTE subquery on its own line                           |

Example `sqlfmt.yaml`:

```yaml
rules:
  uppercase_keywords: true
  uppercase_functions: true
  align_clauses: true
  strip_quotes: true
  format_select_list: true
  require_explicit_as: false
  strip_trailing_whitespace: true
  normalize_not_equal: true
  normalize_boolean: true
  operator_spacing: true
  blank_lines_between_statements: true
  newline_before_and_or: true
  newline_before_join: true
  newline_before_on: true
  indent_case_when: true
  leading_comma: false
  normalize_null_comparison: true
  trailing_newline: true
  newline_before_limit: true
  normalize_order_direction: true
  cte_formatting: true
```

## Features

- Uppercases all SQL keywords and built-in function names.
- Aligns major SQL clauses (SELECT, FROM, WHERE, etc.) for readability.
- Formats long SELECT lists vertically, with consistent indentation.
- Strips quotes from identifiers when safe.
- Designed for Snowflake SQL, but extensible to other dialects with grammar changes.

## Usage

### As a CLI

```
go run main.go < input.sql > output.sql
```

Or build and install:

```
go build -o sqlfmt
./sqlfmt < input.sql > output.sql
```

### As a Library

Import and use the `FormatSQL` function:

```go
import "github.com/jorwoods/sqlfmt/formatter"

formatted := formatter.FormatSQL(rawSQL)
```

## Development

- The formatter logic is in `formatter/rules.go` and `formatter/format.go`.
- The grammar and generated parser/lexer are in the `parser/` directory.
- All formatting rules are driven by the generated grammar—no custom keyword or function lists.
- Tests are in `formatter/format_test.go` and `formatter/config_test.go`.
- Code style is enforced by `.editorconfig` (spaces only).
- **When adding a new rule:**
  - Add a config entry in `sqlfmt.yaml` and `formatter/config.go`.
  - Document the rule in this README.
  - Add or update tests.
  - Use a [Conventional Commits](https://www.conventionalcommits.org/) style commit message.

## License

MIT License. See [LICENSE](LICENSE) for details.
