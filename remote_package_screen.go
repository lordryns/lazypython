package main

import (
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

func updateRemotePackageTable(m *model, pkgs []string) {
	var rows []table.Row
	for _, pkg := range pkgs {
		rows = append(rows, table.Row{pkg})
	}
	m.remotePackageTable.SetRows(rows)
}

func drawPythonRemotePackagesTable(m *model, pkgs []string) {
	columns := []table.Column{
		{Title: "Package", Width: m.window.width / 2},
	}

	var rows []table.Row
	for _, pack := range pkgs {
		rows = append(rows, table.Row{pack})
	}

	m.remotePackageTable = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(m.window.height/2),
	)

	style := table.DefaultStyles()
	style.Header = style.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	style.Selected = style.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	m.remotePackageTable.SetStyles(style)
}

func drawPythonPackageTable(m *model, pman pythonManager) {
	columns := []table.Column{
		{Title: "Package", Width: ((m.window.width + 10) / 2) / 2},
		{Title: "Version", Width: ((m.window.width / 2) / 2) / 2},
	}

	var rows []table.Row
	for _, pack := range pman.packages {
		rows = append(rows, table.Row{pack.path, pack.version})
	}

	m.packageTable = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithWidth(m.window.width/2),
		table.WithHeight(m.window.height/2),
	)

	style := table.DefaultStyles()
	style.Header = style.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	style.Selected = style.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	m.packageTable.SetStyles(style)
}

func drawPythonScriptsTable(m *model, pman pythonManager) {
	columns := []table.Column{
		{Title: "Script Name", Width: m.window.width / 2},
		{Title: "Lines", Width: ((m.window.width / 2) / 2) / 2},
		{Title: "Functions", Width: ((m.window.width / 2) / 2) / 2},
		{Title: "Classes", Width: ((m.window.width / 2) / 2) / 2},
	}

	var rows []table.Row
	for _, script := range pman.scripts {
		rows = append(rows, table.Row{script.path, strconv.Itoa(script.lines), strconv.Itoa(script.functions), strconv.Itoa(script.classes)})
	}

	m.pythonScriptTable = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(m.window.height/2),
	)

	style := table.DefaultStyles()
	style.Header = style.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	style.Selected = style.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	m.pythonScriptTable.SetStyles(style)
}
