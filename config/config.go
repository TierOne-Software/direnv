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
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	AutoApply   bool              `toml:"auto_apply"`
	Environment map[string]string `toml:"environment"`
	Aliases     map[string]string `toml:"aliases"`
	Scripts     map[string]string `toml:"scripts"`
	Hooks       Hooks             `toml:"hooks"`
}

type Hooks struct {
	PreApply  string `toml:"pre_apply"`
	PostApply string `toml:"post_apply"`
	OnLeave   string `toml:"on_leave"`
}

const (
	ConfigFileName      = ".direnv.toml"
	LocalConfigFileName = ".direnv.local.toml"
)

func LoadConfig(path string) (*Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	if cfg.Environment == nil {
		cfg.Environment = make(map[string]string)
	}
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]string)
	}
	if cfg.Scripts == nil {
		cfg.Scripts = make(map[string]string)
	}

	return &cfg, nil
}

func FindConfig(startDir string) (*Config, string, error) {
	dir := startDir

	for {
		configPath := filepath.Join(dir, ConfigFileName)

		if _, err := os.Stat(configPath); err == nil {
			cfg, err := LoadConfig(configPath)
			if err != nil {
				return nil, "", err
			}

			// Try to load local overrides
			localConfigPath := filepath.Join(dir, LocalConfigFileName)
			if _, err := os.Stat(localConfigPath); err == nil {
				localCfg, err := LoadConfig(localConfigPath)
				if err != nil {
					return nil, "", fmt.Errorf("failed to load local config: %w", err)
				}
				cfg = MergeConfigs(cfg, localCfg)
			}

			return cfg, configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return nil, "", nil
}

func MergeConfigs(base, override *Config) *Config {
	merged := &Config{
		AutoApply:   override.AutoApply || base.AutoApply,
		Environment: make(map[string]string),
		Aliases:     make(map[string]string),
		Scripts:     make(map[string]string),
		Hooks:       base.Hooks, // Start with base hooks
	}

	// Copy base values
	for k, v := range base.Environment {
		merged.Environment[k] = v
	}
	for k, v := range base.Aliases {
		merged.Aliases[k] = v
	}
	for k, v := range base.Scripts {
		merged.Scripts[k] = v
	}

	// Override with local values
	for k, v := range override.Environment {
		merged.Environment[k] = v
	}
	for k, v := range override.Aliases {
		merged.Aliases[k] = v
	}
	for k, v := range override.Scripts {
		merged.Scripts[k] = v
	}

	// Override hooks if they exist in local config
	if override.Hooks.PreApply != "" {
		merged.Hooks.PreApply = override.Hooks.PreApply
	}
	if override.Hooks.PostApply != "" {
		merged.Hooks.PostApply = override.Hooks.PostApply
	}
	if override.Hooks.OnLeave != "" {
		merged.Hooks.OnLeave = override.Hooks.OnLeave
	}

	return merged
}
