# Contributing to MRXL

Thank you for taking the time to contribute!

## Getting Started

**Prerequisites:** Go 1.22+

```bash
git clone https://github.com/v420v/mrxl.git
cd mrxl
go mod download
go build ./...
go test ./...
```

## Project Structure

```
cmd/           # CLI entry point
internal/
  ast/         # AST type definitions for each diagram
  parser/      # Mermaid text → AST
  gen/         # AST → Excel (.xlsx)
examples/      # Sample .mmd files
```

Adding a new diagram type requires three files: `internal/ast/<type>.go`, `internal/parser/<type>.go`, and `internal/gen/<type>.go`, plus registration in `parser/parser.go` and `gen/gen.go`.

## Making Changes

1. Fork the repository and create a branch from `main`.
2. Make your changes. Add or update tests in `internal/parser/` as appropriate.
3. Run `go test ./...` and `go build ./...` — both must pass.
4. Open a pull request against `main`.

## Commit Style

This project uses a simple prefix convention (see `.gitmessage`):

```
add: short summary
fix: short summary
refactor: short summary
test: short summary
docs: short summary
chore: short summary
```

Use the imperative mood and keep the summary under 50 characters.

## Reporting Issues

Please use the [issue templates](.github/ISSUE_TEMPLATE/) when opening a bug report or feature request.
