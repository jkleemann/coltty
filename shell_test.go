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
	if !strings.Contains(hook, "&!") {
		t.Error("zsh hook should run in background with &!")
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

func TestShellHookUnsupported(t *testing.T) {
	_, err := ShellHook("fish")
	if err == nil {
		t.Error("expected error for unsupported shell")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("expected 'unsupported shell' error, got: %v", err)
	}
}
