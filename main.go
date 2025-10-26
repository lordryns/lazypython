package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

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

type LogObject struct {
	Level   string
	Time    string
	Message string
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
	logs                              []LogObject
	showLoggingScreen                 bool
	logTable                          table.Model
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
		if ok := toCacheOrNotToCache(); ok {
			return InfoMsg("Packages loaded from cache!")
		}
		fetchPackagesFromIndex()
		remotePackagesIndexedSuccessfully = true

		savePackagesToCache(pythonPackages)
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
	m.info = fmt.Sprintf("%v Installing %v...", m.spinner.View(), m.remotePackageTable.SelectedRow()[0])
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

			if m.showLoggingScreen {
				m.showLoggingScreen = false
			}

		case "ctrl+l":
			m.showLoggingScreen = !m.showLoggingScreen
			m.showHomeScreen = false
			m.openPackageInstallScreen = false
			m.openHelpMenu = false
			if !m.showLoggingScreen {
				m.showHomeScreen = true
			}
			updateLoggingTable(&m)

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
		if msg.isErr {
			m.err = errors.New(msg.content)
			var tnow = time.Now()
			m.logs = append(m.logs, LogObject{Level: "Error", Time: fmt.Sprintf("%v-%v-%v", tnow.Hour(), tnow.Minute(), tnow.Second()),
				Message: msg.content})
			m.info = "Failed to install package! Ctrl + L for logs"
		} else {
			m.info = "Package installed successfully!"
		}
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
					var pkgCount int
					var nonExactCount int
					for _, pkg := range pythonPackages {
						var pkgLower = strings.ToLower(pkg)
						switch {
						case query == pkgLower:
							pkgCount += 1
							exactMatches = append(exactMatches, pkg)
						case strings.HasPrefix(pkgLower, query) && nonExactCount < 29:
							pkgCount += 1
							nonExactCount += 1
							closestMatches = append(closestMatches, pkg)
						case strings.Contains(pkgLower, query) && nonExactCount < 29:
							pkgCount += 1
							nonExactCount += 1
							looseMatches = append(looseMatches, pkg)
						}

						if pkgCount > 30 {
							break
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
	m.logTable, cmd = m.logTable.Update(msg)
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

	if m.showLoggingScreen {
		return drawLoggingPage(&m)
	}

	return lipgloss.NewStyle().Width(m.window.width).Height(m.window.height).Align(lipgloss.Center, lipgloss.Center).Render("Somehow this page showed up even though it isn't supposed to, press the Esc key to return to Home... restart if this persists.")
}

func main() {
	if _, err := tea.NewProgram(initialize(), tea.WithAltScreen()).Run(); err != nil {
		panic(err)
	}
}
