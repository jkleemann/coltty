# `coltty set` Interactive Picker Design

## Summary

Replace the current argument-only `coltty set <scheme>` flow with a default interactive picker when `coltty set` is invoked without arguments. The picker presents a Ghostty-style two-pane interface: a scrollable theme list on the left and a live preview terminal on the right. Arrow-key navigation updates the in-app preview only. Pressing `Enter` persists the selected theme to `.coltty.toml` and applies it. Pressing `Esc` cancels and restores the terminal colors that were active before entering the picker.

The existing `coltty set <scheme>` command remains available as the non-interactive path for scripts and direct usage.

## Goals

- Make `coltty set` discoverable and pleasant for interactive theme selection.
- Provide a richer preview than the current `coltty schemes` output.
- Keep preview changes reversible until the user confirms.
- Support large theme collections with visible fuzzy filtering.
- Add lightweight preference features through favorites and usage badges.

## Non-Goals

- Changing the semantics of `coltty set <scheme>`.
- Reordering the main list based on usage counts.
- Building a general-purpose theme browser outside the terminal picker.
- Making home-directory usage scanning a hard dependency for opening the picker.

## CLI Behavior

- `coltty set` opens the interactive picker.
- `coltty set <scheme>` keeps the current direct-write behavior.
- In picker mode:
  - `Up`/`Down` moves selection.
  - Typing updates a visible fuzzy filter query.
  - `Backspace` edits the query.
  - `Enter` writes `.coltty.toml` and applies the chosen theme.
  - `Esc` clears the active filter first; if the filter is empty, it cancels and restores the original colors.
  - `f` toggles favorite status for the selected theme.
  - `Tab` switches the left pane between `all` and `favorites`.

## User Experience

### Layout

The picker uses a full-screen terminal UI with two panes:

- Left pane:
  - visible `Filter:` line
  - scrollable list of scheme names
  - favorite marker
  - optional usage badge such as `used in 7 dirs`
  - active selection highlight
- Right pane:
  - one integrated preview terminal, not separate cards
  - palette strip
  - large Zig code sample with varied syntax highlighting
  - smaller `less` sample
  - smaller markdown sample
  - brief shell/status lines that reinforce preview behavior

### Preview Semantics

- Moving selection repaints only the in-app preview.
- Preview changes are transient.
- `Enter` persists and then applies the final choice to the real terminal.
- `Esc` restores the colors active before launching the picker.

### Initial Selection

- If the current directory config resolves to a named scheme, select that scheme initially.
- If the current directory uses inline overrides, infer the closest known theme and select it as an approximate match.

## Data Model

The picker state should track:

- complete scheme list from built-in and user-defined schemes
- current filter query
- filtered match list and ranking
- current selection index
- current view mode: `all` or `favorites`
- favorites set
- usage metadata keyed by scheme name
- initial resolved scheme or inferred closest match
- snapshot needed to restore terminal state and colors on cancel

## Favorites

- Favorites are stored separately from user theme definitions in `~/.config/coltty/favorites.toml`.
- The file is UI state, not configuration for scheme definitions.
- Favorites can be toggled from the picker with `f`.
- `Tab` switches between all schemes and favorites-only view.

## Filtering

- Filtering is visible and discoverable, not hidden.
- Matching is case-insensitive fuzzy matching on scheme names.
- Ranking should prefer exact prefix matches first, then stronger fuzzy matches, then weaker matches.
- If the current selection is filtered out, move selection to the top remaining match.
- When there are no matches, keep the current preview stable until a valid selection exists again.

## Current Theme Resolution

### Named Scheme

If `.coltty.toml` refers to a named scheme, resolve it directly from the same lookup logic used by the existing `set` command.

### Inline Overrides

If `.coltty.toml` uses inline colors, compute the closest known theme and initialize the picker selection to that theme. This is approximate and should be surfaced in the picker state as inferred rather than exact.

## Usage Metadata

- Scan under `~` for `.coltty.toml` usage data.
- Usage counts are secondary metadata only and are displayed as badges.
- Usage data must not change primary ordering logic.
- If the scan fails, is too slow, or is unavailable, the picker still opens without badges.

## Architecture

Recommended approach: build a thin full-screen TUI over reusable command logic.

### Shared backend extraction

Refactor the current `set` implementation in `main.go` into reusable functions for:

- scheme enumeration
- scheme lookup
- current directory config resolution
- inline override nearest-match inference
- `.coltty.toml` persistence
- final adapter application

This preserves one backend for both `coltty set <scheme>` and the interactive picker.

### Picker layer

Add a picker-focused unit that handles:

- terminal lifecycle
- input handling
- fuzzy filtering
- rendering
- favorites load/save
- usage scan and badge attachment
- transient preview application
- terminal restore on cancel or exit

Keep packages flat unless a new package boundary is justified by cohesion.

## Terminal Lifecycle

The picker should:

- enter alternate screen mode
- hide and restore the cursor
- repeatedly render without corrupting the user shell
- restore terminal state on normal exit and best-effort interrupt paths
- restore original colors on cancel

If the terminal environment does not support required interactive behavior, fail clearly and direct the user to `coltty set <scheme>`.

## Error Handling

- Failure to load usage metadata should degrade gracefully.
- Failure to load favorites should warn and continue with an empty favorites set.
- Failure to save favorites should surface a warning without corrupting picker state.
- Failure to apply transient preview should surface as an in-picker warning if possible; the picker should remain usable.
- Failure to persist the final choice should leave the original config unchanged.

## Testing

Prioritize tests around state transitions and shared backend behavior:

- fuzzy matching and ranking
- filter editing and no-match handling
- selection movement on filtered result sets
- favorites file load/save and toggle behavior
- view switching between `all` and `favorites`
- current named-scheme resolution
- nearest-known-theme inference for inline overrides
- cancel restoring previous colors
- confirm persisting `.coltty.toml` and applying the selected theme

UI rendering tests can stay lighter and focus on high-signal snapshots or specific string expectations. The main behavioral coverage should live below the renderer.

## Rollout Notes

- Keep the non-interactive path unchanged to avoid breaking scripts.
- Treat favorites as v1 scope.
- Treat usage badges as best-effort metadata in v1.
- Do not make usage-based ordering part of the first implementation.
