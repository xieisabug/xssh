# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
# Build the project
go build -o xssh

# Run the application
./xssh

# Test the application with Go modules
go mod tidy
```

## Architecture Overview

XSSH is a Terminal User Interface (TUI) SSH connection manager built with Go and the Bubbletea framework. The application provides a comprehensive SSH host management system with automatic key setup.

### Core Architecture

**Entry Point Flow:**
- `main.go` initializes the Bubbletea TUI program
- After TUI exits, checks if a host was selected for connection
- Uses `syscall.Exec` to replace the current process with SSH for native terminal experience

**State Management:**
- `internal/ui/model.go` contains the main application state and Bubbletea Model
- Uses a `ViewMode` enum to manage different UI states (List, Add, Edit, Delete, Auth selection, etc.)
- Central state includes SSH config, host lists, form data, and UI state

**Multi-Mode UI System:**
The application operates in several distinct modes:
- `ModeList`: Main host list with search/filter capabilities
- `ModeAdd/ModeEdit`: Form-based host configuration
- `ModeAuthSelect`: Choose between password or SSH key authentication
- `ModeKeySelect`: Select from available SSH private keys
- `ModePasswordInput`: Secure password entry with asterisk masking
- `ModeConnectTest`: Automated SSH connection testing and key setup
- `ModeDelete`: Confirmation dialog for host deletion

**SSH Configuration Management:**
- `internal/config/ssh.go` handles parsing and writing SSH config files
- Supports standard SSH config format with regex-based parsing
- Maintains compatibility with existing SSH configurations

**SSH Connection and Setup:**
- `internal/ssh/client.go`: Basic SSH connection and command building
- `internal/ssh/setup.go`: Advanced SSH setup including:
  - Connection testing with password/key authentication
  - Automatic SSH key pair generation
  - Public key deployment (ssh-copy-id equivalent)
  - Key-based authentication verification

**View Rendering:**
- `internal/ui/views.go` contains specialized render functions for each mode
- Uses Lipgloss for styling and layout
- Implements responsive design patterns for different terminal sizes

### Key Design Patterns

**Form Flow Management:**
Adding a host follows a multi-step wizard pattern:
1. Basic info form (host, user, port, alias)
2. Authentication method selection
3. Credential input (password or key selection)
4. Automated connection testing and SSH key setup
5. Configuration persistence

**SSH Key Automation:**
When using password authentication, the application automatically:
- Tests initial password connection
- Generates RSA 2048-bit key pair if none exists
- Deploys public key to remote `~/.ssh/authorized_keys`
- Verifies key-based authentication works
- Saves configuration with key authentication enabled

**Process Replacement for SSH:**
Instead of spawning SSH as a subprocess, the application uses `syscall.Exec` to replace itself with the SSH process, ensuring perfect terminal compatibility for interactive sessions, full-screen applications, and proper signal handling.

## Module Dependencies

- **Bubbletea**: TUI framework for terminal applications
- **Lipgloss**: Styling and layout for terminal UIs  
- **golang.org/x/crypto/ssh**: SSH client implementation for connection testing
- **atotto/clipboard**: Cross-platform clipboard operations

## Configuration File Handling

The application reads/writes standard SSH config files at `~/.ssh/config`. The parser handles:
- Host declarations with multiple configuration options
- HostName, User, Port, and IdentityFile directives
- Preserves existing configurations and formatting
- Creates configuration if none exists