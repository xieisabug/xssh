package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"xssh/internal/config"
	"xssh/internal/forwarding"
)

// CLIOptions holds all command-line options
type CLIOptions struct {
	ShowHelp          bool
	ShowVersion       bool
	ForwardingRule    *forwarding.ForwardingRule
	HostAlias         string
	ListHosts         bool
	ListForwarding    bool
	StopForwarding    string
	Interactive       bool
	ConnectOnly       bool
}

// ParseArgs parses command line arguments and returns CLIOptions
func ParseArgs() (*CLIOptions, error) {
	opts := &CLIOptions{
		Interactive: true, // Default to interactive mode
	}

	// Custom flag handling since we want to support both -f and --forward formats
	args := os.Args[1:]
	
	for i := 0; i < len(args); i++ {
		arg := args[i]
		
		switch {
		case arg == "-h" || arg == "--help":
			opts.ShowHelp = true
			opts.Interactive = false
			return opts, nil
			
		case arg == "-v" || arg == "--version":
			opts.ShowVersion = true
			opts.Interactive = false
			return opts, nil
			
		case arg == "-l" || arg == "--list":
			opts.ListHosts = true
			opts.Interactive = false
			
		case arg == "--list-forwarding":
			opts.ListForwarding = true
			opts.Interactive = false
			
		case arg == "--stop-forwarding":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("option %s requires an argument", arg)
			}
			i++
			opts.StopForwarding = args[i]
			opts.Interactive = false
			
		case arg == "-c" || arg == "--connect":
			opts.ConnectOnly = true
			opts.Interactive = false
			
		case arg == "-f" || arg == "--forward":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("option %s requires an argument", arg)
			}
			i++
			rule, err := parseForwardingRule(args[i])
			if err != nil {
				return nil, fmt.Errorf("invalid forwarding rule: %v", err)
			}
			opts.ForwardingRule = rule
			opts.Interactive = false
			
			// Next argument might be host alias
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				opts.HostAlias = args[i]
			}
			
		case !strings.HasPrefix(arg, "-"):
			// This is likely a host alias
			opts.HostAlias = arg
			opts.Interactive = false
			
		default:
			return nil, fmt.Errorf("unknown option: %s", arg)
		}
	}
	
	return opts, nil
}

// parseForwardingRule parses a forwarding rule string
// Supports formats:
// - "8080:localhost:80" (local forwarding)
// - "R:8080:localhost:80" (remote forwarding)  
// - "D:1080" (dynamic forwarding/SOCKS proxy)
func parseForwardingRule(ruleStr string) (*forwarding.ForwardingRule, error) {
	parts := strings.Split(ruleStr, ":")
	
	rule := &forwarding.ForwardingRule{
		ID: fmt.Sprintf("cli-%d", len(ruleStr)), // Simple ID generation
	}
	
	if len(parts) == 2 && strings.ToUpper(parts[0]) == "D" {
		// Dynamic forwarding: D:1080
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %s", parts[1])
		}
		rule.Type = forwarding.DynamicForward
		rule.LocalHost = "localhost"
		rule.LocalPort = port
		rule.Description = fmt.Sprintf("SOCKS proxy on port %d", port)
		return rule, nil
	}
	
	if len(parts) == 4 && strings.ToUpper(parts[0]) == "R" {
		// Remote forwarding: R:8080:localhost:80
		localPort, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid local port: %s", parts[1])
		}
		remotePort, err := strconv.Atoi(parts[3])
		if err != nil {
			return nil, fmt.Errorf("invalid remote port: %s", parts[3])
		}
		
		rule.Type = forwarding.RemoteForward
		rule.LocalHost = "localhost"
		rule.LocalPort = localPort
		rule.RemoteHost = parts[2]
		rule.RemotePort = remotePort
		rule.Description = fmt.Sprintf("Remote %d -> %s:%d", localPort, parts[2], remotePort)
		return rule, nil
	}
	
	if len(parts) == 3 {
		// Local forwarding: 8080:localhost:80
		localPort, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid local port: %s", parts[0])
		}
		remotePort, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid remote port: %s", parts[2])
		}
		
		rule.Type = forwarding.LocalForward
		rule.LocalHost = "localhost"
		rule.LocalPort = localPort
		rule.RemoteHost = parts[1]
		rule.RemotePort = remotePort
		rule.Description = fmt.Sprintf("Local %d -> %s:%d", localPort, parts[1], remotePort)
		return rule, nil
	}
	
	return nil, fmt.Errorf("invalid forwarding rule format. Use: [R:]local_port:remote_host:remote_port or D:port")
}

// ShowHelp displays help information
func ShowHelp() {
	fmt.Println("xssh - SSH Connection Manager with Port Forwarding")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  xssh [OPTIONS] [HOST_ALIAS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -h, --help                     Show this help message")
	fmt.Println("  -v, --version                  Show version information")
	fmt.Println("  -l, --list                     List all configured SSH hosts")
	fmt.Println("  -c, --connect HOST             Connect to specified host")
	fmt.Println("  -f, --forward RULE [HOST]      Start port forwarding with specified rule")
	fmt.Println("  --list-forwarding              List all active port forwarding sessions")
	fmt.Println("  --stop-forwarding ID           Stop a specific forwarding session")
	fmt.Println()
	fmt.Println("PORT FORWARDING RULES:")
	fmt.Println("  Local forwarding:    8080:localhost:80")
	fmt.Println("                      Forward local port 8080 to remote localhost:80")
	fmt.Println()
	fmt.Println("  Remote forwarding:   R:8080:localhost:80")
	fmt.Println("                      Forward remote port 8080 to local localhost:80")
	fmt.Println()
	fmt.Println("  Dynamic forwarding:  D:1080")
	fmt.Println("                      Create SOCKS5 proxy on local port 1080")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  xssh                           # Start interactive mode")
	fmt.Println("  xssh myserver                  # Connect to 'myserver' host")
	fmt.Println("  xssh -c myserver               # Connect to 'myserver' host")
	fmt.Println("  xssh -l                        # List all configured hosts")
	fmt.Println("  xssh -f 8080:localhost:80 web  # Forward port 8080 to web server")
	fmt.Println("  xssh -f R:9000:db:5432 proxy   # Remote forward port 9000 to database")
	fmt.Println("  xssh -f D:1080 gateway         # Create SOCKS proxy through gateway")
	fmt.Println("  xssh --list-forwarding         # Show active forwarding sessions")
	fmt.Println("  xssh --stop-forwarding cli-123 # Stop forwarding session")
}

// ShowVersion displays version information
func ShowVersion() {
	fmt.Println("xssh v1.0.0")
	fmt.Println("SSH Connection Manager with Port Forwarding")
	fmt.Println("Built with Go and Bubbletea TUI framework")
}

// ListHosts displays all configured SSH hosts
func ListHosts() error {
	sshConfig, err := config.LoadSSHConfig()
	if err != nil {
		return fmt.Errorf("failed to load SSH config: %v", err)
	}
	
	if len(sshConfig.Hosts) == 0 {
		fmt.Println("No SSH hosts configured.")
		fmt.Println("Run 'xssh' to enter interactive mode and add hosts.")
		return nil
	}
	
	fmt.Println("Configured SSH Hosts:")
	fmt.Println()
	
	for _, host := range sshConfig.Hosts {
		fmt.Printf("  %s\n", host.Name)
		fmt.Printf("    Host: %s@%s:%s\n", host.User, host.Host, host.Port)
		if host.Identity != "" {
			fmt.Printf("    Key:  %s\n", host.Identity)
		}
		fmt.Println()
	}
	
	return nil
}