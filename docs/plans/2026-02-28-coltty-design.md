# Coltty Design

A Rust CLI tool and shell hook that automatically switches terminal color schemes based on your current directory, giving each project a distinct visual identity at a glance.

## Core Flow

1. User `cd`s into a directory.
2. Shell hook invokes `coltty apply`.
3. Coltty walks up the directory tree looking for `.coltty.toml` config files.
4. Nearest config wins, with inheritance from parent configs.
5. If no config is found anywhere, applies the global default scheme.
6. The resolved color scheme is sent to the terminal via the appropriate adapter (Ghostty first).

## Config File Format

### Global Config

Located at `~/.config/coltty/config.toml`. Defines named schemes and the default.

```toml
[default]
scheme = "calm"

[schemes.calm]
foreground = "#c0caf5"
background = "#1a1b26"
cursor = "#c0caf5"
palette = [
    "#15161e", "#f7768e", "#9ece6a", "#e0af68",
    "#7aa2f7", "#bb9af7", "#7dcfff", "#a9b1d6",
    "#414868", "#f7768e", "#9ece6a", "#e0af68",
    "#7aa2f7", "#bb9af7", "#7dcfff", "#c0caf5",
]

[schemes.danger]
foreground = "#f8f8f2"
background = "#3b0a0a"
cursor = "#ff5555"
palette = [
    "#282a36", "#ff5555", "#50fa7b", "#f1fa8c",
    "#bd93f9", "#ff79c6", "#8be9fd", "#f8f8f2",
    "#6272a4", "#ff6e6e", "#69ff94", "#ffffa5",
    "#d6acff", "#ff92df", "#a4ffff", "#ffffff",
]
```

### Per-Directory Config

A `.coltty.toml` file placed in any directory.

```toml
# Simple: reference a named scheme
scheme = "danger"
```

```toml
# Or reference a scheme with overrides
scheme = "calm"

[overrides]
background = "#1e2030"
```

### Resolution Order

Current directory, then parent, then grandparent, and so on up to `~/.config/coltty/config.toml` default. The first `.coltty.toml` found wins. If it references a named scheme, that scheme is looked up from the global config.

## Ghostty Adapter

Ghostty supports live config reloading. When Ghostty detects its config file has changed, it automatically applies the new colors.

### Strategy

1. Coltty writes the resolved color values to a Ghostty-compatible config fragment file at `~/.config/coltty/ghostty-colors`.
2. The user adds a one-time `config-file = /Users/<user>/.config/coltty/ghostty-colors` line to their Ghostty config.
3. On each `cd`, coltty regenerates that fragment file with the new palette.
4. Ghostty picks up the change automatically via its file-watching.

### Why File-Based Over OSC Escape Sequences

- Ghostty's native config format means colors survive new splits and tabs in the same session.
- No flickering or partial-apply issues.
- Clean separation: coltty never talks directly to the terminal, just writes a file.

### Generated Fragment

The file at `~/.config/coltty/ghostty-colors`:

```
foreground = #c0caf5
background = #1a1b26
cursor-color = #c0caf5
palette = 0=#15161e
palette = 1=#f7768e
palette = 2=#9ece6a
palette = 3=#e0af68
palette = 4=#7aa2f7
palette = 5=#bb9af7
palette = 6=#7dcfff
palette = 7=#a9b1d6
palette = 8=#414868
palette = 9=#f7768e
palette = 10=#9ece6a
palette = 11=#e0af68
palette = 12=#7aa2f7
palette = 13=#bb9af7
palette = 14=#7dcfff
palette = 15=#c0caf5
```

### Future Adapter Trait

```rust
trait TerminalAdapter {
    fn apply(&self, scheme: &ResolvedScheme) -> Result<()>;
    fn detect(&self) -> bool;
}
```

`detect()` checks if the terminal is active (e.g., `$TERM_PROGRAM == ghostty`), so coltty can auto-select the right adapter.

## Shell Integration

A thin shell hook that fires on every directory change. Installed by adding one line to `.zshrc` or `.bashrc`.

### Zsh

```zsh
eval "$(coltty init zsh)"
```

Expands to:

```zsh
coltty_chpwd() {
    coltty apply --quiet 2>/dev/null &!
}
chpwd_functions+=(coltty_chpwd)
```

### Bash

```bash
eval "$(coltty init bash)"
```

### Details

- Runs in the background (`&!` in zsh) so it never slows down the prompt.
- `--quiet` suppresses output unless there's an error.
- `coltty apply` is also available as a manual command for debugging or forcing a refresh.
- `coltty apply --dry-run` prints what it would do without changing anything.

## CLI Commands

- `coltty init <shell>` — print the shell hook code for the given shell
- `coltty apply` — manually apply the scheme for the current directory
- `coltty apply --quiet` — apply silently (used by the shell hook)
- `coltty apply --dry-run` — print the resolved scheme without applying
- `coltty show` — print the resolved scheme for the current directory, showing which config file matched and what colors were resolved
- `coltty schemes` — list all named schemes from the global config

## Project Structure

```
coltty/
├── Cargo.toml
├── src/
│   ├── main.rs          # CLI entry point (clap)
│   ├── config.rs         # TOML parsing, scheme resolution
│   ├── resolver.rs       # Directory walk + config merging
│   ├── adapter/
│   │   ├── mod.rs        # TerminalAdapter trait
│   │   └── ghostty.rs    # Ghostty adapter
│   └── shell.rs          # Shell hook generation
```

### Dependencies

- `clap` — CLI argument parsing
- `serde` + `toml` — config parsing
- `dirs` — XDG/home directory resolution

## Error Handling

- **Config parse errors:** Print a warning to stderr, fall back to default scheme. Never leave the terminal in a broken state.
- **Missing global config:** Silently use a hardcoded default scheme on first run, or run `coltty init` to generate one.
- **No adapter detected:** Warn once, do nothing.
- **Permission errors writing the fragment file:** Warn to stderr, no-op.

## Out of Scope

These are intentionally excluded from the initial version:

- No GUI or TUI for picking colors.
- No scheme importing from other tools (iTerm2 themes, base16, etc.).
- No per-tab or per-pane tracking.
- No daemon mode — just a fast binary invoked on each `cd`.
