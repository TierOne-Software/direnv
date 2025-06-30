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
	"sort"
	"strings"

	"github.com/tierone-software/direnv/config"
)

type DiffType int

const (
	Added DiffType = iota
	Modified
	Removed
	Unchanged
)

type EnvDiff struct {
	Key      string
	OldValue string
	NewValue string
	Type     DiffType
}

type AliasDiff struct {
	Name     string
	OldValue string
	NewValue string
	Type     DiffType
}

type ScriptDiff struct {
	Name    string
	Content string
	Type    DiffType
}

type ConfigDiff struct {
	Environment []EnvDiff
	Aliases     []AliasDiff
	Scripts     []ScriptDiff
}

func GenerateDiff(cfg *config.Config, baseDir string) (*ConfigDiff, error) {
	diff := &ConfigDiff{}

	// Get current environment
	currentEnv := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			currentEnv[parts[0]] = parts[1]
		}
	}

	// Compare environment variables
	targetEnv := make(map[string]string)
	for key, value := range cfg.Environment {
		expandedValue := expandEnvVar(value, baseDir)
		targetEnv[key] = expandedValue
	}

	// Find all environment keys
	allEnvKeys := make(map[string]bool)
	for key := range currentEnv {
		allEnvKeys[key] = true
	}
	for key := range targetEnv {
		allEnvKeys[key] = true
	}

	for key := range allEnvKeys {
		currentVal, hasCurrentVal := currentEnv[key]
		targetVal, hasTargetVal := targetEnv[key]

		// Only show relevant environment changes
		if !hasCurrentVal && hasTargetVal {
			diff.Environment = append(diff.Environment, EnvDiff{
				Key:      key,
				NewValue: targetVal,
				Type:     Added,
			})
		} else if hasCurrentVal && hasTargetVal && currentVal != targetVal {
			diff.Environment = append(diff.Environment, EnvDiff{
				Key:      key,
				OldValue: currentVal,
				NewValue: targetVal,
				Type:     Modified,
			})
		}
		// Skip showing removals for now since they're mostly system vars
	}

	// Compare aliases (we don't have current aliases, so all are new)
	for name, command := range cfg.Aliases {
		diff.Aliases = append(diff.Aliases, AliasDiff{
			Name:     name,
			NewValue: command,
			Type:     Added,
		})
	}

	// Compare scripts (we don't have current scripts, so all are new)
	for name, content := range cfg.Scripts {
		diff.Scripts = append(diff.Scripts, ScriptDiff{
			Name:    name,
			Content: content,
			Type:    Added,
		})
	}

	// Sort for consistent output
	sort.Slice(diff.Environment, func(i, j int) bool {
		return diff.Environment[i].Key < diff.Environment[j].Key
	})
	sort.Slice(diff.Aliases, func(i, j int) bool {
		return diff.Aliases[i].Name < diff.Aliases[j].Name
	})
	sort.Slice(diff.Scripts, func(i, j int) bool {
		return diff.Scripts[i].Name < diff.Scripts[j].Name
	})

	return diff, nil
}

func (d *ConfigDiff) Format() string {
	var output []string

	if len(d.Environment) > 0 {
		output = append(output, "Environment variables:")
		for _, envDiff := range d.Environment {
			switch envDiff.Type {
			case Added:
				output = append(output, fmt.Sprintf("  + %s=%s", envDiff.Key, envDiff.NewValue))
			case Modified:
				output = append(output, fmt.Sprintf("  ~ %s=%s → %s", envDiff.Key, envDiff.OldValue, envDiff.NewValue))
			case Removed:
				output = append(output, fmt.Sprintf("  - %s=%s", envDiff.Key, envDiff.OldValue))
			}
		}
		output = append(output, "")
	}

	if len(d.Aliases) > 0 {
		output = append(output, "Aliases:")
		for _, aliasDiff := range d.Aliases {
			output = append(output, fmt.Sprintf("  + alias %s=%s", aliasDiff.Name, aliasDiff.NewValue))
		}
		output = append(output, "")
	}

	if len(d.Scripts) > 0 {
		output = append(output, "Scripts (functions):")
		for _, scriptDiff := range d.Scripts {
			lines := strings.Split(scriptDiff.Content, "\n")
			firstLine := strings.TrimSpace(lines[0])
			if len(lines) > 1 {
				output = append(output, fmt.Sprintf("  + function %s() { %s... }", scriptDiff.Name, firstLine))
			} else {
				output = append(output, fmt.Sprintf("  + function %s() { %s }", scriptDiff.Name, firstLine))
			}
		}
	}

	if len(output) == 0 {
		return "No changes would be made.\n"
	}

	return strings.Join(output, "\n") + "\n"
}

func isSystemVar(key string) bool {
	systemVars := []string{
		"PATH", "HOME", "USER", "SHELL", "PWD", "OLDPWD", "TERM", "LANG", "LC_ALL",
		"SSH_CONNECTION", "SSH_CLIENT", "SSH_TTY", "_", "SHLVL", "PS1", "PS2",
	}

	for _, sysVar := range systemVars {
		if key == sysVar {
			return true
		}
	}

	// Skip most system/shell variables that start with underscore or are ALL_CAPS
	if strings.HasPrefix(key, "_") || strings.HasPrefix(key, "XDG_") {
		return true
	}

	return false
}
