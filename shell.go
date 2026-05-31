package main

import "fmt"

// ShellHook returns the shell hook code for the given shell.
func ShellHook(shell string) (string, error) {
	switch shell {
	case "zsh":
		return zshHook(), nil
	case "bash":
		return bashHook(), nil
	case "fish":
		return fishHook(), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s (supported: zsh, bash, fish)", shell)
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
    if [[ "${__coltty_last_pwd:-}" != "$PWD" ]]; then
        __coltty_last_pwd="$PWD"
        coltty apply --quiet 2>/dev/null
    fi
}
if [[ ! "$PROMPT_COMMAND" =~ __coltty_prompt_command ]]; then
    PROMPT_COMMAND="__coltty_prompt_command;${PROMPT_COMMAND}"
fi
`
}

func fishHook() string {
	return `if not functions --query coltty_chpwd
    function coltty_chpwd --on-variable PWD
        coltty apply --quiet 2>/dev/null
    end
end
`
}
