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

# ─── Next steps ───────────────────────────────────────────────────────────────

cat <<'EOF'

  ✓ coltty installed successfully!

  Next steps:

  1. Add the shell hook to your shell profile:

     # zsh (~/.zshrc)
     eval "$(coltty init zsh)"

     # bash (~/.bashrc)
     eval "$(coltty init bash)"

  2. Create a global config at ~/.config/coltty/config.toml:

     [default]
     scheme = "calm"

     [schemes.calm]
     foreground = "#c0caf5"
     background = "#1a1b26"
     cursor = "#c0caf5"

  3. Tag a project directory:

     echo 'scheme = "calm"' > ~/projects/myapp/.coltty.toml

  4. cd into the directory and watch the colors change!

  See https://github.com/jkleemann/coltty for full documentation.

EOF
