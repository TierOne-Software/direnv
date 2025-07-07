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

import "fmt"

func GetInitScript(shellType ShellType) string {
	switch shellType {
	case Bash:
		return bashInitScript
	case Zsh:
		return zshInitScript
	default:
		return fmt.Sprintf("# Shell type '%s' is not yet supported\n", shellType)
	}
}

const bashInitScript = `# direnv - Directory Environment Manager
# Add this to your ~/.bashrc

export DIRENV_SHELL=bash

_direnv_check() {
    # Prevent recursive calls
    [[ "${_DIRENV_IN_PROGRESS:-}" == "1" ]] && return
    
    if [[ "${DIRENV_AUTO_APPLY:-}" =~ ^1$ ]]; then
        if [[ -f ".direnv.toml" ]]; then
            export _DIRENV_IN_PROGRESS=1
            eval "$(direnv apply 2>/dev/null)"
            unset _DIRENV_IN_PROGRESS
        fi
    fi
}

_direnv_cd() {
    builtin cd "$@" && _direnv_check
}

_direnv_pushd() {
    builtin pushd "$@" && _direnv_check
}

_direnv_popd() {
    builtin popd "$@" && _direnv_check
}

direnv-apply() {
    eval "$(direnv apply)"
}

direnv-restore() {
    direnv restore
}

direnv-info() {
    direnv info
}

direnv-enable() {
    direnv enable
    echo "Auto-apply enabled"
}

direnv-disable() {
    direnv disable
    echo "Auto-apply disabled"
}

# Override cd, pushd, popd
alias cd='_direnv_cd'
alias pushd='_direnv_pushd'
alias popd='_direnv_popd'

# Initial check for current directory
_direnv_check

# Load completions if available
if command -v direnv >/dev/null 2>&1; then
    eval "$(direnv completion bash 2>/dev/null)"
fi
`

const zshInitScript = `# direnv - Directory Environment Manager
# Add this to your ~/.zshrc

export DIRENV_SHELL=zsh

_direnv_check() {
    # Prevent recursive calls
    [[ "${_DIRENV_IN_PROGRESS:-}" == "1" ]] && return
    
    if [[ "${DIRENV_AUTO_APPLY:-}" =~ ^1$ ]]; then
        if [[ -f ".direnv.toml" ]]; then
            export _DIRENV_IN_PROGRESS=1
            eval "$(direnv apply 2>/dev/null)"
            unset _DIRENV_IN_PROGRESS
        fi
    fi
}

# Use zsh's native chpwd hook
autoload -U add-zsh-hook
add-zsh-hook chpwd _direnv_check

direnv-apply() {
    eval "$(direnv apply)"
}

direnv-restore() {
    direnv restore
}

direnv-info() {
    direnv info
}

direnv-enable() {
    direnv enable
    echo "Auto-apply enabled"
}

direnv-disable() {
    direnv disable
    echo "Auto-apply disabled"
}

# For manual directory changes that might not trigger chpwd
_direnv_cd() {
    builtin cd "$@" && _direnv_check
}

_direnv_pushd() {
    builtin pushd "$@" && _direnv_check
}

_direnv_popd() {
    builtin popd "$@" && _direnv_check
}

# Optional aliases (zsh's chpwd hook should handle most cases)
alias cd='_direnv_cd'
alias pushd='_direnv_pushd'
alias popd='_direnv_popd'

# Initial check for current directory
_direnv_check

# Load completions if available
if command -v direnv >/dev/null 2>&1; then
    eval "$(direnv completion zsh 2>/dev/null)"
fi

# Apply cd completions to direnv aliases
compdef _cd _direnv_cd
compdef _cd _direnv_pushd
compdef _cd _direnv_popd
`
