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
	"strings"
	"testing"

	"github.com/TierOne-Software/direnv/config"
)

func TestGenerateDiff(t *testing.T) {
	// Set some known environment variables
	os.Setenv("TEST_EXISTING", "old_value")
	defer os.Unsetenv("TEST_EXISTING")

	cfg := &config.Config{
		Environment: map[string]string{
			"TEST_EXISTING": "new_value",
			"TEST_NEW":      "added_value",
		},
		Aliases: map[string]string{
			"test_alias": "echo test",
		},
		Scripts: map[string]string{
			"test_script": "echo script",
		},
	}

	diff, err := GenerateDiff(cfg, "/test")
	if err != nil {
		t.Fatalf("GenerateDiff failed: %v", err)
	}

	// Check that we detect the modification
	found := false
	for _, envDiff := range diff.Environment {
		if envDiff.Key == "TEST_EXISTING" && envDiff.Type == Modified {
			found = true
			if envDiff.OldValue != "old_value" || envDiff.NewValue != "new_value" {
				t.Errorf("Expected old_value → new_value, got %s → %s", envDiff.OldValue, envDiff.NewValue)
			}
		}
	}
	if !found {
		t.Error("Expected to find TEST_EXISTING modification in diff")
	}

	// Check that aliases are detected
	if len(diff.Aliases) != 1 || diff.Aliases[0].Name != "test_alias" {
		t.Error("Expected test_alias in diff")
	}

	// Check that scripts are detected
	if len(diff.Scripts) != 1 || diff.Scripts[0].Name != "test_script" {
		t.Error("Expected test_script in diff")
	}
}

func TestDiffFormat(t *testing.T) {
	diff := &ConfigDiff{
		Environment: []EnvDiff{
			{Key: "NEW_VAR", NewValue: "value", Type: Added},
			{Key: "MOD_VAR", OldValue: "old", NewValue: "new", Type: Modified},
		},
		Aliases: []AliasDiff{
			{Name: "test", NewValue: "echo test", Type: Added},
		},
		Scripts: []ScriptDiff{
			{Name: "build", Content: "make build", Type: Added},
		},
	}

	output := diff.Format()

	if !strings.Contains(output, "+ NEW_VAR=value") {
		t.Error("Expected + NEW_VAR=value in output")
	}

	if !strings.Contains(output, "~ MOD_VAR=old → new") {
		t.Error("Expected ~ MOD_VAR=old → new in output")
	}

	if !strings.Contains(output, "+ alias test=echo test") {
		t.Error("Expected alias in output")
	}

	if !strings.Contains(output, "+ function build()") {
		t.Error("Expected function in output")
	}
}
