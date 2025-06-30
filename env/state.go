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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type State struct {
	Environment map[string]string `json:"environment"`
	Aliases     map[string]string `json:"aliases"`
	Directory   string            `json:"directory"`
	OnLeaveHook string            `json:"on_leave_hook"`
}

var stateFile string
var stateDir string

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get home directory: %v", err))
	}

	// Use ~/.config/direnv/ directory for state files
	stateDir = filepath.Join(homeDir, ".config", "direnv")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create direnv config directory: %v", err))
	}

	// Use parent shell PID for consistent state across direnv command invocations
	shellPID := getShellPID()
	stateFile = filepath.Join(stateDir, fmt.Sprintf("state_%d.json", shellPID))
}

// getShellPID returns the PID of the parent shell
func getShellPID() int {
	// Try to get from environment variable (set by shell integration)
	if shellPIDStr := os.Getenv("DIRENV_SHELL_PID"); shellPIDStr != "" {
		if shellPID, err := strconv.Atoi(shellPIDStr); err == nil {
			return shellPID
		}
	}

	// Fallback to parent process PID
	return os.Getppid()
}

func GetCurrentState() (*State, error) {
	state := &State{
		Environment: make(map[string]string),
		Aliases:     make(map[string]string),
	}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			state.Environment[parts[0]] = parts[1]
		}
	}

	return state, nil
}

func SaveState(state *State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(stateFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func SaveStateWithHook(state *State, directory string, onLeaveHook string) error {
	state.Directory = directory
	state.OnLeaveHook = onLeaveHook
	return SaveState(state)
}

func LoadSavedState() (*State, error) {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

func RestoreState() error {
	state, err := LoadSavedState()
	if err != nil {
		return err
	}
	if state == nil {
		return nil
	}

	currentState, err := GetCurrentState()
	if err != nil {
		return err
	}

	for key := range currentState.Environment {
		if _, exists := state.Environment[key]; !exists {
			if err := os.Unsetenv(key); err != nil {
				return fmt.Errorf("failed to unset %s: %w", key, err)
			}
		}
	}

	for key, value := range state.Environment {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	if err := os.Remove(stateFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove state file: %w", err)
	}

	return nil
}

func HasSavedState() bool {
	_, err := os.Stat(stateFile)
	return err == nil
}

func ExecuteOnLeaveHook() error {
	state, err := LoadSavedState()
	if err != nil || state == nil {
		return nil // No state, nothing to do
	}

	if state.OnLeaveHook != "" && state.Directory != "" {
		if err := ExecuteScript("on_leave", state.OnLeaveHook, state.Directory); err != nil {
			return fmt.Errorf("on-leave hook failed: %w", err)
		}
	}

	return nil
}
