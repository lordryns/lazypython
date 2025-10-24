package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var remotePackagesIndexedSuccessfully = false

type pythonPackage struct {
	path    string
	version string
}

type pythonScript struct {
	path      string
	lines     int
	functions int
	classes   int
}

type pythonManager struct {
	version  string
	packages []pythonPackage
	scripts  []pythonScript
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
	localPackages                     []pythonPackage
	pythonScriptTable                 table.Model
	focusOnLocalPackageTable          bool
	remotePackageSelected             PackageInfo
}

type InfoMsg string

func updateSpinnerType(m *model) {
	var spinners = []spinner.Spinner{spinner.Dot, spinner.Globe, spinner.Line, spinner.MiniDot, spinner.Jump, spinner.Ellipsis, spinner.Meter, spinner.Monkey, spinner.Moon, spinner.Points, spinner.Pulse}
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
	var m = model{spinner: _spinner, info: "Hello from Lazypython", packageInput: installEntry, showHomeScreen: true, managerInUse: "pip", focusOnLocalPackageTable: true}
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
		m.info = "Done!"
		remotePackagesIndexedSuccessfully = true
		return InfoMsg("Remote packages indexed successfully!")
	}
}

func fetchPackagesAsync() tea.Cmd {
	return func() tea.Msg {
		var pman, err = generatePackageDetails()
		return LoadedPythonManager{pacman: pman, err: err}
	}
}

func runInstallCommandAndRespondAsync(m *model) tea.Cmd {
	return func() tea.Msg {
		var res = runInstallCommandAndRespond(m.managerInUse, m.remotePackageTable.SelectedRow()[0])
		return res
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

		case "p":
			if m.showHomeScreen {
				if m.managerInUse == "pip" {
					m.managerInUse = "uv"
				} else {
					m.managerInUse = "pip"
				}
			}

		case "ctrl+p":
			drawPythonRemotePackagesTable(&m, m.filteredPackages)
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
				m.remotePackageSelected = PackageInfo{}
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
			if m.showHomeScreen {
				m.focusOnLocalPackageTable = !m.focusOnLocalPackageTable
				if m.focusOnLocalPackageTable {
					m.pythonScriptTable.Blur()
					m.packageTable.Focus()
					m.packageTable.SetCursor(0)
					m.pythonScriptTable.SetCursor(-1)
				} else {
					m.pythonScriptTable.Focus()
					m.packageTable.Blur()
					m.pythonScriptTable.SetCursor(0)
					m.packageTable.SetCursor(-1)
				}
			}
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

		case "enter":
			if m.openPackageInstallScreen {
				if m.remotePackageTable.Focused() {
					m.remotePackageSelected = getPackageInfo(m.remotePackageTable.SelectedRow()[0])
				}
			}

		case "ctrl+a":
			if m.openPackageInstallScreen {
				if m.remotePackageTable.Focused() {
					m.info = fmt.Sprintf("%v Installing %v...", m.spinner.View(), m.remotePackageTable.SelectedRow()[0])
					return m, runInstallCommandAndRespondAsync(&m)
				}
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

	case InstallResponseObject:
		updateSpinnerType(&m)
		m.info = lipgloss.NewStyle().Foreground(lipgloss.Color(func() string {
			if msg.isErr {
				return "1"
			} else {
				return "2"
			}
		}())).Render(msg.content)

	case LoadedPythonManager:
		updateSpinnerType(&m)
		drawPythonPackageTable(&m, msg.pacman)
		drawPythonScriptsTable(&m, msg.pacman)
		m.localPackages = msg.pacman.packages
		m.err = msg.err
		if msg.err != nil {
			m.info = fmt.Sprintf("err: %v", msg.err.Error())
		}
		m.loadingState = false
		m.showPackageTable = true

	case InfoMsg:
		m.remotePackagesIndexedSuccessfully = true
		m.info = string(msg)
		if len(pythonPackages) < 50 {
			m.info = "Indexing process failed! restart the app to resolve"
		}

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)

		updateRemotePackageTable(&m, m.filteredPackages)

		if !m.remotePackagesIndexedSuccessfully {
			m.info = fmt.Sprintf("%v Indexing remote packages on PYPI...", m.spinner.View())
		}

		if m.openPackageInstallScreen {
			if m.packageInput.Focused() {
				var query = strings.TrimSpace(m.packageInput.Value())
				m.filteredPackages = nil
				if query != "" {
					query = strings.ToLower(query)
					var exactMatches []string
					var closestMatches []string
					var looseMatches []string
					for _, pkg := range pythonPackages {
						switch {
						case query == strings.ToLower(pkg):
							exactMatches = append(exactMatches, pkg)
						case strings.HasPrefix(strings.ToLower(pkg), query):
							closestMatches = append(closestMatches, pkg)
						case strings.Contains(strings.ToLower(pkg), query):
							looseMatches = append(looseMatches, pkg)
						}
					}
					m.filteredPackages = nil
					m.filteredPackages = append(m.filteredPackages, exactMatches...)
					m.filteredPackages = append(m.filteredPackages, closestMatches...)
					m.filteredPackages = append(m.filteredPackages, looseMatches...)
					m.remotePackageTable.SetCursor(0)
				}
			}
		} else {
			m.filteredPackages = nil
		}

		return m, cmd
	}
	var cmd tea.Cmd
	if m.focusOnLocalPackageTable {
		m.packageTable, cmd = m.packageTable.Update(msg)
	} else {
		m.pythonScriptTable, cmd = m.pythonScriptTable.Update(msg)
	}

	if m.openPackageInstallScreen {
		m.packageInput, cmd = m.packageInput.Update(msg)
	}

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
			Render("HELP\nUse Ctrl + h or the Esc key to close this screen\nCtrl + c to exit the application\nCtrl + p to find (and install) a package\nUse p to toggle package managers while in home screen")
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
	var borderColor = lipgloss.Color("63")

	var infoText = lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		Render(m.info)

	var helpText = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render("Press Ctrl + H for help")

	var header = lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().
			Width(m.window.width/2).
			Align(lipgloss.Left).
			Render(infoText),
		lipgloss.NewStyle().
			Width(m.window.width/2).
			Align(lipgloss.Right).
			Render(helpText),
	)

	header = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(m.window.width - 2).
		Render(header)

	mainHeight := int(float64(m.window.height) * 0.45)

	var tableView = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(m.window.width/2 - 2).
		Height(mainHeight).
		Render(m.packageTable.View())

	var infoFrame = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(m.window.width/2 - 2).
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

	scriptHeight := int(float64(m.window.height) * 0.25)
	if scriptHeight < 6 {
		scriptHeight = 6
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
		Render(fmt.Sprintf("Type to filter | Enter to install | Esc to cancel * %s", lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(m.info)))

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

func main() {
	if _, err := tea.NewProgram(initialize(), tea.WithAltScreen()).Run(); err != nil {
		panic(err)
	}
}

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
