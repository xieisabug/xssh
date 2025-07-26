package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"xssh/internal/config"
	"xssh/internal/forwarding"
	"xssh/internal/ssh"
)

// ViewMode represents the current UI mode
type ViewMode int

const (
	ModeList ViewMode = iota
	ModeAdd
	ModeEdit
	ModeDelete
	ModeAuthSelect
	ModeKeySelect
	ModePasswordInput
	ModeKeyPasswordInput
	ModeConnectTest
	ModeKeySetup
	ModeForwardingSelect
	ModeForwardingAdd
	ModeForwardingList
	ModeRemoteHostSelect
)

// AuthType represents authentication method
type AuthType int

const (
	AuthPassword AuthType = iota
	AuthKey
)

// FormField represents current form field being edited
type FormField int

const (
	FieldHost FormField = iota
	FieldUser
	FieldPort
	FieldAlias
	FieldPassword
	FieldLocalHost
	FieldLocalPort
	FieldRemoteHost
	FieldRemotePort
	FieldDescription
)

// FormData holds data for add/edit forms
type FormData struct {
	Host        string
	User        string
	Port        string
	Identity    string
	Alias       string
	Password    string
	KeyPassword string
	AuthType    AuthType
	
	// Port forwarding fields
	LocalHost    string
	LocalPort    string
	RemoteHost   string
	RemotePort   string
	Description  string
	UseExistingHost bool // Whether to use an existing SSH host as remote host
	SelectedRemoteHostIndex int // Index of selected remote host from hosts list
}

// Model represents the application state
type Model struct {
	sshConfig     *config.SSHConfig
	hosts         []config.SSHHost
	filteredHosts []config.SSHHost
	cursor        int
	searchMode    bool   // Whether we're in search input mode
	filterQuery   string
	showHelp      bool   // Whether to show detailed help
	height        int
	width         int
	message       string
	messageType   string // "success", "error", "info"
	selectedHost  *config.SSHHost // Host to connect to when exiting
	
	// Form state
	viewMode      ViewMode
	formData      FormData
	currentField  FormField
	editIndex     int // Index of host being edited
	keyFiles      []string // Available SSH key files
	keyCursor     int // Cursor for key selection
	setupProgress string // Progress message for setup
	isSetupDone   bool // Whether setup completed successfully
	
	// Port forwarding state
	forwardingManager *forwarding.ForwardingManager
	forwardingType    forwarding.ForwardingType
	selectedHostIndex int // Index of selected host for forwarding
}

// NewModel creates a new model
func NewModel() Model {
	sshConfig, err := config.LoadSSHConfig()
	if err != nil {
		sshConfig = &config.SSHConfig{Hosts: []config.SSHHost{}}
	}

	return Model{
		sshConfig:         sshConfig,
		hosts:             sshConfig.Hosts,
		filteredHosts:     sshConfig.Hosts,
		cursor:            0,
		searchMode:        false,
		filterQuery:       "",
		showHelp:          false,
		message:           "",
		messageType:       "",
		selectedHost:      nil,
		viewMode:          ModeList,
		formData:          FormData{Port: "22", AuthType: AuthPassword},
		currentField:      FieldHost,
		editIndex:         -1,
		keyFiles:          []string{},
		keyCursor:         0,
		setupProgress:     "",
		isSetupDone:       false,
		forwardingManager: forwarding.NewManager(),
		selectedHostIndex: -1,
	}
}

// Init implements the tea.Model interface
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements the tea.Model interface
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	case tea.KeyMsg:
		switch m.viewMode {
		case ModeList:
			if m.searchMode {
				return m.handleSearchMode(msg)
			}
			return m.handleListMode(msg)
		case ModeAdd, ModeEdit:
			return m.handleFormMode(msg)
		case ModeDelete:
			return m.handleDeleteMode(msg)
		case ModeAuthSelect:
			return m.handleAuthSelectMode(msg)
		case ModeKeySelect:
			return m.handleKeySelectMode(msg)
		case ModePasswordInput:
			return m.handlePasswordInputMode(msg)
		case ModeKeyPasswordInput:
			return m.handleKeyPasswordInputMode(msg)
		case ModeConnectTest:
			return m.handleConnectTestMode(msg)
		case ModeKeySetup:
			return m.handleKeySetupMode(msg)
		case ModeForwardingSelect:
			return m.handleForwardingSelectMode(msg)
		case ModeForwardingAdd:
			return m.handleForwardingAddMode(msg)
		case ModeForwardingList:
			return m.handleForwardingListMode(msg)
		case ModeRemoteHostSelect:
			return m.handleRemoteHostSelectMode(msg)
		}
		return m.handleListMode(msg)

	case string:
		// Handle connection test results
		if msg == "connection_success" {
			m.setupProgress = "Connection successful! SSH keys configured."
			m.isSetupDone = true
		} else if strings.HasPrefix(msg, "connection_error:") {
			errorMsg := strings.TrimPrefix(msg, "connection_error:")
			m.setupProgress = fmt.Sprintf("Error: %s", errorMsg)
			m.message = errorMsg
			m.messageType = "error"
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear message on any key press
	m.message = ""
	m.messageType = ""

	switch msg.String() {
	case "esc":
		// Exit search mode
		m.searchMode = false
		
	case "enter":
		// Exit search mode and keep current filter
		m.searchMode = false
		
	case "backspace":
		if len(m.filterQuery) > 0 {
			m.filterQuery = m.filterQuery[:len(m.filterQuery)-1]
			m.filterHosts()
		}
		
	case "ctrl+c":
		return m, tea.Quit
		
	default:
		// Handle regular character input for filtering
		if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
			m.filterQuery += msg.String()
			m.filterHosts()
		}
	}
	
	return m, nil
}

func (m Model) handleListMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear message on any key press
	m.message = ""
	m.messageType = ""

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	
	case "down", "j":
		if m.cursor < len(m.filteredHosts)-1 {
			m.cursor++
		}
	
	case ":":
		// Enter search mode
		m.searchMode = true
	
	case "a":
		// Add new host
		m.viewMode = ModeAdd
		m.formData = FormData{Port: "22", AuthType: AuthPassword}
		m.currentField = FieldHost
	
	case "e":
		// Edit selected host
		if len(m.filteredHosts) > 0 {
			host := m.filteredHosts[m.cursor]
			m.viewMode = ModeEdit
			m.editIndex = m.findHostIndex(host.Name)
			m.formData = FormData{
				Host:     host.Host,
				User:     host.User,
				Port:     host.Port,
				Identity: host.Identity,
				Alias:    host.Name,
				AuthType: AuthPassword,
			}
			if host.Identity != "" {
				m.formData.AuthType = AuthKey
			}
			m.currentField = FieldHost
		}
	
	case "d":
		// Delete selected host
		if len(m.filteredHosts) > 0 {
			m.viewMode = ModeDelete
		}
	
	case "f":
		// Port forwarding for selected host
		if len(m.filteredHosts) > 0 {
			m.selectedHostIndex = m.cursor
			m.viewMode = ModeForwardingSelect
		}
	
	case "enter":
		if len(m.filteredHosts) > 0 {
			host := m.filteredHosts[m.cursor]
			// Store the selected host and quit
			m.selectedHost = &host
			return m, tea.Quit
		}
	
	case "c":
		if len(m.filteredHosts) > 0 {
			host := m.filteredHosts[m.cursor]
			if err := ssh.CopySSHCommand(host); err != nil {
				m.message = "Failed to copy command"
				m.messageType = "error"
			} else {
				m.message = "SSH command copied to clipboard!"
				m.messageType = "success"
			}
		}
	
	case "esc":
		// Clear filter
		m.filterQuery = ""
		m.filteredHosts = m.hosts
		m.cursor = 0
		// Also close help if open
		m.showHelp = false
	
	case "?", "h", "m":
		// Toggle help display
		m.showHelp = !m.showHelp
	}
	
	return m, nil
}

// renderBasicHelp renders the condensed help text
func (m Model) renderBasicHelp() string {
	if m.searchMode {
		return "Type to search • ESC: exit search • Enter: confirm • Ctrl+C: quit"
	}
	return "↑/j↓: nav • Enter: connect • a: add • e: edit • d: del • f: forward • :: search • ?: help • q: quit"
}

// renderDetailedHelp renders the full help overlay
func (m Model) renderDetailedHelp() string {
	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width).
		Align(lipgloss.Center)
	
	content.WriteString(headerStyle.Render("KEYBOARD SHORTCUTS") + "\n\n")
	
	// Create sections
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginTop(1)
	
	itemStyle := lipgloss.NewStyle().
		MarginLeft(2)
	
	// Navigation section
	content.WriteString(sectionStyle.Render("NAVIGATION") + "\n")
	content.WriteString(itemStyle.Render("↑/k, ↓/j         Navigate up/down") + "\n")
	content.WriteString(itemStyle.Render("Enter            Connect to selected host") + "\n")
	content.WriteString(itemStyle.Render("ESC              Clear filter or close help") + "\n\n")
	
	// Host Management section  
	content.WriteString(sectionStyle.Render("HOST MANAGEMENT") + "\n")
	content.WriteString(itemStyle.Render("a                Add new host") + "\n")
	content.WriteString(itemStyle.Render("e                Edit selected host") + "\n")  
	content.WriteString(itemStyle.Render("d                Delete selected host") + "\n")
	content.WriteString(itemStyle.Render("c                Copy SSH command to clipboard") + "\n\n")
	
	// Advanced Features section
	content.WriteString(sectionStyle.Render("ADVANCED FEATURES") + "\n")
	content.WriteString(itemStyle.Render("f                Port forwarding menu") + "\n")
	content.WriteString(itemStyle.Render(":                Search/filter hosts") + "\n\n")
	
	// General section
	content.WriteString(sectionStyle.Render("GENERAL") + "\n")
	content.WriteString(itemStyle.Render("?, h, m          Toggle this help") + "\n")
	content.WriteString(itemStyle.Render("q, Ctrl+C        Quit application") + "\n\n")
	
	// Footer
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(m.width).
		Align(lipgloss.Center).
		MarginTop(1)
	
	content.WriteString(footerStyle.Render("Press ESC or ? to close help"))
	
	return content.String()
}

func (m *Model) filterHosts() {
	if m.filterQuery == "" {
		m.filteredHosts = m.hosts
		m.cursor = 0
		return
	}

	m.filteredHosts = []config.SSHHost{}
	query := strings.ToLower(m.filterQuery)
	
	for _, host := range m.hosts {
		if strings.Contains(strings.ToLower(host.Name), query) ||
			strings.Contains(strings.ToLower(host.Host), query) ||
			strings.Contains(strings.ToLower(host.User), query) {
			m.filteredHosts = append(m.filteredHosts, host)
		}
	}
	
	// Reset cursor to top
	m.cursor = 0
}

// findHostIndex finds the index of a host by name in the main hosts slice
func (m Model) findHostIndex(name string) int {
	for i, host := range m.hosts {
		if host.Name == name {
			return i
		}
	}
	return -1
}

// handleFormMode handles input for Add/Edit forms
func (m Model) handleFormMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel form
		m.viewMode = ModeList
		return m, nil
	
	case "tab", "down":
		// Next field
		switch m.currentField {
		case FieldHost:
			m.currentField = FieldUser
		case FieldUser:
			m.currentField = FieldPort
		case FieldPort:
			// Go to auth selection
			m.viewMode = ModeAuthSelect
		case FieldAlias:
			// Go to password input or connection test
			if m.formData.AuthType == AuthPassword {
				m.currentField = FieldPassword
				m.viewMode = ModePasswordInput
			} else {
				// For key auth, go to connection test
				return m.startConnectionTest()
			}
		}
	
	case "shift+tab", "up":
		// Previous field
		switch m.currentField {
		case FieldUser:
			m.currentField = FieldHost
		case FieldPort:
			m.currentField = FieldUser
		case FieldAlias:
			m.currentField = FieldPort
		}
	
	case "enter":
		// Next field or save
		if m.currentField == FieldAlias {
			// Go to password input or connection test
			if m.formData.AuthType == AuthPassword {
				m.currentField = FieldPassword
				m.viewMode = ModePasswordInput
				return m, nil
			} else {
				// For key auth, go to connection test
				return m.startConnectionTest()
			}
		}
		// Trigger tab behavior
		return m.handleFormMode(tea.KeyMsg{Type: tea.KeyTab})
	
	case "backspace":
		// Delete character from current field
		switch m.currentField {
		case FieldHost:
			if len(m.formData.Host) > 0 {
				m.formData.Host = m.formData.Host[:len(m.formData.Host)-1]
			}
		case FieldUser:
			if len(m.formData.User) > 0 {
				m.formData.User = m.formData.User[:len(m.formData.User)-1]
			}
		case FieldPort:
			if len(m.formData.Port) > 0 {
				m.formData.Port = m.formData.Port[:len(m.formData.Port)-1]
			}
		case FieldAlias:
			if len(m.formData.Alias) > 0 {
				m.formData.Alias = m.formData.Alias[:len(m.formData.Alias)-1]
			}
		}
	
	default:
		// Add character to current field
		if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
			switch m.currentField {
			case FieldHost:
				m.formData.Host += msg.String()
			case FieldUser:
				m.formData.User += msg.String()
			case FieldPort:
				m.formData.Port += msg.String()
			case FieldAlias:
				m.formData.Alias += msg.String()
			}
		}
	}
	
	return m, nil
}

// handleDeleteMode handles delete confirmation
func (m Model) handleDeleteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm delete
		if len(m.filteredHosts) > 0 {
			hostToDelete := m.filteredHosts[m.cursor]
			m.sshConfig.RemoveHost(hostToDelete.Name)
			if err := m.sshConfig.Save(); err != nil {
				m.message = fmt.Sprintf("Failed to save config: %v", err)
				m.messageType = "error"
			} else {
				m.message = fmt.Sprintf("Host '%s' deleted", hostToDelete.Name)
				m.messageType = "success"
				// Reload hosts
				m.hosts = m.sshConfig.Hosts
				m.filteredHosts = m.hosts
				if m.cursor >= len(m.filteredHosts) {
					m.cursor = len(m.filteredHosts) - 1
				}
				if m.cursor < 0 {
					m.cursor = 0
				}
			}
		}
		m.viewMode = ModeList
	
	case "n", "N", "esc":
		// Cancel delete
		m.viewMode = ModeList
	}
	
	return m, nil
}

// handleAuthSelectMode handles authentication type selection
func (m Model) handleAuthSelectMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ModeAdd
		if m.editIndex >= 0 {
			m.viewMode = ModeEdit
		}
		m.currentField = FieldPort
	
	case "1":
		m.formData.AuthType = AuthPassword
		m.formData.Identity = ""
		m.currentField = FieldAlias
		m.viewMode = ModeAdd
		if m.editIndex >= 0 {
			m.viewMode = ModeEdit
		}
	
	case "2":
		m.formData.AuthType = AuthKey
		// Load available SSH keys
		m.loadSSHKeys()
		if len(m.keyFiles) > 0 {
			m.viewMode = ModeKeySelect
			m.keyCursor = 0
		} else {
			m.message = "No SSH keys found in ~/.ssh/"
			m.messageType = "error"
			m.viewMode = ModeAdd
			if m.editIndex >= 0 {
				m.viewMode = ModeEdit
			}
		}
	}
	
	return m, nil
}

// handleKeySelectMode handles SSH key selection
func (m Model) handleKeySelectMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ModeAuthSelect
	
	case "up", "k":
		if m.keyCursor > 0 {
			m.keyCursor--
		}
	
	case "down", "j":
		if m.keyCursor < len(m.keyFiles)-1 {
			m.keyCursor++
		}
	
	case "enter":
		if len(m.keyFiles) > 0 {
			m.formData.Identity = m.keyFiles[m.keyCursor]
			// Check if key needs a password by trying to parse it
			if m.checkKeyNeedsPassword(m.formData.Identity) {
				m.viewMode = ModeKeyPasswordInput
			} else {
				m.currentField = FieldAlias
				m.viewMode = ModeAdd
				if m.editIndex >= 0 {
					m.viewMode = ModeEdit
				}
			}
		}
	}
	
	return m, nil
}

// View implements the tea.Model interface
func (m Model) View() string {
	if m.height == 0 {
		return "Loading..."
	}

	// Handle different view modes
	switch m.viewMode {
	case ModeAdd, ModeEdit:
		return m.renderFormView()
	case ModeDelete:
		return m.renderDeleteView()
	case ModeAuthSelect:
		return m.renderAuthSelectView()
	case ModeKeySelect:
		return m.renderKeySelectView()
	case ModePasswordInput:
		return m.renderPasswordInputView()
	case ModeKeyPasswordInput:
		return m.renderKeyPasswordInputView()
	case ModeConnectTest, ModeKeySetup:
		return m.renderConnectTestView()
	case ModeForwardingSelect:
		return m.renderForwardingSelectView()
	case ModeForwardingAdd:
		return m.renderForwardingAddView()
	case ModeForwardingList:
		return m.renderForwardingListView()
	case ModeRemoteHostSelect:
		return m.renderRemoteHostSelectView()
	default:
		return m.renderListView()
	}
}

// renderListView renders the main host list view
func (m Model) renderListView() string {
	// Define styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Height(m.height - 8). // Leave space for header, filter, and help
		Width(m.width - 4)

	filterStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true)

	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#999999")).
		Italic(true).
		Align(lipgloss.Center)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(m.width)

	messageStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center)

	// Build the view
	var content strings.Builder

	// Header
	header := headerStyle.Render("SSH Connection Manager")
	content.WriteString(header + "\n\n")

	// Filter display
	var filterDisplay string
	if m.searchMode {
		filterDisplay = fmt.Sprintf("Search: %s", m.filterQuery)
		if m.filterQuery != "" {
			filterDisplay += "█"
		} else {
			filterDisplay += "█"
		}
	} else {
		if m.filterQuery != "" {
			filterDisplay = fmt.Sprintf("Filtered by: %s", m.filterQuery)
		} else {
			filterDisplay = "Press ':' to search"
		}
	}
	content.WriteString(filterStyle.Render(filterDisplay) + "\n\n")

	// Host list panel
	var listContent strings.Builder
	
	if len(m.filteredHosts) == 0 {
		if m.filterQuery == "" {
			listContent.WriteString(emptyStyle.Render("No SSH hosts configured"))
		} else {
			listContent.WriteString(emptyStyle.Render("No hosts match your filter"))
		}
	} else {
		// Add table header
		listContent.WriteString(m.formatTableHeader() + "\n")
		
		// Add host rows
		for i, host := range m.filteredHosts {
			cursor := "  "
			if m.cursor == i {
				cursor = "▶ "
			}

			hostDisplay := fmt.Sprintf("%s%s", cursor, m.formatTableRow(host))
			
			if m.cursor == i {
				listContent.WriteString(selectedStyle.Render(hostDisplay) + "\n")
			} else {
				listContent.WriteString(hostDisplay + "\n")
			}
		}
	}

	panel := panelStyle.Render(listContent.String())
	content.WriteString(panel + "\n")

	// Message
	if m.message != "" {
		var msgStyle lipgloss.Style
		switch m.messageType {
		case "success":
			msgStyle = messageStyle.Foreground(lipgloss.Color("#00FF00"))
		case "error":
			msgStyle = messageStyle.Foreground(lipgloss.Color("#FF0000"))
		default:
			msgStyle = messageStyle.Foreground(lipgloss.Color("#FFFF00"))
		}
		content.WriteString(msgStyle.Render(m.message) + "\n")
	}

	// Help
	content.WriteString(helpStyle.Render(m.renderBasicHelp()))
	
	// Show detailed help overlay if requested
	if m.showHelp {
		overlayStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(2).
			Width(m.width - 8).
			MaxHeight(m.height - 4)
		
		overlay := overlayStyle.Render(m.renderDetailedHelp())
		
		// Position overlay in center of screen
		overlayLines := strings.Split(overlay, "\n")
		startY := (m.height - len(overlayLines)) / 2
		if startY < 0 {
			startY = 0
		}
		
		// Add padding to bring overlay to center
		centeredOverlay := strings.Repeat("\n", startY) + overlay
		content.WriteString("\n" + centeredOverlay)
	}

	return content.String()
}

// calculateColumnWidths calculates optimal column widths for the host table
func (m Model) calculateColumnWidths() (int, int, int, int, int) {
	if len(m.filteredHosts) == 0 {
		// Default widths when no hosts
		return 15, 18, 12, 6, 8
	}
	
	// Find maximum widths needed for each column
	maxName, maxHost, maxUser, maxPort := 4, 4, 4, 4 // Minimum header widths
	
	for _, host := range m.filteredHosts {
		if len(host.Name) > maxName {
			maxName = len(host.Name)
		}
		
		if len(host.Host) > maxHost {
			maxHost = len(host.Host)
		}
		
		if len(host.User) > maxUser {
			maxUser = len(host.User)
		}
		
		if len(host.Port) > maxPort {
			maxPort = len(host.Port)
		}
	}
	
	// Calculate available width (subtract cursor space, borders, padding)
	availableWidth := m.width - 8 // Account for borders and padding
	
	// Reserve space for cursor and separators
	cursorWidth := 2
	sepWidth := 3 * 3 // 3 separators, each 3 chars wide (" │ ")
	authWidth := 8    // Fixed width for auth type column
	
	usableWidth := availableWidth - cursorWidth - sepWidth - authWidth
	
	// Distribute remaining width among columns with priority: Name > Host > User > Port
	nameWidth := maxName
	hostWidth := maxHost
	userWidth := maxUser
	portWidth := maxPort
	
	totalNeeded := nameWidth + hostWidth + userWidth + portWidth
	
	if totalNeeded > usableWidth {
		// Need to truncate columns, prioritize Name and Host
		if usableWidth >= 40 {
			nameWidth = min(maxName, usableWidth/4)
			hostWidth = min(maxHost, usableWidth/3)
			userWidth = min(maxUser, (usableWidth-nameWidth-hostWidth)/2)
			portWidth = usableWidth - nameWidth - hostWidth - userWidth
		} else {
			// Very narrow terminal
			nameWidth = min(maxName, usableWidth/2)
			hostWidth = usableWidth - nameWidth
			userWidth = 0
			portWidth = 0
		}
	} else if totalNeeded < usableWidth {
		// Extra space available, expand columns proportionally
		extra := usableWidth - totalNeeded
		nameWidth += extra / 3
		hostWidth += extra / 3
		userWidth += extra - (extra/3)*2
	}
	
	return max(nameWidth, 4), max(hostWidth, 4), max(userWidth, 4), max(portWidth, 4), authWidth
}

// formatTableHeader creates a formatted table header
func (m Model) formatTableHeader() string {
	nameWidth, hostWidth, userWidth, portWidth, authWidth := m.calculateColumnWidths()
	
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4"))
	
	name := padAndTruncate("NAME", nameWidth)
	host := padAndTruncate("HOST", hostWidth)  
	user := padAndTruncate("USER", userWidth)
	port := padAndTruncate("PORT", portWidth)
	auth := padAndTruncate("AUTH", authWidth)
	
	var header string
	if userWidth > 0 && portWidth > 0 {
		header = fmt.Sprintf("  %s │ %s │ %s │ %s │ %s", name, host, user, port, auth)
	} else if userWidth > 0 {
		header = fmt.Sprintf("  %s │ %s │ %s │ %s", name, host, user, auth)
	} else {
		header = fmt.Sprintf("  %s │ %s │ %s", name, host, auth)
	}
	
	return headerStyle.Render(header)
}

// formatTableRow formats a single host as a table row
func (m Model) formatTableRow(host config.SSHHost) string {
	nameWidth, hostWidth, userWidth, portWidth, authWidth := m.calculateColumnWidths()
	
	name := padAndTruncate(host.Name, nameWidth)
	hostAddr := padAndTruncate(host.Host, hostWidth)
	user := padAndTruncate(host.User, userWidth)
	port := padAndTruncate(host.Port, portWidth)
	
	// Determine auth type
	authType := "PWD"
	if host.Identity != "" {
		authType = "KEY"
	}
	auth := padAndTruncate(authType, authWidth)
	
	if userWidth > 0 && portWidth > 0 {
		return fmt.Sprintf("%s │ %s │ %s │ %s │ %s", name, hostAddr, user, port, auth)
	} else if userWidth > 0 {
		return fmt.Sprintf("%s │ %s │ %s │ %s", name, hostAddr, user, auth)
	} else {
		return fmt.Sprintf("%s │ %s │ %s", name, hostAddr, auth)
	}
}

// padAndTruncate pads or truncates a string to the specified width
func padAndTruncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	
	if len(s) > width {
		if width <= 3 {
			return s[:width]
		}
		return s[:width-3] + "..."
	}
	
	return fmt.Sprintf("%-*s", width, s)
}

// Helper function for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetSelectedHost returns the host selected for connection, if any
func (m Model) GetSelectedHost() *config.SSHHost {
	return m.selectedHost
}

// loadSSHKeys loads available SSH private key files from ~/.ssh/
func (m *Model) loadSSHKeys() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	
	sshDir := filepath.Join(homeDir, ".ssh")
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return
	}
	
	m.keyFiles = []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		// Skip public keys and known_hosts, config files
		if strings.HasSuffix(name, ".pub") || 
		   name == "known_hosts" || 
		   name == "config" ||
		   name == "authorized_keys" {
			continue
		}
		
		fullPath := filepath.Join(sshDir, name)
		m.keyFiles = append(m.keyFiles, fullPath)
	}
}

// saveHost saves the current form data as a new or updated host
func (m Model) saveHost() (tea.Model, tea.Cmd) {
	// Validate required fields
	if m.formData.Host == "" {
		m.message = "Host address is required"
		m.messageType = "error"
		return m, nil
	}
	
	if m.formData.Alias == "" {
		m.message = "Alias is required"
		m.messageType = "error"
		return m, nil
	}
	
	// Default port if empty
	port := m.formData.Port
	if port == "" {
		port = "22"
	}
	
	// Create new host config
	newHost := config.SSHHost{
		Name:     m.formData.Alias,
		Host:     m.formData.Host,
		User:     m.formData.User,
		Port:     port,
		Identity: m.formData.Identity,
	}
	
	if m.viewMode == ModeEdit && m.editIndex >= 0 {
		// Update existing host
		oldName := m.hosts[m.editIndex].Name
		m.sshConfig.RemoveHost(oldName)
		m.sshConfig.AddHost(newHost)
		m.message = fmt.Sprintf("Host '%s' updated", newHost.Name)
	} else {
		// Add new host
		// Check if alias already exists
		for _, host := range m.hosts {
			if host.Name == newHost.Name {
				m.message = fmt.Sprintf("Host alias '%s' already exists", newHost.Name)
				m.messageType = "error"
				return m, nil
			}
		}
		m.sshConfig.AddHost(newHost)
		m.message = fmt.Sprintf("Host '%s' added", newHost.Name)
	}
	
	// Save to file
	if err := m.sshConfig.Save(); err != nil {
		m.message = fmt.Sprintf("Failed to save config: %v", err)
		m.messageType = "error"
		return m, nil
	}
	
	m.messageType = "success"
	
	// Reload hosts and return to list
	m.hosts = m.sshConfig.Hosts
	m.filteredHosts = m.hosts
	m.viewMode = ModeList
	m.editIndex = -1
	
	return m, nil
}

// handlePasswordInputMode handles password input
func (m Model) handlePasswordInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ModeAdd
		if m.editIndex >= 0 {
			m.viewMode = ModeEdit
		}
		m.currentField = FieldAlias
	
	case "enter":
		// Start connection test
		return m.startConnectionTest()
	
	case "backspace":
		if len(m.formData.Password) > 0 {
			m.formData.Password = m.formData.Password[:len(m.formData.Password)-1]
		}
	
	default:
		// Add character to password field
		if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
			m.formData.Password += msg.String()
		}
	}
	
	return m, nil
}

// handleConnectTestMode handles the connection testing phase
func (m Model) handleConnectTestMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.isSetupDone {
			// Setup completed, return to list
			return m.saveHostAndReturn()
		} else {
			// Cancel setup, return to form
			m.viewMode = ModePasswordInput
			if m.formData.AuthType == AuthKey {
				m.viewMode = ModeAdd
				if m.editIndex >= 0 {
					m.viewMode = ModeEdit
				}
				m.currentField = FieldAlias
			}
		}
	
	case "enter":
		if m.isSetupDone {
			// Setup completed, save and return to list
			return m.saveHostAndReturn()
		}
	}
	
	return m, nil
}

// handleKeySetupMode handles SSH key setup phase
func (m Model) handleKeySetupMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ModeConnectTest
	case "enter":
		if m.isSetupDone {
			return m.saveHostAndReturn()
		}
	}
	
	return m, nil
}

// handleKeyPasswordInputMode handles SSH private key password input
func (m Model) handleKeyPasswordInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ModeKeySelect
	
	case "enter":
		// Continue to alias field
		m.currentField = FieldAlias
		m.viewMode = ModeAdd
		if m.editIndex >= 0 {
			m.viewMode = ModeEdit
		}
	
	case "backspace":
		if len(m.formData.KeyPassword) > 0 {
			m.formData.KeyPassword = m.formData.KeyPassword[:len(m.formData.KeyPassword)-1]
		}
	
	default:
		// Add character to key password field
		if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
			m.formData.KeyPassword += msg.String()
		}
	}
	
	return m, nil
}

// checkKeyNeedsPassword checks if an SSH private key is encrypted
func (m Model) checkKeyNeedsPassword(keyPath string) bool {
	// Read the key file
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return false
	}
	
	// Check if the key contains encryption headers
	keyContent := string(keyData)
	// Look for encrypted key markers
	return strings.Contains(keyContent, "Proc-Type: 4,ENCRYPTED") || 
		   strings.Contains(keyContent, "-----BEGIN ENCRYPTED PRIVATE KEY-----")
}

// startConnectionTest begins the connection test process
func (m Model) startConnectionTest() (tea.Model, tea.Cmd) {
	m.viewMode = ModeConnectTest
	m.setupProgress = "Testing connection..."
	m.isSetupDone = false
	
	// Create a command to test the connection
	return m, tea.Cmd(func() tea.Msg {
		return m.testConnection()
	})
}

// testConnection tests SSH connection and sets up keys if needed
func (m Model) testConnection() tea.Msg {
	// Create host config for testing
	host := config.SSHHost{
		Name:     m.formData.Alias,
		Host:     m.formData.Host,
		User:     m.formData.User,
		Port:     m.formData.Port,
		Identity: m.formData.Identity,
	}
	
	var result ssh.SetupResult
	
	// Test connection based on auth type
	if m.formData.AuthType == AuthKey && m.formData.Identity != "" {
		// Test key-based connection with or without password
		result = ssh.TestConnectionWithKeyPassword(host, m.formData.KeyPassword)
	} else {
		// Test password connection and set up keys
		result = ssh.TestConnection(host, m.formData.Password)
	}
	
	if result.Success {
		// Update form data with generated key path if applicable
		if m.formData.AuthType == AuthPassword && host.Identity == "" {
			// SSH key was generated, update identity path
			homeDir, _ := os.UserHomeDir()
			m.formData.Identity = filepath.Join(homeDir, ".ssh", "id_rsa")
			m.formData.AuthType = AuthKey
		}
		return "connection_success"
	} else {
		return fmt.Sprintf("connection_error:%s", result.Message)
	}
}

// saveHostAndReturn saves the host and returns to list
func (m Model) saveHostAndReturn() (tea.Model, tea.Cmd) {
	return m.saveHost()
}

// handleForwardingSelectMode handles forwarding type selection
func (m Model) handleForwardingSelectMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ModeList
	
	case "1":
		m.forwardingType = forwarding.LocalForward
		m.formData = FormData{
			LocalHost:  "localhost",
			LocalPort:  "",
			RemoteHost: "",
			RemotePort: "",
		}
		m.currentField = FieldLocalPort
		m.viewMode = ModeForwardingAdd
	
	case "2":
		m.forwardingType = forwarding.RemoteForward
		m.formData = FormData{
			LocalHost:  "localhost",
			LocalPort:  "",
			RemoteHost: "localhost",
			RemotePort: "",
		}
		m.currentField = FieldRemotePort
		m.viewMode = ModeForwardingAdd
	
	case "3":
		m.forwardingType = forwarding.DynamicForward
		m.formData = FormData{
			LocalHost: "localhost",
			LocalPort: "",
		}
		m.currentField = FieldLocalPort
		m.viewMode = ModeForwardingAdd
	
	case "l":
		// Show active forwarding list
		m.viewMode = ModeForwardingList
	}
	
	return m, nil
}

// handleForwardingAddMode handles forwarding add form
func (m Model) handleForwardingAddMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ModeForwardingSelect
	
	case "enter":
		// Handle special case for remote host selection
		if m.currentField == FieldRemoteHost && m.forwardingType == forwarding.LocalForward {
			// Go to remote host selection mode
			m.cursor = 0 // Reset cursor for host selection
			m.viewMode = ModeRemoteHostSelect
			return m, nil
		}
		// Start the forwarding
		return m.startForwarding()
	
	case "tab", "down":
		// Next field based on forwarding type
		switch m.forwardingType {
		case forwarding.LocalForward:
			switch m.currentField {
			case FieldLocalPort:
				m.currentField = FieldRemoteHost
			case FieldRemoteHost:
				m.currentField = FieldRemotePort
			case FieldRemotePort:
				m.currentField = FieldDescription
			}
		case forwarding.RemoteForward:
			switch m.currentField {
			case FieldRemotePort:
				m.currentField = FieldLocalPort
			case FieldLocalPort:
				m.currentField = FieldDescription
			}
		case forwarding.DynamicForward:
			switch m.currentField {
			case FieldLocalPort:
				m.currentField = FieldDescription
			}
		}
	
	case "backspace":
		// Delete character from current field
		switch m.currentField {
		case FieldLocalPort:
			if len(m.formData.LocalPort) > 0 {
				m.formData.LocalPort = m.formData.LocalPort[:len(m.formData.LocalPort)-1]
			}
		case FieldRemoteHost:
			if len(m.formData.RemoteHost) > 0 {
				m.formData.RemoteHost = m.formData.RemoteHost[:len(m.formData.RemoteHost)-1]
			}
		case FieldRemotePort:
			if len(m.formData.RemotePort) > 0 {
				m.formData.RemotePort = m.formData.RemotePort[:len(m.formData.RemotePort)-1]
			}
		case FieldDescription:
			if len(m.formData.Description) > 0 {
				m.formData.Description = m.formData.Description[:len(m.formData.Description)-1]
			}
		}
	
	default:
		// Add character to current field
		if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
			switch m.currentField {
			case FieldLocalPort:
				m.formData.LocalPort += msg.String()
			case FieldRemoteHost:
				m.formData.RemoteHost += msg.String()
			case FieldRemotePort:
				m.formData.RemotePort += msg.String()
			case FieldDescription:
				m.formData.Description += msg.String()
			}
		}
	}
	
	return m, nil
}

// handleForwardingListMode handles the forwarding list view
func (m Model) handleForwardingListMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.viewMode = ModeList
	
	case "s":
		// Stop selected forwarding
		sessions := m.forwardingManager.GetAllSessions()
		if m.cursor < len(sessions) {
			session := sessions[m.cursor]
			if err := m.forwardingManager.StopForwarding(session.Rule.ID); err != nil {
				m.message = fmt.Sprintf("Failed to stop forwarding: %v", err)
				m.messageType = "error"
			} else {
				m.message = "Forwarding stopped"
				m.messageType = "success"
			}
		}
	
	case "a":
		// Add new forwarding
		m.viewMode = ModeForwardingSelect
	
	case "up", "k":
		sessions := m.forwardingManager.GetAllSessions()
		if m.cursor > 0 && len(sessions) > 0 {
			m.cursor--
		}
	
	case "down", "j":
		sessions := m.forwardingManager.GetAllSessions()
		if m.cursor < len(sessions)-1 {
			m.cursor++
		}
	}
	
	return m, nil
}

// startForwarding starts a new port forwarding session
func (m Model) startForwarding() (tea.Model, tea.Cmd) {
	// Validate inputs
	if m.formData.LocalPort == "" {
		m.message = "Local port is required"
		m.messageType = "error"
		return m, nil
	}
	
	if m.forwardingType != forwarding.DynamicForward {
		if m.formData.RemoteHost == "" {
			m.message = "Remote host is required"
			m.messageType = "error"
			return m, nil
		}
		if m.formData.RemotePort == "" {
			m.message = "Remote port is required"
			m.messageType = "error"
			return m, nil
		}
	}
	
	// Parse ports
	localPort := 0
	remotePort := 0
	if _, err := fmt.Sscanf(m.formData.LocalPort, "%d", &localPort); err != nil {
		m.message = "Invalid local port"
		m.messageType = "error"
		return m, nil
	}
	
	if m.forwardingType != forwarding.DynamicForward {
		if _, err := fmt.Sscanf(m.formData.RemotePort, "%d", &remotePort); err != nil {
			m.message = "Invalid remote port"
			m.messageType = "error"
			return m, nil
		}
	}
	
	// Determine the actual remote host address
	actualRemoteHost := m.formData.RemoteHost
	if m.formData.UseExistingHost && m.formData.SelectedRemoteHostIndex < len(m.hosts) {
		// Use the actual host address from the selected SSH host
		selectedHost := m.hosts[m.formData.SelectedRemoteHostIndex]
		actualRemoteHost = selectedHost.Host
	}
	
	// Create forwarding rule
	rule := forwarding.ForwardingRule{
		ID:          fmt.Sprintf("%s-%d-%d", m.forwardingType.String(), localPort, time.Now().Unix()),
		Type:        m.forwardingType,
		LocalHost:   m.formData.LocalHost,
		LocalPort:   localPort,
		RemoteHost:  actualRemoteHost,
		RemotePort:  remotePort,
		Description: m.formData.Description,
	}
	
	// Get selected host
	if m.selectedHostIndex < 0 || m.selectedHostIndex >= len(m.filteredHosts) {
		m.message = "No host selected"
		m.messageType = "error"
		return m, nil
	}
	
	host := m.filteredHosts[m.selectedHostIndex]
	
	// Start forwarding
	if err := m.forwardingManager.StartForwarding(rule, host, m.formData.KeyPassword); err != nil {
		m.message = fmt.Sprintf("Failed to start forwarding: %v", err)
		m.messageType = "error"
		return m, nil
	}
	
	m.message = fmt.Sprintf("Port forwarding started: %s", rule.Description)
	m.messageType = "success"
	m.viewMode = ModeForwardingList
	
	return m, nil
}

// Note: Forwarding view functions (renderForwardingSelectView, renderForwardingAddView, renderForwardingListView)
// are defined in forwarding_views.go for better code organization

// handleRemoteHostSelectMode handles remote host selection
func (m Model) handleRemoteHostSelectMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ModeForwardingAdd
	
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	
	case "down", "j":
		// +1 for manual input option
		if m.cursor < len(m.hosts) {
			m.cursor++
		}
	
	case "enter":
		if m.cursor == len(m.hosts) {
			// Manual input option selected
			m.formData.UseExistingHost = false
			m.formData.RemoteHost = ""
			m.currentField = FieldRemoteHost
		} else {
			// Existing host selected
			if m.cursor < len(m.hosts) {
				selectedHost := m.hosts[m.cursor]
				m.formData.UseExistingHost = true
				m.formData.SelectedRemoteHostIndex = m.cursor
				m.formData.RemoteHost = selectedHost.Host
				m.currentField = FieldRemotePort
			}
		}
		m.viewMode = ModeForwardingAdd
	}
	
	return m, nil
}