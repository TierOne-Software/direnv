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
	"path/filepath"
	"strings"
)

type ShellType string

const (
	Bash    ShellType = "bash"
	Zsh     ShellType = "zsh"
	Fish    ShellType = "fish"
	Unknown ShellType = "unknown"
)

func Detect() ShellType {
	shellEnv := os.Getenv("SHELL")
	if shellEnv == "" {
		return Unknown
	}

	shellName := filepath.Base(shellEnv)

	switch {
	case strings.Contains(shellName, "bash"):
		return Bash
	case strings.Contains(shellName, "zsh"):
		return Zsh
	case strings.Contains(shellName, "fish"):
		return Fish
	default:
		return Unknown
	}
}

func GetConfigFile(shellType ShellType) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch shellType {
	case Bash:
		return filepath.Join(homeDir, ".bashrc")
	case Zsh:
		return filepath.Join(homeDir, ".zshrc")
	case Fish:
		return filepath.Join(homeDir, ".config", "fish", "config.fish")
	default:
		return ""
	}
}

func SetAutoApply(enabled bool) error {
	// For environment variable approach, we just give instructions
	if enabled {
		fmt.Println("To enable auto-apply, run:")
		fmt.Println("  export DIRENV_AUTO_APPLY=1")
		fmt.Println("Add this to your shell config (~/.bashrc or ~/.zshrc) to make it permanent.")
	} else {
		fmt.Println("To disable auto-apply, run:")
		fmt.Println("  export DIRENV_AUTO_APPLY=0")
		fmt.Println("Or remove the variable entirely with: unset DIRENV_AUTO_APPLY")
	}
	return nil
}

func IsAutoApplyEnabled() bool {
	autoApply := os.Getenv("DIRENV_AUTO_APPLY")
	return autoApply == "1"
}
