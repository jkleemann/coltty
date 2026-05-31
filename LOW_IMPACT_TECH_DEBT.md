# Low-Impact Technical Debt

A holding area for findings that are valid but not worth fixing now. Items here are candidates for future grilling sessions or can be pulled into the refinement plan if priorities shift.

---

## LD-001: Skip No-Op Applies (was P2.16)

**Date:** 2026-05-30  
**Original finding:** Rapidly cd-ing between directories spams OSC sequences. No optimization for "same scheme as last time."  
**Why deferred:** P0.2 (bash hook fix) eliminates the main source of spam. After that fix, a typical user only triggers `coltty apply` on actual directory changes, making this optimization marginal.  
**When to revisit:** If users report performance issues with many `cd` operations in scripts, or if we add a daemon mode.  
**Proposed approach:** Cache last-applied scheme name + directory in `$XDG_CACHE_HOME/coltty/last-scheme` (requires P0.3 XDG support). Skip `Apply()` if resolved scheme is identical.

---

## LD-002: Import Name Derivation Collision (was P1.13)

**Date:** 2026-05-30  
**Original finding:** "My Theme" and "my-theme" both become "my-theme" during import. `--append` silently overwrites.  
**Why deferred:** Low user impact. Most users import themes one at a time and review before appending.  
**When to revisit:** If we add bulk import (`coltty import --batch`) where collisions are more likely.  
**Proposed approach:** Before appending, check if derived name exists. If so, append `-2`, `-3`, etc. or prompt for a new name.

---

## LD-003: install.sh XDG-Aware Config Paths

**Date:** 2026-05-30  
**Original finding:** `install.sh` hardcodes shell config paths (`~/.zshrc`, `~/.bashrc`, `~/.config/fish/config.fish`). It does not respect `$ZDOTDIR` (zsh), `$BASH_ENV` (bash), or `$XDG_CONFIG_HOME` (fish).  
**Why deferred:** The current hardcoded paths work for 95%+ of users. Adding XDG awareness adds complexity to a simple install script.  
**When to revisit:** If users report that the hook was added to the wrong file due to custom config locations.  
**Proposed approaches:**
- **B:** Use `${XDG_CONFIG_HOME:-$HOME/.config}` for fish, `${ZDOTDIR:-$HOME}` for zsh.
- **C:** Query each shell for its actual config directory (e.g., `fish -c 'echo $__fish_config_dir'`).

---

*Add new items at the top. Include date, original finding, rationale for deferral, and revisit trigger.*
