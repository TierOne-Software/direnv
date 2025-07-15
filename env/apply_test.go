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

func TestExpandEnvVar(t *testing.T) {
	baseDir := "/test/project"
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "/usr/bin:/bin")
	defer os.Setenv("PATH", originalPath)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "expand PROJECT_ROOT",
			input:    "$PROJECT_ROOT/bin",
			expected: "/test/project/bin",
		},
		{
			name:     "expand PATH",
			input:    "$PATH:/custom/bin",
			expected: "/usr/bin:/bin:/custom/bin",
		},
		{
			name:     "combined expansion",
			input:    "$PATH:$PROJECT_ROOT/bin",
			expected: "/usr/bin:/bin:/test/project/bin",
		},
		{
			name:     "no expansion needed",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandEnvVar(tt.input, baseDir)
			if result != tt.expected {
				t.Errorf("expandEnvVar() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestApplyConfig(t *testing.T) {
	originalPath := os.Getenv("PATH")
	originalCC := os.Getenv("CC")
	defer func() {
		os.Setenv("PATH", originalPath)
		if originalCC != "" {
			os.Setenv("CC", originalCC)
		} else {
			os.Unsetenv("CC")
		}
	}()

	os.Setenv("PATH", "/usr/bin")

	cfg := &config.Config{
		Environment: map[string]string{
			"CC":   "gcc-11",
			"PATH": "$PATH:$PROJECT_ROOT/bin",
		},
	}

	if err := ApplyConfig(cfg, "/project"); err != nil {
		t.Fatalf("ApplyConfig failed: %v", err)
	}

	if os.Getenv("CC") != "gcc-11" {
		t.Errorf("Expected CC=gcc-11, got %s", os.Getenv("CC"))
	}

	expectedPath := "/usr/bin:/project/bin"
	if os.Getenv("PATH") != expectedPath {
		t.Errorf("Expected PATH=%s, got %s", expectedPath, os.Getenv("PATH"))
	}
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "'simple'"},
		{"with spaces", "'with spaces'"},
		{"with'quote", "'with'\"'\"'quote'"},
		{"", "''"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := shellQuote(tt.input)
			if result != tt.expected {
				t.Errorf("shellQuote(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExportForShell(t *testing.T) {
	cfg := &config.Config{
		Environment: map[string]string{
			"TEST_VAR": "value",
			"PATH":     "$PATH:/test",
		},
		Aliases: map[string]string{
			"ll": "ls -la",
		},
		Scripts: map[string]string{
			"build": "echo Building...",
		},
	}

	result := ExportForShell(cfg, "/project", "bash")

	if !strings.Contains(result, "export TEST_VAR='value'") {
		t.Error("Expected export TEST_VAR='value' in output")
	}

	if !strings.Contains(result, "alias ll='ls -la'") {
		t.Error("Expected alias ll='ls -la' in output")
	}

	if !strings.Contains(result, "build() {") {
		t.Error("Expected build function definition in output")
	}
}
