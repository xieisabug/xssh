package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SSHHost represents a single SSH host configuration
type SSHHost struct {
	Name     string
	Host     string
	User     string
	Port     string
	Identity string
}

// SSHConfig holds all SSH hosts
type SSHConfig struct {
	Hosts []SSHHost
	Path  string
}

// LoadSSHConfig reads and parses SSH config file
func LoadSSHConfig() (*SSHConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".ssh", "config")
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create empty config if it doesn't exist
			return &SSHConfig{
				Hosts: []SSHHost{},
				Path:  configPath,
			}, nil
		}
		return nil, err
	}
	defer file.Close()

	config := &SSHConfig{
		Hosts: []SSHHost{},
		Path:  configPath,
	}

	scanner := bufio.NewScanner(file)
	var currentHost *SSHHost

	hostRegex := regexp.MustCompile(`^Host\s+(.+)$`)
	hostNameRegex := regexp.MustCompile(`^\s*HostName\s+(.+)$`)
	userRegex := regexp.MustCompile(`^\s*User\s+(.+)$`)
	portRegex := regexp.MustCompile(`^\s*Port\s+(.+)$`)
	identityRegex := regexp.MustCompile(`^\s*IdentityFile\s+(.+)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if matches := hostRegex.FindStringSubmatch(line); matches != nil {
			// Save previous host if exists
			if currentHost != nil {
				config.Hosts = append(config.Hosts, *currentHost)
			}
			
			// Start new host
			hostName := strings.TrimSpace(matches[1])
			currentHost = &SSHHost{
				Name: hostName,
				Host: hostName, // Default to name
				Port: "22",     // Default port
			}
		} else if currentHost != nil {
			if matches := hostNameRegex.FindStringSubmatch(line); matches != nil {
				currentHost.Host = strings.TrimSpace(matches[1])
			} else if matches := userRegex.FindStringSubmatch(line); matches != nil {
				currentHost.User = strings.TrimSpace(matches[1])
			} else if matches := portRegex.FindStringSubmatch(line); matches != nil {
				currentHost.Port = strings.TrimSpace(matches[1])
			} else if matches := identityRegex.FindStringSubmatch(line); matches != nil {
				currentHost.Identity = strings.TrimSpace(matches[1])
			}
		}
	}

	// Don't forget the last host
	if currentHost != nil {
		config.Hosts = append(config.Hosts, *currentHost)
	}

	return config, scanner.Err()
}

// SaveSSHConfig writes the config back to file
func (c *SSHConfig) Save() error {
	file, err := os.Create(c.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, host := range c.Hosts {
		fmt.Fprintf(writer, "Host %s\n", host.Name)
		fmt.Fprintf(writer, "    HostName %s\n", host.Host)
		if host.User != "" {
			fmt.Fprintf(writer, "    User %s\n", host.User)
		}
		if host.Port != "22" && host.Port != "" {
			fmt.Fprintf(writer, "    Port %s\n", host.Port)
		}
		if host.Identity != "" {
			fmt.Fprintf(writer, "    IdentityFile %s\n", host.Identity)
		}
		fmt.Fprintln(writer)
	}

	return nil
}

// AddHost adds a new host to the configuration at the beginning
func (c *SSHConfig) AddHost(host SSHHost) {
	c.Hosts = append([]SSHHost{host}, c.Hosts...)
}

// RemoveHost removes a host by name
func (c *SSHConfig) RemoveHost(name string) {
	for i, host := range c.Hosts {
		if host.Name == name {
			c.Hosts = append(c.Hosts[:i], c.Hosts[i+1:]...)
			break
		}
	}
}

// UpdateHost updates an existing host
func (c *SSHConfig) UpdateHost(name string, updatedHost SSHHost) {
	for i, host := range c.Hosts {
		if host.Name == name {
			c.Hosts[i] = updatedHost
			break
		}
	}
}