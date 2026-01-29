# AGENTS.md

## Codebase Map

- **main.go**: CLI entry point. Runs `cmd.Execute()`.
- **cmd/root.go**: Cobra CLI setup. Handles file input/output and invokes the formatter.
- **formatter/**: Core formatting logic and configuration.
  - `format.go`, `rules.go`, `alias_as.go`, `cte_refactor.go`: Formatting rules and orchestration.
  - `config.go`: Loads/parses `sqlfmt.yaml` config, defines rule toggles.
  - `config_test.go`, `format_test.go`: Unit tests for config and formatting.
- **parser/**: ANTLR-generated Snowflake SQL grammar, lexer, and parser. Do not edit generated files directly.
- **sqlfmt.yaml**: Example config for rule toggles (see README for all options).
- **.editorconfig**: Enforces spaces-only indentation and other style rules.
- **README.md**: Project overview, rule documentation, and contribution guide.
- **.github/copilot-instructions.md**: Agent-specific and contributor instructions.

## Directory Roles & Naming

- **formatter/**: All formatting logic, config, and tests. Naming is explicit: `*_test.go` for tests, `config.go` for config, `rules.go` for rule logic.
- **parser/**: Only grammar and generated parser/lexer for Snowflake SQL.
- **cmd/**: CLI wiring and argument handling.
- **.github/**: Agent and automation instructions.

## Configuration Wiring

- Rules are toggled in `sqlfmt.yaml` under the `rules:` key.
- `formatter/config.go` defines the `Config` struct and loads config, searching up the directory tree.
- All new rules must be added to `sqlfmt.yaml`, `formatter/config.go`, and documented in `README.md`.

## Test Locations

- All tests are in `formatter/format_test.go` and `formatter/config_test.go`.
- Tests cover config parsing, rule application, and edge cases.

## Local Norms

- **Spaces only**: No tabs for indentation or alignment (enforced by `.editorconfig`).
- **Conventional Commits**: All commit messages must follow [Conventional Commits](https://www.conventionalcommits.org/) style.
- **Rule changes**: Every new formatting rule must have a config entry, code, tests, and documentation.
- **Config loader**: Always searches up the directory tree for `sqlfmt.yaml`.
- **User-facing features**: Must be documented in `README.md`.
- **Tests required**: All new features must have corresponding tests.
- **Grammar-driven**: No hand-maintained keyword lists; always use the grammar.

## Self-correction

- If the code map above is discovered to be stale, update it immediately.
- If the user gives a correction about how work should be done in this repo, add it to "Local Norms" (or another clearly labeled section) so future sessions inherit it.
