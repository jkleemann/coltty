# AGENTS.md

Guidelines for AI agents working in this repository.

## Go Conventions

- Use `gofmt` and `go vet` before committing. Code that doesn't pass both should not be committed.
- Keep packages flat unless there's a clear reason to nest. Prefer `adapter/ghostty.go` over `adapter/ghostty/ghostty.go`.
- Use the standard `error` return pattern. Don't panic except for truly unrecoverable programmer errors.
- Prefer the standard library over third-party packages. Only add dependencies when they provide substantial value (e.g., `cobra` for CLI, `BurntSushi/toml` for TOML parsing).
- Name interfaces by what they do, not what they are: `TerminalAdapter`, not `ITerminalAdapter`.
- Test files live next to the code they test: `config.go` and `config_test.go` in the same package.
- Use table-driven tests for functions with multiple input/output cases.
- Don't use `init()` functions. Pass dependencies explicitly.

## Committing

- Run `go build ./...` and `go test ./...` before every commit. Do not commit code that doesn't compile or has failing tests.
- Write commit messages that explain *why*, not *what*. The diff shows what changed.
- Keep commits atomic: one logical change per commit. Don't mix refactoring with new features.
- Never commit generated files, binaries, or editor artifacts.

## Verification

- Before claiming any task is complete, run:
  ```bash
  go build ./...
  go vet ./...
  go test ./...
  ```
- All three must pass with zero errors. Warnings from `go vet` are treated as errors.
- When adding a new feature, write tests that cover the happy path and at least one error case.
- When fixing a bug, write a test that reproduces the bug before writing the fix.
- After modifying the Ghostty adapter, manually verify the generated fragment file is valid Ghostty config syntax.
- After modifying shell integration, verify the generated hook code is valid shell syntax by running it through `bash -n` or `zsh -n`.
