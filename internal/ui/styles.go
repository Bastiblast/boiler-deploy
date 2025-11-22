package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette
	primaryColor   = lipgloss.Color("#00D9FF")
	secondaryColor = lipgloss.Color("#7D56F4")
	accentColor    = lipgloss.Color("#FF6B9D")
	successColor   = lipgloss.Color("#00FF87")
	warningColor   = lipgloss.Color("#FFD700")
	errorColor     = lipgloss.Color("#FF5555")
	mutedColor     = lipgloss.Color("#626262")

	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(secondaryColor).
			Padding(0, 1).
			MarginBottom(1)

	// Menu item styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true).
				PaddingLeft(2)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			PaddingLeft(4)

	// Info box style
	infoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2).
			MarginTop(1)

	// Help style
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Success style
	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	// Info/subtitle style
	infoStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginBottom(1)
)
