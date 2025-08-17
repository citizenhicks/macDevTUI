package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InstallConfig represents the configuration for the installer
type InstallConfig struct {
	Homebrew HombrewConfig  `json:"homebrew"`
	Shell    ShellConfig    `json:"shell"`
	DevTools DevToolsConfig `json:"devtools"`
	Dotfiles DotfilesConfig `json:"dotfiles"`
	Terminal TerminalConfig `json:"terminal"`
}

// HombrewConfig contains Homebrew-related configuration
type HombrewConfig struct {
	Install       bool     `json:"install"`
	BrewfilePaths []string `json:"brewfile_paths"`
}

// ShellConfig contains shell setup configuration
type ShellConfig struct {
	Install       bool       `json:"install"`
	RequiredTools []string   `json:"required_tools"`
	ShellFiles    []string   `json:"shell_files"`
	ThemeFile     string     `json:"theme_file"`
	InitCommands  [][]string `json:"init_commands"`
}

// DevToolsConfig contains development tools configuration
type DevToolsConfig struct {
	Install     bool       `json:"install"`
	Languages   []Language `json:"languages"`
	GlobalTools [][]string `json:"global_tools"`
	VerifyTools []string   `json:"verify_tools"`
}

// Language represents a programming language configuration
type Language struct {
	Name     string     `json:"name"`
	Enabled  bool       `json:"enabled"`
	Commands [][]string `json:"commands"`
}

// DotfilesConfig contains dotfiles restoration configuration
type DotfilesConfig struct {
	Install  bool              `json:"install"`
	Mappings map[string]string `json:"mappings"`
}

// TerminalConfig contains terminal configuration
type TerminalConfig struct {
	Install     bool              `json:"install"`
	ConfigFiles map[string]string `json:"config_files"`
}

// LoadConfig loads configuration from JSON file
func LoadConfig() (*InstallConfig, error) {
	// Get current directory and home directory safely
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPaths := []string{
		filepath.Join(currentDir, "install-config.json"),
		filepath.Join(currentDir, "/config/install-config.json"),
		filepath.Join(homeDir, ".config", "install-config.json"),
	}

	for _, configPath := range configPaths {
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				return nil, err
			}

			var config InstallConfig
			if err := json.Unmarshal(data, &config); err != nil {
				return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
			}

			// Validate the configuration
			if err := config.Validate(); err != nil {
				return nil, fmt.Errorf("invalid configuration in %s: %w", configPath, err)
			}

			return &config, nil
		}
	}

	// Return error if no config file found
	return nil, fmt.Errorf("no configuration file found in expected locations: %v", configPaths)
}

// SaveConfig saves configuration to JSON file
func (c *InstallConfig) SaveConfig(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Validate ensures the configuration is valid and safe to use
func (c *InstallConfig) Validate() error {
	// Validate Homebrew config
	if c.Homebrew.Install && len(c.Homebrew.BrewfilePaths) == 0 {
		return fmt.Errorf("homebrew is enabled but no brewfile paths specified")
	}

	// Validate shell config
	if c.Shell.Install {
		if len(c.Shell.RequiredTools) == 0 {
			return fmt.Errorf("shell is enabled but no required tools specified")
		}
		// Check for potentially dangerous commands in shell init
		for _, cmd := range c.Shell.InitCommands {
			if len(cmd) == 0 {
				return fmt.Errorf("empty command in shell init commands")
			}
			// Basic security check - prevent obviously dangerous commands
			dangerousCommands := []string{"rm", "sudo", "chmod", "chown", "dd", "mkfs", "fdisk", "killall", "kill"}
			for _, dangerous := range dangerousCommands {
				if cmd[0] == dangerous {
					return fmt.Errorf("potentially dangerous command in shell init: %s", cmd[0])
				}
			}
			// Check for dangerous flags like -rf in rm commands
			for _, arg := range cmd {
				if strings.Contains(arg, "-rf") || strings.Contains(arg, "--force") {
					return fmt.Errorf("potentially dangerous flag found in command: %s", arg)
				}
			}
		}
	}

	// Validate development tools config
	if c.DevTools.Install {
		for _, lang := range c.DevTools.Languages {
			if lang.Name == "" {
				return fmt.Errorf("language name cannot be empty")
			}
			if lang.Enabled && len(lang.Commands) == 0 {
				return fmt.Errorf("language %s is enabled but has no commands", lang.Name)
			}
		}
	}

	// Validate dotfiles config
	if c.Dotfiles.Install && len(c.Dotfiles.Mappings) == 0 {
		return fmt.Errorf("dotfiles is enabled but no mappings specified")
	}

	// Validate terminal config
	if c.Terminal.Install && len(c.Terminal.ConfigFiles) == 0 {
		return fmt.Errorf("terminal is enabled but no config files specified")
	}

	return nil
}
