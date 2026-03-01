#!/bin/sh
# install.sh — Install coltty
# https://github.com/jkleemann/coltty
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/jkleemann/coltty/main/install.sh | sh
#   # or
#   sh install.sh
#
# Environment variables:
#   INSTALL_DIR  — override install directory (default: $GOPATH/bin or ~/.local/bin)

set -e

REPO="github.com/jkleemann/coltty"
REPO_URL="https://github.com/jkleemann/coltty.git"

usage() {
    cat <<'EOF'
Usage: install.sh [--help]

Install coltty — automatic terminal color scheme switching.

Methods (tried in order):
  1. go install (if Go >= 1.21 is available)
  2. git clone + go build (fallback)

Environment variables:
  INSTALL_DIR   Override install directory
                Default: $GOPATH/bin (go install) or ~/.local/bin (manual build)

Examples:
  curl -fsSL https://raw.githubusercontent.com/jkleemann/coltty/main/install.sh | sh
  INSTALL_DIR=/usr/local/bin sh install.sh
EOF
}

info() {
    printf '  %s\n' "$@"
}

warn() {
    printf '  warning: %s\n' "$@" >&2
}

err() {
    printf '  error: %s\n' "$@" >&2
    exit 1
}

# ─── Parse flags ──────────────────────────────────────────────────────────────

for arg in "$@"; do
    case "$arg" in
        --help|-h)
            usage
            exit 0
            ;;
        *)
            printf 'Unknown option: %s\n\n' "$arg" >&2
            usage >&2
            exit 1
            ;;
    esac
done

# ─── Detect platform ─────────────────────────────────────────────────────────

OS="$(uname -s)"
ARCH="$(uname -m)"
info "Platform: ${OS}/${ARCH}"

# ─── Check for Go ────────────────────────────────────────────────────────────

HAS_GO=false
if command -v go >/dev/null 2>&1; then
    GO_VERSION="$(go version | awk '{print $3}' | sed 's/go//')"
    info "Found Go ${GO_VERSION}"
    HAS_GO=true
else
    warn "Go not found. Coltty requires Go to install."
    warn "Install Go from https://go.dev/dl/ and re-run this script."
    exit 1
fi

# ─── Install via go install ──────────────────────────────────────────────────

info ""
info "Installing coltty..."

if [ -n "$INSTALL_DIR" ]; then
    info "Using INSTALL_DIR=${INSTALL_DIR}"
    GOBIN="$INSTALL_DIR" go install "${REPO}@latest"
    BINARY="${INSTALL_DIR}/coltty"
else
    go install "${REPO}@latest"
    # Determine where go install put the binary
    if [ -n "$GOBIN" ]; then
        BINARY="${GOBIN}/coltty"
    elif [ -n "$GOPATH" ]; then
        BINARY="${GOPATH}/bin/coltty"
    else
        BINARY="$(go env GOPATH)/bin/coltty"
    fi
fi

# ─── Verify binary ───────────────────────────────────────────────────────────

if [ ! -f "$BINARY" ]; then
    warn "Binary not found at ${BINARY}"
    warn "go install may have placed it elsewhere. Check 'go env GOPATH'."
    exit 1
fi

info "Installed: ${BINARY}"

# Check if binary is on PATH
if ! command -v coltty >/dev/null 2>&1; then
    warn ""
    warn "coltty is not on your PATH."
    BIN_DIR="$(dirname "$BINARY")"
    warn "Add this to your shell profile:"
    warn "  export PATH=\"${BIN_DIR}:\$PATH\""
    warn ""
fi

# ─── Add shell hook ─────────────────────────────────────────────────────────

HOOK_LINE='eval "$(coltty init zsh)"'
HOOK_LINE_BASH='eval "$(coltty init bash)"'

# Detect current shell and pick the right rc file.
CURRENT_SHELL="$(basename "${SHELL:-/bin/sh}")"
case "$CURRENT_SHELL" in
    zsh)
        RC_FILE="$HOME/.zshrc"
        HOOK="$HOOK_LINE"
        ;;
    bash)
        RC_FILE="$HOME/.bashrc"
        HOOK="$HOOK_LINE_BASH"
        ;;
    *)
        RC_FILE=""
        ;;
esac

if [ -n "$RC_FILE" ]; then
    if [ -f "$RC_FILE" ] && grep -qF "coltty init" "$RC_FILE"; then
        info "Shell hook already present in ${RC_FILE}"
    else
        printf '\n# coltty — automatic terminal color schemes\n%s\n' "$HOOK" >> "$RC_FILE"
        info "Added shell hook to ${RC_FILE}"
    fi
else
    warn "Could not detect shell (got: ${CURRENT_SHELL})."
    warn "Manually add the hook to your shell profile:"
    warn "  zsh:  eval \"\$(coltty init zsh)\"   → ~/.zshrc"
    warn "  bash: eval \"\$(coltty init bash)\"  → ~/.bashrc"
fi

# ─── Next steps ───────────────────────────────────────────────────────────────

cat <<'EOF'

  ✓ coltty installed successfully!

  Next steps:

  1. Tag a project directory:

     echo 'scheme = "dracula"' > ~/projects/myapp/.coltty.toml

  2. Open a new shell (or run: source ~/.zshrc) and cd into the directory!

  See https://github.com/jkleemann/coltty for full documentation.

EOF
