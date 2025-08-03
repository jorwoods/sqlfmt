# sqlfmt

A future-proof SQL formatter for Snowflake SQL, driven by ANTLR grammar.

## Overview

**sqlfmt** is a command-line tool and Go library for formatting SQL code written for the Snowflake data platform. It uses a parser and lexer generated from the official Snowflake SQL grammar (ANTLR), ensuring that all keyword, function, and clause logic is always in sync with the grammar—no hand-maintained keyword lists or brittle regexes.

- **Grammar-driven:** All keyword and function detection is based on the generated parser/lexer, not custom slices or maps.
- **Consistent formatting:** Uppercases keywords and function names, aligns clauses, and formats SELECT lists for readability.
- **Spaces only:** All output uses spaces for alignment and indentation (never tabs). See `.editorconfig` for enforcement.
- **Robust and maintainable:** Changes to the grammar are automatically reflected in formatting logic, making the tool future-proof.

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
- Tests are in `formatter/format_test.go`.
- Code style is enforced by `.editorconfig` (spaces only).

## License

MIT License. See [LICENSE](LICENSE) for details.
