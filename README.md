# Coltty

Automatically switch terminal color schemes based on your current directory.

Coltty is a small Go CLI that hooks into your shell's `cd` command. When you change directories, it walks up the directory tree looking for `.coltty.toml` config files, resolves a color scheme, and applies it to your terminal instantly. Each project gets a distinct visual identity at a glance.

## Supported Terminals

| Terminal | Mechanism | Extra Setup | Extended Colors |
|----------|-----------|-------------|-----------------|
| [Ghostty](https://ghostty.org) | OSC + config fragment | One-time `config-file` line | — |
| [iTerm2](https://iterm2.com) | OSC + proprietary OSC 1337 | — | tab, bold, selection, presets |
| [Terminal.app](https://support.apple.com/guide/terminal) | AppleScript profile switching | Pre-create profiles | — |
| [Alacritty](https://alacritty.org) | OSC | — | — |
| [Kitty](https://sw.kovidgoyal.net/kitty/) | OSC | — | — |
| [WezTerm](https://wezfurlong.org/wezterm/) | OSC | — | — |
| [Hyper](https://hyper.is) | OSC | — | — |
| [Tabby](https://tabby.sh) | OSC | — | — |
| [Konsole](https://konsole.kde.org) | OSC | — | — |
| [GNOME Terminal](https://wiki.gnome.org/Apps/Terminal) | OSC | — | — |
| VTE-based (Tilix, Terminator, xfce4-terminal, ...) | OSC | — | — |
| [st](https://st.suckless.org) | OSC | — | — |
| [urxvt](http://software.schmorp.de/pkg/rxvt-unicode.html) | OSC | — | — |
| [xterm](https://invisible-island.net/xterm/) | OSC | — | — |
| [foot](https://codeberg.org/dnkl/foot) | OSC | — | — |
| tmux | Automatic DCS passthrough | — | — |
| GNU Screen | **Not supported** | — | — |

## Install

### Script

```bash
curl -fsSL https://raw.githubusercontent.com/jkleemann/coltty/main/install.sh | sh
```

### go install

```bash
go install github.com/jkleemann/coltty@latest
```

### Build from source

```bash
git clone https://github.com/jkleemann/coltty.git
cd coltty
go build -o coltty .
cp coltty ~/.local/bin/   # or anywhere on your PATH
```

## Quick Start

### 1. Add the shell hook

**Zsh** — add to `~/.zshrc`:

```zsh
eval "$(coltty init zsh)"
```

**Bash** — add to `~/.bashrc`:

```bash
eval "$(coltty init bash)"
```

### 2. Tag a directory

Coltty ships with 8 built-in color schemes — no config file needed. Drop a `.coltty.toml` in any project:

```toml
scheme = "dracula"
```

Or use a scheme with overrides:

```toml
scheme = "nord"

[overrides]
background = "#1e2030"
```

### 3. Try it

```bash
cd ~/projects/my-project
# Terminal colors change instantly
```

### 4. (Optional) Create a global config

To define custom schemes or set a default, create `~/.config/coltty/config.toml`:

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
```

User-defined schemes with the same name as a built-in will override it.

## Built-in Schemes

Coltty ships with 8 built-in color schemes available without any config file. All meet WCAG AA contrast requirements (>=4.5:1 fg/bg ratio).

| Name | Aesthetic | Background | Foreground |
|------|-----------|------------|------------|
| `gruvbox` | Warm retro | `#282828` | `#ebdbb2` |
| `nord` | Cool arctic | `#2e3440` | `#d8dee9` |
| `dracula` | Purple-tinted | `#282a36` | `#f8f8f2` |
| `solarized-dark` | Teal classic | `#002b36` | `#839496` |
| `catppuccin` | Pastel dark | `#1e1e2e` | `#cdd6f4` |
| `one-dark` | Atom-style | `#282c34` | `#abb2bf` |
| `rose-pine` | Muted floral | `#191724` | `#e0def4` |
| `kanagawa` | Japanese ink | `#1f1f28` | `#dcd7ba` |

Each includes full 16-color ANSI palette and cursor color. Use `coltty schemes` to see all available schemes.

## Configuration

### Global config (`~/.config/coltty/config.toml`)

Full example with all fields:

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

# Extended colors (iTerm2 only — silently ignored by other terminals)
bold = "#e0af68"
tab = "#1a1b26"
selection_foreground = "#c0caf5"
selection_background = "#33467c"

# Switch to an iTerm2 preset by name instead of setting individual colors
iterm_preset = "Solarized Dark"

# Switch Terminal.app to this profile name (must already exist in Terminal.app)
terminal_app_profile = "My Dark Theme"
```

### Per-directory config (`.coltty.toml`)

Placed in any directory. Coltty walks up from the current directory and uses the first one found.

```toml
scheme = "danger"

[overrides]
background = "#1e2030"
```

### Configuration Reference

| Field | Type | Description | Used by |
|-------|------|-------------|---------|
| `foreground` | string | Foreground color (hex) | All OSC terminals, Ghostty |
| `background` | string | Background color (hex) | All OSC terminals, Ghostty |
| `cursor` | string | Cursor color (hex) | All OSC terminals, Ghostty |
| `palette` | string[] | 16 ANSI colors (hex) | All OSC terminals, Ghostty |
| `bold` | string | Bold text color (hex) | iTerm2 |
| `tab` | string | Tab bar color (hex) | iTerm2 |
| `selection_foreground` | string | Selection foreground (hex) | iTerm2 |
| `selection_background` | string | Selection background (hex) | iTerm2 |
| `iterm_preset` | string | iTerm2 color preset name | iTerm2 |
| `terminal_app_profile` | string | Terminal.app profile name | Terminal.app |

## Terminal Setup

### Ghostty

Ghostty is unique: in addition to live OSC color changes, coltty writes a config fragment file so that **new windows and tabs** also pick up the current scheme. Add this line to your Ghostty config (`~/.config/ghostty/config`):

```
config-file = /Users/<you>/.config/coltty/ghostty-colors
```

Replace `<you>` with your username, or use the full path from `echo ~/.config/coltty/ghostty-colors`.

The fragment is written to `~/.config/coltty/ghostty-colors` and Ghostty watches it for changes. Existing terminals get instant color changes via OSC; the fragment handles persistence for new terminals.

### iTerm2

Works out of the box — no extra setup needed.

iTerm2 supports extended color fields beyond the standard foreground/background/cursor/palette:

- `tab` — tab bar color
- `bold` — bold text color
- `selection_foreground` / `selection_background` — selection colors
- `iterm_preset` — switch to a named iTerm2 color preset

These fields are set via iTerm2's proprietary OSC 1337 sequences and are silently ignored by other terminals.

### Terminal.app

Terminal.app does not support OSC color-setting sequences. Instead, coltty switches to a named **settings profile** via AppleScript.

**Setup**: create Terminal.app profiles (Terminal > Settings > Profiles) matching your scheme names. For example, if your scheme is named `calm`, create a Terminal.app profile called "calm".

To use a different profile name than the scheme name, set `terminal_app_profile` in your scheme:

```toml
[schemes.calm]
foreground = "#c0caf5"
background = "#1a1b26"
cursor = "#c0caf5"
terminal_app_profile = "My Custom Profile"
```

### Alacritty / Kitty / WezTerm / xterm / foot

No extra setup needed. Colors are applied instantly via standard OSC escape sequences (OSC 10, 11, 12, 4). Just install coltty and add the shell hook.

### tmux

Transparent. If the `TMUX` environment variable is set, coltty automatically wraps all OSC sequences in DCS passthrough so they reach the outer terminal. No configuration needed.

### GNU Screen

Not supported. GNU Screen does not pass through OSC escape sequences. Coltty prints a warning if it detects a Screen session.

## Commands

| Command | Description |
|---------|-------------|
| `coltty init <shell>` | Print the shell hook code (`zsh` or `bash`) |
| `coltty apply` | Apply the scheme for the current directory |
| `coltty apply --quiet` | Apply silently (used by the shell hook) |
| `coltty apply --dry-run` | Print what would be applied without changing anything |
| `coltty show` | Show the resolved scheme and which config matched |
| `coltty schemes` | List all available schemes (built-in and user-defined) |

## How It Works

```
 cd ~/projects/myapp
        │
        ▼
 Shell hook runs: coltty apply --quiet
        │
        ▼
 Walk up directory tree looking for .coltty.toml
 ~/projects/myapp/.coltty.toml  ← found! scheme = "danger"
        │
        ▼
 Look up "danger" in ~/.config/coltty/config.toml
 Fall back to built-in schemes if not in user config
 Apply any [overrides] from the per-directory config
        │
        ▼
 Detect terminal (Ghostty? iTerm2? Terminal.app? Alacritty? ...)
        │
        ├─ Ghostty:       OSC sequences + write config fragment
        ├─ iTerm2:        OSC sequences + OSC 1337 extensions
        ├─ Terminal.app:  AppleScript profile switch
        ├─ Other OSC:     OSC 10/11/12/4 sequences
        └─ tmux:          Wrap in DCS passthrough, then above
```

If no `.coltty.toml` is found walking up, the global default scheme is applied.
