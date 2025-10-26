package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func drawHomeScreen(m *model) string {
	if m.window.width < 40 {
		m.window.width = 40
	}
	if m.window.height < 20 {
		m.window.height = 20
	}

	var borderColor = lipgloss.Color("63")

	var infoText = lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		Render(m.info)
	var helpText = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render("Press Ctrl + H for help")

	var headerWidth = m.window.width - 4
	if headerWidth < 20 {
		headerWidth = 20
	}

	var header = lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().
			Width(headerWidth/2).
			Align(lipgloss.Left).
			Render(infoText),
		lipgloss.NewStyle().
			Width(headerWidth/2).
			Align(lipgloss.Right).
			Render(helpText),
	)
	header = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(m.window.width - 2).
		Render(header)

	var headerHeight = lipgloss.Height(header)

	var footerHeight = 3
	var availableHeight = m.window.height - headerHeight - footerHeight

	mainHeight := int(float64(availableHeight) * 0.60)
	if mainHeight < 8 {
		mainHeight = 8
	}

	var halfWidth = m.window.width/2 - 3
	if halfWidth < 15 {
		halfWidth = 15
	}

	var tableView = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(halfWidth).
		Height(mainHeight).
		Render(m.packageTable.View())

	var infoFrame = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(halfWidth).
		Height(mainHeight).
		Render(fmt.Sprintf(
			"Python Version: %v\nInstalled Packages: %v\nPackage Manager: %v",
			getPythonVersion(), len(m.localPackages), m.managerInUse,
		))

	var mainContent = lipgloss.JoinHorizontal(
		lipgloss.Top,
		tableView,
		infoFrame,
	)

	var mainContentHeight = lipgloss.Height(mainContent)

	scriptHeight := availableHeight - mainContentHeight - 2
	if scriptHeight < 5 {
		scriptHeight = 5
	}

	m.pythonScriptTable.SetHeight(scriptHeight)

	var pythonScriptsSection = lipgloss.NewStyle().
		MarginTop(0).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(m.window.width - 2).
		Height(scriptHeight).
		Render(m.pythonScriptTable.View())

	var footerText = lipgloss.NewStyle().
		Width(m.window.width - 2).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("240")).
		Render("j k Navigate | Tab Switch | Ctrl+C Quit")

	var footer = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(m.window.width - 2).
		Render(footerText)

	var screen = lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		mainContent,
		pythonScriptsSection,
		footer,
	)

	return screen
}
