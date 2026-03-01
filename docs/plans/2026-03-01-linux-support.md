# Part 3: Linux Terminal Support

## Context

Coltty currently supports terminals commonly found on macOS (Ghostty, iTerm2, Terminal.app) plus cross-platform terminals that support standard OSC sequences (Alacritty, Kitty, WezTerm, xterm, foot). Linux desktops have several popular terminals that aren't yet detected, even though most of them support the same OSC 10/11/12/4 escape sequences coltty already uses.

The goal is to add detection for Linux-native terminals so `coltty` works out of the box on typical Linux desktops without falling through to "no supported terminal detected."

## Research Findings

### OSC Support Matrix

Nearly every mainstream Linux terminal supports standard OSC 10/11/12/4 sequences for dynamic color changes. The key question per-terminal is **how to detect it**, not whether the color mechanism works.

| Terminal | Engine | OSC 4/10/11/12 | Detection Env Var | Notes |
|----------|--------|:-:|---|---|
| **GNOME Terminal** | VTE | Yes | `VTE_VERSION` | Most popular Linux terminal |
| **Tilix** | VTE | Yes | `TILIX_ID` | Also sets `VTE_VERSION` |
| **Terminator** | VTE | Yes | `TERMINATOR_UUID` | Also sets `VTE_VERSION` |
| **xfce4-terminal** | VTE | Yes | `VTE_VERSION` | No unique env var |
| **LXTerminal** | VTE | Yes | `VTE_VERSION` | No unique env var |
| **sakura** | VTE | Yes | `VTE_VERSION` | No unique env var |
| **Guake** | VTE | Yes | `VTE_VERSION` | Drop-down, no unique env var |
| **Konsole** | Qt | Yes | `KONSOLE_DBUS_SESSION` | KDE's default terminal |
| **Yakuake** | Konsole KPart | Yes | `KONSOLE_DBUS_SESSION` | KDE drop-down, same engine as Konsole |
| **st** | Custom | Yes | `TERM=st-256color` | Suckless terminal |
| **urxvt** | Custom | Yes | `TERM` contains `rxvt` | Also has proprietary OSC 708 for border |
| **Hyper** | xterm.js | Yes | `TERM_PROGRAM=Hyper` | Electron-based |
| **Tabby** | xterm.js | Yes | `TERM_PROGRAM=Tabby` | Electron-based (formerly Terminus) |
| **Terminology** | Custom | Yes | Hard to detect (`TERM=xterm`) | Enlightenment terminal |
| **cool-retro-term** | qtermwidget | **No** | Hard to detect | Not feasible to support |

### Key Insight: VTE Covers Most Linux Terminals

All GNOME/GTK-based terminals use the VTE (Virtual Terminal Emulator) library which handles OSC parsing at the library level. Detecting `VTE_VERSION` as a **catch-all** covers GNOME Terminal, Tilix, Terminator, xfce4-terminal, LXTerminal, sakura, Guake, MATE Terminal, elementary Terminal, and any other VTE-based terminal.

Note: `VTE_VERSION` is set by `/etc/profile.d/vte.sh` (or `vte-2.91.sh`) which runs in login shells. It may not be set in all environments, but it's the most reliable detection method available for VTE terminals.

## Approach

All new terminals use standard OSC sequences — the existing `OSCAdapter` pattern handles them with zero new color-change code. Each terminal only needs a constructor with a detection function, identical to the existing `NewAlacrittyAdapter()` / `NewFootAdapter()` pattern.

### New Adapters (all in `adapter/terminals.go`)

```go
// Konsole (also covers Yakuake which embeds Konsole's KPart)
func NewKonsoleAdapter() *OSCAdapter {
    return &OSCAdapter{
        TermName:   "konsole",
        DetectFunc: func() bool { return os.Getenv("KONSOLE_DBUS_SESSION") != "" },
    }
}

// Hyper (Electron-based, uses xterm.js)
func NewHyperAdapter() *OSCAdapter {
    return &OSCAdapter{
        TermName:   "hyper",
        DetectFunc: func() bool { return os.Getenv("TERM_PROGRAM") == "Hyper" },
    }
}

// Tabby (Electron-based, uses xterm.js, formerly Terminus)
func NewTabbyAdapter() *OSCAdapter {
    return &OSCAdapter{
        TermName:   "tabby",
        DetectFunc: func() bool { return os.Getenv("TERM_PROGRAM") == "Tabby" },
    }
}

// st (suckless terminal)
func NewStAdapter() *OSCAdapter {
    return &OSCAdapter{
        TermName: "st",
        DetectFunc: func() bool {
            term := os.Getenv("TERM")
            return term == "st-256color" || term == "st" ||
                   term == "st-meta-256color" || term == "st-meta"
        },
    }
}

// urxvt (rxvt-unicode)
func NewUrxvtAdapter() *OSCAdapter {
    return &OSCAdapter{
        TermName: "urxvt",
        DetectFunc: func() bool {
            term := os.Getenv("TERM")
            return strings.HasPrefix(term, "rxvt")
        },
    }
}

// VTE-based terminals (GNOME Terminal, Tilix, Terminator, xfce4-terminal, etc.)
// This is a catch-all that should be last in the adapter list.
func NewVTEAdapter() *OSCAdapter {
    return &OSCAdapter{
        TermName:   "vte",
        DetectFunc: func() bool { return os.Getenv("VTE_VERSION") != "" },
    }
}
```

### Detection Priority Order (updated `AllAdapters()`)

More specific detectors must come before generic ones. The VTE adapter is a catch-all and must come last among the new adapters (but before any future "unknown terminal" fallback).

```go
func AllAdapters() []TerminalAdapter {
    return []TerminalAdapter{
        // macOS-specific (most specific first)
        NewGhosttyAdapter(""),
        NewITermAdapter(),
        NewTerminalAppAdapter(),
        // Cross-platform (TERM_PROGRAM detection)
        NewAlacrittyAdapter(),
        NewKittyAdapter(),
        NewWezTermAdapter(),
        NewHyperAdapter(),
        NewTabbyAdapter(),
        // Linux-specific (env var detection)
        NewKonsoleAdapter(),     // KONSOLE_DBUS_SESSION
        // TERM-based detection (less specific)
        NewXtermAdapter(),       // XTERM_VERSION
        NewFootAdapter(),        // TERM=foot*
        NewStAdapter(),          // TERM=st*
        NewUrxvtAdapter(),       // TERM=rxvt*
        // VTE catch-all (must be last — matches any VTE terminal)
        NewVTEAdapter(),         // VTE_VERSION
    }
}
```

### Files to Modify

**`adapter/terminals.go`** — Add 6 new constructor functions:
- `NewKonsoleAdapter()` — detects via `KONSOLE_DBUS_SESSION`
- `NewHyperAdapter()` — detects via `TERM_PROGRAM=Hyper`
- `NewTabbyAdapter()` — detects via `TERM_PROGRAM=Tabby`
- `NewStAdapter()` — detects via `TERM=st*`
- `NewUrxvtAdapter()` — detects via `TERM=rxvt*`
- `NewVTEAdapter()` — detects via `VTE_VERSION` (catch-all for GNOME Terminal, Tilix, Terminator, xfce4-terminal, etc.)
- Add `"strings"` import for `strings.HasPrefix`

**`adapter/adapter.go`** — Update `AllAdapters()` to include the new adapters in correct priority order.

**`adapter/terminals_test.go`** — Add detection tests for each new adapter (env var set → detected, env var unset → not detected).

**`README.md`** — Update the "Supported Terminals" table to list new terminals.

### Not Supported (with rationale)

| Terminal | Reason |
|----------|--------|
| **cool-retro-term** | Does not support OSC color sequences (qtermwidget backend). No feasible workaround. |
| **Terminology** | OSC works but detection is unreliable — sets `TERM=xterm` with no unique env var. Would require process tree inspection which is fragile. Low market share. May work via VTE fallback if user sets `VTE_VERSION` manually, or via future generic TERM fallback. |

### Future Consideration: Generic `COLORTERM` Fallback

Many terminals set `COLORTERM=truecolor` or `COLORTERM=24bit`. A low-priority "generic OSC" adapter detecting on `COLORTERM` could serve as an absolute last-resort fallback for terminals we don't explicitly detect. This would be placed after the VTE adapter and would cover edge cases like Terminology. Not recommended for this initial implementation — better to be explicit about what we support.

## Verification

- `go test ./...` — all tests pass, including new detection tests
- On GNOME desktop: `VTE_VERSION` is set, `coltty apply` works in GNOME Terminal
- On KDE desktop: `KONSOLE_DBUS_SESSION` is set, `coltty apply` works in Konsole
- `coltty apply --dry-run` resolves correctly in each detected terminal
- README lists all newly supported terminals
