/*
 * Copyright 2025 TierOne Software
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package shell

import (
	"fmt"
	"os"

	"github.com/tierone-software/direnv/config"
)

func GetCompletionScript(shellType ShellType) string {
	switch shellType {
	case Bash:
		return bashCompletionScript
	case Zsh:
		return zshCompletionScript
	default:
		return fmt.Sprintf("# Completion for shell type '%s' is not yet supported\n", shellType)
	}
}

func ListAvailableScripts() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cfg, _, err := config.FindConfig(cwd)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return []string{}, nil
	}

	scripts := make([]string, 0, len(cfg.Scripts))
	for name := range cfg.Scripts {
		scripts = append(scripts, name)
	}
	return scripts, nil
}

const bashCompletionScript = `# direnv bash completion
_direnv() {
    local cur prev opts scripts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    opts="apply diff info enable disable init completion restore run"

    case "${prev}" in
        run)
            # Complete with available scripts
            if [ -f ".direnv.toml" ]; then
                scripts=$(direnv completion scripts 2>/dev/null)
                COMPREPLY=($(compgen -W "${scripts}" -- ${cur}))
            fi
            return 0
            ;;
        *)
            ;;
    esac

    COMPREPLY=($(compgen -W "${opts}" -- ${cur}))
    return 0
}

complete -F _direnv direnv

# Complete scripts directly when they're available as functions
_direnv_script_completion() {
    if [ -f ".direnv.toml" ]; then
        local scripts=$(direnv completion scripts 2>/dev/null)
        for script in $scripts; do
            complete -F _empty_completion $script 2>/dev/null || true
        done
    fi
}

_empty_completion() {
    COMPREPLY=()
}

# Register script completions when direnv is applied
if command -v direnv >/dev/null 2>&1; then
    _direnv_script_completion
fi
`

const zshCompletionScript = `# direnv zsh completion
#compdef direnv

_direnv() {
    local context state line
    typeset -A opt_args

    local -a commands
    commands=(
        'apply:Apply directory environment'
        'diff:Show what would change'
        'info:Show current status'
        'enable:Enable auto-apply'
        'disable:Disable auto-apply'
        'init:Initialize shell integration'
        'completion:Generate shell completion'
        'restore:Restore previous environment'
        'run:Run a script from the config'
    )

    _arguments \
        '1: :->commands' \
        '*: :->args' \
        && return 0

    case $state in
        commands)
            _describe 'direnv commands' commands
            ;;
        args)
            case $words[2] in
                run)
                    if [[ -f ".direnv.toml" ]]; then
                        local -a scripts
                        scripts=(${(f)"$(direnv completion scripts 2>/dev/null)"})
                        _describe 'scripts' scripts
                    fi
                    ;;
            esac
            ;;
    esac
}

# Function to complete scripts when they're available as functions
_direnv_register_scripts() {
    if [[ -f ".direnv.toml" ]]; then
        local scripts=(${(f)"$(direnv completion scripts 2>/dev/null)"})
        for script in $scripts; do
            compdef '_message "direnv script"' $script 2>/dev/null || true
        done
    fi
}

# Register completions for direnv
compdef _direnv direnv

# Register script completions when direnv is applied
if command -v direnv >/dev/null 2>&1; then
    _direnv_register_scripts
fi
`
