package main

import (
	"fmt"
	"strings"
	"time"
)

// InstallStatus represents the status of an installation step
type InstallStatus int

const (
	StatusReady InstallStatus = iota
	StatusInProgress
	StatusComplete
	StatusError
)

func (s InstallStatus) String() string {
	switch s {
	case StatusReady:
		return "● Ready"
	case StatusInProgress:
		return "◐ Installing..."
	case StatusComplete:
		return "✓ Complete"
	case StatusError:
		return "✗ Error"
	default:
		return "Unknown"
	}
}

// SetupStep represents a single setup step
type SetupStep struct {
	ID          string
	Title       string
	Icon        string
	Description string
	Items       []string
	EstTime     time.Duration
	Status      InstallStatus
	Error       string
	Progress    int // 0-100
	Enabled     bool
}

// getConfigurableSteps creates steps based on loaded configuration
func getConfigurableSteps(config *InstallConfig) []SetupStep {
	if config == nil {
		return []SetupStep{} // Return empty steps if no config
	}
	
	var steps []SetupStep
	
	// Homebrew step
	if config.Homebrew.Install {
		steps = append(steps, SetupStep{
			ID:          "homebrew",
			Title:       "Homebrew & Packages",
			Icon:        "▶",
			Description: "Install Homebrew and packages from Brewfile",
			Items:       []string{
				"Homebrew package manager",
				fmt.Sprintf("Brewfile locations: %d paths configured", len(config.Homebrew.BrewfilePaths)),
				"Packages from: " + strings.Join(config.Homebrew.BrewfilePaths, ", "),
			},
			EstTime: 15 * time.Minute,
			Status:  StatusReady,
			Enabled: true,
		})
	}
	
	// Terminal step
	if config.Terminal.Install {
		var terminalFiles []string
		for src, dest := range config.Terminal.ConfigFiles {
			terminalFiles = append(terminalFiles, fmt.Sprintf("%s → %s", src, dest))
		}
		
		steps = append(steps, SetupStep{
			ID:          "terminal",
			Title:       "Terminal Configuration", 
			Icon:        "▶",
			Description: "Configure terminal applications with Catppuccin theme",
			Items:       terminalFiles,
			EstTime:     2 * time.Minute,
			Status:      StatusReady,
			Enabled:     true,
		})
	}
	
	// Shell step
	if config.Shell.Install {
		shellItems := []string{
			fmt.Sprintf("Required tools: %s", strings.Join(config.Shell.RequiredTools, ", ")),
			fmt.Sprintf("Shell files: %s", strings.Join(config.Shell.ShellFiles, ", ")),
			fmt.Sprintf("Theme file: %s", config.Shell.ThemeFile),
			fmt.Sprintf("Init commands: %d configured", len(config.Shell.InitCommands)),
		}
		
		steps = append(steps, SetupStep{
			ID:          "shell",
			Title:       "Shell & Prompt Setup",
			Icon:        "▶", 
			Description: "Configure Zsh with Oh-My-Posh and productivity tools",
			Items:       shellItems,
			EstTime:     3 * time.Minute,
			Status:      StatusReady,
			Enabled:     true,
		})
	}
	
	// DevTools step
	if config.DevTools.Install {
		var devItems []string
		for _, lang := range config.DevTools.Languages {
			status := "disabled"
			if lang.Enabled {
				status = "enabled"
			}
			devItems = append(devItems, fmt.Sprintf("%s: %s (%d commands)", lang.Name, status, len(lang.Commands)))
		}
		devItems = append(devItems, fmt.Sprintf("Verify tools: %s", strings.Join(config.DevTools.VerifyTools, ", ")))
		
		steps = append(steps, SetupStep{
			ID:          "devtools",
			Title:       "Development Tools",
			Icon:        "▶",
			Description: "Configure development environments and toolchains", 
			Items:       devItems,
			EstTime:     5 * time.Minute,
			Status:      StatusReady,
			Enabled:     true,
		})
	}
	
	// Dotfiles step
	if config.Dotfiles.Install {
		var dotfileItems []string
		for src, dest := range config.Dotfiles.Mappings {
			dotfileItems = append(dotfileItems, fmt.Sprintf("%s → %s", src, dest))
		}
		
		steps = append(steps, SetupStep{
			ID:          "dotfiles",
			Title:       "Restore Dotfiles",
			Icon:        "▶",
			Description: "Copy configuration files to their destinations",
			Items:       dotfileItems,
			EstTime:     1 * time.Minute,
			Status:      StatusReady,
			Enabled:     true,
		})
	}
	
	// Always add verify step
	steps = append(steps, SetupStep{
		ID:          "verify",
		Title:       "Verify Installation",
		Icon:        "▶",
		Description: "Test that all tools are properly installed and accessible",
		Items:       []string{
			"Check tool availability in PATH",
			"Validate configurations", 
			"Generate installation report",
		},
		EstTime: 2 * time.Minute,
		Status:  StatusReady,
		Enabled: true,
	})
	
	return steps
}

// Helper functions for configuration and reporting
func getTotalConfiguredSteps(config *InstallConfig) int {
	count := 0
	if config.Homebrew.Install { count++ }
	if config.Terminal.Install { count++ }
	if config.Shell.Install { count++ }
	if config.DevTools.Install { count++ }
	if config.Dotfiles.Install { count++ }
	count++ // Always include verify step
	return count
}

func getStepDisplayName(stepID string) string {
	switch stepID {
	case "homebrew": return "Homebrew & Packages"
	case "terminal": return "Terminal Configuration" 
	case "shell": return "Shell & Prompt Setup"
	case "devtools": return "Development Tools"
	case "dotfiles": return "Restore Dotfiles"
	case "verify": return "Verify Installation"
	default: return stepID
	}
}

// KeyBinding represents keyboard shortcuts
type KeyBinding struct {
	QWERTY    []string
	ColemakDH []string
	Action    string
}

// GetKeyBindings returns the keyboard shortcuts for both layouts
func GetKeyBindings() []KeyBinding {
	return []KeyBinding{
		{
			QWERTY:    []string{"↑", "k"},
			ColemakDH: []string{"↑", "u"},
			Action:    "Navigate up",
		},
		{
			QWERTY:    []string{"↓", "j"},
			ColemakDH: []string{"↓", "e"},
			Action:    "Navigate down",
		},
		{
			QWERTY:    []string{"→", "l"},
			ColemakDH: []string{"→", "i"},
			Action:    "Navigate right",
		},
		{
			QWERTY:    []string{"←", "h"},
			ColemakDH: []string{"←", "n"},
			Action:    "Navigate left",
		},
		{
			QWERTY:    []string{"Enter", "Space"},
			ColemakDH: []string{"Enter", "Space"},
			Action:    "Select/Toggle",
		},
		{
			QWERTY:    []string{"Tab"},
			ColemakDH: []string{"Tab"},
			Action:    "Next pane",
		},
		{
			QWERTY:    []string{"q", "Esc"},
			ColemakDH: []string{"q", "Esc"},
			Action:    "Quit",
		},
	}
}

// FormatEstimatedTime formats duration in a human-readable way
func FormatEstimatedTime(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	minutes := int(d.Minutes())
	if minutes == 1 {
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", minutes)
}