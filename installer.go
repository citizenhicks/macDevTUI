package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	currentDir, _ = os.Getwd()
	homeDir, _    = os.UserHomeDir()
	logger        *log.Logger
)

// initLogger sets up logging to a file
func initLogger() {
	logFile, err := os.OpenFile(filepath.Join(currentDir, "macdevtui.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file:", err)
	}
	logger = log.New(logFile, "", log.LstdFlags)
	logger.Println("=== MacDevTUI Session Started ===")
}

// expandPath expands path variables like $HOME, ~, and {{.HOME}}
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}
	if path == "~" {
		return homeDir
	}

	if strings.Contains(path, "$HOME") {
		path = strings.ReplaceAll(path, "$HOME", homeDir)
	}

	if strings.Contains(path, "{{.HOME}}") {
		path = strings.ReplaceAll(path, "{{.HOME}}", homeDir)
	}

	return path
}

// expandPaths expands path variables in a slice of strings
func expandPaths(paths []string) []string {
	expanded := make([]string, len(paths))
	for i, path := range paths {
		expanded[i] = expandPath(path)
	}
	return expanded
}

// expandCommands expands path variables in command arguments
func expandCommands(commands [][]string) [][]string {
	expanded := make([][]string, len(commands))
	for i, cmd := range commands {
		expanded[i] = expandPaths(cmd)
	}
	return expanded
}

// InstallMsg represents an installation progress message
type InstallMsg struct {
	StepID   string
	Status   InstallStatus
	Progress int
	Message  string
	Error    error
}

// DotfilesStatus tracks dotfiles installation status
type DotfilesStatus struct {
	CopiedFiles    []string
	MissingFiles   []string
	IsCleanInstall bool
}

// Global variables to track executed steps and dotfiles status
var (
	executedSteps  []string
	dotfilesStatus DotfilesStatus
)

// StartInstallation begins the installation process for enabled steps
func (m Model) StartInstallation() tea.Cmd {
	return func() tea.Msg {
		// Initialize logger
		initLogger()
		logger.Println("Starting installation process")
		
		// Reset executed steps tracking
		executedSteps = []string{}

		// Process each enabled step sequentially
		for _, step := range m.steps {
			if !step.Enabled {
				logger.Printf("Skipping disabled step: %s", step.ID)
				continue
			}

			// Track that this step is being executed
			logger.Printf("Executing step: %s", step.ID)
			executedSteps = append(executedSteps, step.ID)

			var err error
			switch step.ID {
			case "homebrew":
				err = installHomebrew()
			case "terminal":
				err = configureTerminal()
			case "shell":
				err = configureShell()
			case "devtools":
				err = installDevTools()
			case "dotfiles":
				err = restoreDotfiles()
			case "verify":
				err = verifyInstallation()
			}

			if err != nil {
				logger.Printf("Step %s failed: %v", step.ID, err)
				return InstallMsg{
					StepID:  step.ID,
					Status:  StatusError,
					Error:   err,
					Message: fmt.Sprintf("Failed: %s", err.Error()),
				}
			} else {
				logger.Printf("Step %s completed successfully", step.ID)
			}
		}

		// All steps completed successfully
		message := "All installations complete!"
		if dotfilesStatus.IsCleanInstall && len(dotfilesStatus.MissingFiles) > 0 {
			message = fmt.Sprintf("Clean install complete! (%d dotfiles not found)", len(dotfilesStatus.MissingFiles))
		}

		// Generate report after all installations complete
		logger.Println("All installations complete, generating report")
		generateReportAfterInstallation(executedSteps)

		return InstallMsg{
			Message:  message,
			Progress: 100,
		}
	}
}

// installHomebrew installs Homebrew and packages
func installHomebrew() error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !config.Homebrew.Install {
		return nil // Skip if disabled
	}

	// Check if Homebrew is already installed
	if _, err := exec.LookPath("brew"); err != nil {
		// Install Homebrew
		cmd := exec.Command("/bin/bash", "-c",
			`/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install Homebrew: %w", err)
		}
	}

	// Use configured Brewfile paths (expand any path variables)
	for _, brewPath := range config.Homebrew.BrewfilePaths {
		expandedPath := expandPath(brewPath)
		if _, err := os.Stat(expandedPath); err == nil {
			cmd := exec.Command("brew", "bundle", "--file="+expandedPath)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to install packages from Brewfile %s: %w", expandedPath, err)
			}
			return nil
		}
	}
	return fmt.Errorf("no Brewfile found in expected locations: %v", config.Homebrew.BrewfilePaths)
}

// configureTerminal sets up Kitty and Tmux configurations
func configureTerminal() error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !config.Terminal.Install {
		return nil // Skip if disabled
	}

	// Copy configured terminal files
	for srcRelPath, destRelPath := range config.Terminal.ConfigFiles {
		srcPath := filepath.Join(currentDir, srcRelPath)
		destPath := filepath.Join(homeDir, destRelPath)

		// Create destination directory
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
		}

		// Copy file
		if err := copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to copy %s to %s: %w", srcPath, destPath, err)
		}
	}

	return nil
}

// configureShell sets up Zsh with Oh-My-Posh and tools
func configureShell() error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !config.Shell.Install {
		return nil // Skip if disabled
	}

	// Check if required tools are installed
	for _, tool := range config.Shell.RequiredTools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s is not installed: %w", tool, err)
		}
	}

	// Ensure .config directory exists
	if err := os.MkdirAll(filepath.Join(homeDir, ".config"), 0755); err != nil {
		return fmt.Errorf("failed to create .config directory: %w", err)
	}

	// Copy configured shell files
	for _, file := range config.Shell.ShellFiles {
		srcFile := filepath.Join(currentDir, file)
		destFile := filepath.Join(homeDir, file)
		if err := copyFile(srcFile, destFile); err != nil {
			return fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	// Copy Oh-My-Posh theme file
	if config.Shell.ThemeFile != "" {
		srcTheme := filepath.Join(currentDir, config.Shell.ThemeFile)
		destTheme := filepath.Join(homeDir, ".config", config.Shell.ThemeFile)
		if err := copyFile(srcTheme, destTheme); err != nil {
			return fmt.Errorf("failed to copy theme file: %w", err)
		}
	}

	// Run configured initialization commands (expand any path variables)
	expandedCommands := expandCommands(config.Shell.InitCommands)
	for _, cmd := range expandedCommands {
		if len(cmd) == 0 {
			continue
		}
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			return fmt.Errorf("failed to run command %v: %w", cmd, err)
		}
	}

	return nil
}

// installDevTools configures development environment
func installDevTools() error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !config.DevTools.Install {
		return nil // Skip if disabled
	}

	// Configure each enabled language
	for _, lang := range config.DevTools.Languages {
		if !lang.Enabled {
			continue // Skip disabled languages
		}

		for _, cmd := range lang.Commands {
			if len(cmd) == 0 {
				continue
			}

			if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
				return fmt.Errorf("failed to configure %s with command %v: %w", lang.Name, cmd, err)
			}
		}
	}

	// Run global tools installation
	for _, cmd := range config.DevTools.GlobalTools {
		if len(cmd) == 0 {
			continue
		}

		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			return fmt.Errorf("failed to install global tool %v: %w", cmd, err)
		}
	}

	// Verify all tools are accessible
	for _, tool := range config.DevTools.VerifyTools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s is not installed or not in PATH: %w", tool, err)
		}
	}

	return nil
}

// restoreDotfiles copies all configuration files
func restoreDotfiles() error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !config.Dotfiles.Install {
		return nil // Skip if disabled
	}

	var copiedFiles []string
	var missingFiles []string

	// Copy configured dotfiles
	for srcRelPath, destRelPath := range config.Dotfiles.Mappings {
		srcPath := filepath.Join(currentDir, srcRelPath)
		destPath := filepath.Join(homeDir, destRelPath)

		if info, err := os.Stat(srcPath); err == nil {
			if info.IsDir() {
				if err := copyDir(srcPath, destPath); err != nil {
					return fmt.Errorf("failed to copy directory %s to %s: %w", srcPath, destPath, err)
				}
			} else {
				if err := copyFile(srcPath, destPath); err != nil {
					return fmt.Errorf("failed to copy file %s to %s: %w", srcPath, destPath, err)
				}
			}
			copiedFiles = append(copiedFiles, srcRelPath)
		} else {
			missingFiles = append(missingFiles, srcRelPath)
		}
	}

	// Store dotfiles status for reporting
	dotfilesStatus = DotfilesStatus{
		CopiedFiles:    copiedFiles,
		MissingFiles:   missingFiles,
		IsCleanInstall: len(missingFiles) > 0,
	}

	return nil
}

// verifyInstallation checks that everything is working
func verifyInstallation() error {
	logger.Println("Starting verification step")
	config, err := LoadConfig()
	if err != nil {
		logger.Printf("Failed to load config for verification: %v", err)
		return fmt.Errorf("failed to load config: %w", err)
	}

	var allTools []string

	// Collect tools to verify from enabled sections
	if config.Homebrew.Install {
		allTools = append(allTools, "brew")
	}

	if config.Shell.Install {
		allTools = append(allTools, config.Shell.RequiredTools...)
	}

	if config.DevTools.Install {
		allTools = append(allTools, config.DevTools.VerifyTools...)
	}

	// Remove duplicates
	toolSet := make(map[string]bool)
	var uniqueTools []string
	for _, tool := range allTools {
		if !toolSet[tool] {
			toolSet[tool] = true
			uniqueTools = append(uniqueTools, tool)
		}
	}

	// Verify each tool
	var failures []string
	for _, tool := range uniqueTools {
		if _, err := exec.LookPath(tool); err != nil {
			failures = append(failures, tool)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("verification failed for: %s", strings.Join(failures, ", "))
	}

	// Generate installation report using actually executed steps
	// Add verify to executed steps since verification always runs
	actuallyExecuted := append([]string{}, executedSteps...) // Copy the global executed steps
	actuallyExecuted = append(actuallyExecuted, "verify")    // Add verify since it's running now
	generateInstallationReport(uniqueTools, actuallyExecuted)
	return nil
}

// generateReportAfterInstallation creates a report after installation completes
func generateReportAfterInstallation(executedSteps []string) {
	logger.Println("Starting report generation after installation")
	config, err := LoadConfig()
	if err != nil {
		logger.Printf("Failed to load config for report: %v", err)
		return
	}

	// Get verified tools based on what was executed
	var verifiedTools []string
	for _, stepID := range executedSteps {
		switch stepID {
		case "homebrew":
			if config.Homebrew.Install {
				verifiedTools = append(verifiedTools, "brew")
			}
		case "shell":
			if config.Shell.Install {
				verifiedTools = append(verifiedTools, config.Shell.RequiredTools...)
			}
		case "devtools":
			if config.DevTools.Install {
				verifiedTools = append(verifiedTools, config.DevTools.VerifyTools...)
			}
		}
	}

	generateInstallationReport(verifiedTools, executedSteps)
}

// generateInstallationReport creates a dynamic summary of what was actually installed
func generateInstallationReport(verifiedTools []string, executedSteps []string) {
	logger.Println("Starting report generation")
	config, err := LoadConfig()
	if err != nil {
		logger.Printf("Failed to load config for report: %v", err)
		return // Skip report if config fails
	}

	reportPath := filepath.Join(currentDir, "macdevtui-report.md")
	logger.Printf("Creating report at: %s", reportPath)

	report := []string{
		"# MacDevTUI Installation Report",
		"",
		"> *Mac development environment installer with Catppuccin theming*",
		"",
		fmt.Sprintf("**Generated:** %s", time.Now().Format("2006-01-02 15:04:05")),
		fmt.Sprintf("**Configuration:** `%s`", getCurrentConfigPath()),
		"",
		"## ✅ Verified Tools",
		"",
	}

	for _, tool := range verifiedTools {
		report = append(report, fmt.Sprintf("- `%s`", tool))
	}

	// Helper function to check if step was executed
	wasExecuted := func(stepID string) bool {
		for _, executed := range executedSteps {
			if executed == stepID {
				return true
			}
		}
		return false
	}

	// Add sections based on what was actually executed
	if config.Homebrew.Install && wasExecuted("homebrew") {
		report = append(report, []string{
			"",
			"## 🍺 Homebrew Packages",
			"",
			fmt.Sprintf("- Installed from Brewfile (searched %d locations)", len(config.Homebrew.BrewfilePaths)),
			"- Run `brew list` to see all installed packages",
		}...)
	}

	if config.Terminal.Install && wasExecuted("terminal") {
		report = append(report, []string{
			"",
			"## 💻 Terminal Configuration",
			"",
		}...)
		for src, dest := range config.Terminal.ConfigFiles {
			report = append(report, fmt.Sprintf("- `%s` → `~/%s`", src, dest))
		}
	}

	if config.Shell.Install && wasExecuted("shell") {
		report = append(report, []string{
			"",
			"## 🐚 Shell Setup",
			"",
			fmt.Sprintf("- **Required tools:** `%s`", strings.Join(config.Shell.RequiredTools, "`, `")),
			fmt.Sprintf("- **Shell files:** `%s`", strings.Join(config.Shell.ShellFiles, "`, `")),
			fmt.Sprintf("- **Theme:** `%s`", config.Shell.ThemeFile),
			fmt.Sprintf("- **Initialization commands:** %d executed", len(config.Shell.InitCommands)),
		}...)
	}

	if config.DevTools.Install && wasExecuted("devtools") {
		report = append(report, []string{
			"",
			"## 🔧 Development Tools",
			"",
		}...)

		for _, lang := range config.DevTools.Languages {
			if lang.Enabled {
				report = append(report, fmt.Sprintf("- **%s:** configured (%d commands executed)", lang.Name, len(lang.Commands)))
			}
		}
	}

	if config.Dotfiles.Install && wasExecuted("dotfiles") {
		report = append(report, []string{
			"",
			"## 📁 Dotfiles",
			"",
		}...)

		if dotfilesStatus.IsCleanInstall {
			report = append(report, "⚠️ **Clean Install** - Some dotfiles were not found")
			report = append(report, "")
		}

		if len(dotfilesStatus.CopiedFiles) > 0 {
			report = append(report, "### ✅ Copied Files")
			for _, file := range dotfilesStatus.CopiedFiles {
				report = append(report, fmt.Sprintf("- `%s` → `~/%s`", file, config.Dotfiles.Mappings[file]))
			}
			report = append(report, "")
		}

		if len(dotfilesStatus.MissingFiles) > 0 {
			report = append(report, "### ❌ Missing Files")
			for _, file := range dotfilesStatus.MissingFiles {
				report = append(report, fmt.Sprintf("- `%s` (not found)", file))
			}
			report = append(report, "")
		}
	}

	report = append(report, []string{
		"",
		"## 🚀 Next Steps",
		"",
		"1. **Restart your terminal** to load new configurations",
		"2. **Test your configured tools** and verify functionality",
		"3. **Customize further** by editing config files",
		"4. **Run health checks** for installed package managers",
		"",
		"## 📄 Files",
		"",
		fmt.Sprintf("- **Report:** `%s`", reportPath),
		fmt.Sprintf("- **Config:** `%s`", getCurrentConfigPath()),
		"",
		"---",
		"",
		"✨ **Installation completed successfully!** ✨",
	}...)

	content := strings.Join(report, "\n")
	err = os.WriteFile(reportPath, []byte(content), 0644)
	if err != nil {
		logger.Printf("Failed to write report: %v", err)
	} else {
		logger.Printf("Report successfully written to: %s", reportPath)
	}
}

// getCurrentConfigPath returns the path of the currently loaded config
func getCurrentConfigPath() string {
	configPaths := []string{
		filepath.Join(currentDir, "install-config.json"),
		filepath.Join(currentDir, "config.json"),
		filepath.Join(homeDir, ".config", "mac-reinstaller", "config.json"),
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "default configuration (no config file found)"
}

// Utility functions for file operations
func copyFile(src, dest string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	return os.WriteFile(dest, data, 0644)
}

func copyDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return copyFile(path, destPath)
	})
}
