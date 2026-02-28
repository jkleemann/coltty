package main

import "fmt"

// ShellHook returns the shell hook code for the given shell.
func ShellHook(shell string) (string, error) {
	switch shell {
	case "zsh":
		return zshHook(), nil
	case "bash":
		return bashHook(), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s (supported: zsh, bash)", shell)
	}
}

func zshHook() string {
	return `coltty_chpwd() {
    coltty apply --quiet 2>/dev/null
}
chpwd_functions+=(coltty_chpwd)
`
}

func bashHook() string {
	return `__coltty_prompt_command() {
    coltty apply --quiet 2>/dev/null
}
if [[ ! "$PROMPT_COMMAND" =~ __coltty_prompt_command ]]; then
    PROMPT_COMMAND="__coltty_prompt_command;${PROMPT_COMMAND}"
fi
`
}
