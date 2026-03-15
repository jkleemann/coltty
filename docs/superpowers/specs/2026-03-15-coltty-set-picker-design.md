# `coltty set` Bubble Tea Picker Design

## Summary

Replace the current custom interactive `coltty set` picker implementation with a Bubble Tea based TUI while preserving the full approved feature scope:

- default interactive picker on `coltty set`
- direct `coltty set <scheme>` path unchanged
- visible fuzzy filter
- favorites and `all` / `favorites` toggle
- usage badges from scanning `~`
- initial selection from named scheme or closest inferred theme
- transient live preview on selection change
- `Enter` persists `.coltty.toml`
- `Esc` clears filter first, then restores original terminal colors and exits
- richer semantic preview styling inspired by Ghostty's own theme preview language

The current custom picker is not a stable foundation. It redraws plain text manually, has no real pane layout, and mixes preview side effects into the event loop. Bubble Tea is the replacement architecture, not an optional enhancement.

## Goals

- Make `coltty set` behave like a real TUI, with stable panes and predictable redraws.
- Preserve the existing approved feature set without regressing interaction behavior.
- Keep preview changes reversible until the user confirms.
- Separate in-app preview rendering from real terminal apply/restore side effects.
- Improve maintainability by using standard TUI primitives instead of custom raw terminal control.
- Make the preview pane visually informative rather than flat, with semantic syntax accents that show off a theme's palette.

## Non-Goals

- Changing the semantics of `coltty set <scheme>`.
- Reducing scope to a minimal picker first and adding features back later.
- Reordering the main list based on usage counts.
- Making home-directory usage scanning a hard dependency for opening the picker.

## CLI Behavior

- `coltty set` opens the Bubble Tea picker.
- `coltty set <scheme>` keeps the current direct-write behavior.
- In picker mode:
  - arrow keys move selection
  - typing updates a visible fuzzy filter field
  - `Backspace` edits the filter
  - `Enter` writes `.coltty.toml` and applies the chosen theme
  - `Esc` clears the active filter first; if the filter is empty, it cancels and restores the original colors
  - `f` toggles favorite status for the selected theme
  - `Tab` switches between `all` and `favorites`

## User Experience

### Layout

The picker uses a full-screen Bubble Tea interface with two panes:

- Left pane:
  - visible filter input
  - current mode indicator
  - scrollable theme list
  - favorite marker
  - usage badge such as `used in 7 dirs`
  - stable active selection highlight
- Right pane:
  - one integrated preview terminal, not separate cards
  - palette strip
  - large Zig code sample with varied syntax highlighting
  - smaller `less` sample with semantic accents
  - smaller markdown sample with lighter semantic accents
  - short help or status line

The key requirement is stable composition. Content must remain anchored in panes instead of drifting or wrapping across the terminal unpredictably.

### Preview Styling

The right pane should not render as mostly plain text. It should use semantic color roles derived from the currently selected theme:

- Zig:
  - stronger accents for keywords, function names, builtins, strings, and important punctuation or headings
- `less` block:
  - semantic accents for section headers, command lines, and emphasized values
- markdown block:
  - lighter accents for headings, bullets, and inline-code-like fragments

The target is closer to Ghostty's own preview feel: enough color structure to communicate a theme's character, but still readable and controlled.

### Preview Semantics

- Moving selection applies a transient real terminal preview and updates the right pane view.
- Preview changes remain transient until `Enter`.
- `Enter` persists and leaves the final theme applied.
- `Esc` restores the colors active before launching the picker.

### Initial Selection

- If the current directory config resolves to a named scheme, select that scheme initially.
- If the current directory uses inline overrides, infer the closest known theme and select it as an approximate match.

## Data Model

The picker model should track:

- all available schemes from built-in and user-defined sources
- filter input state
- filtered and ranked results
- current selection
- current mode: `all` or `favorites`
- favorites set
- usage metadata keyed by scheme name
- resolved initial scheme or inferred closest match
- original terminal state needed for restore on cancel
- any warning or status message shown in the UI
- semantic preview color roles derived from the selected theme

## Favorites

- Favorites are stored separately from user theme definitions in `~/.config/coltty/favorites.toml`.
- The file is UI state, not theme definition.
- Favorites can be toggled with `f`.
- `Tab` switches between all themes and favorites-only view.

## Filtering

- Filtering is visible and discoverable, not hidden.
- Matching is case-insensitive fuzzy matching on scheme names.
- Ranking should prefer exact prefix matches first, then stronger fuzzy matches, then weaker matches.
- If the current selection is filtered out, move selection to the top remaining match.
- If there are no matches, the UI should show that state clearly and keep the current preview stable until a valid selection exists again.

## Current Theme Resolution

### Named Scheme

If `.coltty.toml` refers to a named scheme, resolve it directly from the same lookup logic used by the existing direct `set` path.

### Inline Overrides

If `.coltty.toml` uses inline colors, compute the closest known theme and initialize the picker selection to that theme. This is approximate and should be surfaced in picker state as inferred rather than exact.

## Usage Metadata

- Scan under `~` for `.coltty.toml` usage data.
- Usage counts are secondary metadata only and are displayed as badges.
- Usage data must not change primary ordering logic.
- If the scan fails or is unavailable, the picker still opens without badges.

## Architecture

Recommended approach: Bubble Tea with `bubbles` and `lipgloss`.

### Dependencies

Add:

- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/bubbles`
- `github.com/charmbracelet/lipgloss`

These are justified because the current problem is primarily one of terminal UI architecture, not business logic.

### Shared Backend

Keep and reuse the existing backend helpers for:

- scheme enumeration
- scheme lookup
- current directory config resolution
- inline override nearest-match inference
- `.coltty.toml` persistence
- final adapter application
- favorites persistence
- usage scanning

The direct and interactive `set` paths should continue to share this backend.

### Bubble Tea Picker Layer

Replace the current custom picker runtime with a Bubble Tea model that handles:

- window sizing
- filter input widget
- list state and selection
- favorites and view-mode toggling
- preview view rendering
- help and status text
- message-driven preview side effects
- cancel/confirm exit paths
- semantic preview styling derived from the selected scheme

### Preview Styling Helper

Add a small preview-style helper that maps a selected scheme into reusable semantic roles, for example:

- base text
- muted text
- keyword
- function or symbol accent
- string
- heading
- bullet or marker

The renderer should use these roles consistently instead of hardcoding ad hoc colors into each sample line.

### Side-Effect Boundary

Keep preview application and restore logic behind a narrow interface so the Bubble Tea update loop remains testable. The model should emit intent, while a small integration layer performs:

- transient apply on selection change
- restore on cancel
- final apply on confirm
- persistence of favorites and `.coltty.toml`

### Removal of Custom Picker Path

The manual raw-terminal code path should be removed once the Bubble Tea picker is in place. Reusing the current plain-text renderer or custom ANSI layout path would preserve the existing root cause.

## Terminal Lifecycle

Bubble Tea should own terminal setup and teardown rather than the current custom `stty raw` path.

Requirements:

- enter alternate screen mode
- handle resize cleanly
- hide and restore cursor correctly
- restore original colors on cancel
- leave the confirmed theme applied on `Enter`

If interactive terminal features are unavailable, fail clearly and direct the user to `coltty set <scheme>`.

## Error Handling

- Failure to load usage metadata should degrade gracefully.
- Failure to load favorites should warn and continue with an empty favorites set.
- Failure to save favorites should surface a warning without corrupting picker state.
- Failure to apply transient preview should surface as an in-app warning if possible; the picker should remain usable.
- Failure to persist the final choice should leave the original config unchanged.
- Failure to start the Bubble Tea program should fail clearly and point the user to the direct `set <scheme>` fallback.

## Testing

Prioritize tests around model behavior and integration boundaries:

- fuzzy matching and ranking
- filter editing and no-match handling
- selection movement on filtered result sets
- favorites toggle and persistence
- view switching between `all` and `favorites`
- current named-scheme resolution
- nearest-known-theme inference for inline overrides
- transient preview apply/restore flow
- confirm persisting `.coltty.toml` and applying the selected theme
- Bubble Tea view output for key layout markers and pane content
- preview styling coverage for Zig, `less`, and markdown sections
- non-interactive fallback if the TUI cannot start

Avoid overfitting tests to exact ANSI output. Prefer model-state assertions and high-signal view checks.

## Workflow Note

User requested an agent guidance rule at `~/AGENTS.md`:

- when a task involves building or replacing a terminal UI, explicitly ask whether the project should use Bubble Tea components
- recommend Bubble Tea by default for Go TUI work unless there is a concrete reason not to

At the time of writing this spec, `/Users/jkleemann/AGENTS.md` does not exist. Creating or updating that file should be included in the implementation planning scope for this change.

## Rollout Notes

- Keep the non-interactive path unchanged to avoid breaking scripts.
- Preserve full feature scope in the Bubble Tea rewrite; do not cut favorites, usage badges, or inferred initial selection from v1.
- Treat the current custom picker implementation as disposable once the Bubble Tea version is working.
