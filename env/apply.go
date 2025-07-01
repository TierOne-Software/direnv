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

package env

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tierone-software/direnv/config"
)

func ApplyConfig(cfg *config.Config, baseDir string) error {
	for key, value := range cfg.Environment {
		expandedValue := expandEnvVar(value, baseDir)
		if err := os.Setenv(key, expandedValue); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", key, err)
		}
	}

	return nil
}

func expandEnvVar(value string, baseDir string) string {
	result := value

	result = strings.ReplaceAll(result, "$PROJECT_ROOT", baseDir)

	result = os.ExpandEnv(result)

	return result
}

func ExecuteScript(scriptName, scriptContent string, baseDir string, args ...string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	// Build the script with positional parameters set
	fullScript := scriptContent
	if len(args) > 0 {
		// Prepend set -- to set positional parameters
		quotedArgs := make([]string, len(args))
		for i, arg := range args {
			quotedArgs[i] = shellQuote(arg)
		}
		fullScript = fmt.Sprintf("set -- %s\n%s", strings.Join(quotedArgs, " "), scriptContent)
	}

	cmd := exec.Command(shell, "-c", fullScript)
	cmd.Dir = baseDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	originalPwd := os.Getenv("PWD")
	cmd.Env = append(os.Environ(),
		"PROJECT_ROOT="+baseDir,
		"PWD="+baseDir,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("script '%s' failed: %w", scriptName, err)
	}

	if originalPwd != "" {
		os.Setenv("PWD", originalPwd)
	}

	return nil
}

func ExportForShell(cfg *config.Config, baseDir string, shellType string) string {
	var exports []string

	// Execute pre-apply hook first
	if cfg.Hooks.PreApply != "" {
		exports = append(exports, fmt.Sprintf("(\n    cd %s\n%s\n)", shellQuote(baseDir), indent(cfg.Hooks.PreApply, "    ")))
	}

	for key, value := range cfg.Environment {
		expandedValue := expandEnvVar(value, baseDir)
		if shellType == "fish" {
			exports = append(exports, fmt.Sprintf("set -gx %s %s", key, shellQuote(expandedValue)))
		} else {
			exports = append(exports, fmt.Sprintf("export %s=%s", key, shellQuote(expandedValue)))
		}
	}

	for name, command := range cfg.Aliases {
		exports = append(exports, fmt.Sprintf("alias %s=%s", name, shellQuote(command)))
	}

	for name, script := range cfg.Scripts {
		// Inject PROJECT_ROOT into the function and pass all arguments
		funcDef := fmt.Sprintf("%s() {\n    local PROJECT_ROOT=%s\n    (\n        cd \"$PROJECT_ROOT\"\n        set -- \"$@\"\n%s\n    )\n}", name, shellQuote(baseDir), indent(script, "        "))
		exports = append(exports, funcDef)
	}

	// Execute post-apply hook last
	if cfg.Hooks.PostApply != "" {
		exports = append(exports, fmt.Sprintf("(\n    cd %s\n%s\n)", shellQuote(baseDir), indent(cfg.Hooks.PostApply, "    ")))
	}

	return strings.Join(exports, "\n")
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

func indent(s string, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
