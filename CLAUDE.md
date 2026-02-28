# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Coltty is a Go CLI tool and shell hook that automatically switches terminal color schemes based on the current directory. It walks up the directory tree looking for `.coltty.toml` config files, resolves a color scheme, and applies it via terminal-specific adapters (Ghostty first).

## Build & Test

```bash
go build -o coltty .
go test ./...
go test -run TestName ./...   # single test
```

## Architecture

See `docs/plans/2026-02-28-coltty-design.md` for the full design specification.

Key concepts:
- **Config resolution**: walk from current directory up to root, first `.coltty.toml` wins, named schemes looked up from global config at `~/.config/coltty/config.toml`
- **Adapter interface**: `TerminalAdapter` with `Apply()` and `Detect()` methods, per-terminal implementations. Ghostty adapter writes a config fragment file that Ghostty watches for changes.
- **Shell hook**: thin `chpwd`/`PROMPT_COMMAND` hook that calls `coltty apply --quiet` in the background on every `cd`
