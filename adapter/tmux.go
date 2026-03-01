package adapter

import (
	"os"
	"strings"
)

// InTmux returns true if running inside a tmux session.
func InTmux() bool {
	return os.Getenv("TMUX") != ""
}

// InScreen returns true if running inside a GNU Screen session.
func InScreen() bool {
	return os.Getenv("STY") != ""
}

// WrapTmuxPassthrough wraps OSC sequences in DCS passthrough so they reach
// the outer terminal when running inside tmux.
// Each OSC sequence is wrapped individually: \033Ptmux;\033<escaped-sequence>\033\\
func WrapTmuxPassthrough(oscOutput string) string {
	sequences := splitOSCSequences(oscOutput)
	var b strings.Builder
	for _, seq := range sequences {
		// Double all ESC characters inside the sequence
		escaped := strings.ReplaceAll(seq, "\033", "\033\033")
		b.WriteString("\033Ptmux;")
		b.WriteString(escaped)
		b.WriteString("\033\\")
	}
	return b.String()
}

// splitOSCSequences splits a string of concatenated OSC sequences into
// individual sequences. Each sequence starts with ESC] and ends with ST (ESC\).
func splitOSCSequences(s string) []string {
	var sequences []string
	for len(s) > 0 {
		// Find start of OSC: \033]
		start := strings.Index(s, "\033]")
		if start == -1 {
			break
		}
		// Find the ST terminator: \033\  (after the opening \033])
		end := strings.Index(s[start+2:], "\033\\")
		if end == -1 {
			// Unterminated sequence; include the rest
			sequences = append(sequences, s[start:])
			break
		}
		// end is relative to start+2, and we need to include the ST itself (\033\\)
		seqEnd := start + 2 + end + 2
		sequences = append(sequences, s[start:seqEnd])
		s = s[seqEnd:]
	}
	return sequences
}
