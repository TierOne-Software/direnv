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

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	configContent := `
auto_apply = true

[environment]
TEST_INTEGRATION_VAR = "integration_value"
PROJECT_NAME = "test_project"
PATH = "$PATH:$PROJECT_ROOT/bin"

[aliases]
test-alias = "echo test alias works"

[scripts]
hello = "echo Hello from direnv script"
`

	configPath := filepath.Join(tmpDir, ".direnv.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	direnvBinary := filepath.Join(originalDir, "direnv")

	cmd := exec.Command(direnvBinary, "info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run direnv info: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "Config: "+configPath) {
		t.Errorf("Expected config path in output, got: %s", output)
	}

	cmd = exec.Command(direnvBinary, "apply")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run direnv apply: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "export TEST_INTEGRATION_VAR") {
		t.Errorf("Expected environment export in output, got: %s", output)
	}

	cmd = exec.Command(direnvBinary, "run", "hello")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run direnv run hello: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "Hello from direnv script") {
		t.Errorf("Expected script output, got: %s", output)
	}
}
