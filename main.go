package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type pythonPackage struct {
	path    string
	version string
}

type pythonManager struct {
	version  string
	packages []pythonPackage
}

type dimension struct {
	width  int
	height int
}
type model struct {
	packageTable     table.Model
	showPackageTable bool
	window           dimension
	err              error
	loadingState     bool
}

func initialize() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return nil
}

type LoadedPythonManager struct {
	pacman pythonManager
	err    error
}

func fetchPackagesAsync() tea.Cmd {
	return func() tea.Msg {
		var pman, err = generatePackageDetails()
		return LoadedPythonManager{pacman: pman, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.window.width = msg.Width
		m.window.height = msg.Height

		m.loadingState = true
		return m, fetchPackagesAsync()

	case LoadedPythonManager:
		drawPythonPackageTable(&m, msg.pacman)
		m.err = msg.err
		m.loadingState = false
		m.showPackageTable = true
	}

	if m.showPackageTable {
		var cmd tea.Cmd
		m.packageTable, cmd = m.packageTable.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	// this is for window size logic, we're returning this if the window is too small
	if m.window.width < 60 || m.window.height < 30 {
		return lipgloss.NewStyle().Width(m.window.width).Height(m.window.height).Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Terminal size too small!\nMust be at least 60, 30\nCurrent: %v, %v", m.window.width, m.window.height))

	}

	if m.loadingState {
		return lipgloss.NewStyle().Width(m.window.width).Height(m.window.height).Align(lipgloss.Center, lipgloss.Center).Render("Loading...")
	}
	return lipgloss.NewStyle().Width(m.window.width).Height(m.window.height).Align(lipgloss.Center).
		Render(fmt.Sprintf("%v\n%v", getPythonVersion(), m.packageTable.View()))

}

func main() {
	if _, err := tea.NewProgram(initialize(), tea.WithAltScreen()).Run(); err != nil {
		panic(err)
	}
}

func drawPythonPackageTable(m *model, pman pythonManager) {
	columns := []table.Column{
		{Title: "Package", Width: m.window.width / 2},
		{Title: "Version", Width: (m.window.width / 2) / 2},
	}

	var rows []table.Row
	for _, pack := range pman.packages {
		rows = append(rows, table.Row{pack.path, pack.version})
	}

	m.packageTable = table.New(
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
	m.packageTable.SetStyles(style)
}
