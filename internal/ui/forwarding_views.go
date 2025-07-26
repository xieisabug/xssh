package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"xssh/internal/forwarding"
)

// renderForwardingSelectView renders the forwarding type selection
func (m Model) renderForwardingSelectView() string {
	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)
	
	header := headerStyle.Render("Select Port Forwarding Type")
	content.WriteString(header + "\n\n")
	
	// Host info
	if m.selectedHostIndex >= 0 && m.selectedHostIndex < len(m.filteredHosts) {
		host := m.filteredHosts[m.selectedHostIndex]
		infoStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			Width(m.width - 4)
		
		info := fmt.Sprintf("Target Host: %s (%s@%s:%s)", host.Name, host.User, host.Host, host.Port)
		content.WriteString(infoStyle.Render(info) + "\n\n")
	}
	
	// Options
	optionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(m.width - 8).
		Margin(1, 2)
	
	option1 := optionStyle.Render("1. Local Forward (-L)\n   Forward local port to remote host through SSH tunnel")
	option2 := optionStyle.Render("2. Remote Forward (-R)\n   Forward remote port to local host")
	option3 := optionStyle.Render("3. Dynamic Forward (-D)\n   Create SOCKS5 proxy on local port")
	optionList := optionStyle.Render("L. List Active Forwardings\n   View and manage active port forwarding sessions")
	
	content.WriteString(option1 + "\n")
	content.WriteString(option2 + "\n")
	content.WriteString(option3 + "\n")
	content.WriteString(optionList + "\n\n")
	
	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(m.width)
	
	help := "1/2/3: select forwarding type • L: list active • ESC: back"
	content.WriteString(helpStyle.Render(help))
	
	return content.String()
}

// renderForwardingAddView renders the forwarding configuration form
func (m Model) renderForwardingAddView() string {
	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)
	
	title := fmt.Sprintf("Configure %s Forwarding", m.forwardingType.String())
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
	
	// Show different fields based on forwarding type
	switch m.forwardingType {
	case forwarding.LocalForward:
		// Local Port
		localPortValue := m.formData.LocalPort
		if m.currentField == FieldLocalPort {
			localPortValue += "█"
		}
		localPortField := "Local Port: "
		if m.currentField == FieldLocalPort {
			localPortField = activeFieldStyle.Render(localPortField + localPortValue)
		} else {
			localPortField = fieldStyle.Render(localPortField + localPortValue)
		}
		content.WriteString(localPortField + "\n\n")
		
		// Remote Host
		remoteHostValue := m.formData.RemoteHost
		var remoteHostDisplay string
		
		if m.formData.UseExistingHost && m.formData.SelectedRemoteHostIndex < len(m.hosts) {
			// Show selected host info
			selectedHost := m.hosts[m.formData.SelectedRemoteHostIndex]
			remoteHostDisplay = fmt.Sprintf("%s (%s)", remoteHostValue, selectedHost.Name)
		} else if m.formData.RemoteHost != "" {
			// Show manual input
			remoteHostDisplay = remoteHostValue
		} else {
			// Show prompt to select
			remoteHostDisplay = "Press Enter to select host"
		}
		
		if m.currentField == FieldRemoteHost {
			if m.formData.RemoteHost == "" {
				remoteHostDisplay += " █"
			} else {
				remoteHostDisplay += "█"
			}
		}
		
		remoteHostField := "Remote Host: "
		if m.currentField == FieldRemoteHost {
			remoteHostField = activeFieldStyle.Render(remoteHostField + remoteHostDisplay)
		} else {
			remoteHostField = fieldStyle.Render(remoteHostField + remoteHostDisplay)
		}
		content.WriteString(remoteHostField + "\n\n")
		
		// Remote Port
		remotePortValue := m.formData.RemotePort
		if m.currentField == FieldRemotePort {
			remotePortValue += "█"
		}
		remotePortField := "Remote Port: "
		if m.currentField == FieldRemotePort {
			remotePortField = activeFieldStyle.Render(remotePortField + remotePortValue)
		} else {
			remotePortField = fieldStyle.Render(remotePortField + remotePortValue)
		}
		content.WriteString(remotePortField + "\n\n")
		
	case forwarding.RemoteForward:
		// Remote Port
		remotePortValue := m.formData.RemotePort
		if m.currentField == FieldRemotePort {
			remotePortValue += "█"
		}
		remotePortField := "Remote Port: "
		if m.currentField == FieldRemotePort {
			remotePortField = activeFieldStyle.Render(remotePortField + remotePortValue)
		} else {
			remotePortField = fieldStyle.Render(remotePortField + remotePortValue)
		}
		content.WriteString(remotePortField + "\n\n")
		
		// Local Port
		localPortValue := m.formData.LocalPort
		if m.currentField == FieldLocalPort {
			localPortValue += "█"
		}
		localPortField := "Local Port: "
		if m.currentField == FieldLocalPort {
			localPortField = activeFieldStyle.Render(localPortField + localPortValue)
		} else {
			localPortField = fieldStyle.Render(localPortField + localPortValue)
		}
		content.WriteString(localPortField + "\n\n")
		
	case forwarding.DynamicForward:
		// Local Port only
		localPortValue := m.formData.LocalPort
		if m.currentField == FieldLocalPort {
			localPortValue += "█"
		}
		localPortField := "SOCKS5 Port: "
		if m.currentField == FieldLocalPort {
			localPortField = activeFieldStyle.Render(localPortField + localPortValue)
		} else {
			localPortField = fieldStyle.Render(localPortField + localPortValue)
		}
		content.WriteString(localPortField + "\n\n")
	}
	
	// Description field (always shown)
	descValue := m.formData.Description
	if m.currentField == FieldDescription {
		descValue += "█"
	}
	descField := "Description: "
	if m.currentField == FieldDescription {
		descField = activeFieldStyle.Render(descField + descValue)
	} else {
		descField = fieldStyle.Render(descField + descValue)
	}
	content.WriteString(descField + "\n\n")
	
	// Example command
	exampleStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#888888")).
		Padding(1, 2).
		Width(m.width - 4).
		Foreground(lipgloss.Color("#888888"))
	
	var example string
	switch m.forwardingType {
	case forwarding.LocalForward:
		if m.formData.LocalPort != "" && m.formData.RemoteHost != "" && m.formData.RemotePort != "" {
			var hostInfo string
			if m.formData.UseExistingHost && m.formData.SelectedRemoteHostIndex < len(m.hosts) {
				selectedHost := m.hosts[m.formData.SelectedRemoteHostIndex]
				hostInfo = fmt.Sprintf("via %s", selectedHost.Name)
			} else {
				hostInfo = "via SSH tunnel"
			}
			example = fmt.Sprintf("Equivalent: ssh -L %s:%s:%s user@host (%s)", 
				m.formData.LocalPort, m.formData.RemoteHost, m.formData.RemotePort, hostInfo)
		} else {
			example = "Example: ssh -L 8080:google.com:80 user@host"
		}
	case forwarding.RemoteForward:
		if m.formData.RemotePort != "" && m.formData.LocalPort != "" {
			example = fmt.Sprintf("Equivalent: ssh -R %s:localhost:%s user@host", m.formData.RemotePort, m.formData.LocalPort)
		} else {
			example = "Example: ssh -R 8080:localhost:3000 user@host"
		}
	case forwarding.DynamicForward:
		if m.formData.LocalPort != "" {
			example = fmt.Sprintf("Equivalent: ssh -D %s user@host", m.formData.LocalPort)
		} else {
			example = "Example: ssh -D 1080 user@host"
		}
	}
	content.WriteString(exampleStyle.Render(example) + "\n\n")
	
	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(m.width)
	
	var help string
	if m.currentField == FieldRemoteHost && m.forwardingType == forwarding.LocalForward {
		help = "Tab: next field • Enter: select remote host • ESC: back"
	} else {
		help = "Tab: next field • Enter: start forwarding • ESC: back"
	}
	content.WriteString(helpStyle.Render(help))
	
	return content.String()
}

// renderForwardingListView renders the list of active forwarding sessions
func (m Model) renderForwardingListView() string {
	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)
	
	header := headerStyle.Render("Active Port Forwarding Sessions")
	content.WriteString(header + "\n\n")
	
	// Get active sessions
	sessions := m.forwardingManager.GetAllSessions()
	
	if len(sessions) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999")).
			Italic(true).
			Align(lipgloss.Center).
			Width(m.width)
		
		content.WriteString(emptyStyle.Render("No active port forwarding sessions") + "\n\n")
	} else {
		// Session list
		selectedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true)
		
		sessionStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			Width(m.width - 4).
			Margin(0, 0, 1, 0)
		
		for i, session := range sessions {
			cursor := "  "
			if m.cursor == i {
				cursor = "▶ "
			}
			
			// Format session info
			var sessionInfo string
			switch session.Rule.Type {
			case forwarding.LocalForward:
				sessionInfo = fmt.Sprintf("%s%s: Local:%d → %s:%d",
					cursor, session.Rule.Type.String(),
					session.Rule.LocalPort, session.Rule.RemoteHost, session.Rule.RemotePort)
			case forwarding.RemoteForward:
				sessionInfo = fmt.Sprintf("%s%s: Remote:%d → Local:%d",
					cursor, session.Rule.Type.String(),
					session.Rule.RemotePort, session.Rule.LocalPort)
			case forwarding.DynamicForward:
				sessionInfo = fmt.Sprintf("%s%s: SOCKS5 on port %d",
					cursor, session.Rule.Type.String(), session.Rule.LocalPort)
			}
			
			if session.Rule.Description != "" {
				sessionInfo += fmt.Sprintf(" (%s)", session.Rule.Description)
			}
			
			// Add statistics
			uptime := session.GetUptime()
			rxRate, txRate := session.GetTransferRate()
			statsInfo := fmt.Sprintf("\nUptime: %v | Connections: %d active, %d total",
				uptime.Round(time.Second),
				session.Stats.ActiveConnections,
				session.Stats.ConnectionCount)
			
			if session.Stats.BytesReceived > 0 || session.Stats.BytesSent > 0 {
				statsInfo += fmt.Sprintf("\nTraffic: ↓%.1fKB (%.1fKB/s) ↑%.1fKB (%.1fKB/s)",
					float64(session.Stats.BytesReceived)/1024, rxRate/1024,
					float64(session.Stats.BytesSent)/1024, txRate/1024)
			}
			
			if session.Stats.ErrorCount > 0 {
				statsInfo += fmt.Sprintf("\nErrors: %d (Last: %s)",
					session.Stats.ErrorCount, session.Stats.LastError)
			}
			
			sessionDisplay := sessionInfo + statsInfo
			
			sessionBox := sessionStyle.Render(sessionDisplay)
			if m.cursor == i {
				sessionBox = selectedStyle.Render(sessionBox)
			}
			
			content.WriteString(sessionBox + "\n")
		}
	}
	
	// Summary statistics if there are sessions
	if len(sessions) > 0 {
		summaryStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FF00")).
			Padding(1, 2).
			Width(m.width - 4).
			Bold(true)
		
		totalConnections := int64(0)
		totalBytes := int64(0)
		totalErrors := int64(0)
		
		for _, session := range sessions {
			totalConnections += session.Stats.ConnectionCount
			totalBytes += session.Stats.BytesReceived + session.Stats.BytesSent
			totalErrors += session.Stats.ErrorCount
		}
		
		summary := fmt.Sprintf("Summary: %d sessions | %d total connections | %.1f MB transferred | %d errors",
			len(sessions), totalConnections, float64(totalBytes)/(1024*1024), totalErrors)
		
		content.WriteString(summaryStyle.Render(summary) + "\n\n")
	}
	
	// Message
	if m.message != "" {
		messageStyle := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center)
		
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
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(m.width)
	
	help := "↑/k: up • ↓/j: down • s: stop selected • a: add new • ESC/q: back"
	content.WriteString(helpStyle.Render(help))
	
	return content.String()
}

// renderRemoteHostSelectView renders the remote host selection view
func (m Model) renderRemoteHostSelectView() string {
	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)
	
	header := headerStyle.Render("Select Remote Host")
	content.WriteString(header + "\n\n")
	
	// Instructions
	infoStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(m.width - 4)
	
	info := "Choose an existing SSH host as the remote host, or select 'Manual Input' to enter a custom host address."
	content.WriteString(infoStyle.Render(info) + "\n\n")
	
	// Host list
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true)
	
	hostStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width - 8).
		Margin(0, 2)
	
	// Show existing hosts
	for i, host := range m.hosts {
		cursor := "  "
		if m.cursor == i {
			cursor = "▶ "
		}
		
		hostDisplay := fmt.Sprintf("%s%s (%s@%s:%s)", cursor, host.Name, host.User, host.Host, host.Port)
		
		if m.cursor == i {
			content.WriteString(selectedStyle.Render(hostStyle.Render(hostDisplay)) + "\n")
		} else {
			content.WriteString(hostStyle.Render(hostDisplay) + "\n")
		}
	}
	
	// Manual input option
	cursor := "  "
	if m.cursor == len(m.hosts) {
		cursor = "▶ "
	}
	
	manualOption := fmt.Sprintf("%s📝 Manual Input (Enter custom host address)", cursor)
	manualStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(0, 1).
		Width(m.width - 8).
		Margin(1, 2).
		Italic(true)
	
	if m.cursor == len(m.hosts) {
		content.WriteString(selectedStyle.Render(manualStyle.Render(manualOption)) + "\n\n")
	} else {
		content.WriteString(manualStyle.Render(manualOption) + "\n\n")
	}
	
	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(m.width)
	
	help := "↑/k: up • ↓/j: down • Enter: select • ESC: back"
	content.WriteString(helpStyle.Render(help))
	
	return content.String()
}