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

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigWithHooks(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `
auto_apply = true

[environment]
TEST_VAR = "test"

[hooks]
pre_apply = "echo pre"
post_apply = "echo post"
on_leave = "echo bye"
`

	configPath := filepath.Join(tmpDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Hooks.PreApply != "echo pre" {
		t.Errorf("Expected pre_apply hook, got %s", cfg.Hooks.PreApply)
	}

	if cfg.Hooks.PostApply != "echo post" {
		t.Errorf("Expected post_apply hook, got %s", cfg.Hooks.PostApply)
	}

	if cfg.Hooks.OnLeave != "echo bye" {
		t.Errorf("Expected on_leave hook, got %s", cfg.Hooks.OnLeave)
	}
}

func TestMergeConfigsWithHooks(t *testing.T) {
	base := &Config{
		AutoApply:   false,
		Environment: map[string]string{"BASE": "value"},
		Hooks: Hooks{
			PreApply:  "echo base pre",
			PostApply: "echo base post",
		},
	}

	override := &Config{
		AutoApply:   true,
		Environment: map[string]string{"OVERRIDE": "value"},
		Hooks: Hooks{
			PreApply: "echo override pre",
			OnLeave:  "echo override leave",
		},
	}

	merged := MergeConfigs(base, override)

	if merged.Hooks.PreApply != "echo override pre" {
		t.Errorf("Expected overridden pre_apply hook, got %s", merged.Hooks.PreApply)
	}

	if merged.Hooks.PostApply != "echo base post" {
		t.Errorf("Expected base post_apply hook, got %s", merged.Hooks.PostApply)
	}

	if merged.Hooks.OnLeave != "echo override leave" {
		t.Errorf("Expected override on_leave hook, got %s", merged.Hooks.OnLeave)
	}
}
