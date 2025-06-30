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

package shell

import (
	"os"
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		shellPath string
		expected  ShellType
	}{
		{"/bin/bash", Bash},
		{"/usr/bin/bash", Bash},
		{"/bin/zsh", Zsh},
		{"/usr/bin/zsh", Zsh},
		{"/usr/local/bin/fish", Fish},
		{"/bin/sh", Unknown},
		{"", Unknown},
	}

	originalShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", originalShell)

	for _, tt := range tests {
		t.Run(tt.shellPath, func(t *testing.T) {
			if tt.shellPath == "" {
				os.Unsetenv("SHELL")
			} else {
				os.Setenv("SHELL", tt.shellPath)
			}

			result := Detect()
			if result != tt.expected {
				t.Errorf("Detect() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetConfigFile(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get user home directory")
	}

	tests := []struct {
		shellType ShellType
		contains  string
	}{
		{Bash, ".bashrc"},
		{Zsh, ".zshrc"},
		{Fish, ".config/fish/config.fish"},
	}

	for _, tt := range tests {
		t.Run(string(tt.shellType), func(t *testing.T) {
			result := GetConfigFile(tt.shellType)
			if result == "" {
				t.Error("Expected non-empty config file path")
			}
			if homeDir != "" && result != "" && result[:len(homeDir)] != homeDir {
				t.Errorf("Expected config file to be in home directory")
			}
		})
	}
}
