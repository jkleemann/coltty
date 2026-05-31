# Coltty Refinement Plan

Generated from comprehensive codebase review. All findings categorized by priority and organized into actionable work items.

---

## P0 — Critical (Fix Before Next Release)

### P0.1 Fix go.mod Go Version
**File:** `go.mod`  
**Problem:** Declares `go 1.25.7`, which does not exist. Breaks `go install` for users on stable Go toolchains.  
**Fix:** Change to the minimum actually required (likely `1.21` or `1.22`). Verify by testing `go install` on a clean machine.  
**Verification:** `go mod tidy && go build ./...`

### P0.2 Fix Bash Hook to Only Fire on Directory Changes
**File:** `shell.go`  
**Problem:** Uses `PROMPT_COMMAND`, which fires on every prompt render (including pressing Enter). This wastes CPU and may cause flicker. The zsh hook correctly uses `chpwd_functions`.  
**Fix:** Track `$PWD` in a variable and only run `coltty apply` when it changes:
```bash
__coltty_prompt_command() {
    if [[ "$__coltty_last_pwd" != "$PWD" ]]; then
        coltty apply --quiet 2>/dev/null
        __coltty_last_pwd="$PWD"
    fi
}
```
**Verification:** Run `bash -n` on generated hook. Test by pressing Enter in bash — no `coltty apply` should run.

### P0.3 Respect `$XDG_CONFIG_HOME`
**File:** `config.go:173`  
**Problem:** Hardcodes `~/.config/coltty/config.toml`, ignoring the XDG Base Directory spec. Users with custom `$XDG_CONFIG_HOME` have configs in the wrong place.  
**Fix:** Check `$XDG_CONFIG_HOME` first:
```go
func globalConfigPath() string {
    if globalConfigPathOverride != "" {
        return globalConfigPathOverride
    }
    configDir := os.Getenv("XDG_CONFIG_HOME")
    if configDir == "" {
        home, _ := os.UserHomeDir()
        configDir = filepath.Join(home, ".config")
    }
    return filepath.Join(configDir, "coltty", "config.toml")
}
```
**Verification:** Test with `XDG_CONFIG_HOME=/tmp/xdg go test ./...`.

### P0.4 Make Global Config Append Atomic
**File:** `cmd_import.go:174-182`  
**Problem:** `appendToGlobalConfig` reads config, modifies in memory, then writes back with `os.Create`. Concurrent `coltty import --append` calls can race and lose data.  
**Fix:** Write to a temp file in the same directory, then rename. Clean up temp file on error:
```go
tmpPath := configPath + ".tmp"
f, err := os.Create(tmpPath)
if err != nil { return err }
encErr := toml.NewEncoder(f).Encode(cfg)
f.Close()
if encErr != nil {
    os.Remove(tmpPath)
    return encErr
}
return os.Rename(tmpPath, configPath)
```
**Verification:** Write a test that spawns two parallel `appendToGlobalConfig` calls and assert no data loss.

### P0.5 Add Hex Color Validation Before Emitting OSC
**File:** `adapter/osc.go`, `adapter/color.go`  
**Problem:** OSC adapters blindly emit any string from config. Malformed colors like `"#gggggg"` or `"not-a-color"` are sent to the terminal, which silently ignores them. Users get no feedback.  
**Fix:** Extract a `ValidateHex(color string) error` helper (reuse logic from `HexToTerminalAppRGB`). Add a `Validate()` method on `ResolvedScheme` that checks all hex fields and returns a slice of validation errors. Call it in `Resolve()` before returning. In `applyCmd`, print warnings to stderr for each invalid color but still apply the scheme. Built-in schemes are hardcoded and trusted; only user configs get validated.  
**Verification:** Add test for `Resolve` with invalid hex in a scheme — should return the scheme plus validation warnings, not an error.

---

## P1 — Important (Should Fix Soon)

### P1.6 Archive or Update Stale Design Docs
**Files:** `docs/plans/2026-02-28-coltty-design.md`, `docs/plans/2026-03-01-linux-support.md`  
**Problem:** The design docs contain multiple inaccuracies vs. actual code:
- Shows `&!` (backgrounding) in zsh hook — removed in actual code
- Lists `adrg/xdg` as dependency — not in go.mod
- Says "No scheme importing" is out of scope — feature now exists
- Shows outdated project structure
**Fix:** Either (A) update docs to reflect reality, or (B) move to `docs/archive/` and add a header noting they are historical.  
**Verification:** Read through each doc and verify every code snippet matches current source.

### P1.7 Add Fish Shell Support
**File:** `shell.go`  
**Problem:** Fish is the default shell on many systems. `install.sh` silently skips fish users.  
**Fix:** Add fish hook using `function __coltty_cwd --on-variable PWD`:
```fish
function __coltty_cwd --on-variable PWD
    coltty apply --quiet 2>/dev/null
end
```
Update `ShellHook()` to accept `"fish"`, and update `install.sh` to detect fish and append to `~/.config/fish/config.fish` (hardcoded path for now; see `LOW_IMPACT_TECH_DEBT.md` for XDG-aware path discussion).  
**Verification:** `fish -n` on generated hook. Test in actual fish shell.

### P1.8 Fix Terminal.app Setup to Respect `terminal_app_profile` Override
**File:** `cmd_setup.go`  
**Problem:** `setup terminal-app` creates profiles using the scheme name. It never checks if the scheme has `terminal_app_profile` set. The `apply` command respects this override; `setup` does not. Inconsistent UX.  
**Fix:** In the setup loop, check `s.TerminalAppProfile` and use it as the profile name if set.  
**Verification:** Add test `TestSetupTerminalAppWithProfileOverride` — scheme with `terminal_app_profile` should create a profile with that name, not the scheme name.

### P1.9 Handle Empty Scheme Name with Overrides Gracefully
**File:** `config.go:231-243`  
**Problem:** If `.coltty.toml` has `[overrides]` but no `scheme =`, `dirCfg.Scheme` is `""`. The code skips lookup and applies overrides to an empty base `Scheme`. Result: only override fields are set, everything else is blank → broken terminal.  
**Fix:** When `dirCfg.Scheme == ""`, fall back to the global default scheme before applying overrides.  
**Verification:** Add test: `.coltty.toml` with only `[overrides]background = "#222"` should resolve to default scheme with overridden background.

### P1.10 Add Tests for Uncovered Functions
**Files:** `main.go`, `config.go`  
**Problem:** Several functions have zero test coverage:
- `toAdapterScheme()` — main.go:313
- `printScheme()` — main.go:303
- `formatInlineConfig()` — main.go:255
- `lookupScheme()` — main.go:242
- `applyOverrides()` — config.go:279
- `getDefaultScheme()` — config.go:264
- `LoadGlobalConfigFrom()` with malformed TOML
- `LoadDirConfig()` with malformed TOML
- `importFile()` with unknown format
**Fix:** Add table-driven tests for each. For error-path tests, feed invalid TOML / invalid formats and assert correct error messages.  
**Verification:** `go test -coverprofile=c.out ./... && go tool cover -func=c.out` — aim for >85% coverage.

### P1.11 Fix Test Fragility in main_test.go
**File:** `main_test.go`  
**Problem:** `executeCommand` mutates global `rootCmd` state. Tests that set flags (like `--inline`, `--append`) must manually reset them in `defer`. One missed reset breaks later tests.  
**Fix:** Create a `newTestRootCmd()` factory that builds a fresh command tree for each test, avoiding the global `rootCmd` singleton. Tests call `newTestRootCmd().Execute()` instead of `rootCmd.Execute()`. No production code changes needed. Remove all manual `defer` flag resets from individual tests.  
**Verification:** All tests pass with the new factory. No test depends on global `rootCmd` state.

### P1.12 Add Error Return to `OSCEmitter.Emit`
**File:** `adapter/osc.go:17`  
**Problem:** `Emit()` silently ignores write failures (`fmt.Fprint` returns error, discarded). In rare cases (e.g., stdout closed), the user gets no feedback.  
**Fix:** Change signature to `func (e *OSCEmitter) Emit(scheme *ResolvedScheme) error`. Return `fmt.Fprint` error. Update all callers.  
**Verification:** Add test with a writer that returns an error — `Emit` should propagate it.

### P1.13 Fix Import Name Derivation Collision
**File:** `cmd_import.go:61-67`  
**Problem:** "My Theme" and "my-theme" both become "my-theme". The `--append` silently overwrites existing schemes.  
**Fix:** Before appending, check if the derived name already exists and prompt (or append a suffix like `-2`) instead of silently overwriting.  
**Verification:** Test importing two files that derive to the same name — second should not clobber first without warning.

---

## P2 — Enhancements (Nice to Have)

### P2.14 Add `coltty unset` Command
**File:** new command in `main.go` or new file  
**Problem:** There is `coltty set` but no way to remove a per-directory config. Users must manually `rm .coltty.toml`.  
**Fix:** Add `coltty unset` that deletes `.coltty.toml` in the current directory and applies the parent/default scheme.  
**Verification:** Test that `unset` removes the file and `Resolve()` then finds parent config or default.

### P2.15 Add Windows Terminal Adapter
**File:** `adapter/terminals.go`, `adapter/adapter.go`  
**Problem:** Windows Terminal is not supported, even though it handles OSC 10/11/12/4. It sets `WT_SESSION`.  
**Fix:** One-line addition:
```go
func NewWindowsTerminalAdapter() *OSCAdapter {
    return &OSCAdapter{
        TermName:   "windows-terminal",
        DetectFunc: func() bool { return os.Getenv("WT_SESSION") != "" },
    }
}
```
Add to `AllAdapters()` and add test. Update README.  
**Verification:** Test detection with `WT_SESSION` set.

### P2.16 Resolve AGENTS.md `init()` Contradiction
**File:** `AGENTS.md`  
**Problem:** The convention doc says "Don't use init() functions" but the codebase uses `init()` in three files for cobra command registration.  
**Fix:** Either (A) remove the rule from AGENTS.md and document why `init()` is acceptable for cobra registration, or (B) refactor to explicit registration in `main()`. Cobra's idiomatic usage is `init()`, so prefer (A).  
**Verification:** AGENTS.md accurately reflects codebase conventions.

### P2.17 Add Generic `COLORTERM` Fallback Adapter
**File:** `adapter/terminals.go`, `adapter/adapter.go`  
**Problem:** Terminals not explicitly detected (e.g., Terminology) fall through to "no supported terminal detected." Many terminals set `COLORTERM=truecolor` or `24bit`.  
**Fix:** Add a lowest-priority adapter that detects on `COLORTERM` and emits standard OSC sequences. Label it `"generic"` in output.  
**Verification:** Test detection with `COLORTERM=truecolor` and no other env vars set.

### P2.18 Improve Ghostty Fragment Path Error Handling
**File:** `adapter/ghostty.go:23`  
**Problem:** `os.UserHomeDir()` error is ignored. If `$HOME` is unset, fragment path becomes relative.  
**Fix:** Handle the error in `Apply()`, not in the constructor (to avoid breaking `AllAdapters()` signature). If `FragmentPath` is relative and `os.UserHomeDir()` failed, return an error from `Apply()` instead of writing to the current directory.  
**Verification:** Test with `HOME=""` — `Apply()` should return error, not create fragment in current directory.

### P2.19 Prevent Terminal.app Adapter from Launching Terminal.app
**File:** `adapter/terminal_app.go`  
**Problem:** AppleScript `tell application "Terminal"` launches the app if not running. A user in iTerm2 with accidentally-set `TERM_PROGRAM=Apple_Terminal` will have Terminal.app spawned.  
**Fix:** Prefix the script with a running check. Note: `System Events` requires accessibility permissions, which may fail silently. A safer approach is to also verify `TERM_PROGRAM_VERSION` is set (Terminal.app sets it), reducing false positives from misconfigured env:
```applescript
tell application "System Events" to if not (exists process "Terminal") then return
```
**Verification:** Test with Terminal.app not running — script should be a no-op.

### P2.20 Disallow iTerm2 Preset + Individual Colors Together
**File:** `adapter/iterm2.go`  
**Problem:** If both `iterm_preset` and individual colors are configured, both are emitted. iTerm2 behavior is undefined — preset may override colors or vice versa.  
**Fix:** In `emitITermExtras`, if `iterm_preset` is set, skip emitting individual color extras. Or document the precedence.  
**Verification:** Test that scheme with both preset and tab color only emits the preset sequence.

### P2.21 Improve install.sh Duplicate Detection
**File:** `install.sh:149`  
**Problem:** Duplicate detection uses `grep -qF "coltty init"`. If a user manually edits the line, a duplicate gets added.  
**Fix:** Use a more unique sentinel:
```bash
if [ -f "$RC_FILE" ] && grep -qF "# coltty-auto-installed" "$RC_FILE"; then
```
And include that comment above the hook line.  
**Verification:** Run install.sh twice — second run should report "already present."

### P2.22 Use `$HOME` in README Ghostty Config Example
**File:** `README.md:286`  
**Problem:** Shows `/Users/<you>/.config/coltty/ghostty-colors` which requires manual editing.  
**Fix:** Change to `config-file = $HOME/.config/coltty/ghostty-colors` (Ghostty expands env vars).  
**Verification:** Confirm Ghostty accepts `$HOME` in `config-file` directive.

---

## Appendix A: Review Methodology

- All `.go` files read and analyzed
- All `_test.go` files read and coverage gaps identified
- All docs and shell scripts read
- `go build ./...`, `go vet ./...`, `go test ./...` all passed at time of review
- Findings categorized by: correctness, performance, UX, test coverage, documentation, security

## Appendix B: Quick Stats

| Metric | Value |
|--------|-------|
| Total Go files | 22 |
| Total test files | 11 |
| Lines of Go code | ~2,800 |
| Test coverage (estimated) | ~65% (happy path heavy) |
| Dependencies | 4 direct + 4 indirect |
| Built-in schemes | 8 |
| Supported terminals | 14 + tmux |
| Shells supported | 2 (zsh, bash) |

## Appendix C: Cross-References Between Items

| Items | Relationship |
|-------|-------------|
| P0.5 ↔ P1.12 | P0.5 adds hex validation at the call site; P1.12 changes `Emit` to return errors. Can be done in either order. |
| P0.2 ↔ (removed P2.16) | P0.2 fixes bash spam; the skip-no-op optimization was dropped as marginal after P0.2. See `LOW_IMPACT_TECH_DEBT.md`. |
| P1.8 ↔ P2.19 | Both involve iTerm2 extended colors behavior. Consider doing together. |
| P1.7 ↔ P2.21 | Both touch `install.sh` and shell integration. Good to batch. |

---

*Plan generated from comprehensive review. Each item includes file references, problem description, proposed fix, and verification steps. Work through P0 first, then P1, then P2.*
