# Coltty

Automatically switch terminal color schemes based on your current directory.

Coltty is a small Go CLI that hooks into your shell's `cd` command. When you change directories, it walks up the directory tree looking for `.coltty.toml` config files, resolves a color scheme, and applies it to your terminal instantly via OSC escape sequences. Each project gets a distinct visual identity at a glance.

Currently supports **Ghostty** (more terminals can be added via the adapter interface).

## Install

```bash
go build -o coltty .
cp coltty ~/.local/bin/   # or anywhere on your PATH
```

## Setup

### 1. Create a global config

Create `~/.config/coltty/config.toml` with your named schemes:

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

### 2. Add the shell hook

**Zsh** — add to `~/.zshrc`:

```zsh
eval "$(coltty init zsh)"
```

**Bash** — add to `~/.bashrc`:

```bash
eval "$(coltty init bash)"
```

### 3. Ghostty setup

Add this line to your Ghostty config (`~/.config/ghostty/config`):

```
config-file = /Users/<you>/.config/coltty/ghostty-colors
```

This ensures new Ghostty windows and tabs pick up the current scheme. Live color changes in existing terminals are handled via OSC escape sequences.

### 4. Tag your projects

Drop a `.coltty.toml` in any directory:

```toml
scheme = "danger"
```

Or use a scheme with overrides:

```toml
scheme = "calm"

[overrides]
background = "#1e2030"
```

Now `cd` into that directory and watch the colors change.

## Commands

| Command | Description |
|---------|-------------|
| `coltty init <shell>` | Print the shell hook code |
| `coltty apply` | Apply the scheme for the current directory |
| `coltty apply --quiet` | Apply silently (used by the shell hook) |
| `coltty apply --dry-run` | Print what would be applied without changing anything |
| `coltty show` | Show the resolved scheme and which config matched |
| `coltty schemes` | List all named schemes from the global config |

## How it works

1. You `cd` into a directory
2. The shell hook runs `coltty apply --quiet`
3. Coltty walks up from the current directory looking for `.coltty.toml`
4. The first one found wins — its scheme is looked up from the global config
5. If none is found, the global default scheme is used
6. Colors are applied via OSC escape sequences (immediate) and written to a Ghostty config fragment (for new windows/tabs)
