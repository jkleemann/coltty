package adapter

import (
	"os"
	"strings"
)

// NewAlacrittyAdapter returns an OSCAdapter for Alacritty.
func NewAlacrittyAdapter() *OSCAdapter {
	return &OSCAdapter{
		TermName:   "alacritty",
		DetectFunc: func() bool { return os.Getenv("TERM_PROGRAM") == "Alacritty" },
	}
}

// NewKittyAdapter returns an OSCAdapter for Kitty.
func NewKittyAdapter() *OSCAdapter {
	return &OSCAdapter{
		TermName:   "kitty",
		DetectFunc: func() bool { return os.Getenv("TERM_PROGRAM") == "kitty" },
	}
}

// NewWezTermAdapter returns an OSCAdapter for WezTerm.
func NewWezTermAdapter() *OSCAdapter {
	return &OSCAdapter{
		TermName:   "wezterm",
		DetectFunc: func() bool { return os.Getenv("TERM_PROGRAM") == "WezTerm" },
	}
}

// NewXtermAdapter returns an OSCAdapter for xterm.
func NewXtermAdapter() *OSCAdapter {
	return &OSCAdapter{
		TermName:   "xterm",
		DetectFunc: func() bool { return os.Getenv("XTERM_VERSION") != "" },
	}
}

// NewFootAdapter returns an OSCAdapter for foot.
func NewFootAdapter() *OSCAdapter {
	return &OSCAdapter{
		TermName: "foot",
		DetectFunc: func() bool {
			term := os.Getenv("TERM")
			return term == "foot" || term == "foot-extra"
		},
	}
}

// NewKonsoleAdapter returns an OSCAdapter for Konsole (also covers Yakuake).
func NewKonsoleAdapter() *OSCAdapter {
	return &OSCAdapter{
		TermName:   "konsole",
		DetectFunc: func() bool { return os.Getenv("KONSOLE_DBUS_SESSION") != "" },
	}
}

// NewHyperAdapter returns an OSCAdapter for Hyper.
func NewHyperAdapter() *OSCAdapter {
	return &OSCAdapter{
		TermName:   "hyper",
		DetectFunc: func() bool { return os.Getenv("TERM_PROGRAM") == "Hyper" },
	}
}

// NewTabbyAdapter returns an OSCAdapter for Tabby (formerly Terminus).
func NewTabbyAdapter() *OSCAdapter {
	return &OSCAdapter{
		TermName:   "tabby",
		DetectFunc: func() bool { return os.Getenv("TERM_PROGRAM") == "Tabby" },
	}
}

// NewStAdapter returns an OSCAdapter for st (suckless terminal).
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

// NewUrxvtAdapter returns an OSCAdapter for urxvt (rxvt-unicode).
func NewUrxvtAdapter() *OSCAdapter {
	return &OSCAdapter{
		TermName: "urxvt",
		DetectFunc: func() bool {
			return strings.HasPrefix(os.Getenv("TERM"), "rxvt")
		},
	}
}

// NewVTEAdapter returns an OSCAdapter for VTE-based terminals
// (GNOME Terminal, Tilix, Terminator, xfce4-terminal, LXTerminal, etc.).
// This is a catch-all and should be last in the adapter list.
func NewVTEAdapter() *OSCAdapter {
	return &OSCAdapter{
		TermName:   "vte",
		DetectFunc: func() bool { return os.Getenv("VTE_VERSION") != "" },
	}
}
