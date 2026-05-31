package main

import (
	"strings"
	"testing"
)

func TestShellHookZsh(t *testing.T) {
	hook, err := ShellHook("zsh")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(hook, "coltty_chpwd") {
		t.Error("zsh hook should contain coltty_chpwd function")
	}
	if !strings.Contains(hook, "chpwd_functions") {
		t.Error("zsh hook should register with chpwd_functions")
	}
	if !strings.Contains(hook, "coltty apply --quiet") {
		t.Error("zsh hook should call coltty apply --quiet")
	}
	if strings.Contains(hook, "&") {
		t.Error("zsh hook should run in foreground (no &) so OSC sequences reach the terminal")
	}
}

func TestShellHookBash(t *testing.T) {
	hook, err := ShellHook("bash")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(hook, "__coltty_prompt_command") {
		t.Error("bash hook should contain __coltty_prompt_command function")
	}
	if !strings.Contains(hook, "PROMPT_COMMAND") {
		t.Error("bash hook should set PROMPT_COMMAND")
	}
	if !strings.Contains(hook, "coltty apply --quiet") {
		t.Error("bash hook should call coltty apply --quiet")
	}
}

func TestBashHookOnlyRunsOnDirectoryChange(t *testing.T) {
	hook, err := ShellHook("bash")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(hook, "${__coltty_last_pwd:-}") {
		t.Error("bash hook should use ${__coltty_last_pwd:-} for set -u compatibility")
	}
	if !strings.Contains(hook, "$PWD") {
		t.Error("bash hook should reference $PWD")
	}
	if !strings.Contains(hook, "__coltty_last_pwd=") {
		t.Error("bash hook should assign __coltty_last_pwd")
	}
	// The hook should only call coltty apply inside a conditional, not unconditionally
	lines := strings.Split(hook, "\n")
	applyLine := -1
	for i, line := range lines {
		if strings.Contains(line, "coltty apply --quiet") {
			applyLine = i
			break
		}
	}
	if applyLine == -1 {
		t.Fatal("bash hook should contain coltty apply --quiet")
	}
	// Check that the apply line is indented (inside a block) rather than at top level
	applyLineTrimmed := strings.TrimLeft(lines[applyLine], " \t")
	if applyLineTrimmed == lines[applyLine] {
		t.Error("bash hook should call coltty apply --quiet inside a conditional block, not at top level")
	}
}

func TestShellHookFish(t *testing.T) {
	hook, err := ShellHook("fish")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(hook, "function coltty_chpwd") {
		t.Error("fish hook should contain function coltty_chpwd")
	}
	if !strings.Contains(hook, "--on-variable PWD") {
		t.Error("fish hook should use --on-variable PWD")
	}
	if !strings.Contains(hook, "coltty apply --quiet") {
		t.Error("fish hook should call coltty apply --quiet")
	}
	if !strings.Contains(hook, "functions --query coltty_chpwd") {
		t.Error("fish hook should guard against duplicate registration with functions --query")
	}
}

func TestZshHookDuplicateGuard(t *testing.T) {
	hook, err := ShellHook("zsh")
	if err != nil {
		t.Fatal(err)
	}
	// The hook should check if coltty_chpwd is already in chpwd_functions
	// before appending, to avoid duplicate registration on repeated sourcing.
	if !strings.Contains(hook, "chpwd_functions[(I)coltty_chpwd]") {
		t.Error("zsh hook should guard against duplicate registration with chpwd_functions[(I)coltty_chpwd]")
	}
}

func TestBashHookAnchoredRegex(t *testing.T) {
	hook, err := ShellHook("bash")
	if err != nil {
		t.Fatal(err)
	}
	// The hook should use an exact-match check for __coltty_prompt_command,
	// not an unanchored regex that could match my__coltty_prompt_command.
	if strings.Contains(hook, "=~ __coltty_prompt_command") {
		t.Error("bash hook should not use unanchored =~ regex for duplicate check")
	}
	// Should use a case statement or similarly robust check
	if !strings.Contains(hook, "case") {
		t.Error("bash hook should use case statement for exact matching")
	}
}

func TestShellHookUnsupported(t *testing.T) {
	_, err := ShellHook("tcsh")
	if err == nil {
		t.Error("expected error for unsupported shell")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("expected 'unsupported shell' error, got: %v", err)
	}
}
