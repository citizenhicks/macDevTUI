package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Version information (set by build process)
var (
	Version   = "dev"      // Set by -ldflags "-X main.Version=..."
	BuildTime = "unknown"  // Set by -ldflags "-X main.BuildTime=..."
	GitCommit = "unknown"  // Set by -ldflags "-X main.GitCommit=..."
)

// KeyboardLayout represents the keyboard layout preference
type KeyboardLayout int

const (
	QWERTY KeyboardLayout = iota
	ColemakDH
)

// Notification represents a popup notification
type Notification struct {
	Title   string
	Message string
	Type    string // "info", "success", "error"
}

// Model represents the main application state
type Model struct {
	steps           []SetupStep
	selectedStep    int
	keyboardLayout  KeyboardLayout
	width           int
	height          int
	installing      bool
	showHelp        bool
	currentProgress int    // 0-100
	currentMessage  string // What's happening now
	config          *InstallConfig
	notification    *Notification // Current notification to show
}

// NewModel creates a new application model
func NewModel() Model {
	config, err := LoadConfig()
	var steps []SetupStep
	var notification *Notification
	
	if err != nil {
		// Show config error as notification
		notification = &Notification{
			Title:   "Configuration Error",
			Message: fmt.Sprintf("Failed to load configuration: %s\nPress q to quit", err.Error()),
			Type:    "error",
		}
		// Create empty steps to avoid crashes
		steps = []SetupStep{}
	} else {
		steps = getConfigurableSteps(config)
	}

	return Model{
		steps:           steps,
		selectedStep:    0,
		keyboardLayout:  ColemakDH, // Default to QWERTY
		width:           0,       // Will be set by tea.WindowSizeMsg
		height:          0,       // Will be set by tea.WindowSizeMsg
		installing:      false,
		showHelp:        false,
		currentProgress: 0,
		currentMessage:  "Ready to install",
		config:          config,
		notification:    notification,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeypress(msg)

	case InstallMsg:
		// Handle installation progress messages
		if msg.StepID != "" {
			for i, step := range m.steps {
				if step.ID == msg.StepID {
					m.steps[i].Status = msg.Status
					if msg.Error != nil {
						m.steps[i].Error = msg.Error.Error()
					}
					break
				}
			}
		}

		// Handle errors by showing notification
		if msg.Error != nil {
			m.installing = false
			m.notification = &Notification{
				Title:   "Installation Error",
				Message: fmt.Sprintf("Step '%s' failed: %s\nPress Enter to dismiss", msg.StepID, msg.Error.Error()),
				Type:    "error",
			}
		}

		// Update progress and message
		if msg.Progress > 0 {
			m.currentProgress = msg.Progress
		}
		if msg.Message != "" {
			m.currentMessage = msg.Message
		}

		// Check if installation is complete
		if strings.Contains(msg.Message, "complete!") {
			m.installing = false
			m.currentProgress = 100
			// Mark all enabled steps as complete
			for i, step := range m.steps {
				if step.Enabled {
					m.steps[i].Status = StatusComplete
				}
			}
			// Show completion notification
			m.notification = &Notification{
				Title:   "Installation Complete!",
				Message: fmt.Sprintf("Report saved to: ./macdevtui-report.md\nPress Enter to dismiss"),
				Type:    "success",
			}
		}
		return m, nil
	}

	return m, nil
}

// handleKeypress processes keyboard input
func (m Model) handleKeypress(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	// Handle notification dismissal first
	if m.notification != nil && (key == "enter" || key == "esc") {
		m.notification = nil
		return m, nil
	}


	// Global shortcuts
	switch key {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "?":
		m.showHelp = !m.showHelp
		return m, nil
	case "c":
		// Toggle keyboard layout
		if m.keyboardLayout == QWERTY {
			m.keyboardLayout = ColemakDH
		} else {
			m.keyboardLayout = QWERTY
		}
		return m, nil
	case "s", "S":
		// START installation (only if no config errors)
		if !m.installing && m.notification == nil {
			m.installing = true
			return m, m.StartInstallation()
		}
		return m, nil
	}

	// Navigation based on keyboard layout
	if m.keyboardLayout == QWERTY {
		switch key {
		case "up", "k":
			if m.selectedStep > 0 {
				m.selectedStep--
			}
		case "down", "j":
			if m.selectedStep < len(m.steps)-1 {
				m.selectedStep++
			}
		case "enter", " ":
			return m.toggleStep()
		}
	} else { // Colemak-DH
		switch key {
		case "up", "u":
			if m.selectedStep > 0 {
				m.selectedStep--
			}
		case "down", "e":
			if m.selectedStep < len(m.steps)-1 {
				m.selectedStep++
			}
		case "enter", " ":
			return m.toggleStep()
		}
	}

	return m, nil
}

// toggleStep toggles the enabled state of the current step
func (m Model) toggleStep() (Model, tea.Cmd) {
	if m.selectedStep >= 0 && m.selectedStep < len(m.steps) {
		m.steps[m.selectedStep].Enabled = !m.steps[m.selectedStep].Enabled
	}
	return m, nil
}

// View implements tea.Model
func (m Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	// Wait for window size message
	if m.width == 0 || m.height == 0 {
		return "Loading..." // Wait for window size message
	}

	// Handle very small terminals
	if m.width < 50 || m.height < 10 {
		return "Terminal too small. Please resize to at least 50x10."
	}

	// Calculate pane dimensions responsively
	var navWidth int
	if m.width < 80 {
		// Small terminals: use fixed minimal navigation
		navWidth = 25
	} else if m.width < 120 {
		// Medium terminals: use proportional with minimum
		navWidth = m.width / 4
		if navWidth < 25 {
			navWidth = 25
		}
	} else {
		// Large terminals: use 1/3 but cap at reasonable size
		navWidth = m.width / 3
		if navWidth > 50 {
			navWidth = 50
		}
	}
	
	detailWidth := m.width - navWidth - 6 // Account for borders and spacing
	if detailWidth < 20 {
		// If terminal is too narrow, adjust navigation width
		detailWidth = 20
		navWidth = m.width - detailWidth - 6
	}

	headerText := fmt.Sprintf("MacDevTUI v%s - Mac Development Environment Installer", Version)
	header := headerStyle.Render(headerText)

	// Render notification banner if present
	var notificationBanner string
	if m.notification != nil {
		notificationBanner = m.renderNotificationBanner()
	}
	
	// Calculate content height accounting for notification banner
	contentHeight := m.height - 6 // Account for header and footer
	if notificationBanner != "" {
		contentHeight -= 1 // Account for notification banner height
	}

	// Render navigation pane
	navContent := m.renderNavigation(contentHeight)
	navPane := navPaneStyle.Width(navWidth).Height(contentHeight).Render(navContent)

	// Render detail pane
	detailContent := m.renderDetails(detailWidth)
	detailPane := detailPaneStyle.Width(detailWidth).Height(contentHeight).Render(detailContent)

	// Combine panes horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, navPane, detailPane)

	// Render footer
	footer := m.renderFooter()

	// Combine all sections vertically
	var sections []string
	sections = append(sections, header)
	if notificationBanner != "" {
		sections = append(sections, notificationBanner)
	}
	sections = append(sections, content, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderNavigation renders the left navigation pane
func (m Model) renderNavigation(height int) string {
	var items []string

	for i, step := range m.steps {
		icon := "○"
		if step.Enabled {
			icon = "●"
		}

		status := ""
		switch step.Status {
		case StatusComplete:
			status = " ✓"
		case StatusInProgress:
			status = " ◐"
		case StatusError:
			status = " ✗"
		}

		text := fmt.Sprintf("%s %s%s", icon, step.Title, status)

		if i == m.selectedStep {
			items = append(items, navItemSelectedStyle.Render("▶ "+text))
		} else {
			items = append(items, navItemStyle.Render("  "+text))
		}
	}

	return strings.Join(items, "\n")
}

// renderDetails renders the right detail pane
func (m Model) renderDetails(paneWidth int) string {
	if m.selectedStep < 0 || m.selectedStep >= len(m.steps) {
		return "Invalid selection"
	}

	step := m.steps[m.selectedStep]

	// Title
	title := detailTitleStyle.Render(fmt.Sprintf("%s %s", step.Icon, step.Title))

	// Description
	description := step.Description

	// Items list
	var itemsList []string
	for _, item := range step.Items {
		itemsList = append(itemsList, "• "+item)
	}
	itemsContent := strings.Join(itemsList, "\n")
	// Make the box responsive to available width
	responsiveBoxStyle := detailBoxStyle.Width(paneWidth - 8) // Account for pane padding and border
	itemsBox := responsiveBoxStyle.Render(itemsContent)

	// Status and timing
	statusText := fmt.Sprintf("Status: %s", step.Status.String())
	if step.Status == StatusReady {
		statusText = statusReadyStyle.Render(statusText)
	} else if step.Status == StatusInProgress {
		statusText = statusProgressStyle.Render(statusText)
	} else if step.Status == StatusComplete {
		statusText = statusCompleteStyle.Render(statusText)
	} else if step.Status == StatusError {
		statusText = statusErrorStyle.Render(statusText)
	}

	timeText := fmt.Sprintf("Est. time: %s", FormatEstimatedTime(step.EstTime))

	statusInfo := lipgloss.JoinVertical(lipgloss.Left,
		statusText,
		timeText,
	)

	// Error message if present
	errorMsg := ""
	if step.Error != "" {
		errorMsg = statusErrorStyle.Render("Error: " + step.Error)
	}

	// Progress bar (if installing) - use detail pane width
	progressBar := m.renderProgressBar(paneWidth - 4) // Use detail pane width minus padding

	// Combine all sections
	sections := []string{title, description, itemsBox, statusInfo}
	if errorMsg != "" {
		sections = append(sections, errorMsg)
	}
	if progressBar != "" {
		sections = append(sections, progressBar)
	}

	return strings.Join(sections, "\n\n")
}

// renderFooter renders the bottom instruction bar
func (m Model) renderFooter() string {
	layout := "QWERTY"
	if m.keyboardLayout == ColemakDH {
		layout = "Colemak-DH"
	}

	var keys string
	
	// Check if content overflows and scrolling is available
	contentOverflows := false
	// Skip overflow check in footer to avoid recursion
	
	if m.installing {
		keys = "Installation in progress... • q: Quit"
	} else if m.notification != nil && m.notification.Type == "error" {
		keys = "Configuration error - Installation disabled • q: Quit"
	} else if contentOverflows {
		keys = "↑/↓: Scroll • j/k: Navigate steps • Space: Toggle • S: START • q: Quit"
	} else if m.keyboardLayout == QWERTY {
		keys = "↑/↓ or k/j: Navigate • Space: Toggle • S: START • c: Layout • ?: Help • q: Quit"
	} else {
		keys = "↑/↓ or u/e: Navigate • Space: Toggle • S: START • c: Layout • ?: Help • q: Quit"
	}

	footerText := fmt.Sprintf("%s | %s", layout, keys)
	return footerStyle.Width(m.width - 2).Render(footerText)
}

// renderNotificationBanner renders a tab-style notification banner
func (m Model) renderNotificationBanner() string {
	// Choose style based on notification type
	var style lipgloss.Style
	switch m.notification.Type {
	case "success":
		style = notificationBannerSuccessStyle
	case "error":
		style = notificationBannerErrorStyle
	default:
		style = notificationBannerStyle
	}

	// Create banner content (single line)
	content := fmt.Sprintf("%s - %s [Press Enter to dismiss]",
		m.notification.Title,
		strings.ReplaceAll(m.notification.Message, "\n", " "))

	// Render full-width banner
	return style.Width(m.width).Render(content)
}

// renderHelp renders the help screen
func (m Model) renderHelp() string {
	helpContent := []string{
		headerStyle.Render(fmt.Sprintf("Help - MacDevTUI v%s", Version)),
		"",
		fmt.Sprintf("Version: %s | Build: %s | Commit: %s", Version, BuildTime, GitCommit),
		"",
		"This tool helps you reinstall your Mac development environment with a consistent",
		"Catppuccin Mocha theme across all tools and configurations.",
		"",
		"Navigation:",
	}

	bindings := GetKeyBindings()
	for _, binding := range bindings {
		qwerty := strings.Join(binding.QWERTY, ", ")
		colemak := strings.Join(binding.ColemakDH, ", ")
		helpContent = append(helpContent,
			fmt.Sprintf("  %s: %s (QWERTY) | %s (Colemak-DH)",
				binding.Action, qwerty, colemak))
	}

	helpContent = append(helpContent, []string{
		"",
		"Additional Commands:",
		"  c: Toggle keyboard layout",
		"  ?: Show/hide this help",
		"",
		"Steps:",
		"  • Use ○/● to toggle steps on/off",
		"  • Selected steps will be installed in order",
		"  • Press Space or Enter to toggle the current step",
		"",
		"Press ? again to close help, or q to quit.",
	}...)

	return strings.Join(helpContent, "\n")
}

// renderProgressBar creates a visual progress bar
func (m Model) renderProgressBar(width int) string {
	if !m.installing {
		return ""
	}
	
	barWidth := width - 4 // Account for brackets and padding
	if barWidth < 10 {
		barWidth = 10
	}
	
	filledCount := int(float64(barWidth) * float64(m.currentProgress) / 100.0)
	emptyCount := barWidth - filledCount
	
	filled := progressBarStyle.Render(strings.Repeat("█", filledCount))
	empty := progressBarEmptyStyle.Render(strings.Repeat("░", emptyCount))
	
	progressText := fmt.Sprintf("%d%%", m.currentProgress)
	message := statusMessageStyle.Render(m.currentMessage)
	
	bar := fmt.Sprintf("[%s%s] %s", filled, empty, progressText)
	
	return lipgloss.JoinVertical(lipgloss.Left, bar, message)
}

func main() {
	// Set up signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	
	// Handle signals in a goroutine
	go func() {
		<-c
		p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}

