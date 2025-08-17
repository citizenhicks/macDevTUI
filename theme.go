package main

import "github.com/charmbracelet/lipgloss"

// CatppuccinMocha colors
var (
	Rosewater = "#f5e0dc"
	Flamingo  = "#f2cdcd"
	Pink      = "#f5c2e7"
	Mauve     = "#cba6f7"
	Red       = "#f38ba8"
	Maroon    = "#eba0ac"
	Peach     = "#fab387"
	Yellow    = "#f9e2af"
	Green     = "#a6e3a1"
	Teal      = "#94e2d5"
	Sky       = "#89dceb"
	Sapphire  = "#74c7ec"
	Blue      = "#89b4fa"
	Lavender  = "#b4befe"
	Text      = "#cdd6f4"
	Subtext1  = "#bac2de"
	Subtext0  = "#a6adc8"
	Overlay2  = "#9399b2"
	Overlay1  = "#7f849c"
	Overlay0  = "#6c7086"
	Surface2  = "#585b70"
	Surface1  = "#45475a"
	Surface0  = "#313244"
	Base      = "#1e1e2e"
	Mantle    = "#181825"
	Crust     = "#11111b"
)

// Theme styles
var (
	// Base styles
	baseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Text))

	// Header style
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Mauve)).
			Bold(true).
			Padding(0, 1).
			Margin(0, 0, 1, 0)

	// Navigation pane styles
	navPaneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Text)).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(Surface2))

	navItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Subtext1)).
			Padding(0, 1)

	navItemSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Mauve)).
				Bold(true).
				Padding(0, 1)

	// Detail pane styles
	detailPaneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Text)).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(Surface2))

	detailTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Blue)).
				Bold(true).
				Margin(0, 0, 1, 0)

	detailBoxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Text)).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(Overlay0)).
			Margin(1, 0)

	// Status styles
	statusReadyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Yellow)).
				Bold(true)

	statusProgressStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Blue)).
				Bold(true)

	statusCompleteStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Green)).
				Bold(true)

	statusErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Red)).
				Bold(true)

	// Footer style
	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Subtext0)).
			Padding(0, 1).
			Margin(1, 0, 0, 0)

	// Progress bar styles
	progressBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Green))

	progressBarEmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Surface2))

	// Status message style
	statusMessageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Blue)).
			Italic(true).
			Margin(0, 0, 1, 0)

	// Notification banner styles (tab-like)
	notificationBannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Crust)).
			Background(lipgloss.Color(Peach)).
			Padding(0, 1).
			Bold(true)

	notificationBannerSuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Crust)).
			Background(lipgloss.Color(Green)).
			Padding(0, 1).
			Bold(true)

	notificationBannerErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Crust)).
			Background(lipgloss.Color(Red)).
			Padding(0, 1).
			Bold(true)
)

