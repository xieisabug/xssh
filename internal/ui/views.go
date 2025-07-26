package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderFormView renders the Add/Edit form
func (m Model) renderFormView() string {
	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)
	
	title := "Add New Host"
	if m.viewMode == ModeEdit {
		title = "Edit Host"
	}
	header := headerStyle.Render(title)
	content.WriteString(header + "\n\n")
	
	// Form fields
	fieldStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(40)
	
	activeFieldStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(0, 1).
		Width(40).
		Bold(true)
	
	// Host field
	hostValue := m.formData.Host
	if m.currentField == FieldHost {
		hostValue += "█"
	}
	hostField := "Host Address: "
	if m.currentField == FieldHost {
		hostField = activeFieldStyle.Render(hostField + hostValue)
	} else {
		hostField = fieldStyle.Render(hostField + hostValue)
	}
	content.WriteString(hostField + "\n\n")
	
	// User field
	userValue := m.formData.User
	if m.currentField == FieldUser {
		userValue += "█"
	}
	userField := "Username: "
	if m.currentField == FieldUser {
		userField = activeFieldStyle.Render(userField + userValue)
	} else {
		userField = fieldStyle.Render(userField + userValue)
	}
	content.WriteString(userField + "\n\n")
	
	// Port field
	portValue := m.formData.Port
	if m.currentField == FieldPort {
		portValue += "█"
	}
	portField := "Port: "
	if m.currentField == FieldPort {
		portField = activeFieldStyle.Render(portField + portValue)
	} else {
		portField = fieldStyle.Render(portField + portValue)
	}
	content.WriteString(portField + "\n\n")
	
	// Show authentication info
	authInfo := "Authentication: "
	if m.formData.AuthType == AuthKey && m.formData.Identity != "" {
		authInfo += fmt.Sprintf("SSH Key (%s)", filepath.Base(m.formData.Identity))
	} else {
		authInfo += "Password"
	}
	content.WriteString(fieldStyle.Render(authInfo) + "\n\n")
	
	// Alias field
	aliasValue := m.formData.Alias
	if m.currentField == FieldAlias {
		aliasValue += "█"
	}
	aliasField := "Alias: "
	if m.currentField == FieldAlias {
		aliasField = activeFieldStyle.Render(aliasField + aliasValue)
	} else {
		aliasField = fieldStyle.Render(aliasField + aliasValue)
	}
	content.WriteString(aliasField + "\n\n")
	
	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(m.width)
	
	help := "Tab/↓: next field • Shift+Tab/↑: prev field • Enter: save • ESC: cancel"
	content.WriteString(helpStyle.Render(help))
	
	return content.String()
}

// renderDeleteView renders the delete confirmation
func (m Model) renderDeleteView() string {
	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#FF6B6B")).
		Padding(0, 1).
		Width(m.width)
	
	header := headerStyle.Render("Delete Host")
	content.WriteString(header + "\n\n")
	
	if len(m.filteredHosts) > 0 {
		host := m.filteredHosts[m.cursor]
		
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true).
			Align(lipgloss.Center).
			Width(m.width)
		
		warning := fmt.Sprintf("Are you sure you want to delete '%s'?", host.Name)
		content.WriteString(warningStyle.Render(warning) + "\n\n")
		
		// Show host details
		detailStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF6B6B")).
			Padding(1, 2).
			Width(m.width - 4)
		
		details := fmt.Sprintf("Host: %s\nUser: %s\nPort: %s", host.Host, host.User, host.Port)
		if host.Identity != "" {
			details += fmt.Sprintf("\nKey: %s", host.Identity)
		}
		
		content.WriteString(detailStyle.Render(details) + "\n\n")
	}
	
	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(m.width).
		Align(lipgloss.Center)
	
	help := "Y: confirm delete • N/ESC: cancel"
	content.WriteString(helpStyle.Render(help))
	
	return content.String()
}

// renderAuthSelectView renders authentication type selection
func (m Model) renderAuthSelectView() string {
	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)
	
	header := headerStyle.Render("Select Authentication Method")
	content.WriteString(header + "\n\n")
	
	// Options
	optionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(40).
		Margin(1, 0)
	
	option1 := optionStyle.Render("1. Password Authentication")
	option2 := optionStyle.Render("2. SSH Key Authentication")
	
	content.WriteString(option1 + "\n")
	content.WriteString(option2 + "\n\n")
	
	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(m.width)
	
	help := "1: password • 2: SSH key • ESC: back"
	content.WriteString(helpStyle.Render(help))
	
	return content.String()
}

// renderKeySelectView renders SSH key selection
func (m Model) renderKeySelectView() string {
	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)
	
	header := headerStyle.Render("Select SSH Key")
	content.WriteString(header + "\n\n")
	
	// Key list
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true)
	
	for i, keyFile := range m.keyFiles {
		cursor := "  "
		if m.keyCursor == i {
			cursor = "▶ "
		}
		
		keyName := filepath.Base(keyFile)
		keyDisplay := fmt.Sprintf("%s%s", cursor, keyName)
		
		if m.keyCursor == i {
			content.WriteString(selectedStyle.Render(keyDisplay) + "\n")
		} else {
			content.WriteString(keyDisplay + "\n")
		}
	}
	
	content.WriteString("\n")
	
	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(m.width)
	
	help := "↑/k: up • ↓/j: down • Enter: select • ESC: back"
	content.WriteString(helpStyle.Render(help))
	
	return content.String()
}

// renderPasswordInputView renders password input form
func (m Model) renderPasswordInputView() string {
	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)
	
	header := headerStyle.Render("Enter Password")
	content.WriteString(header + "\n\n")
	
	// Form info
	infoStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(m.width - 4)
	
	info := fmt.Sprintf("Host: %s\nUser: %s\nPort: %s", 
		m.formData.Host, m.formData.User, m.formData.Port)
	content.WriteString(infoStyle.Render(info) + "\n\n")
	
	// Password field
	fieldStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(0, 1).
		Width(40).
		Bold(true)
	
	// Show asterisks for password
	passwordDisplay := strings.Repeat("*", len(m.formData.Password)) + "█"
	passwordField := fieldStyle.Render("Password: " + passwordDisplay)
	content.WriteString(passwordField + "\n\n")
	
	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(m.width)
	
	help := "Type password • Enter: test connection • ESC: back"
	content.WriteString(helpStyle.Render(help))
	
	return content.String()
}

// renderKeyPasswordInputView renders SSH private key password input form
func (m Model) renderKeyPasswordInputView() string {
	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)
	
	header := headerStyle.Render("Enter SSH Key Password")
	content.WriteString(header + "\n\n")
	
	// Form info
	infoStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(m.width - 4)
	
	info := fmt.Sprintf("SSH Key: %s\nHost: %s\nUser: %s\nPort: %s", 
		filepath.Base(m.formData.Identity), m.formData.Host, m.formData.User, m.formData.Port)
	content.WriteString(infoStyle.Render(info) + "\n\n")
	
	// Password field
	fieldStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(0, 1).
		Width(40).
		Bold(true)
	
	// Show asterisks for password
	passwordDisplay := strings.Repeat("*", len(m.formData.KeyPassword)) + "█"
	passwordField := fieldStyle.Render("Key Password: " + passwordDisplay)
	content.WriteString(passwordField + "\n\n")
	
	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(m.width)
	
	help := "Type password • Enter: continue • ESC: back"
	content.WriteString(helpStyle.Render(help))
	
	return content.String()
}

// renderConnectTestView renders connection test and setup progress
func (m Model) renderConnectTestView() string {
	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)
	
	header := headerStyle.Render("Setting up SSH Connection")
	content.WriteString(header + "\n\n")
	
	// Host info
	infoStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(m.width - 4)
	
	info := fmt.Sprintf("Host: %s\nUser: %s\nPort: %s\nAuth: %s", 
		m.formData.Host, m.formData.User, m.formData.Port,
		map[AuthType]string{AuthPassword: "Password", AuthKey: "SSH Key"}[m.formData.AuthType])
	content.WriteString(infoStyle.Render(info) + "\n\n")
	
	// Progress
	progressStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF00")).
		Padding(1, 2).
		Width(m.width - 4).
		Align(lipgloss.Center)
	
	if m.isSetupDone {
		progressStyle = progressStyle.BorderForeground(lipgloss.Color("#00FF00"))
		content.WriteString(progressStyle.Render("✓ Setup completed successfully!") + "\n\n")
	} else {
		progressStyle = progressStyle.BorderForeground(lipgloss.Color("#FFFF00"))
		content.WriteString(progressStyle.Render("⏳ " + m.setupProgress) + "\n\n")
	}
	
	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(m.width)
	
	var help string
	if m.isSetupDone {
		help = "Enter: save and continue • ESC: cancel"
	} else {
		help = "Please wait... • ESC: cancel"
	}
	content.WriteString(helpStyle.Render(help))
	
	return content.String()
}