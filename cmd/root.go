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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tierone-software/direnv/config"
	"github.com/tierone-software/direnv/env"
	"github.com/tierone-software/direnv/shell"
)

func Execute() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: direnv <command> [args]\n\nCommands:\n  apply     - Apply directory environment\n  diff      - Show what would change\n  info      - Show current status\n  enable    - Enable auto-apply\n  disable   - Disable auto-apply\n  init      - Initialize shell integration\n  completion - Generate shell completion\n  doctor    - Diagnose configuration issues\n  cleanup   - Clean up orphaned state files\n  restore   - Restore previous environment\n  run       - Run a script from the config")
	}

	command := os.Args[1]

	switch command {
	case "apply":
		return applyCommand()
	case "diff":
		return diffCommand()
	case "info":
		return infoCommand()
	case "enable":
		return enableCommand()
	case "disable":
		return disableCommand()
	case "init":
		return initCommand()
	case "completion":
		return completionCommand()
	case "doctor":
		return doctorCommand()
	case "cleanup":
		return cleanupCommand()
	case "restore":
		return restoreCommand()
	case "run":
		if len(os.Args) < 3 {
			return fmt.Errorf("usage: direnv run <script-name>")
		}
		return runScriptCommand(os.Args[2])
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func applyCommand() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	cfg, configPath, err := config.FindConfig(cwd)
	if err != nil {
		return fmt.Errorf("failed to find config: %w", err)
	}
	if cfg == nil {
		return fmt.Errorf("no .direnv.toml found in current or parent directories")
	}

	configDir := filepath.Dir(configPath)

	// Execute on-leave hook from previous environment if it exists
	if err := env.ExecuteOnLeaveHook(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: on-leave hook failed: %v\n", err)
	}

	// Save current state before applying new environment
	state, err := env.GetCurrentState()
	if err != nil {
		return fmt.Errorf("failed to get current state: %w", err)
	}

	if err := env.SaveStateWithHook(state, configDir, cfg.Hooks.OnLeave); err != nil {
		return fmt.Errorf("failed to save current state: %w", err)
	}

	shellType := shell.Detect()

	// Output shell commands for evaluation
	output := env.ExportForShell(cfg, configDir, string(shellType))
	fmt.Print(output)

	return nil
}

func diffCommand() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	cfg, configPath, err := config.FindConfig(cwd)
	if err != nil {
		return fmt.Errorf("failed to find config: %w", err)
	}
	if cfg == nil {
		fmt.Println("No .direnv.toml found in current or parent directories")
		return nil
	}

	configDir := filepath.Dir(configPath)

	diff, err := env.GenerateDiff(cfg, configDir)
	if err != nil {
		return fmt.Errorf("failed to generate diff: %w", err)
	}

	fmt.Printf("Changes that would be applied from %s:\n\n", configPath)
	fmt.Print(diff.Format())

	return nil
}

func infoCommand() error {
	shellType := shell.Detect()
	fmt.Printf("Shell: %s\n", shellType)

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	cfg, configPath, err := config.FindConfig(cwd)
	if err != nil {
		return fmt.Errorf("failed to find config: %w", err)
	}

	if configPath != "" {
		fmt.Printf("Config: %s\n", configPath)

		// Check for local config
		localConfigPath := filepath.Join(filepath.Dir(configPath), config.LocalConfigFileName)
		if _, err := os.Stat(localConfigPath); err == nil {
			fmt.Printf("Local overrides: %s\n", localConfigPath)
		}

		// Show config summary
		if cfg != nil {
			envCount := len(cfg.Environment)
			aliasCount := len(cfg.Aliases)
			scriptCount := len(cfg.Scripts)
			fmt.Printf("Environment: %d variables, %d aliases, %d scripts\n", envCount, aliasCount, scriptCount)

			if cfg.Hooks.PreApply != "" || cfg.Hooks.PostApply != "" || cfg.Hooks.OnLeave != "" {
				fmt.Printf("Hooks: ")
				hooks := []string{}
				if cfg.Hooks.PreApply != "" {
					hooks = append(hooks, "pre-apply")
				}
				if cfg.Hooks.PostApply != "" {
					hooks = append(hooks, "post-apply")
				}
				if cfg.Hooks.OnLeave != "" {
					hooks = append(hooks, "on-leave")
				}
				fmt.Printf("%s\n", strings.Join(hooks, ", "))
			}
		}
	} else {
		fmt.Println("Config: none found")
	}

	// Enhanced state information
	if env.HasSavedState() {
		activeDir, inActiveDir := env.GetActiveEnvironmentInfo()
		if activeDir != "" {
			if inActiveDir {
				fmt.Printf("Active environment: %s (current)\n", activeDir)
			} else {
				fmt.Printf("Active environment: %s (outside)\n", activeDir)
			}

			state, err := env.LoadSavedState()
			if err == nil && state != nil && state.OnLeaveHook != "" {
				fmt.Println("On-leave hook: configured")
			}
		} else {
			fmt.Println("State: environment modified")
		}
		fmt.Printf("State file: ~/.direnv_state_%d.json\n", os.Getpid())
	} else {
		fmt.Println("State: clean")
	}

	// Auto-apply status
	if shell.IsAutoApplyEnabled() {
		fmt.Println("Auto-apply: enabled")
	} else {
		fmt.Println("Auto-apply: disabled")
	}

	return nil
}

func enableCommand() error {
	return shell.SetAutoApply(true)
}

func disableCommand() error {
	return shell.SetAutoApply(false)
}

func initCommand() error {
	shellType := shell.Detect()
	script := shell.GetInitScript(shellType)
	fmt.Print(script)
	return nil
}

func completionCommand() error {
	if len(os.Args) >= 3 && os.Args[2] == "scripts" {
		// Special case: list available scripts for completion
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		cfg, _, err := config.FindConfig(cwd)
		if err != nil || cfg == nil {
			return nil // No error, just no scripts
		}

		for name := range cfg.Scripts {
			fmt.Println(name)
		}
		return nil
	}

	shellType := shell.Detect()
	script := shell.GetCompletionScript(shellType)
	fmt.Print(script)
	return nil
}

func restoreCommand() error {
	// Execute on-leave hook before restoring
	if err := env.ExecuteOnLeaveHook(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: on-leave hook failed: %v\n", err)
	}

	if err := env.RestoreState(); err != nil {
		return fmt.Errorf("failed to restore state: %w", err)
	}
	fmt.Println("Environment restored")
	return nil
}

func cleanupCommand() error {
	if err := env.CleanupOrphanedState(); err != nil {
		return fmt.Errorf("failed to cleanup: %w", err)
	}
	fmt.Println("Cleanup completed")
	return nil
}

func runScriptCommand(scriptName string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	cfg, configPath, err := config.FindConfig(cwd)
	if err != nil {
		return fmt.Errorf("failed to find config: %w", err)
	}
	if cfg == nil {
		return fmt.Errorf("no .direnv.toml found in current or parent directories")
	}

	script, exists := cfg.Scripts[scriptName]
	if !exists {
		return fmt.Errorf("script '%s' not found in config", scriptName)
	}

	configDir := filepath.Dir(configPath)
	return env.ExecuteScript(scriptName, script, configDir)
}
