package tui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	CheckedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	ProgressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("63"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)
