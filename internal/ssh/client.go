package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/atotto/clipboard"
	"xssh/internal/config"
)

// ConnectToHost connects to SSH host using system ssh command
// This will properly handle terminal I/O and restore terminal state
func ConnectToHost(host config.SSHHost) error {
	args := []string{"ssh"}

	if host.User != "" {
		args = append(args, "-l", host.User)
	}

	if host.Port != "22" && host.Port != "" {
		args = append(args, "-p", host.Port)
	}

	if host.Identity != "" {
		args = append(args, "-i", host.Identity)
	}

	args = append(args, host.Host)

	// Find ssh binary
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh command not found: %v", err)
	}

	// Use syscall.Exec to replace current process with SSH
	// This ensures proper terminal handling and I/O
	return syscall.Exec(sshPath, args, os.Environ())
}

// BuildSSHCommand builds the SSH command string for a host
func BuildSSHCommand(host config.SSHHost) string {
	var parts []string
	parts = append(parts, "ssh")

	if host.User != "" {
		parts = append(parts, "-l", host.User)
	}

	if host.Port != "22" && host.Port != "" {
		parts = append(parts, "-p", host.Port)
	}

	if host.Identity != "" {
		parts = append(parts, "-i", host.Identity)
	}

	parts = append(parts, host.Host)

	return strings.Join(parts, " ")
}

// CopySSHCommand copies SSH command to clipboard
func CopySSHCommand(host config.SSHHost) error {
	command := BuildSSHCommand(host)
	return clipboard.WriteAll(command)
}

// ExecSSH replaces current process with SSH connection
// Deprecated: Use ConnectToHost instead
func ExecSSH(host config.SSHHost) error {
	return ConnectToHost(host)
}