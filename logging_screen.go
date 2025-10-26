package main

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

func updateLoggingTable(m *model) {
	var rows []table.Row
	for _, log := range m.logs {
		rows = append(rows, table.Row{log.Level, log.Time, log.Message})
	}
	m.logTable.SetRows(rows)
}

func drawLoggingPage(m *model) string {
	if len(m.logTable.Columns()) == 0 {
		var columns = []table.Column{
			{Title: "Level", Width: (m.window.width / 2) / 2},
			{Title: "Time", Width: (m.window.width / 2) / 2},
			{Title: "Message", Width: m.window.width / 2},
		}

		var rows []table.Row
		for _, log := range m.logs {
			rows = append(rows, table.Row{log.Level, log.Time, log.Message})
		}

		m.logTable = table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(m.window.height-5),
		)

		var s = table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(true)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)

		m.logTable.SetStyles(s)
	}

	var header = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Padding(1, 0).
		Render("Application Logs")

	var footer = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(1, 0).
		Render("j/k: navigate • Esc: Home")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.logTable.View(),
		footer,
	)
}
