package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"xssh/internal/cli"
	"xssh/internal/config"
	"xssh/internal/forwarding"
	"xssh/internal/ssh"
	"xssh/internal/ui"
)

func main() {
	// Parse command line arguments
	opts, err := cli.ParseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Use 'xssh --help' for usage information.\n")
		os.Exit(1)
	}

	// Handle non-interactive modes
	if !opts.Interactive {
		if err := handleNonInteractiveMode(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Start interactive TUI mode
	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	
	model, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	// Check if we need to connect to a host
	if finalModel, ok := model.(ui.Model); ok {
		if selectedHost := finalModel.GetSelectedHost(); selectedHost != nil {
			// Connect to the selected host
			fmt.Printf("Connecting to %s...\n", selectedHost.Name)
			if err := ssh.ConnectToHost(*selectedHost); err != nil {
				fmt.Printf("Failed to connect: %v\n", err)
				os.Exit(1)
			}
		}
	}
}

// handleNonInteractiveMode handles command-line only operations
func handleNonInteractiveMode(opts *cli.CLIOptions) error {
	if opts.ShowHelp {
		cli.ShowHelp()
		return nil
	}

	if opts.ShowVersion {
		cli.ShowVersion()
		return nil
	}

	if opts.ListHosts {
		return cli.ListHosts()
	}

	if opts.ListForwarding {
		return listActiveForwarding()
	}

	if opts.StopForwarding != "" {
		return stopForwardingSession(opts.StopForwarding)
	}

	if opts.ForwardingRule != nil {
		return handlePortForwarding(opts.ForwardingRule, opts.HostAlias)
	}

	if opts.HostAlias != "" {
		return connectToHostByAlias(opts.HostAlias)
	}

	return nil
}

// listActiveForwarding lists all active port forwarding sessions
func listActiveForwarding() error {
	manager := forwarding.NewManager()
	sessions := manager.GetAllSessions()
	
	if len(sessions) == 0 {
		fmt.Println("No active port forwarding sessions.")
		return nil
	}
	
	fmt.Println("Active Port Forwarding Sessions:")
	fmt.Println()
	
	for _, session := range sessions {
		fmt.Printf("  %s (%s)\n", session.Rule.ID, session.Rule.Type.String())
		fmt.Printf("    %s\n", session.Rule.Description)
		fmt.Printf("    Active: %v, Uptime: %v\n", session.IsActive(), session.GetUptime().Round(time.Second))
		fmt.Printf("    Connections: %d active, %d total\n", 
			session.Stats.ActiveConnections, session.Stats.ConnectionCount)
		if session.Stats.BytesReceived > 0 || session.Stats.BytesSent > 0 {
			fmt.Printf("    Data: %d bytes received, %d bytes sent\n", 
				session.Stats.BytesReceived, session.Stats.BytesSent)
		}
		fmt.Println()
	}
	
	return nil
}

// stopForwardingSession stops a specific port forwarding session
func stopForwardingSession(sessionID string) error {
	manager := forwarding.NewManager()
	
	// Check if session exists
	if _, exists := manager.GetSession(sessionID); !exists {
		return fmt.Errorf("forwarding session '%s' not found", sessionID)
	}
	
	// Stop the session
	if err := manager.StopForwarding(sessionID); err != nil {
		return fmt.Errorf("failed to stop forwarding session: %v", err)
	}
	
	fmt.Printf("Stopped port forwarding session: %s\n", sessionID)
	return nil
}

// handlePortForwarding starts a port forwarding session
func handlePortForwarding(rule *forwarding.ForwardingRule, hostAlias string) error {
	if hostAlias == "" {
		return fmt.Errorf("host alias is required for port forwarding")
	}
	
	// Load SSH config to find the host
	sshConfig, err := config.LoadSSHConfig()
	if err != nil {
		return fmt.Errorf("failed to load SSH config: %v", err)
	}
	
	var targetHost *config.SSHHost
	for _, host := range sshConfig.Hosts {
		if host.Name == hostAlias {
			targetHost = &host
			break
		}
	}
	
	if targetHost == nil {
		return fmt.Errorf("host '%s' not found in SSH config", hostAlias)
	}
	
	// Start port forwarding
	manager := forwarding.NewManager()
	fmt.Printf("Starting port forwarding: %s\n", rule.Description)
	fmt.Printf("Connecting to %s@%s:%s\n", targetHost.User, targetHost.Host, targetHost.Port)
	
	if err := manager.StartForwarding(*rule, *targetHost, ""); err != nil {
		return fmt.Errorf("failed to start port forwarding: %v", err)
	}
	
	fmt.Printf("Port forwarding active. Press Ctrl+C to stop.\n")
	
	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Wait for interrupt signal
	<-sigChan
	fmt.Printf("\nShutting down port forwarding...\n")
	manager.StopForwarding(rule.ID)
	
	return nil
}

// connectToHostByAlias connects to a specific host by alias
func connectToHostByAlias(alias string) error {
	// Load SSH config to find the host
	sshConfig, err := config.LoadSSHConfig()
	if err != nil {
		return fmt.Errorf("failed to load SSH config: %v", err)
	}
	
	var targetHost *config.SSHHost
	for _, host := range sshConfig.Hosts {
		if host.Name == alias {
			targetHost = &host
			break
		}
	}
	
	if targetHost == nil {
		return fmt.Errorf("host '%s' not found in SSH config", alias)
	}
	
	// Connect to the host
	fmt.Printf("Connecting to %s...\n", targetHost.Name)
	if err := ssh.ConnectToHost(*targetHost); err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	
	return nil
}