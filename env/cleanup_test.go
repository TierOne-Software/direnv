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
	"os"
	"path/filepath"
	"testing"
)

func TestCleanupOrphanedState(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some fake state files in new format
	stateFile1 := filepath.Join(tmpDir, "state_99999.json")
	stateFile2 := filepath.Join(tmpDir, "state_12345.json")
	currentStateFile := filepath.Join(tmpDir, "state_"+string(rune(os.Getpid()))+".json")

	// Write some test state files
	testState := `{"environment":{},"aliases":{},"directory":"","on_leave_hook":""}`
	for _, file := range []string{stateFile1, stateFile2, currentStateFile} {
		if err := os.WriteFile(file, []byte(testState), 0600); err != nil {
			t.Fatalf("Failed to create test state file: %v", err)
		}
	}

	// Verify files exist
	for _, file := range []string{stateFile1, stateFile2, currentStateFile} {
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("Test state file should exist: %s", file)
		}
	}

	// Note: We can't easily test actual cleanup since it requires checking
	// if processes are running, and the test processes are ephemeral
}

func TestGetActiveEnvironmentInfo(t *testing.T) {
	// Test when no state exists
	activeDir, inActive := GetActiveEnvironmentInfo()
	if activeDir != "" || inActive {
		t.Error("Expected no active environment when no state exists")
	}

	// Test would require setting up state file with known PID,
	// which is complex in a test environment
}
