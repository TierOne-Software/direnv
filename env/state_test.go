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

func TestStateOperations(t *testing.T) {
	tmpDir := t.TempDir()
	originalStateFile := stateFile
	stateFile = filepath.Join(tmpDir, "test_state.json")
	defer func() { stateFile = originalStateFile }()

	originalEnv := os.Getenv("TEST_VAR")
	defer func() {
		if originalEnv != "" {
			os.Setenv("TEST_VAR", originalEnv)
		} else {
			os.Unsetenv("TEST_VAR")
		}
	}()

	os.Setenv("TEST_VAR", "original_value")

	state, err := GetCurrentState()
	if err != nil {
		t.Fatalf("Failed to get current state: %v", err)
	}

	if state.Environment["TEST_VAR"] != "original_value" {
		t.Errorf("Expected TEST_VAR=original_value, got %s", state.Environment["TEST_VAR"])
	}

	if err := SaveState(state); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	if !HasSavedState() {
		t.Error("Expected saved state to exist")
	}

	os.Setenv("TEST_VAR", "modified_value")
	os.Setenv("NEW_VAR", "new_value")

	loadedState, err := LoadSavedState()
	if err != nil {
		t.Fatalf("Failed to load saved state: %v", err)
	}

	if loadedState.Environment["TEST_VAR"] != "original_value" {
		t.Errorf("Expected saved TEST_VAR=original_value, got %s", loadedState.Environment["TEST_VAR"])
	}

	if err := RestoreState(); err != nil {
		t.Fatalf("Failed to restore state: %v", err)
	}

	if os.Getenv("TEST_VAR") != "original_value" {
		t.Errorf("Expected TEST_VAR to be restored to original_value, got %s", os.Getenv("TEST_VAR"))
	}

	if os.Getenv("NEW_VAR") != "" {
		t.Errorf("Expected NEW_VAR to be unset, got %s", os.Getenv("NEW_VAR"))
	}

	if HasSavedState() {
		t.Error("Expected saved state to be removed after restore")
	}
}
