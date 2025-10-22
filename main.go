package main

import (
	"fmt"
	"math/rand"

	"github.com/charmbracelet/bubbles/spinner"
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
	spinner          spinner.Model
}

func updateSpinnerType(m *model) {
	var spinners = []spinner.Spinner{spinner.Dot, spinner.Globe, spinner.Line, spinner.MiniDot, spinner.Jump, spinner.Ellipsis, spinner.Hamburger, spinner.Meter, spinner.Monkey, spinner.Moon, spinner.Points, spinner.Pulse}
	m.spinner.Spinner = spinners[rand.Intn(len(spinners)-1)]
}

func initialize() model {
	var _spinner = spinner.New()
	_spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	return model{spinner: _spinner}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
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
		updateSpinnerType(&m)
		drawPythonPackageTable(&m, msg.pacman)
		m.err = msg.err
		m.loadingState = false
		m.showPackageTable = true

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
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
		return lipgloss.NewStyle().Width(m.window.width).Height(m.window.height).Align(lipgloss.Center, lipgloss.Center).Render(fmt.Sprintf("%s Loading...", m.spinner.View()))
	}

	var infoText = lipgloss.NewStyle().Align(lipgloss.Left).Render("Hello from Lazypython!")
	var helpText = lipgloss.NewStyle().Align(lipgloss.Right).Render("Use Ctrl + c to quit")

	jointText := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(m.window.width/2).Render(infoText),
		lipgloss.NewStyle().Width(m.window.width/2).Align(lipgloss.Right).Render(helpText),
	)

	var bottomText = lipgloss.NewStyle().Width(m.window.width).Height(m.window.height / 2).AlignVertical(lipgloss.Bottom).Render(jointText)

	var tableAndHeader = lipgloss.NewStyle().Width(m.window.width).Align(lipgloss.Center).
		Render(fmt.Sprintf("%v\n%v", getPythonVersion(), m.packageTable.View()))

	return fmt.Sprintf("%v\n%v", tableAndHeader, bottomText)

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
