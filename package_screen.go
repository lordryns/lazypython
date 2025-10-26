package main

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

func drawPackageInstallScreen(m *model) string {
	var header = lipgloss.NewStyle().
		Width(m.window.width - 10).
		Foreground(lipgloss.Color("229")).
		Bold(true).
		Align(lipgloss.Center).
		Render("Install Python Package")

	header = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Render(header)

	m.packageInput.Width = m.window.width - 27
	var inputBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		MarginTop(1).
		Render("Package name: " + m.packageInput.View())

	var tableHeight = m.window.height/2 - 20
	var tableBox = lipgloss.NewStyle().
		Width(m.window.width/2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Height(tableHeight).
		Render(m.remotePackageTable.View())

	var packageInfoBox = lipgloss.NewStyle().
		Width(m.window.width/2-10).
		Height((m.window.height/2)+1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Render(
			fmt.Sprintf(
				"Package name: %v\nPackage version: %v\n\nAuthor email: %v\n\nSummary: %v\n\nSize: %v bytes\n\nDownloads (Last Week): %v\n\nDownloads (Last Month): %v",
				m.remotePackageSelected.Info.Name,
				m.remotePackageSelected.Info.Version,
				m.remotePackageSelected.Info.AuthorEmail,
				m.remotePackageSelected.Info.Summary,
				func() string {
					if files, ok := m.remotePackageSelected.Releases[m.remotePackageSelected.Info.Version]; ok && len(files) > 0 {
						return strconv.Itoa(files[0].Size)
					}
					return "Unknown"
				}(),
				m.remotePackageSelected.Downloads.LastWeek,
				m.remotePackageSelected.Downloads.LastMonth,
			))

	var jointBox = lipgloss.JoinHorizontal(lipgloss.Center, tableBox, packageInfoBox)
	var footer = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Width(m.window.width - 8).
		Render(fmt.Sprintf("Type to filter | Ctrl+A to install | Esc to cancel * %s", lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(m.info)))

	screen := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		inputBox,
		jointBox,
		footer,
	)

	return lipgloss.NewStyle().
		Width(m.window.width).
		Height(m.window.height).
		Padding(1, 2).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Top).
		Render(screen)
}
