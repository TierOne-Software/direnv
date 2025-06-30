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

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `
auto_apply = true

[environment]
CC = "gcc-11"
CXX = "g++-11"
PATH = "$PATH:/custom/bin"

[aliases]
build = "make build"
test = "make test"

[scripts]
setup = "echo Setting up..."
`

	configPath := filepath.Join(tmpDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if !cfg.AutoApply {
		t.Error("Expected auto_apply to be true")
	}

	if cfg.Environment["CC"] != "gcc-11" {
		t.Errorf("Expected CC=gcc-11, got %s", cfg.Environment["CC"])
	}

	if cfg.Aliases["build"] != "make build" {
		t.Errorf("Expected build alias to be 'make build', got %s", cfg.Aliases["build"])
	}

	if cfg.Scripts["setup"] != "echo Setting up..." {
		t.Errorf("Expected setup script, got %s", cfg.Scripts["setup"])
	}
}

func TestFindConfig(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub", "dir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	configContent := `auto_apply = false`
	configPath := filepath.Join(tmpDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, foundPath, err := FindConfig(subDir)
	if err != nil {
		t.Fatalf("Failed to find config: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected to find config, got nil")
	}

	if foundPath != configPath {
		t.Errorf("Expected config path %s, got %s", configPath, foundPath)
	}

	if cfg.AutoApply {
		t.Error("Expected auto_apply to be false")
	}
}

func TestFindConfigNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	cfg, foundPath, err := FindConfig(tmpDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cfg != nil {
		t.Error("Expected no config, got one")
	}

	if foundPath != "" {
		t.Errorf("Expected empty path, got %s", foundPath)
	}
}

func TestLocalConfigMerging(t *testing.T) {
	tmpDir := t.TempDir()

	baseConfig := `
auto_apply = false

[environment]
CC = "gcc"
CXX = "g++"

[aliases]
build = "make build"
test = "make test"
`

	localConfig := `
auto_apply = true

[environment]
CC = "gcc-11"
DEBUG = "true"

[aliases]
test = "go test ./..."

[scripts]
setup = "echo local setup"
`

	configPath := filepath.Join(tmpDir, ConfigFileName)
	localConfigPath := filepath.Join(tmpDir, LocalConfigFileName)

	if err := os.WriteFile(configPath, []byte(baseConfig), 0644); err != nil {
		t.Fatalf("Failed to write base config: %v", err)
	}

	if err := os.WriteFile(localConfigPath, []byte(localConfig), 0644); err != nil {
		t.Fatalf("Failed to write local config: %v", err)
	}

	cfg, foundPath, err := FindConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to find config: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected to find merged config, got nil")
	}

	if foundPath != configPath {
		t.Errorf("Expected config path %s, got %s", configPath, foundPath)
	}

	// Test merged values
	if !cfg.AutoApply {
		t.Error("Expected auto_apply to be true (from local)")
	}

	if cfg.Environment["CC"] != "gcc-11" {
		t.Errorf("Expected CC=gcc-11 (from local), got %s", cfg.Environment["CC"])
	}

	if cfg.Environment["CXX"] != "g++" {
		t.Errorf("Expected CXX=g++ (from base), got %s", cfg.Environment["CXX"])
	}

	if cfg.Environment["DEBUG"] != "true" {
		t.Errorf("Expected DEBUG=true (from local), got %s", cfg.Environment["DEBUG"])
	}

	if cfg.Aliases["build"] != "make build" {
		t.Errorf("Expected build alias from base, got %s", cfg.Aliases["build"])
	}

	if cfg.Aliases["test"] != "go test ./..." {
		t.Errorf("Expected test alias to be overridden by local, got %s", cfg.Aliases["test"])
	}

	if cfg.Scripts["setup"] != "echo local setup" {
		t.Errorf("Expected setup script from local, got %s", cfg.Scripts["setup"])
	}
}
