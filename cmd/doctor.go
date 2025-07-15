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
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TierOne-Software/direnv/config"
	"github.com/TierOne-Software/direnv/env"
	"github.com/TierOne-Software/direnv/shell"
)

type DiagnosticResult struct {
	Status  string // "✓", "✗", "⚠"
	Message string
}

func doctorCommand() error {
	fmt.Println("direnv doctor - Diagnostic Information")
	fmt.Println("====================================")

	results := []DiagnosticResult{}

	// Check shell detection
	shellType := shell.Detect()
	if shellType == shell.Unknown {
		results = append(results, DiagnosticResult{"✗", "Shell detection failed - unknown shell"})
	} else {
		results = append(results, DiagnosticResult{"✓", fmt.Sprintf("Shell detected: %s", shellType)})
	}

	// Check for config file
	cwd, err := os.Getwd()
	if err != nil {
		results = append(results, DiagnosticResult{"✗", fmt.Sprintf("Failed to get current directory: %v", err)})
	} else {
		cfg, configPath, err := config.FindConfig(cwd)
		if err != nil {
			results = append(results, DiagnosticResult{"✗", fmt.Sprintf("Config search failed: %v", err)})
		} else if cfg == nil {
			results = append(results, DiagnosticResult{"⚠", "No .direnv.toml found in current or parent directories"})
		} else {
			results = append(results, DiagnosticResult{"✓", fmt.Sprintf("Config found: %s", configPath)})

			// Check for local overrides
			localConfigPath := filepath.Join(filepath.Dir(configPath), config.LocalConfigFileName)
			if _, err := os.Stat(localConfigPath); err == nil {
				results = append(results, DiagnosticResult{"✓", fmt.Sprintf("Local config found: %s", localConfigPath)})
			} else {
				results = append(results, DiagnosticResult{"ℹ", "No local config (.direnv.local.toml) found"})
			}

			// Validate config syntax
			if err := validateConfig(cfg); err != nil {
				results = append(results, DiagnosticResult{"✗", fmt.Sprintf("Config validation failed: %v", err)})
			} else {
				results = append(results, DiagnosticResult{"✓", "Config syntax is valid"})
			}
		}
	}

	// Check auto-apply status
	if shell.IsAutoApplyEnabled() {
		results = append(results, DiagnosticResult{"✓", "Auto-apply is enabled"})
	} else {
		results = append(results, DiagnosticResult{"ℹ", "Auto-apply is disabled"})
	}

	// Check shell integration
	shellConfigFile := shell.GetConfigFile(shellType)
	if shellConfigFile == "" {
		results = append(results, DiagnosticResult{"⚠", "Unable to determine shell config file"})
	} else {
		if checkShellIntegration(shellConfigFile) {
			results = append(results, DiagnosticResult{"✓", "Shell integration appears to be installed"})
		} else {
			results = append(results, DiagnosticResult{"⚠", fmt.Sprintf("Shell integration not found in %s", shellConfigFile)})
			results = append(results, DiagnosticResult{"ℹ", "Run: direnv init >> " + shellConfigFile})
		}
	}

	// Check environment state
	if env.HasSavedState() {
		results = append(results, DiagnosticResult{"ℹ", "Previous environment state exists"})
		results = append(results, DiagnosticResult{"ℹ", "Run 'direnv restore' to clean up if needed"})
	} else {
		results = append(results, DiagnosticResult{"✓", "No saved environment state"})
	}

	// Check direnv binary location
	if direnvPath, err := exec.LookPath("direnv"); err != nil {
		results = append(results, DiagnosticResult{"⚠", "direnv not found in PATH"})
	} else {
		results = append(results, DiagnosticResult{"✓", fmt.Sprintf("direnv binary: %s", direnvPath)})
	}

	// Print all results
	fmt.Println()
	for _, result := range results {
		fmt.Printf("%s %s\n", result.Status, result.Message)
	}

	// Summary
	fmt.Println()
	var errorCount, warningCount, successCount int
	for _, result := range results {
		switch result.Status {
		case "✓":
			successCount++
		case "✗":
			errorCount++
		case "⚠":
			warningCount++
		}
	}

	fmt.Printf("Summary: %d checks passed, %d warnings, %d errors\n", successCount, warningCount, errorCount)

	if errorCount > 0 {
		fmt.Println("\n⚠ Please fix the errors above before using direnv")
		return fmt.Errorf("doctor found %d error(s)", errorCount)
	}

	if warningCount > 0 {
		fmt.Println("\n⚠ Some issues were found, but direnv should still work")
	} else {
		fmt.Println("\n✓ Everything looks good!")
	}

	return nil
}

func validateConfig(cfg *config.Config) error {
	// Check for common issues
	for key, value := range cfg.Environment {
		if key == "" {
			return fmt.Errorf("empty environment variable name")
		}
		if strings.Contains(key, " ") {
			return fmt.Errorf("environment variable name contains spaces: %s", key)
		}
		_ = value // value can be empty, that's fine
	}

	for name, command := range cfg.Aliases {
		if name == "" {
			return fmt.Errorf("empty alias name")
		}
		if command == "" {
			return fmt.Errorf("empty alias command for: %s", name)
		}
	}

	for name, script := range cfg.Scripts {
		if name == "" {
			return fmt.Errorf("empty script name")
		}
		if script == "" {
			return fmt.Errorf("empty script content for: %s", name)
		}
	}

	return nil
}

func checkShellIntegration(configFile string) bool {
	content, err := os.ReadFile(configFile)
	if err != nil {
		return false
	}

	contentStr := string(content)
	return strings.Contains(contentStr, "_direnv_check") ||
		strings.Contains(contentStr, "direnv init") ||
		strings.Contains(contentStr, "DIRENV_SHELL")
}
