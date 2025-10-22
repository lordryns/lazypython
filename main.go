package main

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
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
	packageTable                      table.Model
	showPackageTable                  bool
	window                            dimension
	err                               error
	loadingState                      bool
	spinner                           spinner.Model
	info                              string
	managerInUse                      string
	openHelpMenu                      bool
	openPackageInstallScreen          bool
	packageInput                      textinput.Model
	showHomeScreen                    bool
	filteredPackages                  []string
	remotePackageTable                table.Model
	focusedOnRemotePackageTable       bool
	remotePackageTableIndex           int
	remotePackagesIndexedSuccessfully bool
}

type InfoMsg string

func updateSpinnerType(m *model) {
	var spinners = []spinner.Spinner{spinner.Dot, spinner.Globe, spinner.Line, spinner.MiniDot, spinner.Jump, spinner.Ellipsis, spinner.Hamburger, spinner.Meter, spinner.Monkey, spinner.Moon, spinner.Points, spinner.Pulse}
	var gen = rand.Intn(len(spinners) - 1)
	m.spinner.Spinner = spinners[gen]
}

func initialize() model {

	var _spinner = spinner.New()
	_spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	var installEntry = textinput.New()
	installEntry.CharLimit = -1
	installEntry.Focus()
	installEntry.Placeholder = "Enter package name..."
	var m = model{spinner: _spinner, info: "Hello from Lazypython", packageInput: installEntry, showHomeScreen: true}
	updateSpinnerType(&m)

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchPackagesFromindexAsync(&m))
}

type LoadedPythonManager struct {
	pacman pythonManager
	err    error
}

func fetchPackagesFromindexAsync(m *model) tea.Cmd {
	return func() tea.Msg {
		fetchPackagesFromIndex()
		m.remotePackagesIndexedSuccessfully = true
		return InfoMsg("Remote packages indexed successfully!")
	}
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

		case "ctrl+h":
			m.openHelpMenu = !m.openHelpMenu
			m.showPackageTable = false
			if m.openHelpMenu {
				m.showHomeScreen = true
			}

		case "ctrl+p":
			m.openPackageInstallScreen = !m.openPackageInstallScreen
			m.showPackageTable = false
			if m.openPackageInstallScreen {
				m.showHomeScreen = true
			}

		case "esc":
			if m.openHelpMenu {
				m.openHelpMenu = false
			}
			if m.openPackageInstallScreen {
				m.openPackageInstallScreen = false
			}

			if !m.openHelpMenu || !m.openPackageInstallScreen {
				m.showHomeScreen = true
			}

		case "down":
			if m.openPackageInstallScreen {
				if m.packageInput.Focused() {
					m.packageInput.Blur()
					m.remotePackageTable.Focus()
					m.focusedOnRemotePackageTable = true
				}

			}

		case "up":
			if m.openPackageInstallScreen {
				if m.remotePackageTable.Cursor() < 1 && m.remotePackageTable.Focused() {
					m.focusedOnRemotePackageTable = false
					m.remotePackageTable.Blur()
					m.packageInput.Focus()
				}
			}

		case "tab":
			if m.openPackageInstallScreen {
				if m.packageInput.Focused() {
					m.packageInput.Blur()
					m.remotePackageTable.Focus()
				} else {
					m.packageInput.Focus()
					m.remotePackageTable.Blur()
				}

				m.focusedOnRemotePackageTable = !m.focusedOnRemotePackageTable
			}
		}

	case tea.WindowSizeMsg:
		m.window.width = msg.Width
		m.window.height = msg.Height

		if !m.showHomeScreen {
			return m, nil
		}
		m.loadingState = true
		return m, fetchPackagesAsync()

	case LoadedPythonManager:

		updateSpinnerType(&m)
		drawPythonPackageTable(&m, msg.pacman)
		m.err = msg.err
		if msg.err != nil {
			m.info = fmt.Sprintf("err: %v", msg.err.Error())
		}
		m.loadingState = false
		m.showPackageTable = true

	case InfoMsg:
		m.info = string(msg)

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)

		if !m.remotePackagesIndexedSuccessfully {
			m.info = fmt.Sprintf("%v Indexing remote packages on PYPI...", m.spinner.View())
		}
		if m.openPackageInstallScreen {
			if m.packageInput.Focused() {
				var query = strings.TrimSpace(m.packageInput.Value())
				m.filteredPackages = nil
				if query != "" {
					for _, pkg := range pythonPackages {
						if strings.Contains(pkg, query) {
							m.filteredPackages = append(m.filteredPackages, pkg)
						}
					}
				}
			}

			drawPythonRemotePackagesTable(&m, m.filteredPackages)
		}
		return m, cmd
	}

	var cmd tea.Cmd
	m.packageTable, cmd = m.packageTable.Update(msg)
	m.packageInput, cmd = m.packageInput.Update(msg)
	m.remotePackageTable, cmd = m.remotePackageTable.Update(msg)

	return m, cmd
}

func (m model) View() string {
	if m.window.width < 60 || m.window.height < 30 {
		return lipgloss.NewStyle().Width(m.window.width).Height(m.window.height).Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Terminal size too small!\nMust be at least 60, 30\nCurrent: %v, %v", m.window.width, m.window.height))

	}

	if m.loadingState {
		return lipgloss.NewStyle().Width(m.window.width).Height(m.window.height).Align(lipgloss.Center, lipgloss.Center).Render(fmt.Sprintf("%s Loading...", m.spinner.View()))
	}

	if m.openHelpMenu {
		return lipgloss.NewStyle().Width(m.window.width).Height(m.window.height).Align(lipgloss.Center, lipgloss.Center).
			Render("HELP\nUse Ctrl + h or the Esc key to close this screen\nCtrl + c to exit the application\nCtrl + p to find (and install) a package")
	}

	if m.openPackageInstallScreen {
		return drawPackageInstallScreen(&m)
	}

	if m.showHomeScreen {
		return drawHomeScreen(&m)
	}

	return lipgloss.NewStyle().Width(m.window.width).Height(m.window.height).Align(lipgloss.Center, lipgloss.Center).Render("Somehow this page showed up even though it isn't supposed to, press the Esc key to return to Home... restart if this persists.")
}

func drawHomeScreen(m *model) string {
	var infoText = lipgloss.NewStyle().Align(lipgloss.Left).Render(m.info)
	var helpText = lipgloss.NewStyle().Align(lipgloss.Right).Render("Use Ctrl + h to open help")

	jointText := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(m.window.width/2).Render(infoText),
		lipgloss.NewStyle().Width(m.window.width/2).Align(lipgloss.Right).Render(helpText),
	)

	var bottomText = lipgloss.NewStyle().Width(m.window.width).Height((m.window.height / 2) - 3).AlignVertical(lipgloss.Bottom).Render(jointText)

	var tableAndHeader = lipgloss.NewStyle().Width(m.window.width).Align(lipgloss.Center).
		Render(fmt.Sprintf("%v\n%v", getPythonVersion(), m.packageTable.View()))

	return fmt.Sprintf("%v\n%v", tableAndHeader, bottomText)
}
func drawPackageInstallScreen(m *model) string {
	m.packageInput.Width = m.window.width
	var inputStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Width(m.window.width - 5).
		Render(m.packageInput.View())

	return lipgloss.NewStyle().Width(m.window.width).AlignHorizontal(lipgloss.Center).
		Render(fmt.Sprintf("%v\n%v\n", inputStyle, m.remotePackageTable.View()))
}

func main() {
	if _, err := tea.NewProgram(initialize(), tea.WithAltScreen()).Run(); err != nil {
		panic(err)
	}
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
