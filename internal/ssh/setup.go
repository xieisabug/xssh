package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
	"xssh/internal/config"
)

// SetupResult represents the result of SSH setup
type SetupResult struct {
	Success bool
	Message string
	Error   error
}

// TestConnection tests SSH connection and performs setup if needed
func TestConnection(host config.SSHHost, password string) SetupResult {
	// First, test if we can connect
	if host.Identity != "" {
		// Test key-based connection
		return testKeyConnection(host)
	} else {
		// Test password connection and set up keys
		return testPasswordConnectionAndSetupKeys(host, password)
	}
}

// TestConnectionWithKeyPassword tests SSH connection with key password
func TestConnectionWithKeyPassword(host config.SSHHost, keyPassword string) SetupResult {
	if host.Identity != "" {
		// Test key-based connection with password
		return testKeyConnectionWithPassword(host, keyPassword)
	} else {
		return SetupResult{
			Success: false,
			Message: "No SSH key specified",
			Error:   fmt.Errorf("no SSH key specified"),
		}
	}
}

// testKeyConnection tests SSH key-based connection
func testKeyConnection(host config.SSHHost) SetupResult {
	return testKeyConnectionWithPassword(host, "")
}

// testKeyConnectionWithPassword tests SSH key-based connection with optional password
func testKeyConnectionWithPassword(host config.SSHHost, keyPassword string) SetupResult {
	// Read private key
	keyData, err := os.ReadFile(host.Identity)
	if err != nil {
		return SetupResult{
			Success: false,
			Message: fmt.Sprintf("Failed to read private key: %v", err),
			Error:   err,
		}
	}

	// Parse private key
	var key ssh.Signer
	if keyPassword != "" {
		// Try to parse encrypted key with password
		key, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(keyPassword))
	} else {
		// Try to parse unencrypted key
		key, err = ssh.ParsePrivateKey(keyData)
	}
	
	if err != nil {
		return SetupResult{
			Success: false,
			Message: fmt.Sprintf("Failed to parse private key: %v", err),
			Error:   err,
		}
	}

	// Create SSH client config
	config := &ssh.ClientConfig{
		User: host.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production, use proper host key checking
		Timeout:         10 * time.Second,
	}

	// Test connection
	client, err := ssh.Dial("tcp", host.Host+":"+host.Port, config)
	if err != nil {
		return SetupResult{
			Success: false,
			Message: fmt.Sprintf("Failed to connect with SSH key: %v", err),
			Error:   err,
		}
	}
	defer client.Close()

	return SetupResult{
		Success: true,
		Message: "SSH key connection successful",
	}
}

// testPasswordConnectionAndSetupKeys tests password connection and sets up SSH keys
func testPasswordConnectionAndSetupKeys(host config.SSHHost, password string) SetupResult {
	// First, test password connection
	config := &ssh.ClientConfig{
		User: host.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production, use proper host key checking
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", host.Host+":"+host.Port, config)
	if err != nil {
		return SetupResult{
			Success: false,
			Message: fmt.Sprintf("Failed to connect with password: %v", err),
			Error:   err,
		}
	}
	client.Close()

	// If password connection works, set up SSH keys
	return setupSSHKeys(host, password)
}

// setupSSHKeys sets up SSH key authentication
func setupSSHKeys(host config.SSHHost, password string) SetupResult {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return SetupResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get home directory: %v", err),
			Error:   err,
		}
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	privateKeyPath := filepath.Join(sshDir, "id_rsa")
	publicKeyPath := filepath.Join(sshDir, "id_rsa.pub")

	// Check if SSH key already exists
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		// Generate SSH key pair
		result := generateSSHKeyPair(privateKeyPath, publicKeyPath)
		if !result.Success {
			return result
		}
	}

	// Copy public key to remote server using ssh-copy-id equivalent
	return copyPublicKey(host, password, publicKeyPath)
}

// generateSSHKeyPair generates a new SSH key pair
func generateSSHKeyPair(privateKeyPath, publicKeyPath string) SetupResult {
	// Use ssh-keygen command to generate key pair
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048", "-f", privateKeyPath, "-N", "")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return SetupResult{
			Success: false,
			Message: fmt.Sprintf("Failed to generate SSH key: %v\nOutput: %s", err, output),
			Error:   err,
		}
	}

	return SetupResult{
		Success: true,
		Message: "SSH key pair generated successfully",
	}
}

// copyPublicKey copies the public key to the remote server
func copyPublicKey(host config.SSHHost, password string, publicKeyPath string) SetupResult {
	// Read public key
	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return SetupResult{
			Success: false,
			Message: fmt.Sprintf("Failed to read public key: %v", err),
			Error:   err,
		}
	}

	// Connect to remote server with password
	config := &ssh.ClientConfig{
		User: host.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	client, err := ssh.Dial("tcp", host.Host+":"+host.Port, config)
	if err != nil {
		return SetupResult{
			Success: false,
			Message: fmt.Sprintf("Failed to connect to remote server: %v", err),
			Error:   err,
		}
	}
	defer client.Close()

	// Create SSH session
	session, err := client.NewSession()
	if err != nil {
		return SetupResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create SSH session: %v", err),
			Error:   err,
		}
	}
	defer session.Close()

	// Create .ssh directory and authorized_keys file on remote server
	commands := []string{
		"mkdir -p ~/.ssh",
		"chmod 700 ~/.ssh",
		fmt.Sprintf("echo '%s' >> ~/.ssh/authorized_keys", string(publicKey)),
		"chmod 600 ~/.ssh/authorized_keys",
	}

	for _, cmd := range commands {
		session, err := client.NewSession()
		if err != nil {
			return SetupResult{
				Success: false,
				Message: fmt.Sprintf("Failed to create session for command '%s': %v", cmd, err),
				Error:   err,
			}
		}

		err = session.Run(cmd)
		session.Close()

		if err != nil {
			return SetupResult{
				Success: false,
				Message: fmt.Sprintf("Failed to execute command '%s': %v", cmd, err),
				Error:   err,
			}
		}
	}

	// Test key-based connection
	privateKeyPath := filepath.Join(filepath.Dir(publicKeyPath), "id_rsa")
	testHost := host
	testHost.Identity = privateKeyPath

	return testKeyConnection(testHost)
}