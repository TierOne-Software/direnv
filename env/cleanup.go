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
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func CleanupOrphanedState() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Find all state_*.json files in ~/.config/direnv/
	configDir := filepath.Join(homeDir, ".config", "direnv")
	pattern := filepath.Join(configDir, "state_*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find state files: %w", err)
	}

	cleaned := 0
	for _, stateFile := range matches {
		// Extract PID from filename (state_<PID>.json)
		base := filepath.Base(stateFile)
		if !strings.HasPrefix(base, "state_") || !strings.HasSuffix(base, ".json") {
			continue
		}

		pidStr := strings.TrimPrefix(base, "state_")
		pidStr = strings.TrimSuffix(pidStr, ".json")

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue // Skip malformed filenames
		}

		// Check if process is still running
		if !isProcessRunning(pid) {
			// Process is dead, remove the state file
			if err := os.Remove(stateFile); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove orphaned state file %s: %v\n", stateFile, err)
			} else {
				cleaned++
			}
		}
	}

	if cleaned > 0 {
		fmt.Printf("Cleaned up %d orphaned state file(s)\n", cleaned)
	}

	return nil
}

// isProcessRunning checks if a process with the given PID is still running
func isProcessRunning(pid int) bool {
	// Send signal 0 to check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func GetActiveEnvironmentInfo() (string, bool) {
	state, err := LoadSavedState()
	if err != nil || state == nil {
		return "", false
	}

	if state.Directory == "" {
		return "", false
	}

	// Check if we're currently in the active directory or a subdirectory
	cwd, err := os.Getwd()
	if err != nil {
		return state.Directory, true
	}

	// Check if current directory is under the active environment directory
	relPath, err := filepath.Rel(state.Directory, cwd)
	if err != nil || filepath.IsAbs(relPath) || len(relPath) >= 2 && relPath[:2] == ".." {
		// We're outside the active environment
		return state.Directory, false
	}

	return state.Directory, true
}
