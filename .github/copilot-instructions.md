# Copilot Instructions for sqlfmt

## Rule Management
- **Every new formatting rule must have a corresponding config entry** in `sqlfmt.yaml` and the `Config` struct in `formatter/config.go`.
- The `README.md` must include an up-to-date list and description of all available rules and their config keys.
- When adding a new rule, update:
  - `sqlfmt.yaml` (example config)
  - `formatter/config.go` (struct and loader)
  - `README.md` (rule documentation)
  - Add or update tests in `formatter/config_test.go` and/or other relevant test files.

## Testing
- **All new functionality must have corresponding tests.**
- Tests should cover config parsing, rule application, and edge cases for the new rule.

## Commit Messages
- **All commit messages must follow the [Conventional Commits](https://www.conventionalcommits.org/) style.**
  - Example: `feat: add config option for keyword lowercasing`
  - Example: `fix: respect strip_quotes config in identifier output`
  - Example: `test: add tests for new align_clauses rule`

## General Guidance
- Keep the codebase grammar-driven and future-proof.
- Never use tabs for indentation or alignment; always use spaces (enforced by `.editorconfig`).
- Ensure the config loader searches up the directory tree for `sqlfmt.yaml`.
- All user-facing features should be documented in `README.md`.
